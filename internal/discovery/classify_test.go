package discovery

import "testing"

func TestClassify(t *testing.T) {
	home, _ := homeDirForTest(t)

	tests := []struct {
		name       string
		proc       ProcInfo
		parent     *ProcInfo
		wantIsDev  bool
		wantConf   Confidence
	}{
		{
			name:      "system daemon under /usr/libexec",
			proc:      ProcInfo{Exe: "/usr/libexec/rapportd", Cwd: "/"},
			wantIsDev: false,
			wantConf:  ConfidenceHigh,
		},
		{
			name:      "macOS system framework path",
			proc:      ProcInfo{Exe: "/System/Library/CoreServices/ControlCenter", Cwd: "/"},
			wantIsDev: false,
			wantConf:  ConfidenceHigh,
		},
		{
			name:      "no cwd available (permission denied)",
			proc:      ProcInfo{Exe: "/usr/bin/something", Cwd: ""},
			wantIsDev: false,
			wantConf:  ConfidenceNone,
		},
		{
			name:      "dev server: home cwd + shell parent",
			proc:      ProcInfo{Exe: "/usr/local/bin/node", Cwd: home + "/projects/events-web"},
			parent:    &ProcInfo{Exe: "/bin/zsh"},
			wantIsDev: true,
			wantConf:  ConfidenceHigh,
		},
		{
			name:      "home cwd, no known parent",
			proc:      ProcInfo{Exe: "/usr/local/bin/node", Cwd: home + "/projects/events-web"},
			parent:    &ProcInfo{Exe: "/sbin/launchd"},
			wantIsDev: true,
			wantConf:  ConfidenceMedium,
		},
		{
			name:      "shell parent but cwd outside home",
			proc:      ProcInfo{Exe: "/usr/local/bin/somecli", Cwd: "/opt/tooling"},
			parent:    &ProcInfo{Exe: "/bin/bash"},
			wantIsDev: true,
			wantConf:  ConfidenceLow,
		},
		{
			name:      "no signals at all",
			proc:      ProcInfo{Exe: "/opt/tooling/bin/thing", Cwd: "/opt/tooling"},
			parent:    &ProcInfo{Exe: "/sbin/launchd"},
			wantIsDev: false,
			wantConf:  ConfidenceLow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Classify(tt.proc, tt.parent)
			if got.IsDev != tt.wantIsDev {
				t.Errorf("IsDev = %v, want %v (reason: %s)", got.IsDev, tt.wantIsDev, got.Reason)
			}
			if got.Confidence != tt.wantConf {
				t.Errorf("Confidence = %v, want %v (reason: %s)", got.Confidence, tt.wantConf, got.Reason)
			}
		})
	}
}

func homeDirForTest(t *testing.T) (string, error) {
	t.Helper()
	home, err := osUserHomeDirForTest()
	if err != nil {
		t.Fatalf("could not resolve home dir for test: %v", err)
	}
	return home, nil
}