package stable_diffusion

import "testing"

// The master-802 upstream resync removed several dedicated C struct fields and
// replaced them with free-form spec strings. These tests pin the Go-side
// translations to the exact strings upstream's own CLI produces, so the
// convenience booleans keep their pre-resync meaning.

func TestPrependBackendAssignment(t *testing.T) {
	cases := []struct {
		name       string
		spec       string
		assignment string
		want       string
	}{
		{"empty spec", "", "*=cpu", "*=cpu"},
		{"existing spec keeps its position after the assignment", "vulkan0", "te=cpu", "te=cpu,vulkan0"},
		{"stacked assignments", "te=cpu,vulkan0", "vae=cpu", "vae=cpu,te=cpu,vulkan0"},
	}
	for _, c := range cases {
		if got := prependBackendAssignment(c.spec, c.assignment); got != c.want {
			t.Errorf("%s: prependBackendAssignment(%q, %q) = %q, want %q", c.name, c.spec, c.assignment, got, c.want)
		}
	}
}

func TestBuildRefImageArgs(t *testing.T) {
	cases := []struct {
		name             string
		autoResize       bool
		increaseRefIndex bool
		extra            string
		want             string
	}{
		// The Go zero value (autoResize=false) matches the previous binding's
		// effective default, which unconditionally wrote the booleans into the
		// C struct: no auto-resize → resize_before_vae=0.
		{"zero values", false, false, "", "resize_before_vae=0"},
		{"auto-resize on", true, false, "", ""},
		{"increase ref index", true, true, "", "ref_index_mode=increase"},
		{"both flags translated", false, true, "", "resize_before_vae=0,ref_index_mode=increase"},
		{"extra args appended last", false, true, "k=v", "resize_before_vae=0,ref_index_mode=increase,k=v"},
		{"extra args alone", true, false, "k=v", "k=v"},
	}
	for _, c := range cases {
		if got := buildRefImageArgs(c.autoResize, c.increaseRefIndex, c.extra); got != c.want {
			t.Errorf("%s: buildRefImageArgs(%v, %v, %q) = %q, want %q", c.name, c.autoResize, c.increaseRefIndex, c.extra, got, c.want)
		}
	}
}
