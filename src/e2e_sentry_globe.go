package opennox

import (
	"fmt"
	"image"
	"math"
	"unsafe"

	"github.com/opennox/libs/noximage"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func e2eSentryMode(mode string) (owner, target string, ok bool) {
	switch mode {
	case "world-to-player":
		return "world", "player", true
	case "world-to-npc":
		return "world", "NPC", true
	case "world-to-monster":
		return "world", "Troll", true
	case "player-to-npc":
		return "player", "NPC", true
	case "npc-to-player":
		return "NPC", "player", true
	case "npc-to-monster":
		return "NPC", "Troll", true
	default:
		return "", "", false
	}
}

// Count the stock solid upper beam, away from both source and target sprites.
// The ray endpoints come from the live server ray; the native viewport and
// framebuffer are read, never redrawn, masked or replaced by this observer.
func e2eSentryBeamPixels(pix *noximage.Image16, from, to image.Point, color uint16) (int, bool) {
	if pix == nil || color == 0 {
		return 0, false
	}
	dx, dy := float64(to.X-from.X), float64(to.Y-from.Y)
	distance := math.Hypot(dx, dy)
	if distance < 80 {
		return 0, false
	}
	seen := make(map[image.Point]bool)
	count := 0
	for step := 32; step <= 72; step++ {
		p := image.Pt(from.X+int(math.Round(dx*float64(step)/distance)), from.Y+int(math.Round(dy*float64(step)/distance))-22)
		for y := -1; y <= 1; y++ {
			for x := -1; x <= 1; x++ {
				point := p.Add(image.Pt(x, y))
				if !point.In(pix.Rect) {
					return 0, false
				}
				if !seen[point] && pix.Pix[pix.PixOffset(point.X, point.Y)] == color {
					count++
				}
				seen[point] = true
			}
		}
	}
	return count, true
}

type e2eSentryFixture struct {
	mode, ownerKind, targetKind        string
	host, target, owner, npc, globe    *server.Object
	original, direction                types.Pointf
	hostHP, hostMax, health, stoppedHP uint16
	angle                              float32
	active, stopped, clientHit         bool
	hits, baselinePixels, beamPixels   int
	color                              uint16
	carry, armor                       uint32
	frame, hitFrame                    uint32
}

func (f *e2eSentryFixture) prepare() {
	f.host = noxServer.Players.HostUnit()
	if f.host == nil || f.host.HealthData == nil || f.host.UpdateData == nil || f.host.ControllingPlayer() == nil ||
		f.host.HealthData.Cur == 0 || f.host.Buffs != 0 || f.host.Flags().HasAny(object.FlagDead|object.FlagDestroyed|object.FlagNoUpdate) {
		e2eError(fmt.Errorf("Sentry fixture requires a live unenchanted host"))
		return
	}
	f.original, f.hostHP, f.hostMax = f.host.PosVec, f.host.HealthData.Cur, f.host.HealthData.Max
	origin, direction, err := e2eWarriorAbilityArena(f.original, f.host.Shape.Circle.R+16, func(from, to types.Pointf) bool {
		return e2eWarriorLaneClear(f.host, from, to)
	})
	if err != nil {
		e2eError(err)
		return
	}
	// A slight diagonal preserves a nonempty packet visibility rectangle even
	// in a horizontal/vertical arena. The target stays inside its circle radius.
	f.angle = float32(math.Atan2(float64(direction.Y), float64(direction.X)) + 0.025)
	f.direction = types.Ptf(float32(math.Cos(float64(f.angle))), float32(math.Sin(float64(f.angle))))
	side := types.Ptf(-direction.Y, direction.X)
	asObjectS(f.host).SetPos(origin.Add(side.Mul(48)))
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	f.target = f.host
	if f.targetKind != "player" {
		f.target = noxServer.NewObjectByTypeID(f.targetKind)
		if f.target == nil {
			e2eError(fmt.Errorf("Sentry fixture has no stock %s", f.targetKind))
			return
		}
		var master *server.Object
		if f.ownerKind == "NPC" {
			// Unowned world monsters are allies of unowned NPCs. A normal
			// player-owned summon supplies the hostile owner relationship.
			master = f.host
		}
		noxServer.CreateObjectAt(f.target, master, origin.Add(direction.Mul(112)))
	} else {
		asObjectS(f.host).SetPos(origin.Add(direction.Mul(112)))
	}
	if f.ownerKind == "player" {
		f.owner = f.host
	} else if f.ownerKind == "NPC" {
		f.npc = noxServer.NewObjectByTypeID("NPC")
		if f.npc == nil {
			e2eError(fmt.Errorf("Sentry fixture has no stock NPC owner"))
			return
		}
		noxServer.CreateObjectAt(f.npc, nil, origin.Add(side.Mul(64)))
		f.owner = f.npc
	}
	noxServer.ObjectsAddPending()
	if f.owner != nil && !noxServer.S().IsEnemyTo(f.owner, f.target) {
		e2eError(fmt.Errorf("Sentry owned fixture requires a hostile target: mode=%s", f.mode))
		return
	}
	for _, unit := range []*server.Object{f.target, f.npc} {
		if unit == nil {
			continue
		}
		if !e2eObjectInWorld(unit) || unit.HealthData == nil || unit.UpdateData == nil || unit.Damage == nil || unit.Buffs != 0 {
			e2eError(fmt.Errorf("Sentry stock unit was not initialized: %p", unit))
			return
		}
		if unit == f.target && unit.Flags().HasAny(object.FlagShort|object.FlagBelow|object.FlagNoCollide) {
			e2eError(fmt.Errorf("Sentry fixture target is excluded by the stock regular-game ray gate: %s flags=%x", f.targetKind, unit.Flags()))
			return
		}
		if unit != f.host {
			unit.UpdateDataMonster().SetAggression(0)
			unit.ClearActionStack()
			unit.MonsterPushAction(ai.ACTION_WAIT, noxServer.Frame()+250000)
		}
	}
	// Durable HP, stock map placement, waiting AI and the map-style beam angle
	// are setup. No damage callback, HP loss, hit marker, ray endpoint, packet,
	// drawable, buff, player state or pixel result is supplied by this fixture.
	asObjectS(f.target).SetMaxHealth(20000)
	f.health = f.target.HealthData.Cur
	if f.target == f.host {
		ud := f.target.UpdateDataPlayer()
		f.carry, f.armor = ud.Field21, ud.Field57
	} else {
		ud := f.target.UpdateDataMonster()
		f.carry, f.armor = ud.Field1, ud.Field518
	}
	f.globe = noxServer.NewObjectByTypeID("SentryGlobe")
	if f.globe == nil || f.globe.UpdateData == nil {
		e2eError(fmt.Errorf("Sentry fixture has no stock SentryGlobe"))
		return
	}
	asObjectS(f.globe).Enable(false)
	data := f.globe.UpdateDataSentry()
	data.Field0, data.Field4, data.Field8 = math.Float32bits(f.angle), math.Float32bits(f.angle), 0
	noxServer.CreateObjectAt(f.globe, f.owner, origin)
	noxServer.ObjectsAddPending()
	if !e2eObjectInWorld(f.globe) || f.globe.Update == nil || f.globe.Update != f.globe.ObjectTypeC().Update ||
		f.globe.Class() != f.globe.ObjectTypeC().Class() || f.globe.Class()&(object.ClassSimple|object.ClassImmobile) != object.ClassSimple|object.ClassImmobile ||
		f.globe.ObjOwner != f.owner || f.globe.IsEnabled() || f.health != 20000 {
		e2eError(fmt.Errorf("Sentry stock globe lost initialization/owner/disabled gate: globe=%p owner=%p/%p enabled=%t health=%d", f.globe, f.globe.ObjOwner, f.owner, f.globe.IsEnabled(), f.health))
		return
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, p := range []unsafe.Pointer{unsafe.Pointer(f.host), f.host.UpdateData, unsafe.Pointer(f.target), f.target.UpdateData, unsafe.Pointer(f.globe), f.globe.UpdateData} {
			if uintptr(p) <= math.MaxUint32 {
				e2eError(fmt.Errorf("Sentry native stock pointer below 4 GiB: %p", p))
				return
			}
		}
	}
	noxServer.Audio.OnSound(func(id sound.ID, kind int, target *server.Object, _ types.Pointf) {
		if !f.active || id != sound.SoundSentryRayHit || target != f.target {
			return
		}
		f.hits++
		// Monster updates consume Injured in the same normal tick. Observe
		// the damage-side flag at the real hit audio, before that consumption.
		if f.target != f.host && !f.target.UpdateDataMonster().StatusFlags.Has(object.MonStatusInjured) {
			e2eError(fmt.Errorf("Sentry did not mark the actual target injured at hit"))
			return
		}
		wantHP := int(f.health) - 500*f.hits
		// Durable HP naturally regenerates between ticks. Keep actual damage
		// and its frame instead of disabling regeneration for the fixture.
		if kind != 0 || f.hits > 12 || f.hits == 1 && f.target.HealthData.Cur != f.health-500 || int(f.target.HealthData.Cur) < wantHP ||
			int(f.target.HealthData.Cur) > wantHP+10*int(noxServer.Frame()-f.frame) ||
			f.target.Obj130 != f.globe || f.target.Field131 != uint32(object.DamageZapRay) {
			e2eError(fmt.Errorf("Sentry real hit mismatch: mode=%s count=%d kind=%d HP=%d/%d source=%p/%p type=%d", f.mode, f.hits, kind, f.target.HealthData.Cur, f.health, f.target.Obj130, f.globe, f.target.Field131))
			return
		}
		e2eLog.Printf("SENTRY ACTUAL HIT: mode=%s count=%d frame=%d HP=%d->%d source=%p type=16", f.mode, f.hits, noxServer.Frame(), f.health, f.target.HealthData.Cur, f.globe)
	})
	e2eLog.Printf("SENTRY PREPARED: mode=%s globe=%p update=%p target=%p update=%p owner=%p HP=%d angle=%g", f.mode, f.globe, f.globe.UpdateData, f.target, f.target.UpdateData, f.owner, f.health, f.angle)
}

func (f *e2eSentryFixture) pixels(endpoint types.Pointf) (int, bool) {
	vp := noxClient.Viewport()
	from := image.Pt(int(math.RoundToEven(float64(f.globe.PosVec.X))), int(math.RoundToEven(float64(f.globe.PosVec.Y))))
	to := image.Pt(int(math.RoundToEven(float64(endpoint.X))), int(math.RoundToEven(float64(endpoint.Y))))
	return e2eSentryBeamPixels(noxClient.r.PixBuffer(), vp.ToScreenPos(from), vp.ToScreenPos(to), f.color)
}

func (f *e2eSentryFixture) start() {
	f.color = uint16(memmap.Uint32(0x5D4594, 1321536))
	count, ok := f.pixels(f.globe.PosVec.Add(f.direction.Mul(128)))
	if !ok || f.target.HealthData.Cur != f.health || f.hits != 0 || f.globe.IsEnabled() {
		e2eError(fmt.Errorf("Sentry disabled baseline invalid: mode=%s pixels=%d ok=%t color=%x HP=%d", f.mode, count, ok, f.color, f.target.HealthData.Cur))
		return
	}
	f.baselinePixels, f.frame, f.active = count, noxServer.Frame(), true
	asObjectS(f.globe).Enable(true)
	e2eLog.Printf("SENTRY ENABLED: mode=%s baseline-pixels=%d color=%x frame=%d", f.mode, count, f.color, f.frame)
}

func (f *e2eSentryFixture) renderedHit() bool {
	if f.hits == 0 {
		return false
	}
	count, ok := f.pixels(f.globe.Pos39)
	if !ok || count < f.baselinePixels+20 {
		return false
	}
	var marker, markerType, carry, armor uint32
	if f.target == f.host {
		ud := f.target.UpdateDataPlayer()
		marker, markerType, carry, armor = ud.Field76, ud.Field75, ud.Field21, ud.Field57
	} else {
		ud := f.target.UpdateDataMonster()
		marker, markerType, carry, armor = ud.Field547, ud.Field546, ud.Field1, ud.Field518
	}
	wantMarker, wantType := uint32(1), uint32(f.globe.TypeInd)
	if f.owner == nil {
		wantMarker, wantType = 2, 16
	}
	if marker != wantMarker || markerType != wantType || carry != f.carry || armor != f.armor ||
		f.globe.UpdateDataSentry().Field0 != math.Float32bits(f.angle) || !f.globe.Flags().Has(object.FlagMarked) {
		e2eError(fmt.Errorf("Sentry marker/armor/carry mismatch: mode=%s marker=%d/%d want=%d/%d carry=%x/%x armor=%x/%x", f.mode, marker, markerType, wantMarker, wantType, carry, f.carry, armor, f.armor))
		return true
	}
	f.beamPixels, f.stoppedHP, f.stopped = count, f.target.HealthData.Cur, true
	f.hitFrame = f.target.Frame134
	path, err := e2eWriteMagicFrame("", noxClient.r.CopyPixBuffer())
	if err != nil {
		e2eError(err)
		return true
	}
	e2eLog.Printf("SENTRY LIVE FRAME: mode=%s path=%s", f.mode, path)
	e2eLog.Printf("SENTRY RENDERED HIT: mode=%s hits=%d HP=%d beam-pixels=%d baseline=%d endpoint=%v marker=%d/%d armor/carry=unchanged", f.mode, f.hits, f.stoppedHP, count, f.baselinePixels, f.globe.Pos39, marker, markerType)
	// Stop through the ordinary object enable API only after both an actual
	// stock update hit and the normally networked/rendered beam were observed.
	asObjectS(f.globe).Enable(false)
	return true
}

func (f *e2eSentryFixture) clientResult() bool {
	dr := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target)))
	if dr == nil {
		return false
	}
	delta, ok := legacy.HealthChangeForDrawable(dr.NetCode32)
	if !ok || delta >= 0 {
		return false
	}
	if int32(delta)%500 != 0 || -int32(delta) > int32(500*f.hits) || f.target.HealthData.Cur < f.stoppedHP || f.target.Frame134 != f.hitFrame {
		e2eError(fmt.Errorf("Sentry client result mismatch: mode=%s delta=%d hits=%d HP=%d/%d", f.mode, delta, f.hits, f.target.HealthData.Cur, f.stoppedHP))
		return true
	}
	f.clientHit = true
	e2eLog.Printf("SENTRY CLIENT DAMAGE: mode=%s delta=%d actual-HP=%d drawable=%p", f.mode, delta, f.stoppedHP, dr)
	return true
}

func (f *e2eSentryFixture) ended() bool {
	if !f.stopped || !f.clientHit || f.globe.IsEnabled() || noxServer.Frame()-f.frame < 30 {
		return false
	}
	count, ok := f.pixels(f.globe.Pos39)
	if !ok || count > f.baselinePixels+4 {
		return false
	}
	if f.target.HealthData.Cur < f.stoppedHP || f.target.Frame134 != f.hitFrame || f.globe.UpdateDataSentry().Field0 != f.globe.UpdateDataSentry().Field4 || legacy.SentryRayCount4C5020() != 0 {
		e2eError(fmt.Errorf("Sentry disabled/consumed queue regression: mode=%s HP=%d/%d rays=%d", f.mode, f.target.HealthData.Cur, f.stoppedHP, legacy.SentryRayCount4C5020()))
		return true
	}
	f.active = false
	e2eLog.Printf("SENTRY STOPPED: mode=%s hits=%d HP=%d beam-pixels=%d baseline=%d queue=0 elapsed=%d", f.mode, f.hits, f.stoppedHP, count, f.baselinePixels, noxServer.Frame()-f.frame)
	noxServer.DelayedDelete(f.globe)
	if f.target != f.host {
		noxServer.DelayedDelete(f.target)
	}
	if f.npc != nil {
		noxServer.DelayedDelete(f.npc)
	}
	asObjectS(f.host).SetPos(f.original)
	if f.target == f.host {
		asObjectS(f.host).SetMaxHealth(int(f.hostMax))
		asObjectS(f.host).SetHealth(int(f.hostHP))
	}
	return true
}

// CheckSentryGlobe uses a real initialized stock map hazard and normal server
// ticks, damage, audio, network delivery and rendering. It does not prove a
// particular stock-map spawn or possession input; those remain separate work.
func (sc *e2eScenario) CheckSentryGlobe(mode, name string) {
	owner, target, ok := e2eSentryMode(mode)
	if !ok {
		e2eError(fmt.Errorf("invalid Sentry scenario %q", mode))
		return
	}
	f := &e2eSentryFixture{mode: mode, ownerKind: owner, targetKind: target}
	sc.addWhen(0, name+" prepare stock hazard", 1200, func() bool {
		unit := noxServer.Players.HostUnit()
		return nox_client_isConnected() && unit != nil && noxClient.ClientPlayerUnit() != nil && unit.Buffs == 0
	}, f.prepare)
	sc.Wait(12, name+" publish disabled fixture")
	sc.add(0, name+" enable ordinary hazard", f.start)
	sc.addWhen(1, name+" real damage and rendered ray", 12, f.renderedHit, func() {})
	sc.addWhen(1, name+" client damage", 120, f.clientResult, func() {})
	sc.addWhen(1, name+" disabled ray cleanup", 120, f.ended, func() {})
	sc.Wait(3, name+" fixture cleanup")
}
