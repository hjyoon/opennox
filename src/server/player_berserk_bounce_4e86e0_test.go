package server

import (
	"math"
	"reflect"
	"testing"
)

type playerBerserkBounceTestObject4E86E0 struct {
	name string
	mass float32
	vel  [2]float32
}

func TestPlayerBerserkBounce4E86E0OrderAndResults(t *testing.T) {
	player := &playerBerserkBounceTestObject4E86E0{name: "player", mass: 2, vel: [2]float32{2, 4}}
	other := &playerBerserkBounceTestObject4E86E0{name: "other", mass: 3, vel: [2]float32{6, 8}}
	var events []string
	playerBerserkBounce4E86E0(player, other, playerBerserkBounceHooks4E86E0[*playerBerserkBounceTestObject4E86E0]{
		mass: func(obj *playerBerserkBounceTestObject4E86E0) float32 {
			events = append(events, "mass:"+obj.name)
			return obj.mass
		},
		velocity: func(obj *playerBerserkBounceTestObject4E86E0, axis int) float32 {
			events = append(events, "velocity:"+obj.name+string(rune('X'+axis)))
			return obj.vel[axis]
		},
		store: func(obj *playerBerserkBounceTestObject4E86E0, axis int, value float32) {
			events = append(events, "store:"+obj.name+string(rune('X'+axis)))
			obj.vel[axis] = value
		},
	})
	wantEvents := []string{
		"mass:player", "mass:other",
		"velocity:playerY", "velocity:otherY",
		"velocity:otherX", "velocity:playerX",
		"velocity:otherY", "velocity:playerY",
		"velocity:playerX", "velocity:otherX",
		"store:playerX", "store:playerY", "store:otherY", "store:otherX",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	for name, tc := range map[string]struct {
		got      float32
		wantBits uint32
	}{
		"player X": {player.vel[0], 0x40d9999a},
		"player Y": {player.vel[1], 0x410ccccd},
		"other X":  {other.vel[0], 0x40333333},
		"other Y":  {other.vel[1], 0x4099999a},
	} {
		if got := math.Float32bits(tc.got); got != tc.wantBits {
			t.Errorf("%s bits = %08x, want %08x", name, got, tc.wantBits)
		}
	}
}

func TestPlayerBerserkBounce4E86E0NilGuardsReadNothing(t *testing.T) {
	obj := &playerBerserkBounceTestObject4E86E0{}
	hooks := playerBerserkBounceHooks4E86E0[*playerBerserkBounceTestObject4E86E0]{
		mass:     func(*playerBerserkBounceTestObject4E86E0) float32 { t.Fatal("mass read"); return 0 },
		velocity: func(*playerBerserkBounceTestObject4E86E0, int) float32 { t.Fatal("velocity read"); return 0 },
		store:    func(*playerBerserkBounceTestObject4E86E0, int, float32) { t.Fatal("store") },
	}
	playerBerserkBounce4E86E0((*playerBerserkBounceTestObject4E86E0)(nil), obj, hooks)
	playerBerserkBounce4E86E0(obj, (*playerBerserkBounceTestObject4E86E0)(nil), hooks)
}

func TestPlayerBerserkBounce4E86E0EqualAndZeroMasses(t *testing.T) {
	for _, tc := range []struct {
		name       string
		mass       float32
		wantPlayer [2]float32
		wantOther  [2]float32
		wantNaN    bool
	}{
		{name: "equal masses swap velocities", mass: 2, wantPlayer: [2]float32{6, 8}, wantOther: [2]float32{2, 4}},
		{name: "zero total mass follows x87 unordered arithmetic", wantNaN: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			player := &playerBerserkBounceTestObject4E86E0{mass: tc.mass, vel: [2]float32{2, 4}}
			other := &playerBerserkBounceTestObject4E86E0{mass: tc.mass, vel: [2]float32{6, 8}}
			playerBerserkBounce4E86E0(player, other, playerBerserkBounceHooks4E86E0[*playerBerserkBounceTestObject4E86E0]{
				mass:     func(obj *playerBerserkBounceTestObject4E86E0) float32 { return obj.mass },
				velocity: func(obj *playerBerserkBounceTestObject4E86E0, axis int) float32 { return obj.vel[axis] },
				store:    func(obj *playerBerserkBounceTestObject4E86E0, axis int, value float32) { obj.vel[axis] = value },
			})
			if tc.wantNaN {
				for i, value := range [4]float32{player.vel[0], player.vel[1], other.vel[0], other.vel[1]} {
					if !math.IsNaN(float64(value)) {
						t.Errorf("velocity %d = %v, want NaN", i, value)
					}
				}
				return
			}
			if player.vel != tc.wantPlayer || other.vel != tc.wantOther {
				t.Fatalf("velocities = %v/%v, want %v/%v", player.vel, other.vel, tc.wantPlayer, tc.wantOther)
			}
		})
	}
}
