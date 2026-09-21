package opennox

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func newPixieCastTestObject() *server.Object {
	return &server.Object{UpdateData: unsafe.Pointer(new(server.PixieUpdateData))}
}

func TestCastPixiesNative540440CreatesMissingPixies(t *testing.T) {
	owner := &server.Object{}
	target := &server.Object{}
	caster := &server.Object{PosVec: types.Ptf(100, 200)}
	caster.Shape.Circle.R = 6
	pixies := []*server.Object{newPixieCastTestObject(), newPixieCastTestObject()}
	created := make([]*server.Object, 0, len(pixies))
	positions := make([]types.Pointf, 0, len(pixies))
	directions := []int{0, 64, 128}
	lifetimes := []int{30, 90}
	directionIndex, lifetimeIndex := 0, 0
	traceCalls := 0
	audioCalls := 0

	got := castPixiesNative540440(spell.SPELL_PIXIE_SWARM, owner, caster, 2, pixieCastHooks540440{
		pixieType: func() int { return 77 },
		countOwned: func(gotOwner *server.Object, typeInd int32) int {
			if gotOwner != owner || typeInd != 77 {
				t.Fatalf("countOwned(%p, %d), want (%p, 77)", gotOwner, typeInd, owner)
			}
			return 0
		},
		desired: func(levelIndex int) int {
			if levelIndex != 1 {
				t.Fatalf("level index = %d, want 1", levelIndex)
			}
			return 3
		},
		randomInt: func(min, max int) int {
			switch {
			case min == 0 && max == 255:
				value := directions[directionIndex]
				directionIndex++
				return value
			case min == 30 && max == 90:
				value := lifetimes[lifetimeIndex]
				lifetimeIndex++
				return value
			default:
				t.Fatalf("random range = %d..%d", min, max)
				return 0
			}
		},
		traceRay: func(from, to types.Pointf) bool {
			if from != caster.PosVec {
				t.Fatalf("trace origin = %v, want %v", from, caster.PosVec)
			}
			traceCalls++
			return traceCalls != 2
		},
		newPixie: func() *server.Object {
			return pixies[len(created)]
		},
		createAt: func(pixie, gotOwner *server.Object, pos types.Pointf) {
			if gotOwner != owner {
				t.Fatalf("create owner = %p, want %p", gotOwner, owner)
			}
			pixie.ObjOwner = gotOwner
			pixie.PosVec = pos
			created = append(created, pixie)
			positions = append(positions, pos)
		},
		findTarget: func(pixie, gotOwner *server.Object) *server.Object {
			if gotOwner != owner || pixie != pixies[len(created)-1] {
				t.Fatalf("findTarget(%p, %p) after create %#v", pixie, gotOwner, created)
			}
			return target
		},
		frame:    func() uint32 { return 100 },
		tickRate: func() uint32 { return 30 },
		playCastAud: func(spellID spell.ID, gotCaster *server.Object) {
			audioCalls++
			if spellID != spell.SPELL_PIXIE_SWARM || gotCaster != caster {
				t.Fatalf("audio = (%d, %p), want (%d, %p)", spellID, gotCaster, spell.SPELL_PIXIE_SWARM, caster)
			}
		},
	})
	if got != 1 {
		t.Fatalf("cast result = %d, want 1", got)
	}
	if traceCalls != 3 || len(created) != 2 || audioCalls != 1 {
		t.Fatalf("cast calls = trace:%d create:%d audio:%d, want 3/2/1", traceCalls, len(created), audioCalls)
	}

	wantDirections := []server.Dir16{0, 128}
	wantLifetimes := []uint32{100 + 30*30, 100 + 30*90}
	for i, pixie := range created {
		cosine, sine := server.SinCosDir(byte(wantDirections[i]))
		wantPos := caster.PosVec.Add(types.Ptf(10*cosine, 10*sine))
		if math.Abs(float64(positions[i].X-wantPos.X)) > 0.0001 || math.Abs(float64(positions[i].Y-wantPos.Y)) > 0.0001 {
			t.Errorf("pixie %d position = %v, want %v", i, positions[i], wantPos)
		}
		update := pixie.UpdateDataPixie()
		if pixie.ObjOwner != owner || update.Owner != owner || update.Target != target {
			t.Errorf("pixie %d pointers = object owner:%p update owner:%p target:%p", i, pixie.ObjOwner, update.Owner, update.Target)
		}
		if pixie.Direction1 != wantDirections[i] || pixie.Direction2 != wantDirections[i] || pixie.VelVec != (types.Pointf{}) {
			t.Errorf("pixie %d motion = directions %d/%d velocity %v", i, pixie.Direction1, pixie.Direction2, pixie.VelVec)
		}
		if pixie.Pos39 != caster.PosVec || update.SpellID != int32(spell.SPELL_PIXIE_SWARM) ||
			update.Deadline != wantLifetimes[i] || update.LastOwnerVisibleFrame != 100 {
			t.Errorf("pixie %d state = pos39:%v spell:%d deadline:%d visible:%d", i,
				pixie.Pos39, update.SpellID, update.Deadline, update.LastOwnerVisibleFrame)
		}
	}
}

func TestCastPixiesNative540440AtCapDoesNothing(t *testing.T) {
	owner, caster := &server.Object{}, &server.Object{}
	called := false
	got := castPixiesNative540440(spell.SPELL_PIXIE_SWARM, owner, caster, 1, pixieCastHooks540440{
		pixieType:  func() int { return 5 },
		countOwned: func(*server.Object, int32) int { return 2 },
		desired:    func(int) int { return 2 },
		randomInt: func(int, int) int {
			called = true
			return 0
		},
		playCastAud: func(spell.ID, *server.Object) { called = true },
	})
	if got != 1 || called {
		t.Fatalf("at-cap cast = result:%d side-effects:%t, want 1/false", got, called)
	}
}

func TestCastPixiesNative540440RejectsNilObjects(t *testing.T) {
	if got := castPixiesNative540440(spell.SPELL_PIXIE_SWARM, nil, &server.Object{}, 1, pixieCastHooks540440{}); got != 0 {
		t.Fatalf("nil-owner result = %d, want 0", got)
	}
	if got := castPixiesNative540440(spell.SPELL_PIXIE_SWARM, &server.Object{}, nil, 1, pixieCastHooks540440{}); got != 0 {
		t.Fatalf("nil-caster result = %d, want 0", got)
	}
}
