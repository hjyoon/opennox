package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestScriptObjectCachePreservesNativePointersAndLRUCapacity(t *testing.T) {
	const firstID = int32(0x3b9aca00)

	// A lookup initializes the historical lazy cache. Keep this test isolated
	// from any entries installed by another legacy test.
	_ = Nox_xxx_script_511C50(int(firstID))
	Sub_511E20()
	defer Sub_511E20()

	objects := make([]*server.Object, 17)
	for i := range objects {
		objects[i] = &server.Object{ScriptIDVal: firstID + int32(i)}
		Nox_xxx_scriptPrepareFoundUnit_511D70(objects[i])
	}

	if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(objects[16])) <= uintptr(^uint32(0)) {
		t.Fatalf("test object pointer %#x did not exercise the native 64-bit range", uintptr(unsafe.Pointer(objects[16])))
	}
	if got := Nox_xxx_script_511C50(int(firstID)); got != nil {
		t.Fatalf("oldest object survived the 16-entry LRU: got %p", got)
	}
	if got := Nox_xxx_script_511C50(int(firstID + 16)); got != objects[16] {
		t.Fatalf("newest object = %p, want native pointer %p", got, objects[16])
	}

	objects[16].ObjFlags = 0x20
	if got := Nox_xxx_script_511C50(int(firstID + 16)); got != nil {
		t.Fatalf("destroyed cached object matched: %p", got)
	}
	objects[16].ObjFlags = 0
	Sub_511DE0(objects[16])
	if got := Nox_xxx_script_511C50(int(firstID + 16)); got != nil {
		t.Fatalf("removed cached object matched: %p", got)
	}
}

func TestSpellPowerUsesNativeObjectUpdateLayouts(t *testing.T) {
	const spellID = 51
	player := &server.Player{}
	player.SpellLvl[spellID] = 7
	playerUpdate := &server.PlayerUpdateData{Player: player}
	playerObject := &server.Object{ObjClass: 4, UpdateData: unsafe.Pointer(playerUpdate)}
	if got := spellGetPowerNative4FE7B0(spellID, playerObject); got != 7 {
		t.Fatalf("player spell power = %d, want 7", got)
	}

	monsterUpdate := &server.MonsterUpdateData{Field510: 5}
	monsterObject := &server.Object{ObjClass: 2, UpdateData: unsafe.Pointer(monsterUpdate)}
	if got := spellGetPowerNative4FE7B0(spellID, monsterObject); got != 5 {
		t.Fatalf("monster spell power = %d, want 5", got)
	}
	if got := spellGetPowerNative4FE7B0(spellID, &server.Object{}); got != 3 {
		t.Fatalf("non-unit spell power = %d, want 3", got)
	}
	if got := spellGetPowerNative4FE7B0(spellID, nil); got != 2 {
		t.Fatalf("nil caster spell power = %d, want 2", got)
	}
}
