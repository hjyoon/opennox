package legacy

import (
	"bytes"
	"math"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/internal/cfixture"
	"github.com/opennox/opennox/v1/server"
)

func game4Owned[T comparable](t *testing.T, value T) *T {
	t.Helper()
	p, free := alloc.NewHigh(value)
	*p = value // alloc.New allocates, but deliberately does not copy its argument.
	t.Cleanup(free)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(p)) <= math.MaxUint32 {
		t.Fatalf("native allocation below 4GiB: %p", p)
	}
	return p
}

func game4PlayerBank(t *testing.T) (*server.Object, *server.PlayerUpdateData, *server.Player) {
	t.Helper()
	p := game4Owned(t, server.Player{})
	u := game4Owned(t, server.PlayerUpdateData{Player: p})
	o := game4Owned(t, server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(u)})
	return o, u, p
}

func TestGame4GoldCEntryNativeWordsAndUnsignedBoundaries(t *testing.T) {
	o, _, p := game4PlayerBank(t)
	p.ProtPlayerGold = 0x123 // Original protection function returns this opaque word.
	for _, tc := range []struct {
		initial, value, want uint32
		op                   int
	}{
		{math.MaxUint32, 2, 1, 0}, {4, 0xffffffff, 3, 0}, {7, 8, 0, 1}, {math.MaxUint32, 0x80000000, 0x7fffffff, 1},
		{7, 0x80000000, 0, 2}, {math.MaxUint32, 1, 0, 2}, {7, 0xfffffffe, 5, 2},
	} {
		p.GoldVal = tc.initial
		result := cfixture.Game4GoldC(o, tc.op, tc.value)
		if p.GoldVal != tc.want {
			t.Fatalf("gold %+v = %#x", tc, p.GoldVal)
		}
		if tc.op < 2 && result != 0x123 {
			t.Fatalf("opaque return=%#x", result)
		}
		if got := cfixture.Game4GoldC(o, 3, 0); got != uintptr(tc.want) {
			t.Fatalf("player getter=%#x", got)
		}
		if got := cfixture.Game4GoldC(o, 4, 0); got != uintptr(tc.want) {
			t.Fatalf("object getter=%#x", got)
		}
	}
	p.ProtPlayerGold = math.MaxUint32
	p.GoldVal = 2
	if got := cfixture.Game4GoldC(o, 0, math.MaxUint32); got != ^uintptr(0) || p.GoldVal != 1 {
		t.Fatalf("signed protection result after add = %#x, gold = %d", got, p.GoldVal)
	}
	if got := cfixture.Game4GoldC(o, 1, 2); got != ^uintptr(0) || p.GoldVal != 0 {
		t.Fatalf("signed protection result after subtract = %#x, gold = %d", got, p.GoldVal)
	}
	p.GoldVal = 0xdeadbeef
	o.ObjClass = object.ClassMonster
	if got := cfixture.Game4GoldC(o, 3, 0); got != 0xdeadbeef {
		t.Fatalf("player getter incorrectly class gated: %#x", got)
	}
	if got := cfixture.Game4GoldC(o, 4, 0); got != 0 {
		t.Fatalf("object class gate=%#x", got)
	}
	cfixture.Game4GoldC(o, 2, 7)
	if p.GoldVal != 0xdeadbeef {
		t.Fatal("nonplayer setter changed Gold")
	}
}

func TestGame4WaypointCEntryCallbackOrderAndNativeLinks(t *testing.T) {
	o, u, p := game4PlayerBank(t)
	a, b, c := game4Owned(t, server.Object{}), game4Owned(t, server.Object{}), game4Owned(t, server.Object{})
	old := game4WaypointServices
	t.Cleanup(func() { game4WaypointServices = old })
	var events []string
	r := game4WaypointRuntime{
		delete: func(w *server.Object) {
			runtime.GC()
			if w == a {
				events = append(events, "delete-a")
				u.CustomWaypoints[1] = c
				o.UpdateData = nil
			} else if w == c {
				events = append(events, "delete-c")
			} else {
				t.Fatalf("unexpected delete %p", w)
			}
		},
		move: func(w *server.Object, pos types.Pointf) {
			if w != b || pos != (types.Pointf{X: 1.25, Y: -2.5}) {
				t.Fatal("move lost object/position")
			}
			events = append(events, "move")
		},
		newObject: func(name string) *server.Object {
			if name != "PlayerWaypoint" {
				t.Fatal(name)
			}
			events = append(events, "new")
			return b
		},
		create: func(w, owner *server.Object, pos types.Pointf) {
			if w != b || owner != o || u.CustomWaypoints[2] != b || pos.X != 1.25 {
				t.Fatal("create linkage")
			}
			events = append(events, "create")
		},
	}
	game4WaypointServices = func() game4WaypointRuntime { return r }
	u.CustomWaypoints = [3]*server.Object{a, b, nil}
	u.CustomWaypointRead = 2
	u.CustomWaypointWrite = 1
	cfixture.Game4WaypointC(o, 0, 0, 0)
	if !reflect.DeepEqual(events, []string{"delete-a", "delete-c"}) || u.CustomWaypoints != ([3]*server.Object{}) || u.CustomWaypointRead != 0 || u.CustomWaypointWrite != 0 {
		t.Fatalf("cleanup: %v %+v", events, u)
	}
	o.UpdateData = unsafe.Pointer(u)
	u.CustomWaypointWrite = 2
	cfixture.Game4WaypointC(o, 1, 1.25, -2.5)
	cfixture.Game4WaypointC(o, 1, 1.25, -2.5)
	if !reflect.DeepEqual(events, []string{"delete-a", "delete-c", "new", "create", "move"}) {
		t.Fatal(events)
	}
	p.Field3680 = 3
	cfixture.Game4WaypointC(o, 1, 1.25, -2.5)
	if len(events) != 5 {
		t.Fatal("observer created/moved waypoint")
	}
	u.CustomWaypointRead = 2
	if cfixture.Game4WaypointC(o, 2, 0, 0) != 1 {
		t.Fatal("presence lost high pointer")
	}
}

func TestGame4WaypointFollowCFloatBoundaryAndLiveCursor(t *testing.T) {
	o, u, _ := game4PlayerBank(t)
	w := game4Owned(t, server.Object{})
	old := game4WaypointServices
	t.Cleanup(func() { game4WaypointServices = old })
	u.CustomWaypoints[0] = w
	var deleted int
	game4WaypointServices = func() game4WaypointRuntime {
		return game4WaypointRuntime{
			delete: func(got *server.Object) {
				if got != w {
					t.Fatal("wrong waypoint")
				}
				deleted++
				u.CustomWaypointRead = 1
			},
			buffed: func(got *server.Object) bool {
				if got != o {
					t.Fatal("buff pointer")
				}
				return true
			},
			confused: func(got *server.Object) server.Dir16 { return 17 },
		}
	}
	i := uintptr(8 * 17)
	xp, yp := memmap.PtrFloat32(0x587000, 194136+i), memmap.PtrFloat32(0x587000, 194140+i)
	xold, yold := *xp, *yp
	t.Cleanup(func() { *xp, *yp = xold, yold })
	*xp, *yp = 0.3125, -0.75
	// Branch and force words observed from GAME.EXE 004F9ADA..004F9B07
	// and 004F9B5F..004F9B87 with startup x87 control word 0x0e7f.
	tests := []struct {
		pos, target types.Pointf
		steer       bool
	}{
		{target: types.Pointf{9.999999, 0}},
		{target: types.Pointf{10, 0}, steer: true},
		{target: types.Pointf{0, 10}, steer: true},
		{target: types.Pointf{6, 8}, steer: true},
		{target: types.Pointf{-6, -8}, steer: true},
		{target: types.Pointf{0.001, 9.999999}},
		{target: types.Pointf{100000, 0}, steer: true},
		{},
		{target: types.Pointf{8.07499885559082, 5.898677349090576}, steer: true},
		{
			pos:    types.Pointf{-0.035242218524217606, -0.34607747197151184},
			target: types.Pointf{5.693376064300537, 7.850441932678223},
		},
	}
	for _, tc := range tests {
		u.CustomWaypointRead = 0
		u.CustomWaypoints[0] = w
		u.CustomWaypoints[1] = w
		o.PosVec = tc.pos
		w.PosVec = tc.target
		o.ForceVec = types.Pointf{X: 0.1, Y: -0.2}
		o.SpeedCur = 1.7
		wantForce := [2]uint32{0x3dcccccd, 0xbe4ccccd}
		if tc.steer {
			wantForce = [2]uint32{0x3f219999, 0xbfbccccd}
		}
		if got := cfixture.Game4WaypointC(o, 3, 0, 0); got != bool2int(tc.steer) {
			t.Fatalf("steer from %+v to %+v = %d, want %t", tc.pos, tc.target, got, tc.steer)
		}
		gotForce := [2]uint32{math.Float32bits(o.ForceVec.X), math.Float32bits(o.ForceVec.Y)}
		if gotForce != wantForce || tc.steer && o.Direction2 != 17 {
			t.Fatalf("force from %+v to %+v = %08x, want %08x, direction %d", tc.pos, tc.target, gotForce, wantForce, o.Direction2)
		}
		if !tc.steer && (u.CustomWaypoints[0] != w || u.CustomWaypoints[1] != nil) {
			t.Fatal("callback cursor/cache ordering")
		}
	}
	if deleted == 0 {
		t.Fatal("near branch not exercised")
	}
}

func TestGame4RewardCEntryPacketsBoundsAndGuideReload(t *testing.T) {
	o, u, p := game4PlayerBank(t)
	p.PlayerInd = 7
	target := game4Owned(t, server.Object{ObjClass: object.ClassMonster, NetCode: 0xabcd1234})
	old := game4RewardServices
	t.Cleanup(func() { game4RewardServices = old })
	var packets [][]byte
	queueResult := int32(-23)
	r := game4RewardRuntime{send: func(ind int, b []byte) int {
		if ind != 7 {
			t.Fatal(ind)
		}
		packets = append(packets, bytes.Clone(b))
		runtime.GC()
		return int(queueResult)
	}}
	game4RewardServices = func() game4RewardRuntime { return r }
	for kind := 0; kind < 3; kind++ {
		if got := cfixture.Game4RewardC(o, kind, target, 0xfe); got != uintptr(queueResult) {
			t.Fatal(got)
		}
		want := []byte{0xf0, byte(30 + kind), 0xfe, 0x34, 0x12}
		if !bytes.Equal(packets[len(packets)-1], want) {
			t.Fatal(packets)
		}
	}
	if cfixture.Game4RewardC(o, 8, nil, 0) != 6 || len(packets) != 3 {
		t.Fatal("invalid subtype")
	}
	invalidResult := int32(-9)
	if got := cfixture.Game4RewardC(o, -7, nil, 0); got != uintptr(invalidResult) {
		t.Fatalf("negative subtype returned %#x, want %#x", got, uintptr(invalidResult))
	}
	if got := cfixture.Game4RewardC(target, 1, nil, 0); got != uintptr(unsafe.Pointer(target)) {
		t.Fatalf("nonplayer reward returned %#x, want recipient %p", got, target)
	}
	if got := cfixture.Game4RewardC(nil, 1, nil, 0); got != 0 || len(packets) != 3 {
		t.Fatalf("nil recipient returned %#x with %d packets", got, len(packets))
	}
	for _, n := range []int{0, 1, 48, 49, 255, 256, 257} {
		before := len(packets)
		cfixture.Game4StringC(o, strings.Repeat("x", n), 0xf7)
		if n == 1 || n == 48 {
			if len(packets) != before+1 {
				t.Fatal(n)
			}
			b := packets[len(packets)-1]
			if len(b) != 52 || b[0] != 0xf0 || b[1] != 33 || b[51] != 0xf7 || !bytes.Equal(b[2:2+n], bytes.Repeat([]byte{'x'}, n)) || b[2+n] != 0 {
				t.Fatalf("packet n%d %x", n, b)
			}
			for _, padding := range b[3+n : 51] {
				if padding != 0 {
					t.Fatal("uninitialized packet padding")
				}
			}
		} else if len(packets) != before {
			t.Fatalf("unsafe name accepted length%d", n)
		}
	}
	other := game4Owned(t, server.Player{})
	other.BeastScrollLvl[4] = 9
	r.name = func(got *server.Object) string {
		if got != target {
			t.Fatal("creature ptr")
		}
		u.Player = other
		o.UpdateData = nil
		return "name"
	}
	r.guide = func(name string) int {
		if name != "name" {
			t.Fatal(name)
		}
		return 4
	}
	r.bonus = func() float64 { return 1.25 }
	r.floatInt = game3MapFloatInt
	if got := cfixture.Game4GuideC(o, target, nil); got != 9 {
		t.Fatalf("cached update/reloaded player=%d", got)
	}
	o.UpdateData = unsafe.Pointer(u)
	var guarded = struct {
		before uint32
		damage int32
		after  uint32
	}{0xaabbccdd, 8, 0xdeadbeef}
	var pin runtime.Pinner
	pin.Pin(&guarded)
	defer pin.Unpin()
	if got := cfixture.Game4DamageC(o, target, &guarded.damage); got != 10 || guarded.damage != 10 || guarded.before != 0xaabbccdd || guarded.after != 0xdeadbeef {
		t.Fatalf("damage %+v return%d", guarded, got)
	}
	// No guide must not touch the damage pointer or resolve a game-data bonus.
	o.UpdateData = unsafe.Pointer(u)
	r.guide = func(string) int { return 0 }
	r.bonus = func() float64 { t.Fatal("eager bonus"); return 0 }
	if got := cfixture.Game4DamageC(o, target, nil); got != 0 {
		t.Fatal(got)
	}
}

func TestGame4PhonemeCEntryAllSignedInputsAndSelection(t *testing.T) {
	p := game4Owned(t, server.Player{})
	old := game4PhonemeServices
	t.Cleanup(func() { game4PhonemeServices = old })
	male := map[int8]int32{0: 193, 1: 186, 2: 187, 3: 192, 5: 188, 6: 191, 7: 190, 8: 189}
	for _, isServer := range []bool{false, true} {
		for _, mode := range []int{0, 1, 2, 3} {
			calls := 0
			game4PhonemeServices = func() game4PhonemeRuntime {
				return game4PhonemeRuntime{server: isServer, class: func(code uint32, s bool) object.Class {
					if code != 0xfedcba98 || s != isServer {
						t.Fatal("netcode/mode width")
					}
					if mode == 0 {
						return object.ClassMonster
					}
					return object.ClassPlayer
				}, player: func(code uint32) *server.Player {
					calls++
					runtime.GC()
					if mode == 3 {
						return nil
					}
					p.Info().SetIsFemale(byte(mode - 1))
					return p
				}}
			}
			for n := -128; n < 128; n++ {
				want := male[int8(n)]
				if mode == 2 && want != 0 {
					want += 8
				}
				if mode == 3 {
					want = 0
				}
				if got := cfixture.Game4PhonemeC(0xfedcba98, int8(n)); got != want {
					t.Fatalf("server%v mode%d phoneme%d=%d want%d", isServer, mode, n, got, want)
				}
			}
			if mode == 0 && calls != 0 || mode != 0 && calls != 256 {
				t.Fatalf("player resolution calls%d", calls)
			}
		}
	}
}

func TestGame4AnimationCEntryBitPriorityAndActionGates(t *testing.T) {
	for bit := uint(2); bit < 27; bit++ {
		p := memmap.PtrUint32(0x587000, uintptr(215824+4*bit))
		old := *p
		*p = 0xf0000000 + uint32(bit)
		defer func() { *p = old }()
		for _, mask := range []uint32{1 << bit, (1 << bit) | 0xf8000003} {
			if got := uint32(cfixture.Game4AnimationC(mask)); got != *p {
				t.Fatalf("bit%d mask%x=%x", bit, mask, got)
			}
		}
	}
	if cfixture.Game4AnimationC(0xf8000003) != 0 {
		t.Fatal("excluded bits")
	}
	o, u, _ := game4PlayerBank(t)
	if cfixture.Game4WaypointC(o, 4, 0, 0) != 1 || cfixture.Game4WaypointC(o, 5, 0, 0) != 1 {
		t.Fatal("idle gates")
	}
	u.State = server.PlayerState23
	if cfixture.Game4WaypointC(o, 5, 0, 0) != 0 {
		t.Fatal("attack state gate")
	}
}
