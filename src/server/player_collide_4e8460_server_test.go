package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestPlayerCollide4E8460NativeDeathTransferUsesPowerArrayAndLiveRates(t *testing.T) {
	player := &Object{ObjClass: object.ClassPlayer}
	player.Buffs = uint32(1) << ENCHANT_DEATH
	player.BuffsDur[ENCHANT_DEATH] = 419
	player.BuffsPower[ENCHANT_DEATH] = 0xab
	other := &Object{ObjClass: object.ClassPlayer}
	collisionWord := uint32(0xa5a55a5a)
	frames := []uint32{30, 0x10001}
	var events []string

	playerCollideNative4E8460(player, other, unsafe.Pointer(&collisionWord), playerCollideNativeDeps4E8460{
		abilityActive: func(got *Object, ability Ability) bool {
			if got != player || ability != AbilityBerserk {
				t.Fatalf("ability args = (%p, %v), want (%p, %v)", got, ability, player, AbilityBerserk)
			}
			events = append(events, "ability")
			return false
		},
		frameRate: func() uint32 {
			events = append(events, "frame")
			value := frames[0]
			frames = frames[1:]
			return value
		},
		applyEnchant: func(got *Object, enchant EnchantID, duration, power uint32) {
			if got != other || enchant != ENCHANT_DEATH || duration != 15 || power != 0xab {
				t.Fatalf("apply args = (%p, %v, %d, %#x), want (%p, %v, 15, 0xab)",
					got, enchant, duration, power, other, ENCHANT_DEATH)
			}
			events = append(events, "apply")
		},
		disableEnchant: func(got *Object, enchant EnchantID) {
			if got != player || enchant != ENCHANT_DEATH {
				t.Fatalf("disable args = (%p, %v), want (%p, %v)", got, enchant, player, ENCHANT_DEATH)
			}
			events = append(events, "disable")
		},
	})

	if want := []string{"ability", "frame", "frame", "apply", "disable"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if len(frames) != 0 {
		t.Fatalf("unused frame rates = %v", frames)
	}
	if collisionWord != 0xa5a55a5a {
		t.Fatalf("ignored collision word changed: %#x", collisionWord)
	}
}

func TestPlayerCollide4E8460NativeWallFieldAndPreviousPosition(t *testing.T) {
	wall := &Wall{Tile1: 9}
	update := &PlayerUpdateData{CollisionWall: wall}
	player := &Object{
		ObjClass:   object.ClassPlayer,
		PosVec:     types.Ptf(10, 20),
		PrevPos:    types.Ptf(-3, 7),
		UpdateData: unsafe.Pointer(update),
	}
	var events []string
	playerCollideNative4E8460(player, nil, nil, playerCollideNativeDeps4E8460{
		abilityActive: func(got *Object, ability Ability) bool {
			if got != player || ability != AbilityBerserk {
				t.Fatal("ability arguments changed")
			}
			events = append(events, "ability")
			return true
		},
		setState: func(got *Object, state PlayerState) {
			if got != player || state != PlayerState(playerCollideState4E8460) {
				t.Fatal("state arguments changed")
			}
			events = append(events, "state")
		},
		earthquake: func(pos types.Pointf, amount int) {
			if pos != player.PosVec || amount != 10 {
				t.Fatalf("earthquake = (%v, %d), want (%v, 10)", pos, amount, player.PosVec)
			}
			events = append(events, "earthquake")
		},
		disableAbility: func(got *Object, ability Ability) {
			if got != player || ability != AbilityBerserk {
				t.Fatal("disable ability arguments changed")
			}
			events = append(events, "disable")
		},
		wallFlags: func(tile uint8) uint32 {
			if tile != wall.Tile1 {
				t.Fatalf("wall tile = %d, want %d", tile, wall.Tile1)
			}
			events = append(events, "wall")
			return 0
		},
		move: func(got *Object, pos types.Pointf) {
			if got != player || pos != player.PrevPos {
				t.Fatalf("move = (%p, %v), want (%p, %v)", got, pos, player, player.PrevPos)
			}
			events = append(events, "move")
		},
	})
	if want := []string{"ability", "state", "earthquake", "disable", "wall", "move"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestPlayerBerserkBounce4E86E0NativeObjectFields(t *testing.T) {
	player := &Object{Mass: 2, VelVec: types.Ptf(2, 4)}
	other := &Object{Mass: 3, VelVec: types.Ptf(6, 8)}
	playerBerserkBounceNative4E86E0(player, other)
	for name, tc := range map[string]struct {
		got      float32
		wantBits uint32
	}{
		"player X": {player.VelVec.X, 0x40d9999a},
		"player Y": {player.VelVec.Y, 0x410ccccd},
		"other X":  {other.VelVec.X, 0x40333333},
		"other Y":  {other.VelVec.Y, 0x4099999a},
	} {
		if got := math.Float32bits(tc.got); got != tc.wantBits {
			t.Errorf("%s bits = %08x, want %08x", name, got, tc.wantBits)
		}
	}
}

func TestPlayerCollide4E8460NativeLayoutFields(t *testing.T) {
	ptrSize := unsafe.Sizeof(uintptr(0))
	type field struct {
		name       string
		got        uintptr
		want32     uintptr
		wantNative uintptr
	}
	fields := []field{
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), 8, 12},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), 16, 20},
		{"Object.PosVec", unsafe.Offsetof(Object{}.PosVec), 56, 60},
		{"Object.NewPos", unsafe.Offsetof(Object{}.NewPos), 64, 68},
		{"Object.PrevPos", unsafe.Offsetof(Object{}.PrevPos), 72, 76},
		{"Object.VelVec", unsafe.Offsetof(Object{}.VelVec), 80, 84},
		{"Object.Mass", unsafe.Offsetof(Object{}.Mass), 120, 124},
		{"Object.Buffs", unsafe.Offsetof(Object{}.Buffs), 340, 344},
		{"Object.BuffsDur", unsafe.Offsetof(Object{}.BuffsDur), 344, 348},
		{"Object.BuffsPower", unsafe.Offsetof(Object{}.BuffsPower), 408, 412},
		{"Object.HealthData", unsafe.Offsetof(Object{}.HealthData), 556, 616},
		{"Object.Damage", unsafe.Offsetof(Object{}.Damage), 716, 808},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), 748, 872},
		{"PlayerUpdateData.CollisionWall", unsafe.Offsetof(PlayerUpdateData{}.CollisionWall), 296, 376},
		{"Wall.Data", unsafe.Offsetof(Wall{}.Data), 28, 40},
		{"Wall.ClientData", unsafe.Offsetof(Wall{}.ClientData), 32, 48},
	}
	for _, f := range fields {
		want := f.wantNative
		if ptrSize == 4 {
			want = f.want32
		}
		if f.got != want {
			t.Errorf("%s offset = %d, want %d for %d-bit", f.name, f.got, want, ptrSize*8)
		}
	}
	if unsafe.Offsetof(HealthData{}.Cur) != 0 || unsafe.Offsetof(HealthData{}.Max) != 4 {
		t.Fatalf("HealthData current/max offsets = %d/%d, want 0/4",
			unsafe.Offsetof(HealthData{}.Cur), unsafe.Offsetof(HealthData{}.Max))
	}
	if unsafe.Offsetof(Wall{}.Tile1) != 1 {
		t.Fatalf("Wall.Tile1 offset = %d, want 1", unsafe.Offsetof(Wall{}.Tile1))
	}
	wantWallSize := uintptr(56)
	if ptrSize == 4 {
		wantWallSize = 36
	}
	if got := unsafe.Sizeof(Wall{}); got != wantWallSize {
		t.Fatalf("Wall size = %d, want %d", got, wantWallSize)
	}
}
