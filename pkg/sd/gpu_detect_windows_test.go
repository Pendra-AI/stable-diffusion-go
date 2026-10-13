//go:build windows

package sd

import "testing"

// TestDisplayAdapters_ReadsTheDeviceManager runs the real SetupAPI read. Every
// Windows machine, a CI runner included, has at least one display-class device
// (the basic or virtual adapter when there is no GPU).
func TestDisplayAdapters_ReadsTheDeviceManager(t *testing.T) {
	adapters, err := displayAdapters()
	if err != nil {
		t.Fatalf("displayAdapters: %v", err)
	}
	if len(adapters) == 0 {
		t.Fatal("no display adapter listed; expected at least the basic one")
	}
	for _, a := range adapters {
		t.Logf("adapter %q %v -> vendor %q", a.name, a.hardwareIDs, adapterGPUVendor(a))
	}
	// GetGPUName must answer from that read, with no wmic on the PATH.
	t.Setenv("PATH", "")
	got, err := GetGPUName()
	if err != nil {
		t.Fatalf("GetGPUName without a PATH: %v", err)
	}
	if want := gpuVendorFromAdapters(adapters); got != want {
		t.Errorf("GetGPUName = %q, want %q", got, want)
	}
}
