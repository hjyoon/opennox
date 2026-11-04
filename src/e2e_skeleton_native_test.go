package opennox

import "testing"

func TestE2ESkeletonNativeCommandIsolation(t *testing.T) {
	if e2ePlay == "" && consoleCommandAt(noxConsole, "e2e-skeleton-native") != nil {
		t.Fatal("Skeleton diagnostic command leaked into ordinary play")
	}
}
