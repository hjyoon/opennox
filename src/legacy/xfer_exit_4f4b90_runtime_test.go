package legacy

import (
	"bytes"
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	objectlib "github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestExitXferNativeLayout4F4B90(t *testing.T) {
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
	if got := unsafe.Sizeof(server.ExitCollideData{}); got != 88 {
		t.Errorf("Exit collide-data size = %d, want 88", got)
	}
	if got := unsafe.Offsetof(server.ExitCollideData{}.MapName); got != 0 {
		t.Errorf("Exit map-name offset = %d, want 0", got)
	}
	if got := unsafe.Offsetof(server.ExitCollideData{}.DestinationX); got != 80 {
		t.Errorf("Exit destination-X offset = %d, want 80", got)
	}
	if got := unsafe.Offsetof(server.ExitCollideData{}.DestinationY); got != 84 {
		t.Errorf("Exit destination-Y offset = %d, want 84", got)
	}
}

func TestExitXferNativeWrite4F4B90PreservesPointersAndUnboundedWire(t *testing.T) {
	object, freeObject := alloc.New(server.Object{})
	defer freeObject()
	rawData, freeData := alloc.Calloc(1, 96)
	defer freeData()
	dataBytes := unsafe.Slice((*byte)(rawData), 96)
	for i := 0; i < 90; i++ {
		dataBytes[i] = 'A'
	}
	dataBytes[90] = 0
	id, freeID := alloc.CString("exit")
	defer freeID()

	assertObjectMapNativePointer4F4530(t, "object", unsafe.Pointer(object))
	assertObjectMapNativePointer4F4530(t, "Exit collide data", rawData)
	assertObjectMapNativePointer4F4530(t, "ID", unsafe.Pointer(id))

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
		mapNameSize  = uint32(91)
	)
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
	object.CollideData = rawData

	path := filepath.Join(t.TempDir(), "exit-write.bin")
	cf, err := cryptfile.OpenFile(path, cryptfile.WriteOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	setObjectMapRuntimeGlobals4F4530(t, cf, gameFrame)
	if got := Nox_xxx_XFerExitNative4F4B90(cf, object); got != 1 {
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
	writeU16(exitXferCurrentVersion4F4B90)
	writeU16(objectMapCurrentVersion4F4530)
	writeU32(extent)
	writeI32(scriptID)
	writeU32(math.Float32bits(positionX))
	writeU32(math.Float32bits(positionY))
	want.WriteByte(math.MaxUint8)
	writeU32(flags & objectMapFlagsMask4F4530)
	want.WriteByte(uint8(len("exit")))
	want.WriteString("exit")
	want.WriteByte(7)
	want.WriteByte(0)
	writeU16(0)
	writeU32(status & objectMapStatusMask4F4530)
	writeU16(1)
	writeU32(0)
	writeU32(handlerFlags)
	writeI32(int32(objectFrame - gameFrame))
	writeU32(mapNameSize)
	want.Write(dataBytes[:mapNameSize])
	want.Write(dataBytes[80:88])

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
	if object.IDPtr != unsafe.Pointer(id) || object.CollideData != rawData {
		t.Errorf("native pointers changed: ID=%p collide=%p", object.IDPtr, object.CollideData)
	}
}

func TestExitXferNativeRead4F4B90RestoresEntryState(t *testing.T) {
	const (
		extent        = uint32(0x55667788)
		scriptID      = int32(0x10203040)
		positionX     = float32(-321.25)
		positionY     = float32(654.5)
		serialized    = uint32(0x01400102)
		status        = uint32(0x12)
		handlerFlags  = uint32(0x55667788)
		frameDelta    = int32(0x01020304)
		originalFlags = uint32(0x80000040)
		originalState = uint32(0xa5)
		originalCount = uint32(0xfedcba98)
		destinationX  = float32(15.5)
		destinationY  = float32(-22.25)
	)
	const mapName = "War01a"

	var payload bytes.Buffer
	writeU16 := func(value uint16) { _ = binary.Write(&payload, binary.LittleEndian, value) }
	writeU32 := func(value uint32) { _ = binary.Write(&payload, binary.LittleEndian, value) }
	writeI32 := func(value int32) { writeU32(uint32(value)) }
	writeU16(exitXferCurrentVersion4F4B90)
	writeU16(objectMapCurrentVersion4F4530)
	writeU32(extent)
	writeI32(scriptID)
	writeU32(math.Float32bits(positionX))
	writeU32(math.Float32bits(positionY))
	payload.WriteByte(math.MaxUint8)
	writeU32(serialized)
	payload.WriteByte(0)
	payload.WriteByte(9)
	payload.WriteByte(0)
	writeU16(0)
	writeU32(status)
	writeU16(1)
	writeU32(0)
	writeU32(handlerFlags)
	writeI32(frameDelta)
	writeU32(uint32(len(mapName) + 1))
	payload.WriteString(mapName)
	payload.WriteByte(0)
	writeU32(math.Float32bits(destinationX))
	writeU32(math.Float32bits(destinationY))

	path := filepath.Join(t.TempDir(), "exit-read.bin")
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
	data, freeData := alloc.New(server.ExitCollideData{})
	defer freeData()
	oldID, freeOldID := alloc.CString("")
	defer freeOldID()
	copy(data.MapName[:], "entry")
	data.DestinationX = -1
	data.DestinationY = -2
	object.ObjClass = objectlib.ClassPlayer | objectlib.ClassMissile
	object.IDPtr = unsafe.Pointer(oldID)
	object.ObjFlags = objectlib.Flags(originalFlags)
	object.Field5 = originalState
	object.Field34 = originalCount
	object.ScriptPickup.Func = -1
	object.CollideData = unsafe.Pointer(data)

	assertObjectMapNativePointer4F4530(t, "object", unsafe.Pointer(object))
	assertObjectMapNativePointer4F4530(t, "Exit collide data", unsafe.Pointer(data))
	assertObjectMapNativePointer4F4530(t, "old ID", unsafe.Pointer(oldID))
	if got := Nox_xxx_XFerExitNative4F4B90(cf, object); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if object.Field34 != originalCount {
		t.Fatalf("Field34 = %#08x, want entry value %#08x", object.Field34, originalCount)
	}
	if object.IDPtr != unsafe.Pointer(oldID) || object.CollideData != unsafe.Pointer(data) {
		t.Fatalf("native pointers changed: ID=%p collide=%p", object.IDPtr, object.CollideData)
	}
	if got := data.MapName[:len(mapName)+1]; !bytes.Equal(got, append([]byte(mapName), 0)) {
		t.Fatalf("Exit map name = %q, want %q", got, mapName)
	}
	if data.DestinationX != destinationX || data.DestinationY != destinationY {
		t.Fatalf("Exit destination = (%v,%v), want (%v,%v)",
			data.DestinationX, data.DestinationY, destinationX, destinationY)
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
