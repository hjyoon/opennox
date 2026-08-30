package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestPlayerAdjustStaminaNative4F7DB0Layout(t *testing.T) {
	wantObjectClass := uintptr(8)
	wantObjectUpdate := uintptr(748)
	wantPlayer := uintptr(276)
	wantPlayerIndex := uintptr(2064)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectClass = 12
		wantObjectUpdate = 872
		wantPlayer = 336
		wantPlayerIndex = 2068
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantObjectClass},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantObjectUpdate},
		{"PlayerUpdateData.Stamina", unsafe.Offsetof(PlayerUpdateData{}.Stamina), 91},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantPlayer},
		{"Player.PlayerInd", unsafe.Offsetof(Player{}.PlayerInd), wantPlayerIndex},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestPlayerAdjustStaminaNative4F7DB0PreservesHighPointersAndReportOrder(t *testing.T) {
	player := &Player{PlayerInd: 0xfe}
	update := &PlayerUpdateData{Stamina: 5, Player: player}
	unit := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}

	var pin runtime.Pinner
	pin.Pin(player)
	pin.Pin(update)
	pin.Pin(unit)
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]uintptr{
			"unit":   uintptr(unsafe.Pointer(unit)),
			"update": uintptr(unsafe.Pointer(update)),
			"player": uintptr(unsafe.Pointer(player)),
		} {
			if pointer <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, pointer)
			}
		}
	}

	reports := 0
	playerAdjustStaminaNative4F7DB0(unit, 0xff, playerAdjustStaminaNativeDeps4F7DB0{
		reportStamina: func(index uint8, gotUnit *Object) {
			reports++
			if index != 0xfe || gotUnit != unit || update.Stamina != 6 {
				t.Fatalf("report = index:%d unit:%p stamina:%d", index, gotUnit, update.Stamina)
			}
		},
	})
	if update.Stamina != 6 || reports != 1 {
		t.Fatalf("stamina=%d reports=%d, want 6/1", update.Stamina, reports)
	}
	runtime.KeepAlive(player)
	runtime.KeepAlive(update)
	runtime.KeepAlive(unit)
}

func TestPlayerAdjustStaminaNative4F7DB0NonPlayerDoesNotReadUpdate(t *testing.T) {
	unit := &Object{ObjClass: object.ClassMonster}
	playerAdjustStaminaNative4F7DB0(unit, 1, playerAdjustStaminaNativeDeps4F7DB0{
		reportStamina: func(uint8, *Object) {
			t.Fatal("non-player stamina reported")
		},
	})
}
