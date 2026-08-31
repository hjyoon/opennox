package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/common/sound"
)

func TestPlayerScheduledSpellNative4FB0E0Layouts(t *testing.T) {
	wantObjectUpdate := uintptr(748)
	wantTrapSpells := uintptr(192)
	wantTrapCount := uintptr(212)
	wantPosX := uintptr(220)
	wantPosY := uintptr(224)
	wantPlayer := uintptr(276)
	wantPlayerIndex := uintptr(2064)
	wantArgSize := uintptr(12)
	wantArgPos := uintptr(4)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectUpdate = 872
		wantTrapSpells = 244
		wantTrapCount = 264
		wantPosX = 272
		wantPosY = 280
		wantPlayer = 336
		wantPlayerIndex = 2068
		wantArgSize = 16
		wantArgPos = 8
	}

	for _, check := range []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantObjectUpdate},
		{"PlayerUpdateData.TrapSpells", unsafe.Offsetof(PlayerUpdateData{}.TrapSpells), wantTrapSpells},
		{"PlayerUpdateData.TrapSpellsCnt", unsafe.Offsetof(PlayerUpdateData{}.TrapSpellsCnt), wantTrapCount},
		{"PlayerUpdateData.Field55", unsafe.Offsetof(PlayerUpdateData{}.Field55), wantPosX},
		{"PlayerUpdateData.Field56", unsafe.Offsetof(PlayerUpdateData{}.Field56), wantPosY},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantPlayer},
		{"Player.PlayerInd", unsafe.Offsetof(Player{}.PlayerInd), wantPlayerIndex},
		{"SpellAcceptArg size", unsafe.Sizeof(SpellAcceptArg{}), wantArgSize},
		{"SpellAcceptArg.Obj", unsafe.Offsetof(SpellAcceptArg{}.Obj), 0},
		{"SpellAcceptArg.Pos", unsafe.Offsetof(SpellAcceptArg{}.Pos), wantArgPos},
	} {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestPlayerDoScheduledSpellNative4FB0E0PreservesPointersAndCoordinates(t *testing.T) {
	player := &Player{PlayerInd: 7}
	update := &PlayerUpdateData{
		TrapSpells:    [5]uint32{11, 22, 33, 44, 55},
		TrapSpellsCnt: 0xaabbcc03,
		Field55:       -2147483648,
		Field56:       16777217,
		Player:        player,
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wideX := int64(1)<<32 | 123
		wideY := -(int64(1) << 32) - 456
		update.Field55 = int(wideX)
		update.Field56 = int(wideY)
	}
	unit := &Object{UpdateData: unsafe.Pointer(update)}
	target := &Object{}

	var checkedUnit *Object
	var castUnit *Object
	var castArg *SpellAcceptArg
	var castID spell.ID
	deps := playerScheduledSpellNativeDeps4FB0E0{
		checkSpell: func(gotUnit *Object, id uint32, bypass int32) int32 {
			checkedUnit = gotUnit
			if id != 11 || bypass != 0 {
				t.Fatalf("check = %d/%d, want 11/0", id, bypass)
			}
			update.TrapSpells[0] = 99
			return 0
		},
		informText: func(index ntype.PlayerInd, code byte, value int) {
			t.Fatal("successful cast informed an error")
		},
		audioEvent: func(id sound.ID, gotUnit *Object, kind int, code uint32) {
			t.Fatal("successful cast played fizzle audio")
		},
		castSpell: func(id spell.ID, gotUnit *Object, arg *SpellAcceptArg) {
			castID = id
			castUnit = gotUnit
			castArg = arg
			update.TrapSpellsCnt = update.TrapSpellsCnt&^0xff | 2
		},
	}

	if got := playerDoScheduledSpellNative4FB0E0(unit, target, deps); got != 1 {
		t.Fatalf("native result = %d, want 1", got)
	}
	if checkedUnit != unit || castUnit != unit || castID != 99 {
		t.Fatalf("native pointers/id = %p/%p/%d, want %p/%p/99", checkedUnit, castUnit, castID, unit, unit)
	}
	if castArg == nil || castArg.Obj != target {
		t.Fatalf("native arg = %#v, want target %p", castArg, target)
	}
	if got, want := castArg.Pos.X, float32(int32(update.Field55)); got != want {
		t.Fatalf("native X = %g, want %g", got, want)
	}
	if got, want := castArg.Pos.Y, float32(int32(update.Field56)); got != want {
		t.Fatalf("native Y = %g, want %g", got, want)
	}
	if got, want := update.TrapSpells, ([5]uint32{22, 0, 33, 44, 55}); got != want {
		t.Fatalf("native spells = %v, want %v", got, want)
	}
	if got := update.TrapSpellsCnt; got != 0xaabbcc01 {
		t.Fatalf("native count = %#x, want 0xaabbcc01", got)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, ptr := range map[string]unsafe.Pointer{
			"unit": unsafe.Pointer(unit), "target": unsafe.Pointer(target),
			"update": unsafe.Pointer(update), "player": unsafe.Pointer(player),
			"spell-arg": unsafe.Pointer(castArg),
		} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatalf("%s pointer = %p, want native address above 4 GiB", name, ptr)
			}
		}
	}
	runtime.KeepAlive(player)
	runtime.KeepAlive(update)
	runtime.KeepAlive(unit)
	runtime.KeepAlive(target)
	runtime.KeepAlive(castArg)
}

func TestPlayerDoScheduledSpellQueueNative4FB1D0MapsRejection(t *testing.T) {
	player := &Player{PlayerInd: 0xfe}
	update := &PlayerUpdateData{
		TrapSpells:    [5]uint32{71, 72, 73, 74, 75},
		TrapSpellsCnt: 0x12345602,
		Player:        player,
	}
	unit := &Object{UpdateData: unsafe.Pointer(update)}
	target := &Object{}

	var informed bool
	var sounded bool
	deps := playerScheduledSpellNativeDeps4FB0E0{
		checkSpell: func(gotUnit *Object, id uint32, bypass int32) int32 {
			if gotUnit != unit || id != 72 || bypass != 0 {
				t.Fatalf("check = %p/%d/%d", gotUnit, id, bypass)
			}
			return 17
		},
		informText: func(index ntype.PlayerInd, code byte, value int) {
			informed = true
			if index != 0xfe || code != 0 || value != 17 {
				t.Fatalf("inform = %d/%d/%d", index, code, value)
			}
		},
		audioEvent: func(id sound.ID, gotUnit *Object, kind int, code uint32) {
			sounded = true
			if id != sound.SoundPermanentFizzle || gotUnit != unit || kind != 0 || code != 0 {
				t.Fatalf("audio = %d/%p/%d/%d", id, gotUnit, kind, code)
			}
		},
		castSpell: func(spell.ID, *Object, *SpellAcceptArg) {
			t.Fatal("rejected spell was cast")
		},
	}

	if got := playerDoScheduledSpellQueueNative4FB1D0(unit, target, deps); got != 1 {
		t.Fatalf("native LIFO result = %d, want 1", got)
	}
	if !informed || !sounded {
		t.Fatalf("rejection callbacks = inform:%v audio:%v", informed, sounded)
	}
	if got, want := update.TrapSpells, ([5]uint32{71, 72, 73, 74, 75}); got != want {
		t.Fatalf("native LIFO spells = %v, want %v", got, want)
	}
	if got := update.TrapSpellsCnt; got != 0x12345601 {
		t.Fatalf("native LIFO count = %#x, want 0x12345601", got)
	}
}
