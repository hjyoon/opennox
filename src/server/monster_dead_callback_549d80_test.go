package server

import (
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

func TestMonsterDeadCallback549D80ExplosionAndSimpleCallbacks(t *testing.T) {
	unit := &Object{PosVec: types.Ptf(12, 34)}
	var events []string
	runtime := MonsterDeadCallbackRuntime549D80{
		CoopMode: func() bool {
			events = append(events, "coop")
			return true
		},
		PushUnits: func(pos types.Pointf, outer, inner, force float32, source *Object) {
			events = append(events, fmt.Sprintf("push:%v:%g:%g:%g:%t", pos, outer, inner, force, source == unit))
		},
		DamageUnits: func(pos types.Pointf, outer, inner float32, damage int, damageType object.DamageType, source *Object) {
			events = append(events, fmt.Sprintf("damage:%v:%g:%g:%d:%d:%t", pos, outer, inner, damage, damageType, source == unit))
		},
		SparkExplosion: func(pos types.Pointf, intensity byte) {
			events = append(events, fmt.Sprintf("spark:%v:%d", pos, intensity))
		},
		PointFX: func(effect netmsg.Op, pos types.Pointf) {
			events = append(events, fmt.Sprintf("point:%d:%v", effect, pos))
		},
		Audio: func(id sound.ID, got *Object) {
			events = append(events, fmt.Sprintf("audio:%d:%t", id, got == unit))
		},
		DelayedDelete: func(got *Object) {
			events = append(events, fmt.Sprintf("delete:%t", got == unit))
		},
		BomberDead: func(got *Object) {
			events = append(events, fmt.Sprintf("bomber:%t", got == unit))
		},
	}
	if !MonsterDeadCallbackNative549D80(unit, MonsterDeadCallbackEmberDemon549D80, runtime) {
		t.Fatal("EmberDemon callback was not handled")
	}
	want := []string{
		"coop",
		"push:{12 34}:96:10:100:true",
		"damage:{12 34}:96:10:30:7:true",
		"spark:{12 34}:128",
		"audio:42:true",
		"delete:true",
	}
	if fmt.Sprint(events) != fmt.Sprint(want) {
		t.Fatalf("EmberDemon events = %v, want %v", events, want)
	}

	events = nil
	if !MonsterDeadCallbackNative549D80(unit, MonsterDeadCallbackDemon549E00, runtime) {
		t.Fatal("Demon callback was not handled")
	}
	want = []string{
		"push:{12 34}:150:10:150:true",
		"damage:{12 34}:150:10:148:7:true",
		"spark:{12 34}:255",
		"audio:42:true",
		"delete:true",
	}
	if fmt.Sprint(events) != fmt.Sprint(want) {
		t.Fatalf("Demon events = %v, want %v", events, want)
	}

	for _, kind := range []MonsterDeadCallbackKind549D80{MonsterDeadCallbackImp549E70, MonsterDeadCallbackSpider54A250} {
		events = nil
		if !MonsterDeadCallbackNative549D80(unit, kind, runtime) ||
			fmt.Sprint(events) != fmt.Sprint([]string{"point:129:{12 34}"}) {
			t.Fatalf("point callback %d events = %v", kind, events)
		}
	}
	events = nil
	if !MonsterDeadCallbackNative549D80(unit, MonsterDeadCallbackBomber54A150, runtime) ||
		fmt.Sprint(events) != fmt.Sprint([]string{"bomber:true"}) {
		t.Fatalf("Bomber events = %v", events)
	}
}

func monsterDeadDebrisTestRuntime549D80(
	t *testing.T,
	unit *Object,
	names *[]string,
	objects *[]*Object,
) MonsterDeadCallbackRuntime549D80 {
	t.Helper()
	return MonsterDeadCallbackRuntime549D80{
		RandomInt: func(minimum, _ int) int { return minimum },
		RandomFloat: func(minimum, _ float32) float32 {
			return minimum
		},
		PointFX: func(netmsg.Op, types.Pointf) {},
		Audio:   func(sound.ID, *Object) {},
		NewObjectByTypeID: func(name string) *Object {
			*names = append(*names, name)
			obj := new(Object)
			*objects = append(*objects, obj)
			return obj
		},
		RandomReachablePoint: func(radius float32, center types.Pointf) types.Pointf {
			return center.Add(types.Ptf(radius, -radius))
		},
		CreateObjectAt: func(obj, owner *Object, position types.Pointf) {
			if owner != nil {
				t.Fatalf("debris owner = %p, want nil", owner)
			}
			obj.PosVec = position
		},
		Raise: func(obj *Object, height float32) { obj.ZVal = height },
		ApplyForce: func(obj *Object, origin types.Pointf, force float64) {
			if origin != unit.PosVec {
				t.Fatalf("force origin = %v, want %v", origin, unit.PosVec)
			}
			obj.Mass = float32(force)
		},
		DecaySetTime: func(obj *Object, delay uint32) { obj.Field32 = delay },
		TickRate:     func() uint32 { return 30 },
	}
}

func TestMonsterDeadCallbackMechGolem549E90Debris(t *testing.T) {
	unit := &Object{PosVec: types.Ptf(50, 60)}
	var names []string
	var objects []*Object
	runtime := monsterDeadDebrisTestRuntime549D80(t, unit, &names, &objects)
	var gotSound sound.ID
	var gotFX netmsg.Op
	runtime.Audio = func(id sound.ID, got *Object) {
		if got != unit {
			t.Fatalf("audio object = %p", got)
		}
		gotSound = id
	}
	runtime.PointFX = func(effect netmsg.Op, pos types.Pointf) {
		if pos != unit.PosVec {
			t.Fatalf("FX position = %v", pos)
		}
		gotFX = effect
	}
	if !MonsterDeadCallbackNative549D80(unit, MonsterDeadCallbackMechGolem549E90, runtime) {
		t.Fatal("MechGolem callback was not handled")
	}
	wantNames := []string{"MechGear", "MechHead", "MechArm", "MechGear", "MechLeg", "MechLeg", "MechGear", "MechGear"}
	if fmt.Sprint(names) != fmt.Sprint(wantNames) {
		t.Fatalf("part names = %v, want %v", names, wantNames)
	}
	if gotSound != sound.SoundMechGolemDie || gotFX != netmsg.MSG_FX_SMOKE_BLAST {
		t.Fatalf("sound/FX = %d/%d", gotSound, gotFX)
	}
	for i, obj := range objects {
		if obj.PosVec != (types.Ptf(80, 30)) || obj.ZVal != 10 || obj.Field27 != -2 ||
			obj.Field29 != math.Float32bits(2) || !obj.ObjFlags.Has(object.FlagBouncy) || obj.Field32 != 300 {
			t.Fatalf("part %d = %#v", i, obj)
		}
	}
}

func TestMonsterDeadCallbackGolem549FA0CyclesParts(t *testing.T) {
	unit := &Object{PosVec: types.Ptf(10, 20)}
	var names []string
	var objects []*Object
	runtime := monsterDeadDebrisTestRuntime549D80(t, unit, &names, &objects)
	coopCalls := 0
	runtime.CoopMode = func() bool {
		coopCalls++
		return false
	}
	index := uint32(11)
	runtime.GolemPartIndex = func() uint32 { return index }
	runtime.SetGolemPartIndex = func(next uint32) { index = next }
	if !MonsterDeadCallbackNative549D80(unit, MonsterDeadCallbackGolem549FA0, runtime) {
		t.Fatal("Golem callback was not handled")
	}
	wantNames := []string{"SmallRock", "BigRock", "MediumRock", "SmallRock", "MediumRock", "SmallRock"}
	if fmt.Sprint(names) != fmt.Sprint(wantNames) {
		t.Fatalf("part names = %v, want %v", names, wantNames)
	}
	wantHeights := []float32{5, 1, 3, 5, 3, 5}
	for i, obj := range objects {
		if obj.Field29 != math.Float32bits(wantHeights[i]) || obj.Mass != 5 || obj.Field32 != 150 ||
			!obj.ObjFlags.Has(object.FlagBouncy) {
			t.Fatalf("part %d = %#v", i, obj)
		}
	}
	if index != 5 || coopCalls != 7 {
		t.Fatalf("index/coop calls = %d/%d, want 5/7", index, coopCalls)
	}
}

func TestMonsterDeadCallbackSkeleton54A310DebrisAndDrops(t *testing.T) {
	unit := &Object{PosVec: types.Ptf(2, 3)}
	var names []string
	var objects []*Object
	runtime := monsterDeadDebrisTestRuntime549D80(t, unit, &names, &objects)
	intCalls := 0
	runtime.RandomInt = func(minimum, _ int) int {
		intCalls++
		if intCalls == 1 {
			return 20
		}
		return minimum
	}
	runtime.CoopMode = func() bool { return false }
	index := uint32(1)
	runtime.SkeletonPartIndex = func() uint32 { return index }
	runtime.SetSkeletonPartIndex = func(next uint32) { index = next }
	var drops []MonsterDieDrop54A390
	runtime.DropItem = func(*Object, MonsterDieDrop54A390) {
		drops = append(drops, MonsterDieDrop54A390{})
	}
	if !MonsterDeadCallbackNative549D80(unit, MonsterDeadCallbackSkeleton54A310, runtime) {
		t.Fatal("Skeleton callback was not handled")
	}
	wantNames := []string{"Skull", "LegBone", "ArmBone", "LegBone", "ArmBone", "LegBone"}
	if fmt.Sprint(names) != fmt.Sprint(wantNames) {
		t.Fatalf("debris names = %v, want %v", names, wantNames)
	}
	if len(drops) != 0 || index != 0 {
		t.Fatalf("drops/index = %v/%d", drops, index)
	}
	if objects[0].ZVal != 40 || objects[0].Field32 != 60 || objects[0].Field29 != math.Float32bits(4) {
		t.Fatalf("skull = %#v", objects[0])
	}
	for i, obj := range objects[1:] {
		if obj.ZVal != 10 || obj.Field27 != -2 || obj.Field29 != math.Float32bits(4) ||
			obj.Mass != 5 || obj.Field32 != 60 || !obj.ObjFlags.Has(object.FlagBouncy) {
			t.Fatalf("bone %d = %#v", i, obj)
		}
	}

	for _, test := range []struct {
		kind MonsterDeadCallbackKind549D80
		roll int
		want MonsterDieDrop54A390
	}{
		{
			kind: MonsterDeadCallbackSkeleton54A310,
			roll: 21,
			want: MonsterDieDrop54A390{TypeID: "Sword", Modifiers: [4]string{"WeaponPower1", "Material2"}},
		},
		{
			kind: MonsterDeadCallbackSkeletonLord54A750,
			roll: 51,
			want: MonsterDieDrop54A390{TypeID: "SteelShield", Modifiers: [4]string{"", "Material2"}},
		},
	} {
		t.Run(fmt.Sprintf("kind-%d-roll-%d", test.kind, test.roll), func(t *testing.T) {
			calls := 0
			rt := runtime
			rt.RandomInt = func(int, int) int { return test.roll }
			rt.CoopMode = func() bool { return true }
			rt.NewObjectByTypeID = func(string) *Object {
				calls++
				return nil
			}
			var got MonsterDieDrop54A390
			rt.DropItem = func(gotUnit *Object, drop MonsterDieDrop54A390) {
				if gotUnit != unit {
					t.Fatalf("drop unit = %p", gotUnit)
				}
				got = drop
			}
			if !MonsterDeadCallbackNative549D80(unit, test.kind, rt) {
				t.Fatal("skeleton drop callback was not handled")
			}
			if calls != 1 || got != test.want {
				t.Fatalf("allocation/drop = %d/%#v, want 1/%#v", calls, got, test.want)
			}
		})
	}
}

func TestMonsterDeadCallbackTroll54A270Cloud(t *testing.T) {
	unit := &Object{PosVec: types.Ptf(7, 8)}
	data := new(ToxicCloudUpdateData)
	cloud := &Object{UpdateData: unsafe.Pointer(data)}
	var createdOwner *Object
	var createdPosition types.Pointf
	var gotAudio sound.ID
	runtime := MonsterDeadCallbackRuntime549D80{
		NewObjectByTypeID: func(typeID string) *Object {
			if typeID != "SmallToxicCloud" {
				t.Fatalf("cloud type = %q", typeID)
			}
			return cloud
		},
		CreateObjectAt: func(got, owner *Object, position types.Pointf) {
			if got != cloud {
				t.Fatalf("created cloud = %p", got)
			}
			createdOwner, createdPosition = owner, position
		},
		Audio: func(id sound.ID, got *Object) {
			if got != unit {
				t.Fatalf("audio object = %p", got)
			}
			gotAudio = id
		},
		BalanceFloat: func(key string) float64 {
			if key != "SmallToxicCloudLifetime" {
				t.Fatalf("balance key = %q", key)
			}
			return 1.25
		},
		TickRate: func() uint32 { return 30 },
	}
	if !MonsterDeadCallbackNative549D80(unit, MonsterDeadCallbackTroll54A270, runtime) {
		t.Fatal("Troll callback was not handled")
	}
	if createdOwner != unit || createdPosition != unit.PosVec || gotAudio != sound.SoundTrollFlatus || data.Duration != 38 {
		t.Fatalf("cloud = owner:%p pos:%v audio:%d duration:%d", createdOwner, createdPosition, gotAudio, data.Duration)
	}
}

func TestMonsterDeadCallback549D80RejectsInvalidInputs(t *testing.T) {
	unit := new(Object)
	if MonsterDeadCallbackNative549D80(nil, MonsterDeadCallbackImp549E70, MonsterDeadCallbackRuntime549D80{}) {
		t.Fatal("nil unit was accepted")
	}
	if MonsterDeadCallbackNative549D80(unit, 0, MonsterDeadCallbackRuntime549D80{}) {
		t.Fatal("invalid callback kind was accepted")
	}
	if MonsterDeadCallbackNative549D80(unit, MonsterDeadCallbackImp549E70, MonsterDeadCallbackRuntime549D80{}) {
		t.Fatal("missing point-FX callback was accepted")
	}
	if MonsterDeadCallbackNative549D80(unit, MonsterDeadCallbackMechGolem549E90, MonsterDeadCallbackRuntime549D80{}) {
		t.Fatal("missing debris callbacks were accepted")
	}
}
