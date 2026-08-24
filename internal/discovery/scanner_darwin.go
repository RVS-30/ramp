//go:build darwin

package discovery

/*
#define PRIVATE 1
#define __APPLE_API_PRIVATE 1

#include <libproc.h>
#include <sys/proc_info.h>
#include <sys/socket.h>
#include <sys/types.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <stdlib.h>
*/
import "C"

import (
	"context"
	"fmt"
	"unsafe"
)

type darwinScanner struct{}

// newOSScanner returns the darwin implementation of Scanner.
func newOSScanner() Scanner {
	return &darwinScanner{}
}

// ListPIDs asks libproc for every visible PID. Two-call idiom: first
// call with a nil buffer returns the required buffer size in bytes;
// second call fills it. We pad the size a little since processes can
// come and go between the two calls (race, expected and harmless).
func (d *darwinScanner) ListPIDs(ctx context.Context) ([]int, error) {
	size := C.proc_listpids(C.PROC_ALL_PIDS, 0, nil, 0)
	if size <= 0 {
		return nil, fmt.Errorf("proc_listpids: size query failed")
	}

	bufBytes := int(size) + int(size)/5 + 64
	buf := C.malloc(C.size_t(bufBytes))
	if buf == nil {
		return nil, fmt.Errorf("proc_listpids: allocation failed")
	}
	defer C.free(buf)

	written := C.proc_listpids(C.PROC_ALL_PIDS, 0, buf, C.int(bufBytes))
	if written <= 0 {
		return nil, fmt.Errorf("proc_listpids: fill call failed")
	}

	count := int(written) / int(unsafe.Sizeof(C.pid_t(0)))
	pids := unsafe.Slice((*C.pid_t)(buf), count)

	out := make([]int, 0, count)
	for _, p := range pids {
		if p > 0 {
			out = append(out, int(p))
		}
	}
	return out, nil
}

// Enrich fills exe path, cwd, ppid, and listening ports for one PID.
// Missing/unreadable fields are left as zero values, not errors —
// permission-denied on another user's process is expected and
// handled by the classifier, not here.
func (d *darwinScanner) Enrich(ctx context.Context, pid int) (ProcInfo, error) {
	info := ProcInfo{PID: pid}

	pathBuf := make([]byte, C.PROC_PIDPATHINFO_MAXSIZE)
	n := C.proc_pidpath(C.int(pid), unsafe.Pointer(&pathBuf[0]), C.PROC_PIDPATHINFO_MAXSIZE)
	if n <= 0 {
		return ProcInfo{}, ErrProcessVanished
	}
	info.Exe = string(pathBuf[:n])

	var bsdInfo C.struct_proc_bsdinfo
	bn := C.proc_pidinfo(C.int(pid), C.PROC_PIDTBSDINFO, 0,
		unsafe.Pointer(&bsdInfo), C.int(unsafe.Sizeof(bsdInfo)))
	if bn > 0 {
		info.PPID = int(bsdInfo.pbi_ppid)
	}

	var vnodeInfo C.struct_proc_vnodepathinfo
	vn := C.proc_pidinfo(C.int(pid), C.PROC_PIDVNODEPATHINFO, 0,
		unsafe.Pointer(&vnodeInfo), C.int(unsafe.Sizeof(vnodeInfo)))
	if vn > 0 {
		info.Cwd = C.GoString(&vnodeInfo.pvi_cdir.vip_path[0])
	}

	info.Ports = listeningTCPPorts(pid)

	return info, nil
}

// listeningTCPPorts returns the local ports this PID is listening on
// (TCP sockets in LISTEN state only).
func listeningTCPPorts(pid int) []int {
	size := C.proc_pidinfo(C.int(pid), C.PROC_PIDLISTFDS, 0, nil, 0)
	if size <= 0 {
		return nil
	}

	buf := C.malloc(C.size_t(size))
	if buf == nil {
		return nil
	}
	defer C.free(buf)

	n := C.proc_pidinfo(C.int(pid), C.PROC_PIDLISTFDS, 0, buf, size)
	if n <= 0 {
		return nil
	}

	fdCount := int(n) / int(unsafe.Sizeof(C.struct_proc_fdinfo{}))
	fds := unsafe.Slice((*C.struct_proc_fdinfo)(buf), fdCount)

	var ports []int
	seen := make(map[int]bool)

	for _, fd := range fds {
		if fd.proc_fdtype != C.PROX_FDTYPE_SOCKET {
			continue
		}

		var sockInfo C.struct_socket_fdinfo
		sn := C.proc_pidfdinfo(C.int(pid), fd.proc_fd, C.PROC_PIDFDSOCKETINFO,
			unsafe.Pointer(&sockInfo), C.int(unsafe.Sizeof(sockInfo)))
		if sn <= 0 {
			continue
		}

		if sockInfo.psi.soi_kind != C.SOCKINFO_TCP {
			continue
		}
		tcpInfo := (*C.struct_tcp_sockinfo)(unsafe.Pointer(&sockInfo.psi.soi_proto[0]))
		if tcpInfo.tcpsi_state != C.TSI_S_LISTEN {
			continue
		}

		rawPort := uint16(tcpInfo.tcpsi_ini.insi_lport)
		localPort := int(rawPort>>8 | rawPort<<8)
		if localPort > 0 && !seen[localPort] {
			seen[localPort] = true
			ports = append(ports, localPort)
		}
	}

	return ports
}
