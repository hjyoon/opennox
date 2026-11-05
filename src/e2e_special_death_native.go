package opennox

import (
	"context"
	"fmt"
	"math"
	"unsafe"

	"github.com/opennox/libs/console"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

// Playback-only direct-C diagnostic. Stock constructors and world services
// execute; source updates are held and the generator script is disabled. The
// temporary Quest gate exercises real credit in this private runtime, without
// claiming a natural kill, script execution or earned campaign progression.
func init() {
	if e2ePlay == "" {
		return
	}
	var host *server.Object
	var hp uint16
	var start uint32
	var created, rocks, generators []*server.Object
	active, visible := false, false
	sources := make(map[*server.Object]string)
	sounds := make(map[string]int)
	calls := 0
	fail := func(message string) { e2eError(fmt.Errorf("special death C diagnostic: %s", message)) }
	noxConsole.Register(&console.Command{Token: "e2e-special-death-native", HelpID: "nohelp", Flags: console.Server | console.NoHelp, Func: func(_ context.Context, _ *console.Console, args []string) bool {
		if len(args) != 1 {
			return false
		}
		switch args[0] {
		case "start":
			if active || host != nil || noxServer.Players.HostUnit() == nil {
				fail("one hosted start required")
				return true
			}
			host = noxServer.Players.HostUnit()
			if host.HealthData == nil || host.HealthData.Cur == 0 || host.UpdateData == nil || noxServer.TickRate() != 30 {
				fail("living host, complete native update and stock FPS30 required")
				return true
			}
			hp, start, active = host.HealthData.Cur, noxServer.Frame(), true
			from, _, err := e2eAIFirstAttackArena(host, 30, false)
			if err != nil {
				e2eError(err)
				return true
			}
			noxServer.Audio.OnSound(func(id sound.ID, kind int, source *server.Object, pos types.Pointf) {
				name, ok := sources[source]
				if !active || !ok || (id != 757 && id != 1000) {
					return
				}
				if kind != 0 || pos != source.PosVec || (name == "generator" && id != 1000) || (name != "generator" && id != 757) {
					fail("actual sound ID/source/position")
					return
				}
				sounds[name]++
				e2eLog.Printf("SPECIAL DEATH NATIVE AUDIO: case=%s sound=%d source=%p frame=%d actual-server-event=true", name, id, source, noxServer.Frame())
			})
			prepare := func(handler, label string, owner *server.Object) *server.Object {
				death, size, ok := server.ObjectDeathHandler(handler)
				if !ok || death == nil || size != 0 {
					fail("missing stock death handler " + handler)
					return nil
				}
				var u *server.Object
				for _, typ := range noxServer.Types.List() {
					if typ.Death == death {
						u = noxServer.NewObjectByTypeID(typ.ID())
						break
					}
				}
				if u == nil || u.Death != death {
					fail("missing actual stock constructor for " + handler)
					return nil
				}
				noxServer.CreateObjectAt(u, owner, from)
				noxServer.ObjectsAddPending()
				u.Update = nil
				sources[u] = label
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(u)) <= math.MaxUint32 {
					fail("stock source pointer must exceed4GiB")
					return nil
				}
				e2eLog.Printf("SPECIAL DEATH NATIVE SOURCE: case=%s handler=%s type=%s source=%p update=%p class=%#x stock-constructor=true", label, handler, u.ObjectTypeC().ID(), u, u.UpdateData, uint32(u.ObjClass))
				return u
			}
			call := func(entry int, u *server.Object, label string) bool {
				if !legacy.SpecialDeathActualDefaultC54E460(entry, u) {
					fail("actual typed public C symbol/default services required")
					return false
				}
				calls++
				e2eLog.Printf("SPECIAL DEATH NATIVE C ENTRY: case=%s entry=%d source=%p update=%p default-world=true stock-source=true", label, entry, u, u.UpdateData)
				return true
			}
			marker := prepare("MarkerDie", "marker", host)
			if marker == nil {
				return true
			}
			update := (*server.PlayerUpdateData)(host.UpdateData)
			oldSlots := update.Field29
			defer func() { update.Field29 = oldSlots }()
			update.Field29[0], update.Field29[2] = marker, marker
			if !call(0, marker, "marker") {
				return true
			}
			if !marker.Flags().Has(object.FlagDestroyed) || update.Field29[0] != nil || update.Field29[1] != oldSlots[1] || update.Field29[2] != marker || update.Field29[3] != oldSlots[3] {
				fail("first matching native marker slot only and actual delayed delete")
				return true
			}
			update.Field29 = oldSlots
			e2eLog.Printf("SPECIAL DEATH NATIVE MARKER: first-slot-cleared=true duplicate-slot-preserved=true neighboring-slots-preserved=true deleted=true original-slots-restored=true")
			index := legacy.BoulderDeathIndex54E4B0()
			oldIndex := *index
			defer func() { *index = oldIndex }()
			*index = 11
			boulder := prepare("BoulderDie", "boulder", nil)
			if boulder == nil {
				return true
			}
			before := noxServer.Objs.Pending
			if !call(1, boulder, "boulder") {
				return true
			}
			for u := noxServer.Objs.Pending; u != before; u = u.ObjNext {
				if u == nil {
					fail("native Boulder creation list")
					return true
				}
				rocks = append(rocks, u)
				created = append(created, u)
			}
			if len(rocks) < 20 || len(rocks) > 30 || *index != (11+uint32(len(rocks)))%12 || !boulder.Flags().Has(object.FlagDestroyed) {
				fail("actual Boulder count/shared C cache/delete")
				return true
			}
			table := []string{"BigRock", "MediumRock", "SmallRock", "MediumRock", "SmallRock", "SmallRock", "MediumRock", "SmallRock", "SmallRock", "SmallRock", "SmallRock", "SmallRock"}
			for i, u := range rocks {
				want := table[(11+len(rocks)-1-i)%12]
				mass := map[string]float32{"BigRock": 1, "MediumRock": 3, "SmallRock": 5}[want]
				if u.ObjectTypeC().ID() != want || u.ObjOwner != nil || !u.Flags().Has(object.FlagBouncy) || u.Field29 != math.Float32bits(mass) || u.Field27 < -2 || u.Field27 > 0 || u.ZVal < 10 || u.ZVal > 70 {
					fail("actual stock names/unsigned masses/motion/ownership")
					return true
				}
			}
			e2eLog.Printf("SPECIAL DEATH NATIVE BOULDER: pieces=%d initial-index=11 final-index=%d names=true masses=true motion=true deleted=true", len(rocks), *index)
			failed := prepare("BoulderDie", "boulder-lookup-failure", nil)
			if failed == nil {
				return true
			}
			*index = 11
			slot := memmap.PtrPtr(0x587000, 291512+4*11)
			oldName := *slot
			before = noxServer.Objs.Pending
			ok := func() bool {
				*slot = nil
				defer func() { *slot = oldName }()
				return call(1, failed, "boulder-lookup-failure")
			}()
			if !ok {
				return true
			}
			if failed.Flags().Has(object.FlagDestroyed) || *index != 11 || noxServer.Objs.Pending != before || *slot != oldName {
				fail("real nil-name constructor failure must return without deletion/cache changes")
				return true
			}
			e2eLog.Printf("SPECIAL DEATH NATIVE BOULDER FAILURE: constructor-name=nil pieces=0 cache=11 deleted=false stock-table-restored=true OOM-credit=false")
			noxServer.DelayedDelete(failed)
			gen := prepare("MonsterGeneratorDie", "generator", nil)
			if gen == nil || gen.UpdateData == nil {
				fail("complete stock generator update required")
				return true
			}
			g := (*server.MonsterGenUpdateData)(gen.UpdateData)
			g.Field56, g.FuncInd60 = math.MaxUint32, 0
			gen.Obj130 = host
			creditBefore := legacy.SpecialDeathActualPlayerWords54E460(host)
			before = noxServer.Objs.Pending
			oldQuest := noxflags.HasGame(noxflags.GameModeQuest)
			ok = func() bool {
				noxflags.SetGame(noxflags.GameModeQuest)
				defer func() {
					if !oldQuest {
						noxflags.UnsetGame(noxflags.GameModeQuest)
					}
				}()
				return call(2, gen, "generator")
			}()
			if !ok {
				return true
			}
			creditAfter := legacy.SpecialDeathActualPlayerWords54E460(host)
			if creditAfter != [4]uint32{creditBefore[0] + 1, creditBefore[1] | 8, creditBefore[2], creditBefore[3]} || !gen.Flags().Has(object.FlagDestroyed) || noxflags.HasGame(noxflags.GameModeQuest) != oldQuest {
				fail("actual typed credit/delete and restored Quest gate")
				return true
			}
			for u := noxServer.Objs.Pending; u != before; u = u.ObjNext {
				if u == nil || u.ObjectTypeC().ID() != "DestroyedGenerator" || u.PosVec != gen.PosVec || u.ObjOwner != nil {
					fail("actual stock generator remnant constructor/live position/create")
					return true
				}
				generators = append(generators, u)
				created = append(created, u)
			}
			if len(generators) != 1 || host.HealthData.Cur != hp {
				fail("one remnant and unchanged living host")
				return true
			}
			e2eLog.Printf("SPECIAL DEATH NATIVE GENERATOR: remnants=1 live-position=true deleted=true credit-before=%d credit-after=%d flags-before=%d flags-after=%d real-script-service=disabled quest-gate-restored=true", creditBefore[0], creditAfter[0], creditBefore[1], creditAfter[1])
			for _, u := range created {
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(u)) <= math.MaxUint32 {
					fail("all native constructor results must exceed4GiB")
					return true
				}
			}
			noxServer.TickHook(func() {
				if !active || visible {
					return
				}
				count := func(units []*server.Object) int {
					n := 0
					for _, u := range units {
						if noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(u))) != nil {
							n++
						}
					}
					return n
				}
				r, g := count(rocks), count(generators)
				if r > 0 && g == 1 {
					path, err := e2eWriteMagicFrame("", noxClient.r.CopyPixBuffer())
					if err != nil {
						e2eError(err)
						return
					}
					visible = true
					e2eLog.Printf("SPECIAL DEATH NATIVE VISIBLE: rocks=%d generators=%d frame=%d path=%s", r, g, noxServer.Frame(), path)
				} else if noxServer.Frame()-start > 120 {
					active = false
					fail("stock debris/remnant client replication timed out")
				}
			})
		case "finish":
			if !active || !visible || calls != 4 || sounds["boulder"] != 1 || sounds["boulder-lookup-failure"] != 1 || sounds["generator"] != 1 || host == nil || host.HealthData.Cur != hp {
				fail(fmt.Sprintf("incomplete calls=%d visible=%v sounds=%v", calls, visible, sounds))
				return true
			}
			path, err := e2eWriteMagicFrame("", noxClient.r.CopyPixBuffer())
			if err != nil {
				e2eError(err)
				return true
			}
			active = false
			for _, u := range created {
				noxServer.DelayedDelete(u)
			}
			e2eLog.Printf("SPECIAL DEATH NATIVE PASS: entries=3 calls=4 first-marker-only=true stock-rocks=%d constructor-lookup-failure-preserved=true generator-remnant=1 real-player-credit=true client-rocks-and-generator=true sound-events=3 living-host-HP=%d unchanged direct-C-diagnostic=true earned-campaign-credit=false physical-audio-credit=false path=%s", len(rocks), hp, path)
		default:
			return false
		}
		return true
	}})
}
