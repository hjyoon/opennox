package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
)

func requireNativeSpellDurationCancelPointers4FE9D0(t *testing.T, values ...unsafe.Pointer) {
	t.Helper()
	if unsafe.Sizeof(uintptr(0)) != 8 {
		return
	}
	for i, value := range values {
		if value == nil || uintptr(value) <= math.MaxUint32 {
			t.Fatalf("pointer %d = %p, want native address above 4 GiB", i, value)
		}
	}
}

func TestSpellDurationCancel4FE9D0NativeLayouts(t *testing.T) {
	wantObjectClass := uintptr(8)
	wantObjectUpdate := uintptr(748)
	wantUpdatePlayer := uintptr(276)
	wantPlayerIndex := uintptr(2064)
	wantDurSize := uintptr(120)
	wantCaster := uintptr(16)
	wantTarget := uintptr(48)
	wantFlags := uintptr(88)
	wantSub108 := uintptr(108)
	wantNext := uintptr(116)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectClass = 12
		wantObjectUpdate = 872
		wantUpdatePlayer = 336
		wantPlayerIndex = 2068
		wantDurSize = 184
		wantCaster = 24
		wantTarget = 72
		wantFlags = 120
		wantSub108 = 160
		wantNext = 176
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"DurSpell size", unsafe.Sizeof(DurSpell{}), wantDurSize},
		{"DurSpell.Spell", unsafe.Offsetof(DurSpell{}.Spell), 4},
		{"DurSpell.Caster16", unsafe.Offsetof(DurSpell{}.Caster16), wantCaster},
		{"DurSpell.Target48", unsafe.Offsetof(DurSpell{}.Target48), wantTarget},
		{"DurSpell.Flags88", unsafe.Offsetof(DurSpell{}.Flags88), wantFlags},
		{"DurSpell.Sub108", unsafe.Offsetof(DurSpell{}.Sub108), wantSub108},
		{"DurSpell.Next", unsafe.Offsetof(DurSpell{}.Next), wantNext},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantObjectClass},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantObjectUpdate},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantUpdatePlayer},
		{"Player.PlayerInd", unsafe.Offsetof(Player{}.PlayerInd), wantPlayerIndex},
		{"spell width", unsafe.Sizeof(DurSpell{}.Spell), 4},
		{"flags width", unsafe.Sizeof(DurSpell{}.Flags88), 4},
		{"player index width", unsafe.Sizeof(Player{}.PlayerInd), 1},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
	if got := uint32(spell.SPELL_CHAIN_LIGHTNING); got != spellDurationCancelChainLightning4FE9D0 {
		t.Fatalf("Chain Lightning = %d, want %d", got, spellDurationCancelChainLightning4FE9D0)
	}
}

func TestSpellDurationCancelNative4FE9D0PreservesPointersAndLiveReloads(t *testing.T) {
	root := &DurSpell{Spell: 8, Flags88: 0x12345640}
	child := new(DurSpell)
	stale := new(DurSpell)
	live := new(DurSpell)
	caster := &Object{ObjClass: object.ClassPlayer | object.Class(0x80)}
	target := new(Object)
	update := new(PlayerUpdateData)
	player := &Player{PlayerInd: 7}
	root.Caster16 = caster
	child.Target48 = target
	child.Next = stale
	caster.UpdateData = unsafe.Pointer(update)
	update.Player = player
	requireNativeSpellDurationCancelPointers4FE9D0(t,
		unsafe.Pointer(root), unsafe.Pointer(child), unsafe.Pointer(stale), unsafe.Pointer(live),
		unsafe.Pointer(caster), unsafe.Pointer(target), unsafe.Pointer(update), unsafe.Pointer(player),
	)

	var events []string
	deps := spellDurationCancelNativeDeps4FE9D0{
		loadCaster: func(record *DurSpell) *Object {
			events = append(events, "caster")
			return record.Caster16
		},
		loadClassLowByte: func(value *Object) byte {
			events = append(events, "class")
			return byte(value.ObjClass)
		},
		loadSpell: func(record *DurSpell) uint32 {
			events = append(events, "spell")
			return record.Spell
		},
		loadUpdate: func(value *Object) *PlayerUpdateData {
			events = append(events, "update")
			return (*PlayerUpdateData)(value.UpdateData)
		},
		loadPlayer: func(value *PlayerUpdateData) *Player {
			events = append(events, "player")
			return value.Player
		},
		loadPlayerIndex: func(value *Player) byte {
			events = append(events, "index")
			return value.PlayerInd
		},
		reportSpellStat: func(index byte, spellID uint32, status byte) {
			events = append(events, "report")
			if index != 7 || spellID != 8 || status != 15 {
				t.Fatalf("report = %d/%d/%d, want 7/8/15", index, spellID, status)
			}
			root.Spell = spellDurationCancelChainLightning4FE9D0
			root.Sub108 = child
		},
		loadSub108: func(record *DurSpell) *DurSpell {
			events = append(events, "sub108")
			return record.Sub108
		},
		loadTarget: func(record *DurSpell) *Object {
			events = append(events, "target")
			return record.Target48
		},
		stopRay: func(record *DurSpell, gotTarget *Object) {
			events = append(events, "stop")
			if record != child || gotTarget != target {
				t.Fatalf("stop ray = %p/%p, want %p/%p", record, gotTarget, child, target)
			}
			child.Next = live
			root.Flags88 = 0x89abcdee
		},
		loadNext: func(record *DurSpell) *DurSpell {
			events = append(events, "next")
			return record.Next
		},
		loadFlagsLowByte: func(record *DurSpell) byte {
			events = append(events, "flags")
			return byte(record.Flags88)
		},
		storeFlagsLowByte: func(record *DurSpell, flags byte) {
			events = append(events, "store")
			record.Flags88 = record.Flags88&^0xff | uint32(flags)
		},
	}

	got := spellDurationCancelNative4FE9D0(root, deps)
	wantEvents := []string{
		"caster", "class", "spell", "update", "player", "index", "report",
		"spell", "sub108", "target", "stop", "next", "target", "next", "flags", "store",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %q, want %q", events, wantEvents)
	}
	if child.Next != live {
		t.Fatalf("child.Next = %p, want live %p", child.Next, live)
	}
	if got != 0xef || root.Flags88 != 0x89abcdef {
		t.Fatalf("result/flags = %#x/%#x, want 0xef/0x89abcdef", got, root.Flags88)
	}
	runtime.KeepAlive(root)
	runtime.KeepAlive(child)
	runtime.KeepAlive(stale)
	runtime.KeepAlive(live)
	runtime.KeepAlive(caster)
	runtime.KeepAlive(target)
	runtime.KeepAlive(update)
	runtime.KeepAlive(player)
}

func TestSpellDurationCancel4FE9D0ServerBindingPreservesHighFlagBytes(t *testing.T) {
	record := &DurSpell{
		Spell:    8,
		Caster16: &Object{ObjClass: object.ClassMonster},
		Flags88:  0xfedcba20,
	}
	var durations SpellsDuration
	got := durations.SpellDurationCancel4FE9D0(record)
	if got != 0x21 || record.Flags88 != 0xfedcba21 {
		t.Fatalf("result/flags = %#x/%#x, want 0x21/0xfedcba21", got, record.Flags88)
	}
	durations.CancelSpell(record)
	if record.Flags88 != 0xfedcba21 {
		t.Fatalf("alias flags = %#x, want 0xfedcba21", record.Flags88)
	}
}

func TestSpellDurationCancel4FE9D0NativeDoesNotGuardNilRecord(t *testing.T) {
	var durations SpellsDuration
	defer func() {
		if recover() == nil {
			t.Fatal("nil duration record did not fault")
		}
	}()
	durations.SpellDurationCancel4FE9D0(nil)
}
