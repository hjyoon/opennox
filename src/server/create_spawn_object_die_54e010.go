package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// CreateSpawnObjectDeathData54E010 is the shared 132-byte parser result used
// by CreateObjectDie and SpawnObjectDie. GAME.EXE stores the spawned object's
// type name at byte 0 and the optional sound ID at byte 128.
type CreateSpawnObjectDeathData54E010 struct {
	TypeID [128]byte
	Sound  uint32
}

// CreateSpawnObjectDeathRuntime54E010 supplies the object factory and effects
// reached by the two death callbacks.
type CreateSpawnObjectDeathRuntime54E010 struct {
	NewObjectByTypeID func(string) *Object
	CreateAt          func(*Object, types.Pointf)
	Audio             func(uint32, *Object)
	DelayedDelete     func(*Object)
}

func createSpawnObjectDeath54E010(source *Object, runtime CreateSpawnObjectDeathRuntime54E010) {
	// Cache the data pointer before the factory call, as GAME.EXE does. The
	// sound remains a live load from that cached record after object creation.
	data := (*CreateSpawnObjectDeathData54E010)(source.DeathData)
	created := runtime.NewObjectByTypeID(alloc.GoString(&data.TypeID[0]))
	if created != nil {
		runtime.CreateAt(created, source.PosVec)
	}
	if data.Sound != 0 {
		runtime.Audio(data.Sound, source)
	}
}

// CreateObjectDieNative54E010 restores GAME.EXE 0054E010 without narrowing
// the source object or its shared death-data pointer to an ABI32 integer.
func CreateObjectDieNative54E010(source *Object, runtime CreateSpawnObjectDeathRuntime54E010) {
	createSpawnObjectDeath54E010(source, runtime)
	runtime.DelayedDelete(source)
}

// SpawnObjectDieNative54E070 restores GAME.EXE 0054E070. Despite the callback
// name, its final side effect is setting DEAD (0x8000), not DESTROYED (0x20).
func SpawnObjectDieNative54E070(source *Object, runtime CreateSpawnObjectDeathRuntime54E010) int16 {
	createSpawnObjectDeath54E010(source, runtime)
	source.ObjFlags |= object.FlagDead
	return int16(uint16(source.ObjFlags))
}
