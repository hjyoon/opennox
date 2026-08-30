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

func TestObjectReadOldNativeLayout4F4170(t *testing.T) {
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
		{name: "Field34", got: unsafe.Offsetof(server.Object{}.Field34), pe32: 136, wide: 140},
		{name: "InvNextItem", got: unsafe.Offsetof(server.Object{}.InvNextItem), pe32: 496, wide: 528},
		{name: "InvFirstItem", got: unsafe.Offsetof(server.Object{}.InvFirstItem), pe32: 504, wide: 544},
		{name: "Field128", got: unsafe.Offsetof(server.Object{}.Field128), pe32: 512, wide: 560},
		{name: "Field129", got: unsafe.Offsetof(server.Object{}.Field129), pe32: 516, wide: 568},
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
}

func TestObjectReadOldNativeWrite4F4170PreservesHighPointersAndWireWidths(t *testing.T) {
	obj := new(server.Object)
	first := new(server.Object)
	second := new(server.Object)
	id := append([]byte("legacy-id"), 0)
	idPtr := unsafe.Pointer(&id[0])
	var pin runtime.Pinner
	pin.Pin(obj)
	pin.Pin(first)
	pin.Pin(second)
	pin.Pin(&id[0])
	defer pin.Unpin()

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]unsafe.Pointer{
			"object": unsafe.Pointer(obj),
			"first":  unsafe.Pointer(first),
			"second": unsafe.Pointer(second),
			"ID":     idPtr,
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("%s pointer = %p, want native address above PE32 range", name, pointer)
			}
		}
	}

	const (
		extent    = uint32(0x11223344)
		flags     = uint32(0xfbadcafe)
		positionX = float32(123.25)
		positionY = float32(-456.5)
		scriptID  = int32(-0x1020304)
		status    = uint32(0x00000004)
	)
	obj.Extent = extent
	obj.ObjClass = objectlib.ClassPlayer
	obj.ObjFlags = objectlib.Flags(flags)
	obj.Field5 = 0xa5
	obj.PosVec.X = positionX
	obj.PosVec.Y = positionY
	obj.IDPtr = idPtr
	obj.TeamVal.ID = server.TeamID(7)
	obj.ScriptIDVal = scriptID
	obj.InvFirstItem = first
	first.InvNextItem = second

	path := filepath.Join(t.TempDir(), "object-old-v5.bin")
	cf, err := cryptfile.OpenFile(path, cryptfile.WriteOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	if got := objectReadOldNative4F4170(cf, obj, 5, 40); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if err := cf.Close(); err != nil {
		t.Fatal(err)
	}

	var want bytes.Buffer
	writeU32 := func(value uint32) { _ = binary.Write(&want, binary.LittleEndian, value) }
	writeI32 := func(value int32) { writeU32(uint32(value)) }
	writeU16 := func(value uint16) { _ = binary.Write(&want, binary.LittleEndian, value) }
	writeU32(extent)
	writeU32(flags & objectReadOldFlagsMask4F4170)
	writeU32(math.Float32bits(positionX))
	writeU32(math.Float32bits(positionY))
	want.WriteByte(uint8(len("legacy-id")))
	want.WriteString("legacy-id")
	want.WriteByte(7)
	want.WriteByte(2)
	writeI32(scriptID)
	writeU16(0)
	writeU32(status)

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want.Bytes()) {
		t.Fatalf("wire payload = %x, want %x", got, want.Bytes())
	}
	if uint32(obj.ObjFlags) != flags {
		t.Fatalf("post-write flags = %#08x, want %#08x", uint32(obj.ObjFlags), flags)
	}
	if obj.Field5 != 0xa5 {
		t.Fatalf("post-write status = %#08x, want %#08x", obj.Field5, uint32(0xa5))
	}
	runtime.KeepAlive(id)
}

func TestObjectReadOldNativeRead4F4170RestoresOldPositionAndNativeID(t *testing.T) {
	const (
		extent       = uint32(0x55667788)
		serialized   = uint32(0x00000102)
		positionX    = int32(-321)
		positionY    = int32(654)
		teamID       = uint8(9)
		inventory    = uint8(17)
		scriptID     = int32(0x10203040)
		status       = uint32(0x00000012)
		originalFlag = uint32(0x80000040)
	)

	var payload bytes.Buffer
	writeU32 := func(value uint32) { _ = binary.Write(&payload, binary.LittleEndian, value) }
	writeI32 := func(value int32) { writeU32(uint32(value)) }
	writeU32(extent)
	writeU32(serialized)
	writeI32(positionX)
	writeI32(positionY)
	payload.WriteByte(uint8(len("old-object")))
	payload.WriteString("old-object")
	payload.WriteByte(teamID)
	payload.WriteByte(inventory)
	writeU32(uint32(scriptID))
	writeU32(0) // Object versions below five store the owned count as a dword.
	writeU32(status)

	path := filepath.Join(t.TempDir(), "object-old-v3.bin")
	if err := os.WriteFile(path, payload.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	cf, err := cryptfile.OpenFile(path, cryptfile.ReadOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cf.Close() }()

	obj, freeObj := alloc.New(server.Object{})
	defer freeObj()
	obj.ObjClass = objectlib.ClassPlayer
	obj.ObjFlags = objectlib.Flags(originalFlag)
	obj.Field5 = 0xa5
	obj.Field34 = math.MaxUint32

	if got := objectReadOldNative4F4170(cf, obj, 3, 40); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if obj.IDPtr == nil {
		t.Fatal("read did not allocate the object ID")
	}
	defer alloc.FreePtr(obj.IDPtr)

	if obj.Extent != extent {
		t.Errorf("extent = %#08x, want %#08x", obj.Extent, extent)
	}
	wantFlags := originalFlag&objectReadOldFlagsKeepMask4F4170 |
		objectReadOldPreserveFlag4F4170 | serialized
	if uint32(obj.ObjFlags) != wantFlags {
		t.Errorf("flags = %#08x, want %#08x", uint32(obj.ObjFlags), wantFlags)
	}
	if obj.PosVec.X != float32(positionX) || obj.PosVec.Y != float32(positionY) {
		t.Errorf("position = (%v, %v), want (%v, %v)", obj.PosVec.X, obj.PosVec.Y, float32(positionX), float32(positionY))
	}
	if obj.NewPos != obj.PosVec {
		t.Errorf("new position = %v, want %v", obj.NewPos, obj.PosVec)
	}
	if got := alloc.GoString((*byte)(obj.IDPtr)); got != "old-object" {
		t.Errorf("ID = %q, want %q", got, "old-object")
	}
	if uint8(obj.TeamVal.ID) != teamID {
		t.Errorf("team ID = %d, want %d", obj.TeamVal.ID, teamID)
	}
	if obj.Field34 != uint32(inventory) {
		t.Errorf("inventory count = %d, want %d", obj.Field34, inventory)
	}
	if obj.ScriptIDVal != scriptID {
		t.Errorf("script ID = %#08x, want %#08x", uint32(obj.ScriptIDVal), uint32(scriptID))
	}
	wantStatus := uint32(0xa5)&^objectReadOldStatusMask4F4170 | status
	if obj.Field5 != wantStatus {
		t.Errorf("status = %#08x, want %#08x", obj.Field5, wantStatus)
	}
}
