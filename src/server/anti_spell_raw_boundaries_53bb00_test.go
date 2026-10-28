package server

import (
	"github.com/opennox/libs/types"
	"testing"
)

func TestAntiSpellRawBoundaries53BB00(t *testing.T) {
	for _, name := range []string{"root-nil", "root-nil-update-alive", "root-nil-update-expired", "candidate-nil", "missile-nil", "destroyed-candidate-nil-missile", "wrong-class-nil-missile", "circle-nil-origin"} {
		t.Run(name, func(t *testing.T) {
			rt := AntiSpellProjectileRuntime53BB00{Frame: func() uint32 { return 20 }, FPS: func() uint32 { return 30 }}
			mustFault := name == "root-nil" || name == "root-nil-update-alive" || name == "candidate-nil" || name == "missile-nil" || name == "circle-nil-origin"
			if mustFault {
				defer func() {
					if recover() == nil {
						t.Fatal("original raw pointer precondition did not fault")
					}
				}()
			}
			switch name {
			case "root-nil":
				AntiSpellProjectileUpdate53BB00(nil, rt)
				AntiSpellProjectileRaw53BB00(nil, rt)
			case "root-nil-update-alive":
				m := new(Object)
				AntiSpellProjectileUpdate53BB00(m, rt)
				AntiSpellProjectileRaw53BB00(m, rt)
			case "root-nil-update-expired":
				m := new(Object)
				deleted := false
				rt.Frame = func() uint32 { return 1000 }
				rt.DelayedDelete = func(o *Object) {
					if o != m {
						t.Fatal("original expiry pointer")
					}
					deleted = true
				}
				AntiSpellProjectileRaw53BB00(m, rt)
				if !deleted {
					t.Fatal("expired original must delete without reading nil update data")
				}
			case "candidate-nil":
				AntiSpellCandidateRaw53BD10(nil, new(Object), rt)
			case "missile-nil":
				AntiSpellCandidateRaw53BD10(new(Object), nil, rt)
			case "destroyed-candidate-nil-missile":
				AntiSpellCandidateRaw53BD10(&Object{ObjFlags: 0x20}, nil, rt)
			case "wrong-class-nil-missile":
				AntiSpellCandidateRaw53BD10(&Object{ObjClass: 1}, nil, rt)
			case "circle-nil-origin":
				s := new(Server)
				s.Map.EachMissileInCirclePtr518170((*types.Pointf)(nil), 600, func(*Object) bool { return true })
			}
		})
	}
}

func TestAntiSpellSharedStateServerIsolation53BD10(t *testing.T) {
	a, b := new(Server), new(Server)
	x, y := a.AntiSpellTargetState53BD10(), b.AntiSpellTargetState53BD10()
	if x == y || x.Target != nil || x.BestDistance != 0 || y.Target != nil || y.BestDistance != 0 {
		t.Fatal("fresh native server search state")
	}
	x.Target = new(Object)
	x.BestDistance = 100_000_000
	if a.AntiSpellTargetState53BD10() != x || y.Target != nil || y.BestDistance != 0 {
		t.Fatal("server-owned shared state was lost or leaked to another world")
	}
}
