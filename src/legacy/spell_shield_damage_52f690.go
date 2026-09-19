package legacy

import (
	"encoding/binary"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

func spellShieldFXNative523670(s *server.Server, target, source *server.Object) {
	if s == nil || target == nil {
		return
	}
	var packet [4]byte
	packet[0] = 0x80
	binary.LittleEndian.PutUint16(packet[1:3], uint16(s.GetUnitNetCode(target)))
	direction := int(target.Direction1)
	if source != nil {
		direction = Nox_xxx_math_509ED0(target.PosVec.Sub(source.PosVec))
	}
	packet[3] = byte(Nox_xxx_math_509EA0(direction))
	s.Nox_xxx_netSendFxAllCli_523030(target.PosVec, packet[:])
}

func spellShieldDamageNative52F690(s *server.Server, target *server.Object, amount int32) {
	server.SpellShieldDamage52F690(
		s.Spells.Dur.SpellDurationFirst4FE930(), target, amount,
		server.SpellShieldDamageRuntime52F690{
			Cancel: func(record *server.DurSpell) {
				_ = s.Spells.Dur.SpellDurationCancel4FE9D0(record)
			},
			BuffOff: Nox_xxx_spellBuffOff_4FF5B0,
		},
	)
}

func spellShieldReduceDamageNative52F710(
	s *server.Server,
	target *server.Object,
	damage *int32,
	typ object.DamageType,
	source *server.Object,
) {
	server.SpellShieldReduceDamage52F710(
		target, damage, typ, source,
		server.SpellShieldReduceDamageRuntime52F710{
			Audio: func(id int, object *server.Object) {
				s.Audio.EventObj(sound.ID(id), object, 0, 0)
			},
			ShieldFX:  func(target, source *server.Object) { spellShieldFXNative523670(s, target, source) },
			CurrentHP: server.UnitGetHP4EE780,
			SetHP:     Nox_xxx_unitSetHP_4E4560,
			ShieldDamage: func(target *server.Object, amount int32) {
				spellShieldDamageNative52F690(s, target, amount)
			},
			FrontBlock: func(target, source *server.Object) bool {
				return Nox_server_testTwoPointsAndDirection_4E6E50(
					target.PosVec, int16(target.Direction1), source.PrevPos,
				)&1 != 0
			},
			Frame: s.Frame,
		},
	)
}
