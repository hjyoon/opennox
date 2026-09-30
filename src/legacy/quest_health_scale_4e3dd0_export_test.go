package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestQuestHealthScale4E3DD0CEntryDelegatesWholeBody(t *testing.T) {
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	oldServer, oldCall := GetServer, questHealthScaleCall4E3DD0
	GetServer = func() Server { return &questCanJoinLegacyServer4E4100{srv: srv} }
	t.Cleanup(func() { GetServer, questHealthScaleCall4E3DD0 = oldServer, oldCall })
	// Preserve globals touched by the old C body, so the red test is isolated.
	ready := memmap.PtrUint32(0x5D4594, 1563932)
	oldReady := *ready
	*ready = 1
	damage, health := memmap.PtrFloat32(0x587000, 202032), memmap.PtrFloat32(0x587000, 202036)
	oldDamage, oldHealth := *damage, *health
	t.Cleanup(func() { *ready, *damage, *health = oldReady, oldDamage, oldHealth })
	calls := 0
	questHealthScaleCall4E3DD0 = func() int16 { calls++; return -1234 }
	if got := questHealthScaleCEntry4E3DD0(); got != -1234 || calls != 1 {
		t.Fatalf("C entry=%d calls=%d, want -1234/1", got, calls)
	}
}

func TestQuestHealthScale4E3DD0CEntryNativeWorldAndRetainedSetter(t *testing.T) {
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	srv.Types.Free()
	oldServer, oldCall := GetServer, questHealthScaleCall4E3DD0
	GetServer = func() Server { return &questCanJoinLegacyServer4E4100{srv: srv} }
	t.Cleanup(func() { GetServer, questHealthScaleCall4E3DD0 = oldServer, oldCall })
	base := 3
	if err := srv.Types.ReadObjectType(&things.Thing{Name: "QuestHealthCFixture", Health: &base}); err != nil {
		t.Fatal(err)
	}
	typ := srv.Types.ByID("QuestHealthCFixture")
	gen, freeGen := alloc.New(server.Object{})
	monster, freeMonster := alloc.New(server.Object{})
	genHP, freeGenHP := alloc.New(server.HealthData{})
	monsterHP, freeMonsterHP := alloc.New(server.HealthData{})
	update, freeUpdate := alloc.New(server.MonsterUpdateData{})
	definition, freeDefinition := alloc.New(server.MonsterDef{})
	for _, free := range []func(){freeGen, freeMonster, freeGenHP, freeMonsterHP, freeUpdate, freeDefinition} {
		t.Cleanup(free)
	}
	*genHP, *monsterHP = server.HealthData{Cur: 3, Max: 3}, server.HealthData{Cur: 3, Max: 3}
	*gen = server.Object{ObjClass: object.Class(0x20000), TypeInd: uint16(typ.Ind()), HealthData: genHP, ObjNext: monster}
	*monster = server.Object{ObjClass: object.ClassMonster, TypeInd: uint16(typ.Ind()), HealthData: monsterHP, UpdateData: unsafe.Pointer(update)}
	definition.HealthQuest72, update.MonsterDef = 0x12340003, definition
	srv.Objs.SetObjects(gen)
	t.Cleanup(func() { srv.Objs.SetObjects(nil) })
	if unsafe.Sizeof(uintptr(0)) > 4 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(gen), unsafe.Pointer(monster), unsafe.Pointer(genHP), unsafe.Pointer(monsterHP), unsafe.Pointer(update), unsafe.Pointer(definition), typ.Health().C()} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatalf("native pointer=%p, want above 4 GiB", ptr)
			}
		}
	}
	ready, init := uint32(1), float32(1)
	coeff := float32(0)
	r := questHealthScaleRuntime4E3DD0()
	// Scalar seams isolate this C-entry test from the private balance assets.
	// With no balance file, GeneratorMaxHealth is zero; monsters are not capped.
	r.Ready, r.DamageInit, r.HealthInit, r.DamageCoeff, r.HealthCoeff = &ready, &init, &init, &coeff, &coeff
	r.Difficulty = func() float64 { return 2 }
	r.StoreDamageScale, r.StoreHealthScale = func(float32) {}, func(float32) {}
	r.HealthScale = func() float64 { return 2.5 }
	// r.SetHP still calls the actual typed C setter, including NeedSync.
	questHealthScaleCall4E3DD0 = func() int16 { return srv.QuestHealthScale4E3DD0(r) }
	wantReturn := uint16(uintptr(unsafe.Pointer(&update.HealthGraph103[0])) + unsafe.Sizeof(update.HealthGraph103))
	if got := questHealthScaleCEntry4E3DD0(); uint16(got) != wantReturn {
		t.Fatalf("short return=%04x want=%04x", uint16(got), wantReturn)
	}
	if genHP.Cur != 0 || genHP.Max != 0 || monsterHP.Cur != 8 || monsterHP.Max != 8 {
		t.Fatalf("C-set health=%+v/%+v", genHP, monsterHP)
	}
	if gen.Field38 != math.MaxUint32 || monster.Field38 != math.MaxUint32 {
		t.Fatal("retained C setter did not mark native objects for synchronization")
	}
	for i, value := range update.HealthGraph103 {
		if value != 8 {
			t.Fatalf("native history[%d]=%d", i, value)
		}
	}
	// Exercise the production adapter too: original scalar globals and all
	// retained C curve readers/writers, without any callback substitution.
	production := questHealthScaleRuntime4E3DD0()
	oldReady := *production.Ready
	globals := []*float32{
		production.DamageInit, production.HealthInit, production.DamageCoeff, production.HealthCoeff,
		memmap.PtrFloat32(0x587000, 202024), memmap.PtrFloat32(0x587000, 202032), memmap.PtrFloat32(0x587000, 202036),
	}
	saved := make([]float32, len(globals))
	for i, ptr := range globals {
		saved[i] = *ptr
		*ptr = 0
	}
	*production.Ready = 0
	t.Cleanup(func() {
		*production.Ready = oldReady
		for i, ptr := range globals {
			*ptr = saved[i]
		}
	})
	*genHP, *monsterHP = server.HealthData{Cur: 3, Max: 3}, server.HealthData{Cur: 3, Max: 3}
	questHealthScaleCall4E3DD0 = oldCall
	if got := questHealthScaleCEntry4E3DD0(); uint16(got) != wantReturn {
		t.Fatalf("production return=%04x", uint16(got))
	}
	if *production.Ready != 1 || genHP.Cur != 0 || genHP.Max != 0 || monsterHP.Cur != 1 || monsterHP.Max != 1 {
		t.Fatalf("production ready=%d hp=%+v/%+v", *production.Ready, genHP, monsterHP)
	}
	for i, value := range update.HealthGraph103 {
		if value != 1 {
			t.Fatalf("production history[%d]=%d", i, value)
		}
	}
}
