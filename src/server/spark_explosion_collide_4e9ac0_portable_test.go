package server

import (
	"reflect"
	"testing"
)

// This standard-library-only contract is also compiled as a standalone test
// package for every supported GOOS/GOARCH tuple.
func TestSparkExplosionCollidePortableContract4E9AC0(t *testing.T) {
	var events []string
	sparkExplosionCollide4E9AC0(1, 0, 99, sparkExplosionCollideHooks4E9AC0[int, uint8]{
		loadCollideData: func(source int) uint8 {
			events = append(events, "data")
			return 12
		},
		loadPower: func(power uint8) uint8 {
			events = append(events, "power")
			return power
		},
		gameFlagsCheck: func(flag uint32) int32 {
			if flag == sparkExplosionQuestFlag4E9AC0 {
				events = append(events, "quest")
				return 0
			}
			events = append(events, "coop")
			return 1
		},
		mapPushUnits: func(int, float32, float32, float32, int, int32, int32) {
			events = append(events, "push")
		},
		mapDamageUnits: func(int, float32, float32, int32, uint32, int, int) {
			events = append(events, "damage")
		},
		sparkFX: func(int, uint8) {
			events = append(events, "fx")
		},
		audio: func(uint32, int, int32, uint32) {
			events = append(events, "audio")
		},
		scorch: func(int, int32) {
			events = append(events, "scorch")
		},
		delayedDelete: func(int) {
			events = append(events, "delete")
		},
	})
	want := []string{
		"data", "quest", "power", "push", "coop", "power",
		"damage", "power", "fx", "audio", "scorch", "delete",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}
