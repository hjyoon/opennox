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

func TestRewardMarkerXferNativeLayout4F74D0(t *testing.T) {
	type field struct {
		name string
		got  uintptr
		want uintptr
	}
	wide := unsafe.Sizeof(uintptr(0)) == 8
	wantObjectSize := uintptr(780)
	wantField34 := uintptr(136)
	wantInitData := uintptr(692)
	if wide {
		wantObjectSize = 928
		wantField34 = 140
		wantInitData = 760
	}
	fields := []field{
		{name: "Object size", got: unsafe.Sizeof(server.Object{}), want: wantObjectSize},
		{name: "Object.Field34", got: unsafe.Offsetof(server.Object{}.Field34), want: wantField34},
		{name: "Object.InitData", got: unsafe.Offsetof(server.Object{}.InitData), want: wantInitData},
		{name: "RewardMarkerInitData size", got: unsafe.Sizeof(server.RewardMarkerInitData{}), want: 220},
		{name: "CategoryMask", got: unsafe.Offsetof(server.RewardMarkerInitData{}.CategoryMask), want: 0},
		{name: "RewardFlags", got: unsafe.Offsetof(server.RewardMarkerInitData{}.RewardFlags), want: 4},
		{name: "Field5", got: unsafe.Offsetof(server.RewardMarkerInitData{}.Field5), want: 5},
		{name: "Spells", got: unsafe.Offsetof(server.RewardMarkerInitData{}.Spells), want: 8},
		{name: "Abilities", got: unsafe.Offsetof(server.RewardMarkerInitData{}.Abilities), want: 145},
		{name: "Guides", got: unsafe.Offsetof(server.RewardMarkerInitData{}.Guides), want: 151},
		{name: "Field192", got: unsafe.Offsetof(server.RewardMarkerInitData{}.Field192), want: 192},
		{name: "Field196", got: unsafe.Offsetof(server.RewardMarkerInitData{}.Field196), want: 196},
		{name: "Field200", got: unsafe.Offsetof(server.RewardMarkerInitData{}.Field200), want: 200},
		{name: "Field204", got: unsafe.Offsetof(server.RewardMarkerInitData{}.Field204), want: 204},
		{name: "Field208", got: unsafe.Offsetof(server.RewardMarkerInitData{}.Field208), want: 208},
		{name: "ChanceMode", got: unsafe.Offsetof(server.RewardMarkerInitData{}.ChanceMode), want: 212},
		{name: "Field216", got: unsafe.Offsetof(server.RewardMarkerInitData{}.Field216), want: 216},
	}
	for _, field := range fields {
		if field.got != field.want {
			t.Errorf("%s native layout = %d, want %d", field.name, field.got, field.want)
		}
	}
}

func TestRewardMarkerXferRuntimeNameTables4F74D0(t *testing.T) {
	deps := rewardMarkerXferRuntimeDeps4F74D0(nil)
	for id := 0; id < rewardMarkerXferSpellCount4F74D0; id++ {
		name := deps.spellName(id)
		if name == "" || deps.spellID(name) != id {
			t.Errorf("spell %d name/round-trip = %q/%d", id, name, deps.spellID(name))
		}
	}
	for id, name := range server.AbilityNames {
		if got := deps.abilityName(id); got != name || deps.abilityID(got) != id {
			t.Errorf("ability %d name/round-trip = %q/%d, want %q/%d", id, got, deps.abilityID(got), name, id)
		}
	}
	for id := 0; id < rewardMarkerXferGuideCount4F74D0; id++ {
		name := deps.guideName(id)
		if name == "" || deps.guideID(name) != id {
			t.Errorf("guide %d name/round-trip = %q/%d", id, name, deps.guideID(name))
		}
	}
	if deps.spellID("SPELL_UNKNOWN") != 0 || deps.abilityID("ABILITY_UNKNOWN") != 0 || deps.guideID("GuideUnknown") != 0 {
		t.Fatal("unknown names did not resolve to invalid ID zero")
	}
	if got := deps.guideName(-1); got != "" {
		t.Errorf("guideName(-1) = %q, want empty", got)
	}
	if got := deps.guideName(rewardMarkerXferGuideCount4F74D0); got != "" {
		t.Errorf("guideName(41) = %q, want empty", got)
	}
}

func rewardMarkerXferNativeTestDeps4F74D0() rewardMarkerXferNativeDeps4F74D0 {
	return rewardMarkerXferNativeDeps4F74D0{
		spellName: func(id int) string {
			return map[int]string{1: "S_ONE", 2: "S_TWO"}[id]
		},
		spellID: func(name string) int {
			return map[string]int{"S_ONE": 1, "S_TWO": 2}[name]
		},
		abilityName: func(id int) string {
			return map[int]string{3: "A_THREE"}[id]
		},
		abilityID: func(name string) int {
			return map[string]int{"A_THREE": 3}[name]
		},
		guideName: func(id int) string {
			return map[int]string{4: "G_FOUR"}[id]
		},
		guideID: func(name string) int {
			return map[string]int{"G_FOUR": 4}[name]
		},
		transferInventory: func(uint16, *server.Object, int32) int32 {
			panic("unexpected RewardMarkerXfer inventory call")
		},
	}
}

func rewardMarkerXferNativeSpecificPayload4F74D0() []byte {
	var payload bytes.Buffer
	writeU16 := func(value uint16) { _ = binary.Write(&payload, binary.LittleEndian, value) }
	writeU32 := func(value uint32) { _ = binary.Write(&payload, binary.LittleEndian, value) }
	writeName := func(name string) {
		payload.WriteByte(uint8(len(name)))
		payload.WriteString(name)
	}
	writeU32(0x11223344)
	payload.Write([]byte{0xa5, 0xb6, 0xc7, 0xd8})
	writeU16(2)
	writeName("S_ONE")
	writeName("S_TWO")
	writeU16(1)
	writeName("A_THREE")
	writeU16(1)
	writeName("G_FOUR")
	writeU32(0x22222222) // Field196 precedes Field192 on the wire.
	writeU32(0x11111111)
	writeU32(0x33333333)
	writeU32(0x44444444)
	writeU32(0x55555555)
	writeU32(0x66666666)
	payload.WriteByte(0xdd)
	return payload.Bytes()
}

func TestRewardMarkerXferNativeRoundTrip4F74D0PreservesPointersAndWire(t *testing.T) {
	object, freeObject := alloc.New(server.Object{})
	defer freeObject()
	data, freeData := alloc.New(server.RewardMarkerInitData{})
	defer freeData()
	id, freeID := alloc.CString("reward-marker")
	defer freeID()
	for name, pointer := range map[string]unsafe.Pointer{
		"object":                unsafe.Pointer(object),
		"RewardMarker InitData": unsafe.Pointer(data),
		"ID":                    unsafe.Pointer(id),
	} {
		assertObjectMapNativePointer4F4530(t, name, pointer)
	}

	data.CategoryMask = 0x11223344
	data.RewardFlags = 0xa5
	data.Field5 = [3]byte{0xb6, 0xc7, 0xd8}
	data.Spells[1], data.Spells[2] = 1, 1
	data.Abilities[3] = 1
	data.Guides[4] = 1
	data.Field192 = 0x11111111
	data.Field196 = 0x22222222
	data.Field200 = 0x33333333
	data.Field204 = 0x44444444
	data.Field208 = 0x55555555
	data.ChanceMode = 0x66666666
	data.Field216 = 0xaabbccdd
	object.ObjClass = objectlib.ClassPlayer | objectlib.ClassMissile
	object.IDPtr = unsafe.Pointer(id)
	object.ScriptPickup.Func = -1
	object.InitData = unsafe.Pointer(data)

	path := filepath.Join(t.TempDir(), "reward-marker.bin")
	writeFile, err := cryptfile.OpenFile(path, cryptfile.WriteOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	setObjectMapRuntimeGlobals4F4530(t, writeFile, 0)
	if got := rewardMarkerXferNative4F74D0(writeFile, object, rewardMarkerXferNativeTestDeps4F74D0()); got != 1 {
		_ = writeFile.Close()
		t.Fatalf("write result = %d, want 1", got)
	}
	if err := writeFile.Close(); err != nil {
		t.Fatal(err)
	}
	wire, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(wire) < 2 || binary.LittleEndian.Uint16(wire[:2]) != rewardMarkerXferCurrentVersion4F74D0 {
		t.Fatalf("wire version prefix = %x, want 63", wire)
	}
	wantSpecific := rewardMarkerXferNativeSpecificPayload4F74D0()
	if !bytes.HasSuffix(wire, wantSpecific) {
		t.Fatalf("RewardMarker wire suffix = %x, want exact suffix %x", wire, wantSpecific)
	}
	if object.InitData != unsafe.Pointer(data) || object.Field34 != 0 || data.Field216 != 0xaabbccdd {
		t.Fatalf("write pointers/fields changed: InitData=%p Field34=%#x Field216=%#x", object.InitData, object.Field34, data.Field216)
	}

	readObject, freeReadObject := alloc.New(server.Object{})
	defer freeReadObject()
	readData, freeReadData := alloc.New(server.RewardMarkerInitData{})
	defer freeReadData()
	oldID, freeOldID := alloc.CString("")
	defer freeOldID()
	readData.Spells[9] = 7
	readData.Field216 = 0xaabbcc00
	readObject.ObjClass = objectlib.ClassPlayer | objectlib.ClassMissile
	readObject.IDPtr = unsafe.Pointer(oldID)
	readObject.ScriptPickup.Func = -1
	readObject.InitData = unsafe.Pointer(readData)
	for name, pointer := range map[string]unsafe.Pointer{
		"read object":                unsafe.Pointer(readObject),
		"read RewardMarker InitData": unsafe.Pointer(readData),
	} {
		assertObjectMapNativePointer4F4530(t, name, pointer)
	}

	readFile, err := cryptfile.OpenFile(path, cryptfile.ReadOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = readFile.Close() }()
	setObjectMapRuntimeGlobals4F4530(t, readFile, 0)
	if got := rewardMarkerXferNative4F74D0(readFile, readObject, rewardMarkerXferNativeTestDeps4F74D0()); got != 1 {
		t.Fatalf("read result = %d, want 1", got)
	}
	if readObject.InitData != unsafe.Pointer(readData) || readObject.Field34 != 0 {
		t.Fatalf("read InitData/Field34 = %p/%#x, want cached pointer/zero", readObject.InitData, readObject.Field34)
	}
	if readData.CategoryMask != data.CategoryMask || readData.RewardFlags != data.RewardFlags || readData.Field5 != data.Field5 {
		t.Fatalf("read header = %#x/%#x/%x, want %#x/%#x/%x",
			readData.CategoryMask, readData.RewardFlags, readData.Field5,
			data.CategoryMask, data.RewardFlags, data.Field5)
	}
	if readData.Spells[1] != 1 || readData.Spells[2] != 1 || readData.Spells[9] != 7 ||
		readData.Abilities[3] != 1 || readData.Guides[4] != 1 {
		t.Fatalf("read lists did not set resolved IDs while retaining existing flags")
	}
	if readData.Field192 != data.Field192 || readData.Field196 != data.Field196 ||
		readData.Field200 != data.Field200 || readData.Field204 != data.Field204 ||
		readData.Field208 != data.Field208 || readData.ChanceMode != data.ChanceMode ||
		readData.Field216 != 0xaabbccdd {
		t.Fatalf("read suffix = %+v, want original values and preserved Field216 high bytes", *readData)
	}
}

func TestRewardMarkerXferExport4F74D0PreservesNativePointerAndResult(t *testing.T) {
	object, freeObject := alloc.New(server.Object{})
	defer freeObject()
	assertObjectMapNativePointer4F4530(t, "object", unsafe.Pointer(object))

	old := rewardMarkerXferCall4F74D0
	t.Cleanup(func() { rewardMarkerXferCall4F74D0 = old })
	calls := 0
	rewardMarkerXferCall4F74D0 = func(_ *cryptfile.CryptFile, gotObject *server.Object) int32 {
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

	if got := rewardMarkerXferExportCall4F74D0(object); got != math.MinInt32 {
		t.Fatalf("first result = %d, want %d", got, int32(math.MinInt32))
	}
	if got := rewardMarkerXferExportCall4F74D0(nil); got != math.MaxInt32 {
		t.Fatalf("second result = %d, want %d", got, int32(math.MaxInt32))
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
	runtime.KeepAlive(object)
}
