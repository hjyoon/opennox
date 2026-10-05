package opennox

import (
	"fmt"
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

// These names and bytes describe only the stock test fixtures. Gameplay
// admits DANGEROUS/non-unit IMPALE hazards without a type-name special case.
func e2ePlayerSpikeKind(typeID string) (damage uint8, switched, ok bool) {
	switch typeID {
	case "Spike", "PeriodicSpike":
		return 2, true, true
	case "SpikeBlock", "SpikeBlockImmobile":
		return 3, false, true
	case "RotatingSpikes", "RotatingSpikesImmobile":
		return 8, false, true
	default:
		return 0, false, false
	}
}

type e2ePlayerSpikeFixture struct {
	typeID                      string
	damage                      uint8
	switched                    bool
	host, hazard                *server.Object
	original, origin, direction types.Pointf
	data                        server.DamageCollideData
	health, stoppedHP           uint16
	baselineFrame, frame        uint32
}

func (f *e2ePlayerSpikeFixture) prepare() {
	f.host = noxServer.Players.HostUnit()
	if f.host == nil || f.host.HealthData == nil || f.host.UpdateData == nil || f.host.ControllingPlayer() == nil ||
		f.host.HealthData.Cur == 0 || f.host.Buffs != 0 || f.host.Flags().HasAny(object.FlagDead|object.FlagDestroyed|object.FlagNoUpdate) {
		e2eError(fmt.Errorf("spike fixture requires a live unenchanted host"))
		return
	}
	f.original = f.host.PosVec
	var err error
	f.origin, f.direction, err = e2eWarriorAbilityArena(f.original, f.host.Shape.Circle.R+65, func(from, to types.Pointf) bool {
		return e2eWarriorLaneClear(f.host, from, to)
	})
	if err != nil {
		e2eError(err)
		return
	}
	f.hazard = noxServer.NewObjectByTypeID(f.typeID)
	if f.hazard == nil || f.hazard.CollideData == nil {
		e2eError(fmt.Errorf("spike fixture has no stock %s collide data", f.typeID))
		return
	}
	if f.switched {
		asObjectS(f.hazard).Enable(false)
	}
	noxServer.CreateObjectAt(f.hazard, nil, f.origin)
	noxServer.ObjectsAddPending()
	f.data = *(*server.DamageCollideData)(f.hazard.CollideData)
	collide, size, registered := server.ObjectCollideHandler("DamageCollide")
	if !e2eObjectInWorld(f.hazard) || f.hazard.ObjOwner != nil || f.hazard.Class() != f.hazard.ObjectTypeC().Class() ||
		f.hazard.Shape != f.hazard.ObjectTypeC().Shape ||
		!f.hazard.Class().Has(object.ClassDangerous) || f.hazard.Class().HasAny(object.MaskUnits|object.ClassWeapon|object.ClassWand|object.ClassMissile) ||
		f.hazard.Collide == nil || f.hazard.Collide != f.hazard.ObjectTypeC().Collide ||
		!registered || size != 8 || f.hazard.Collide != collide ||
		f.data.Damage != f.damage || f.data.DamageType != int32(object.DamageImpale) {
		e2eError(fmt.Errorf("stock spike initialization mismatch: type=%s hazard=%p class=%x collide=%p data=%+v", f.typeID, f.hazard, f.hazard.Class(), f.hazard.Collide, f.data))
		return
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, p := range []unsafe.Pointer{unsafe.Pointer(f.host), f.host.UpdateData, unsafe.Pointer(f.host.HealthData), unsafe.Pointer(f.host.ControllingPlayer()), unsafe.Pointer(f.hazard), f.hazard.CollideData} {
			if uintptr(p) <= math.MaxUint32 {
				e2eError(fmt.Errorf("spike native stock pointer below 4 GiB: %p", p))
				return
			}
		}
	}
	// Place a stationary contact fixture. Switch hazards first overlap while
	// disabled; other shapes start outside their collision bounds. SetPos and
	// the ordinary collision queue do not call DamageCollide or supply HP loss.
	offset := float32(80)
	if f.switched {
		offset = 6
	}
	asObjectS(f.host).SetPos(f.origin.Add(f.direction.Mul(offset)))
	legacy.Nox_xxx_unitHasCollideOrUpdateFn_537610(f.host)
	f.health, f.baselineFrame = f.host.HealthData.Cur, f.host.Frame134
	e2eLog.Printf("SPIKE PREPARED: type=%s hazard=%p collide=%p player=%p HP=%d data=%+v switched=%t", f.typeID, f.hazard, f.hazard.Collide, f.host, f.health, f.data, f.switched)
}

func (f *e2ePlayerSpikeFixture) start() {
	if f.host.HealthData.Cur < f.health || f.host.Frame134 != f.baselineFrame ||
		f.switched && (f.hazard.IsEnabled() || !f.hazard.Flags().Has(object.FlagNoCollide)) {
		e2eError(fmt.Errorf("spike safe baseline mismatch: type=%s HP=%d/%d hit-frame=%d/%d enabled=%t flags=%x", f.typeID, f.host.HealthData.Cur, f.health, f.host.Frame134, f.baselineFrame, f.hazard.IsEnabled(), f.hazard.Flags()))
		return
	}
	if f.switched {
		asObjectS(f.hazard).Enable(true)
	} else {
		asObjectS(f.host).SetPos(f.origin.Add(f.direction.Mul(6)))
	}
	f.health, f.frame = f.host.HealthData.Cur, noxServer.Frame()
	legacy.Nox_xxx_unitHasCollideOrUpdateFn_537610(f.host)
	e2eLog.Printf("SPIKE CONTACT: type=%s HP=%d frame=%d safe-baseline=passed enabled=%t", f.typeID, f.health, f.frame, f.hazard.IsEnabled())
}

func (f *e2ePlayerSpikeFixture) damaged() bool {
	if f.host.HealthData.Cur >= f.health {
		return false
	}
	update := f.host.UpdateDataPlayer()
	if f.host.HealthData.Cur == 0 || f.host.Flags().HasAny(object.FlagDead|object.FlagDestroyed) ||
		f.host.Obj130 != f.hazard || f.host.Field131 != uint32(object.DamageImpale) || f.host.Frame134 < f.frame ||
		update.Field76 != 2 || update.Field75 != uint32(object.DamageImpale) ||
		*(*server.DamageCollideData)(f.hazard.CollideData) != f.data || f.hazard.Flags().Has(object.FlagNoCollide) {
		e2eError(fmt.Errorf("spike actual damage mismatch: type=%s HP=%d/%d source=%p/%p type=%d marker=%d/%d frame=%d/%d flags=%x", f.typeID, f.host.HealthData.Cur, f.health, f.host.Obj130, f.hazard, f.host.Field131, update.Field76, update.Field75, f.host.Frame134, f.frame, f.hazard.Flags()))
		return true
	}
	f.stoppedHP = f.host.HealthData.Cur
	e2eLog.Printf("SPIKE ACTUAL DAMAGE: type=%s HP=%d->%d loss=%d frame=%d source=%p type=3 marker=2/3 stock-data=unchanged", f.typeID, f.health, f.stoppedHP, f.health-f.stoppedHP, f.host.Frame134, f.hazard)
	// Only after real spatial collision and damage have been observed, remove
	// this fixture through the ordinary deletion queue. Never replace callbacks.
	noxServer.DelayedDelete(f.hazard)
	return true
}

func (f *e2ePlayerSpikeFixture) clientResult() bool {
	drawable := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.host)))
	if drawable == nil {
		return false
	}
	delta, ok := legacy.HealthChangeForDrawable(drawable.NetCode32)
	if !ok || delta >= 0 {
		return false
	}
	// A fresh display-list baseline was required before this contact. Do not
	// inject a packet or damage-number record; inspect the normal consumer.
	if f.stoppedHP >= f.health || f.host.HealthData.Cur == 0 {
		e2eError(fmt.Errorf("spike client damage has no live server hit: type=%s delta=%d HP=%d/%d", f.typeID, delta, f.stoppedHP, f.health))
		return true
	}
	path, err := e2eWriteMagicFrame("", noxClient.r.CopyPixBuffer())
	if err != nil {
		e2eError(err)
		return true
	}
	e2eLog.Printf("SPIKE CLIENT DAMAGE: type=%s delta=%d actual-HP=%d drawable=%p", f.typeID, delta, f.stoppedHP, drawable)
	e2eLog.Printf("SPIKE LIVE FRAME: type=%s path=%s", f.typeID, path)
	return true
}

// CheckPlayerSpikeCollision observes stock circle/box and active/inactive
// collision gates, normal player damage, networking and live rendering. Map
// placement is a fixture, not a claim about every campaign spawn/input path.
func (sc *e2eScenario) CheckPlayerSpikeCollision(typeID, name string) {
	damage, switched, ok := e2ePlayerSpikeKind(typeID)
	if !ok {
		e2eError(fmt.Errorf("invalid player spike fixture %q", typeID))
		return
	}
	f := &e2ePlayerSpikeFixture{typeID: typeID, damage: damage, switched: switched}
	sc.addWhen(0, name+" prepare stock hazard", 1200, func() bool {
		unit := noxServer.Players.HostUnit()
		if !nox_client_isConnected() || unit == nil || unit.Buffs != 0 || noxClient.ClientPlayerUnit() == nil {
			return false
		}
		_, stale := legacy.HealthChangeForDrawable(uint32(noxServer.GetUnitNetCode(unit)))
		return !stale
	}, f.prepare)
	sc.Wait(6, name+" publish safe baseline")
	sc.add(0, name+" ordinary contact", f.start)
	sc.addWhen(1, name+" actual collision damage", 120, f.damaged, func() {})
	sc.addWhen(1, name+" client damage", 120, f.clientResult, func() {})
	sc.add(0, name+" restore fixture position", func() {
		asObjectS(f.host).SetPos(f.original)
	})
	sc.Wait(3, name+" fixture cleanup")
}
