package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"
)

func defaultFieldGuideUseNativeDeps53F930() fieldGuideUseNativeDeps53F930 {
	return fieldGuideUseNativeDeps53F930{
		guideByName:       func(string) int32 { return 0 },
		gameFlagsCheck:    func(uint32) int32 { return 0 },
		primaryMessage:    func(*Object, string, uint8) {},
		awardGuide:        func(*Object, int32, int32) int32 { return 0 },
		delayedDeleteItem: func(*Object) {},
	}
}

func TestUseFieldGuideNative53F930PreservesPointersAndScalars(t *testing.T) {
	player := &Player{}
	player.info[66] = fieldGuideUseConjurerClass53F930
	update := &PlayerUpdateData{Player: player}
	owner := &Object{ObjClass: 0xf4, UpdateData: unsafe.Pointer(update)}
	data := new(FieldGuideUseData)
	data.SetCreature("CarnivorousPlant")
	item := new(Object)
	item.UseData.SetPtr(unsafe.Pointer(data))

	var (
		awardOwner  *Object
		awardArgs   [2]int32
		deletedItem *Object
	)
	deps := defaultFieldGuideUseNativeDeps53F930()
	deps.guideByName = func(name string) int32 {
		if name != "CarnivorousPlant" {
			t.Fatalf("guide name = %q", name)
		}
		return 24
	}
	deps.gameFlagsCheck = func(mask uint32) int32 {
		if mask != fieldGuideUseQuestFlag53F930 {
			t.Fatalf("flag mask = %#x, want %#x", mask, fieldGuideUseQuestFlag53F930)
		}
		return math.MinInt32
	}
	deps.awardGuide = func(gotOwner *Object, guide, notify int32) int32 {
		awardOwner = gotOwner
		awardArgs = [2]int32{guide, notify}
		return math.MinInt32
	}
	deps.delayedDeleteItem = func(gotItem *Object) {
		deletedItem = gotItem
	}
	deps.primaryMessage = func(*Object, string, uint8) {
		t.Fatal("successful use emitted a priority message")
	}

	if got := useFieldGuideNative53F930(owner, item, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if awardOwner != owner || awardArgs != [2]int32{24, 1} {
		t.Fatalf("award = %p/%v, want %p/[24 1]", awardOwner, awardArgs, owner)
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

func TestUseFieldGuideNative53F930UsesLivePlayerForDuplicateGate(t *testing.T) {
	initial := &Player{}
	initial.info[66] = fieldGuideUseConjurerClass53F930
	replacement := &Player{}
	replacement.BeastScrollLvl[1] = math.MaxUint32
	update := &PlayerUpdateData{Player: initial}
	owner := &Object{ObjClass: 4, UpdateData: unsafe.Pointer(update)}
	data := new(FieldGuideUseData)
	data.SetCreature("Bat")
	item := new(Object)
	item.UseData.SetPtr(unsafe.Pointer(data))

	deps := defaultFieldGuideUseNativeDeps53F930()
	deps.guideByName = func(string) int32 { return 1 }
	deps.gameFlagsCheck = func(uint32) int32 {
		update.Player = replacement
		return 0
	}
	messageSeen := false
	deps.primaryMessage = func(gotOwner *Object, message string, value uint8) {
		messageSeen = true
		if gotOwner != owner || message != fieldGuideUseDuplicateMessage53F930 || value != 0 {
			t.Fatalf("message = %p/%q/%d", gotOwner, message, value)
		}
	}
	deps.awardGuide = func(*Object, int32, int32) int32 {
		t.Fatal("duplicate guide was awarded")
		return 0
	}
	deps.delayedDeleteItem = func(*Object) {
		t.Fatal("duplicate guide item was deleted")
	}

	if got := useFieldGuideNative53F930(owner, item, deps); got != 0 || !messageSeen {
		t.Fatalf("result/message = %d/%t, want 0/true", got, messageSeen)
	}
	runtime.KeepAlive(initial)
	runtime.KeepAlive(replacement)
	runtime.KeepAlive(update)
	runtime.KeepAlive(owner)
	runtime.KeepAlive(data)
	runtime.KeepAlive(item)
}

func TestFieldGuideUse53F930NativeLayout(t *testing.T) {
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"FieldGuideUseData size", unsafe.Sizeof(FieldGuideUseData{}), 64},
		{"FieldGuideUseData.CreatureBuf", unsafe.Offsetof(FieldGuideUseData{}.CreatureBuf), 0},
		{"beast guide level width", unsafe.Sizeof(Player{}.BeastScrollLvl[0]), 4},
		{"beast guide level count", uintptr(len(Player{}.BeastScrollLvl)), 41},
		{"protection width", unsafe.Sizeof(Player{}.Prot4640), 4},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}
