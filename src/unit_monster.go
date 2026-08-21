package opennox

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func objectMonsterInit(obj *server.Object) {
	obj.Server().MonsterInit4F0040(obj, server.MonsterInitRuntime4F0040{
		SetHealth: legacy.Nox_xxx_unitSetHP_4E4560,
	})
}

func nox_xxx_monsterCreateFn_54C480(u *server.Object) {
	s := noxServer
	ud := u.UpdateDataMonster()
	name := s.Types.ByInd(int(u.TypeInd)).ID()
	ud.SoundSet122 = legacy.Nox_xxx_getDefaultSoundSet_424350(name)
	def := legacy.Nox_xxx_monsterDefByTT_517560(int(u.TypeInd))
	ud.MonsterDef = def
	if def != nil {
		h := u.HealthData
		u.Experience = float32(def.Experience64)
		h.Cur = uint16(def.Health68)
		h.Field2 = uint16(def.Health68)
		h.Max = uint16(def.Health68)
		speed := float64(def.Speed76) / 32
		u.SpeedBase = float32(speed)
		u.SpeedCur = float32(speed)
		ud.RetreatLevel = def.RetreatRatio80
		ud.Field335 = 1
		ud.ResumeLevel = def.ResumeRatio84
		ud.Field337 = 1
		ud.StatusFlags = def.StatusFlags92
		ud.Field361 = 1
		ud.FleeRange = def.FleeRange88
	}
	ud.AIStackInd = 0
	ud.AIStack[0].Action = uint32(ai.ACTION_IDLE)
	ud.AIAction340 = 0
	ud.Aggression = 0.5
	ud.Aggression2 = 0.5
	ud.Field330 = 0.5
	ud.Field332 = 0.5
	ud.SightRange = 150.0
	ud.Field329 = 30.0
	ud.Field333 = math.MaxUint8
	ud.Field331 = 30
	ud.Field338 = 1.0

	ud.ScriptLookingForEnemy = server.ScriptCallback{Func: -1}
	ud.ScriptEnemySighted = server.ScriptCallback{Func: -1}
	ud.ScriptChangeFocus = server.ScriptCallback{Func: -1}
	ud.ScriptIsHit = server.ScriptCallback{Func: -1}
	ud.ScriptRetreat = server.ScriptCallback{Func: -1}
	ud.ScriptDeath = server.ScriptCallback{Func: -1}
	ud.ScriptCollision = server.ScriptCallback{Func: -1}
	ud.ScriptHearEnemy = server.ScriptCallback{Func: -1}
	ud.ScriptEndOfWaypoint = server.ScriptCallback{Func: -1}
	ud.ScriptLostEnemy = server.ScriptCallback{Func: -1}

	fps := s.TickRate()
	ud.Field510 = 1
	ud.Field362_0 = 0
	ud.Field362_2 = uint16(fps / 2)
	ud.Field364_0 = uint16(fps * 3)
	ud.Field364_2 = uint16(fps * 10)
	ud.Field366_0 = uint16(fps * 3)
	ud.Field366_2 = uint16(fps * 6)
	ud.Field368_0 = 0
	ud.Field368_2 = uint16(fps * 3)
	ud.Field370_0 = 0
	ud.Field370_2 = uint16(fps * 6)
	ud.DialogStartFunc = -1
	ud.DialogEndFunc = -1
	ud.Field0 = 0xDEADFACE
	legacy.Nox_xxx_monsterAutoSpells_54C0C0(u)
	if u.SubClass().AsMonster().Has(object.MonsterShopkeeper) {
		idata := u.InitData
		*(*float32)(unsafe.Add(idata, 1716)) = 1.0
		*(*float32)(unsafe.Add(idata, 1720)) = 0.33333
	}
	if u.SubClass().AsMonster().Has(object.MonsterFemaleNPC) {
		for i := range ud.Color {
			ud.Color[i] = server.Color3{R: 210, G: 174, B: 121}
		}
	}
	ud.Field1 = 0
	ud.Field72 = 0
	ud.Field73 = 0
}
