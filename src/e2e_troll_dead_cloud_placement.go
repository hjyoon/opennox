package opennox

import (
	"context"
	"fmt"

	"github.com/opennox/libs/console"
	"github.com/opennox/libs/types"
)

// The first Troll cloud scene retained a genuine placement failure beside a
// cramped BluDeath spawn. This separate E2E-only setup searches other existing
// stock PlayerStart neighborhoods; it does not alter that frozen scene or its
// health, poison, death, immunity, sound, lifetime or rendering assertions.
func init() {
	if e2ePlay == "" {
		return
	}
	noxConsole.Register(&console.Command{
		Token: "e2e-troll-cloud-place", HelpID: "nohelp", Flags: console.Server | console.NoHelp,
		Func: func(_ context.Context, _ *console.Console, args []string) bool {
			if len(args) != 0 {
				return false
			}
			host := noxServer.Players.HostUnit()
			if host == nil {
				e2eError(fmt.Errorf("Troll cloud placement requires the actual host"))
				return true
			}
			radius := host.Shape.Circle.R + 40
			for start := noxServer.Objs.First(); start != nil; start = start.Next() {
				if start.ObjectTypeC().ID() != "PlayerStart" {
					continue
				}
				origin, direction, err := e2eWarriorAbilityArena(start.PosVec, radius,
					func(from, to types.Pointf) bool { return e2eWarriorLaneClear(host, from, to) })
				if err != nil {
					continue
				}
				asObjectS(host).SetPos(origin)
				host.VelVec, host.ForceVec, host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
				e2eLog.Printf("TROLL CLOUD EXPLICIT PLACEMENT: stock-start=%p start=%v origin=%v direction=%v validated-radius=%g only-position=true", start, start.PosVec, origin, direction, radius)
				return true
			}
			e2eError(fmt.Errorf("Troll cloud has no stock PlayerStart neighborhood with the unchanged radius=%g", radius))
			return true
		},
	})
}
