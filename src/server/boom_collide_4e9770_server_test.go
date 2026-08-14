package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/common/sound"
)

func defaultBoomCollideNativeDeps4E9770(balance *boomCollideBalance4E9770) boomCollideNativeDeps4E9770 {
	return boomCollideNativeDeps4E9770{
		balance:        balance,
		gameDataFloat:  func(string) float64 { return 0 },
		floatToInt:     playerCollideRound4E8460,
		gameFlagsCheck: func(uint32) int32 { return 0 },
		findParent:     func(obj *Object) *Object { return obj },
		isEnemy:        func(*Object, *Object) int32 { return 1 },
		pointFX:        func(uint32, types.Pointf) {},
		inversion:      func(*Object, *Object) int32 { return 0 },
		changeOwner:    func(*Object, *Object) {},
		checkDirection: func(types.Pointf, int16, types.Pointf) int32 { return 0 },
		audio:          func(uint32, *Object, int32, uint32) {},
		targetDamage:   func(*Object, *Object, *Object, int32, object.DamageType) int32 { return 0 },
		scorch:         func(types.Pointf, int32) {},
		wallReflect:    spellProjectileWallReflect57B810,
		traceHitPoint:  func() *ntype.Point32 { return nil },
		damageMap:      func(int32, int32, int32, object.DamageType, *Object) {},
		mapDamageUnits: func(types.Pointf, float32, float32, int32, object.DamageType, *Object, *Object) {},
		mapPushUnits:   func(types.Pointf, float32, float32, float32, *Object, int32, int32) {},
		delayedDelete:  func(*Object) {},
	}
}

func TestBoomCollide4E9770NativeInitializesCacheAndUsesLiveCommonState(t *testing.T) {
	source := &Object{PosVec: types.Ptf(10, 20)}
	cache := &boomCollideBalance4E9770{}
	deps := defaultBoomCollideNativeDeps4E9770(cache)
	values := map[string]float64{
		boomCollideDamageBalance4E9770:    11.5,
		boomCollideSplashBalance4E9770:    22.5,
		boomCollideRangeBalance4E9770:     33.25,
		boomCollidePushRangeBalance4E9770: 44.5,
		boomCollideForceBalance4E9770:     55.75,
	}
	var events []string
	deps.gameDataFloat = func(name string) float64 {
		events = append(events, "balance:"+name)
		return values[name]
	}
	deps.pointFX = func(id uint32, pos types.Pointf) {
		if id != uint32(netmsg.MSG_FX_LESSER_EXPLOSION) || pos != (types.Ptf(10, 20)) {
			t.Fatalf("point FX = (%d, %v)", id, pos)
		}
		events = append(events, "point-fx")
	}
	deps.mapDamageUnits = func(pos types.Pointf, radius, inner float32, damage int32, damageType object.DamageType, gotSource, excluded *Object) {
		if pos != (types.Ptf(10, 20)) || radius != 33.25 || inner != 5 || damage != 22 || damageType != object.DamageExplosion || gotSource != source || excluded != nil {
			t.Fatalf("map damage args = (%v,%g,%g,%d,%v,%p,%p)", pos, radius, inner, damage, damageType, gotSource, excluded)
		}
		events = append(events, "map-damage")
		source.PosVec = types.Ptf(-3, 7)
		cache.Force = 66.5
		cache.PushRange = 77.25
	}
	deps.mapPushUnits = func(pos types.Pointf, first, second, force float32, gotSource *Object, arg6, arg7 int32) {
		if pos != (types.Ptf(-3, 7)) || first != 77.25 || second != 77.25 || force != 66.5 || gotSource != source || arg6 != 0 || arg7 != 0 {
			t.Fatalf("map push args = (%v,%g,%g,%g,%p,%d,%d)", pos, first, second, force, gotSource, arg6, arg7)
		}
		events = append(events, "map-push")
	}
	deps.audio = func(id uint32, obj *Object, kind int32, code uint32) {
		if id != uint32(sound.SoundMagicMissileDetonate) || obj != source || kind != 0 || code != 0 {
			t.Fatalf("audio args = (%d,%p,%d,%d)", id, obj, kind, code)
		}
		events = append(events, "audio")
	}
	deps.delayedDelete = func(obj *Object) {
		if obj != source {
			t.Fatalf("delete object = %p, want %p", obj, source)
		}
		events = append(events, "delete")
	}

	boomCollideNative4E9770(source, nil, nil, deps)
	wantEvents := []string{
		"balance:MagicMissileDamage", "balance:MagicMissileSplashDamage",
		"balance:MagicMissileRange", "balance:MagicMissilePushRange", "balance:MagicMissileForce",
		"point-fx", "map-damage", "map-push", "audio", "delete",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %q, want %q", events, wantEvents)
	}
	if cache.Ready != 1 || cache.DirectDamage != 12 || cache.SplashDamage != 22 || cache.Range != 33.25 || cache.PushRange != 77.25 || cache.Force != 66.5 {
		t.Fatalf("cache = %+v", *cache)
	}
}

func TestBoomCollide4E9770NativePlayerEnchantUsesLiveFields(t *testing.T) {
	source := &Object{PosVec: types.Ptf(1, 2)}
	target := &Object{
		ObjClass:   object.ClassPlayer,
		PosVec:     types.Ptf(30, 40),
		Direction1: Dir16(0xff81),
		Buffs:      uint32(1) << boomCollideInversionEnchant4E9770,
	}
	cache := &boomCollideBalance4E9770{Ready: 1}
	deps := defaultBoomCollideNativeDeps4E9770(cache)
	var events []string
	deps.pointFX = func(uint32, types.Pointf) { events = append(events, "fx") }
	deps.inversion = func(gotTarget, gotSource *Object) int32 {
		if gotTarget != target || gotSource != source {
			t.Fatal("inversion object arguments changed")
		}
		events = append(events, "inversion")
		return 0
	}
	deps.checkDirection = func(first types.Pointf, direction int16, second types.Pointf) int32 {
		if first != target.PosVec || direction != -127 || second != source.PosVec {
			t.Fatalf("direction args = (%v,%d,%v)", first, direction, second)
		}
		events = append(events, "direction")
		return 3
	}
	deps.changeOwner = func(gotSource, gotTarget *Object) {
		if gotSource != source || gotTarget != target {
			t.Fatal("change-owner object arguments changed")
		}
		events = append(events, "owner")
	}
	deps.audio = func(id uint32, obj *Object, kind int32, code uint32) {
		if id != boomCollideReflectAudio4E9770 || obj != target || kind != 0 || code != 0 {
			t.Fatal("reflection audio arguments changed")
		}
		events = append(events, "audio")
	}
	deps.targetDamage = func(*Object, *Object, *Object, int32, object.DamageType) int32 {
		t.Fatal("reflected projectile applied damage")
		return 0
	}

	boomCollideNative4E9770(source, target, nil, deps)
	if want := []string{"fx", "inversion", "direction", "owner", "audio"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %q, want %q", events, want)
	}
}

func TestBoomCollide4E9770NativeDamageCallbackThenLiveScorchPosition(t *testing.T) {
	source := &Object{PosVec: types.Ptf(1, 2)}
	owner := &Object{}
	target := &Object{ObjClass: object.ClassImmobile, PosVec: types.Ptf(3, 4)}
	cache := &boomCollideBalance4E9770{Ready: 1, DirectDamage: 47, SplashDamage: 5, Range: 6, PushRange: 7, Force: 8}
	deps := defaultBoomCollideNativeDeps4E9770(cache)
	var events []string
	deps.findParent = func(got *Object) *Object {
		if got != source {
			t.Fatal("find-parent source changed")
		}
		events = append(events, "parent")
		return owner
	}
	deps.targetDamage = func(gotTarget, gotParent, gotSource *Object, damage int32, damageType object.DamageType) int32 {
		if gotTarget != target || gotParent != owner || gotSource != source || damage != 47 || damageType != object.DamageExplosion {
			t.Fatal("target damage arguments changed")
		}
		events = append(events, "damage")
		target.PosVec = types.Ptf(9, 10)
		return -1
	}
	deps.scorch = func(pos types.Pointf, kind int32) {
		if pos != (types.Ptf(9, 10)) || kind != 0 {
			t.Fatalf("scorch = (%v,%d), want ((9,10),0)", pos, kind)
		}
		events = append(events, "scorch")
	}
	deps.mapDamageUnits = func(types.Pointf, float32, float32, int32, object.DamageType, *Object, *Object) {
		events = append(events, "map-damage")
	}
	deps.mapPushUnits = func(types.Pointf, float32, float32, float32, *Object, int32, int32) {
		events = append(events, "map-push")
	}
	deps.audio = func(uint32, *Object, int32, uint32) { events = append(events, "audio") }
	deps.delayedDelete = func(*Object) { events = append(events, "delete") }

	boomCollideNative4E9770(source, target, nil, deps)
	if want := []string{"parent", "damage", "scorch", "map-damage", "map-push", "audio", "delete"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %q, want %q", events, want)
	}
}

func TestBoomCollide4E9770NativeWallReflectDirectionAndMapDamage(t *testing.T) {
	source := &Object{PosVec: types.Ptf(5, 6), VelVec: types.Ptf(6, -8)}
	normal := types.Ptf(1, -1)
	point := &ntype.Point32{X: 123, Y: -456}
	cache := &boomCollideBalance4E9770{Ready: 1, DirectDamage: 99}
	deps := defaultBoomCollideNativeDeps4E9770(cache)
	deps.traceHitPoint = func() *ntype.Point32 { return point }
	var calls int
	deps.damageMap = func(x, y, damage int32, damageType object.DamageType, gotSource *Object) {
		calls++
		if x != point.X || y != point.Y || damage != 99 || damageType != object.DamageExplosion || gotSource != source {
			t.Fatalf("map damage args = (%d,%d,%d,%v,%p)", x, y, damage, damageType, gotSource)
		}
	}

	boomCollideNative4E9770(source, nil, &normal, deps)
	if calls != 1 {
		t.Fatalf("map damage calls = %d, want 1", calls)
	}
	if source.VelVec != (types.Ptf(-4, 3)) {
		t.Fatalf("velocity = %v, want (-4,3)", source.VelVec)
	}
	if want := Dir16(directionFromVector509ED0(-8, 6)); source.Direction2 != want {
		t.Fatalf("Direction2 = %d, want %d", source.Direction2, want)
	}
}

func TestBoomCollide4E9770NativeLayoutAndConstants(t *testing.T) {
	if unsafe.Sizeof(boomCollideBalance4E9770{}) != 24 {
		t.Fatalf("balance cache size = %d, want 24", unsafe.Sizeof(boomCollideBalance4E9770{}))
	}
	fields := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"ready", unsafe.Offsetof(boomCollideBalance4E9770{}.Ready), 0},
		{"direct", unsafe.Offsetof(boomCollideBalance4E9770{}.DirectDamage), 4},
		{"splash", unsafe.Offsetof(boomCollideBalance4E9770{}.SplashDamage), 8},
		{"range", unsafe.Offsetof(boomCollideBalance4E9770{}.Range), 12},
		{"push range", unsafe.Offsetof(boomCollideBalance4E9770{}.PushRange), 16},
		{"force", unsafe.Offsetof(boomCollideBalance4E9770{}.Force), 20},
	}
	for _, field := range fields {
		if field.got != field.want {
			t.Errorf("%s offset = %d, want %d", field.name, field.got, field.want)
		}
	}
	if uint32(netmsg.MSG_FX_LESSER_EXPLOSION) != boomCollidePointFX4E9770 {
		t.Fatalf("lesser-explosion opcode = %d, want %d", netmsg.MSG_FX_LESSER_EXPLOSION, boomCollidePointFX4E9770)
	}
	if uint32(sound.SoundMagicMissileDetonate) != boomCollideDetonateAudio4E9770 {
		t.Fatalf("MagicMissileDetonate sound = %d, want %d", sound.SoundMagicMissileDetonate, boomCollideDetonateAudio4E9770)
	}
	if uint32(object.DamageExplosion) != boomCollideExplosionDamageType4E9770 {
		t.Fatalf("explosion damage type = %d, want %d", object.DamageExplosion, boomCollideExplosionDamageType4E9770)
	}
	if math.Float32bits(math.Float32frombits(boomCollideVelocityScaleBits4E9770)) != 0x3f000000 {
		t.Fatal("velocity scale bits changed")
	}
}
