//go:build windows

package sd

import (
	"errors"
	"fmt"

	"golang.org/x/sys/windows"
)

// displayClassGUID is the device setup class of display adapters,
// {4d36e968-e325-11ce-bfc1-08002be10318}.
var displayClassGUID = windows.GUID{
	Data1: 0x4d36e968, Data2: 0xe325, Data3: 0x11ce,
	Data4: [8]byte{0xbf, 0xc1, 0x08, 0x00, 0x2b, 0xe1, 0x03, 0x18},
}

// displayAdapters lists the display adapters that are present, read from the
// device manager (SetupAPI). It runs no program, so it does not depend on the
// PATH or on wmic. A data-centre card with no monitor output is a
// display-class device too, so it is listed.
func displayAdapters() ([]displayAdapter, error) {
	devs, err := windows.SetupDiGetClassDevsEx(&displayClassGUID, "", 0, windows.DIGCF_PRESENT, 0, "")
	if err != nil {
		return nil, fmt.Errorf("list display adapters: %w", err)
	}
	defer devs.Close()

	var out []displayAdapter
	for i := 0; ; i++ {
		info, err := devs.EnumDeviceInfo(i)
		if errors.Is(err, windows.ERROR_NO_MORE_ITEMS) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list display adapters: %w", err)
		}
		var a displayAdapter
		if v, err := devs.DeviceRegistryProperty(info, windows.SPDRP_DEVICEDESC); err == nil {
			a.name, _ = v.(string)
		}
		if v, err := devs.DeviceRegistryProperty(info, windows.SPDRP_HARDWAREID); err == nil {
			a.hardwareIDs, _ = v.([]string)
		}
		out = append(out, a)
	}
	return out, nil
}
