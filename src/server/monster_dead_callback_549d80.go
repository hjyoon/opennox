package server

import (
	"math"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

// MonsterDeadCallbackKind549D80 identifies the ten monster.bin DEAD_FUNCTION
// callbacks implemented at GAME.EXE 00549D80-0054A750.
type MonsterDeadCallbackKind549D80 uint8

const (
	MonsterDeadCallbackEmberDemon549D80 MonsterDeadCallbackKind549D80 = iota + 1
	MonsterDeadCallbackDemon549E00
	MonsterDeadCallbackImp549E70
	MonsterDeadCallbackMechGolem549E90
	MonsterDeadCallbackGolem549FA0
	MonsterDeadCallbackBomber54A150
	MonsterDeadCallbackSpider54A250
	MonsterDeadCallbackTroll54A270
	MonsterDeadCallbackSkeleton54A310
	MonsterDeadCallbackSkeletonLord54A750
)

var monsterDeadMechParts549E90 = [...]string{
	"MechGear", "MechHead", "MechArm", "MechGear",
	"MechLeg", "MechLeg", "MechGear", "MechGear",
}

var monsterDeadGolemParts549FA0 = [...]struct {
	typeID string
	height float32
}{
	{"BigRock", 1},
	{"MediumRock", 3},
	{"SmallRock", 5},
	{"MediumRock", 3},
	{"SmallRock", 5},
	{"SmallRock", 5},
	{"MediumRock", 3},
	{"SmallRock", 5},
	{"SmallRock", 5},
	{"SmallRock", 5},
	{"SmallRock", 5},
	{"SmallRock", 5},
}

var monsterDeadSkeletonParts54A4C0 = [...]string{"ArmBone", "LegBone"}

// MonsterDeadCallbackRuntime549D80 supplies world effects for the native
// DEAD_FUNCTION implementations. Stateful debris indices remain backed by
// their original C globals so saves and repeated deaths keep the same cycle.
type MonsterDeadCallbackRuntime549D80 struct {
	CoopMode             func() bool
	RandomInt            func(minimum, maximum int) int
	RandomFloat          func(minimum, maximum float32) float32
	PushUnits            func(types.Pointf, float32, float32, float32, *Object)
	DamageUnits          func(types.Pointf, float32, float32, int, object.DamageType, *Object)
	SparkExplosion       func(types.Pointf, byte)
	PointFX              func(netmsg.Op, types.Pointf)
	Audio                func(sound.ID, *Object)
	DelayedDelete        func(*Object)
	BomberDead           func(*Object)
	NewObjectByTypeID    func(string) *Object
	RandomReachablePoint func(float32, types.Pointf) types.Pointf
	CreateObjectAt       func(*Object, *Object, types.Pointf)
	Raise                func(*Object, float32)
	ApplyForce           func(*Object, types.Pointf, float64)
	DecaySetTime         func(*Object, uint32)
	TickRate             func() uint32
	BalanceFloat         func(string) float64
	DropItem             func(*Object, MonsterDieDrop54A390)
	GolemPartIndex       func() uint32
	SetGolemPartIndex    func(uint32)
	SkeletonPartIndex    func() uint32
	SetSkeletonPartIndex func(uint32)
}

func validMonsterDeadCallbackKind549D80(kind MonsterDeadCallbackKind549D80) bool {
	return kind >= MonsterDeadCallbackEmberDemon549D80 && kind <= MonsterDeadCallbackSkeletonLord54A750
}

func monsterDeadDebrisDeps549D80(runtime MonsterDeadCallbackRuntime549D80) bool {
	return runtime.RandomInt != nil && runtime.RandomFloat != nil && runtime.PointFX != nil &&
		runtime.Audio != nil && runtime.NewObjectByTypeID != nil && runtime.RandomReachablePoint != nil &&
		runtime.CreateObjectAt != nil && runtime.Raise != nil && runtime.DecaySetTime != nil &&
		runtime.TickRate != nil
}

// MonsterDeadCallbackNative549D80 restores all registered DEAD_FUNCTION
// callbacks without passing Object pointers through the original PE32 integer
// arguments. A false result means required runtime wiring was unavailable;
// ordinary allocation failure is an original, successfully handled outcome.
func MonsterDeadCallbackNative549D80(
	unit *Object,
	kind MonsterDeadCallbackKind549D80,
	runtime MonsterDeadCallbackRuntime549D80,
) bool {
	if unit == nil || !validMonsterDeadCallbackKind549D80(kind) {
		return false
	}
	switch kind {
	case MonsterDeadCallbackEmberDemon549D80:
		if runtime.CoopMode == nil || runtime.PushUnits == nil || runtime.DamageUnits == nil ||
			runtime.SparkExplosion == nil || runtime.Audio == nil || runtime.DelayedDelete == nil {
			return false
		}
		damage := 96
		if runtime.CoopMode() {
			damage = 30
		}
		position := unit.PosVec
		runtime.PushUnits(position, 96, 10, 100, unit)
		runtime.DamageUnits(position, 96, 10, damage, object.DamageExplosion, unit)
		runtime.SparkExplosion(position, 128)
		runtime.Audio(sound.ID(42), unit)
		runtime.DelayedDelete(unit)
		return true
	case MonsterDeadCallbackDemon549E00:
		if runtime.PushUnits == nil || runtime.DamageUnits == nil || runtime.SparkExplosion == nil ||
			runtime.Audio == nil || runtime.DelayedDelete == nil {
			return false
		}
		position := unit.PosVec
		runtime.PushUnits(position, 150, 10, 150, unit)
		runtime.DamageUnits(position, 150, 10, 148, object.DamageExplosion, unit)
		runtime.SparkExplosion(position, 255)
		runtime.Audio(sound.ID(42), unit)
		runtime.DelayedDelete(unit)
		return true
	case MonsterDeadCallbackImp549E70, MonsterDeadCallbackSpider54A250:
		if runtime.PointFX == nil {
			return false
		}
		runtime.PointFX(netmsg.MSG_FX_BLUE_SPARKS, unit.PosVec)
		return true
	case MonsterDeadCallbackMechGolem549E90:
		return monsterDeadMechGolem549E90(unit, runtime)
	case MonsterDeadCallbackGolem549FA0:
		return monsterDeadGolem549FA0(unit, runtime)
	case MonsterDeadCallbackBomber54A150:
		if runtime.BomberDead == nil {
			return false
		}
		runtime.BomberDead(unit)
		return true
	case MonsterDeadCallbackTroll54A270:
		return monsterDeadTroll54A270(unit, runtime)
	case MonsterDeadCallbackSkeleton54A310, MonsterDeadCallbackSkeletonLord54A750:
		return monsterDeadSkeleton54A310(unit, kind, runtime)
	default:
		return false
	}
}

func monsterDeadMechGolem549E90(unit *Object, runtime MonsterDeadCallbackRuntime549D80) bool {
	if !monsterDeadDebrisDeps549D80(runtime) {
		return false
	}
	runtime.Audio(sound.SoundMechGolemDie, unit)
	runtime.PointFX(netmsg.MSG_FX_SMOKE_BLAST, unit.PosVec)
	for _, typeID := range monsterDeadMechParts549E90 {
		part := runtime.NewObjectByTypeID(typeID)
		if part == nil {
			break
		}
		position := runtime.RandomReachablePoint(30, unit.PosVec)
		runtime.CreateObjectAt(part, nil, position)
		runtime.Raise(part, runtime.RandomFloat(10, 70))
		part.Field27 = runtime.RandomFloat(-2, 0)
		part.Field29 = math.Float32bits(2)
		part.ObjFlags |= object.FlagBouncy
		seconds := runtime.RandomInt(10, 20)
		runtime.DecaySetTime(part, runtime.TickRate()*uint32(seconds))
	}
	return true
}

func monsterDeadGolem549FA0(unit *Object, runtime MonsterDeadCallbackRuntime549D80) bool {
	if !monsterDeadDebrisDeps549D80(runtime) || runtime.ApplyForce == nil || runtime.CoopMode == nil ||
		runtime.GolemPartIndex == nil || runtime.SetGolemPartIndex == nil {
		return false
	}
	runtime.Audio(sound.SoundGolemDie, unit)
	runtime.PointFX(netmsg.MSG_FX_SMOKE_BLAST, unit.PosVec)
	count := 6
	if runtime.CoopMode() {
		count = runtime.RandomInt(20, 30)
	}
	index := runtime.GolemPartIndex() % uint32(len(monsterDeadGolemParts549FA0))
	for range count {
		spec := monsterDeadGolemParts549FA0[index]
		part := runtime.NewObjectByTypeID(spec.typeID)
		if part == nil {
			break
		}
		position := runtime.RandomReachablePoint(30, unit.PosVec)
		runtime.CreateObjectAt(part, nil, position)
		runtime.Raise(part, runtime.RandomFloat(10, 70))
		part.Field27 = runtime.RandomFloat(-2, 0)
		part.ObjFlags |= object.FlagBouncy
		part.Field29 = math.Float32bits(spec.height)
		runtime.ApplyForce(part, unit.PosVec, float64(runtime.RandomFloat(5, 20)))
		var seconds int
		if runtime.CoopMode() {
			seconds = runtime.RandomInt(10, 20)
		} else {
			seconds = runtime.RandomInt(5, 10)
		}
		runtime.DecaySetTime(part, runtime.TickRate()*uint32(seconds))
		index = (index + 1) % uint32(len(monsterDeadGolemParts549FA0))
		runtime.SetGolemPartIndex(index)
	}
	return true
}

func monsterDeadTroll54A270(unit *Object, runtime MonsterDeadCallbackRuntime549D80) bool {
	if runtime.NewObjectByTypeID == nil || runtime.CreateObjectAt == nil || runtime.Audio == nil ||
		runtime.BalanceFloat == nil || runtime.TickRate == nil {
		return false
	}
	cloud := runtime.NewObjectByTypeID("SmallToxicCloud")
	if cloud == nil {
		return true
	}
	runtime.CreateObjectAt(cloud, unit, unit.PosVec)
	runtime.Audio(sound.SoundTrollFlatus, unit)
	lifetime := float32(runtime.BalanceFloat("SmallToxicCloudLifetime"))
	duration := poisonGasTrapRound4EB910(poisonGasTrapMultiply4EB910(lifetime, runtime.TickRate()))
	if cloud.UpdateData != nil {
		(*ToxicCloudUpdateData)(cloud.UpdateData).Duration = duration
	}
	return true
}

func monsterDeadSkeleton54A310(
	unit *Object,
	kind MonsterDeadCallbackKind549D80,
	runtime MonsterDeadCallbackRuntime549D80,
) bool {
	if !monsterDeadDebrisDeps549D80(runtime) || runtime.ApplyForce == nil || runtime.CoopMode == nil || runtime.DropItem == nil ||
		runtime.SkeletonPartIndex == nil || runtime.SetSkeletonPartIndex == nil {
		return false
	}
	roll := runtime.RandomInt(0, 100)
	monsterDeadSkeletonDebris54A4C0(unit, runtime)
	if roll <= 20 {
		return true
	}
	drop := MonsterDieDrop54A390{
		TypeID:    "Sword",
		Modifiers: [4]string{"WeaponPower1", "Material2"},
	}
	if roll > 50 {
		drop.TypeID = "WoodenShield"
		if kind == MonsterDeadCallbackSkeletonLord54A750 {
			drop.TypeID = "SteelShield"
		}
		drop.Modifiers = [4]string{"", "Material2"}
	}
	if runtime.CoopMode() {
		runtime.DropItem(unit, drop)
	}
	return true
}

func monsterDeadSkeletonDebris54A4C0(unit *Object, runtime MonsterDeadCallbackRuntime549D80) {
	runtime.Audio(sound.SoundSkeletonDie, unit)
	runtime.PointFX(netmsg.MSG_FX_SMOKE_BLAST, unit.PosVec)
	skull := runtime.NewObjectByTypeID("Skull")
	if skull == nil {
		return
	}
	position := runtime.RandomReachablePoint(20, unit.PosVec)
	runtime.CreateObjectAt(skull, nil, position)
	runtime.Raise(skull, 40)
	skull.Field27 = runtime.RandomFloat(-2, 0)
	skull.Field29 = math.Float32bits(4)
	skull.ObjFlags |= object.FlagBouncy
	runtime.ApplyForce(skull, unit.PosVec, float64(runtime.RandomFloat(5, 25)))

	var skullSeconds, partCount int
	if runtime.CoopMode() {
		skullSeconds = runtime.RandomInt(10, 20)
		partCount = runtime.RandomInt(10, 20)
	} else {
		skullSeconds = runtime.RandomInt(2, 5)
		partCount = runtime.RandomInt(5, 10)
	}
	runtime.DecaySetTime(skull, runtime.TickRate()*uint32(skullSeconds))

	index := runtime.SkeletonPartIndex() % uint32(len(monsterDeadSkeletonParts54A4C0))
	for range partCount {
		part := runtime.NewObjectByTypeID(monsterDeadSkeletonParts54A4C0[index])
		if part == nil {
			break
		}
		position = runtime.RandomReachablePoint(20, unit.PosVec)
		runtime.CreateObjectAt(part, nil, position)
		runtime.Raise(part, runtime.RandomFloat(10, 35))
		part.Field27 = runtime.RandomFloat(-2, 0)
		part.Field29 = math.Float32bits(4)
		part.ObjFlags |= object.FlagBouncy
		runtime.ApplyForce(part, unit.PosVec, float64(runtime.RandomFloat(5, 25)))
		var seconds int
		if runtime.CoopMode() {
			seconds = runtime.RandomInt(10, 20)
		} else {
			seconds = runtime.RandomInt(2, 5)
		}
		runtime.DecaySetTime(part, runtime.TickRate()*uint32(seconds))
		index = (index + 1) % uint32(len(monsterDeadSkeletonParts54A4C0))
		runtime.SetSkeletonPartIndex(index)
	}
}
