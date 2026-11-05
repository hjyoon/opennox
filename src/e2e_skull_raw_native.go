package opennox

import (
	"context"
	"fmt"
	"math"
	"unsafe"

	"github.com/opennox/libs/console"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

// Explicit raw-C diagnostic, enabled only during playback. The stock trap's
// ordinary update is held while the five public entries are exercised.
// Prepared records do not supply natural trap activation or campaign credit.
func init() {
	if e2ePlay == "" {
		return
	}
	var host, source, control *server.Object
	var hp uint16
	var start uint32
	var trapWire uint16
	active, returned, visible := false, false, false
	sounds, controls := 0, 0
	entries := map[int]bool{}
	noxConsole.Register(&console.Command{Token: "e2e-skull-raw-native", HelpID: "nohelp", Flags: console.Server | console.NoHelp, Func: func(_ context.Context, _ *console.Console, args []string) bool {
		if len(args) != 1 {
			return false
		}
		fail := func(message string) { e2eError(fmt.Errorf("Skull raw C diagnostic: %s", message)) }
		switch args[0] {
		case "start":
			if active || source != nil || noxServer.Players.HostUnit() == nil {
				fail("one hosted start required")
				return true
			}
			host = noxServer.Players.HostUnit()
			if host.HealthData == nil || host.HealthData.Cur == 0 || host.Poison540 != 0 || noxServer.TickRate() != 30 || noxServer.Balance.Float("ArrowTrapDamage") != 25 {
				fail("living clean host and original damage25/FPS30 required")
				return true
			}
			hp, start, active = host.HealthData.Cur, noxServer.Frame(), true
			source = noxServer.NewObjectByTypeID("ArrowTrap1")
			if source == nil || source.UpdateData == nil {
				fail("stock native ArrowTrap1 and52-byte update required")
				return true
			}
			noxServer.CreateObjectAt(source, nil, types.Ptf(host.PosVec.X+32, host.PosVec.Y))
			source.Update = nil
			source.ObjFlags |= object.FlagEnabled | object.FlagNoCollide
			source.Shape.Kind, source.Shape.Circle.R = server.ShapeKindCircle, 14
			source.Direction1, source.Direction2 = 128, 77
			id := uint32(noxServer.Types.IndByID("MercArcherArrow"))
			if id == 0 || noxServer.Audio.Field12(889) <= 0 {
				fail("original projectile numeric type and audio descriptor required")
				return true
			}
			if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(source)) <= math.MaxUint32 || uintptr(source.UpdateData) <= math.MaxUint32) {
				fail("native source and update must exceed4GiB")
				return true
			}
			state := memmap.PtrUint32(0x5D4594, 2491764)
			call := func(entry int, target *server.Object) (uint32, bool) {
				value, ok := legacy.SkullRawActualDefaultC54F9A0(entry, source, target, id)
				if !ok {
					fail("actual typed C entry/default services required")
					return 0, false
				}
				entries[entry] = true
				e2eLog.Printf("SKULL RAW NATIVE C ENTRY: entry=%d source=%p update=%p default-world=true prepared-stock=ArrowTrap1", entry, source, source.UpdateData)
				return value, true
			}
			if !noxServer.IsEnemyTo(source, host) || !noxServer.MapTraceVision(host, source) || legacy.Nox_server_testTwoPointsAndDirection_4E6E50(source.PosVec, int16(source.Direction1), host.PosVec)&1 == 0 {
				fail("independent real hostility/map/facing controls")
				return true
			}
			*state = 99
			if _, ok := call(4, host); !ok {
				return true
			}
			if *state != 1 {
				fail("actual raw candidate must accept real host")
				return true
			}
			source.Direction1 = 0
			found, ok := call(3, nil)
			if !ok {
				return true
			}
			d := (*server.SkullUpdateData)(source.UpdateData)
			d.TargetReady, d.FireDelay = 9, 19
			if _, ok = call(2, nil); !ok {
				return true
			}
			expectedState, expectedDelay := uint8(0), uint32(19)
			if found != 0 {
				expectedState, expectedDelay = 1, 0
			}
			if d.TargetReady != expectedState || d.FireDelay != expectedDelay {
				fail("actual scan/wrapper shared marker and state transition")
				return true
			}
			noxServer.Audio.OnSound(func(id sound.ID, kind int, unit *server.Object, position types.Pointf) {
				if !active || unit != source && unit != control {
					return
				}
				if !returned || id != 889 || kind != 0 || position != unit.PosVec || unit.Flags().Has(object.FlagDestroyed) {
					fail("actual queued audio identity and after-entry observation")
					return
				}
				if unit == source {
					sounds++
				} else {
					controls++
				}
				e2eLog.Printf("SKULL RAW NATIVE C AUDIO: control=%v source=%p after-entry=true", unit == control, unit)
			})
			control = noxServer.NewObjectByTypeID("RedPotion")
			if control == nil {
				fail("independent audio control source")
				return true
			}
			noxServer.CreateObjectAt(control, nil, types.Ptf(source.PosVec.X, source.PosVec.Y+24))
			noxServer.Audio.EventObj(889, control, 0, 0)
			verifyProjectile := func() bool {
				p := noxServer.Objs.Pending
				if p == nil || p.ObjOwner != source || uint32(p.TypeInd) != id || p.CollideData == nil || p.Direction1 != 0 || p.Direction2 != 0 || p.PosVec != types.Ptf(source.PosVec.X+18, source.PosVec.Y) || p.VelVec.X != p.SpeedCur || p.VelVec.Y != 0 {
					fail("actual C-created stock projectile owner/position/direction/velocity")
					return false
				}
				c := (*server.ProjectileCollideData)(p.CollideData)
				if c.Damage != 25 || c.Field4 != 25 {
					fail("original stock balance25 conversion and both collide dwords")
					return false
				}
				if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(p)) <= math.MaxUint32 || uintptr(p.CollideData) <= math.MaxUint32) {
					fail("actual native projectile/collide must exceed4GiB")
					return false
				}
				e2eLog.Printf("SKULL RAW NATIVE C PROJECTILE: source=%p object=%p collide=%p original-stock-damage=25 native-velocity=true owner=true", source, p, p.CollideData)
				return true
			}
			*d = server.SkullUpdateData{ScanDelay: 17, FireDelay: 0, TargetReady: 1, ProjectileType: id, Enabled: 1}
			value, ok := call(0, nil)
			if !ok {
				return true
			}
			if value != 29 || d.ScanDelay != 16 || d.FireDelay != 29 || !verifyProjectile() {
				fail("actual raw update fire/timers")
				return true
			}
			if _, ok = call(1, nil); !ok || !verifyProjectile() {
				return true
			}
			returned = true
			if sounds != 0 || controls != 0 {
				fail("console audio must queue until the next reset")
				return true
			}
			trapWire = uint16(noxServer.GetUnitNetCode(source))
			noxServer.TickHook(func() {
				if !active || visible {
					return
				}
				if noxClient.Objs.ByNetCode(trapWire) != nil {
					path, err := e2eWriteMagicFrame("", noxClient.r.CopyPixBuffer())
					if err != nil {
						e2eError(err)
						return
					}
					visible = true
					e2eLog.Printf("SKULL RAW NATIVE C VISIBLE: trap-wire=%d frame=%d path=%s", trapWire, noxServer.Frame(), path)
				} else if noxServer.Frame()-start > 120 {
					fail("stock trap client replication")
				}
			})
		case "finish":
			if !active || !returned || !visible || len(entries) != 5 || sounds != 2 || controls != 1 || host.HealthData.Cur != hp || host.Poison540 != 0 {
				fail(fmt.Sprintf("incomplete entries=%d visible=%v sounds=%d controls=%d", len(entries), visible, sounds, controls))
				return true
			}
			path, err := e2eWriteMagicFrame("", noxClient.r.CopyPixBuffer())
			if err != nil {
				e2eError(err)
				return true
			}
			active = false
			noxServer.DelayedDelete(source)
			noxServer.DelayedDelete(control)
			e2eLog.Printf("SKULL RAW NATIVE C PASS: entries=5 default-world=all original-stock-damage=25 native-projectiles=2 client-trap=true sound-events=2 control-sound-events=1 queued-control=true living-host-HP=%d unchanged prepared-stock=ArrowTrap1 held-ordinary-update=true earned-campaign-credit=false path=%s", hp, path)
		default:
			return false
		}
		return true
	}})
}
