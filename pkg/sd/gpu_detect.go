package sd

import "strings"

// displayAdapter is one display adapter the operating system has a present
// device for.
type displayAdapter struct {
	// name is the device description ("NVIDIA L4").
	name string
	// hardwareIDs are the bus identifiers ("PCI\VEN_10DE&DEV_27B8&..."). A
	// software adapter (Microsoft Basic Display, Hyper-V video) has none on the
	// PCI bus.
	hardwareIDs []string
}

// pciGPUVendors maps a PCI vendor id to the name GetGPUName reports for it.
var pciGPUVendors = []struct{ id, vendor string }{
	{"VEN_10DE", "NVIDIA"},
	{"VEN_1002", "AMD"},
	{"VEN_1022", "AMD"},
	{"VEN_8086", "Intel"},
}

// gpuVendorFromAdapters names the vendor of the hardware GPU among adapters:
// "NVIDIA", "AMD", "Intel", or "" when there is none. With several GPUs NVIDIA
// and AMD win over Intel, whose adapter is usually the integrated one, so a
// laptop that lists its integrated GPU first still gets its discrete card's
// build. The PCI vendor id decides, so a virtual adapter never counts.
func gpuVendorFromAdapters(adapters []displayAdapter) string {
	found := map[string]bool{}
	for _, a := range adapters {
		if v := adapterGPUVendor(a); v != "" {
			found[v] = true
		}
	}
	for _, v := range []string{"NVIDIA", "AMD", "Intel"} {
		if found[v] {
			return v
		}
	}
	return ""
}

func adapterGPUVendor(a displayAdapter) string {
	paravirtual := false
	for _, id := range a.hardwareIDs {
		id = strings.ToUpper(id)
		if !strings.HasPrefix(id, `PCI\`) {
			continue
		}
		for _, p := range pciGPUVendors {
			if strings.Contains(id, p.id) {
				return p.vendor
			}
		}
		// A host GPU shared into a Hyper-V guest is a PCI device of Microsoft's
		// that carries the host card's name.
		if strings.Contains(id, "VEN_1414") {
			paravirtual = true
		}
	}
	if !paravirtual {
		return ""
	}
	name := strings.ToUpper(a.name)
	switch {
	case strings.Contains(name, "NVIDIA"):
		return "NVIDIA"
	case strings.Contains(name, "AMD"), strings.Contains(name, "RADEON"):
		return "AMD"
	case strings.Contains(name, "INTEL"):
		return "Intel"
	}
	return ""
}
