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
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type glyphXferNativeInventoryCall4F5890 struct {
	version uint16
	object  *server.Object
	count   int32
}

func TestGlyphXferNativeLayout4F5890(t *testing.T) {
	type field struct {
		name string
		got  uintptr
		pe32 uintptr
		wide uintptr
	}
	fields := []field{
		{name: "Object size", got: unsafe.Sizeof(server.Object{}), pe32: 780, wide: 928},
		{name: "Object.Direction1", got: unsafe.Offsetof(server.Object{}.Direction1), pe32: 124, wide: 128},
		{name: "Object.Direction2", got: unsafe.Offsetof(server.Object{}.Direction2), pe32: 126, wide: 130},
		{name: "Object.Field34", got: unsafe.Offsetof(server.Object{}.Field34), pe32: 136, wide: 140},
		{name: "Object.InitData", got: unsafe.Offsetof(server.Object{}.InitData), pe32: 692, wide: 760},
		{name: "GlyphInitData size", got: unsafe.Sizeof(server.GlyphInitData{}), pe32: 36, wide: 40},
		{name: "GlyphInitData point", got: unsafe.Offsetof(server.GlyphInitData{}.SpellArg) + unsafe.Offsetof(server.SpellAcceptArg{}.Pos), pe32: 28, wide: 32},
		{name: "GlyphInitData target", got: unsafe.Offsetof(server.GlyphInitData{}.SpellArg) + unsafe.Offsetof(server.SpellAcceptArg{}.Obj), pe32: 24, wide: 24},
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
		{name: "GlyphInitData.Spells", got: unsafe.Offsetof(server.GlyphInitData{}.Spells), want: 0},
		{name: "GlyphInitData.SpellsCnt", got: unsafe.Offsetof(server.GlyphInitData{}.SpellsCnt), want: 20},
		{name: "GlyphInitData.SpellArg", got: unsafe.Offsetof(server.GlyphInitData{}.SpellArg), want: 24},
		{name: "Pointf size", got: unsafe.Sizeof(types.Pointf{}), want: 8},
		{name: "Pointf.X", got: unsafe.Offsetof(types.Pointf{}.X), want: 0},
		{name: "Pointf.Y", got: unsafe.Offsetof(types.Pointf{}.Y), want: 4},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestGlyphXferNativeWrite4F5890UsesCachedDataAndExactWire(t *testing.T) {
	object, freeObject := alloc.New(server.Object{})
	defer freeObject()
	data, freeData := alloc.New(server.GlyphInitData{})
	defer freeData()
	liveData, freeLiveData := alloc.New(server.GlyphInitData{})
	defer freeLiveData()
	target, freeTarget := alloc.New(server.Object{})
	defer freeTarget()
	id, freeID := alloc.CString("glyph")
	defer freeID()

	for name, pointer := range map[string]unsafe.Pointer{
		"object":          unsafe.Pointer(object),
		"Glyph init data": unsafe.Pointer(data),
		"live Glyph data": unsafe.Pointer(liveData),
		"target object":   unsafe.Pointer(target),
		"ID":              unsafe.Pointer(id),
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
		targetX      = float32(77.25)
		targetY      = float32(-88.5)
	)
	*data = server.GlyphInitData{
		Spells:    [5]uint32{10, 11, 0xcccccccc, 0xdddddddd, 0xeeeeeeee},
		SpellsCnt: 0xa1b2c302,
		SpellArg: server.SpellAcceptArg{
			Obj: target,
			Pos: types.Pointf{X: targetX, Y: targetY},
		},
	}
	*liveData = server.GlyphInitData{
		Spells:    [5]uint32{99},
		SpellsCnt: 0x55667701,
		SpellArg:  server.SpellAcceptArg{Obj: target, Pos: types.Pointf{X: 1, Y: 2}},
	}
	object.ObjClass = objectlib.ClassPlayer | objectlib.ClassMissile
	object.Extent = extent
	object.ScriptIDVal = scriptID
	object.PosVec = types.Pointf{X: positionX, Y: positionY}
	object.ObjFlags = objectlib.Flags(flags)
	object.IDPtr = unsafe.Pointer(id)
	object.TeamVal.ID = server.TeamID(7)
	object.Field5 = status
	object.ScriptPickup = server.ScriptCallback{Flags: handlerFlags, Func: -1}
	object.Field34 = objectFrame
	object.Direction1 = server.Dir16(0xabcd)
	object.Direction2 = server.Dir16(0x1234)
	object.InitData = unsafe.Pointer(data)

	path := filepath.Join(t.TempDir(), "glyph-write.bin")
	cf, err := cryptfile.OpenFile(path, cryptfile.WriteOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	setObjectMapRuntimeGlobals4F4530(t, cf, gameFrame)

	var spellNameIDs []uint32
	deps := glyphXferNativeDeps4F5890{
		spellID: func(string) uint32 {
			t.Fatal("write mode parsed a spell name")
			return 0
		},
		spellName: func(value uint32) string {
			spellNameIDs = append(spellNameIDs, value)
			if len(spellNameIDs) == 1 {
				object.InitData = unsafe.Pointer(liveData)
			}
			switch value {
			case 10:
				return "SPELL_ALPHA"
			case 11:
				return "SPELL_BETA"
			default:
				t.Fatalf("spell name ID = %d", value)
				return ""
			}
		},
		transferInventory: func(uint16, *server.Object, int32) int32 {
			t.Fatal("write mode transferred inventory")
			return 0
		},
	}
	if got := glyphXferNative4F5890(cf, object, deps); got != 1 {
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
	writeU16(glyphXferCurrentVersion4F5890)
	writeU16(objectMapCurrentVersion4F4530)
	writeU32(extent)
	writeI32(scriptID)
	writeU32(math.Float32bits(positionX))
	writeU32(math.Float32bits(positionY))
	want.WriteByte(math.MaxUint8)
	writeU32(flags & objectMapFlagsMask4F4530)
	want.WriteByte(uint8(len("glyph")))
	want.WriteString("glyph")
	want.WriteByte(7)
	want.WriteByte(0)
	writeU16(0)
	writeU32(status & objectMapStatusMask4F4530)
	writeU16(1)
	writeU32(0)
	writeU32(handlerFlags)
	writeI32(int32(objectFrame - gameFrame))
	want.WriteByte(0xcd)
	writeU32(math.Float32bits(targetX))
	writeU32(math.Float32bits(targetY))
	want.WriteByte(2)
	want.WriteByte(uint8(len("SPELL_ALPHA")))
	want.WriteString("SPELL_ALPHA")
	want.WriteByte(uint8(len("SPELL_BETA")))
	want.WriteString("SPELL_BETA")

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want.Bytes()) {
		t.Fatalf("wire payload = %x, want %x", got, want.Bytes())
	}
	if !reflect.DeepEqual(spellNameIDs, []uint32{10, 10, 11, 11}) {
		t.Fatalf("spell name IDs = %v, want each cached spell reloaded", spellNameIDs)
	}
	if object.InitData != unsafe.Pointer(liveData) || object.Field34 != objectFrame {
		t.Fatalf("live InitData/Field34 = %p/%#x, want %p/%#x",
			object.InitData, object.Field34, liveData, objectFrame)
	}
	if data.SpellArg.Obj != target || data.SpellsCnt != 0xa1b2c302 || liveData.Spells[0] != 99 {
		t.Fatalf("cached/live Glyph data changed = %+v/%+v", *data, *liveData)
	}
}

func TestGlyphXferNativeRead4F5890ClearsNativeTargetAndRestoresCount(t *testing.T) {
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
		directionLow   = uint8(0x7e)
		targetX        = float32(-17.75)
		targetY        = float32(91.125)
	)

	var payload bytes.Buffer
	writeU16 := func(value uint16) { _ = binary.Write(&payload, binary.LittleEndian, value) }
	writeU32 := func(value uint32) { _ = binary.Write(&payload, binary.LittleEndian, value) }
	writeI32 := func(value int32) { writeU32(uint32(value)) }
	writeU16(glyphXferCurrentVersion4F5890)
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
	payload.WriteByte(directionLow)
	writeU32(math.Float32bits(targetX))
	writeU32(math.Float32bits(targetY))
	payload.WriteByte(2)
	payload.WriteByte(uint8(len("SPELL_ONE")))
	payload.WriteString("SPELL_ONE")
	payload.WriteByte(uint8(len("SPELL_TWO")))
	payload.WriteString("SPELL_TWO")

	path := filepath.Join(t.TempDir(), "glyph-read.bin")
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
	data, freeData := alloc.New(server.GlyphInitData{})
	defer freeData()
	liveData, freeLiveData := alloc.New(server.GlyphInitData{})
	defer freeLiveData()
	target, freeTarget := alloc.New(server.Object{})
	defer freeTarget()
	oldID, freeOldID := alloc.CString("")
	defer freeOldID()

	*data = server.GlyphInitData{
		Spells:    [5]uint32{0x11111111, 0x22222222, 0x33333333, 0x44444444, 0x55555555},
		SpellsCnt: 0xaabbcc00,
		SpellArg:  server.SpellAcceptArg{Obj: target, Pos: types.Pointf{X: 1, Y: 2}},
	}
	liveData.SpellArg.Obj = target
	object.ObjClass = objectlib.ClassPlayer | objectlib.ClassMissile
	object.IDPtr = unsafe.Pointer(oldID)
	object.ObjFlags = objectlib.Flags(originalFlags)
	object.Field5 = originalState
	object.Field34 = originalCount
	object.Direction1 = server.Dir16(0xab00)
	object.Direction2 = server.Dir16(0x1234)
	object.ScriptPickup.Func = -1
	object.InitData = unsafe.Pointer(data)

	for name, pointer := range map[string]unsafe.Pointer{
		"object":          unsafe.Pointer(object),
		"Glyph init data": unsafe.Pointer(data),
		"live Glyph data": unsafe.Pointer(liveData),
		"target object":   unsafe.Pointer(target),
		"old ID":          unsafe.Pointer(oldID),
	} {
		assertObjectMapNativePointer4F4530(t, name, pointer)
	}

	var inventoryCalls []glyphXferNativeInventoryCall4F5890
	var parsedNames []string
	deps := glyphXferNativeDeps4F5890{
		spellID: func(name string) uint32 {
			parsedNames = append(parsedNames, name)
			switch name {
			case "SPELL_ONE":
				return 101
			case "SPELL_TWO":
				return 202
			default:
				t.Fatalf("spell name = %q", name)
				return 0
			}
		},
		spellName: func(uint32) string {
			t.Fatal("read mode looked up a spell name")
			return ""
		},
		transferInventory: func(version uint16, gotObject *server.Object, count int32) int32 {
			inventoryCalls = append(inventoryCalls, glyphXferNativeInventoryCall4F5890{
				version: version, object: gotObject, count: count,
			})
			if data.SpellArg.Obj != nil {
				t.Fatalf("inventory transfer observed uncleared target %p", data.SpellArg.Obj)
			}
			gotObject.InitData = unsafe.Pointer(liveData)
			gotObject.Field34 = 0x11223344
			return 1
		},
	}
	if got := glyphXferNative4F5890(cf, object, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}

	wantInventory := []glyphXferNativeInventoryCall4F5890{{
		version: glyphXferCurrentVersion4F5890,
		object:  object,
		count:   int32(inventoryCount),
	}}
	if !reflect.DeepEqual(inventoryCalls, wantInventory) {
		t.Fatalf("inventory calls = %#v, want %#v", inventoryCalls, wantInventory)
	}
	if !reflect.DeepEqual(parsedNames, []string{"SPELL_ONE", "SPELL_TWO"}) {
		t.Fatalf("parsed names = %v", parsedNames)
	}
	if data.Spells != [5]uint32{101, 202, 0x33333333, 0x44444444, 0x55555555} {
		t.Fatalf("spells = %#v", data.Spells)
	}
	if data.SpellsCnt != 0xaabbcc02 {
		t.Fatalf("spell count = %#x, want high bytes preserved", data.SpellsCnt)
	}
	if data.SpellArg.Obj != nil || liveData.SpellArg.Obj != target {
		t.Fatalf("cached/live targets = %p/%p, want nil/%p", data.SpellArg.Obj, liveData.SpellArg.Obj, target)
	}
	if data.SpellArg.Pos != (types.Pointf{X: targetX, Y: targetY}) {
		t.Fatalf("target point = %v, want (%v,%v)", data.SpellArg.Pos, targetX, targetY)
	}
	wantDirection := server.Dir16(0xab00 | uint16(directionLow))
	if object.Direction1 != wantDirection || object.Direction2 != wantDirection {
		t.Fatalf("directions = %#x/%#x, want %#x", object.Direction1, object.Direction2, wantDirection)
	}
	if object.InitData != unsafe.Pointer(liveData) || object.Field34 != originalCount {
		t.Fatalf("live InitData/Field34 = %p/%#x, want %p/%#x",
			object.InitData, object.Field34, liveData, originalCount)
	}
	if object.IDPtr != unsafe.Pointer(oldID) {
		t.Fatalf("zero-length ID changed native pointer: got %p want %p", object.IDPtr, oldID)
	}
}

func TestGlyphXferExport4F5890PreservesNativePointersAndResult(t *testing.T) {
	object, freeObject := alloc.New(server.Object{})
	defer freeObject()
	context, freeContext := alloc.New(uint64(0))
	defer freeContext()

	assertObjectMapNativePointer4F4530(t, "object", unsafe.Pointer(object))
	assertObjectMapNativePointer4F4530(t, "context", unsafe.Pointer(context))

	old := glyphXferCall4F5890
	t.Cleanup(func() { glyphXferCall4F5890 = old })
	calls := 0
	glyphXferCall4F5890 = func(
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

	if got := glyphXferExportCall4F5890(object, unsafe.Pointer(context)); got != math.MinInt32 {
		t.Fatalf("first result = %d, want %d", got, int32(math.MinInt32))
	}
	if got := glyphXferExportCall4F5890(object, nil); got != math.MaxInt32 {
		t.Fatalf("second result = %d, want %d", got, int32(math.MaxInt32))
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
	runtime.KeepAlive(object)
	runtime.KeepAlive(context)
}
