package server

import (
	"unsafe"

	"github.com/opennox/libs/types"
)

// MonsterActionCastRuntime5413B0 supplies the side effects of GAME.EXE
// 00541300..0054148F without narrowing object or spell-argument pointers.
type MonsterActionCastRuntime5413B0 struct {
	CastSpell   func(id int32, caster *Object, arg *SpellAcceptArg)
	AudioEvent  func(id uint32, unit *Object)
	RandomFloat func(min, max float32) float64
}

func (s *Server) monsterCastRandomFloat541490(runtime MonsterActionCastRuntime5413B0, min, max float32) float64 {
	if runtime.RandomFloat != nil {
		return runtime.RandomFloat(min, max)
	}
	return logicRandomFloat416030(s.Rand.Logic, min, max)
}

// MonsterActionCast5413B0 follows the original cast-frame gate and builds the
// native-width SpellAcceptArg. On PE32 the legacy caller still uses the C path.
func (s *Server) MonsterActionCast5413B0(unit *Object, mode int, runtime MonsterActionCastRuntime5413B0) {
	if unit == nil || unit.UpdateData == nil {
		return
	}
	update := unit.UpdateDataMonster()
	index := int(update.AIStackInd)
	if index < 0 || index >= len(update.AIStack) {
		return
	}
	head := &update.AIStack[index]
	if update.Field120_2 != 0 || update.MonsterDef == nil {
		return
	}
	if uint32(update.Field120_1) != update.MonsterDef.MissileAttackFrame216 {
		if update.Field120_1 == 1 && update.SoundSet122 != nil && runtime.AudioEvent != nil {
			runtime.AudioEvent(*(*uint32)(unsafe.Add(update.SoundSet122, 56)), unit)
		}
		return
	}

	arg := SpellAcceptArg{}
	switch mode {
	case 0:
		arg.Obj = head.ArgObj(2)
		if arg.Obj == nil {
			return
		}
		arg.Pos = s.monsterCastRandomRecoil541490(update, arg.Obj, runtime)
	case 1:
		arg.Pos = head.ArgPos(2)
	default:
		return
	}
	if unit.HasEnchant(ENCHANT_ANTI_MAGIC) {
		return
	}
	unit.Direction2 = DirFromVec(arg.Pos.Sub(unit.PosVec))
	if runtime.CastSpell != nil {
		runtime.CastSpell(int32(head.ArgU32(0)), unit, &arg)
	}
}

func (s *Server) monsterCastRandomRecoil541490(update *MonsterUpdateData, target *Object, runtime MonsterActionCastRuntime5413B0) types.Pointf {
	variance := float32(1) - update.Field330
	p := target.PosVec
	recoil := s.monsterCastRandomFloat541490(runtime, variance, variance+1)
	p.X = float32(float64(p.X) - recoil*float64(target.VelVec.X)*6)
	p.Y = float32(float64(p.Y) - recoil*float64(target.VelVec.Y)*6)
	spread := variance*float32(0.80000001) + 0.2
	p.X = float32(s.monsterCastRandomFloat541490(runtime, -60, 60)*float64(spread) + float64(p.X))
	p.Y = float32(s.monsterCastRandomFloat541490(runtime, -60, 60)*float64(spread) + float64(p.Y))
	return p
}
