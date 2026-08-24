package discovery

import "os"

func osUserHomeDirForTest() (string, error) {
	return os.UserHomeDir()
}