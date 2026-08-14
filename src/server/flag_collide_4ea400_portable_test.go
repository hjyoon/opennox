package server

import "testing"

func TestFlagCollidePortableContract4EA400(t *testing.T) {
	events := make([]string, 0, 4)
	flags := map[uint32]int32{0x20: -1, 0x40: 1}
	flagCollide4EA400(1, 2, 3, flagCollideHooks4EA400[int, int]{
		loadFlags: func(target int) uint32 {
			events = append(events, "flags")
			if target != 2 {
				t.Fatalf("flags target = %d", target)
			}
			return 0
		},
		hasGameFlag: func(mask uint32) int32 {
			events = append(events, "mode")
			return flags[mask]
		},
		loadClassLow: func(target int) uint8 {
			events = append(events, "class")
			return 4
		},
		pickupCTF: func(source, target, collision int) {
			events = append(events, "ctf")
			if source != 1 || target != 2 || collision != 3 {
				t.Fatalf("CTF args = %d/%d/%d", source, target, collision)
			}
		},
		loadGameBallCache: func() uint32 { t.Fatal("CTF read FlagBall cache"); return 0 },
		lookupGameBall:    func(string) uint32 { t.Fatal("CTF looked up GameBall"); return 0 },
		storeGameBall:     func(uint32) { t.Fatal("CTF stored GameBall") },
		loadTypeInd:       func(int) uint16 { t.Fatal("CTF read TypeInd"); return 0 },
		pickupGameBall:    func(int, int, int) { t.Fatal("CTF called FlagBall") },
	})
	want := []string{"flags", "mode", "class", "ctf"}
	if len(events) != len(want) {
		t.Fatalf("events = %v", events)
	}
	for i := range want {
		if events[i] != want[i] {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}
}
