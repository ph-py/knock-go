//go:build !windows

package cli

func init() {
	// Virtual terminal processing is natively supported on Unix/Linux/macOS.
}
