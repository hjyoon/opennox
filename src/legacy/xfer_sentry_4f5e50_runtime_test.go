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

type sentryXferNativeInventoryCall4F5E50 struct {
	version uint16
	object  *server.Object
	count   int32
}

func TestSentryXferNativeLayout4F5E50(t *testing.T) {
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

	if got := unsafe.Sizeof(server.SentryUpdateData{}); got != sentryXferUpdateSize4F5E50 {
		t.Errorf("Sentry update-data size = %d, want %d", got, sentryXferUpdateSize4F5E50)
	}
	for _, field := range []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{name: "Field0", got: unsafe.Offsetof(server.SentryUpdateData{}.Field0), want: 0},
		{name: "Field4", got: unsafe.Offsetof(server.SentryUpdateData{}.Field4), want: 4},
		{name: "Field8", got: unsafe.Offsetof(server.SentryUpdateData{}.Field8), want: 8},
	} {
		if field.got != field.want {
			t.Errorf("Sentry update-data %s offset = %d, want %d", field.name, field.got, field.want)
		}
	}
}

func TestSentryXferNativeWrite4F5E50PreservesFixedRecordAndExactWire(t *testing.T) {
	object, freeObject := alloc.New(server.Object{})
	defer freeObject()
	data, freeData := alloc.New(server.SentryUpdateData{})
	defer freeData()
	id, freeID := alloc.CString("sentry")
	defer freeID()

	for name, pointer := range map[string]unsafe.Pointer{
		"object":             unsafe.Pointer(object),
		"Sentry update data": unsafe.Pointer(data),
		"ID":                 unsafe.Pointer(id),
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
		field4       = uint32(0x55667788)
		field8       = uint32(0x99aabbcc)
	)
	data.Field0 = field0
	data.Field4 = field4
	data.Field8 = field8
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

	path := filepath.Join(t.TempDir(), "sentry-write.bin")
	cf, err := cryptfile.OpenFile(path, cryptfile.WriteOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	setObjectMapRuntimeGlobals4F4530(t, cf, gameFrame)

	gameFlagCalls := 0
	deps := sentryXferNativeDeps4F5E50{
		gameFlags: func(mask uint32) int32 {
			gameFlagCalls++
			if mask != sentryXferGameMask4F5E50 {
				t.Fatalf("game flag mask = %#x, want %#x", mask, sentryXferGameMask4F5E50)
			}
			return 1
		},
		transferInventory: func(uint16, *server.Object, int32) int32 {
			t.Fatal("write mode transferred inventory")
			return 0
		},
	}
	if got := sentryXferNative4F5E50(cf, object, deps); got != 1 {
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
	writeU16(sentryXferCurrentVersion4F5E50)
	writeU16(objectMapCurrentVersion4F4530)
	writeU32(extent)
	writeI32(scriptID)
	writeU32(math.Float32bits(positionX))
	writeU32(math.Float32bits(positionY))
	want.WriteByte(math.MaxUint8)
	writeU32(flags & objectMapFlagsMask4F4530)
	want.WriteByte(uint8(len("sentry")))
	want.WriteString("sentry")
	want.WriteByte(7)
	want.WriteByte(0)
	writeU16(0)
	writeU32(status & objectMapStatusMask4F4530)
	writeU16(1)
	writeU32(0)
	writeU32(handlerFlags)
	writeI32(int32(objectFrame - gameFrame))
	writeU32(field4)
	writeU32(field8)
	writeU32(field4)

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want.Bytes()) {
		t.Fatalf("wire payload = %x, want %x", got, want.Bytes())
	}
	if gameFlagCalls != 1 {
		t.Errorf("game flag calls = %d, want 1", gameFlagCalls)
	}
	if data.Field0 != field4 || data.Field4 != field4 || data.Field8 != field8 {
		t.Fatalf("fixed update record = %+v, want {%#x %#x %#x}", *data, field4, field4, field8)
	}
	if object.Field34 != objectFrame || object.UpdateData != unsafe.Pointer(data) {
		t.Fatalf("Field34/UpdateData = %#x/%p, want %#x/%p",
			object.Field34, object.UpdateData, objectFrame, data)
	}
}

func TestSentryXferNativeRead4F5E50UsesCachedRecordAndRestoresCount(t *testing.T) {
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
		serializedField0 = uint32(0xa1b2c3d4)
		serializedField4 = uint32(0x10293847)
		serializedField8 = uint32(0x89abcdef)
	)

	var payload bytes.Buffer
	writeU16 := func(value uint16) { _ = binary.Write(&payload, binary.LittleEndian, value) }
	writeU32 := func(value uint32) { _ = binary.Write(&payload, binary.LittleEndian, value) }
	writeI32 := func(value int32) { writeU32(uint32(value)) }
	writeU16(sentryXferCurrentVersion4F5E50)
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
	writeU32(serializedField4)
	writeU32(serializedField8)
	writeU32(serializedField0)

	path := filepath.Join(t.TempDir(), "sentry-read.bin")
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
	data, freeData := alloc.New(server.SentryUpdateData{})
	defer freeData()
	liveData, freeLiveData := alloc.New(server.SentryUpdateData{})
	defer freeLiveData()
	oldID, freeOldID := alloc.CString("")
	defer freeOldID()
	data.Field0 = 0x0badcafe
	data.Field4 = 0x11111111
	data.Field8 = 0x22222222
	liveData.Field0 = 0x33333333
	liveData.Field4 = 0x44444444
	liveData.Field8 = 0x55555555
	object.ObjClass = objectlib.ClassPlayer | objectlib.ClassMissile
	object.IDPtr = unsafe.Pointer(oldID)
	object.ObjFlags = objectlib.Flags(originalFlags)
	object.Field5 = originalState
	object.Field34 = originalCount
	object.ScriptPickup.Func = -1
	object.UpdateData = unsafe.Pointer(data)

	for name, pointer := range map[string]unsafe.Pointer{
		"object":                  unsafe.Pointer(object),
		"Sentry update data":      unsafe.Pointer(data),
		"live Sentry update data": unsafe.Pointer(liveData),
		"old ID":                  unsafe.Pointer(oldID),
	} {
		assertObjectMapNativePointer4F4530(t, name, pointer)
	}

	var inventoryCalls []sentryXferNativeInventoryCall4F5E50
	deps := sentryXferNativeDeps4F5E50{
		gameFlags: func(uint32) int32 {
			t.Fatal("read mode called game flags")
			return 0
		},
		transferInventory: func(version uint16, gotObject *server.Object, count int32) int32 {
			inventoryCalls = append(inventoryCalls, sentryXferNativeInventoryCall4F5E50{
				version: version,
				object:  gotObject,
				count:   count,
			})
			gotObject.UpdateData = unsafe.Pointer(liveData)
			gotObject.Field34 = 0x11223344
			return 1
		},
	}
	if got := sentryXferNative4F5E50(cf, object, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}

	if len(inventoryCalls) != 1 || inventoryCalls[0] != (sentryXferNativeInventoryCall4F5E50{
		version: sentryXferCurrentVersion4F5E50,
		object:  object,
		count:   int32(inventoryCount),
	}) {
		t.Fatalf("inventory calls = %+v", inventoryCalls)
	}
	if data.Field0 != serializedField0 || data.Field4 != serializedField4 || data.Field8 != serializedField8 {
		t.Fatalf("cached update record = %+v, want {%#x %#x %#x}",
			*data, serializedField0, serializedField4, serializedField8)
	}
	if *liveData != (server.SentryUpdateData{
		Field0: 0x33333333,
		Field4: 0x44444444,
		Field8: 0x55555555,
	}) || object.UpdateData != unsafe.Pointer(liveData) {
		t.Fatalf("live update record/UpdateData = %+v/%p, want unchanged/%p",
			*liveData, object.UpdateData, liveData)
	}
	if object.Field34 != originalCount {
		t.Fatalf("Field34 = %#x, want entry %#x", object.Field34, originalCount)
	}
	if object.IDPtr != unsafe.Pointer(oldID) {
		t.Fatalf("zero-length ID changed native pointer: got %p want %p", object.IDPtr, oldID)
	}
}

func TestSentryXferExport4F5E50PreservesNativePointerAndResult(t *testing.T) {
	object, freeObject := alloc.New(server.Object{})
	defer freeObject()
	assertObjectMapNativePointer4F4530(t, "object", unsafe.Pointer(object))

	old := sentryXferCall4F5E50
	t.Cleanup(func() { sentryXferCall4F5E50 = old })
	calls := 0
	sentryXferCall4F5E50 = func(_ *cryptfile.CryptFile, gotObject *server.Object) int32 {
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

	if got := sentryXferExportCall4F5E50(object); got != math.MinInt32 {
		t.Fatalf("first result = %d, want %d", got, int32(math.MinInt32))
	}
	if got := sentryXferExportCall4F5E50(nil); got != math.MaxInt32 {
		t.Fatalf("second result = %d, want %d", got, int32(math.MaxInt32))
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
	runtime.KeepAlive(object)
}
