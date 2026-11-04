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

// This explicit E2E-only diagnostic calls the four public C entries in a
// hosted stock map with their default services. It does not replace natural
// callbacks or claim earned Skeleton kills/campaign progression.
func init() {
	if e2ePlay == "" {
		return
	}
	var host *server.Object
	var bones []*server.Object
	var created []*server.Object
	var hp uint16
	var start uint32
	sounds := 0
	visible := false
	active := false
	noxConsole.Register(&console.Command{Token: "e2e-skeleton-native", HelpID: "nohelp", Flags: console.Server | console.NoHelp, Func: func(_ context.Context, _ *console.Console, args []string) bool {
		if len(args) != 1 {
			return false
		}
		switch args[0] {
		case "start":
			if active || host != nil || noxServer.Players.HostUnit() == nil {
				e2eError(fmt.Errorf("Skeleton C fixture requires one hosted start"))
				return true
			}
			host = noxServer.Players.HostUnit()
			if host.HealthData == nil || host.HealthData.Cur == 0 {
				e2eError(fmt.Errorf("Skeleton C fixture requires living stock host"))
				return true
			}
			hp = host.HealthData.Cur
			start = noxServer.Frame()
			active = true
			noxServer.Audio.OnSound(func(id sound.ID, kind int, source *server.Object, _ types.Pointf) {
				if active && id == 356 && source == host {
					if kind != 0 {
						e2eError(fmt.Errorf("Skeleton sound kind"))
					}
					sounds++
				}
			})
			alreadyCoop := noxflags.HasGame(noxflags.GameModeCoop)
			noxflags.SetGame(noxflags.GameModeCoop)
			defer func() {
				if !alreadyCoop {
					noxflags.UnsetGame(noxflags.GameModeCoop)
				}
			}()
			for _, entry := range []int{0, 3, 2, 1} {
				before := noxServer.Objs.Pending
				value, ok := legacy.SkeletonDeathActualDefaultC54A310(entry, host, "SteelShield", [4]string{"", "Material2", "", ""}, 0)
				if !ok || value != 1 {
					e2eError(fmt.Errorf("actual default C entry %d rejected", entry))
					return true
				}
				pieces, skulls, parts, items := 0, 0, 0, 0
				for o := noxServer.Objs.Pending; o != before; o = o.ObjNext {
					if o == nil || o.ObjectTypeC() == nil || o.ObjOwner != nil {
						e2eError(fmt.Errorf("Skeleton C creation list/owner"))
						return true
					}
					if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(o)) <= math.MaxUint32 {
						e2eError(fmt.Errorf("Skeleton created pointer below4GiB"))
						return true
					}
					created = append(created, o)
					pieces++
					name := o.ObjectTypeC().ID()
					switch name {
					case "Skull":
						skulls++
					case "ArmBone", "LegBone":
						parts++
						bones = append(bones, o)
					default:
						items++
						if entry == 1 {
							attrs := o.InitDataModifier()
							if name != "SteelShield" || attrs == nil || attrs.Modifiers[1] == nil || attrs.Modifiers[1].Name() != "Material2" || attrs.Modifiers[0] != nil || attrs.Modifiers[2] != nil || attrs.Modifiers[3] != nil || attrs.Field16 != 0 {
								e2eError(fmt.Errorf("actual default C stock material modifier"))
								return true
							}
						}
					}
					if name == "Skull" || name == "ArmBone" || name == "LegBone" {
						if !o.Flags().Has(object.FlagBouncy) || o.Field29 != math.Float32bits(4) || o.Field27 < -2 || o.Field27 > 0 {
							e2eError(fmt.Errorf("actual C debris motion"))
							return true
						}
					}
				}
				if entry == 1 {
					if pieces != 1 || items != 1 {
						e2eError(fmt.Errorf("actual C shield creation"))
						return true
					}
				} else if skulls != 1 || parts < 10 || parts > 20 || items > 1 {
					e2eError(fmt.Errorf("actual C debris count entry=%d skulls=%d bones=%d items=%d", entry, skulls, parts, items))
					return true
				}
				e2eLog.Printf("SKELETON NATIVE C ENTRY: entry=%d source=%p pieces=%d bones=%d items=%d index=%d default-world=true", entry, host, pieces, parts, items, *legacy.SkeletonDeathIndex54A310())
			}
			if host.HealthData.Cur != hp {
				e2eError(fmt.Errorf("actual C host health changed"))
				return true
			}
			// Console calls can queue sounds outside the audio loop. The finish
			// step checks all three after normal frame processing replays them.
			noxServer.TickHook(func() {
				if !active || visible {
					return
				}
				count := 0
				for _, o := range bones {
					if noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(o))) != nil {
						count++
					}
				}
				if count > 0 {
					path, err := e2eWriteMagicFrame("", noxClient.r.CopyPixBuffer())
					if err != nil {
						e2eError(err)
						return
					}
					visible = true
					e2eLog.Printf("SKELETON NATIVE C VISIBLE: bones=%d frame=%d path=%s", count, noxServer.Frame(), path)
				} else if noxServer.Frame()-start > 240 {
					active = false
					e2eError(fmt.Errorf("actual C bones did not replicate to client"))
				}
			})
		case "finish":
			if !active || !visible || sounds != 3 || host == nil || host.HealthData.Cur != hp {
				e2eError(fmt.Errorf("Skeleton C fixture incomplete"))
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
			e2eLog.Printf("SKELETON NATIVE C PASS: entries=4 default-world=all stock-material=true client-bones=true sound-events=3 living-host-HP=%d unchanged direct-C-diagnostic=true earned-campaign-credit=false path=%s", hp, path)
		default:
			return false
		}
		return true
	}})
}
