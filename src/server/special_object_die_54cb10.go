package server

import (
	"encoding/binary"
	"math"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

// PolypDeathRuntime54CB10 supplies the object lifecycle effects used by
// PolypDie. The ToxicCloud update record itself remains owned by Object.
type PolypDeathRuntime54CB10 struct {
	NewObjectByTypeID func(string) *Object
	CreateObjectAt    func(*Object, *Object, types.Pointf)
	BalanceFloat      func(string) float64
	TickRate          func() uint32
	Audio             func(sound.ID, *Object)
	DelayedDelete     func(*Object)
}

// PolypDieNative54CB10 restores GAME.EXE 0054CB10 without narrowing the
// source pointer to the original PE32 int parameter. GAME.EXE caches the
// cloud's update-data pointer before CreateAt, so this implementation does too.
func PolypDieNative54CB10(source *Object, runtime PolypDeathRuntime54CB10) {
	cloud := runtime.NewObjectByTypeID(poisonGasTrapCloudType4EB910)
	if cloud != nil {
		update := (*ToxicCloudUpdateData)(cloud.UpdateData)
		runtime.CreateObjectAt(cloud, nil, source.PosVec)
		lifetime := float32(runtime.BalanceFloat(poisonGasTrapLifetime4EB910))
		duration := poisonGasTrapRound4EB910(poisonGasTrapMultiply4EB910(lifetime, runtime.TickRate()))
		update.Duration = duration
	}
	runtime.Audio(sound.SoundPolypExplode, source)
	runtime.DelayedDelete(source)
}

// MarkerDeathRuntime54E460 supplies the owner lookup and world effects used
// by MarkerDie.
type MarkerDeathRuntime54E460 struct {
	FindOwnerChainPlayer func(*Object) *Object
	PointFX              func(netmsg.Op, types.Pointf)
	DelayedDelete        func(*Object)
}

// MarkerDieNative54E460 restores GAME.EXE 0054E460. Only the first matching
// player marker slot is cleared before the smoke effect and deferred delete.
func MarkerDieNative54E460(source *Object, runtime MarkerDeathRuntime54E460) {
	player := runtime.FindOwnerChainPlayer(source)
	if player != nil {
		update := (*PlayerUpdateData)(player.UpdateData)
		if update != nil {
			for index, marker := range update.Field29 {
				if marker == source {
					update.Field29[index] = nil
					break
				}
			}
		}
	}
	runtime.PointFX(netmsg.MSG_FX_SMOKE_BLAST, source.PosVec)
	runtime.DelayedDelete(source)
}

// BoulderDeathRuntime54E4B0 supplies the random source, debris factory, and
// world effects used by BoulderDie.
type BoulderDeathRuntime54E4B0 struct {
	RandomInt            func(int, int) int
	RandomFloat          func(float32, float32) float32
	Audio                func(sound.ID, *Object)
	PointFX              func(netmsg.Op, types.Pointf)
	NewObjectByTypeID    func(string) *Object
	RandomReachablePoint func(float32, types.Pointf) types.Pointf
	CreateObjectAt       func(*Object, *Object, types.Pointf)
	Raise                func(*Object, float32)
	ApplyForce           func(*Object, types.Pointf, float64)
	DecaySetTime         func(*Object, uint32)
	TickRate             func() uint32
	PartIndex            func() uint32
	SetPartIndex         func(uint32)
	DelayedDelete        func(*Object)
}

// BoulderDieNative54E4B0 restores GAME.EXE 0054E4B0. It returns false when a
// debris allocation fails; in that case the original callback exits without
// scheduling deletion of the boulder. A true result means the full loop (or a
// non-positive count) completed and the source was scheduled for deletion.
func BoulderDieNative54E4B0(source *Object, runtime BoulderDeathRuntime54E4B0) bool {
	runtime.Audio(sound.SoundWallDestroyedStone, source)
	runtime.PointFX(netmsg.MSG_FX_SMOKE_BLAST, source.PosVec)
	count := runtime.RandomInt(20, 30)
	if count <= 0 {
		runtime.DelayedDelete(source)
		return true
	}

	index := runtime.PartIndex() % uint32(len(monsterDeadGolemParts549FA0))
	for range count {
		spec := monsterDeadGolemParts549FA0[index]
		part := runtime.NewObjectByTypeID(spec.typeID)
		if part == nil {
			return false
		}
		position := runtime.RandomReachablePoint(30, source.PosVec)
		runtime.CreateObjectAt(part, nil, position)
		runtime.Raise(part, runtime.RandomFloat(10, 70))
		part.Field27 = runtime.RandomFloat(-2, 0)
		part.ObjFlags |= object.FlagBouncy
		part.Field29 = math.Float32bits(spec.height)
		runtime.ApplyForce(part, source.PosVec, float64(runtime.RandomFloat(5, 20)))
		seconds := runtime.RandomInt(45, 75)
		runtime.DecaySetTime(part, runtime.TickRate()*uint32(seconds))
		index = (index + 1) % uint32(len(monsterDeadGolemParts549FA0))
		runtime.SetPartIndex(index)
	}
	runtime.DelayedDelete(source)
	return true
}

// MonsterGeneratorDeathRuntime54E630 supplies the script, Quest, network, and
// object lifecycle services reached by MonsterGeneratorDie.
type MonsterGeneratorDeathRuntime54E630 struct {
	Frame                func() uint32
	SetQuestTimer        func(uint32)
	SetQuestMode         func(int32)
	ScriptCallback       func(*ScriptCallback, *Object, *Object, ScriptEventType)
	Audio                func(sound.ID, *Object)
	SendFX               func(types.Pointf, []byte)
	QuestMode            func() bool
	FindOwnerChainPlayer func(*Object) *Object
	NewObjectByTypeID    func(string) *Object
	CreateObjectAt       func(*Object, *Object, types.Pointf)
	DelayedDelete        func(*Object)
}

func monsterGeneratorBreakFXPacket523200(position types.Pointf, intensity byte) [7]byte {
	var packet [7]byte
	packet[0] = 0xf0
	packet[1] = 25
	binary.LittleEndian.PutUint16(packet[2:4], uint16(poisonGasTrapRound4EB910(position.X)))
	binary.LittleEndian.PutUint16(packet[4:6], uint16(poisonGasTrapRound4EB910(position.Y)))
	packet[6] = intensity
	return packet
}

// RecordMonsterGeneratorDestroyed applies GAME.EXE sub_4D61B0's player
// bookkeeping after the caller has resolved a live player unit.
func (p *Player) RecordMonsterGeneratorDestroyed() {
	if p == nil {
		return
	}
	p.field4668++
	p.field4692 |= 8
}

func recordMonsterGeneratorDestroyed4D61B0(playerUnit *Object) {
	if playerUnit == nil || uint8(playerUnit.ObjClass)&uint8(object.ClassPlayer) == 0 ||
		playerUnit.ObjFlags.Has(object.FlagDestroyed) {
		return
	}
	update := (*PlayerUpdateData)(playerUnit.UpdateData)
	if update == nil || update.Player == nil {
		return
	}
	update.Player.RecordMonsterGeneratorDestroyed()
}

// MonsterGeneratorDieNative54E630 restores GAME.EXE 0054E630 with
// native-width Object references. The generator callback record is cached
// before any Quest calls, matching the original update-data load order.
func MonsterGeneratorDieNative54E630(source *Object, runtime MonsterGeneratorDeathRuntime54E630) {
	update := (*MonsterGenUpdateData)(source.UpdateData)
	runtime.SetQuestTimer(runtime.Frame())
	runtime.SetQuestMode(0)
	runtime.ScriptCallback(
		(*ScriptCallback)(unsafe.Pointer(&update.Field56)),
		source.Obj130,
		source,
		NoxEventGeneratorDead,
	)
	runtime.Audio(sound.SoundMonsterGeneratorDie, source)
	packet := monsterGeneratorBreakFXPacket523200(source.PosVec, 200)
	runtime.SendFX(source.PosVec, packet[:])
	if runtime.QuestMode() && source.Obj130 != nil {
		playerUnit := runtime.FindOwnerChainPlayer(source.Obj130)
		recordMonsterGeneratorDestroyed4D61B0(playerUnit)
	}
	destroyed := runtime.NewObjectByTypeID("DestroyedGenerator")
	if destroyed != nil {
		runtime.CreateObjectAt(destroyed, nil, source.PosVec)
	}
	runtime.DelayedDelete(source)
}
