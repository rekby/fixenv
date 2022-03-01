package fixenv

import "testing"

func TestInterfaceCompatible(t *testing.T) {
	// Test compile and compatible T is subinterface of testing.TB
	var tb testing.TB = t
	var localT T = tb
	localT.Name()
}
