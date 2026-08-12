package server

import (
	"bytes"
	"math"
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestObjectSyncDataMatchesGAMEEXEContract(t *testing.T) {
	tests := []struct {
		name  string
		key   uint32
		setup func(obj *Object, typ *ObjectType)
		want  bool
	}{
		{name: "animation frame zero", key: 0x1},
		{name: "animation frame nonzero", key: 0x1, setup: func(obj *Object, _ *ObjectType) { obj.Field33 = 0x80000000 }, want: true},
		{name: "object health missing", key: 0x2, setup: func(obj *Object, _ *ObjectType) { obj.HealthData = nil }},
		{name: "type health missing", key: 0x2, setup: func(_ *Object, typ *ObjectType) { typ.health = nil }},
		{name: "health equal", key: 0x2},
		{name: "health differs", key: 0x2, setup: func(obj *Object, _ *ObjectType) { obj.HealthData.Cur++ }, want: true},
		{name: "enabled flag equal", key: 0x4},
		{name: "enabled flag differs", key: 0x4, setup: func(obj *Object, _ *ObjectType) { obj.ObjFlags &^= object.FlagEnabled }, want: true},
		{name: "other flag differs", key: 0x4, setup: func(_ *Object, typ *ObjectType) { typ.flags |= object.FlagPending }},
		{name: "extended status equal", key: 0x8},
		{name: "extended status differs", key: 0x8, setup: func(obj *Object, _ *ObjectType) { obj.Field5 ^= 0x80000000 }, want: true},
		{name: "height positive zero", key: 0x40},
		{name: "height negative zero", key: 0x40, setup: func(obj *Object, _ *ObjectType) { obj.ZVal = math.Float32frombits(0x80000000) }},
		{name: "height positive", key: 0x40, setup: func(obj *Object, _ *ObjectType) { obj.ZVal = 1 }, want: true},
		{name: "height negative", key: 0x40, setup: func(obj *Object, _ *ObjectType) { obj.ZVal = -1 }, want: true},
		{name: "height infinity", key: 0x40, setup: func(obj *Object, _ *ObjectType) { obj.ZVal = float32(math.Inf(1)) }, want: true},
		{name: "height NaN is unordered", key: 0x40, setup: func(obj *Object, _ *ObjectType) { obj.ZVal = math.Float32frombits(0x7FC00001) }},
		{name: "buffs empty", key: 0x80},
		{name: "buffs nonempty", key: 0x80, setup: func(obj *Object, _ *ObjectType) { obj.Buffs = 0x80000000 }, want: true},
		{name: "equipment unrelated class", key: 0x200},
		{name: "equipment wand", key: 0x200, setup: func(obj *Object, _ *ObjectType) { obj.ObjClass = object.ClassWand }, want: true},
		{name: "equipment weapon", key: 0x200, setup: func(obj *Object, _ *ObjectType) { obj.ObjClass = object.ClassWeapon }, want: true},
		{name: "equipment armor", key: 0x200, setup: func(obj *Object, _ *ObjectType) { obj.ObjClass = object.ClassArmor }, want: true},
		{name: "equipment flag", key: 0x200, setup: func(obj *Object, _ *ObjectType) { obj.ObjClass = object.ClassFlag }, want: true},
		{name: "monster without NPC subclass", key: 0x400, setup: func(obj *Object, _ *ObjectType) { obj.ObjClass = object.ClassMonster }},
		{name: "NPC", key: 0x400, setup: func(obj *Object, _ *ObjectType) {
			obj.ObjClass, obj.ObjSubClass = object.ClassMonster, object.SubClass(object.MonsterNPC)
		}, want: true},
		{name: "female NPC", key: 0x400, setup: func(obj *Object, _ *ObjectType) {
			obj.ObjClass, obj.ObjSubClass = object.ClassMonster, object.SubClass(object.MonsterFemaleNPC)
		}, want: true},
		{name: "NPC subclass without monster class", key: 0x400, setup: func(obj *Object, _ *ObjectType) { obj.ObjSubClass = object.SubClass(object.MonsterNPC) }},
		{name: "unsupported zero", key: 0},
		{name: "unsupported combined key", key: 0x3},
		{name: "unsupported table entry", key: 0x10},
		{name: "unsupported table boundary", key: 0x20},
		{name: "unsupported middle key", key: 0x100},
		{name: "unsupported high key", key: 0x401},
		{name: "unsupported maximum key", key: math.MaxUint32},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			objHealth := &HealthData{Cur: 0x2468, Max: 0xACE0, Field16: 0x13579BDF}
			typeHealth := &HealthData{Cur: objHealth.Cur, Max: 0xFFFF, Field16: 0x89ABCDEF}
			typ := &ObjectType{
				flags:  object.FlagEnabled,
				Field9: 0x10203040,
				health: typeHealth,
			}
			obj := &Object{
				TypeInd:    1,
				ObjClass:   object.ClassMissile,
				ObjFlags:   object.FlagEnabled,
				Field5:     typ.Field9,
				Field32:    0x11223344,
				Field34:    0x55667788,
				Field110:   0xA5A5A5A5,
				HealthData: objHealth,
				InitData:   unsafe.Pointer(typeHealth),
				UpdateData: unsafe.Pointer(objHealth),
			}
			bindSyncDataTestType(t, obj, typ)
			if tc.setup != nil {
				tc.setup(obj, typ)
			}

			before := bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(obj)), unsafe.Sizeof(*obj)))
			if got := obj.HasSyncData(tc.key); got != tc.want {
				t.Fatalf("HasSyncData(%#x) = %t, want %t", tc.key, got, tc.want)
			}
			after := unsafe.Slice((*byte)(unsafe.Pointer(obj)), unsafe.Sizeof(*obj))
			if !bytes.Equal(after, before) {
				t.Fatal("sync-data query modified the object")
			}
		})
	}
}

func bindSyncDataTestType(t *testing.T, obj *Object, typ *ObjectType) {
	t.Helper()
	h := atomic.AddUintptr(&serverLast, 1)
	s := &Server{handle: h}
	s.Types.byInd = []*ObjectType{nil, typ}
	servers.Store(h, s)
	t.Cleanup(s.Close)
	obj.serverHandle = h
}
