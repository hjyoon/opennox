package opennox

import "testing"

func TestE2ESkullRawNativeCommandIsolation(t *testing.T) {
	if e2ePlay == "" && consoleCommandAt(noxConsole, "e2e-skull-raw-native") != nil {
		t.Fatal("raw Skull diagnostic command leaked into ordinary play")
	}
}
