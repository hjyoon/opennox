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

type elevatorXferNativeInventoryCall4F53D0 struct {
	version uint16
	object  *server.Object
	count   int32
}

func TestElevatorXferNativeLayout4F53D0(t *testing.T) {
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

	if got := unsafe.Sizeof(server.ElevatorUpdateData{}); got != 20 {
		t.Errorf("Elevator update-data size = %d, want 20", got)
	}
	for _, field := range []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{name: "Field_0", got: unsafe.Offsetof(server.ElevatorUpdateData{}.Field_0), want: 0},
		{name: "Field_1", got: unsafe.Offsetof(server.ElevatorUpdateData{}.Field_1), want: 4},
		{name: "Field_2", got: unsafe.Offsetof(server.ElevatorUpdateData{}.Field_2), want: 8},
		{name: "Field_3", got: unsafe.Offsetof(server.ElevatorUpdateData{}.Field_3), want: 12},
		{name: "Field_4", got: unsafe.Offsetof(server.ElevatorUpdateData{}.Field_4), want: 16},
	} {
		if field.got != field.want {
			t.Errorf("Elevator update-data %s offset = %d, want %d", field.name, field.got, field.want)
		}
	}
}

func TestElevatorXferNativeWrite4F53D0PreservesFixedRecordAndWire(t *testing.T) {
	object, freeObject := alloc.New(server.Object{})
	defer freeObject()
	data, freeData := alloc.New(server.ElevatorUpdateData{})
	defer freeData()
	id, freeID := alloc.CString("elevator")
	defer freeID()

	for name, pointer := range map[string]unsafe.Pointer{
		"object":               unsafe.Pointer(object),
		"Elevator update data": unsafe.Pointer(data),
		"ID":                   unsafe.Pointer(id),
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
		field0       = uint32(0x0badcafe)
		linkPE32     = uint32(0x89abcdef)
		shaftExtent  = uint32(0x99aabbcc)
		field3       = byte(0x7d)
		field4       = uint32(0x55667788)
	)
	data.Field_0 = field0
	data.Field_1 = linkPE32
	data.Field_2 = shaftExtent
	data.Field_3 = field3
	data.Field_4 = field4
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
	object.UpdateData = unsafe.Pointer(data)

	path := filepath.Join(t.TempDir(), "elevator-write.bin")
	cf, err := cryptfile.OpenFile(path, cryptfile.WriteOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	setObjectMapRuntimeGlobals4F4530(t, cf, gameFrame)

	deps := elevatorXferNativeDeps4F53D0{
		transferInventory: func(uint16, *server.Object, int32) int32 {
			t.Fatal("write mode transferred inventory")
			return 0
		},
	}
	if got := elevatorXferNative4F53D0(cf, object, deps); got != 1 {
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
	writeU16(elevatorXferCurrentVersion4F53D0)
	writeU16(objectMapCurrentVersion4F4530)
	writeU32(extent)
	writeI32(scriptID)
	writeU32(math.Float32bits(positionX))
	writeU32(math.Float32bits(positionY))
	want.WriteByte(math.MaxUint8)
	writeU32(flags & objectMapFlagsMask4F4530)
	want.WriteByte(uint8(len("elevator")))
	want.WriteString("elevator")
	want.WriteByte(7)
	want.WriteByte(0)
	writeU16(0)
	writeU32(status & objectMapStatusMask4F4530)
	writeU16(1)
	writeU32(0)
	writeU32(handlerFlags)
	writeI32(int32(objectFrame - gameFrame))
	writeU32(shaftExtent)
	writeU32(field4)
	want.WriteByte(field3)

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want.Bytes()) {
		t.Fatalf("wire payload = %x, want %x", got, want.Bytes())
	}
	if data.Field_0 != field0 || data.Field_1 != linkPE32 ||
		data.Field_2 != shaftExtent || data.Field_3 != field3 || data.Field_4 != field4 {
		t.Fatalf("fixed update record changed: %+v", *data)
	}
	if object.Field34 != objectFrame || object.UpdateData != unsafe.Pointer(data) {
		t.Fatalf("Field34/UpdateData = %#x/%p, want %#x/%p",
			object.Field34, object.UpdateData, objectFrame, data)
	}
}

func TestElevatorXferNativeRead4F53D0WritesCachedRecordAndRestoresCount(t *testing.T) {
	const (
		extent           = uint32(0x55667788)
		scriptID         = int32(0x10203040)
		positionX        = float32(-321.25)
		positionY        = float32(654.5)
		serialized       = uint32(0x01400102)
		status           = uint32(0x12)
		handlerFlags     = uint32(0x55667788)
		frameDelta       = int32(0x01020304)
		originalFlags    = uint32(0x80000040)
		originalState    = uint32(0xa5)
		originalCount    = uint32(0xfedcba98)
		inventoryCount   = uint8(3)
		serializedShaft  = uint32(0xa1b2c3d4)
		serializedField4 = uint32(0x10293847)
		serializedField3 = byte(0x6e)
	)

	var payload bytes.Buffer
	writeU16 := func(value uint16) { _ = binary.Write(&payload, binary.LittleEndian, value) }
	writeU32 := func(value uint32) { _ = binary.Write(&payload, binary.LittleEndian, value) }
	writeI32 := func(value int32) { writeU32(uint32(value)) }
	writeU16(elevatorXferCurrentVersion4F53D0)
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
	writeU32(serializedShaft)
	writeU32(serializedField4)
	payload.WriteByte(serializedField3)

	path := filepath.Join(t.TempDir(), "elevator-read.bin")
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
	data, freeData := alloc.New(server.ElevatorUpdateData{})
	defer freeData()
	liveData, freeLiveData := alloc.New(server.ElevatorUpdateData{})
	defer freeLiveData()
	oldID, freeOldID := alloc.CString("")
	defer freeOldID()
	data.Field_0 = 0x0badcafe
	data.Field_1 = 0x89abcdef
	data.Field_2 = 0x11111111
	data.Field_3 = 0x22
	data.Field_4 = 0x33333333
	liveData.Field_2 = 0x44444444
	object.ObjClass = objectlib.ClassPlayer | objectlib.ClassMissile
	object.IDPtr = unsafe.Pointer(oldID)
	object.ObjFlags = objectlib.Flags(originalFlags)
	object.Field5 = originalState
	object.Field34 = originalCount
	object.ScriptPickup.Func = -1
	object.UpdateData = unsafe.Pointer(data)

	for name, pointer := range map[string]unsafe.Pointer{
		"object":                    unsafe.Pointer(object),
		"Elevator update data":      unsafe.Pointer(data),
		"live Elevator update data": unsafe.Pointer(liveData),
		"old ID":                    unsafe.Pointer(oldID),
	} {
		assertObjectMapNativePointer4F4530(t, name, pointer)
	}

	var inventoryCalls []elevatorXferNativeInventoryCall4F53D0
	deps := elevatorXferNativeDeps4F53D0{
		transferInventory: func(version uint16, gotObject *server.Object, count int32) int32 {
			inventoryCalls = append(inventoryCalls, elevatorXferNativeInventoryCall4F53D0{
				version: version, object: gotObject, count: count,
			})
			gotObject.UpdateData = unsafe.Pointer(liveData)
			gotObject.Field34 = 0x11223344
			return 1
		},
	}
	if got := elevatorXferNative4F53D0(cf, object, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}

	wantInventory := []elevatorXferNativeInventoryCall4F53D0{{
		version: elevatorXferCurrentVersion4F53D0,
		object:  object,
		count:   int32(inventoryCount),
	}}
	if !reflect.DeepEqual(inventoryCalls, wantInventory) {
		t.Fatalf("inventory calls = %#v, want %#v", inventoryCalls, wantInventory)
	}
	if data.Field_0 != 0x0badcafe || data.Field_1 != 0x89abcdef ||
		data.Field_2 != serializedShaft || data.Field_3 != serializedField3 ||
		data.Field_4 != serializedField4 {
		t.Fatalf("cached update record = %+v", *data)
	}
	if liveData.Field_2 != 0x44444444 || object.UpdateData != unsafe.Pointer(liveData) {
		t.Fatalf("live extent/UpdateData = %#x/%p, want 0x44444444/%p",
			liveData.Field_2, object.UpdateData, liveData)
	}
	if object.Field34 != originalCount {
		t.Fatalf("Field34 = %#x, want entry %#x", object.Field34, originalCount)
	}
	if object.IDPtr != unsafe.Pointer(oldID) {
		t.Fatalf("zero-length ID changed native pointer: got %p want %p", object.IDPtr, oldID)
	}
}
