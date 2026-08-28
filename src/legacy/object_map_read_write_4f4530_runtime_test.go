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

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestObjectMapReadWriteNativeLayout4F4530(t *testing.T) {
	type field struct {
		name string
		got  uintptr
		pe32 uintptr
		wide uintptr
	}
	fields := []field{
		{name: "size", got: unsafe.Sizeof(server.Object{}), pe32: 780, wide: 928},
		{name: "IDPtr", got: unsafe.Offsetof(server.Object{}.IDPtr), pe32: 0, wide: 0},
		{name: "TypeInd", got: unsafe.Offsetof(server.Object{}.TypeInd), pe32: 4, wide: 8},
		{name: "ObjFlags", got: unsafe.Offsetof(server.Object{}.ObjFlags), pe32: 16, wide: 20},
		{name: "Field5", got: unsafe.Offsetof(server.Object{}.Field5), pe32: 20, wide: 24},
		{name: "Extent", got: unsafe.Offsetof(server.Object{}.Extent), pe32: 40, wide: 44},
		{name: "ScriptIDVal", got: unsafe.Offsetof(server.Object{}.ScriptIDVal), pe32: 44, wide: 48},
		{name: "TeamVal.ID", got: unsafe.Offsetof(server.Object{}.TeamVal) + unsafe.Offsetof(server.ObjectTeam{}.ID), pe32: 52, wide: 56},
		{name: "PosVec", got: unsafe.Offsetof(server.Object{}.PosVec), pe32: 56, wide: 60},
		{name: "NewPos", got: unsafe.Offsetof(server.Object{}.NewPos), pe32: 64, wide: 68},
		{name: "Field32", got: unsafe.Offsetof(server.Object{}.Field32), pe32: 128, wide: 132},
		{name: "Field34", got: unsafe.Offsetof(server.Object{}.Field34), pe32: 136, wide: 140},
		{name: "InvNextItem", got: unsafe.Offsetof(server.Object{}.InvNextItem), pe32: 496, wide: 528},
		{name: "InvFirstItem", got: unsafe.Offsetof(server.Object{}.InvFirstItem), pe32: 504, wide: 544},
		{name: "Field128", got: unsafe.Offsetof(server.Object{}.Field128), pe32: 512, wide: 560},
		{name: "Field129", got: unsafe.Offsetof(server.Object{}.Field129), pe32: 516, wide: 568},
		{name: "Field189", got: unsafe.Offsetof(server.Object{}.Field189), pe32: 756, wide: 888},
		{name: "ScriptPickup", got: unsafe.Offsetof(server.Object{}.ScriptPickup), pe32: 764, wide: 904},
		{name: "ScriptPickup.Func", got: unsafe.Offsetof(server.Object{}.ScriptPickup) + unsafe.Offsetof(server.ScriptCallback{}.Func), pe32: 768, wide: 908},
	}

	wide := unsafe.Sizeof(uintptr(0)) == 8
	for _, field := range fields {
		want := field.pe32
		if wide {
			want = field.wide
		}
		if field.got != want {
			t.Errorf("Object.%s native layout = %d, want %d", field.name, field.got, want)
		}
	}
	if got := unsafe.Sizeof(server.ScriptCallback{}); got != 8 {
		t.Errorf("ScriptCallback native size = %d, want 8", got)
	}
}

type objectMapLegacyServer4F4530 struct {
	Server
	srv *server.Server
}

func (s *objectMapLegacyServer4F4530) S() *server.Server {
	return s.srv
}

func setObjectMapRuntimeGlobals4F4530(t *testing.T, cf *cryptfile.CryptFile, frame uint32) {
	t.Helper()
	oldFile := cryptfile.Global()
	oldFrame := gameFrameHook
	oldFlags := noxflags.GetGame()
	oldGetServer := GetServer
	testServer := &objectMapLegacyServer4F4530{srv: new(server.Server)}
	cryptfile.SetGlobal(cf)
	gameFrameHook = func() uint32 { return frame }
	GetServer = func() Server {
		return testServer
	}
	noxflags.ResetGame()
	t.Cleanup(func() {
		cryptfile.SetGlobal(oldFile)
		gameFrameHook = oldFrame
		GetServer = oldGetServer
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})
}

func assertObjectMapNativePointer4F4530(t *testing.T, name string, pointer unsafe.Pointer) {
	t.Helper()
	if pointer == nil {
		t.Fatalf("%s pointer is nil", name)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
		t.Fatalf("%s pointer = %p, want native address above PE32 range", name, pointer)
	}
}

func TestObjectMapReadWriteNativeWriteV64_4F4530PreservesPointersAndWire(t *testing.T) {
	obj, freeObj := alloc.New(server.Object{})
	defer freeObj()
	first, freeFirst := alloc.New(server.Object{})
	defer freeFirst()
	second, freeSecond := alloc.New(server.Object{})
	defer freeSecond()
	id, freeID := alloc.CString("modern-id")
	defer freeID()

	assertObjectMapNativePointer4F4530(t, "object", unsafe.Pointer(obj))
	assertObjectMapNativePointer4F4530(t, "first inventory object", unsafe.Pointer(first))
	assertObjectMapNativePointer4F4530(t, "second inventory object", unsafe.Pointer(second))
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
	)
	obj.ObjClass = objectlib.ClassPlayer | objectlib.ClassMissile
	obj.Extent = extent
	obj.ScriptIDVal = scriptID
	obj.PosVec.X = positionX
	obj.PosVec.Y = positionY
	obj.ObjFlags = objectlib.Flags(flags)
	obj.IDPtr = unsafe.Pointer(id)
	obj.TeamVal.ID = server.TeamID(7)
	obj.InvFirstItem = first
	first.InvNextItem = second
	obj.Field5 = status
	obj.ScriptPickup = server.ScriptCallback{Flags: handlerFlags, Func: -1}
	obj.Field34 = objectFrame

	path := filepath.Join(t.TempDir(), "object-modern-v64.bin")
	cf, err := cryptfile.OpenFile(path, cryptfile.WriteOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	setObjectMapRuntimeGlobals4F4530(t, cf, gameFrame)
	if got := objectMapReadWriteNative4F4530(cf, obj, 40); got != 1 {
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
	writeU16(objectMapCurrentVersion4F4530)
	writeU32(extent)
	writeI32(scriptID)
	writeU32(math.Float32bits(positionX))
	writeU32(math.Float32bits(positionY))
	want.WriteByte(math.MaxUint8)
	writeU32(flags & objectMapFlagsMask4F4530)
	want.WriteByte(uint8(len("modern-id")))
	want.WriteString("modern-id")
	want.WriteByte(7)
	want.WriteByte(2)
	writeU16(0)
	writeU32(status & objectMapStatusMask4F4530)
	writeU16(1)
	writeU32(0)
	writeU32(handlerFlags)
	writeI32(int32(objectFrame - gameFrame))

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want.Bytes()) {
		t.Fatalf("wire payload = %x, want %x", got, want.Bytes())
	}
	if uint32(obj.ObjFlags) != flags {
		t.Errorf("post-write flags = %#08x, want %#08x", uint32(obj.ObjFlags), flags)
	}
	if obj.Field5 != status {
		t.Errorf("post-write status = %#08x, want %#08x", obj.Field5, status)
	}
	if obj.Field34 != objectFrame {
		t.Errorf("post-write Field34 = %#08x, want %#08x", obj.Field34, objectFrame)
	}
}

func TestObjectMapReadWriteNativeReadV64_4F4530RestoresNativeState(t *testing.T) {
	const (
		extent        = uint32(0x55667788)
		scriptID      = int32(0x10203040)
		positionX     = float32(-321.25)
		positionY     = float32(654.5)
		serialized    = uint32(0x01400102)
		teamID        = uint8(9)
		inventory     = uint8(17)
		status        = uint32(0x12)
		handlerFlags  = uint32(0x55667788)
		frameDelta    = int32(0x01020304)
		originalFlags = uint32(0x80000040)
		originalState = uint32(0xa5)
	)

	var payload bytes.Buffer
	writeU16 := func(value uint16) { _ = binary.Write(&payload, binary.LittleEndian, value) }
	writeU32 := func(value uint32) { _ = binary.Write(&payload, binary.LittleEndian, value) }
	writeI32 := func(value int32) { writeU32(uint32(value)) }
	writeU16(objectMapCurrentVersion4F4530)
	writeU32(extent)
	writeI32(scriptID)
	writeU32(math.Float32bits(positionX))
	writeU32(math.Float32bits(positionY))
	payload.WriteByte(math.MaxUint8)
	writeU32(serialized)
	payload.WriteByte(uint8(len("modern-read")))
	payload.WriteString("modern-read")
	payload.WriteByte(teamID)
	payload.WriteByte(inventory)
	writeU16(0)
	writeU32(status)
	writeU16(1)
	writeU32(0)
	writeU32(handlerFlags)
	writeI32(frameDelta)

	path := filepath.Join(t.TempDir(), "object-modern-v64-read.bin")
	if err := os.WriteFile(path, payload.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	cf, err := cryptfile.OpenFile(path, cryptfile.ReadOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cf.Close() }()
	setObjectMapRuntimeGlobals4F4530(t, cf, 0x89abcdef)

	obj, freeObj := alloc.New(server.Object{})
	defer freeObj()
	oldID, freeOldID := alloc.CString("")
	defer freeOldID()
	obj.ObjClass = objectlib.ClassPlayer | objectlib.ClassMissile
	obj.IDPtr = unsafe.Pointer(oldID) // Forces the original admission result to -1 before wire input.
	obj.ObjFlags = objectlib.Flags(originalFlags)
	obj.Field5 = originalState
	obj.Field34 = math.MaxUint32
	obj.ScriptPickup.Func = -1

	assertObjectMapNativePointer4F4530(t, "object", unsafe.Pointer(obj))
	assertObjectMapNativePointer4F4530(t, "old ID", unsafe.Pointer(oldID))
	if got := objectMapReadWriteNative4F4530(cf, obj, 40); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if obj.IDPtr == unsafe.Pointer(oldID) {
		t.Fatal("read retained the old ID pointer")
	}
	assertObjectMapNativePointer4F4530(t, "allocated ID", obj.IDPtr)
	defer alloc.FreePtr(obj.IDPtr)

	if obj.Extent != extent {
		t.Errorf("extent = %#08x, want %#08x", obj.Extent, extent)
	}
	if obj.ScriptIDVal != scriptID {
		t.Errorf("script ID = %#08x, want %#08x", uint32(obj.ScriptIDVal), uint32(scriptID))
	}
	if obj.PosVec.X != positionX || obj.PosVec.Y != positionY {
		t.Errorf("position = (%v, %v), want (%v, %v)", obj.PosVec.X, obj.PosVec.Y, positionX, positionY)
	}
	if obj.NewPos != obj.PosVec {
		t.Errorf("new position = %v, want %v", obj.NewPos, obj.PosVec)
	}
	if got := alloc.GoString((*byte)(obj.IDPtr)); got != "modern-read" {
		t.Errorf("ID = %q, want %q", got, "modern-read")
	}
	if uint8(obj.TeamVal.ID) != teamID {
		t.Errorf("team ID = %d, want %d", obj.TeamVal.ID, teamID)
	}
	if obj.Field34 != uint32(inventory) {
		t.Errorf("inventory count = %d, want %d", obj.Field34, inventory)
	}
	wantFlags := originalFlags&objectMapFlagsKeepMask4F4530 |
		objectMapPreserveFlag4F4530 | serialized
	if uint32(obj.ObjFlags) != wantFlags {
		t.Errorf("flags = %#08x, want %#08x", uint32(obj.ObjFlags), wantFlags)
	}
	wantStatus := originalState&^objectMapStatusMask4F4530 | status
	if obj.Field5 != wantStatus {
		t.Errorf("status = %#08x, want %#08x", obj.Field5, wantStatus)
	}
	if obj.ScriptPickup.Flags != handlerFlags || obj.ScriptPickup.Func != -1 {
		t.Errorf("script callback = {%#08x, %d}, want {%#08x, -1}", obj.ScriptPickup.Flags, obj.ScriptPickup.Func, handlerFlags)
	}
	if obj.Field32 != uint32(frameDelta) {
		t.Errorf("decay frame = %#08x, want %#08x", obj.Field32, uint32(frameDelta))
	}
}
