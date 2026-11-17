package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

// Original owner corpus register-glyph-0325 calls InventoryPut with a null
// Glyph after creation fails, and a duplicate registration calls neither.
func TestMonsterSpawnOriginalFailedGlyphNative50E030(t *testing.T) {
	s := new(Server)
	defer s.MonsterSpawnFree50D820()
	s.MonsterSpawnInit50D780()
	generator, count := monsterSpawnTestGenerator50D780(0)
	spawned, update := monsterSpawnTestMonster50D780(object.SubClass(object.MonsterBomber))
	var trace []string
	rt := MonsterSpawnRegisterRuntime50E030{
		CreateGlyph: func() *Object { trace = append(trace, "new_glyph"); return nil },
		InventoryPut: func(owner, item *Object, report int32) {
			if owner != spawned || item != nil || report != 1 {
				t.Fatal("original null Glyph inventory arguments")
			}
			trace = append(trace, "inventory_put:null:1")
		},
	}
	if !s.MonsterSpawnRegister50E030(generator, spawned, rt) || !s.MonsterSpawnRegister50E030(generator, spawned, rt) {
		t.Fatal("registration/duplicate")
	}
	if !reflect.DeepEqual(trace, []string{"new_glyph", "inventory_put:null:1"}) || count.ActiveCount != 1 || update.Field548 != generator || update.Field549 == nil {
		t.Fatal("original failed Glyph/duplicate contract", trace, count.ActiveCount)
	}
}

// Original delete-inventory-0353 looks up Glyph before scanning, caches each
// next item before deletion, removes both matches and skips duplicate work.
func TestMonsterSpawnOriginalGlyphInventoryNative50E210(t *testing.T) {
	s := new(Server)
	defer s.MonsterSpawnFree50D820()
	s.MonsterSpawnInit50D780()
	generator, count := monsterSpawnTestGenerator50D780(254)
	spawned, update := monsterSpawnTestMonster50D780(object.SubClass(object.MonsterBomber))
	if !s.MonsterSpawnRegister50E030(generator, spawned, MonsterSpawnRegisterRuntime50E030{}) {
		t.Fatal("fixture registration")
	}
	count.ActiveCount = 254
	items := [3]*Object{{TypeInd: 77}, {TypeInd: 19}, {TypeInd: 77}}
	spawned.InvFirstItem = items[0]
	items[0].InvNextItem = items[1]
	items[1].InvNextItem = items[2]
	lookups := 0
	var deleted []*Object
	rt := MonsterSpawnDeleteRuntime50E210{
		GlyphID: func() int {
			if len(deleted) != 0 {
				t.Fatal("Glyph lookup occurred after deletion")
			}
			lookups++
			return 77
		},
		DelayedDelete: func(obj *Object) { deleted = append(deleted, obj); obj.InvNextItem = nil },
	}
	s.MonsterSpawnDelete50E210(spawned, rt)
	s.MonsterSpawnDelete50E210(spawned, rt)
	if lookups != 1 || !reflect.DeepEqual(deleted, []*Object{items[0], items[2]}) || count.ActiveCount != 253 || update.Field548 != nil || update.Field549 != nil {
		t.Fatal("original cached Glyph deletion contract", lookups, deleted, count.ActiveCount)
	}
}

func TestMonsterSpawnOriginalGlyphLookupBeforeClassFilter(t *testing.T) {
	s := new(Server)
	defer s.MonsterSpawnFree50D820()
	obj := &Object{UpdateData: unsafe.Pointer(new(MonsterUpdateData)), PosVec: types.Ptf(1, 2)}
	lookups := 0
	rt := MonsterSpawnDeleteRuntime50E210{GlyphID: func() int { lookups++; return 77 }}
	s.MonsterSpawnDelete50E210(obj, rt)
	s.MonsterSpawnDelete50E210(obj, rt)
	if lookups != 1 {
		t.Fatal("original eager Glyph cache for non-monster", lookups)
	}
}
