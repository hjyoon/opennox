package legacy

import (
	"bytes"
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	objectlib "github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type moverXferNativeInventoryCall4F5730 struct {
	version uint16
	object  *server.Object
	count   int32
}

func TestMoverXferNativeLayout4F5730(t *testing.T) {
	type field struct {
		name string
		got  uintptr
		pe32 uintptr
		wide uintptr
	}
	fields := []field{
		{name: "Object size", got: unsafe.Sizeof(server.Object{}), pe32: 780, wide: 928},
		{name: "Object.Field34", got: unsafe.Offsetof(server.Object{}.Field34), pe32: 136, wide: 140},
		{name: "Object.SpeedCur", got: unsafe.Offsetof(server.Object{}.SpeedCur), pe32: 544, wide: 604},
		{name: "Object.SpeedBase", got: unsafe.Offsetof(server.Object{}.SpeedBase), pe32: 548, wide: 608},
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

	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{name: "size", got: unsafe.Sizeof(server.MoverUpdateData{}), want: 36},
		{name: "Field_0", got: unsafe.Offsetof(server.MoverUpdateData{}.Field_0), want: 0},
		{name: "Field_1", got: unsafe.Offsetof(server.MoverUpdateData{}.Field_1), want: 4},
		{name: "Field_2", got: unsafe.Offsetof(server.MoverUpdateData{}.Field_2), want: 8},
		{name: "Field_3", got: unsafe.Offsetof(server.MoverUpdateData{}.Field_3), want: 12},
		{name: "Field_4", got: unsafe.Offsetof(server.MoverUpdateData{}.Field_4), want: 16},
		{name: "Field_5", got: unsafe.Offsetof(server.MoverUpdateData{}.Field_5), want: 20},
		{name: "Field_6", got: unsafe.Offsetof(server.MoverUpdateData{}.Field_6), want: 24},
		{name: "Field_7", got: unsafe.Offsetof(server.MoverUpdateData{}.Field_7), want: 28},
		{name: "Field_8", got: unsafe.Offsetof(server.MoverUpdateData{}.Field_8), want: 32},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("Mover update-data %s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestMoverXferNativeWrite4F5730UsesCachedRecordAndNativeWaypointIndexes(t *testing.T) {
	object, freeObject := alloc.New(server.Object{})
	defer freeObject()
	data, freeData := alloc.New(server.MoverUpdateData{})
	defer freeData()
	liveData, freeLiveData := alloc.New(server.MoverUpdateData{})
	defer freeLiveData()
	id, freeID := alloc.CString("mover")
	defer freeID()

	for name, pointer := range map[string]unsafe.Pointer{
		"object":                 unsafe.Pointer(object),
		"Mover update data":      unsafe.Pointer(data),
		"live Mover update data": unsafe.Pointer(liveData),
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
		field0       = byte(0x7d)
		field1       = float32(7.25)
		field2       = int32(-0x1020304)
		field3PE32   = uint32(0x89abcdef)
		field4       = uint32(0x10293847)
		field5PE32   = uint32(0x76543210)
		field6       = uint32(0x55667788)
		field7PE32   = uint32(0xa5b6c7d8)
		field8       = uint32(0x99aabbcc)
		waypoint3    = uint32(0x01020304)
		waypoint5    = uint32(0x05060708)
		speedBase    = float32(11.75)
		speedCur     = float32(6.5)
	)
	*data = server.MoverUpdateData{
		Field_0: field0, Field_1: field1, Field_2: field2,
		Field_3: field3PE32, Field_4: field4,
		Field_5: field5PE32, Field_6: field6,
		Field_7: field7PE32, Field_8: field8,
	}
	liveData.Field_8 = 0x44556677
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
	object.SpeedBase = speedBase
	object.SpeedCur = speedCur
	object.UpdateData = unsafe.Pointer(data)

	path := filepath.Join(t.TempDir(), "mover-write.bin")
	cf, err := cryptfile.OpenFile(path, cryptfile.WriteOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	setObjectMapRuntimeGlobals4F4530(t, cf, gameFrame)

	var waypointSlots []int
	var waypointData []*server.MoverUpdateData
	deps := moverXferNativeDeps4F5730{
		waypointIndex: func(gotObject *server.Object, gotData *server.MoverUpdateData, slot int) uint32 {
			if gotObject != object {
				t.Fatalf("waypoint object = %p, want %p", gotObject, object)
			}
			waypointSlots = append(waypointSlots, slot)
			waypointData = append(waypointData, gotData)
			if slot == 3 {
				object.UpdateData = unsafe.Pointer(liveData)
				return waypoint3
			}
			return waypoint5
		},
		transferInventory: func(uint16, *server.Object, int32) int32 {
			t.Fatal("write mode transferred inventory")
			return 0
		},
	}
	if got := moverXferNative4F5730(cf, object, deps); got != 1 {
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
	writeU16(moverXferCurrentVersion4F5730)
	writeU16(objectMapCurrentVersion4F4530)
	writeU32(extent)
	writeI32(scriptID)
	writeU32(math.Float32bits(positionX))
	writeU32(math.Float32bits(positionY))
	want.WriteByte(math.MaxUint8)
	writeU32(flags & objectMapFlagsMask4F4530)
	want.WriteByte(uint8(len("mover")))
	want.WriteString("mover")
	want.WriteByte(7)
	want.WriteByte(0)
	writeU16(0)
	writeU32(status & objectMapStatusMask4F4530)
	writeU16(1)
	writeU32(0)
	writeU32(handlerFlags)
	writeI32(int32(objectFrame - gameFrame))
	writeU32(math.Float32bits(field1))
	writeI32(field2)
	writeU32(field8)
	want.WriteByte(field0)
	writeU32(waypoint3)
	writeU32(waypoint5)
	writeU32(math.Float32bits(speedBase))
	writeU32(math.Float32bits(speedCur))

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want.Bytes()) {
		t.Fatalf("wire payload = %x, want %x", got, want.Bytes())
	}
	if !reflect.DeepEqual(waypointSlots, []int{3, 5}) ||
		!reflect.DeepEqual(waypointData, []*server.MoverUpdateData{data, data}) {
		t.Fatalf("waypoint slots/data = %v/%p, want [3 5]/cached %p", waypointSlots, waypointData, data)
	}
	if object.UpdateData != unsafe.Pointer(liveData) || object.Field34 != objectFrame {
		t.Fatalf("live UpdateData/Field34 = %p/%#x, want %p/%#x",
			object.UpdateData, object.Field34, liveData, objectFrame)
	}
	wantData := server.MoverUpdateData{
		Field_0: field0, Field_1: field1, Field_2: field2,
		Field_3: field3PE32, Field_4: field4,
		Field_5: field5PE32, Field_6: field6,
		Field_7: field7PE32, Field_8: field8,
	}
	if *data != wantData || liveData.Field_8 != 0x44556677 {
		t.Fatalf("cached/live update records = %+v/%+v, want %+v/live extent unchanged", *data, *liveData, wantData)
	}
}

func TestMoverXferNativeRead4F5730WritesCachedRecordAndRestoresCount(t *testing.T) {
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
		serializedField0 = byte(0x6e)
		serializedField1 = float32(-8.75)
		serializedField2 = int32(-0x1234567)
		serializedField4 = uint32(0xa1b2c3d4)
		serializedField6 = uint32(0x10293847)
		serializedField8 = uint32(0x99aabbcc)
		serializedBase   = float32(15.5)
		serializedCur    = float32(9.25)
	)

	var payload bytes.Buffer
	writeU16 := func(value uint16) { _ = binary.Write(&payload, binary.LittleEndian, value) }
	writeU32 := func(value uint32) { _ = binary.Write(&payload, binary.LittleEndian, value) }
	writeI32 := func(value int32) { writeU32(uint32(value)) }
	writeU16(moverXferCurrentVersion4F5730)
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
	writeU32(math.Float32bits(serializedField1))
	writeI32(serializedField2)
	writeU32(serializedField8)
	payload.WriteByte(serializedField0)
	writeU32(serializedField4)
	writeU32(serializedField6)
	writeU32(math.Float32bits(serializedBase))
	writeU32(math.Float32bits(serializedCur))

	path := filepath.Join(t.TempDir(), "mover-read.bin")
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
	data, freeData := alloc.New(server.MoverUpdateData{})
	defer freeData()
	liveData, freeLiveData := alloc.New(server.MoverUpdateData{})
	defer freeLiveData()
	oldID, freeOldID := alloc.CString("")
	defer freeOldID()
	*data = server.MoverUpdateData{
		Field_0: 0x22, Field_1: 1.5, Field_2: 7,
		Field_3: 0x89abcdef, Field_4: 0x11111111,
		Field_5: 0x76543210, Field_6: 0x22222222,
		Field_7: 0xa5b6c7d8, Field_8: 0x33333333,
	}
	liveData.Field_8 = 0x44444444
	object.ObjClass = objectlib.ClassPlayer | objectlib.ClassMissile
	object.IDPtr = unsafe.Pointer(oldID)
	object.ObjFlags = objectlib.Flags(originalFlags)
	object.Field5 = originalState
	object.Field34 = originalCount
	object.SpeedBase = 2.5
	object.SpeedCur = 3.5
	object.ScriptPickup.Func = -1
	object.UpdateData = unsafe.Pointer(data)

	for name, pointer := range map[string]unsafe.Pointer{
		"object":                 unsafe.Pointer(object),
		"Mover update data":      unsafe.Pointer(data),
		"live Mover update data": unsafe.Pointer(liveData),
		"old ID":                 unsafe.Pointer(oldID),
	} {
		assertObjectMapNativePointer4F4530(t, name, pointer)
	}

	var inventoryCalls []moverXferNativeInventoryCall4F5730
	deps := moverXferNativeDeps4F5730{
		waypointIndex: func(*server.Object, *server.MoverUpdateData, int) uint32 {
			t.Fatal("read mode resolved a native waypoint")
			return 0
		},
		transferInventory: func(version uint16, gotObject *server.Object, count int32) int32 {
			inventoryCalls = append(inventoryCalls, moverXferNativeInventoryCall4F5730{
				version: version, object: gotObject, count: count,
			})
			gotObject.UpdateData = unsafe.Pointer(liveData)
			gotObject.Field34 = 0x11223344
			return 1
		},
	}
	if got := moverXferNative4F5730(cf, object, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}

	wantInventory := []moverXferNativeInventoryCall4F5730{{
		version: moverXferCurrentVersion4F5730,
		object:  object,
		count:   int32(inventoryCount),
	}}
	if !reflect.DeepEqual(inventoryCalls, wantInventory) {
		t.Fatalf("inventory calls = %#v, want %#v", inventoryCalls, wantInventory)
	}
	if data.Field_0 != serializedField0 || data.Field_1 != serializedField1 ||
		data.Field_2 != serializedField2 || data.Field_4 != serializedField4 ||
		data.Field_6 != serializedField6 || data.Field_8 != serializedField8 {
		t.Fatalf("cached update record = %+v", *data)
	}
	if data.Field_3 != 0x89abcdef || data.Field_5 != 0x76543210 || data.Field_7 != 0xa5b6c7d8 {
		t.Fatalf("transient PE32 slots changed on read: %#x/%#x/%#x", data.Field_3, data.Field_5, data.Field_7)
	}
	if object.SpeedBase != serializedBase || object.SpeedCur != serializedCur {
		t.Fatalf("speeds = %v/%v, want base/current %v/%v",
			object.SpeedBase, object.SpeedCur, serializedBase, serializedCur)
	}
	if liveData.Field_8 != 0x44444444 || object.UpdateData != unsafe.Pointer(liveData) {
		t.Fatalf("live extent/UpdateData = %#x/%p, want 0x44444444/%p",
			liveData.Field_8, object.UpdateData, liveData)
	}
	if object.Field34 != originalCount {
		t.Fatalf("Field34 = %#x, want entry %#x", object.Field34, originalCount)
	}
	if object.IDPtr != unsafe.Pointer(oldID) {
		t.Fatalf("zero-length ID changed native pointer: got %p want %p", object.IDPtr, oldID)
	}
}

func TestMoverXferExport4F5730PreservesNativePointersAndResult(t *testing.T) {
	object, freeObject := alloc.New(server.Object{})
	defer freeObject()
	context, freeContext := alloc.New(uint64(0))
	defer freeContext()

	assertObjectMapNativePointer4F4530(t, "object", unsafe.Pointer(object))
	assertObjectMapNativePointer4F4530(t, "context", unsafe.Pointer(context))

	old := moverXferCall4F5730
	t.Cleanup(func() { moverXferCall4F5730 = old })
	calls := 0
	moverXferCall4F5730 = func(
		_ *cryptfile.CryptFile,
		gotObject *server.Object,
		gotContext unsafe.Pointer,
	) int32 {
		calls++
		if gotObject != object {
			t.Fatalf("object = %p, want %p", gotObject, object)
		}
		switch calls {
		case 1:
			if gotContext != unsafe.Pointer(context) {
				t.Fatalf("context = %p, want %p", gotContext, context)
			}
			return math.MinInt32
		case 2:
			if gotContext != nil {
				t.Fatalf("context = %p, want nil", gotContext)
			}
			return math.MaxInt32
		default:
			t.Fatalf("unexpected call %d", calls)
			return 0
		}
	}

	if got := moverXferExportCall4F5730(object, unsafe.Pointer(context)); got != math.MinInt32 {
		t.Fatalf("first result = %d, want %d", got, int32(math.MinInt32))
	}
	if got := moverXferExportCall4F5730(object, nil); got != math.MaxInt32 {
		t.Fatalf("second result = %d, want %d", got, int32(math.MaxInt32))
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
	runtime.KeepAlive(object)
	runtime.KeepAlive(context)
}
