package opennox

import "testing"

func TestE2EGolemNativeCommandIsolation(t *testing.T) {
	if e2ePlay == "" && consoleCommandAt(noxConsole, "e2e-golem-native") != nil {
		t.Fatal("Golem diagnostic command leaked into ordinary play")
	}
}
