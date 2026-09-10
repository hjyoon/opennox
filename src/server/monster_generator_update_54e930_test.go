package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

func TestMonsterGeneratorUpdateNative54E930PreservesPointersAndLifecycle(t *testing.T) {
	definition := &MonsterDef{HealthQuest72: 73}
	waypoint := new(Waypoint)
	enemy := &Object{TypeInd: 91}
	templateUpdate := &MonsterUpdateData{
		MonsterDef:   definition,
		CurrentEnemy: enemy,
	}
	templateUpdate.Waypoints[0] = waypoint
	template := &Object{
		TypeInd:    77,
		ObjClass:   object.ClassMonster,
		Direction1: 23,
		UpdateData: unsafe.Pointer(templateUpdate),
	}

	generatorUpdate := &MonsterGenUpdateData{
		MaxActive: 1,
		Field92:   1,
		Field64:   0x11223344,
		FuncInd68: ^uint32(4),
	}
	generatorUpdate.Field0[0] = template
	generator := &Object{
		ObjClass:   object.ClassMonsterGenerator,
		ObjFlags:   object.FlagEnabled,
		PosVec:     types.Ptf(100, 200),
		UpdateData: unsafe.Pointer(generatorUpdate),
	}
	playerData := &PlayerUpdateData{Player: new(Player)}
	player := &Object{
		ObjClass:   object.ClassPlayer,
		PosVec:     generator.PosVec,
		UpdateData: unsafe.Pointer(playerData),
	}
	spawnedUpdate := new(MonsterUpdateData)
	spawned := &Object{
		TypeInd:    template.TypeInd,
		ObjClass:   object.ClassMonster,
		UpdateData: unsafe.Pointer(spawnedUpdate),
		HealthData: new(HealthData),
	}

	var (
		needSyncCount int
		created       *Object
		createdOwner  *Object
		createdAt     types.Pointf
		callback      *ScriptCallback
		caller        *Object
		trigger       *Object
		event         ScriptEventType
		fx            [4]int32
		fxLength      int16
		audioID       sound.ID
		audioObject   *Object
	)
	deps := monsterGeneratorNativeDeps54E930{
		questStage: func() uint32 { return 0 },
		questGroup: func() int { return 0 },
		balanceFloat: func(name string) float32 {
			if name == "QuestHardcoreStage" {
				return 100
			}
			return 0
		},
		floatToInt: func(value float32) int32 {
			return int32(math.RoundToEven(float64(value)))
		},
		frame: func() uint32 { return 8 },
		needSync: func(got *Object) {
			if got != generator {
				t.Fatalf("NeedSync object = %p, want generator %p", got, generator)
			}
			needSyncCount++
		},
		randomInt:       func(int, int) int { return 0 },
		randomFloat:     func(float32, float32) float64 { return 0 },
		firstPlayerUnit: func() *Object { return player },
		nextPlayerUnit:  func(*Object) *Object { return nil },
		trace:           func(types.Pointf, types.Pointf, MapTraceFlags) bool { return true },
		occupied:        func(types.Pointf) bool { return false },
		tileAllow:       func(*types.Pointf) int32 { return 0 },
		canPlace:        func(types.Pointf) bool { return true },
		newObject: func(typeInd int) *Object {
			if typeInd != int(template.TypeInd) {
				t.Fatalf("new object type = %d, want %d", typeInd, template.TypeInd)
			}
			return spawned
		},
		healthFactor: func() float64 { return 1.5 },
		typeHealth:   func(*Object) *HealthData { return nil },
		typeByName:   func(string) int { return 999 },
		register: func(gotGenerator, gotSpawned *Object) bool {
			return gotGenerator == generator && gotSpawned == spawned
		},
		freeObject: func(*Object) { t.Fatal("registered spawn was freed") },
		createAt: func(got, owner *Object, position types.Pointf) {
			created, createdOwner, createdAt = got, owner, position
		},
		scriptCallback: func(got *ScriptCallback, gotCaller, gotTrigger *Object, gotEvent ScriptEventType) {
			callback, caller, trigger, event = got, gotCaller, gotTrigger, gotEvent
		},
		spawnFX: func(got [4]int32, length int16) {
			fx, fxLength = got, length
		},
		audio: func(id sound.ID, got *Object) {
			audioID, audioObject = id, got
		},
	}

	state := new(monsterGeneratorState54E930)
	monsterGeneratorUpdateNative54E930(generator, state, deps)

	if needSyncCount != 1 || generatorUpdate.Frame88 != 8 {
		t.Fatalf("NeedSync/frame = %d/%d, want 1/8", needSyncCount, generatorUpdate.Frame88)
	}
	if spawnedUpdate.MonsterDef != definition || spawnedUpdate.Waypoints[0] != waypoint ||
		spawnedUpdate.CurrentEnemy != enemy {
		t.Fatalf("copied pointers = %p/%p/%p, want %p/%p/%p",
			spawnedUpdate.MonsterDef, spawnedUpdate.Waypoints[0], spawnedUpdate.CurrentEnemy,
			definition, waypoint, enemy,
		)
	}
	if spawned.Direction1 != template.Direction1 || spawned.Direction2 != template.Direction1 {
		t.Fatalf("spawned directions = %d/%d, want %d", spawned.Direction1, spawned.Direction2, template.Direction1)
	}
	if spawned.HealthData.Cur != 110 || spawned.HealthData.Max != 110 {
		t.Fatalf("spawned health = %d/%d, want 110/110", spawned.HealthData.Cur, spawned.HealthData.Max)
	}
	if created != spawned || createdOwner != nil || createdAt == generator.PosVec {
		t.Fatalf("create = %p/%p/%v, want spawned/nil/non-generator position", created, createdOwner, createdAt)
	}
	wantCallback := (*ScriptCallback)(unsafe.Pointer(&generatorUpdate.Field64))
	if callback != wantCallback || callback.Flags != generatorUpdate.Field64 || callback.Func != -5 ||
		caller != spawned || trigger != generator || event != NoxEventGeneratorSpawn {
		t.Fatalf("callback = %p %#v caller=%p trigger=%p event=%d", callback, callback, caller, trigger, event)
	}
	if fxLength != 10 || fx[0] != 100 || fx[1] != 150 {
		t.Fatalf("spawn FX = %v/%d, want origin 100/150 and length 10", fx, fxLength)
	}
	if audioID != sound.SoundMonsterGeneratorSpawn || audioObject != spawned {
		t.Fatalf("audio = %d/%p, want generator spawn/%p", audioID, audioObject, spawned)
	}
}

func TestMonsterGeneratorCopyCreature54F2B0ClonesAndEquipsInventory(t *testing.T) {
	templateDefinition := new(MonsterDef)
	templateUpdate := &MonsterUpdateData{MonsterDef: templateDefinition}
	weaponAttrs := new(ModifierInitData)
	weapon := &Object{
		TypeInd:  11,
		ObjClass: object.ClassWeapon,
		ObjFlags: object.FlagEquipped,
		InitData: unsafe.Pointer(weaponAttrs),
		InvNextItem: &Object{
			TypeInd:  12,
			ObjClass: object.ClassArmor,
			ObjFlags: object.FlagEquipped,
		},
	}
	template := &Object{
		ObjClass:     object.ClassMonster,
		ObjSubClass:  object.SubClass(0x10),
		Direction1:   17,
		InvFirstItem: weapon,
		UpdateData:   unsafe.Pointer(templateUpdate),
	}
	spawnedUpdate := new(MonsterUpdateData)
	spawned := &Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(spawnedUpdate)}
	weaponClone := &Object{TypeInd: weapon.TypeInd, ObjClass: object.ClassWeapon}
	armorClone := &Object{TypeInd: weapon.InvNextItem.TypeInd, ObjClass: object.ClassArmor}
	var (
		putItems       []*Object
		appliedSource  *ModifierInitData
		appliedTarget  *Object
		equippedWeapon *Object
		equippedArmor  *Object
	)
	deps := monsterGeneratorNativeDeps54E930{
		newObject: func(typeInd int) *Object {
			switch typeInd {
			case int(weapon.TypeInd):
				return weaponClone
			case int(weapon.InvNextItem.TypeInd):
				return armorClone
			default:
				t.Fatalf("unexpected inventory type %d", typeInd)
				return nil
			}
		},
		applyAttrs: func(item *Object, attrs *ModifierInitData) bool {
			appliedTarget, appliedSource = item, attrs
			return true
		},
		inventoryPut: func(owner, item *Object, _ int32) {
			if owner != spawned {
				t.Fatalf("inventory owner = %p, want %p", owner, spawned)
			}
			putItems = append(putItems, item)
			item.InvHolder = owner
			item.InvNextItem = owner.InvFirstItem
			owner.InvFirstItem = item
		},
		equipWeapon: func(owner, item *Object) int {
			if owner != spawned {
				t.Fatalf("weapon owner = %p, want %p", owner, spawned)
			}
			equippedWeapon = item
			return 1
		},
		equipArmor: func(owner, item *Object) int {
			if owner != spawned {
				t.Fatalf("armor owner = %p, want %p", owner, spawned)
			}
			equippedArmor = item
			return 1
		},
	}

	monsterGeneratorCopyCreature54F2B0(template, spawned, deps)

	if spawnedUpdate.MonsterDef != templateDefinition {
		t.Fatalf("MonsterDef pointer = %p, want %p", spawnedUpdate.MonsterDef, templateDefinition)
	}
	if len(putItems) != 2 || putItems[0] != weaponClone || putItems[1] != armorClone {
		t.Fatalf("cloned inventory = %v, want weapon then armor", putItems)
	}
	if appliedTarget != weaponClone || appliedSource != weaponAttrs {
		t.Fatalf("applied attrs = %p/%p, want %p/%p", appliedTarget, appliedSource, weaponClone, weaponAttrs)
	}
	if equippedWeapon != weaponClone || equippedArmor != armorClone {
		t.Fatalf("equipped clones = %p/%p, want %p/%p", equippedWeapon, equippedArmor, weaponClone, armorClone)
	}
	if spawned.Direction1 != template.Direction1 || spawned.Direction2 != template.Direction1 {
		t.Fatalf("spawned directions = %d/%d, want %d", spawned.Direction1, spawned.Direction2, template.Direction1)
	}
}
