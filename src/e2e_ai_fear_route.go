package opennox

import (
	"fmt"
	"image"
	"math"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

// Plan only the test player's mouse waypoints, independently of monster
// retreat pathfinding. The live NPC path/action/position is never supplied.
// A bounded 23-unit grid explores actual clear wall/prop lanes.
func e2eAIFearRoute(from, to types.Pointf, clear, visible func(types.Pointf, types.Pointf) bool) ([]types.Pointf, bool) {
	if clear == nil || visible == nil {
		return nil, false
	}
	for _, value := range []float32{from.X, from.Y, to.X, to.Y} {
		if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
			return nil, false
		}
	}
	if _, walk := e2eAIFearApproach(from, to); !walk {
		return nil, visible(from, to)
	}
	origin := image.Point{}
	queue := []image.Point{origin}
	parents := map[image.Point]image.Point{origin: origin}
	point := func(p image.Point) types.Pointf { return from.Add(types.Ptf(float32(p.X)*23, float32(p.Y)*23)) }
	for head := 0; head < len(queue) && head < 4096; head++ {
		current := queue[head]
		pos := point(current)
		delta := to.Sub(pos)
		if delta.X*delta.X+delta.Y*delta.Y <= 48*48 && visible(pos, to) {
			var path []types.Pointf
			for current != origin {
				path = append(path, point(current))
				current = parents[current]
			}
			for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
				path[i], path[j] = path[j], path[i]
			}
			return path, true
		}
		for _, step := range []image.Point{{X: 1}, {X: -1}, {Y: 1}, {Y: -1}, {X: 1, Y: 1}, {X: 1, Y: -1}, {X: -1, Y: 1}, {X: -1, Y: -1}} {
			next := current.Add(step)
			if next.X < -48 || next.X > 48 || next.Y < -48 || next.Y > 48 {
				continue
			}
			if _, seen := parents[next]; seen || !clear(pos, point(next)) {
				continue
			}
			parents[next] = current
			queue = append(queue, next)
		}
	}
	return nil, false
}

func (f *e2eAIFearFixture) approach() (types.Pointf, bool) {
	if !f.expired || f.hit {
		return types.Pointf{}, false
	}
	host, target := f.combat.host, f.combat.enemy
	for f.playerRouteIndex < len(f.playerRoute) && host.PosVec.Sub(f.playerRoute[f.playerRouteIndex]).Len() <= 6 {
		f.playerRouteIndex++
	}
	if f.playerRouteIndex >= len(f.playerRoute) {
		if _, walk := e2eAIFearApproach(host.PosVec, target.PosVec); !walk {
			return types.Pointf{}, false
		}
		clear := func(from, to types.Pointf) bool {
			side := types.Ptf(-(to.Y - from.Y), to.X-from.X).Normalize().Mul(host.Shape.Circle.R + 4)
			return e2eWarriorLaneClear(host, from, to) &&
				e2eWarriorLaneClear(host, from.Add(side), to.Add(side)) &&
				e2eWarriorLaneClear(host, from.Sub(side), to.Sub(side))
		}
		visible := func(from, to types.Pointf) bool { return noxServer.MapTraceRay(from, to, server.MapTraceFlag1) }
		var ok bool
		f.playerRoute, ok = e2eAIFearRoute(host.PosVec, target.PosVec, clear, visible)
		f.playerRouteIndex = 0
		if !ok || len(f.playerRoute) == 0 {
			e2eError(fmt.Errorf("no independent player walking route after stock fear: %s %v->%v", f.mode, host.PosVec, target.PosVec))
			return types.Pointf{}, false
		}
		e2eLog.Printf("AI FEAR PLAYER WALK: mode=%s after-expiry=true waypoints=%d wall/prop-clear=true attack-input=none", f.mode, len(f.playerRoute))
	}
	return f.playerRoute[f.playerRouteIndex], true
}
