package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestShopItemCostNoMerchant50E3D0BasicModifierAndHealth(t *testing.T) {
	s := &Server{}
	item := &Object{Worth: 76, ObjClass: object.ClassFood}
	if got := s.ShopItemCostNoMerchant50E3D0(item); got != 76 {
		t.Fatalf("basic cost = %d", got)
	}
	item.Worth = 0
	if got := s.ShopItemCostNoMerchant50E3D0(item); got != 1 {
		t.Fatalf("minimum cost = %d", got)
	}
	item.Worth = 1<<24 + 1
	if got := s.ShopItemCostNoMerchant50E3D0(item); got != 1<<24 {
		t.Fatalf("binary32 worth spill = %d", got)
	}
	mod := &ModifierEff{Price20: 31}
	attrs := &ModifierInitData{Modifiers: [4]*ModifierEff{mod}}
	item.ObjClass = object.ClassArmor
	item.Worth = 69
	item.InitData = unsafe.Pointer(attrs)
	item.HealthData = &HealthData{Cur: 1, Max: 2}
	if got := s.ShopItemCostNoMerchant50E3D0(item); got != 50 {
		t.Fatalf("modifier and health cost = %d", got)
	}
}

func TestShopItemCostNoMerchant50E3D0WandCharge(t *testing.T) {
	s := &Server{}
	item := &Object{Worth: 100, ObjClass: object.ClassWand, ObjSubClass: 0x10000}
	data := &WandUseData{Charge: 3, MaxCharge: 4}
	item.UseData.Ptr = unsafe.Pointer(data)
	item.InitData = unsafe.Pointer(&ModifierInitData{})
	if got := s.ShopItemCostNoMerchant50E3D0(item); got != 75 {
		t.Fatalf("wand charge cost = %d", got)
	}
}

func TestShopItemCostNoMerchant50E3D0MissingOptionalData(t *testing.T) {
	s := &Server{}
	for _, class := range []object.Class{object.ClassArmor, object.ClassWeapon, object.ClassWand} {
		item := &Object{Worth: 40, ObjClass: class, ObjSubClass: 0x10000}
		if got := s.ShopItemCostNoMerchant50E3D0(item); got != 40 {
			t.Fatalf("class %v without optional data: cost = %d", class, got)
		}
	}
}
