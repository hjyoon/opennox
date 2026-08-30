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

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type obeliskXferNativeInventoryCall4F6F60 struct {
	version uint16
	object  *server.Object
	count   int32
}

func TestObeliskXferNativeLayout4F6F60(t *testing.T) {
	type field struct {
		name string
		got  uintptr
		pe32 uintptr
		wide uintptr
	}
	fields := []field{
		{name: "Object size", got: unsafe.Sizeof(server.Object{}), pe32: 780, wide: 928},
		{name: "Object.Field34", got: unsafe.Offsetof(server.Object{}.Field34), pe32: 136, wide: 140},
		{name: "Object.UpdateData", got: unsafe.Offsetof(server.Object{}.UpdateData), pe32: 748, wide: 872},
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
		{name: "ObeliskUpdateData size", got: unsafe.Sizeof(server.ObeliskUpdateData{}), want: 4},
		{name: "ObeliskUpdateData.Mana", got: unsafe.Offsetof(server.ObeliskUpdateData{}.Mana), want: 0},
	}
	for _, field := range fixed {
		if field.got != field.want {
			t.Errorf("%s = %d, want %d", field.name, field.got, field.want)
		}
	}
}

func TestObeliskXferNativeWrite4F6F60PreservesPointersAndExactWire(t *testing.T) {
	object, freeObject := alloc.New(server.Object{})
	defer freeObject()
	data, freeData := alloc.New(server.ObeliskUpdateData{Mana: -0x1020304})
	defer freeData()
	*data = server.ObeliskUpdateData{Mana: -0x1020304}
	id, freeID := alloc.CString("obelisk-native")
	defer freeID()

	for name, pointer := range map[string]unsafe.Pointer{
		"object":      unsafe.Pointer(object),
		"update data": unsafe.Pointer(data),
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
	object.ObjClass = objectlib.ClassImmobile
	object.ObjSubClass = objectlib.SubClass(objectlib.OtherVisibleObelisk)
	object.Extent = extent
	object.ScriptIDVal = scriptID
	object.PosVec = types.Pointf{X: positionX, Y: positionY}
	object.ObjFlags = objectlib.Flags(flags)
	object.IDPtr = unsafe.Pointer(id)
	object.TeamVal.ID = server.TeamID(7)
	object.Field5 = status
	object.ScriptPickup = server.ScriptCallback{Flags: handlerFlags, Func: -1}
	object.Field34 = objectFrame
	object.UpdateData = unsafe.Pointer(data)

	path := filepath.Join(t.TempDir(), "obelisk-write.bin")
	cf, err := cryptfile.OpenFile(path, cryptfile.WriteOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	setObjectMapRuntimeGlobals4F4530(t, cf, gameFrame)
	if got := Nox_xxx_XFerObeliskNative4F6F60(cf, object); got != 1 {
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
	writeU16(obeliskXferCurrentVersion4F6F60)
	writeU16(objectMapCurrentVersion4F4530)
	writeU32(extent)
	writeI32(scriptID)
	writeU32(math.Float32bits(positionX))
	writeU32(math.Float32bits(positionY))
	want.WriteByte(math.MaxUint8)
	writeU32(flags & objectMapFlagsMask4F4530)
	want.WriteByte(uint8(len("obelisk-native")))
	want.WriteString("obelisk-native")
	want.WriteByte(7)
	want.WriteByte(0)
	writeU16(0)
	writeU32(status & objectMapStatusMask4F4530)
	writeU16(1)
	writeU32(0)
	writeU32(handlerFlags)
	writeI32(int32(objectFrame - gameFrame))
	writeI32(data.Mana)
	want.WriteByte(0)

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want.Bytes()) {
		t.Fatalf("wire payload = %x, want %x", got, want.Bytes())
	}
	if object.Field34 != objectFrame || object.UpdateData != unsafe.Pointer(data) {
		t.Fatalf("native records changed: Field34=%#x update=%p", object.Field34, object.UpdateData)
	}
	if data.Mana != -0x1020304 {
		t.Fatalf("mana = %d, want %d", data.Mana, int32(-0x1020304))
	}
}

func TestObeliskXferNativeRead4F6F60UsesNativeDrawablesAndRestoresCount(t *testing.T) {
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
		mana           = int32(125)
		minimapWire    = uint8(0x7f)
		streamSentinel = uint8(0x5a)
	)

	var payload bytes.Buffer
	writeU16 := func(value uint16) { _ = binary.Write(&payload, binary.LittleEndian, value) }
	writeU32 := func(value uint32) { _ = binary.Write(&payload, binary.LittleEndian, value) }
	writeI32 := func(value int32) { writeU32(uint32(value)) }
	writeU16(obeliskXferCurrentVersion4F6F60)
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
	writeI32(mana)
	// The live native list computes one; this non-one stream byte must still
	// be consumed and otherwise ignored on read, matching GAME.EXE.
	payload.WriteByte(minimapWire)
	payload.WriteByte(streamSentinel)

	path := filepath.Join(t.TempDir(), "obelisk-read.bin")
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
	data, freeData := alloc.New(server.ObeliskUpdateData{Mana: 1})
	defer freeData()
	*data = server.ObeliskUpdateData{Mana: 1}
	oldID, freeOldID := alloc.CString("")
	defer freeOldID()
	object.ObjClass = objectlib.ClassImmobile
	object.ObjSubClass = objectlib.SubClass(objectlib.OtherVisibleObelisk)
	object.IDPtr = unsafe.Pointer(oldID)
	object.ObjFlags = objectlib.Flags(originalFlags)
	object.Field5 = originalState
	object.Field34 = originalCount
	object.ScriptPickup.Func = -1
	object.UpdateData = unsafe.Pointer(data)

	for name, pointer := range map[string]unsafe.Pointer{
		"object":      unsafe.Pointer(object),
		"update data": unsafe.Pointer(data),
		"existing ID": unsafe.Pointer(oldID),
	} {
		assertObjectMapNativePointer4F4530(t, name, pointer)
	}

	first := new(client.Drawable)
	target := new(client.Drawable)
	first.Field_102 = target

	var inventoryCalls []obeliskXferNativeInventoryCall4F6F60
	var syncCalls int
	var gameFlagCalls int
	var staticCalls int
	var firstCalls int
	deps := obeliskXferRuntimeDeps4F6F60()
	deps.syncManaLevel = func(gotObject *server.Object, level float32) {
		syncCalls++
		if gotObject != object || math.Float32bits(level) != math.Float32bits(200) {
			t.Fatalf("mana sync = object %p level %v, want %p/200", gotObject, level, object)
		}
	}
	deps.gameFlags = func(mask uint32) int32 {
		gameFlagCalls++
		if mask != obeliskXferQuestFlag4F6F60 {
			t.Fatalf("game flag mask = %#x, want %#x", mask, obeliskXferQuestFlag4F6F60)
		}
		return 1
	}
	deps.staticDrawable = func(code uint32) *client.Drawable {
		staticCalls++
		if code != extent {
			t.Fatalf("static drawable code = %#x, want live extent %#x", code, extent)
		}
		return target
	}
	deps.firstMinimap = func() *client.Drawable {
		firstCalls++
		return first
	}
	deps.transferInventory = func(version uint16, gotObject *server.Object, count int32) int32 {
		inventoryCalls = append(inventoryCalls, obeliskXferNativeInventoryCall4F6F60{
			version: version,
			object:  gotObject,
			count:   count,
		})
		gotObject.Field34 = 0x11223344
		return 1
	}

	if got := obeliskXferNative4F6F60(cf, object, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if data.Mana != mana {
		t.Fatalf("mana = %d, want %d", data.Mana, mana)
	}
	if syncCalls != 1 || gameFlagCalls != 1 || staticCalls != 1 || firstCalls != 1 {
		t.Fatalf("dependency calls = sync %d flags %d static %d first %d, want all 1",
			syncCalls, gameFlagCalls, staticCalls, firstCalls)
	}
	if len(inventoryCalls) != 1 || inventoryCalls[0] != (obeliskXferNativeInventoryCall4F6F60{
		version: obeliskXferCurrentVersion4F6F60,
		object:  object,
		count:   int32(inventoryCount),
	}) {
		t.Fatalf("inventory calls = %+v", inventoryCalls)
	}
	if object.Field34 != originalCount {
		t.Fatalf("Field34 = %#x, want entry %#x", object.Field34, originalCount)
	}
	if object.IDPtr != unsafe.Pointer(oldID) || object.UpdateData != unsafe.Pointer(data) {
		t.Fatalf("native object-owned pointers changed")
	}
	if object.Extent != extent || object.ScriptIDVal != scriptID ||
		object.PosVec != (types.Pointf{X: positionX, Y: positionY}) {
		t.Fatalf("common fields = extent %#x script %#x position %+v",
			object.Extent, uint32(object.ScriptIDVal), object.PosVec)
	}
	if got := objectReadOldRWU8Native4F4170(cf, 0); got != streamSentinel {
		t.Fatalf("byte after minimap payload = %#x, want sentinel %#x", got, streamSentinel)
	}
}

func TestObeliskXferExport4F6F60PreservesNativePointerAndResult(t *testing.T) {
	object, freeObject := alloc.New(server.Object{})
	defer freeObject()
	assertObjectMapNativePointer4F4530(t, "object", unsafe.Pointer(object))

	old := obeliskXferCall4F6F60
	t.Cleanup(func() { obeliskXferCall4F6F60 = old })
	calls := 0
	obeliskXferCall4F6F60 = func(_ *cryptfile.CryptFile, gotObject *server.Object) int32 {
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

	if got := obeliskXferExportCall4F6F60(object); got != math.MinInt32 {
		t.Fatalf("first result = %d, want %d", got, int32(math.MinInt32))
	}
	if got := obeliskXferExportCall4F6F60(nil); got != math.MaxInt32 {
		t.Fatalf("second result = %d, want %d", got, int32(math.MaxInt32))
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
	runtime.KeepAlive(object)
}
