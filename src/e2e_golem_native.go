package opennox

import (
	"context"
	"fmt"
	"github.com/opennox/libs/console"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
	"math"
	"unsafe"
)

// Explicit E2E-only direct-C diagnostic with stock world services. The living
// host is a source; no natural Golem kill or earned campaign credit is claimed.
func init() {
	if e2ePlay == "" {
		return
	}
	var host *server.Object
	var created, mechs, rocks []*server.Object
	var hp uint16
	var start uint32
	active, visible := false, false
	sounds := 0
	noxConsole.Register(&console.Command{Token: "e2e-golem-native", HelpID: "nohelp", Flags: console.Server | console.NoHelp, Func: func(_ context.Context, _ *console.Console, args []string) bool {
		if len(args) != 1 {
			return false
		}
		switch args[0] {
		case "start":
			if active || host != nil || noxServer.Players.HostUnit() == nil {
				e2eError(fmt.Errorf("Golem C diagnostic requires one hosted start"))
				return true
			}
			host = noxServer.Players.HostUnit()
			if host.HealthData == nil || host.HealthData.Cur == 0 {
				e2eError(fmt.Errorf("Golem C diagnostic requires living host"))
				return true
			}
			hp = host.HealthData.Cur
			start = noxServer.Frame()
			active = true
			noxServer.Audio.OnSound(func(id sound.ID, kind int, source *server.Object, _ types.Pointf) {
				if active && source == host && (id == 487 || id == 494) {
					if kind != 0 {
						e2eError(fmt.Errorf("Golem sound kind"))
					}
					sounds++
				}
			})
			alreadyCoop := noxflags.HasGame(noxflags.GameModeCoop)
			defer func() {
				if alreadyCoop {
					noxflags.SetGame(noxflags.GameModeCoop)
				} else {
					noxflags.UnsetGame(noxflags.GameModeCoop)
				}
			}()
			index := legacy.GolemDeathIndex549E90()
			oldIndex := *index
			*index = 11
			defer func() { *index = oldIndex }()
			for _, entry := range []string{"mech", "golem-coop", "golem-solo"} {
				mech := entry == "mech"
				if entry == "golem-solo" {
					noxflags.UnsetGame(noxflags.GameModeCoop)
				} else {
					noxflags.SetGame(noxflags.GameModeCoop)
				}
				before := noxServer.Objs.Pending
				initial := *index
				value, ok := legacy.GolemDeathActualDefaultC549E90(mech, host)
				if !ok || value != 1 {
					e2eError(fmt.Errorf("actual default C %s rejected", entry))
					return true
				}
				list := []*server.Object{}
				for o := noxServer.Objs.Pending; o != before; o = o.ObjNext {
					if o == nil || o.ObjectTypeC() == nil || o.ObjOwner != nil || !o.Flags().Has(object.FlagBouncy) {
						e2eError(fmt.Errorf("Golem C creation list/flags/owner"))
						return true
					}
					if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(o)) <= math.MaxUint32 {
						e2eError(fmt.Errorf("Golem C pointer below4GiB"))
						return true
					}
					list = append(list, o)
					created = append(created, o)
				}
				count := len(list)
				if mech {
					if count != 8 || *index != initial {
						e2eError(fmt.Errorf("actual Mech count/cache"))
						return true
					}
				} else if (entry == "golem-solo" && count != 6) || (entry == "golem-coop" && (count < 20 || count > 30)) || *index != (initial+uint32(count))%12 {
					e2eError(fmt.Errorf("actual Golem count/cache %s count=%d", entry, count))
					return true
				}
				for i, o := range list {
					name := o.ObjectTypeC().ID()
					mass := float32(2)
					want := []string{"MechGear", "MechHead", "MechArm", "MechGear", "MechLeg", "MechLeg", "MechGear", "MechGear"}[(count-1-i)%8]
					if mech {
						mechs = append(mechs, o)
					} else {
						want = []string{"BigRock", "MediumRock", "SmallRock", "MediumRock", "SmallRock", "SmallRock", "MediumRock", "SmallRock", "SmallRock", "SmallRock", "SmallRock", "SmallRock"}[(int(initial)+count-1-i)%12]
						mass = map[string]float32{"BigRock": 1, "MediumRock": 3, "SmallRock": 5}[name]
						rocks = append(rocks, o)
					}
					if name != want || mass == 0 || o.Field29 != math.Float32bits(mass) || o.Field27 < -2 || o.Field27 > 0 || o.ZVal < 10 || o.ZVal > 70 {
						e2eError(fmt.Errorf("actual Golem native table/motion/mass %s", name))
						return true
					}
				}
				e2eLog.Printf("GOLEM NATIVE C ENTRY: entry=%s source=%p pieces=%d initial-index=%d final-index=%d default-world=true", entry, host, count, initial, *index)
			}
			if host.HealthData.Cur != hp {
				e2eError(fmt.Errorf("actual Golem C host health changed"))
				return true
			}
			noxServer.TickHook(func() {
				if !active || visible {
					return
				}
				clientCount := func(objects []*server.Object) int {
					n := 0
					for _, o := range objects {
						if noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(o))) != nil {
							n++
						}
					}
					return n
				}
				m, r := clientCount(mechs), clientCount(rocks)
				if m > 0 && r > 0 {
					path, err := e2eWriteMagicFrame("", noxClient.r.CopyPixBuffer())
					if err != nil {
						e2eError(err)
						return
					}
					visible = true
					e2eLog.Printf("GOLEM NATIVE C VISIBLE: mechs=%d rocks=%d frame=%d path=%s", m, r, noxServer.Frame(), path)
				} else if noxServer.Frame()-start > 120 {
					active = false
					e2eError(fmt.Errorf("Golem C debris did not replicate"))
				}
			})
		case "finish":
			if !active || !visible || sounds != 3 || host == nil || host.HealthData.Cur != hp {
				e2eError(fmt.Errorf("Golem C diagnostic incomplete visible=%v sounds=%d", visible, sounds))
				return true
			}
			path, err := e2eWriteMagicFrame("", noxClient.r.CopyPixBuffer())
			if err != nil {
				e2eError(err)
				return true
			}
			active = false
			for _, o := range created {
				noxServer.DelayedDelete(o)
			}
			e2eLog.Printf("GOLEM NATIVE C PASS: entries=2 calls=3 default-world=all stock-tables=true byte-masses=true client-mechs-and-rocks=true sound-events=3 living-host-HP=%d unchanged direct-C-diagnostic=true earned-campaign-credit=false path=%s", hp, path)
		default:
			return false
		}
		return true
	}})
}
