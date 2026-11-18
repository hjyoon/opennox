package server

import (
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"
)

type PixieCastRuntime540440 struct {
	PixieType     func() int
	CountOwned    func(owner *Object, typeID int32) int32
	Desired       func(levelIndex int) float64
	RandomInt     func(min, max int) int
	Direction     func(uint8) types.Pointf
	TraceRay      func(from, to types.Pointf) bool
	NewPixie      func() *Object
	CreateAt      func(pixie, owner *Object, position types.Pointf)
	FindTarget    func(pixie, owner *Object) *Object
	Frame         func() uint32
	TickRate      func() uint32
	PlayCastAudio func(spell.ID, *Object)
}

func CastPixiesNative540440(id spell.ID, owner, caster *Object, level int, r PixieCastRuntime540440) int {
	if owner == nil || caster == nil {
		return 0
	}
	add := pointOnLineAddChop53_57C8A0
	mul := pointOnLineMulChop53_57C8A0
	spill := pointOnLineSpill57C8A0
	radius := float64(spill(add(float64(caster.Shape.Circle.R), 4)))
	current := r.CountOwned(owner, int32(r.PixieType()))
	index := int(int32(level) - 1)
	if current >= x87TruncSignedQwordLow566DCC(r.Desired(index)) {
		return 1
	}
	// The second table read controls the number of attempts. Entering the
	// spawn branch still plays the cast sound when that count becomes zero.
	for attempts := x87TruncSignedQwordLow566DCC(r.Desired(index)) - current; attempts > 0; attempts-- {
		direction := r.RandomInt(0, 255)
		vec := r.Direction(uint8(direction))
		from := caster.PosVec
		spawn := types.Ptf(spill(add(mul(radius, float64(vec.X)), float64(from.X))),
			spill(add(mul(radius, float64(vec.Y)), float64(from.Y))))
		if !r.TraceRay(from, spawn) {
			continue
		}
		pixie := r.NewPixie()
		if pixie == nil {
			continue
		}
		update := pixie.UpdateDataPixie()
		r.CreateAt(pixie, owner, spawn)
		pixie.Direction1, pixie.Direction2 = Dir16(direction), Dir16(direction)
		pixie.VelVec = types.Pointf{}
		update.Target = r.FindTarget(pixie, owner)
		update.Owner = owner
		update.SpellID = int32(id)
		pixie.Pos39 = caster.PosVec
		lifetime := uint32(r.RandomInt(30, 90))
		update.Deadline = lifetime*r.TickRate() + r.Frame()
		update.LastOwnerVisibleFrame = r.Frame()
	}
	r.PlayCastAudio(id, caster)
	return 1
}
