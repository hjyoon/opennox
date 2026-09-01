package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"
)

func defaultSpellRewardUseNativeDeps53F9E0() spellRewardUseNativeDeps53F9E0 {
	return spellRewardUseNativeDeps53F9E0{
		checkSpellClass:   func(uint8, uint8) int32 { return 0 },
		primaryMessage:    func(*Object, string, uint8) {},
		audit:             func(int32, *Object, int32, uint32) {},
		gameFlagsCheck:    func(uint32) int32 { return 0 },
		grantSpell:        func(*Object, int32, int32, int32, int32) int32 { return 0 },
		delayedDeleteItem: func(*Object) {},
	}
}

func TestUseSpellRewardNative53F9E0PreservesPointersAndScalars(t *testing.T) {
	player := &Player{}
	player.info[66] = spellRewardUseWarrior53F9E0
	update := &PlayerUpdateData{Player: player}
	owner := &Object{ObjClass: 0xf4, NetCode: 0x89abcdef, UpdateData: unsafe.Pointer(update)}
	data := &SpellRewardUseData{Spell: 5}
	item := &Object{}
	item.UseData.SetPtr(unsafe.Pointer(data))

	var (
		checkedClass uint8
		checkedSpell uint8
		grantOwner   *Object
		grantArgs    [4]int32
		deletedItem  *Object
	)
	deps := defaultSpellRewardUseNativeDeps53F9E0()
	deps.checkSpellClass = func(class, spellID uint8) int32 {
		checkedClass, checkedSpell = class, spellID
		return 0
	}
	deps.gameFlagsCheck = func(mask uint32) int32 {
		if mask != spellRewardUseQuestMask53F9E0 {
			t.Fatalf("flag mask = %#x, want %#x", mask, spellRewardUseQuestMask53F9E0)
		}
		return math.MinInt32
	}
	deps.grantSpell = func(gotOwner *Object, spellID, notify, quest, override int32) int32 {
		grantOwner = gotOwner
		grantArgs = [4]int32{spellID, notify, quest, override}
		return math.MinInt32
	}
	deps.delayedDeleteItem = func(gotItem *Object) {
		deletedItem = gotItem
	}
	deps.audit = func(int32, *Object, int32, uint32) {
		t.Fatal("successful reward emitted failure sound")
	}

	if got := useSpellRewardNative53F9E0(owner, item, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if checkedClass != spellRewardUseWarrior53F9E0 || checkedSpell != 5 {
		t.Fatalf("class check = %d/%d, want 1/5", checkedClass, checkedSpell)
	}
	if grantOwner != owner || grantArgs != [4]int32{5, 1, 1, 0} {
		t.Fatalf("grant = %p/%v, want %p/[5 1 1 0]", grantOwner, grantArgs, owner)
	}
	if deletedItem != item {
		t.Fatalf("deleted item = %p, want %p", deletedItem, item)
	}

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, ptr := range map[string]uintptr{
			"player": uintptr(unsafe.Pointer(player)),
			"update": uintptr(unsafe.Pointer(update)),
			"owner":  uintptr(unsafe.Pointer(owner)),
			"data":   uintptr(unsafe.Pointer(data)),
			"item":   uintptr(unsafe.Pointer(item)),
		} {
			if ptr <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, ptr)
			}
		}
	}
	runtime.KeepAlive(player)
	runtime.KeepAlive(update)
	runtime.KeepAlive(owner)
	runtime.KeepAlive(data)
	runtime.KeepAlive(item)
}

func TestUseSpellRewardNative53F9E0ClassFailureUsesLiveOwner(t *testing.T) {
	player := &Player{}
	player.info[66] = 0xff
	update := &PlayerUpdateData{Player: player}
	owner := &Object{ObjClass: 4, NetCode: 0x10203040, UpdateData: unsafe.Pointer(update)}
	data := &SpellRewardUseData{Spell: 2}
	item := &Object{}
	item.UseData.SetPtr(unsafe.Pointer(data))

	deps := defaultSpellRewardUseNativeDeps53F9E0()
	deps.primaryMessage = func(gotOwner *Object, message string, value uint8) {
		if gotOwner != owner || message != spellRewardUseFailureMsg53F9E0 || value != 0 {
			t.Fatalf("message = %p/%q/%d", gotOwner, message, value)
		}
		owner.NetCode = math.MaxUint32
	}
	audited := false
	deps.audit = func(id int32, gotOwner *Object, kind int32, code uint32) {
		audited = true
		if id != spellRewardUseFailureSound53F9E0 || gotOwner != owner || kind != spellRewardUseFailureKind53F9E0 || code != math.MaxUint32 {
			t.Fatalf("audit = %d/%p/%d/%#x", id, gotOwner, kind, code)
		}
	}
	deps.checkSpellClass = func(uint8, uint8) int32 {
		t.Fatal("invalid class checked spell class")
		return 0
	}

	if got := useSpellRewardNative53F9E0(owner, item, deps); got != 0 || !audited {
		t.Fatalf("result/audited = %d/%t, want 0/true", got, audited)
	}
	runtime.KeepAlive(player)
	runtime.KeepAlive(update)
	runtime.KeepAlive(owner)
	runtime.KeepAlive(data)
	runtime.KeepAlive(item)
}

func TestSpellRewardUseNativeLayout53F9E0(t *testing.T) {
	if got := unsafe.Sizeof(SpellRewardUseData{}); got != 1 {
		t.Fatalf("SpellRewardUseData size = %d, want 1", got)
	}
	if got := unsafe.Offsetof(SpellRewardUseData{}.Spell); got != 0 {
		t.Fatalf("Spell offset = %d, want 0", got)
	}
	if got := unsafe.Sizeof(Player{}.SpellLvl[0]); got != 4 {
		t.Fatalf("SpellLvl element size = %d, want 4", got)
	}
	if got := len(Player{}.SpellLvl); got != 137 {
		t.Fatalf("SpellLvl count = %d, want 137", got)
	}
}
