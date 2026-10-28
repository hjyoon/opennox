package opennox

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

func TestChainLightningNativeCRealWandAndRayAllocator(t *testing.T) {
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)
	s.spells.duration.Init(s)
	oldServer, oldGet := noxServer, legacy.GetServer
	noxServer, legacy.GetServer = s, func() legacy.Server { return s }
	t.Cleanup(func() { noxServer, legacy.GetServer = oldServer, oldGet })
	if !s.Spells.Dur.Init() {
		t.Fatal("native duration allocator")
	}
	t.Cleanup(s.Spells.Dur.Free)
	caster, freeCaster := alloc.New(server.Object{})
	wand, freeWand := alloc.New(server.Object{})
	target, freeTarget := alloc.New(server.Object{})
	update, freeUpdate := alloc.New(server.PlayerUpdateData{})
	data, freeData := alloc.New(server.WandUseData{})
	record, freeRecord := alloc.New(server.DurSpell{})
	for _, free := range []func(){freeCaster, freeWand, freeTarget, freeUpdate, freeData, freeRecord} {
		t.Cleanup(free)
	}
	*caster, *wand, *target = server.Object{ObjClass: object.ClassPlayer}, server.Object{}, server.Object{}
	*update = server.PlayerUpdateData{EquippedWeapon: wand}
	*data = server.WandUseData{Flags: 0x123456fd, Charge: 7, MaxCharge: 11}
	*record = server.DurSpell{Spell: 43, Level: 3, Caster16: caster, Pos: types.Ptf(11.5, -22.25), Flags88: 0x77556601}
	caster.UpdateData, wand.UseData.Ptr, wand.ObjSubClass = unsafe.Pointer(update), unsafe.Pointer(data), 0x40000
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, pointer := range []unsafe.Pointer{caster.CObj(), wand.CObj(), target.CObj(), unsafe.Pointer(update), unsafe.Pointer(data), record.C()} {
			if uintptr(pointer) <= 0xffffffff {
				t.Fatal("native composition pointer below4GiB")
			}
		}
	}
	if got := ccall.CallIntPtr(legacy.Get_nox_xxx_onStartLightning_52F820(), record.C()); got != 0 {
		t.Fatal("native create", got)
	}
	if s.spells.duration.chainLightningWeapons[record] != wand || record.Field72 != 0 || record.Flags88 != 0x77556603 {
		t.Fatal("real runtime lost native equipped wand or flags")
	}
	// The original buff8 guard returns before a balance read or mana debit.
	caster.Buffs = 1 << 8
	if got := ccall.CallIntPtr(legacy.Get_nox_xxx_onFrameLightning_52F8A0(), record.C()); got != 1 || data.Charge != 7 {
		t.Fatal("native guarded update", got, data.Charge)
	}
	for range 2 {
		s.Spells.Dur.NewLightningSub(record, caster, target)
	}
	record.Sub104, record.Sub108 = record.Sub108, nil
	for range 2 {
		s.Spells.Dur.NewLightningSub(record, target, caster)
	}
	rays := make(map[*server.DurSpell]bool)
	for _, head := range []*server.DurSpell{record.Sub104, record.Sub108} {
		for ray := head; ray != nil; ray = ray.Next {
			rays[ray] = true
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ray.C()) <= 0xffffffff {
				t.Fatal("native ray pointer below4GiB")
			}
		}
	}
	if len(rays) != 4 {
		t.Fatal("both real native ray generations were not allocated")
	}
	got := uint8(ccall.CallIntPtr(legacy.Get_sub_530100(), record.C()))
	if got != 0xf9 || data.Flags != 0x123456f9 || data.Charge != 7 || data.MaxCharge != 11 ||
		record.Sub104 != nil || record.Sub108 != nil || s.spells.duration.chainLightningWeapons[record] != nil {
		t.Fatalf("native destroy return/flags/rays/sidecar = %#x/%#x/%p/%p", got, data.Flags, record.Sub104, record.Sub108)
	}
	// FreeObjectFirst must return every released ray to the real allocator.
	for range 4 {
		ray := s.Spells.Dur.NewRaw()
		if !rays[ray] {
			t.Fatal("released native ray was not returned by the allocator")
		}
		delete(rays, ray)
	}
}
