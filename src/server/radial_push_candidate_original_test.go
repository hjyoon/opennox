package server

import (
	"github.com/opennox/libs/types"
	"testing"
)

func radialHooksOriginal(o radialOpsOriginal) mapPushUnitsAroundHooks52E040[*Object] {
	return mapPushUnitsAroundHooks52E040[*Object]{eachInRect: o.each, isMovable: o.movable, position: o.position, traceRay: o.trace, mass: o.mass, isMissile: o.missile, loadForce: o.loadForce, storeForce: o.storeForce, beforeForce: o.before, wake: o.wake}
}

func TestRadialPushCandidateOriginal(t *testing.T) {
	count := 0
	for _, row := range radialRowsOriginal(t) {
		if row.Mode != "actual-startup" || row.Input.Kind != "candidate" {
			continue
		}
		count++
		t.Run(row.Input.Name, func(t *testing.T) {
			radialFixtureOriginal(t, row, func(o radialOpsOriginal) {
				mapPushUnit52E0E0(o.first, MapPushUnitArguments52E0E0{Origin: func() types.Pointf { return *o.origin() }, Radius: o.radius, Inner: o.inner, ScaledForce: o.force}, radialHooksOriginal(o))
			})
		})
	}
	if count != 104 {
		t.Fatalf("candidate conditions=%d", count)
	}
}
