package discovery

import (
	"context"
	"os"
	"testing"
	"time"
)

func osUserHomeDirForTest() (string, error) {
	return os.UserHomeDir()
}

func contextWithTimeout(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(cancel)
	return ctx
}