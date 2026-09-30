package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/balance"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestQuestHealthScale4E3DD0NativeFieldsAndScalarCache(t *testing.T) {
	srv := New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	srv.Types.Free()
	srv.Balance.file = &balance.File{Global: balance.Config{
		"generatormaxhealth": balance.Float(100), "playerdamagediffinit": balance.Float(1.5),
		"systemhealthdiffinit": balance.Float(1.5), "playerdamagediffcoeff": balance.Float(0.25),
		"systemhealthdiffcoeff": balance.Float(1),
	}}
	base := 3
	if err := srv.Types.ReadObjectType(&things.Thing{Name: "QuestHealthFixture", Health: &base}); err != nil {
		t.Fatal(err)
	}
	typ := srv.Types.ByID("QuestHealthFixture")
	typ.Health().Cur = 1
	gen, freeGen := alloc.New(Object{})
	monster, freeMonster := alloc.New(Object{})
	genHP, freeGenHP := alloc.New(HealthData{})
	monsterHP, freeMonsterHP := alloc.New(HealthData{})
	update, freeUpdate := alloc.New(MonsterUpdateData{})
	definition, freeDefinition := alloc.New(MonsterDef{})
	for _, free := range []func(){freeGen, freeMonster, freeGenHP, freeMonsterHP, freeUpdate, freeDefinition} {
		t.Cleanup(free)
	}
	*genHP, *monsterHP = HealthData{Cur: 3, Max: 3, Field2: 0x1234}, HealthData{Cur: 3, Max: 3, Field2: 0xabcd}
	*gen = Object{ObjClass: object.Class(0x20000), TypeInd: uint16(typ.Ind()), HealthData: genHP, ObjNext: monster}
	*monster = Object{ObjClass: object.ClassMonster, TypeInd: uint16(typ.Ind()), HealthData: monsterHP, UpdateData: unsafe.Pointer(update)}
	definition.HealthQuest72 = 0x12340003
	update.MonsterDef, update.StatusFlags = definition, object.MonsterStatus(0xffffff7f)
	srv.Objs.SetObjects(gen)
	t.Cleanup(func() { srv.Objs.SetObjects(nil) })
	if unsafe.Sizeof(uintptr(0)) > 4 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(gen), unsafe.Pointer(monster), unsafe.Pointer(genHP), unsafe.Pointer(monsterHP), unsafe.Pointer(update), unsafe.Pointer(definition), typ.Health().C()} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatalf("pointer=%p must exceed 4 GiB", ptr)
			}
		}
	}
	var ready uint32
	var damageInit, healthInit, damageCoeff, healthCoeff, damage, factor float32
	writes := 0
	r := QuestHealthScaleRuntime4E3DD0{
		Ready: &ready, DamageInit: &damageInit, HealthInit: &healthInit, DamageCoeff: &damageCoeff, HealthCoeff: &healthCoeff,
		Difficulty:       func() float64 { return 2 },
		StoreDamageScale: func(v float32) { damage = v }, StoreHealthScale: func(v float32) { factor = v },
		HealthScale: func() float64 { return float64(factor) },
		SetHP:       func(obj *Object, value uint16) int32 { writes++; obj.HealthData.Cur = value; return 0 },
	}
	wantReturn := uint16(uintptr(unsafe.Pointer(&update.HealthGraph103[0])) + unsafe.Sizeof(update.HealthGraph103))
	if got := srv.QuestHealthScale4E3DD0(r); uint16(got) != wantReturn {
		t.Fatalf("return=%04x want=%04x", uint16(got), wantReturn)
	}
	if ready != 1 || damageInit != 1.5 || healthInit != 1.5 || damageCoeff != 0.25 || healthCoeff != 1 || damage != 1.75 || factor != 2.5 || writes != 2 {
		t.Fatalf("cache=%d %v/%v/%v/%v scales=%v/%v writes=%d", ready, damageInit, healthInit, damageCoeff, healthCoeff, damage, factor, writes)
	}
	if genHP.Cur != 2 || genHP.Max != 8 || monsterHP.Cur != 8 || monsterHP.Max != 8 || genHP.Field2 != 0x1234 || monsterHP.Field2 != 0xabcd {
		t.Fatalf("native hp=%+v/%+v", genHP, monsterHP)
	}
	for i, value := range update.HealthGraph103 {
		if value != 8 {
			t.Fatalf("history[%d]=%d", i, value)
		}
	}
	// The generator is now injured and must not be reset on the next pass.
	monsterHP.Cur = 7
	srv.QuestHealthScale4E3DD0(r)
	if writes != 2 || genHP.Cur != 2 || monsterHP.Cur != 7 {
		t.Fatalf("injured actors rescaled: writes=%d", writes)
	}
}
