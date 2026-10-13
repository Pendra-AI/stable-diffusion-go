//go:build !windows

package sd

import "errors"

// displayAdapters is Windows-only: the library is a single build on the other
// platforms, so nothing there chooses by GPU vendor.
func displayAdapters() ([]displayAdapter, error) {
	return nil, errors.New("display adapter listing is Windows-only")
}
