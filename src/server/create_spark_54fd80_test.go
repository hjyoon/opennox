package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestCreateSpark54FD80NativeDataAndKinds(t *testing.T) {
	for _, tc := range []struct {
		kind      int
		class     object.Class
		flags     object.Flags
		collision uint32
		field29   float32
	}{
		{0, 0x100, 0x41, 0, 0},
		{1, 0x82100, 0x800001, 3, 7},
		{2, 0x80100, 0x800041, 0, 7},
		{3, 0x80100, 0x1, 0, 0},
		{4, 0x100, 0x1, 0, 0},
	} {
		t.Run(string(rune('0'+tc.kind)), func(t *testing.T) {
			data := new(SparkUpdateData)
			collision := new(uint32)
			spark := &Object{
				ObjClass:    0x2100,
				ObjFlags:    0x800041,
				UpdateData:  unsafe.Pointer(data),
				CollideData: unsafe.Pointer(collision),
			}
			owner := new(Object)
			pos := types.Ptf(12, 34)
			velocity := types.Ptf(-2, 3)
			created, raised := 0, 0
			got := CreateSpark54FD80(SparkCreateRuntime54FD80{
				NewObject: func(id string) *Object {
					if id != "Spark" {
						t.Fatalf("type ID = %q", id)
					}
					return spark
				},
				CreateAt: func(obj, gotOwner *Object, gotPos types.Pointf) {
					if obj != spark || gotOwner != owner || gotPos != pos {
						t.Fatalf("create at %p/%p/%v", obj, gotOwner, gotPos)
					}
					created++
				},
				Frame: func() uint32 { return 123 },
				Raise: func(obj *Object, z float32) {
					if obj != spark || z != 28 {
						t.Fatalf("raise %p/%g", obj, z)
					}
					raised++
				},
			}, pos, tc.kind, 25, velocity, 9, owner)
			if got != spark || created != 1 || raised != 1 {
				t.Fatalf("object/create/raise = %p/%d/%d", got, created, raised)
			}
			if spark.Field34 != 123 || data.LifetimeInitial != 25 || data.LifetimeRemaining != 25 || data.Kind != uint32(tc.kind) {
				t.Fatalf("frame/update = %d/%+v", spark.Field34, *data)
			}
			if spark.ObjClass != tc.class || spark.ObjFlags != tc.flags || *collision != tc.collision ||
				spark.Field27 != 9 || spark.Field29 != math.Float32bits(tc.field29) || spark.VelVec != velocity {
				t.Fatalf("spark class/flags/collision/z/field29/velocity = %#x/%#x/%d/%g/%#x/%v", spark.ObjClass, spark.ObjFlags, *collision, spark.Field27, spark.Field29, spark.VelVec)
			}
		})
	}
}

func TestCreateSpark54FD80MissingObjectOrRecords(t *testing.T) {
	for _, candidate := range []*Object{nil, {}, {UpdateData: unsafe.Pointer(new(SparkUpdateData))}, {CollideData: unsafe.Pointer(new(uint32))}} {
		created := false
		got := CreateSpark54FD80(SparkCreateRuntime54FD80{
			NewObject: func(string) *Object { return candidate },
			CreateAt:  func(*Object, *Object, types.Pointf) { created = true },
		}, types.Pointf{}, 1, 20, types.Pointf{}, 0, nil)
		if got != nil || created {
			t.Fatalf("invalid candidate %p created %p / %v", candidate, got, created)
		}
	}
}
