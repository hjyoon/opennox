package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/things"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestItemArmorDamagePublicCEntries(t *testing.T) {
	s := npcReflectLegacyServer4E17B0(t)
	s.FreeObjectTypes()
	t.Cleanup(s.FreeObjectTypes)
	source, releaseSource := alloc.NewHigh(server.Object{})
	t.Cleanup(releaseSource)
	weapon, releaseWeapon := alloc.NewHigh(server.Object{})
	t.Cleanup(releaseWeapon)
	target, releaseTarget := alloc.NewHigh(server.Object{})
	t.Cleanup(releaseTarget)
	if cgoABIPointerSize == 8 {
		for _, ptr := range []*server.Object{target, source, weapon} {
			if uintptr(unsafe.Pointer(ptr)) <= 0xffffffff {
				t.Fatalf("fixture pointer %p must be above 4 GiB", ptr)
			}
		}
	}
	entries := itemDamagePublicEntries()
	for i, tc := range []struct {
		proc  string
		class object.Class
	}{
		{"BallDamage", 0},
		{"WeaponDamage", object.ClassWeapon},
		{"ArmorDamage", object.ClassArmor},
	} {
		t.Run(tc.proc, func(t *testing.T) {
			name := "NativePublic" + tc.proc
			thing := &things.Thing{Name: name, OnDamage: &things.ProcFunc{Name: tc.proc}}
			if err := s.Types.ReadObjectType(thing); err != nil {
				t.Fatal(err)
			}
			if got := s.Types.ByID(name).Damage; got != entries[i] {
				t.Fatalf("registered callback %p, want public C entry %p", got, entries[i])
			}
			*target = server.Object{ObjClass: tc.class, Material: uint16(object.MaterialMetal)}
			before := *target
			if got := callItemDamagePublicEntry(i, target, source, weapon); got != 0 || *target != before {
				t.Fatalf("public C callback returned %d or mutated an unheld target", got)
			}
			if got := callItemDamagePublicEntry(i, nil, source, weapon); got != 0 {
				t.Fatalf("public C callback returned %d for missing target", got)
			}
		})
	}
}
