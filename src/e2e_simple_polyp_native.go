package opennox

import (
	"context"
	"fmt"
	"math"
	"unsafe"

	"github.com/opennox/libs/console"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

// Direct raw-C diagnostic only. Prepared stock RedPotion objects supply
// valid source records to the three class-unrestricted original entries.
// This does not claim natural ImpEgg/Polyp deaths or earned campaign credit.
func init() {
	if e2ePlay == "" {
		return
	}
	var host, control, egg, cloud *server.Object
	var hp uint16
	var start uint32
	var wire uint16
	active, visible := false, false
	type pendingAudio struct {
		entry    int
		sound    sound.ID
		position types.Pointf
		returned bool
	}
	pending := make(map[*server.Object]*pendingAudio)
	sounds, controlSounds := 0, 0
	queuedControl := false
	noxConsole.Register(&console.Command{Token: "e2e-simple-polyp-native", HelpID: "nohelp", Flags: console.Server | console.NoHelp, Func: func(_ context.Context, _ *console.Console, args []string) bool {
		if len(args) != 1 {
			return false
		}
		switch args[0] {
		case "start":
			if active || host != nil || noxServer.Players.HostUnit() == nil {
				e2eError(fmt.Errorf("simple/Polyp C diagnostic requires one hosted start"))
				return true
			}
			host = noxServer.Players.HostUnit()
			if host.HealthData == nil || host.HealthData.Cur == 0 || host.Poison540 != 0 || noxServer.TickRate() != 30 || noxServer.Balance.Float("ToxicCloudLifetime") != 15 {
				e2eError(fmt.Errorf("simple/Polyp C diagnostic requires living unpoisoned host and original lifetime15/FPS30"))
				return true
			}
			hp, start, active = host.HealthData.Cur, noxServer.Frame(), true
			noxServer.Audio.OnSound(func(id sound.ID, kind int, unit *server.Object, position types.Pointf) {
				record := pending[unit]
				if !active || record == nil {
					return
				}
				// Console commands run before Audio.Reset. The real service queues
				// these events; its callback sees the completed deletion on the
				// following frame. Synchronous audio-before-delete order is covered
				// separately by the actual default-C production tests.
				deleted := record.entry == 1 || record.entry == 2
				if !record.returned || id != record.sound || kind != 0 || position != record.position || unit.Flags().Has(object.FlagDestroyed) != deleted {
					e2eError(fmt.Errorf("actual queued C diagnostic audio identity/deletion: entry=%d returned=%v id=%d kind=%d destroyed=%v", record.entry, record.returned, id, kind, unit.Flags().Has(object.FlagDestroyed)))
					return
				}
				delete(pending, unit)
				if record.entry == -1 {
					controlSounds++
					queuedControl = true
				} else {
					sounds++
				}
				e2eLog.Printf("SIMPLE POLYP NATIVE C AUDIO: entry=%d source=%p queued-after-entry=true destroyed=%v", record.entry, unit, deleted)
			})
			position := types.Ptf(host.PosVec.X+220, host.PosVec.Y+48)
			control = noxServer.NewObjectByTypeID("RedPotion")
			if control == nil || noxServer.Audio.Field12(764) <= 0 || noxServer.Audio.Field12(284) <= 0 || noxServer.Audio.Field12(753) <= 0 {
				e2eError(fmt.Errorf("stock source and original audio descriptors required"))
				return true
			}
			noxServer.CreateObjectAt(control, nil, position)
			controlRecord := &pendingAudio{entry: -1, sound: 764, position: position}
			pending[control] = controlRecord
			// Independent world-service control: no death entry is invoked.
			// This is the same server service reached by the typed C audio export.
			noxServer.Audio.EventObj(764, control, 0, 0)
			controlRecord.returned = true
			if controlSounds != 0 || control.Flags().Has(object.FlagDestroyed) {
				e2eError(fmt.Errorf("console audio control must queue before the next Audio.Reset"))
				return true
			}
			for entry, id := range []sound.ID{764, 284, 753} {
				source := noxServer.NewObjectByTypeID("RedPotion")
				if source == nil {
					e2eError(fmt.Errorf("stock prepared source allocation"))
					return true
				}
				noxServer.CreateObjectAt(source, nil, position)
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(source)) <= math.MaxUint32 {
					e2eError(fmt.Errorf("actual C source below4GiB"))
					return true
				}
				before, flags := noxServer.Objs.Pending, source.ObjFlags
				record := &pendingAudio{entry: entry, sound: id, position: position}
				pending[source] = record
				value, ok := legacy.SimplePolypDeathActualDefaultC54CAE0(entry, source)
				record.returned = true
				if !ok || sounds != 0 {
					e2eError(fmt.Errorf("actual public C entry=%d accepted=%v premature-audio=%d result=%08x", entry, ok, sounds, value))
					return true
				}
				switch entry {
				case 0:
					if source.ObjFlags != flags|object.FlagNoCollide || value != uint32(source.ObjFlags) || source.Flags().Has(object.FlagDestroyed) {
						e2eError(fmt.Errorf("actual ImpEgg C full flags/no deletion"))
						return true
					}
					egg = source
				case 1:
					cloud = noxServer.Objs.Pending
					if cloud == nil || cloud == before || cloud.ObjNext != before || cloud.ObjectTypeC() == nil || cloud.ObjectTypeC().ID() != "ToxicCloud" || cloud.ObjOwner != nil || cloud.PosVec != position || cloud.UpdateData == nil || (*server.ToxicCloudUpdateData)(cloud.UpdateData).Duration != 450 {
						e2eError(fmt.Errorf("actual Polyp C original stock cloud/placement/duration450"))
						return true
					}
					if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(cloud)) <= math.MaxUint32 || uintptr(cloud.UpdateData) <= math.MaxUint32) {
						e2eError(fmt.Errorf("actual cloud/update below4GiB"))
						return true
					}
					wire = uint16(noxServer.GetUnitNetCode(cloud))
					fallthrough
				case 2:
					if !source.Flags().Has(object.FlagDestroyed) || noxServer.Objs.DeletedList != source {
						e2eError(fmt.Errorf("actual C source not delayed-deleted after audio"))
						return true
					}
				}
				e2eLog.Printf("SIMPLE POLYP NATIVE C ENTRY: entry=%d source=%p default-world=true prepared-source=RedPotion original-stock-balance15-FPS30=true", entry, source)
			}
			noxServer.TickHook(func() {
				if !active || visible {
					return
				}
				if noxClient.Objs.ByNetCode(wire) != nil {
					path, err := e2eWriteMagicFrame("", noxClient.r.CopyPixBuffer())
					if err != nil {
						e2eError(err)
						return
					}
					visible = true
					e2eLog.Printf("SIMPLE POLYP NATIVE C VISIBLE: cloud-wire=%d frame=%d path=%s", wire, noxServer.Frame(), path)
				} else if noxServer.Frame()-start > 120 {
					active = false
					e2eError(fmt.Errorf("actual C cloud did not replicate"))
				}
			})
		case "finish":
			if !active || !visible || sounds != 3 || controlSounds != 1 || !queuedControl || len(pending) != 0 || host == nil || host.HealthData.Cur != hp || host.Poison540 != 0 {
				e2eError(fmt.Errorf("simple/Polyp C diagnostic incomplete visible=%v sounds=%d controls=%d pending=%d", visible, sounds, controlSounds, len(pending)))
				return true
			}
			path, err := e2eWriteMagicFrame("", noxClient.r.CopyPixBuffer())
			if err != nil {
				e2eError(err)
				return true
			}
			active = false
			noxServer.DelayedDelete(control)
			noxServer.DelayedDelete(egg)
			noxServer.DelayedDelete(cloud)
			e2eLog.Printf("SIMPLE POLYP NATIVE C PASS: entries=3 default-world=all original-stock-lifetime=450 client-cloud=true sound-events=3 control-sound-events=1 queued-control=true living-host-HP=%d unchanged direct-C-diagnostic=true prepared-source=RedPotion earned-campaign-credit=false path=%s", hp, path)
		default:
			return false
		}
		return true
	}})
}
