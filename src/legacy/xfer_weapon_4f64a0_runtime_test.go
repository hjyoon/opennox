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

type weaponXferNativeInventoryCall4F64A0 struct {
	version uint16
	object  *server.Object
	count   int32
}

func TestWeaponXferNativeLayout4F64A0(t *testing.T) {
	type field struct {
		name string
		got  uintptr
		pe32 uintptr
		wide uintptr
	}
	fields := []field{
		{name: "Object size", got: unsafe.Sizeof(server.Object{}), pe32: 780, wide: 928},
		{name: "Object.Field34", got: unsafe.Offsetof(server.Object{}.Field34), pe32: 136, wide: 140},
		{name: "Object.HealthData", got: unsafe.Offsetof(server.Object{}.HealthData), pe32: 556, wide: 616},
		{name: "Object.InitData", got: unsafe.Offsetof(server.Object{}.InitData), pe32: 692, wide: 760},
		{name: "Object.UseData", got: unsafe.Offsetof(server.Object{}.UseData), pe32: 736, wide: 848},
		{name: "Object.UpdateData", got: unsafe.Offsetof(server.Object{}.UpdateData), pe32: 748, wide: 872},
		{name: "ModifierInitData size", got: unsafe.Sizeof(server.ModifierInitData{}), pe32: 20, wide: 40},
		{name: "ModifierInitData.Field16", got: unsafe.Offsetof(server.ModifierInitData{}.Field16), pe32: 16, wide: 32},
		{name: "Modifier size", got: unsafe.Sizeof(server.Modifier{}), pe32: 88, wide: 112},
		{name: "Modifier.Durability52", got: unsafe.Offsetof(server.Modifier{}.Durability52), pe32: 52, wide: 64},
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

	fixed := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{name: "HealthData size", got: unsafe.Sizeof(server.HealthData{}), want: 20},
		{name: "HealthData.Cur", got: unsafe.Offsetof(server.HealthData{}.Cur), want: 0},
		{name: "HealthData.Field2", got: unsafe.Offsetof(server.HealthData{}.Field2), want: 2},
		{name: "HealthData.Max", got: unsafe.Offsetof(server.HealthData{}.Max), want: 4},
		{name: "WandUseData size", got: unsafe.Sizeof(server.WandUseData{}), want: 116},
		{name: "WandUseData.ProjectileName", got: unsafe.Offsetof(server.WandUseData{}.ProjectileName), want: 4},
		{name: "WandUseData.ProjectileType", got: unsafe.Offsetof(server.WandUseData{}.ProjectileType), want: 84},
		{name: "WandUseData.Sound", got: unsafe.Offsetof(server.WandUseData{}.Sound), want: 88},
		{name: "WandUseData.Spell", got: unsafe.Offsetof(server.WandUseData{}.Spell), want: 92},
		{name: "WandUseData.Flags", got: unsafe.Offsetof(server.WandUseData{}.Flags), want: 96},
		{name: "WandUseData.Cooldown", got: unsafe.Offsetof(server.WandUseData{}.Cooldown), want: 100},
		{name: "WandUseData.LastUsed", got: unsafe.Offsetof(server.WandUseData{}.LastUsed), want: 104},
		{name: "WandUseData.Charge", got: unsafe.Offsetof(server.WandUseData{}.Charge), want: 108},
		{name: "WandUseData.MaxCharge", got: unsafe.Offsetof(server.WandUseData{}.MaxCharge), want: 109},
		{name: "WandUseData.Progress", got: unsafe.Offsetof(server.WandUseData{}.Progress), want: 112},
		{name: "WeaponArmorUpdateData size", got: unsafe.Sizeof(server.WeaponArmorUpdateData{}), want: 8},
		{name: "WeaponArmorUpdateData.Field4", got: unsafe.Offsetof(server.WeaponArmorUpdateData{}.Field4), want: 4},
	}
	for _, field := range fixed {
		if field.got != field.want {
			t.Errorf("%s = %d, want %d", field.name, field.got, field.want)
		}
	}
}

func TestWeaponXferNativeWrite4F64A0PreservesPointersAndExactWire(t *testing.T) {
	object, freeObject := alloc.New(server.Object{})
	defer freeObject()
	initData, freeInitData := alloc.New(server.ModifierInitData{})
	defer freeInitData()
	health, freeHealth := alloc.New(server.HealthData{Cur: 73, Field2: 90, Max: 100})
	defer freeHealth()
	update, freeUpdate := alloc.New(server.WeaponArmorUpdateData{Field0: 0x10203040, Field4: 0xa1b2c3d4})
	defer freeUpdate()
	id, freeID := alloc.CString("weapon-native")
	defer freeID()
	*health = server.HealthData{Cur: 73, Field2: 90, Max: 100}
	*update = server.WeaponArmorUpdateData{Field0: 0x10203040, Field4: 0xa1b2c3d4}

	for name, pointer := range map[string]unsafe.Pointer{
		"object":      unsafe.Pointer(object),
		"init data":   unsafe.Pointer(initData),
		"health data": unsafe.Pointer(health),
		"update data": unsafe.Pointer(update),
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
		objectFrame  = uint32(0x11223344)
		gameFrame    = uint32(0x01020304)
	)
	object.ObjClass = objectlib.ClassWeapon | objectlib.ClassClientPersist
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
	object.InitData = unsafe.Pointer(initData)
	object.HealthData = health
	object.UpdateData = unsafe.Pointer(update)

	path := filepath.Join(t.TempDir(), "weapon-write.bin")
	cf, err := cryptfile.OpenFile(path, cryptfile.WriteOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	setObjectMapRuntimeGlobals4F4530(t, cf, gameFrame)
	if got := Nox_xxx_XFerWeaponNative4F64A0(cf, object); got != 1 {
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
	writeU16(weaponXferCurrentVersion4F64A0)
	writeU16(objectMapCurrentVersion4F4530)
	writeU32(extent)
	writeI32(scriptID)
	writeU32(math.Float32bits(positionX))
	writeU32(math.Float32bits(positionY))
	want.WriteByte(math.MaxUint8)
	writeU32(flags & objectMapFlagsMask4F4530)
	want.WriteByte(uint8(len("weapon-native")))
	want.WriteString("weapon-native")
	want.WriteByte(7)
	want.WriteByte(0)
	writeU16(0)
	writeU32(status & objectMapStatusMask4F4530)
	writeU16(1)
	writeU32(0)
	writeU32(handlerFlags)
	writeI32(int32(objectFrame - gameFrame))
	want.Write([]byte{0, 0, 0, 0})
	writeU16(health.Cur)
	writeU32(update.Field4)

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want.Bytes()) {
		t.Fatalf("wire payload = %x, want %x", got, want.Bytes())
	}
	if object.Field34 != objectFrame || object.InitData != unsafe.Pointer(initData) ||
		object.HealthData != health || object.UpdateData != unsafe.Pointer(update) {
		t.Fatalf("native records changed: Field34=%#x init=%p health=%p update=%p",
			object.Field34, object.InitData, object.HealthData, object.UpdateData)
	}
}

func TestWeaponXferNativeRead4F64A0UsesNativeRecordsAndRestoresCount(t *testing.T) {
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
		newCharge      = uint8(5)
		maximumCharge  = uint8(12)
		newProgress    = int32(80)
		serializedHP   = uint16(120)
		newUpdate      = uint32(0x0badc0de)
	)

	var payload bytes.Buffer
	writeU16 := func(value uint16) { _ = binary.Write(&payload, binary.LittleEndian, value) }
	writeU32 := func(value uint32) { _ = binary.Write(&payload, binary.LittleEndian, value) }
	writeI32 := func(value int32) { writeU32(uint32(value)) }
	writeU16(weaponXferCurrentVersion4F64A0)
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
	payload.Write([]byte{0, 0, 0, 0})
	payload.WriteByte(newCharge)
	payload.WriteByte(maximumCharge)
	writeI32(newProgress)
	writeU16(serializedHP)
	writeU32(newUpdate)

	path := filepath.Join(t.TempDir(), "weapon-read.bin")
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
	initData, freeInitData := alloc.New(server.ModifierInitData{Field16: 0x11223344})
	defer freeInitData()
	health, freeHealth := alloc.New(server.HealthData{Cur: 50, Field2: 90, Max: 100})
	defer freeHealth()
	useData, freeUseData := alloc.New(server.WandUseData{Charge: 9, MaxCharge: maximumCharge, Progress: 100})
	defer freeUseData()
	update, freeUpdate := alloc.New(server.WeaponArmorUpdateData{Field0: 0x10203040, Field4: 0xa1b2c3d4})
	defer freeUpdate()
	oldID, freeOldID := alloc.CString("")
	defer freeOldID()
	*initData = server.ModifierInitData{Field16: 0x11223344}
	*health = server.HealthData{Cur: 50, Field2: 90, Max: 100}
	*useData = server.WandUseData{Charge: 9, MaxCharge: maximumCharge, Progress: 100}
	*update = server.WeaponArmorUpdateData{Field0: 0x10203040, Field4: 0xa1b2c3d4}
	object.ObjClass = objectlib.ClassWand | objectlib.ClassClientPersist
	object.ObjSubClass = objectlib.SubClass(0x00010000)
	object.IDPtr = unsafe.Pointer(oldID)
	object.ObjFlags = objectlib.Flags(originalFlags)
	object.Field5 = originalState
	object.Field34 = originalCount
	object.ScriptPickup.Func = -1
	object.InitData = unsafe.Pointer(initData)
	object.HealthData = health
	object.UseData.Ptr = unsafe.Pointer(useData)
	object.UpdateData = unsafe.Pointer(update)

	for name, pointer := range map[string]unsafe.Pointer{
		"object":      unsafe.Pointer(object),
		"init data":   unsafe.Pointer(initData),
		"health data": unsafe.Pointer(health),
		"use data":    unsafe.Pointer(useData),
		"update data": unsafe.Pointer(update),
		"existing ID": unsafe.Pointer(oldID),
	} {
		assertObjectMapNativePointer4F4530(t, name, pointer)
	}

	var applied server.ModifierInitData
	setHPCalls := make([]uint16, 0, 1)
	var inventoryCalls []weaponXferNativeInventoryCall4F64A0
	deps := weaponXferRuntimeDeps4F64A0()
	deps.modifierIDByName = func(name string) int32 {
		if name != "" {
			t.Fatalf("modifier name = %q, want empty", name)
		}
		return -1
	}
	deps.modifierByID = func(id int32) *server.ModifierEff {
		if id != -1 {
			t.Fatalf("modifier ID = %d, want -1", id)
		}
		return nil
	}
	deps.applyModifiers = func(gotObject *server.Object, attrs *server.ModifierInitData) {
		if gotObject != object {
			t.Fatalf("modifier object = %p, want %p", gotObject, object)
		}
		applied = *attrs
	}
	deps.gameFlag4096 = func() bool { return true }
	deps.unitSetHP = func(gotObject *server.Object, value uint16) {
		if gotObject != object {
			t.Fatalf("HP object = %p, want %p", gotObject, object)
		}
		setHPCalls = append(setHPCalls, value)
		gotObject.HealthData.Cur = value
	}
	deps.switchToSolo = func() int32 { return 1 }
	deps.notMultiplayer = func() int32 {
		t.Fatal("notMultiplayer evaluated after switchToSolo succeeded")
		return 0
	}
	deps.anyTrackedPlayers = func() int32 {
		t.Fatal("anyTrackedPlayers evaluated after switchToSolo succeeded")
		return 0
	}
	deps.projectileClass = func(uint16) *server.Modifier {
		t.Fatal("projectile fallback evaluated on direct HP path")
		return nil
	}
	deps.transferInventory = func(version uint16, gotObject *server.Object, count int32) int32 {
		inventoryCalls = append(inventoryCalls, weaponXferNativeInventoryCall4F64A0{
			version: version,
			object:  gotObject,
			count:   count,
		})
		gotObject.Field34 = 0x11223344
		return 1
	}

	if got := weaponXferNative4F64A0(cf, object, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if applied.Field16 != math.MaxUint32 || applied.Modifiers != ([4]*server.ModifierEff{}) {
		t.Fatalf("applied modifiers = %+v, want nil slots and %#x tail", applied, uint32(math.MaxUint32))
	}
	if useData.Charge != newCharge || useData.MaxCharge != maximumCharge || useData.Progress != uint32(newProgress) {
		t.Fatalf("charge record = %d/%d/%d, want %d/%d/%d",
			useData.Charge, useData.MaxCharge, useData.Progress,
			newCharge, maximumCharge, newProgress)
	}
	if len(setHPCalls) != 1 || setHPCalls[0] != health.Max || health.Cur != health.Max {
		t.Fatalf("HP calls/current = %v/%d, want [%d]/%d", setHPCalls, health.Cur, health.Max, health.Max)
	}
	if update.Field0 != 0x10203040 || update.Field4 != newUpdate {
		t.Fatalf("update record = %#x/%#x, want %#x/%#x", update.Field0, update.Field4, uint32(0x10203040), newUpdate)
	}
	if len(inventoryCalls) != 1 || inventoryCalls[0] != (weaponXferNativeInventoryCall4F64A0{
		version: weaponXferCurrentVersion4F64A0,
		object:  object,
		count:   int32(inventoryCount),
	}) {
		t.Fatalf("inventory calls = %+v", inventoryCalls)
	}
	if object.Field34 != originalCount {
		t.Fatalf("Field34 = %#x, want entry %#x", object.Field34, originalCount)
	}
	if object.IDPtr != unsafe.Pointer(oldID) || object.InitData != unsafe.Pointer(initData) ||
		object.HealthData != health || object.UseData.Ptr != unsafe.Pointer(useData) ||
		object.UpdateData != unsafe.Pointer(update) {
		t.Fatalf("native object-owned pointers changed")
	}
}

func TestWeaponXferExport4F64A0PreservesNativePointerAndResult(t *testing.T) {
	object, freeObject := alloc.New(server.Object{})
	defer freeObject()
	assertObjectMapNativePointer4F4530(t, "object", unsafe.Pointer(object))

	old := weaponXferCall4F64A0
	t.Cleanup(func() { weaponXferCall4F64A0 = old })
	calls := 0
	weaponXferCall4F64A0 = func(_ *cryptfile.CryptFile, gotObject *server.Object) int32 {
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

	if got := weaponXferExportCall4F64A0(object); got != math.MinInt32 {
		t.Fatalf("first result = %d, want %d", got, int32(math.MinInt32))
	}
	if got := weaponXferExportCall4F64A0(nil); got != math.MaxInt32 {
		t.Fatalf("second result = %d, want %d", got, int32(math.MaxInt32))
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
	runtime.KeepAlive(object)
}
