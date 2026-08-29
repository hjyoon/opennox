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

type transporterXferNativeInventoryCall4F5300 struct {
	version uint16
	object  *server.Object
	count   int32
}

func TestTransporterXferNativeLayout4F5300(t *testing.T) {
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

	if got := unsafe.Sizeof(server.TransporterUpdateData{}); got != 20 {
		t.Errorf("Transporter update-data size = %d, want 20", got)
	}
	if got := unsafe.Offsetof(server.TransporterUpdateData{}.TargetPE32); got != 12 {
		t.Errorf("TargetPE32 offset = %d, want 12", got)
	}
	if got := unsafe.Offsetof(server.TransporterUpdateData{}.TargetExtent); got != 16 {
		t.Errorf("TargetExtent offset = %d, want 16", got)
	}
}

func TestTransporterXferNativeWrite4F5300KeepsCachedExtentBesideNativePointer(t *testing.T) {
	object, freeObject := alloc.New(server.Object{})
	defer freeObject()
	data, freeData := alloc.New(server.TransporterUpdateData{})
	defer freeData()
	liveData, freeLiveData := alloc.New(server.TransporterUpdateData{})
	defer freeLiveData()
	id, freeID := alloc.CString("transporter")
	defer freeID()

	for name, pointer := range map[string]unsafe.Pointer{
		"object":                  unsafe.Pointer(object),
		"Transporter update data": unsafe.Pointer(data),
		"live update data":        unsafe.Pointer(liveData),
		"ID":                      unsafe.Pointer(id),
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
		targetExtent = uint32(0x99aabbcc)
	)
	data.TargetExtent = targetExtent
	liveData.TargetExtent = 0x55667788
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

	path := filepath.Join(t.TempDir(), "transporter-write.bin")
	cf, err := cryptfile.OpenFile(path, cryptfile.WriteOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	setObjectMapRuntimeGlobals4F4530(t, cf, gameFrame)

	var targetCalls []*server.TransporterUpdateData
	deps := transporterXferNativeDeps4F5300{
		hasTarget: func(gotObject *server.Object, gotData *server.TransporterUpdateData) bool {
			if gotObject != object {
				t.Fatalf("target object = %p, want %p", gotObject, object)
			}
			targetCalls = append(targetCalls, gotData)
			object.UpdateData = unsafe.Pointer(liveData)
			return true
		},
		transferInventory: func(uint16, *server.Object, int32) int32 {
			t.Fatal("write mode transferred inventory")
			return 0
		},
	}
	if got := transporterXferNative4F5300(cf, object, deps); got != 1 {
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
	writeU16(transporterXferCurrentVersion4F5300)
	writeU16(objectMapCurrentVersion4F4530)
	writeU32(extent)
	writeI32(scriptID)
	writeU32(math.Float32bits(positionX))
	writeU32(math.Float32bits(positionY))
	want.WriteByte(math.MaxUint8)
	writeU32(flags & objectMapFlagsMask4F4530)
	want.WriteByte(uint8(len("transporter")))
	want.WriteString("transporter")
	want.WriteByte(7)
	want.WriteByte(0)
	writeU16(0)
	writeU32(status & objectMapStatusMask4F4530)
	writeU16(1)
	writeU32(0)
	writeU32(handlerFlags)
	writeI32(int32(objectFrame - gameFrame))
	writeU32(targetExtent)

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want.Bytes()) {
		t.Fatalf("wire payload = %x, want %x", got, want.Bytes())
	}
	if !reflect.DeepEqual(targetCalls, []*server.TransporterUpdateData{data}) {
		t.Fatalf("target-data calls = %p, want cached %p", targetCalls, data)
	}
	if object.UpdateData != unsafe.Pointer(liveData) || object.Field34 != objectFrame {
		t.Fatalf("live UpdateData/Field34 = %p/%#x, want %p/%#x",
			object.UpdateData, object.Field34, liveData, objectFrame)
	}
	if data.TargetPE32 != 0 || data.TargetExtent != targetExtent || liveData.TargetExtent != 0x55667788 {
		t.Fatalf("cached PE32/extent/live extent = %#x/%#x/%#x",
			data.TargetPE32, data.TargetExtent, liveData.TargetExtent)
	}
}

func TestTransporterXferNativeRead4F5300WritesCachedExtentAndRestoresCount(t *testing.T) {
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
		serializedTarget = uint32(0xa1b2c3d4)
	)

	var payload bytes.Buffer
	writeU16 := func(value uint16) { _ = binary.Write(&payload, binary.LittleEndian, value) }
	writeU32 := func(value uint32) { _ = binary.Write(&payload, binary.LittleEndian, value) }
	writeI32 := func(value int32) { writeU32(uint32(value)) }
	writeU16(transporterXferCurrentVersion4F5300)
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
	writeU32(serializedTarget)

	path := filepath.Join(t.TempDir(), "transporter-read.bin")
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
	data, freeData := alloc.New(server.TransporterUpdateData{})
	defer freeData()
	liveData, freeLiveData := alloc.New(server.TransporterUpdateData{})
	defer freeLiveData()
	oldID, freeOldID := alloc.CString("")
	defer freeOldID()
	data.TargetPE32 = 0
	liveData.TargetExtent = 0x12345678
	object.ObjClass = objectlib.ClassPlayer | objectlib.ClassMissile
	object.IDPtr = unsafe.Pointer(oldID)
	object.ObjFlags = objectlib.Flags(originalFlags)
	object.Field5 = originalState
	object.Field34 = originalCount
	object.ScriptPickup.Func = -1
	object.UpdateData = unsafe.Pointer(data)

	for name, pointer := range map[string]unsafe.Pointer{
		"object":                  unsafe.Pointer(object),
		"Transporter update data": unsafe.Pointer(data),
		"live update data":        unsafe.Pointer(liveData),
		"old ID":                  unsafe.Pointer(oldID),
	} {
		assertObjectMapNativePointer4F4530(t, name, pointer)
	}

	var inventoryCalls []transporterXferNativeInventoryCall4F5300
	deps := transporterXferNativeDeps4F5300{
		hasTarget: func(*server.Object, *server.TransporterUpdateData) bool {
			t.Fatal("read mode tested native target state")
			return false
		},
		transferInventory: func(version uint16, gotObject *server.Object, count int32) int32 {
			inventoryCalls = append(inventoryCalls, transporterXferNativeInventoryCall4F5300{
				version: version, object: gotObject, count: count,
			})
			gotObject.UpdateData = unsafe.Pointer(liveData)
			gotObject.Field34 = 0x11223344
			return 1
		},
	}
	if got := transporterXferNative4F5300(cf, object, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}

	wantInventory := []transporterXferNativeInventoryCall4F5300{{
		version: transporterXferCurrentVersion4F5300,
		object:  object,
		count:   int32(inventoryCount),
	}}
	if !reflect.DeepEqual(inventoryCalls, wantInventory) {
		t.Fatalf("inventory calls = %#v, want %#v", inventoryCalls, wantInventory)
	}
	if data.TargetPE32 != 0 || data.TargetExtent != serializedTarget {
		t.Fatalf("cached PE32/extent = %#x/%#x, want 0/%#x",
			data.TargetPE32, data.TargetExtent, serializedTarget)
	}
	if liveData.TargetExtent != 0x12345678 || object.UpdateData != unsafe.Pointer(liveData) {
		t.Fatalf("live extent/UpdateData = %#x/%p, want 0x12345678/%p",
			liveData.TargetExtent, object.UpdateData, liveData)
	}
	if object.Field34 != originalCount {
		t.Fatalf("Field34 = %#x, want entry %#x", object.Field34, originalCount)
	}
	if object.IDPtr != unsafe.Pointer(oldID) {
		t.Fatalf("zero-length ID changed native pointer: got %p want %p", object.IDPtr, oldID)
	}
}
