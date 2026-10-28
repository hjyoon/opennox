package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func TestToxicCloudRawMissingDamageCallback53D8C0(t *testing.T) {
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		toxicCloudRawRuntime53D850().Damage(nil, nil, nil, nil, 0, object.DamagePoison)
	}()
	if _, ok := recovered.(interface{ RuntimeError() }); !ok {
		t.Fatalf("raw nil indirect-call fault = %v", recovered)
	}
}

func TestToxicCloudNativeDamageDispatch53D8C0(t *testing.T) {
	newObject := func() *server.Object {
		frame := radialCFrameOriginal(t, unsafe.Sizeof(server.Object{}))
		o := (*server.Object)(unsafe.Pointer(&frame[32]))
		*o = server.Object{}
		return o
	}
	cloud, target, owner, newOwner := newObject(), newObject(), newObject(), newObject()
	frame := radialCFrameOriginal(t, unsafe.Sizeof(server.ToxicCloudUpdateData{}))
	data := (*server.ToxicCloudUpdateData)(unsafe.Pointer(&frame[32]))
	*data = server.ToxicCloudUpdateData{Duration: 2}
	cloud.UpdateData = unsafe.Pointer(data)
	cloud.Field34 = 9
	cloud.ObjOwner = owner
	target.ObjClass = object.ClassPlayer
	fn0, fn1 := toxicCloudDamageProbePtr53D850(0), toxicCloudDamageProbePtr53D850(1)
	if fn0 == nil || fn1 == nil || fn0 == fn1 {
		t.Fatal("distinct native damage callback identities")
	}
	target.Damage = fn0
	calls := 0
	server.RegisterObjectDamageGo("ToxicCloudNativeDamageProbe0", fn0, func(v, s, w *server.Object, d int32, typ object.DamageType) bool {
		if calls != 0 || v != target || s != owner || w != cloud || d != 7 || typ != object.DamagePoison {
			t.Fatal("first native damage arguments/order")
		}
		calls++
		cloud.ObjOwner = newOwner
		target.Damage = fn1
		return false
	})
	server.RegisterObjectDamageGo("ToxicCloudNativeDamageProbe1", fn1, func(v, s, w *server.Object, d int32, typ object.DamageType) bool {
		if calls != 1 || v != target || s != newOwner || w != cloud || d != 7 || typ != object.DamagePoison {
			t.Fatal("reloaded native damage arguments/order")
		}
		calls++
		return false
	})
	defaultDamage := toxicCloudRawRuntime53D850().Damage
	now := uint32(10)
	ownerReads, poisons, circles := 0, 0, 0
	h := server.ToxicCloudRawRuntime53D850{
		Frame: func() uint32 { return now },
		EachInCircle: func(p *types.Pointf, r float32, visit func(*server.Object)) {
			if p != &cloud.PosVec || r != 75 {
				t.Fatal("native circle center/radius")
			}
			circles++
			visit(target)
			visit(target)
		},
		RandomInt: func(a, b int) int {
			if a == 3 && b == 10 {
				return 7
			}
			if a != 5 || b != 10 {
				t.Fatal("random bounds")
			}
			now = 20
			return 6
		},
		TraceRay: func(*types.Rectf, server.MapTraceFlags) int32 { return 1 },
		FindOwnerPlayer: func(o *server.Object) *server.Object {
			if o != cloud {
				t.Fatal("owner source")
			}
			ownerReads++
			return o.ObjOwner
		},
		Damage: defaultDamage,
		IsEnemy: func(s, v *server.Object) int32 {
			if s != newOwner || v != target {
				t.Fatal("post-damage owner reload")
			}
			return 1
		},
		ActivatePoison: func(v *server.Object, a, b int32) int32 {
			if v != target || a != 1 || b != 1 {
				t.Fatal("poison arguments")
			}
			poisons++
			return 1
		},
		DelayedDelete: func(*server.Object) { t.Fatal("live cloud deleted") },
	}
	old := toxicCloudRawRuntime53D850
	toxicCloudRawRuntime53D850 = func() server.ToxicCloudRawRuntime53D850 { return h }
	defer func() { toxicCloudRawRuntime53D850 = old }()
	toxicCloudNativeProbe53D850(0, cloud, target)
	if calls != 2 || poisons != 2 || ownerReads != 4 || circles != 1 || cloud.Field34 != 26 || data.Duration != 1 {
		t.Fatalf("native dispatch calls=%d poison=%d owners=%d circle=%d next=%d duration=%d", calls, poisons, ownerReads, circles, cloud.Field34, data.Duration)
	}
}
