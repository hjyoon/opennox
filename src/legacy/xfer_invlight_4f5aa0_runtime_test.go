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

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type invLightXferNativeInventoryCall4F5AA0 struct {
	version uint16
	object  *server.Object
	count   int32
}

func TestInvLightXferNativeLayout4F5AA0(t *testing.T) {
	type field struct {
		name string
		got  uintptr
		pe32 uintptr
		wide uintptr
	}
	fields := []field{
		{name: "Object size", got: unsafe.Sizeof(server.Object{}), pe32: 780, wide: 928},
		{name: "Object.Field34", got: unsafe.Offsetof(server.Object{}.Field34), pe32: 136, wide: 140},
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
	if got := unsafe.Sizeof(uintptr(0)); got != 4 && got != 8 {
		t.Errorf("native pointer size = %d, want 4 or 8", got)
	}
	if client.DrawableLightXferSize != invLightXferPayloadSize4F5AA0 {
		t.Errorf("drawable light wire size = %d, want %d", client.DrawableLightXferSize, invLightXferPayloadSize4F5AA0)
	}
}

func TestInvLightXferNativeWrite4F5AA0UsesNativeDrawableAndExactWire(t *testing.T) {
	object, freeObject := alloc.New(server.Object{})
	defer freeObject()
	id, freeID := alloc.CString("invisible-light")
	defer freeID()
	drawable := &client.Drawable{
		LightFlags:        0xa1b2c3d4,
		LightIntensity:    17.25,
		LightIntensityRad: 0x11223344,
		LightIntensityU16: 0x55667788,
		LightDir:          0x99aa,
		LightPenumbra:     0xbbcc,
		Field_42:          0xddeeff01,
		Field_43:          0x23456789,
		Field_44:          0xabcdef01,
		Field_65:          0x10203040,
		Field_66:          0x50607080,
		Field_67:          0x90a0b0c0,
		Field_68:          0xd0e0f001,
	}
	drawable.LightColor.R = -17
	drawable.LightColor.G = 257
	drawable.LightColor.B = -65537

	for name, pointer := range map[string]unsafe.Pointer{
		"object":   unsafe.Pointer(object),
		"ID":       unsafe.Pointer(id),
		"drawable": unsafe.Pointer(drawable),
	} {
		assertObjectMapNativePointer4F4530(t, name, pointer)
	}

	const (
		extent       = uint32(0x11223344)
		netCode      = uint32(0x55667788)
		scriptID     = int32(-0x1020304)
		positionX    = float32(123.25)
		positionY    = float32(-456.5)
		flags        = uint32(0x91408162)
		status       = uint32(0xa5)
		handlerFlags = uint32(0xa1b2c3d4)
		objectFrame  = uint32(0x11223344)
		gameFrame    = uint32(0x01020304)
	)
	object.ObjClass = objectlib.ClassPlayer | objectlib.ClassMissile
	object.Extent = extent
	object.NetCode = netCode
	object.ScriptIDVal = scriptID
	object.PosVec.X = positionX
	object.PosVec.Y = positionY
	object.ObjFlags = objectlib.Flags(flags)
	object.IDPtr = unsafe.Pointer(id)
	object.TeamVal.ID = server.TeamID(7)
	object.Field5 = status
	object.ScriptPickup = server.ScriptCallback{Flags: handlerFlags, Func: -1}
	object.Field34 = objectFrame

	path := filepath.Join(t.TempDir(), "invlight-write.bin")
	cf, err := cryptfile.OpenFile(path, cryptfile.WriteOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	setObjectMapRuntimeGlobals4F4530(t, cf, gameFrame)

	gameFlagCalls := 0
	deps := invLightXferNativeDeps4F5AA0{
		gameFlags: func(mask uint32) int32 {
			gameFlagCalls++
			if mask != invLightXferGameMask4F5AA0 {
				t.Fatalf("game flag mask = %#x", mask)
			}
			return 0
		},
		firstDrawable: func() *client.Drawable {
			t.Fatal("non-preview write walked List1")
			return nil
		},
		staticDrawable: func(uint32) *client.Drawable {
			t.Fatal("dynamic object used static drawable lookup")
			return nil
		},
		dynamicDrawable: func(code uint32) *client.Drawable {
			if code != netCode {
				t.Fatalf("dynamic code = %#x, want %#x", code, netCode)
			}
			return drawable
		},
		transferInventory: func(uint16, *server.Object, int32) int32 {
			t.Fatal("write mode transferred inventory")
			return 0
		},
	}
	if got := invLightXferNative4F5AA0(cf, object, deps); got != 1 {
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
	writeU16(invLightXferCurrentVersion4F5AA0)
	writeU16(objectMapCurrentVersion4F4530)
	writeU32(extent)
	writeI32(scriptID)
	writeU32(math.Float32bits(positionX))
	writeU32(math.Float32bits(positionY))
	want.WriteByte(math.MaxUint8)
	writeU32(flags & objectMapFlagsMask4F4530)
	want.WriteByte(uint8(len("invisible-light")))
	want.WriteString("invisible-light")
	want.WriteByte(7)
	want.WriteByte(0)
	writeU16(0)
	writeU32(status & objectMapStatusMask4F4530)
	writeU16(1)
	writeU32(0)
	writeU32(handlerFlags)
	writeI32(int32(objectFrame - gameFrame))
	light := drawable.LightXferData()
	want.Write(invLightXferExpectedWire4F5AA0(light, 60))

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
	if object.Field34 != objectFrame {
		t.Errorf("Field34 = %#x, want %#x", object.Field34, objectFrame)
	}
	runtime.KeepAlive(drawable)
}

func TestInvLightXferNativeRead4F5AA0AppliesToWideField189AndRestoresCount(t *testing.T) {
	const (
		extent         = uint32(0x55667788)
		scriptID       = int32(0x10203040)
		positionX      = float32(-321.25)
		positionY      = float32(654.5)
		serialized     = uint32(0x01400102)
		status         = uint32(0x12)
		handlerFlags   = uint32(0x55667788)
		frameDelta     = int32(0x01020304)
		originalFlags  = uint32(0x80000040)
		originalState  = uint32(0xa5)
		originalCount  = uint32(0xfedcba98)
		inventoryCount = uint8(3)
	)

	var wantLight [invLightXferPayloadSize4F5AA0]byte
	for _, part := range invLightXferParts4F5AA0 {
		for index := part.offset; index < part.offset+part.size; index++ {
			wantLight[index] = byte(index*3 + 1)
		}
	}
	for index := 36; index < 40; index++ {
		wantLight[index] = byte(index*3 + 1)
	}

	var payload bytes.Buffer
	writeU16 := func(value uint16) { _ = binary.Write(&payload, binary.LittleEndian, value) }
	writeU32 := func(value uint32) { _ = binary.Write(&payload, binary.LittleEndian, value) }
	writeI32 := func(value int32) { writeU32(uint32(value)) }
	writeU16(invLightXferCurrentVersion4F5AA0)
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
	payload.Write(invLightXferExpectedWire4F5AA0(wantLight, 60))

	path := filepath.Join(t.TempDir(), "invlight-read.bin")
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
	oldID, freeOldID := alloc.CString("")
	defer freeOldID()
	scriptData, freeScriptData := alloc.Make([]byte(nil), 2572)
	defer freeScriptData()
	for index := range scriptData {
		scriptData[index] = 0xcc
	}
	object.ObjClass = objectlib.ClassPlayer | objectlib.ClassMissile
	object.IDPtr = unsafe.Pointer(oldID)
	object.ObjFlags = objectlib.Flags(originalFlags)
	object.Field5 = originalState
	object.Field34 = originalCount
	object.Field189 = unsafe.Pointer(&scriptData[0])
	object.ScriptPickup.Func = -1

	for name, pointer := range map[string]unsafe.Pointer{
		"object":      unsafe.Pointer(object),
		"old ID":      unsafe.Pointer(oldID),
		"Field189":    object.Field189,
		"light block": unsafe.Add(object.Field189, 2432),
	} {
		assertObjectMapNativePointer4F4530(t, name, pointer)
	}

	var inventoryCalls []invLightXferNativeInventoryCall4F5AA0
	gameFlagCalls := 0
	deps := invLightXferNativeDeps4F5AA0{
		gameFlags: func(mask uint32) int32 {
			gameFlagCalls++
			if mask != invLightXferGameMask4F5AA0 {
				t.Fatalf("game flag mask = %#x", mask)
			}
			return 1
		},
		firstDrawable: func() *client.Drawable {
			t.Fatal("read mode walked List1")
			return nil
		},
		staticDrawable: func(uint32) *client.Drawable {
			t.Fatal("read mode resolved a static drawable")
			return nil
		},
		dynamicDrawable: func(uint32) *client.Drawable {
			t.Fatal("read mode resolved a dynamic drawable")
			return nil
		},
		transferInventory: func(version uint16, gotObject *server.Object, count int32) int32 {
			inventoryCalls = append(inventoryCalls, invLightXferNativeInventoryCall4F5AA0{
				version: version,
				object:  gotObject,
				count:   count,
			})
			object.Field34 = 0x44
			return 1
		},
	}
	if got := invLightXferNative4F5AA0(cf, object, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if object.IDPtr != unsafe.Pointer(oldID) {
		defer alloc.FreePtr(object.IDPtr)
	}

	gotLight := unsafe.Slice((*byte)(unsafe.Add(object.Field189, 2432)), len(wantLight))
	if !bytes.Equal(gotLight, wantLight[:]) {
		t.Fatalf("applied light = %x, want %x", gotLight, wantLight)
	}
	if gameFlagCalls != 1 {
		t.Errorf("game flag calls = %d, want 1", gameFlagCalls)
	}
	if len(inventoryCalls) != 1 || inventoryCalls[0] != (invLightXferNativeInventoryCall4F5AA0{
		version: 60,
		object:  object,
		count:   int32(inventoryCount),
	}) {
		t.Fatalf("inventory calls = %+v", inventoryCalls)
	}
	if object.Field34 != originalCount {
		t.Errorf("Field34 = %#x, want original %#x", object.Field34, originalCount)
	}
	if object.Field189 != unsafe.Pointer(&scriptData[0]) {
		t.Errorf("Field189 changed = %p, want %p", object.Field189, &scriptData[0])
	}
}

func TestInvLightXferExport4F5AA0PreservesNativePointerAndResult(t *testing.T) {
	object, freeObject := alloc.New(server.Object{})
	defer freeObject()
	assertObjectMapNativePointer4F4530(t, "object", unsafe.Pointer(object))

	old := invLightXferCall4F5AA0
	t.Cleanup(func() { invLightXferCall4F5AA0 = old })
	calls := 0
	invLightXferCall4F5AA0 = func(_ *cryptfile.CryptFile, gotObject *server.Object) int32 {
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

	if got := invLightXferExportCall4F5AA0(object); got != math.MinInt32 {
		t.Fatalf("first result = %d, want %d", got, int32(math.MinInt32))
	}
	if got := invLightXferExportCall4F5AA0(nil); got != math.MaxInt32 {
		t.Fatalf("second result = %d, want %d", got, int32(math.MaxInt32))
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
	runtime.KeepAlive(object)
}
