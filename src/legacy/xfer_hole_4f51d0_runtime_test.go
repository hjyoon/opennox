package legacy

import (
	"bytes"
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"unsafe"

	objectlib "github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type holeXferNativeScriptCall4F51D0 struct {
	callback *server.ScriptCallback
	context  unsafe.Pointer
}

type holeXferNativeInventoryCall4F51D0 struct {
	version uint16
	object  *server.Object
	count   int32
}

func TestHoleXferNativeLayout4F51D0(t *testing.T) {
	type field struct {
		name string
		got  uintptr
		pe32 uintptr
		wide uintptr
	}
	fields := []field{
		{name: "Object size", got: unsafe.Sizeof(server.Object{}), pe32: 780, wide: 928},
		{name: "Object.Field34", got: unsafe.Offsetof(server.Object{}.Field34), pe32: 136, wide: 140},
		{name: "Object.CollideData", got: unsafe.Offsetof(server.Object{}.CollideData), pe32: 700, wide: 776},
		{name: "Object.Field189", got: unsafe.Offsetof(server.Object{}.Field189), pe32: 756, wide: 888},
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

	if got := unsafe.Sizeof(server.HoleCollideData{}); got != 28 {
		t.Errorf("Hole collide-data size = %d, want 28", got)
	}
	if got := unsafe.Sizeof(server.ScriptCallback{}); got != 8 {
		t.Errorf("ScriptCallback size = %d, want 8", got)
	}
	for _, field := range []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{name: "Script", got: unsafe.Offsetof(server.HoleCollideData{}.Script), want: 0},
		{name: "DestinationX", got: unsafe.Offsetof(server.HoleCollideData{}.DestinationX), want: 8},
		{name: "DestinationY", got: unsafe.Offsetof(server.HoleCollideData{}.DestinationY), want: 12},
		{name: "DestinationExtent", got: unsafe.Offsetof(server.HoleCollideData{}.DestinationExtent), want: 16},
		{name: "DestinationNetCode", got: unsafe.Offsetof(server.HoleCollideData{}.DestinationNetCode), want: 20},
		{name: "Reserved22", got: unsafe.Offsetof(server.HoleCollideData{}.Reserved22), want: 22},
		{name: "Field24", got: unsafe.Offsetof(server.HoleCollideData{}.Field24), want: 24},
		{name: "ScriptCallback.Func", got: unsafe.Offsetof(server.ScriptCallback{}.Func), want: 4},
	} {
		if field.got != field.want {
			t.Errorf("Hole collide-data %s offset = %d, want %d", field.name, field.got, field.want)
		}
	}

	if got := holeXferScriptContextNative4F51D0(nil); got != nil {
		t.Errorf("nil script-data context = %p, want nil", got)
	}
	scriptData, freeScriptData := alloc.Malloc(1024)
	defer freeScriptData()
	if got, want := holeXferScriptContextNative4F51D0(scriptData), unsafe.Add(scriptData, 128); got != want {
		t.Errorf("script-data context = %p, want %p", got, want)
	}
}

func TestHoleXferNativeWrite4F51D0PreservesPointersAndWire(t *testing.T) {
	object, freeObject := alloc.New(server.Object{})
	defer freeObject()
	data, freeData := alloc.New(server.HoleCollideData{})
	defer freeData()
	liveData, freeLiveData := alloc.New(server.HoleCollideData{})
	defer freeLiveData()
	scriptData, freeScriptData := alloc.Malloc(1024)
	defer freeScriptData()
	liveScriptData, freeLiveScriptData := alloc.Malloc(1024)
	defer freeLiveScriptData()
	id, freeID := alloc.CString("hole")
	defer freeID()

	for name, pointer := range map[string]unsafe.Pointer{
		"object":                 unsafe.Pointer(object),
		"Hole collide data":      unsafe.Pointer(data),
		"live Hole collide data": unsafe.Pointer(liveData),
		"script data":            scriptData,
		"live script data":       liveScriptData,
		"ID":                     unsafe.Pointer(id),
	} {
		assertObjectMapNativePointer4F4530(t, name, pointer)
	}

	const (
		extent            = uint32(0x11223344)
		scriptID          = int32(-0x1020304)
		positionX         = float32(123.25)
		positionY         = float32(-456.5)
		flags             = uint32(0x91408162)
		status            = uint32(0xa5)
		handlerFlags      = uint32(0xa1b2c3d4)
		objectFrame       = uint32(0x11223344)
		gameFrame         = uint32(0x01020304)
		field24           = uint32(0x55667788)
		destinationX      = int32(0x10203040)
		destinationY      = int32(-0x1020304)
		destinationExtent = uint32(0x99aabbcc)
		destinationNet    = uint16(0xddee)
	)
	data.Script = server.ScriptCallback{Flags: 0x01020304, Func: -7}
	data.DestinationX = destinationX
	data.DestinationY = destinationY
	data.DestinationExtent = destinationExtent
	data.DestinationNetCode = destinationNet
	data.Reserved22 = 0xbeef
	data.Field24 = field24
	liveData.Reserved22 = 0xcafe
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
	object.CollideData = unsafe.Pointer(data)
	object.Field189 = scriptData

	path := filepath.Join(t.TempDir(), "hole-write.bin")
	cf, err := cryptfile.OpenFile(path, cryptfile.WriteOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	setObjectMapRuntimeGlobals4F4530(t, cf, gameFrame)

	var scriptCalls []holeXferNativeScriptCall4F51D0
	inventoryCalls := 0
	deps := holeXferNativeDeps4F51D0{
		transferScript: func(callback *server.ScriptCallback, context unsafe.Pointer) {
			scriptCalls = append(scriptCalls, holeXferNativeScriptCall4F51D0{
				callback: callback,
				context:  context,
			})
			object.CollideData = unsafe.Pointer(liveData)
			object.Field189 = liveScriptData
		},
		transferInventory: func(uint16, *server.Object, int32) int32 {
			inventoryCalls++
			return 1
		},
	}
	if got := holeXferNative4F51D0(cf, object, deps); got != 1 {
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
	writeU16(holeXferCurrentVersion4F51D0)
	writeU16(objectMapCurrentVersion4F4530)
	writeU32(extent)
	writeI32(scriptID)
	writeU32(math.Float32bits(positionX))
	writeU32(math.Float32bits(positionY))
	want.WriteByte(math.MaxUint8)
	writeU32(flags & objectMapFlagsMask4F4530)
	want.WriteByte(uint8(len("hole")))
	want.WriteString("hole")
	want.WriteByte(7)
	want.WriteByte(0)
	writeU16(0)
	writeU32(status & objectMapStatusMask4F4530)
	writeU16(1)
	writeU32(0)
	writeU32(handlerFlags)
	writeI32(int32(objectFrame - gameFrame))
	writeU32(field24)
	writeI32(destinationX)
	writeI32(destinationY)
	writeU32(destinationExtent)
	writeU16(destinationNet)

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want.Bytes()) {
		t.Fatalf("wire payload = %x, want %x", got, want.Bytes())
	}
	wantScriptCalls := []holeXferNativeScriptCall4F51D0{{
		callback: &data.Script,
		context:  unsafe.Add(scriptData, 128),
	}}
	if !reflect.DeepEqual(scriptCalls, wantScriptCalls) {
		t.Fatalf("script calls = %#v, want %#v", scriptCalls, wantScriptCalls)
	}
	if inventoryCalls != 0 {
		t.Errorf("inventory calls = %d, want 0 in write mode", inventoryCalls)
	}
	if object.Field34 != objectFrame {
		t.Errorf("Field34 = %#08x, want entry value %#08x", object.Field34, objectFrame)
	}
	if object.CollideData != unsafe.Pointer(liveData) || object.Field189 != liveScriptData {
		t.Errorf("live pointers were not retained: collide=%p script=%p", object.CollideData, object.Field189)
	}
	if data.Reserved22 != 0xbeef || data.Field24 != field24 ||
		data.DestinationX != destinationX || data.DestinationY != destinationY ||
		data.DestinationExtent != destinationExtent || data.DestinationNetCode != destinationNet {
		t.Errorf("write changed cached Hole data: %+v", *data)
	}
	if liveData.Reserved22 != 0xcafe || liveData.Field24 != 0 {
		t.Errorf("write mutated replacement Hole data: %+v", *liveData)
	}
}

func TestHoleXferNativeRead4F51D0PreservesNativePointersAndEntryCaches(t *testing.T) {
	const (
		extent                = uint32(0x55667788)
		scriptID              = int32(0x10203040)
		positionX             = float32(30.5)
		positionY             = float32(-14.25)
		serialized            = uint32(0x01400102)
		status                = uint32(0x12)
		handlerFlags          = uint32(0x55667788)
		frameDelta            = int32(0x01020304)
		originalFlags         = uint32(0x80000040)
		originalState         = uint32(0xa5)
		originalCount         = uint32(0xfedcba98)
		inventoryCount        = uint8(3)
		serializedField24     = uint32(0x44332211)
		serializedX           = int32(0x10293847)
		serializedY           = int32(-0x1234567)
		serializedExtent      = uint32(0xa1b2c3d4)
		serializedDestination = uint16(0x5aa5)
	)

	var payload bytes.Buffer
	writeU16 := func(value uint16) { _ = binary.Write(&payload, binary.LittleEndian, value) }
	writeU32 := func(value uint32) { _ = binary.Write(&payload, binary.LittleEndian, value) }
	writeI32 := func(value int32) { writeU32(uint32(value)) }
	writeU16(holeXferCurrentVersion4F51D0)
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
	writeU32(serializedField24)
	writeI32(serializedX)
	writeI32(serializedY)
	writeU32(serializedExtent)
	writeU16(serializedDestination)

	path := filepath.Join(t.TempDir(), "hole-read.bin")
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
	data, freeData := alloc.New(server.HoleCollideData{})
	defer freeData()
	liveData, freeLiveData := alloc.New(server.HoleCollideData{})
	defer freeLiveData()
	scriptData, freeScriptData := alloc.Malloc(1024)
	defer freeScriptData()
	liveScriptData, freeLiveScriptData := alloc.Malloc(1024)
	defer freeLiveScriptData()
	oldID, freeOldID := alloc.CString("")
	defer freeOldID()

	data.Reserved22 = 0xbeef
	liveData.Reserved22 = 0xcafe
	object.ObjClass = objectlib.ClassPlayer | objectlib.ClassMissile
	object.IDPtr = unsafe.Pointer(oldID)
	object.ObjFlags = objectlib.Flags(originalFlags)
	object.Field5 = originalState
	object.Field34 = originalCount
	object.ScriptPickup.Func = -1
	object.CollideData = unsafe.Pointer(data)
	object.Field189 = scriptData
	for index := range object.Field140 {
		object.Field140[index] = 0xa5000abc | uint32(index<<12)
	}

	for name, pointer := range map[string]unsafe.Pointer{
		"object":                 unsafe.Pointer(object),
		"Hole collide data":      unsafe.Pointer(data),
		"live Hole collide data": unsafe.Pointer(liveData),
		"script data":            scriptData,
		"live script data":       liveScriptData,
		"old ID":                 unsafe.Pointer(oldID),
	} {
		assertObjectMapNativePointer4F4530(t, name, pointer)
	}

	var scriptCalls []holeXferNativeScriptCall4F51D0
	var inventoryCalls []holeXferNativeInventoryCall4F51D0
	deps := holeXferNativeDeps4F51D0{
		transferScript: func(callback *server.ScriptCallback, context unsafe.Pointer) {
			scriptCalls = append(scriptCalls, holeXferNativeScriptCall4F51D0{
				callback: callback,
				context:  context,
			})
			callback.Flags = 0x01020304
			callback.Func = -9
			object.CollideData = unsafe.Pointer(liveData)
			object.Field189 = liveScriptData
		},
		transferInventory: func(version uint16, gotObject *server.Object, count int32) int32 {
			inventoryCalls = append(inventoryCalls, holeXferNativeInventoryCall4F51D0{
				version: version,
				object:  gotObject,
				count:   count,
			})
			gotObject.Field34 = 0x11223344
			return 1
		},
	}
	if got := holeXferNative4F51D0(cf, object, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}

	wantScriptCalls := []holeXferNativeScriptCall4F51D0{{
		callback: &data.Script,
		context:  unsafe.Add(scriptData, 128),
	}}
	if !reflect.DeepEqual(scriptCalls, wantScriptCalls) {
		t.Fatalf("script calls = %#v, want %#v", scriptCalls, wantScriptCalls)
	}
	wantInventoryCalls := []holeXferNativeInventoryCall4F51D0{{
		version: holeXferCurrentVersion4F51D0,
		object:  object,
		count:   int32(inventoryCount),
	}}
	if !reflect.DeepEqual(inventoryCalls, wantInventoryCalls) {
		t.Fatalf("inventory calls = %#v, want %#v", inventoryCalls, wantInventoryCalls)
	}
	if object.Field34 != originalCount {
		t.Fatalf("Field34 = %#08x, want entry value %#08x", object.Field34, originalCount)
	}
	if object.CollideData != unsafe.Pointer(liveData) || object.Field189 != liveScriptData {
		t.Fatalf("live pointers changed unexpectedly: collide=%p script=%p", object.CollideData, object.Field189)
	}
	if object.IDPtr != unsafe.Pointer(oldID) {
		t.Fatalf("ID pointer = %p, want entry pointer %p", object.IDPtr, oldID)
	}
	if data.Script != (server.ScriptCallback{Flags: 0x01020304, Func: -9}) ||
		data.DestinationX != serializedX || data.DestinationY != serializedY ||
		data.DestinationExtent != serializedExtent ||
		data.DestinationNetCode != serializedDestination ||
		data.Reserved22 != 0xbeef || data.Field24 != serializedField24 {
		t.Fatalf("cached Hole data = %+v, want serialized values with reserved22 intact", *data)
	}
	if liveData.Reserved22 != 0xcafe || liveData.Field24 != 0 || liveData.Script != (server.ScriptCallback{}) {
		t.Fatalf("replacement Hole data mutated: %+v", *liveData)
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
