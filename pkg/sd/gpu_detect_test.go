package sd

import "testing"

func TestGPUVendorFromAdapters(t *testing.T) {
	basic := displayAdapter{name: "Microsoft Basic Display Adapter", hardwareIDs: []string{`ROOT\BasicDisplay`}}
	hyperv := displayAdapter{name: "Microsoft Hyper-V Video", hardwareIDs: []string{`VMBUS\{da0a7802-e377-4aac-8e77-0558eb1073f8}`}}
	l4 := displayAdapter{name: "NVIDIA L4", hardwareIDs: []string{`PCI\VEN_10DE&DEV_27B8&SUBSYS_16CA10DE&REV_A1`, `PCI\VEN_10DE&DEV_27B8`}}
	radeon := displayAdapter{name: "AMD Radeon RX 7900 XT", hardwareIDs: []string{`PCI\VEN_1002&DEV_744C&SUBSYS_0E3B1002`}}
	iris := displayAdapter{name: "Intel(R) Iris(R) Xe Graphics", hardwareIDs: []string{`PCI\VEN_8086&DEV_9A49`}}

	for _, tc := range []struct {
		name     string
		adapters []displayAdapter
		want     string
	}{
		{"none", nil, ""},
		{"software adapters only", []displayAdapter{basic, hyperv}, ""},
		{"a headless data-centre card beside the basic adapter", []displayAdapter{basic, l4}, "NVIDIA"},
		{"amd", []displayAdapter{radeon}, "AMD"},
		{"intel", []displayAdapter{iris}, "Intel"},
		{"a discrete card wins over the integrated one listed first", []displayAdapter{iris, l4}, "NVIDIA"},
		{"amd wins over intel", []displayAdapter{iris, radeon}, "AMD"},
		{"lower-case hardware id", []displayAdapter{{name: "x", hardwareIDs: []string{`pci\ven_10de&dev_1234`}}}, "NVIDIA"},
		{"a host gpu shared into a hyper-v guest", []displayAdapter{{name: "NVIDIA GeForce RTX 4090", hardwareIDs: []string{`PCI\VEN_1414&DEV_008E`}}}, "NVIDIA"},
		{"a virtual adapter named after a vendor", []displayAdapter{{name: "NVIDIA Virtual Display", hardwareIDs: []string{`ROOT\NVVAD`}}}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := gpuVendorFromAdapters(tc.adapters); got != tc.want {
				t.Errorf("gpuVendorFromAdapters = %q, want %q", got, tc.want)
			}
		})
	}
}
