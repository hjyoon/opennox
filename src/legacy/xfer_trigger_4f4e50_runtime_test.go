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

type triggerXferNativeScriptCall4F4E50 struct {
	callback *server.ScriptCallback
	context  unsafe.Pointer
}

type triggerXferNativeInventoryCall4F4E50 struct {
	version uint16
	object  *server.Object
	count   int32
}

func TestTriggerXferNativeLayout4F4E50(t *testing.T) {
	type field struct {
		name string
		got  uintptr
		pe32 uintptr
		wide uintptr
	}
	fields := []field{
		{name: "Object size", got: unsafe.Sizeof(server.Object{}), pe32: 780, wide: 928},
		{name: "Object.Field33", got: unsafe.Offsetof(server.Object{}.Field33), pe32: 132, wide: 136},
		{name: "Object.Field34", got: unsafe.Offsetof(server.Object{}.Field34), pe32: 136, wide: 140},
		{name: "Object.Shape", got: unsafe.Offsetof(server.Object{}.Shape), pe32: 172, wide: 176},
		{name: "Object.Shape.Box.W", got: unsafe.Offsetof(server.Object{}.Shape) + unsafe.Offsetof(server.Shape{}.Box) + unsafe.Offsetof(server.ShapeBox{}.W), pe32: 184, wide: 188},
		{name: "Object.Shape.Box.H", got: unsafe.Offsetof(server.Object{}.Shape) + unsafe.Offsetof(server.Shape{}.Box) + unsafe.Offsetof(server.ShapeBox{}.H), pe32: 188, wide: 192},
		{name: "Object.UpdateData", got: unsafe.Offsetof(server.Object{}.UpdateData), pe32: 748, wide: 872},
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

	if got := unsafe.Sizeof(server.TriggerUpdateData{}); got != 60 {
		t.Errorf("Trigger update-data size = %d, want 60", got)
	}
	if got := unsafe.Sizeof(server.ScriptCallback{}); got != 8 {
		t.Errorf("ScriptCallback size = %d, want 8", got)
	}
	for _, field := range []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{name: "Flags", got: unsafe.Offsetof(server.TriggerUpdateData{}.Flags), want: 0},
		{name: "Field4", got: unsafe.Offsetof(server.TriggerUpdateData{}.Field4), want: 4},
		{name: "State", got: unsafe.Offsetof(server.TriggerUpdateData{}.State), want: 8},
		{name: "Field9", got: unsafe.Offsetof(server.TriggerUpdateData{}.Field9), want: 9},
		{name: "ScriptCollide", got: unsafe.Offsetof(server.TriggerUpdateData{}.ScriptCollide), want: 12},
		{name: "ScriptActivate", got: unsafe.Offsetof(server.TriggerUpdateData{}.ScriptActivate), want: 20},
		{name: "ScriptDeactivate", got: unsafe.Offsetof(server.TriggerUpdateData{}.ScriptDeactivate), want: 28},
		{name: "SoundActivate", got: unsafe.Offsetof(server.TriggerUpdateData{}.SoundActivate), want: 36},
		{name: "SoundDeactivate", got: unsafe.Offsetof(server.TriggerUpdateData{}.SoundDeactivate), want: 40},
		{name: "ClassInclude", got: unsafe.Offsetof(server.TriggerUpdateData{}.ClassInclude), want: 44},
		{name: "ClassExclude", got: unsafe.Offsetof(server.TriggerUpdateData{}.ClassExclude), want: 48},
		{name: "TeamInclude", got: unsafe.Offsetof(server.TriggerUpdateData{}.TeamInclude), want: 52},
		{name: "TeamExclude", got: unsafe.Offsetof(server.TriggerUpdateData{}.TeamExclude), want: 53},
		{name: "Colors", got: unsafe.Offsetof(server.TriggerUpdateData{}.Colors), want: 54},
		{name: "ScriptCallback.Func", got: unsafe.Offsetof(server.ScriptCallback{}.Func), want: 4},
	} {
		if field.got != field.want {
			t.Errorf("Trigger update-data %s offset = %d, want %d", field.name, field.got, field.want)
		}
	}

	if got := triggerXferScriptContextNative4F4E50(nil, 256); got != nil {
		t.Errorf("nil script-data context = %p, want nil", got)
	}
	scriptData, freeScriptData := alloc.Malloc(1024)
	defer freeScriptData()
	if got, want := triggerXferScriptContextNative4F4E50(scriptData, 512), unsafe.Add(scriptData, 512); got != want {
		t.Errorf("script-data context = %p, want %p", got, want)
	}
}

func TestTriggerXferNativeWrite4F4E50PreservesPointersAndWire(t *testing.T) {
	object, freeObject := alloc.New(server.Object{})
	defer freeObject()
	update, freeUpdate := alloc.New(server.TriggerUpdateData{})
	defer freeUpdate()
	liveUpdate, freeLiveUpdate := alloc.New(server.TriggerUpdateData{})
	defer freeLiveUpdate()
	scriptData, freeScriptData := alloc.Malloc(1024)
	defer freeScriptData()
	liveScriptData, freeLiveScriptData := alloc.Malloc(1024)
	defer freeLiveScriptData()
	id, freeID := alloc.CString("trigger")
	defer freeID()

	for name, pointer := range map[string]unsafe.Pointer{
		"object":                   unsafe.Pointer(object),
		"Trigger update data":      unsafe.Pointer(update),
		"live Trigger update data": unsafe.Pointer(liveUpdate),
		"script data":              scriptData,
		"live script data":         liveScriptData,
		"ID":                       unsafe.Pointer(id),
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
		width        = float32(12.75)
		height       = float32(-2.9)
		triggerFlags = uint32(0x55667788)
		classInclude = uint32(0x01020304)
		classExclude = uint32(0xa0b0c0d0)
		field33      = uint32(0x8899aabb)
	)
	colors := [6]uint8{1, 3, 5, 7, 9, 11}
	update.Flags = triggerFlags
	update.State = 0x12
	update.Field9 = 0x34
	update.ScriptCollide = server.ScriptCallback{Flags: 0x11111111, Func: 1}
	update.ScriptActivate = server.ScriptCallback{Flags: 0x22222222, Func: 2}
	update.ScriptDeactivate = server.ScriptCallback{Flags: 0x33333333, Func: 3}
	update.ClassInclude = classInclude
	update.ClassExclude = classExclude
	update.TeamInclude = 0x56
	update.TeamExclude = 0x78
	update.Colors = colors
	liveUpdate.Flags = 0xdeadbeef
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
	object.Field33 = field33
	object.Field34 = objectFrame
	object.Shape.Kind = server.ShapeKindBox
	object.Shape.Box.W = width
	object.Shape.Box.H = height
	object.UpdateData = unsafe.Pointer(update)
	object.Field189 = scriptData

	path := filepath.Join(t.TempDir(), "trigger-write.bin")
	cf, err := cryptfile.OpenFile(path, cryptfile.WriteOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	setObjectMapRuntimeGlobals4F4530(t, cf, gameFrame)

	var scriptCalls []triggerXferNativeScriptCall4F4E50
	inventoryCalls := 0
	deps := triggerXferNativeDeps4F4E50{
		transferScript: func(callback *server.ScriptCallback, context unsafe.Pointer) {
			scriptCalls = append(scriptCalls, triggerXferNativeScriptCall4F4E50{
				callback: callback,
				context:  context,
			})
			if len(scriptCalls) == 1 {
				object.UpdateData = unsafe.Pointer(liveUpdate)
				object.Field189 = liveScriptData
			}
		},
		initLegacyScript: func(*server.ScriptCallback) {
			t.Fatal("legacy script initializer called for version 61")
		},
		transferInventory: func(uint16, *server.Object, int32) int32 {
			inventoryCalls++
			return 1
		},
	}
	if got := triggerXferNative4F4E50(cf, object, deps); got != 1 {
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
	writeU16(triggerXferCurrentVersion4F4E50)
	writeU16(objectMapCurrentVersion4F4530)
	writeU32(extent)
	writeI32(scriptID)
	writeU32(math.Float32bits(positionX))
	writeU32(math.Float32bits(positionY))
	want.WriteByte(math.MaxUint8)
	writeU32(flags & objectMapFlagsMask4F4530)
	want.WriteByte(uint8(len("trigger")))
	want.WriteString("trigger")
	want.WriteByte(7)
	want.WriteByte(0)
	writeU16(0)
	writeU32(status & objectMapStatusMask4F4530)
	writeU16(1)
	writeU32(0)
	writeU32(handlerFlags)
	writeI32(int32(objectFrame - gameFrame))
	writeI32(12)
	writeI32(-2)
	want.Write(colors[:])
	writeU32(triggerFlags)
	writeU32(classInclude)
	writeU32(classExclude)
	want.WriteByte(0x56)
	want.WriteByte(0x78)
	want.WriteByte(0x12)
	want.WriteByte(0x34)
	writeU32(field33)

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want.Bytes()) {
		t.Fatalf("wire payload = %x, want %x", got, want.Bytes())
	}
	wantScriptCalls := []triggerXferNativeScriptCall4F4E50{
		{callback: &update.ScriptActivate, context: unsafe.Add(scriptData, 256)},
		{callback: &update.ScriptDeactivate, context: unsafe.Add(scriptData, 384)},
		{callback: &update.ScriptCollide, context: unsafe.Add(scriptData, 512)},
	}
	if !reflect.DeepEqual(scriptCalls, wantScriptCalls) {
		t.Fatalf("script calls = %#v, want %#v", scriptCalls, wantScriptCalls)
	}
	if inventoryCalls != 0 {
		t.Errorf("inventory calls = %d, want 0 in write mode", inventoryCalls)
	}
	if object.Field34 != objectFrame {
		t.Errorf("Field34 = %#08x, want entry value %#08x", object.Field34, objectFrame)
	}
	if object.UpdateData != unsafe.Pointer(liveUpdate) || object.Field189 != liveScriptData {
		t.Errorf("live pointers were not retained: update=%p script=%p", object.UpdateData, object.Field189)
	}
	if update.Flags != triggerFlags || update.ClassInclude != classInclude ||
		update.ClassExclude != classExclude || update.Colors != colors {
		t.Errorf("write changed cached Trigger update data: %+v", *update)
	}
	if liveUpdate.Flags != 0xdeadbeef || liveUpdate.ClassInclude != 0 {
		t.Errorf("write mutated replacement Trigger update data: %+v", *liveUpdate)
	}
}

func TestTriggerXferNativeRead4F4E50PreservesNativePointersAndEntryCaches(t *testing.T) {
	const (
		extent          = uint32(0x55667788)
		scriptID        = int32(0x10203040)
		positionX       = float32(30.5)
		positionY       = float32(-14.25)
		serialized      = uint32(0x01400102)
		status          = uint32(0x12)
		handlerFlags    = uint32(0x55667788)
		frameDelta      = int32(0x01020304)
		originalFlags   = uint32(0x80000040)
		originalState   = uint32(0xa5)
		originalCount   = uint32(0xfedcba98)
		inventoryCount  = uint8(3)
		triggerFlags    = uint32(0x44332211)
		classInclude    = uint32(0x10293847)
		classExclude    = uint32(0xa1b2c3d4)
		serializedFrame = uint32(0x0badf00d)
	)
	colors := [6]uint8{0xa1, 0xb2, 0xc3, 0xd4, 0xe5, 0xf6}

	var payload bytes.Buffer
	writeU16 := func(value uint16) { _ = binary.Write(&payload, binary.LittleEndian, value) }
	writeU32 := func(value uint32) { _ = binary.Write(&payload, binary.LittleEndian, value) }
	writeI32 := func(value int32) { writeU32(uint32(value)) }
	writeU16(triggerXferCurrentVersion4F4E50)
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
	writeI32(70)
	writeI32(-5)
	payload.Write(colors[:])
	writeU32(triggerFlags)
	writeU32(classInclude)
	writeU32(classExclude)
	payload.WriteByte(0x5a)
	payload.WriteByte(0xa5)
	payload.WriteByte(3)
	payload.WriteByte(4)
	writeU32(serializedFrame)

	path := filepath.Join(t.TempDir(), "trigger-read.bin")
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
	update, freeUpdate := alloc.New(server.TriggerUpdateData{})
	defer freeUpdate()
	liveUpdate, freeLiveUpdate := alloc.New(server.TriggerUpdateData{})
	defer freeLiveUpdate()
	scriptData, freeScriptData := alloc.Malloc(1024)
	defer freeScriptData()
	liveScriptData, freeLiveScriptData := alloc.Malloc(1024)
	defer freeLiveScriptData()
	oldID, freeOldID := alloc.CString("")
	defer freeOldID()

	update.Flags = 0xffffffff
	update.ClassInclude = 0xffffffff
	update.ClassExclude = 0xffffffff
	update.TeamInclude = 0xff
	update.TeamExclude = 0xff
	liveUpdate.Flags = 0xdeadbeef
	object.ObjClass = objectlib.ClassPlayer | objectlib.ClassMissile
	object.IDPtr = unsafe.Pointer(oldID)
	object.ObjFlags = objectlib.Flags(originalFlags)
	object.Field5 = originalState
	object.Field33 = 0x11111111
	object.Field34 = originalCount
	object.ScriptPickup.Func = -1
	object.Shape.Kind = server.ShapeKindBox
	object.UpdateData = unsafe.Pointer(update)
	object.Field189 = scriptData
	for index := range object.Field140 {
		object.Field140[index] = 0xa5000abc | uint32(index<<12)
	}

	for name, pointer := range map[string]unsafe.Pointer{
		"object":                   unsafe.Pointer(object),
		"Trigger update data":      unsafe.Pointer(update),
		"live Trigger update data": unsafe.Pointer(liveUpdate),
		"script data":              scriptData,
		"live script data":         liveScriptData,
		"old ID":                   unsafe.Pointer(oldID),
	} {
		assertObjectMapNativePointer4F4530(t, name, pointer)
	}

	var scriptCalls []triggerXferNativeScriptCall4F4E50
	var inventoryCalls []triggerXferNativeInventoryCall4F4E50
	deps := triggerXferNativeDeps4F4E50{
		transferScript: func(callback *server.ScriptCallback, context unsafe.Pointer) {
			scriptCalls = append(scriptCalls, triggerXferNativeScriptCall4F4E50{
				callback: callback,
				context:  context,
			})
			if len(scriptCalls) == 1 {
				object.UpdateData = unsafe.Pointer(liveUpdate)
				object.Field189 = liveScriptData
			}
		},
		initLegacyScript: func(*server.ScriptCallback) {
			t.Fatal("legacy script initializer called for version 61")
		},
		transferInventory: func(version uint16, gotObject *server.Object, count int32) int32 {
			inventoryCalls = append(inventoryCalls, triggerXferNativeInventoryCall4F4E50{
				version: version,
				object:  gotObject,
				count:   count,
			})
			gotObject.Field34 = 0x11223344
			return 1
		},
	}
	if got := triggerXferNative4F4E50(cf, object, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}

	wantScriptCalls := []triggerXferNativeScriptCall4F4E50{
		{callback: &update.ScriptActivate, context: unsafe.Add(scriptData, 256)},
		{callback: &update.ScriptDeactivate, context: unsafe.Add(scriptData, 384)},
		{callback: &update.ScriptCollide, context: unsafe.Add(scriptData, 512)},
	}
	if !reflect.DeepEqual(scriptCalls, wantScriptCalls) {
		t.Fatalf("script calls = %#v, want %#v", scriptCalls, wantScriptCalls)
	}
	wantInventoryCalls := []triggerXferNativeInventoryCall4F4E50{
		{version: triggerXferCurrentVersion4F4E50, object: object, count: int32(inventoryCount)},
	}
	if !reflect.DeepEqual(inventoryCalls, wantInventoryCalls) {
		t.Fatalf("inventory calls = %#v, want %#v", inventoryCalls, wantInventoryCalls)
	}
	if object.Field34 != originalCount {
		t.Fatalf("Field34 = %#08x, want entry value %#08x", object.Field34, originalCount)
	}
	if object.UpdateData != unsafe.Pointer(liveUpdate) || object.Field189 != liveScriptData {
		t.Fatalf("live pointers changed unexpectedly: update=%p script=%p", object.UpdateData, object.Field189)
	}
	if object.IDPtr != unsafe.Pointer(oldID) {
		t.Fatalf("ID pointer = %p, want entry pointer %p", object.IDPtr, oldID)
	}
	if object.Shape.Box.W != 60 || object.Shape.Box.H != -5 {
		t.Errorf("shape = %g x %g, want 60 x -5", object.Shape.Box.W, object.Shape.Box.H)
	}
	if update.Colors != colors || update.Flags != triggerFlags ||
		update.ClassInclude != classInclude || update.ClassExclude != classExclude ||
		update.TeamInclude != 0x5a || update.TeamExclude != 0xa5 ||
		update.State != 3 || update.Field9 != 4 {
		t.Fatalf("cached Trigger update data = %+v, want serialized values", *update)
	}
	if liveUpdate.Flags != 0xdeadbeef || liveUpdate.ClassInclude != 0 || liveUpdate.Colors != [6]uint8{} {
		t.Fatalf("replacement Trigger update data mutated: %+v", *liveUpdate)
	}
	if object.Field33 != serializedFrame || object.Field38 != math.MaxUint32 {
		t.Errorf("animation state = frame %#08x sync %#08x, want %#08x/ffffffff",
			object.Field33, object.Field38, serializedFrame)
	}
	for index, value := range object.Field140 {
		if value&0xfff != 0 || value&0x10000 == 0 || value&0xff000000 != 0xa5000000 {
			t.Errorf("Field140[%d] = %#08x, want cleared low bits, animation marker, and preserved high bits", index, value)
		}
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
