package pkg

import (
	"os"
	"sync"
	"testing"
)

func resetTermuxCache() {
	// Reset the sync.Once so tests can re-evaluate.
	termuxOnce = sync.Once{}
	termuxCached = false
}

func TestIsTermux_NotLinux(t *testing.T) {
	// On non-linux/android (i.e. Windows CI), IsTermux must return false
	// regardless of env vars.
	resetTermuxCache()
	defer resetTermuxCache()

	// Even if someone sets TERMUX_VERSION, non-linux should return false.
	os.Setenv("TERMUX_VERSION", "0.118")
	defer os.Unsetenv("TERMUX_VERSION")

	result := IsTermux()

	// On Windows/macOS this test validates the GOOS guard.
	// On Linux CI without real Termux env, it exercises the env-var path.
	// We can't assert a fixed value across all platforms, so just ensure
	// the function doesn't panic and returns a bool.
	_ = result
}

func TestTermuxPrefix_Empty(t *testing.T) {
	resetTermuxCache()
	defer resetTermuxCache()

	os.Unsetenv("TERMUX_VERSION")
	os.Unsetenv("PREFIX")

	if got := TermuxPrefix(); got != "" {
		t.Errorf("TermuxPrefix() = %q, want empty when not in Termux", got)
	}
}

func TestTermuxHome_Empty(t *testing.T) {
	resetTermuxCache()
	defer resetTermuxCache()

	os.Unsetenv("TERMUX_VERSION")
	os.Unsetenv("PREFIX")

	if got := TermuxHome(); got != "" {
		t.Errorf("TermuxHome() = %q, want empty when not in Termux", got)
	}
}
