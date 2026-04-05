package pkg

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// Termux environment detection utilities.
//
// Termux sets $PREFIX (typically /data/data/com.termux/files/usr) and
// $TERMUX_VERSION.  These functions enable PicoClaw to automatically
// detect the Termux runtime and adapt paths/resource limits accordingly.

var (
	termuxOnce   sync.Once
	termuxCached bool
)

// IsTermux reports whether PicoClaw is running inside a Termux environment.
// The result is cached after the first call.
func IsTermux() bool {
	termuxOnce.Do(func() {
		if runtime.GOOS != "linux" && runtime.GOOS != "android" {
			termuxCached = false
			return
		}
		// Primary signal: $TERMUX_VERSION is always set by the Termux app.
		if os.Getenv("TERMUX_VERSION") != "" {
			termuxCached = true
			return
		}
		// Fallback: $PREFIX points inside the Termux file tree.
		prefix := os.Getenv("PREFIX")
		if prefix != "" && filepath.Base(filepath.Dir(prefix)) == "files" {
			termuxCached = true
			return
		}
		termuxCached = false
	})
	return termuxCached
}

// TermuxPrefix returns the Termux $PREFIX path (e.g. /data/data/com.termux/files/usr),
// or an empty string when not running in Termux.
func TermuxPrefix() string {
	if !IsTermux() {
		return ""
	}
	return os.Getenv("PREFIX")
}

// TermuxHome returns the Termux home directory ($HOME inside Termux,
// typically /data/data/com.termux/files/home), or an empty string
// when not running in Termux.
func TermuxHome() string {
	if !IsTermux() {
		return ""
	}
	// Inside Termux, $HOME is always set.
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	// Derive from PREFIX as fallback: PREFIX/../home
	prefix := os.Getenv("PREFIX")
	if prefix != "" {
		return filepath.Join(filepath.Dir(prefix), "home")
	}
	return ""
}
