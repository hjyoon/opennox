package opennox

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func (s *Server) wandShot53F480(
	owner *server.Object,
	projectileType uint32,
	position types.Pointf,
	direction server.Dir16,
) *server.Object {
	projectile := s.NewObjectByTypeInd(int(projectileType))
	if projectile == nil {
		return nil
	}
	s.CreateObjectAt(projectile, owner, position)
	projectile.SetDir(direction)
	projectile.VelVec = owner.VelVec.Add(direction.Vec().Mul(projectile.SpeedCur))
	return projectile
}

func wandUseConsumeCharge53F290(owner, wand *server.Object, data *server.WandUseData) {
	if data.MaxCharge == 0 {
		return
	}
	data.Charge--
	data.Progress = uint32(data.Charge) * 100 / uint32(data.MaxCharge)
	if owner.Class().Has(object.ClassPlayer) {
		player := owner.UpdateDataPlayer().Player
		legacy.Nox_xxx_netReportCharges_4D82B0(player.PlayerInd, wand, data.Charge, data.MaxCharge)
	}
}

// nox_xxx_useLesserFireballStaff_53F290 restores WandUse without converting
// the owner, wand, UseData, or newly-created projectile pointers to ABI32.
func nox_xxx_useLesserFireballStaff_53F290(owner, wand *server.Object) bool {
	data := wand.UseData.AsWand()
	if data == nil {
		return true
	}
	s := noxServer
	if data.ProjectileType == 0 {
		data.ProjectileType = uint32(s.Types.IndByID(data.Projectile()))
	}
	if data.MaxCharge != 0 && data.Charge == 0 {
		s.Audio.EventObj(sound.SoundDepletedWand, owner, 0, 0)
		return false
	}
	readiness := legacy.Nox_xxx_itemCheckReadinessEffect_4E0960(wand)
	if s.Frame()-data.LastUsed < data.Cooldown-uint32(readiness) {
		return false
	}

	direction := owner.Direction1
	position := owner.PosVec.Add(direction.Vec().Mul(owner.Shape.Circle.R + 4)).Add(owner.VelVec)
	if !s.MapTraceRayAt(owner.PosVec, position, nil, nil, 5) {
		position = owner.PosVec
	}
	s.wandShot53F480(owner, data.ProjectileType, position, direction)
	if data.Flags&1 != 0 {
		s.wandShot53F480(owner, data.ProjectileType, position, server.RoundDir(int(direction)+8))
		s.wandShot53F480(owner, data.ProjectileType, position, server.RoundDir(int(direction)-8))
	}
	s.Audio.EventObj(sound.ID(data.Sound), owner, 0, 0)
	wandUseConsumeCharge53F290(owner, wand, data)
	data.LastUsed = s.Frame()
	return true
}

// nox_xxx_useWandCastSpell_53F4F0 restores WandCastUse with a native
// SpellAcceptArg. In particular, CursorObj remains a full-width Object pointer.
func nox_xxx_useWandCastSpell_53F4F0(owner, wand *server.Object) bool {
	data := wand.UseData.AsWand()
	s := noxServer
	if data.MaxCharge != 0 && data.Charge == 0 {
		s.Audio.EventObj(sound.SoundDepletedWand, owner, 0, 0)
		return false
	}
	readiness := legacy.Nox_xxx_itemCheckReadinessEffect_4E0960(wand)
	if s.Frame()-data.LastUsed < data.Cooldown-uint32(2*readiness) {
		return false
	}

	arg := server.SpellAcceptArg{Obj: owner, Pos: owner.PosVec}
	if owner.Class().Has(object.ClassPlayer) {
		update := owner.UpdateDataPlayer()
		arg.Pos = types.Point2f(update.Player.CursorVec)
		if update.CursorObj != nil {
			arg.Obj = update.CursorObj
		}
	}
	data.Flags |= 4
	if s.Nox_xxx_spellAccept4FD400(spell.ID(data.Spell), owner, owner, owner, &arg, 4) {
		data.LastUsed = s.Frame()
		if !wand.Class().Has(object.ClassWand) || uint32(wand.SubClass())&0x04040000 == 0 {
			wandUseConsumeCharge53F290(owner, wand, data)
		}
	}
	return true
}

// nox_xxx_useFireWand_53F670 restores FireWandUse. Its particle bridge only
// carries primitive values; neither object pointer enters the legacy body.
func nox_xxx_useFireWand_53F670(owner, wand *server.Object) bool {
	s := noxServer
	direction := owner.Direction1.Vec()
	position := owner.PosVec.Add(direction.Mul(owner.Shape.Circle.R * 1.5))
	speed := float32(s.Rand.Logic.FloatClamp(12, 25))
	velocity := direction.Mul(speed)
	velocity.X += float32(s.Rand.Logic.FloatClamp(-2, 2))
	velocity.Y += float32(s.Rand.Logic.FloatClamp(-2, 2))
	legacy.Nox_xxx_createSpark_54FD80(position.X, position.Y, 1, 20, velocity.X, velocity.Y, 0, 0)
	if s.Frame()-wand.Field34 > s.TickRate() {
		s.Audio.EventObj(sound.ID(9), owner, 0, 0)
		wand.Field34 = s.Frame()
	}
	return false
}
