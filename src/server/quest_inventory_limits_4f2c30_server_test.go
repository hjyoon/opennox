package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/balance"
	"github.com/opennox/libs/object"
)

func questInventoryServerAddTypes4F2C30(srv *Server) {
	for index, name := range questInventoryTypeNames4F2C30 {
		questItemServerAddType4F2590(srv, name, 101+index)
	}
}

func questInventoryServerSetLimit4F2C30(srv *Server, value float64) {
	srv.Balance.file = &balance.File{
		Global: balance.Config{
			"forceofnaturestafflimit": balance.Float(value),
		},
		Tags: make(map[balance.Tag]balance.Config),
	}
}

func questInventoryServerAddItems4F2C30(owner *Object, typeInd uint16, count int) {
	for index := 0; index < count; index++ {
		owner.InvFirstItem = &Object{TypeInd: typeInd, InvNextItem: owner.InvFirstItem}
	}
}

func TestQuestInventoryRoundFloat32ToInt32_4F2C30(t *testing.T) {
	for _, test := range []struct {
		name  string
		value float32
		want  int32
	}{
		{name: "zero", value: 0},
		{name: "positive below half", value: 1.49, want: 1},
		{name: "positive tie even", value: 2.5, want: 2},
		{name: "positive tie odd", value: 3.5, want: 4},
		{name: "negative tie even", value: -2.5, want: -2},
		{name: "negative tie odd", value: -3.5, want: -4},
		{name: "largest finite result", value: math.Float32frombits(0x4effffff), want: 2147483520},
		{name: "positive boundary", value: math.Float32frombits(0x4f000000), want: math.MinInt32},
		{name: "negative boundary", value: math.Float32frombits(0xcf000000), want: math.MinInt32},
		{name: "negative out of range", value: math.Float32frombits(0xcf000001), want: math.MinInt32},
		{name: "positive infinity", value: float32(math.Inf(1)), want: math.MinInt32},
		{name: "negative infinity", value: float32(math.Inf(-1)), want: math.MinInt32},
		{name: "nan", value: float32(math.NaN()), want: math.MinInt32},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := questInventoryRoundFloat32ToInt32_4F2C30(test.value); got != test.want {
				t.Fatalf("round(%08x) = %d, want %d", math.Float32bits(test.value), got, test.want)
			}
		})
	}
}

func TestQuestInventoryServerNilAndNonPlayerBypass4F2C30(t *testing.T) {
	srv := new(Server)
	questInventoryServerAddTypes4F2C30(srv)
	if got := srv.QuestInventoryLimits4F2C30(nil); got != 1 {
		t.Fatalf("nil owner = %d, want 1", got)
	}
	for index, typeID := range srv.questInventoryLimits.typeIDs {
		if typeID != uint32(101+index) {
			t.Fatalf("cache[%d] = %d, want %d", index, typeID, 101+index)
		}
	}
	owner := &Object{ObjClass: object.ClassMonster}
	if got := srv.QuestInventoryLimits4F2C30(owner); got != 1 {
		t.Fatalf("non-player owner = %d, want 1", got)
	}
}

func TestQuestInventoryServerUsesNativeObjectLinks4F2C30(t *testing.T) {
	srv := new(Server)
	questInventoryServerAddTypes4F2C30(srv)
	questInventoryServerSetLimit4F2C30(srv, 3.5)
	owner := &Object{ObjClass: object.ClassPlayer}
	for index := 0; index < questInventoryPotionTypes4F2C30; index++ {
		questInventoryServerAddItems4F2C30(owner, uint16(101+index), 9)
	}
	questInventoryServerAddItems4F2C30(owner, uint16(101+questInventoryPotionTypes4F2C30), 4)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(owner)) <= math.MaxUint32 {
		t.Fatalf("owner does not exercise a native high pointer on %s/%s: %p", runtime.GOOS, runtime.GOARCH, owner)
	}
	if got := srv.QuestInventoryLimits4F2C30(owner); got != 1 {
		t.Fatalf("native inventory at rounded boundary = %d, want 1", got)
	}
	questInventoryServerAddItems4F2C30(owner, 106, 1)
	if got := srv.QuestInventoryLimits4F2C30(owner); got != 0 {
		t.Fatalf("ten native ShieldPotion objects = %d, want 0", got)
	}
}

func TestQuestInventoryServerStaffLimitRoundsToEven4F2C30(t *testing.T) {
	for _, test := range []struct {
		name  string
		limit float64
		count int
		want  int32
	}{
		{name: "three point five accepts four", limit: 3.5, count: 4, want: 1},
		{name: "two point five rejects three", limit: 2.5, count: 3},
		{name: "invalid rejects zero", limit: math.Inf(1), count: 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			srv := new(Server)
			questInventoryServerAddTypes4F2C30(srv)
			questInventoryServerSetLimit4F2C30(srv, test.limit)
			owner := &Object{ObjClass: object.ClassPlayer}
			questInventoryServerAddItems4F2C30(owner, uint16(101+questInventoryPotionTypes4F2C30), test.count)
			if got := srv.QuestInventoryLimits4F2C30(owner); got != test.want {
				t.Fatalf("limit %v count %d = %d, want %d", test.limit, test.count, got, test.want)
			}
		})
	}
}
