package opennox

import "testing"

func TestE2ESimplePolypNativeCommandIsolation(t *testing.T) {
	if e2ePlay == "" && consoleCommandAt(noxConsole, "e2e-simple-polyp-native") != nil {
		t.Fatal("simple/Polyp raw C diagnostic leaked into ordinary play")
	}
}
