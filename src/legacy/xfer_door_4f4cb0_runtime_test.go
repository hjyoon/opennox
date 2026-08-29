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

type doorXferNativeAttachCall4F4CB0 struct {
	object *server.Object
	tileX  int32
	tileY  int32
}

type doorXferNativeInventoryCall4F4CB0 struct {
	version uint16
	object  *server.Object
	count   int32
}

func TestDoorXferNativeLayout4F4CB0(t *testing.T) {
	type field struct {
		name string
		got  uintptr
		pe32 uintptr
		wide uintptr
	}
	fields := []field{
		{name: "Object size", got: unsafe.Sizeof(server.Object{}), pe32: 780, wide: 928},
		{name: "Object.PosVec", got: unsafe.Offsetof(server.Object{}.PosVec), pe32: 56, wide: 60},
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

	if got := unsafe.Sizeof(server.DoorUpdateData{}); got != 52 {
		t.Errorf("Door update-data size = %d, want 52", got)
	}
	for _, field := range []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{name: "LockCode", got: unsafe.Offsetof(server.DoorUpdateData{}.LockCode), want: 1},
		{name: "TargetDirection", got: unsafe.Offsetof(server.DoorUpdateData{}.TargetDirection), want: 4},
		{name: "SyncedDirection", got: unsafe.Offsetof(server.DoorUpdateData{}.SyncedDirection), want: 8},
		{name: "CurrentDirection", got: unsafe.Offsetof(server.DoorUpdateData{}.CurrentDirection), want: 12},
		{name: "TileX", got: unsafe.Offsetof(server.DoorUpdateData{}.TileX), want: 16},
		{name: "TileY", got: unsafe.Offsetof(server.DoorUpdateData{}.TileY), want: 20},
		{name: "FractionalDir", got: unsafe.Offsetof(server.DoorUpdateData{}.FractionalDir), want: 40},
	} {
		if field.got != field.want {
			t.Errorf("Door update-data %s offset = %d, want %d", field.name, field.got, field.want)
		}
	}
}

func TestDoorXferNativeWrite4F4CB0PreservesPointersAndWire(t *testing.T) {
	object, freeObject := alloc.New(server.Object{})
	defer freeObject()
	update, freeUpdate := alloc.New(server.DoorUpdateData{})
	defer freeUpdate()
	id, freeID := alloc.CString("door")
	defer freeID()

	assertObjectMapNativePointer4F4530(t, "object", unsafe.Pointer(object))
	assertObjectMapNativePointer4F4530(t, "Door update data", unsafe.Pointer(update))
	assertObjectMapNativePointer4F4530(t, "ID", unsafe.Pointer(id))

	const (
		extent          = uint32(0x11223344)
		scriptID        = int32(-0x1020304)
		positionX       = float32(123.25)
		positionY       = float32(-456.5)
		flags           = uint32(0x91408162)
		status          = uint32(0xa5)
		handlerFlags    = uint32(0xa1b2c3d4)
		objectFrame     = uint32(0x11223344)
		gameFrame       = uint32(0x01020304)
		direction       = int32(-17)
		lockCode        = uint8(0xfe)
		targetDirection = int32(29)
	)
	update.CurrentDirection = direction
	update.LockCode = lockCode
	update.TargetDirection = targetDirection
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
	object.UpdateData = unsafe.Pointer(update)

	path := filepath.Join(t.TempDir(), "door-write.bin")
	cf, err := cryptfile.OpenFile(path, cryptfile.WriteOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	setObjectMapRuntimeGlobals4F4530(t, cf, gameFrame)
	if got := Nox_xxx_XFerDoorNative4F4CB0(cf, object); got != 1 {
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
	writeU16(doorXferCurrentVersion4F4CB0)
	writeU16(objectMapCurrentVersion4F4530)
	writeU32(extent)
	writeI32(scriptID)
	writeU32(math.Float32bits(positionX))
	writeU32(math.Float32bits(positionY))
	want.WriteByte(math.MaxUint8)
	writeU32(flags & objectMapFlagsMask4F4530)
	want.WriteByte(uint8(len("door")))
	want.WriteString("door")
	want.WriteByte(7)
	want.WriteByte(0)
	writeU16(0)
	writeU32(status & objectMapStatusMask4F4530)
	writeU16(1)
	writeU32(0)
	writeU32(handlerFlags)
	writeI32(int32(objectFrame - gameFrame))
	writeI32(direction)
	writeI32(int32(lockCode))
	writeI32(targetDirection)

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want.Bytes()) {
		t.Fatalf("wire payload = %x, want %x", got, want.Bytes())
	}
	if object.Field34 != objectFrame {
		t.Errorf("Field34 = %#08x, want entry value %#08x", object.Field34, objectFrame)
	}
	if object.IDPtr != unsafe.Pointer(id) || object.UpdateData != unsafe.Pointer(update) {
		t.Errorf("native pointers changed: ID=%p update=%p", object.IDPtr, object.UpdateData)
	}
	if update.CurrentDirection != direction || update.LockCode != lockCode ||
		update.TargetDirection != targetDirection {
		t.Errorf("write changed Door update data: %+v", *update)
	}
}

func TestDoorXferNativeRead4F4CB0PreservesNativePointersAndEntryCaches(t *testing.T) {
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
		direction       = int32(7)
		lockCode        = int32(0x1234abcd)
		targetDirection = int32(12)
		inventoryCount  = uint8(3)
	)

	var payload bytes.Buffer
	writeU16 := func(value uint16) { _ = binary.Write(&payload, binary.LittleEndian, value) }
	writeU32 := func(value uint32) { _ = binary.Write(&payload, binary.LittleEndian, value) }
	writeI32 := func(value int32) { writeU32(uint32(value)) }
	writeU16(doorXferCurrentVersion4F4CB0)
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
	writeI32(direction)
	writeI32(lockCode)
	writeI32(targetDirection)

	path := filepath.Join(t.TempDir(), "door-read.bin")
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
	update, freeUpdate := alloc.New(server.DoorUpdateData{})
	defer freeUpdate()
	liveUpdate, freeLiveUpdate := alloc.New(server.DoorUpdateData{})
	defer freeLiveUpdate()
	oldID, freeOldID := alloc.CString("")
	defer freeOldID()
	update.CurrentDirection = -1
	update.TargetDirection = -2
	update.SyncedDirection = -3
	update.TileX = -4
	update.TileY = -5
	update.FractionalDir = -6
	update.LockCode = 0x77
	liveUpdate.CurrentDirection = 99
	object.ObjClass = objectlib.ClassPlayer | objectlib.ClassMissile
	object.IDPtr = unsafe.Pointer(oldID)
	object.ObjFlags = objectlib.Flags(originalFlags)
	object.Field5 = originalState
	object.Field34 = originalCount
	object.ScriptPickup.Func = -1
	object.UpdateData = unsafe.Pointer(update)

	assertObjectMapNativePointer4F4530(t, "object", unsafe.Pointer(object))
	assertObjectMapNativePointer4F4530(t, "Door update data", unsafe.Pointer(update))
	assertObjectMapNativePointer4F4530(t, "live Door update data", unsafe.Pointer(liveUpdate))
	assertObjectMapNativePointer4F4530(t, "old ID", unsafe.Pointer(oldID))

	var directionXLoads []int32
	var directionYLoads []int32
	var attachCalls []doorXferNativeAttachCall4F4CB0
	var inventoryCalls []doorXferNativeInventoryCall4F4CB0
	deps := doorXferNativeDeps4F4CB0{
		loadDirectionX: func(value int32) int32 {
			directionXLoads = append(directionXLoads, value)
			return server.DoorDirectionX(value)
		},
		loadDirectionY: func(value int32) int32 {
			directionYLoads = append(directionYLoads, value)
			return server.DoorDirectionY(value)
		},
		attachWall: func(gotObject *server.Object, tileX, tileY int32) {
			attachCalls = append(attachCalls, doorXferNativeAttachCall4F4CB0{
				object: gotObject,
				tileX:  tileX,
				tileY:  tileY,
			})
			gotObject.UpdateData = unsafe.Pointer(liveUpdate)
		},
		transferInventory: func(version uint16, gotObject *server.Object, count int32) int32 {
			inventoryCalls = append(inventoryCalls, doorXferNativeInventoryCall4F4CB0{
				version: version,
				object:  gotObject,
				count:   count,
			})
			gotObject.Field34 = 0x11223344
			return 1
		},
	}
	if got := doorXferNative4F4CB0(cf, object, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}

	if object.Field34 != originalCount {
		t.Fatalf("Field34 = %#08x, want entry value %#08x", object.Field34, originalCount)
	}
	if object.IDPtr != unsafe.Pointer(oldID) || object.UpdateData != unsafe.Pointer(liveUpdate) {
		t.Fatalf("native pointers changed unexpectedly: ID=%p update=%p", object.IDPtr, object.UpdateData)
	}
	if update.CurrentDirection != direction || update.FractionalDir != 56 ||
		update.TargetDirection != targetDirection || update.SyncedDirection != direction ||
		update.TileX != 2 || update.TileY != 0 || update.LockCode != 0xcd {
		t.Fatalf("cached Door update data = %+v, want direction 7/56 target 12 tiles 2/0 lock cd", *update)
	}
	if liveUpdate.CurrentDirection != 99 || liveUpdate.TargetDirection != 0 ||
		liveUpdate.TileX != 0 || liveUpdate.LockCode != 0 {
		t.Fatalf("live Door update data mutated after pointer replacement: %+v", *liveUpdate)
	}
	if !reflect.DeepEqual(directionXLoads, []int32{targetDirection}) ||
		!reflect.DeepEqual(directionYLoads, []int32{targetDirection}) {
		t.Fatalf("direction loads = %v/%v, want target 12", directionXLoads, directionYLoads)
	}
	if len(attachCalls) != 1 ||
		attachCalls[0].object != object ||
		attachCalls[0].tileX != 2 || attachCalls[0].tileY != 0 {
		t.Fatalf("attach calls = %#v, want native object and tiles 2/0", attachCalls)
	}
	if len(inventoryCalls) != 1 ||
		inventoryCalls[0].object != object ||
		inventoryCalls[0].version != doorXferCurrentVersion4F4CB0 ||
		inventoryCalls[0].count != int32(inventoryCount) {
		t.Fatalf("inventory calls = %#v, want native object version 60 count 3", inventoryCalls)
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
