package legacy

import (
	"bytes"
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"unsafe"

	objectlib "github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type abilityRewardXferNativeInventoryCall4F6240 struct {
	version uint16
	object  *server.Object
	count   int32
}

func TestAbilityRewardXferNativeLayout4F6240(t *testing.T) {
	type field struct {
		name string
		got  uintptr
		pe32 uintptr
		wide uintptr
	}
	fields := []field{
		{name: "Object size", got: unsafe.Sizeof(server.Object{}), pe32: 780, wide: 928},
		{name: "Object.Field34", got: unsafe.Offsetof(server.Object{}.Field34), pe32: 136, wide: 140},
		{name: "Object.UseData", got: unsafe.Offsetof(server.Object{}.UseData), pe32: 736, wide: 848},
	}
	wide := unsafe.Sizeof(uintptr(0)) == 8
	for _, field := range fields {
		want := field.pe32
		if wide {
			want = field.wide
		}
		if field.got != want {
			t.Errorf("%s native layout = %d, want %d", field.name, field.got, want)
		}
	}
	if got := unsafe.Sizeof(server.AbilityRewardUseData{}); got != 1 {
		t.Errorf("AbilityReward use-data size = %d, want 1", got)
	}
	if got := unsafe.Offsetof(server.AbilityRewardUseData{}.Ability); got != 0 {
		t.Errorf("AbilityReward ability offset = %d, want 0", got)
	}
}

func TestAbilityRewardXferRuntimeNameTable4F6240(t *testing.T) {
	deps := abilityRewardXferRuntimeDeps4F6240()
	for id, name := range server.AbilityNames {
		if got := deps.abilityName(uint8(id)); got != name {
			t.Errorf("abilityName(%d) = %q, want %q", id, got, name)
		}
		if got := deps.abilityID(name); got != int32(id) {
			t.Errorf("abilityID(%q) = %d, want %d", name, got, id)
		}
	}
	if got := deps.abilityName(0xff); got != "Ability(255)" {
		t.Errorf("invalid ability name = %q, want Ability(255)", got)
	}
	if got := deps.abilityID("ABILITY_UNKNOWN"); got != int32(server.AbilityInvalid) {
		t.Errorf("unknown ability ID = %d, want invalid zero", got)
	}
}

func TestAbilityRewardXferNativeWrite4F6240PreservesPointersAndExactWire(t *testing.T) {
	object, freeObject := alloc.New(server.Object{})
	defer freeObject()
	data, freeData := alloc.New(server.AbilityRewardUseData{})
	defer freeData()
	id, freeID := alloc.CString("ability-reward")
	defer freeID()

	for name, pointer := range map[string]unsafe.Pointer{
		"object":                 unsafe.Pointer(object),
		"AbilityReward use data": unsafe.Pointer(data),
		"ID":                     unsafe.Pointer(id),
	} {
		assertObjectMapNativePointer4F4530(t, name, pointer)
	}

	const (
		extent       = uint32(0x11223344)
		scriptID     = int32(-0x1020304)
		positionX    = float32(123.25)
		positionY    = float32(-456.5)
		flags        = uint32(0x91408162)
		status       = uint32(0xa5)
		handlerFlags = uint32(0xa1b2c3d4)
		objectFrame  = uint32(0x11223344)
		gameFrame    = uint32(0x01020304)
	)
	data.Ability = uint8(server.AbilityHarpoon)
	object.ObjClass = objectlib.ClassPlayer | objectlib.ClassMissile
	object.Extent = extent
	object.ScriptIDVal = scriptID
	object.PosVec.X = positionX
	object.PosVec.Y = positionY
	object.ObjFlags = objectlib.Flags(flags)
	object.IDPtr = unsafe.Pointer(id)
	object.TeamVal.ID = server.TeamID(7)
	object.Field5 = status
	object.ScriptPickup = server.ScriptCallback{Flags: handlerFlags, Func: -1}
	object.Field34 = objectFrame
	object.UseData.Ptr = unsafe.Pointer(data)

	path := filepath.Join(t.TempDir(), "ability-reward-write.bin")
	cf, err := cryptfile.OpenFile(path, cryptfile.WriteOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	setObjectMapRuntimeGlobals4F4530(t, cf, gameFrame)
	if got := Nox_xxx_XFerAbilityRewardNative4F6240(cf, object); got != 1 {
		_ = cf.Close()
		t.Fatalf("result = %d, want 1", got)
	}
	if err := cf.Close(); err != nil {
		t.Fatal(err)
	}

	var want bytes.Buffer
	writeU16 := func(value uint16) { _ = binary.Write(&want, binary.LittleEndian, value) }
	writeU32 := func(value uint32) { _ = binary.Write(&want, binary.LittleEndian, value) }
	writeI32 := func(value int32) { writeU32(uint32(value)) }
	writeU16(abilityRewardXferCurrentVersion4F6240)
	writeU16(objectMapCurrentVersion4F4530)
	writeU32(extent)
	writeI32(scriptID)
	writeU32(math.Float32bits(positionX))
	writeU32(math.Float32bits(positionY))
	want.WriteByte(math.MaxUint8)
	writeU32(flags & objectMapFlagsMask4F4530)
	want.WriteByte(uint8(len("ability-reward")))
	want.WriteString("ability-reward")
	want.WriteByte(7)
	want.WriteByte(0)
	writeU16(0)
	writeU32(status & objectMapStatusMask4F4530)
	writeU16(1)
	writeU32(0)
	writeU32(handlerFlags)
	writeI32(int32(objectFrame - gameFrame))
	abilityName := server.AbilityHarpoon.String()
	want.WriteByte(uint8(len(abilityName)))
	want.WriteString(abilityName)

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want.Bytes()) {
		t.Fatalf("wire payload = %x, want %x", got, want.Bytes())
	}
	if object.Field34 != objectFrame || object.UseData.Ptr != unsafe.Pointer(data) {
		t.Fatalf("Field34/UseData = %#x/%p, want %#x/%p",
			object.Field34, object.UseData.Ptr, objectFrame, data)
	}
	if object.IDPtr != unsafe.Pointer(id) || data.Ability != uint8(server.AbilityHarpoon) {
		t.Fatalf("ID/ability changed: %p/%d", object.IDPtr, data.Ability)
	}
}

func TestAbilityRewardXferNativeRead4F6240UsesCachedRecordAndRestoresCount(t *testing.T) {
	const (
		extent         = uint32(0x55667788)
		scriptID       = int32(0x10203040)
		positionX      = float32(-321.25)
		positionY      = float32(654.5)
		serialized     = uint32(0x01400102)
		status         = uint32(0x12)
		handlerFlags   = uint32(0x55667788)
		frameDelta     = int32(0x01020304)
		originalFlags  = uint32(0x80000040)
		originalState  = uint32(0xa5)
		originalCount  = uint32(0xfedcba98)
		inventoryCount = uint8(3)
	)

	var payload bytes.Buffer
	writeU16 := func(value uint16) { _ = binary.Write(&payload, binary.LittleEndian, value) }
	writeU32 := func(value uint32) { _ = binary.Write(&payload, binary.LittleEndian, value) }
	writeI32 := func(value int32) { writeU32(uint32(value)) }
	writeU16(abilityRewardXferCurrentVersion4F6240)
	writeU16(objectMapCurrentVersion4F4530)
	writeU32(extent)
	writeI32(scriptID)
	writeU32(math.Float32bits(positionX))
	writeU32(math.Float32bits(positionY))
	payload.WriteByte(math.MaxUint8)
	writeU32(serialized)
	payload.WriteByte(0)
	payload.WriteByte(9)
	payload.WriteByte(inventoryCount)
	writeU16(0)
	writeU32(status)
	writeU16(1)
	writeU32(0)
	writeU32(handlerFlags)
	writeI32(frameDelta)
	abilityName := server.AbilityHarpoon.String()
	payload.WriteByte(uint8(len(abilityName)))
	payload.WriteString(abilityName)

	path := filepath.Join(t.TempDir(), "ability-reward-read.bin")
	if err := os.WriteFile(path, payload.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	cf, err := cryptfile.OpenFile(path, cryptfile.ReadOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cf.Close() }()
	setObjectMapRuntimeGlobals4F4530(t, cf, 0x89abcdef)

	object, freeObject := alloc.New(server.Object{})
	defer freeObject()
	data, freeData := alloc.New(server.AbilityRewardUseData{})
	defer freeData()
	liveData, freeLiveData := alloc.New(server.AbilityRewardUseData{})
	defer freeLiveData()
	oldID, freeOldID := alloc.CString("")
	defer freeOldID()
	data.Ability = uint8(server.AbilityBerserk)
	liveData.Ability = uint8(server.AbilityTreadLightly)
	object.ObjClass = objectlib.ClassPlayer | objectlib.ClassMissile
	object.IDPtr = unsafe.Pointer(oldID)
	object.ObjFlags = objectlib.Flags(originalFlags)
	object.Field5 = originalState
	object.Field34 = originalCount
	object.ScriptPickup.Func = -1
	object.UseData.Ptr = unsafe.Pointer(data)

	for name, pointer := range map[string]unsafe.Pointer{
		"object":                    unsafe.Pointer(object),
		"cached AbilityReward data": unsafe.Pointer(data),
		"live AbilityReward data":   unsafe.Pointer(liveData),
		"zero-length existing ID":   unsafe.Pointer(oldID),
	} {
		assertObjectMapNativePointer4F4530(t, name, pointer)
	}

	var inventoryCalls []abilityRewardXferNativeInventoryCall4F6240
	deps := abilityRewardXferRuntimeDeps4F6240()
	deps.transferInventory = func(version uint16, gotObject *server.Object, count int32) int32 {
		inventoryCalls = append(inventoryCalls, abilityRewardXferNativeInventoryCall4F6240{
			version: version,
			object:  gotObject,
			count:   count,
		})
		gotObject.UseData.Ptr = unsafe.Pointer(liveData)
		gotObject.Field34 = 0x11223344
		return 1
	}
	if got := abilityRewardXferNative4F6240(cf, object, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}

	if len(inventoryCalls) != 1 || inventoryCalls[0] != (abilityRewardXferNativeInventoryCall4F6240{
		version: abilityRewardXferCurrentVersion4F6240,
		object:  object,
		count:   int32(inventoryCount),
	}) {
		t.Fatalf("inventory calls = %+v", inventoryCalls)
	}
	if data.Ability != uint8(server.AbilityHarpoon) {
		t.Fatalf("cached ability = %d, want %d", data.Ability, server.AbilityHarpoon)
	}
	if liveData.Ability != uint8(server.AbilityTreadLightly) || object.UseData.Ptr != unsafe.Pointer(liveData) {
		t.Fatalf("live AbilityReward record/pointer = %d/%p, want %d/%p",
			liveData.Ability, object.UseData.Ptr, server.AbilityTreadLightly, liveData)
	}
	if object.Field34 != originalCount {
		t.Fatalf("Field34 = %#x, want entry %#x", object.Field34, originalCount)
	}
	if object.IDPtr != unsafe.Pointer(oldID) {
		t.Fatalf("zero-length ID changed native pointer: got %p want %p", object.IDPtr, oldID)
	}
	if object.Extent != extent || object.ScriptIDVal != scriptID {
		t.Errorf("extent/script ID = %#08x/%#08x, want %#08x/%#08x",
			object.Extent, uint32(object.ScriptIDVal), extent, uint32(scriptID))
	}
	if object.PosVec.X != positionX || object.PosVec.Y != positionY || object.NewPos != object.PosVec {
		t.Errorf("position/new position = %v/%v, want (%v,%v) mirrored",
			object.PosVec, object.NewPos, positionX, positionY)
	}
}

func TestAbilityRewardXferExport4F6240PreservesNativePointerAndResult(t *testing.T) {
	object, freeObject := alloc.New(server.Object{})
	defer freeObject()
	assertObjectMapNativePointer4F4530(t, "object", unsafe.Pointer(object))

	old := abilityRewardXferCall4F6240
	t.Cleanup(func() { abilityRewardXferCall4F6240 = old })
	calls := 0
	abilityRewardXferCall4F6240 = func(_ *cryptfile.CryptFile, gotObject *server.Object) int32 {
		calls++
		switch calls {
		case 1:
			if gotObject != object {
				t.Fatalf("object = %p, want %p", gotObject, object)
			}
			return math.MinInt32
		case 2:
			if gotObject != nil {
				t.Fatalf("object = %p, want nil", gotObject)
			}
			return math.MaxInt32
		default:
			t.Fatalf("unexpected call %d", calls)
			return 0
		}
	}

	if got := abilityRewardXferExportCall4F6240(object); got != math.MinInt32 {
		t.Fatalf("first result = %d, want %d", got, int32(math.MinInt32))
	}
	if got := abilityRewardXferExportCall4F6240(nil); got != math.MaxInt32 {
		t.Fatalf("second result = %d, want %d", got, int32(math.MaxInt32))
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
	runtime.KeepAlive(object)
}
