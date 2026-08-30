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
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type teamXferNativeInventoryCall4F6D20 struct {
	version uint16
	object  *server.Object
	count   int32
}

func TestTeamXferNativeLayout4F6D20(t *testing.T) {
	type field struct {
		name string
		got  uintptr
		pe32 uintptr
		wide uintptr
	}
	fields := []field{
		{name: "Object size", got: unsafe.Sizeof(server.Object{}), pe32: 780, wide: 928},
		{name: "Object.ObjClass", got: unsafe.Offsetof(server.Object{}.ObjClass), pe32: 8, wide: 12},
		{name: "Object.PosVec", got: unsafe.Offsetof(server.Object{}.PosVec), pe32: 56, wide: 60},
		{name: "Object.Field34", got: unsafe.Offsetof(server.Object{}.Field34), pe32: 136, wide: 140},
		{name: "Object.InitData", got: unsafe.Offsetof(server.Object{}.InitData), pe32: 692, wide: 760},
		{name: "Object.UpdateData", got: unsafe.Offsetof(server.Object{}.UpdateData), pe32: 748, wide: 872},
		{name: "ModifierInitData size", got: unsafe.Sizeof(server.ModifierInitData{}), pe32: 20, wide: 40},
		{name: "ModifierInitData.Field16", got: unsafe.Offsetof(server.ModifierInitData{}.Field16), pe32: 16, wide: 32},
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

	fixed := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{name: "FlagUpdateData size", got: unsafe.Sizeof(server.FlagUpdateData4EA490{}), want: 12},
		{name: "FlagUpdateData.Home", got: unsafe.Offsetof(server.FlagUpdateData4EA490{}.Home), want: 0},
		{name: "FlagUpdateData.State", got: unsafe.Offsetof(server.FlagUpdateData4EA490{}.State), want: 8},
	}
	for _, field := range fixed {
		if field.got != field.want {
			t.Errorf("%s = %d, want %d", field.name, field.got, field.want)
		}
	}
}

func TestTeamXferNativeWrite4F6D20PreservesPointersAndExactWire(t *testing.T) {
	object, freeObject := alloc.New(server.Object{})
	defer freeObject()
	initData, freeInitData := alloc.New(server.ModifierInitData{})
	defer freeInitData()
	update, freeUpdate := alloc.New(server.FlagUpdateData4EA490{
		Home:  types.Pointf{X: 9, Y: 8},
		State: 7,
	})
	defer freeUpdate()
	*update = server.FlagUpdateData4EA490{Home: types.Pointf{X: 9, Y: 8}, State: 7}
	id, freeID := alloc.CString("team-native")
	defer freeID()

	for name, pointer := range map[string]unsafe.Pointer{
		"object":      unsafe.Pointer(object),
		"init data":   unsafe.Pointer(initData),
		"update data": unsafe.Pointer(update),
		"ID":          unsafe.Pointer(id),
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
	object.ObjClass = objectlib.ClassFlag | objectlib.ClassClientPersist
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
	object.InitData = unsafe.Pointer(initData)
	object.UpdateData = unsafe.Pointer(update)

	path := filepath.Join(t.TempDir(), "team-write.bin")
	cf, err := cryptfile.OpenFile(path, cryptfile.WriteOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	setObjectMapRuntimeGlobals4F4530(t, cf, gameFrame)
	if got := Nox_xxx_XFerTeamNative4F6D20(cf, object); got != 1 {
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
	writeU16(teamXferCurrentVersion4F6D20)
	writeU16(objectMapCurrentVersion4F4530)
	writeU32(extent)
	writeI32(scriptID)
	writeU32(math.Float32bits(positionX))
	writeU32(math.Float32bits(positionY))
	want.WriteByte(math.MaxUint8)
	writeU32(flags & objectMapFlagsMask4F4530)
	want.WriteByte(uint8(len("team-native")))
	want.WriteString("team-native")
	want.WriteByte(7)
	want.WriteByte(0)
	writeU16(0)
	writeU32(status & objectMapStatusMask4F4530)
	writeU16(1)
	writeU32(0)
	writeU32(handlerFlags)
	writeI32(int32(objectFrame - gameFrame))
	want.Write([]byte{0, 0, 0, 0})

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want.Bytes()) {
		t.Fatalf("wire payload = %x, want %x", got, want.Bytes())
	}
	if object.Field34 != objectFrame || object.InitData != unsafe.Pointer(initData) ||
		object.UpdateData != unsafe.Pointer(update) {
		t.Fatalf("native records changed: Field34=%#x init=%p update=%p",
			object.Field34, object.InitData, object.UpdateData)
	}
	if *update != (server.FlagUpdateData4EA490{Home: types.Pointf{X: 9, Y: 8}, State: 7}) {
		t.Fatalf("write mode changed flag update: %+v", *update)
	}
}

func TestTeamXferNativeRead4F6D20ResetsFlagHomeAndRestoresCount(t *testing.T) {
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
	writeU16(teamXferCurrentVersion4F6D20)
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
	payload.Write([]byte{0, 0, 0, 0})

	path := filepath.Join(t.TempDir(), "team-read.bin")
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
	initData, freeInitData := alloc.New(server.ModifierInitData{Field16: 0x11223344})
	defer freeInitData()
	update, freeUpdate := alloc.New(server.FlagUpdateData4EA490{
		Home:  types.Pointf{X: 1, Y: 2},
		State: 0x76543210,
	})
	defer freeUpdate()
	*update = server.FlagUpdateData4EA490{
		Home:  types.Pointf{X: 1, Y: 2},
		State: 0x76543210,
	}
	oldID, freeOldID := alloc.CString("")
	defer freeOldID()
	object.ObjClass = objectlib.ClassFlag | objectlib.ClassClientPersist
	object.IDPtr = unsafe.Pointer(oldID)
	object.ObjFlags = objectlib.Flags(originalFlags)
	object.Field5 = originalState
	object.Field34 = originalCount
	object.ScriptPickup.Func = -1
	object.InitData = unsafe.Pointer(initData)
	object.UpdateData = unsafe.Pointer(update)

	for name, pointer := range map[string]unsafe.Pointer{
		"object":      unsafe.Pointer(object),
		"init data":   unsafe.Pointer(initData),
		"update data": unsafe.Pointer(update),
		"existing ID": unsafe.Pointer(oldID),
	} {
		assertObjectMapNativePointer4F4530(t, name, pointer)
	}

	var applied server.ModifierInitData
	modifierLookups := 0
	var inventoryCalls []teamXferNativeInventoryCall4F6D20
	deps := teamXferRuntimeDeps4F6D20()
	deps.modifierIDByName = func(name string) int32 {
		modifierLookups++
		if name != "" {
			t.Fatalf("modifier name = %q, want empty", name)
		}
		return -1
	}
	deps.modifierByID = func(id int32) *server.ModifierEff {
		if id != -1 {
			t.Fatalf("modifier ID = %d, want -1", id)
		}
		return nil
	}
	deps.applyModifiers = func(gotObject *server.Object, attrs *server.ModifierInitData) {
		if gotObject != object {
			t.Fatalf("modifier object = %p, want %p", gotObject, object)
		}
		applied = *attrs
	}
	deps.transferInventory = func(version uint16, gotObject *server.Object, count int32) int32 {
		inventoryCalls = append(inventoryCalls, teamXferNativeInventoryCall4F6D20{
			version: version,
			object:  gotObject,
			count:   count,
		})
		gotObject.Field34 = 0x11223344
		return 1
	}

	if got := teamXferNative4F6D20(cf, object, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if modifierLookups != 4 || applied.Field16 != math.MaxUint32 ||
		applied.Modifiers != ([4]*server.ModifierEff{}) {
		t.Fatalf("modifier lookups/applied = %d/%+v, want 4/nil slots and max tail",
			modifierLookups, applied)
	}
	if update.Home != (types.Pointf{X: positionX, Y: positionY}) || update.State != 0x76543210 {
		t.Fatalf("flag update = %+v, want serialized home and preserved state", *update)
	}
	if len(inventoryCalls) != 1 || inventoryCalls[0] != (teamXferNativeInventoryCall4F6D20{
		version: teamXferCurrentVersion4F6D20,
		object:  object,
		count:   int32(inventoryCount),
	}) {
		t.Fatalf("inventory calls = %+v", inventoryCalls)
	}
	if object.Field34 != originalCount {
		t.Fatalf("Field34 = %#x, want entry %#x", object.Field34, originalCount)
	}
	if object.IDPtr != unsafe.Pointer(oldID) || object.InitData != unsafe.Pointer(initData) ||
		object.UpdateData != unsafe.Pointer(update) {
		t.Fatalf("native object-owned pointers changed")
	}
	if object.Extent != extent || object.ScriptIDVal != scriptID ||
		object.PosVec != (types.Pointf{X: positionX, Y: positionY}) {
		t.Fatalf("common object fields = extent %#x script %#x position %+v",
			object.Extent, uint32(object.ScriptIDVal), object.PosVec)
	}
}

func TestTeamXferExport4F6D20PreservesNativePointerAndResult(t *testing.T) {
	object, freeObject := alloc.New(server.Object{})
	defer freeObject()
	assertObjectMapNativePointer4F4530(t, "object", unsafe.Pointer(object))

	old := teamXferCall4F6D20
	t.Cleanup(func() { teamXferCall4F6D20 = old })
	calls := 0
	teamXferCall4F6D20 = func(_ *cryptfile.CryptFile, gotObject *server.Object) int32 {
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

	if got := teamXferExportCall4F6D20(object); got != math.MinInt32 {
		t.Fatalf("first result = %d, want %d", got, int32(math.MinInt32))
	}
	if got := teamXferExportCall4F6D20(nil); got != math.MaxInt32 {
		t.Fatalf("second result = %d, want %d", got, int32(math.MaxInt32))
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
	runtime.KeepAlive(object)
}
