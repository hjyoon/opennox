package server

import "github.com/opennox/libs/types"

// MonsterProjectileLead533080 follows the original precision-53 / ToZero
// prediction and returns the native target pointer. The two output stores
// retain their original order, including when output aliases object fields.
func MonsterProjectileLead533080(unit, target *Object, speed float32, output *types.Pointf) *Object {
	add, sub := monsterMissileAdd533080, monsterMissileSub533080
	mul, div, spill := monsterMissileMul533080, monsterMissileDiv533080, monsterMissileSpill533080
	load, memory := monsterMissileLoadF32_533080, monsterMissileMemoryF32_533080
	dx := sub(load(target.PosVec.X), memory(unit.PosVec.X))
	dy := sub(load(target.PosVec.Y), memory(unit.PosVec.Y))
	travel := div(monsterMissileSqrt533080(add(mul(dy, dy), mul(dx, dx))), memory(speed))
	output.X = spill(add(mul(travel, memory(target.VelVec.X)), memory(target.PosVec.X)))
	output.Y = spill(add(mul(travel, memory(target.VelVec.Y)), memory(target.PosVec.Y)))
	return target
}

// MonsterProjectileLeadNative533080 has no server or resource dependency.
func (*Server) MonsterProjectileLeadNative533080(unit, target *Object, speed float32, output *types.Pointf) *Object {
	return MonsterProjectileLead533080(unit, target, speed, output)
}
