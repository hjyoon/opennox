package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
)

func TestPlayerSubStaminaNative4F7D30Layout(t *testing.T) {
	wantObjectClass := uintptr(8)
	wantObjectUpdate := uintptr(748)
	wantPlayer := uintptr(276)
	wantPlayerIndex := uintptr(2064)
	wantMonsterStamina := uintptr(1128)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectClass = 12
		wantObjectUpdate = 872
		wantPlayer = 336
		wantPlayerIndex = 2068
		wantMonsterStamina = 1792
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
		{"MonsterUpdateData.Stamina", unsafe.Offsetof(MonsterUpdateData{}.Stamina), wantMonsterStamina},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestPlayerSubStaminaNative4F7D30PreservesHighPointersAndStoreOrder(t *testing.T) {
	player := &Player{PlayerInd: 0xfe}
	update := &PlayerUpdateData{Stamina: 100, Player: player}
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
	got := playerSubStaminaNative4F7D30(unit, 45, playerSubStaminaNativeDeps4F7D30{
		reportStamina: func(index uint8, gotUnit *Object) {
			reports++
			if index != 0xfe || gotUnit != unit || update.Stamina != 55 {
				t.Fatalf("report = index:%d unit:%p stamina:%d", index, gotUnit, update.Stamina)
			}
		},
	})
	if got != 1 || update.Stamina != 55 || reports != 1 {
		t.Fatalf("result = %d stamina=%d reports=%d, want 1/55/1", got, update.Stamina, reports)
	}
	runtime.KeepAlive(player)
	runtime.KeepAlive(update)
	runtime.KeepAlive(unit)
}

func TestPlayerSubStaminaNative4F7D30MonsterNegativeAmountWraps(t *testing.T) {
	update := &MonsterUpdateData{Stamina: 5}
	unit := &Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(update)}
	got := playerSubStaminaNative4F7D30(unit, -1, playerSubStaminaNativeDeps4F7D30{
		reportStamina: func(uint8, *Object) {
			t.Fatal("monster stamina reported")
		},
	})
	if got != 1 || update.Stamina != 6 {
		t.Fatalf("result = %d stamina=%d, want 1/6", got, update.Stamina)
	}
}

func TestPlayerSubStaminaReportNative4D8800LiveClassUpdateAndPacket(t *testing.T) {
	first := &PlayerUpdateData{Stamina: 1}
	second := &PlayerUpdateData{Stamina: 0xab}
	unit := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(first)}
	unit.UpdateData = unsafe.Pointer(second)

	called := 0
	playerSubStaminaReportNative4D8800(0xfe, unit, playerSubStaminaReportNativeDeps4D8800{
		sendReliable: func(recipient int32, packet [2]byte, related *Object, remove int32) int32 {
			called++
			if recipient != 0xfe || packet != [2]byte{byte(netmsg.MSG_REPORT_STAMINA), 0xab} || related != nil || remove != 1 {
				t.Fatalf("send = recipient:%d packet:%v related:%p remove:%d", recipient, packet, related, remove)
			}
			return -1 << 31
		},
	})
	if called != 1 {
		t.Fatalf("send calls = %d, want 1", called)
	}

	unit.ObjClass = object.ClassMonster
	unit.UpdateData = nil
	playerSubStaminaReportNative4D8800(3, unit, playerSubStaminaReportNativeDeps4D8800{
		sendReliable: func(int32, [2]byte, *Object, int32) int32 {
			t.Fatal("non-player report sent")
			return 0
		},
	})
}
