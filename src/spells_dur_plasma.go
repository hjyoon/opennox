package opennox

import (
	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func (sp *spellsDuration) plasmaRuntime531580() server.SpellPlasmaRuntime531580 {
	world := sp.s.S()
	return server.SpellPlasmaRuntime531580{
		HecubahType:     memmap.PtrUint32(0x5D4594, 2487936),
		HecubahOrbType:  memmap.PtrUint32(0x5D4594, 2487940),
		LookupType:      func(name string) uint32 { return uint32(world.Types.IndByID(name)) },
		Frame:           sp.s.Frame,
		TickRate:        world.TickRate,
		Balance:         func(key string) float32 { return float32(sp.s.Balance.Float(key)) },
		ObjectsInCircle: world.EachChainLightningObject52F8A0,
		IsEnemy:         world.IsEnemyTo,
		CanInteract:     func(a, b *server.Object) bool { return world.CanInteract(a, b, 0) },
		Facing: func(a, b *server.Object) int {
			return legacy.Nox_server_testTwoPointsAndDirection_4E6E50(a.PosVec, int16(a.Direction1), b.PosVec)
		},
		Distance:      server.ObjectDistance4E6C00,
		PositionDelta: world.PositionDelta4FEA70,
		StartRay:      world.NetStartDurationRaySpell,
		StopRay:       world.NetStopRaySpell,
		PointFX: func(code uint8, pos types.Pointf) {
			world.Nox_xxx_netSendPointFx_522FF0(netmsg.Op(code), pos)
		},
		Damage: func(target, caster *server.Object, amount int32) {
			target.CallDamage(caster, nil, int(amount), object.DamagePlasma)
		},
		Audio: func(id uint16, target *server.Object) { sp.s.Audio.EventObj(sound.ID(id), target, 0, 0) },
		SetPlayerState: func(caster *server.Object, state server.PlayerState) {
			_ = nox_xxx_playerSetState_4FA020(caster, state)
		},
		ReportCharges: func(update *server.PlayerUpdateData, wand *server.Object, charge, max uint8) {
			legacy.Nox_xxx_netReportCharges_4D82B0(update.Player.PlayerInd, wand, charge, max)
		},
		LoadWeapon: func(record *server.DurSpell) *server.Object { return sp.plasmaWeapons[record] },
		StoreWeapon: func(record *server.DurSpell, wand *server.Object) {
			if wand == nil {
				delete(sp.plasmaWeapons, record)
				return
			}
			if sp.plasmaWeapons == nil {
				sp.plasmaWeapons = make(map[*server.DurSpell]*server.Object)
			}
			sp.plasmaWeapons[record] = wand
		},
		LoadRayTarget: func(record *server.DurSpell) *server.Object { return sp.durationRayTargets[record] },
		StoreRayTarget: func(record *server.DurSpell, target *server.Object) {
			if target == nil {
				delete(sp.durationRayTargets, record)
				return
			}
			if sp.durationRayTargets == nil {
				sp.durationRayTargets = make(map[*server.DurSpell]*server.Object)
			}
			sp.durationRayTargets[record] = target
		},
	}
}
