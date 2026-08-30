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

type monsterGeneratorXferNativeScriptCall4F7130 struct {
	handler *server.ScriptCallback
	context unsafe.Pointer
}

type monsterGeneratorXferNativeInventoryCall4F7130 struct {
	version uint16
	object  *server.Object
	count   int32
}

func TestMonsterGeneratorXferNativeLayout4F7130(t *testing.T) {
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
		{name: "Object.Field189", got: unsafe.Offsetof(server.Object{}.Field189), pe32: 756, wide: 888},
		{name: "MonsterGenUpdateData size", got: unsafe.Sizeof(server.MonsterGenUpdateData{}), pe32: 164, wide: 216},
		{name: "MonsterGenUpdateData.Field48", got: unsafe.Offsetof(server.MonsterGenUpdateData{}.Field48), pe32: 48, wide: 96},
		{name: "MonsterGenUpdateData.FuncInd52", got: unsafe.Offsetof(server.MonsterGenUpdateData{}.FuncInd52), pe32: 52, wide: 100},
		{name: "MonsterGenUpdateData.Field56", got: unsafe.Offsetof(server.MonsterGenUpdateData{}.Field56), pe32: 56, wide: 104},
		{name: "MonsterGenUpdateData.FuncInd60", got: unsafe.Offsetof(server.MonsterGenUpdateData{}.FuncInd60), pe32: 60, wide: 108},
		{name: "MonsterGenUpdateData.Field64", got: unsafe.Offsetof(server.MonsterGenUpdateData{}.Field64), pe32: 64, wide: 112},
		{name: "MonsterGenUpdateData.FuncInd68", got: unsafe.Offsetof(server.MonsterGenUpdateData{}.FuncInd68), pe32: 68, wide: 116},
		{name: "MonsterGenUpdateData.ScriptCollision", got: unsafe.Offsetof(server.MonsterGenUpdateData{}.ScriptCollision), pe32: 72, wide: 120},
		{name: "MonsterGenUpdateData.SpawnRate", got: unsafe.Offsetof(server.MonsterGenUpdateData{}.SpawnRate), pe32: 80, wide: 128},
		{name: "MonsterGenUpdateData.QuestSpawnRate", got: unsafe.Offsetof(server.MonsterGenUpdateData{}.QuestSpawnRate), pe32: 83, wide: 131},
		{name: "MonsterGenUpdateData.ActiveCount", got: unsafe.Offsetof(server.MonsterGenUpdateData{}.ActiveCount), pe32: 86, wide: 134},
		{name: "MonsterGenUpdateData.MaxActive", got: unsafe.Offsetof(server.MonsterGenUpdateData{}.MaxActive), pe32: 87, wide: 135},
		{name: "MonsterGenUpdateData.Frame88", got: unsafe.Offsetof(server.MonsterGenUpdateData{}.Frame88), pe32: 88, wide: 136},
		{name: "MonsterGenUpdateData.Field92", got: unsafe.Offsetof(server.MonsterGenUpdateData{}.Field92), pe32: 92, wide: 140},
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

	data := server.MonsterGenUpdateData{
		Field48: 0x01020304, FuncInd52: 0xf1020304,
		Field56: 0x11121314, FuncInd60: 0xf1121314,
		Field64: 0x21222324, FuncInd68: 0xf1222324,
		ScriptCollision: server.ScriptCallback{Flags: 0x31323334, Func: -0x1020304},
	}
	tests := []struct {
		slot      monsterGeneratorScriptSlot4F7130
		wantFlags uint32
		wantFunc  int32
	}{
		{slot: monsterGeneratorScript48_4F7130, wantFlags: data.Field48, wantFunc: int32(data.FuncInd52)},
		{slot: monsterGeneratorScript56_4F7130, wantFlags: data.Field56, wantFunc: int32(data.FuncInd60)},
		{slot: monsterGeneratorScript72_4F7130, wantFlags: data.ScriptCollision.Flags, wantFunc: data.ScriptCollision.Func},
		{slot: monsterGeneratorScript64_4F7130, wantFlags: data.Field64, wantFunc: int32(data.FuncInd68)},
	}
	for _, test := range tests {
		handler := monsterGeneratorScriptHandler4F7130(&data, test.slot)
		if handler.Flags != test.wantFlags || handler.Func != test.wantFunc {
			t.Errorf("slot %d callback = %#x/%#x, want %#x/%#x",
				test.slot, handler.Flags, uint32(handler.Func), test.wantFlags, uint32(test.wantFunc))
		}
	}
	if got := monsterGeneratorScriptContext4F7130(nil, 1920); got != nil {
		t.Fatalf("nil script context = %p, want nil", got)
	}
	var scriptData [2572]byte
	if got, want := monsterGeneratorScriptContext4F7130(unsafe.Pointer(&scriptData[0]), 2304), unsafe.Pointer(&scriptData[2304]); got != want {
		t.Fatalf("script context = %p, want %p", got, want)
	}
}

func TestMonsterGeneratorXferNativeWrite4F7130PreservesPointersAndExactWire(t *testing.T) {
	object, freeObject := alloc.New(server.Object{})
	defer freeObject()
	data, freeData := alloc.New(server.MonsterGenUpdateData{})
	defer freeData()
	prototypeA, freePrototypeA := alloc.New(server.Object{})
	defer freePrototypeA()
	prototypeB, freePrototypeB := alloc.New(server.Object{})
	defer freePrototypeB()
	scriptData, freeScriptData := alloc.New([2572]byte{})
	defer freeScriptData()
	id, freeID := alloc.CString("monster-generator-native")
	defer freeID()

	for name, pointer := range map[string]unsafe.Pointer{
		"object":      unsafe.Pointer(object),
		"update data": unsafe.Pointer(data),
		"prototype A": unsafe.Pointer(prototypeA),
		"prototype B": unsafe.Pointer(prototypeB),
		"script data": unsafe.Pointer(scriptData),
		"ID":          unsafe.Pointer(id),
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
		gameFrame    = uint32(0x01020304)
	)
	object.ObjClass = objectlib.ClassMonsterGenerator | objectlib.ClassImmobile
	object.Extent = extent
	object.ScriptIDVal = scriptID
	object.PosVec = types.Pointf{X: positionX, Y: positionY}
	object.ObjFlags = objectlib.Flags(flags)
	object.IDPtr = unsafe.Pointer(id)
	object.TeamVal.ID = server.TeamID(7)
	object.Field5 = status
	object.ScriptPickup = server.ScriptCallback{Flags: handlerFlags, Func: -1}
	object.Field34 = 0
	object.UpdateData = unsafe.Pointer(data)
	object.Field189 = unsafe.Pointer(scriptData)
	data.Field0[0] = prototypeA
	data.Field0[5] = prototypeB
	data.Field48, data.FuncInd52 = 0x01020304, 0xf1020304
	data.Field56, data.FuncInd60 = 0x11121314, 0xf1121314
	data.ScriptCollision = server.ScriptCallback{Flags: 0x21222324, Func: -0x1020304}
	data.Field64, data.FuncInd68 = 0x31323334, 0xf1323334
	data.SpawnRate = [3]uint8{1, 2, 3}
	data.ActiveCount = 4
	data.MaxActive = 5
	data.Frame88 = 0x41424344
	data.QuestSpawnRate = [3]uint8{6, 7, 8}
	data.Field92 = 0x51525354

	path := filepath.Join(t.TempDir(), "monster-generator-write.bin")
	cf, err := cryptfile.OpenFile(path, cryptfile.WriteOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	setObjectMapRuntimeGlobals4F4530(t, cf, gameFrame)

	var scriptCalls []monsterGeneratorXferNativeScriptCall4F7130
	var saved []*server.Object
	deps := monsterGeneratorXferNativeDeps4F7130{
		loadTypeName: func(prototype *server.Object) []byte {
			switch prototype {
			case prototypeA:
				return []byte("Imp")
			case prototypeB:
				return []byte("Ogre")
			default:
				t.Fatalf("unexpected prototype %p", prototype)
				return nil
			}
		},
		transferScript: func(handler *server.ScriptCallback, context unsafe.Pointer) int32 {
			scriptCalls = append(scriptCalls, monsterGeneratorXferNativeScriptCall4F7130{handler: handler, context: context})
			handler.Flags = objectReadOldRWU32Native4F4170(cf, handler.Flags)
			handler.Func = objectReadOldRWI32Native4F4170(cf, handler.Func)
			return 0
		},
		saveObject: func(prototype *server.Object) int32 {
			saved = append(saved, prototype)
			marker := uint16(0xa001)
			if prototype == prototypeB {
				marker = 0xa002
			}
			objectReadOldRWU16Native4F4170(cf, marker)
			return 0
		},
		newObjectByTypeName: func([]byte) *server.Object { return nil },
		callObjectXfer:      func(*server.Object) int32 { return 0 },
		transferInventory:   func(uint16, *server.Object, int32) int32 { return 0 },
	}
	if got := monsterGeneratorXferNative4F7130(cf, object, deps); got != 1 {
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
	writeU16(monsterGeneratorXferCurrentVersion4F7130)
	writeU16(objectMapCurrentVersion4F4530)
	writeU32(extent)
	writeI32(scriptID)
	writeU32(math.Float32bits(positionX))
	writeU32(math.Float32bits(positionY))
	want.WriteByte(math.MaxUint8)
	writeU32(flags & objectMapFlagsMask4F4530)
	want.WriteByte(uint8(len("monster-generator-native")))
	want.WriteString("monster-generator-native")
	want.WriteByte(7)
	want.WriteByte(0)
	writeU16(0)
	writeU32(status & objectMapStatusMask4F4530)
	writeU16(1)
	writeU32(0)
	writeU32(handlerFlags)
	writeI32(-int32(gameFrame))
	want.WriteByte(3)
	want.Write([]byte{1, 2, 3})
	want.WriteByte(4)
	want.WriteByte(5)
	writeU32(0x41424344)
	writeU32(0x01020304)
	writeI32(int32(data.FuncInd52))
	writeU32(0x11121314)
	writeI32(int32(data.FuncInd60))
	writeU32(0x21222324)
	writeI32(-0x1020304)
	writeU32(0x31323334)
	writeI32(int32(data.FuncInd68))
	want.WriteByte(3)
	want.WriteByte(1)
	want.WriteByte(3)
	want.WriteString("Imp")
	writeU16(0xa001)
	want.WriteByte(1)
	want.WriteByte(4)
	want.WriteString("Ogre")
	writeU16(0xa002)
	want.WriteByte(0)
	want.WriteByte(3)
	want.Write([]byte{6, 7, 8})
	writeU32(0x51525354)

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want.Bytes()) {
		t.Fatalf("wire payload = %x, want %x", got, want.Bytes())
	}
	wantContexts := []unsafe.Pointer{
		unsafe.Add(unsafe.Pointer(scriptData), 1920),
		unsafe.Add(unsafe.Pointer(scriptData), 2048),
		unsafe.Add(unsafe.Pointer(scriptData), 2176),
		unsafe.Add(unsafe.Pointer(scriptData), 2304),
	}
	if len(scriptCalls) != 4 {
		t.Fatalf("script calls = %d, want 4", len(scriptCalls))
	}
	for index := range scriptCalls {
		if scriptCalls[index].context != wantContexts[index] {
			t.Errorf("script context[%d] = %p, want %p", index, scriptCalls[index].context, wantContexts[index])
		}
	}
	if !reflect.DeepEqual(saved, []*server.Object{prototypeA, prototypeB}) {
		t.Fatalf("saved objects = %v", saved)
	}
	if object.UpdateData != unsafe.Pointer(data) || object.Field189 != unsafe.Pointer(scriptData) || object.Field34 != 0 {
		t.Fatalf("native pointers/count changed: update=%p script=%p Field34=%#x", object.UpdateData, object.Field189, object.Field34)
	}
}

func TestMonsterGeneratorXferNativeRead4F7130StoresNativePointersAndExactWire(t *testing.T) {
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
		streamSentinel = uint8(0x5a)
	)

	var payload bytes.Buffer
	writeU16 := func(value uint16) { _ = binary.Write(&payload, binary.LittleEndian, value) }
	writeU32 := func(value uint32) { _ = binary.Write(&payload, binary.LittleEndian, value) }
	writeI32 := func(value int32) { writeU32(uint32(value)) }
	writeU16(monsterGeneratorXferCurrentVersion4F7130)
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
	payload.WriteByte(3)
	payload.Write([]byte{11, 12, 13})
	payload.WriteByte(14)
	payload.WriteByte(15)
	writeU32(0x61626364)
	for index := int32(0); index < 4; index++ {
		writeU32(0x71000000 + uint32(index))
		writeI32(-100 - index)
	}
	payload.WriteByte(2)
	payload.WriteByte(2)
	payload.WriteByte(3)
	payload.WriteString("Imp")
	writeU16(0x1111)
	writeU32(0xaaaaaaaa)
	payload.WriteByte(0xe1)
	payload.WriteByte(3)
	payload.WriteString("Bat")
	writeU16(0x2222)
	writeU32(0xbbbbbbbb)
	payload.WriteByte(0xe2)
	payload.WriteByte(1)
	payload.WriteByte(4)
	payload.WriteString("Ogre")
	writeU16(0x3333)
	writeU32(0xcccccccc)
	payload.WriteByte(0xe3)
	payload.WriteByte(3)
	payload.Write([]byte{21, 22, 23})
	writeU32(0x81828384)
	payload.WriteByte(streamSentinel)

	path := filepath.Join(t.TempDir(), "monster-generator-read.bin")
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
	data, freeData := alloc.New(server.MonsterGenUpdateData{})
	defer freeData()
	imp, freeImp := alloc.New(server.Object{})
	defer freeImp()
	bat, freeBat := alloc.New(server.Object{})
	defer freeBat()
	ogre, freeOgre := alloc.New(server.Object{})
	defer freeOgre()
	scriptData, freeScriptData := alloc.New([2572]byte{})
	defer freeScriptData()
	oldID, freeOldID := alloc.CString("")
	defer freeOldID()
	object.ObjClass = objectlib.ClassMonsterGenerator | objectlib.ClassImmobile
	object.IDPtr = unsafe.Pointer(oldID)
	object.ObjFlags = objectlib.Flags(originalFlags)
	object.Field5 = originalState
	object.Field34 = originalCount
	object.ScriptPickup.Func = -1
	object.UpdateData = unsafe.Pointer(data)
	object.Field189 = unsafe.Pointer(scriptData)

	for name, pointer := range map[string]unsafe.Pointer{
		"object":       unsafe.Pointer(object),
		"update data":  unsafe.Pointer(data),
		"script data":  unsafe.Pointer(scriptData),
		"created Imp":  unsafe.Pointer(imp),
		"created Bat":  unsafe.Pointer(bat),
		"created Ogre": unsafe.Pointer(ogre),
		"existing ID":  unsafe.Pointer(oldID),
	} {
		assertObjectMapNativePointer4F4530(t, name, pointer)
	}

	var scriptCalls []monsterGeneratorXferNativeScriptCall4F7130
	var xferCalls []*server.Object
	var inventoryCalls []monsterGeneratorXferNativeInventoryCall4F7130
	deps := monsterGeneratorXferNativeDeps4F7130{
		loadTypeName: func(*server.Object) []byte { return nil },
		transferScript: func(handler *server.ScriptCallback, context unsafe.Pointer) int32 {
			scriptCalls = append(scriptCalls, monsterGeneratorXferNativeScriptCall4F7130{handler: handler, context: context})
			handler.Flags = objectReadOldRWU32Native4F4170(cf, handler.Flags)
			handler.Func = objectReadOldRWI32Native4F4170(cf, handler.Func)
			return 0
		},
		saveObject: func(*server.Object) int32 { return 0 },
		newObjectByTypeName: func(name []byte) *server.Object {
			switch cStringBytes528DB0(name) {
			case "Imp":
				return imp
			case "Bat":
				return bat
			case "Ogre":
				return ogre
			default:
				return nil
			}
		},
		callObjectXfer: func(created *server.Object) int32 {
			xferCalls = append(xferCalls, created)
			want := uint8(0xe1 + len(xferCalls) - 1)
			if got := objectReadOldRWU8Native4F4170(cf, 0); got != want {
				t.Fatalf("xfer marker = %#x, want %#x", got, want)
			}
			return 1
		},
		transferInventory: func(version uint16, gotObject *server.Object, count int32) int32 {
			inventoryCalls = append(inventoryCalls, monsterGeneratorXferNativeInventoryCall4F7130{
				version: version, object: gotObject, count: count,
			})
			gotObject.Field34 = 0x11223344
			return 1
		},
	}
	if got := monsterGeneratorXferNative4F7130(cf, object, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if data.SpawnRate != [3]uint8{11, 12, 13} || data.ActiveCount != 14 || data.MaxActive != 15 || data.Frame88 != 0x61626364 {
		t.Fatalf("prefix data = spawn %v active %d max %d frame %#x", data.SpawnRate, data.ActiveCount, data.MaxActive, data.Frame88)
	}
	callbacks := []*server.ScriptCallback{
		monsterGeneratorScriptHandler4F7130(data, monsterGeneratorScript48_4F7130),
		monsterGeneratorScriptHandler4F7130(data, monsterGeneratorScript56_4F7130),
		monsterGeneratorScriptHandler4F7130(data, monsterGeneratorScript72_4F7130),
		monsterGeneratorScriptHandler4F7130(data, monsterGeneratorScript64_4F7130),
	}
	for index, callback := range callbacks {
		if callback.Flags != 0x71000000+uint32(index) || callback.Func != -100-int32(index) {
			t.Errorf("callback[%d] = %#x/%d", index, callback.Flags, callback.Func)
		}
	}
	wantContexts := []unsafe.Pointer{
		unsafe.Add(unsafe.Pointer(scriptData), 1920),
		unsafe.Add(unsafe.Pointer(scriptData), 2048),
		unsafe.Add(unsafe.Pointer(scriptData), 2176),
		unsafe.Add(unsafe.Pointer(scriptData), 2304),
	}
	if len(scriptCalls) != len(wantContexts) {
		t.Fatalf("script calls = %d, want %d", len(scriptCalls), len(wantContexts))
	}
	for index := range scriptCalls {
		if scriptCalls[index].context != wantContexts[index] {
			t.Errorf("script context[%d] = %p, want %p", index, scriptCalls[index].context, wantContexts[index])
		}
	}
	if data.Field0[0] != imp || data.Field0[1] != bat || data.Field0[4] != ogre {
		t.Fatalf("prototype pointers = %p/%p/%p, want %p/%p/%p", data.Field0[0], data.Field0[1], data.Field0[4], imp, bat, ogre)
	}
	if !reflect.DeepEqual(xferCalls, []*server.Object{imp, bat, ogre}) {
		t.Fatalf("xfer calls = %v", xferCalls)
	}
	if data.QuestSpawnRate != [3]uint8{21, 22, 23} || data.Field92 != 0x81828384 {
		t.Fatalf("suffix data = quest %v Field92 %#x", data.QuestSpawnRate, data.Field92)
	}
	wantInventory := []monsterGeneratorXferNativeInventoryCall4F7130{{
		version: 63, object: object, count: int32(inventoryCount),
	}}
	if !reflect.DeepEqual(inventoryCalls, wantInventory) || object.Field34 != originalCount {
		t.Fatalf("inventory/Field34 = %+v/%#x", inventoryCalls, object.Field34)
	}
	if object.UpdateData != unsafe.Pointer(data) || object.Field189 != unsafe.Pointer(scriptData) || object.IDPtr != unsafe.Pointer(oldID) {
		t.Fatalf("native object pointers changed")
	}
	if object.Extent != extent || object.ScriptIDVal != scriptID || object.PosVec != (types.Pointf{X: positionX, Y: positionY}) {
		t.Fatalf("common fields = extent %#x script %#x position %+v", object.Extent, uint32(object.ScriptIDVal), object.PosVec)
	}
	if got := objectReadOldRWU8Native4F4170(cf, 0); got != streamSentinel {
		t.Fatalf("byte after generator payload = %#x, want %#x", got, streamSentinel)
	}
}

func TestMonsterGeneratorXferExport4F7130PreservesNativePointerAndResult(t *testing.T) {
	object, freeObject := alloc.New(server.Object{})
	defer freeObject()
	assertObjectMapNativePointer4F4530(t, "object", unsafe.Pointer(object))

	old := monsterGeneratorXferCall4F7130
	t.Cleanup(func() { monsterGeneratorXferCall4F7130 = old })
	calls := 0
	monsterGeneratorXferCall4F7130 = func(_ *cryptfile.CryptFile, gotObject *server.Object) int32 {
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

	if got := monsterGeneratorXferExportCall4F7130(object); got != math.MinInt32 {
		t.Fatalf("first result = %d, want %d", got, int32(math.MinInt32))
	}
	if got := monsterGeneratorXferExportCall4F7130(nil); got != math.MaxInt32 {
		t.Fatalf("second result = %d, want %d", got, int32(math.MaxInt32))
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
	runtime.KeepAlive(object)
}
