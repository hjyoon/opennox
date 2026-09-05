package opennox

import (
	"unsafe"

	"github.com/opennox/libs/player"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/client/noxrender"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type serverSpells struct {
	duration spellsDuration
	missiles spellMissiles
	walls    spellWall
}

func (sp *serverSpells) Init(s *Server) {
	sp.duration.Init(s)
	sp.missiles.Init(s)
	sp.walls.Init(s)
}

func (sp *serverSpells) Free() {
	sp.walls.Free()
	sp.missiles.Free()
	sp.duration.Free()
}

const phonemeLeafNativeSize = 40 + 10*(unsafe.Sizeof(uintptr(0))-4)

var _ = [1]struct{}{}[phonemeLeafNativeSize-unsafe.Sizeof(server.PhonemeLeaf{})]

func nox_xxx_spellAwardAll1_4EFD80(p *server.Player) {
	noxServer.BeastScrollAwardAll4EFD80(p, server.BeastScrollAwardAllRuntime4EFD80{
		ResetProtection: func(token uint32, value int32) {
			legacy.Nox_xxx_playerResetProtectionCRC_56F7D0(token, int(value))
		},
		AwardProtection: func(token uint32, index, level int32) {
			legacy.Nox_xxx_playerAwardSpellProtectionCRC_56FCE0(token, int(index), int(level))
		},
		ApplyProtection: func(token uint32, levels *[41]uint32, count int32) {
			legacy.Nox_xxx_playerApplyProtectionCRC_56FD50(
				token,
				unsafe.Pointer(&levels[0]),
				int(count),
			)
		},
	})
}

func nox_xxx_spellAwardAll2_4EFC80(p *server.Player) {
	noxServer.SpellAwardAll4EFC80(p, server.SpellAwardAllRuntime4EFC80{
		ResetProtection: func(token uint32, value int32) {
			legacy.Nox_xxx_playerResetProtectionCRC_56F7D0(token, int(value))
		},
		AwardProtection: func(token uint32, index, level int32) {
			legacy.Nox_xxx_playerAwardSpellProtectionCRC_56FCE0(token, int(index), int(level))
		},
		GrantSpell: func(unit *server.Object, spellID, a3, a4, a5 int32) {
			_ = legacy.Nox_xxx_spellGrantToPlayer_4FB550(
				unit,
				spell.ID(spellID),
				int(a3),
				int(a4),
				int(a5),
			)
		},
		ApplyProtection: func(token uint32, levels *[137]uint32, count int32) {
			legacy.Nox_xxx_playerApplyProtectionCRC_56FD50(
				token,
				unsafe.Pointer(&levels[0]),
				int(count),
			)
		},
	})
}

func nox_xxx_spellAwardAll3_4EFE10(p *server.Player) {
	noxServer.WarriorAbilityAwardAll4EFE10(p, server.WarriorAbilityAwardAllRuntime4EFE10{
		AwardProtection: func(token uint32, index, level int32) {
			legacy.Nox_xxx_playerAwardSpellProtectionCRC_56FCE0(token, int(index), int(level))
		},
	})
}

func (s *Server) spellGrantRuntime4FB550() server.SpellGrantRuntime4FB550 {
	return server.SpellGrantRuntime4FB550{
		AwardProtection: func(token uint32, spellID, level int32) {
			legacy.Nox_xxx_playerAwardSpellProtectionCRC_56FCE0(token, int(spellID), int(level))
		},
		SendLineMessage: legacy.Nox_xxx_netSendLineMessage_4D9EB0,
		ShopExit:        s.shopExitNative50F4C0,
	}
}

// SpellGrantToPlayer4FB550 supplies the root-owned legacy services to the
// native-width server model of GAME.EXE 004FB550.
func (s *Server) SpellGrantToPlayer4FB550(
	unit *server.Object,
	spellID, notify, shop, override int32,
) int32 {
	return s.Server.SpellGrantToPlayer4FB550(
		unit,
		spellID,
		notify,
		shop,
		override,
		s.spellGrantRuntime4FB550(),
	)
}

func nox_xxx_spellTitle_424930(ind int) (string, bool) {
	sp := noxServer.Spells.DefByInd(spell.ID(ind))
	if sp == nil || !sp.IsValid() {
		return "", false
	}
	return sp.Title, true
}

func nox_xxx_spellDescription_424A30(ind int) (string, bool) {
	sp := noxServer.Spells.DefByInd(spell.ID(ind))
	if sp == nil {
		return "", false
	}
	return sp.Desc, true
}

func nox_xxx_spellIcon_424A90(ind int) unsafe.Pointer {
	sp := noxServer.Spells.DefByInd(spell.ID(ind))
	if sp == nil {
		return nil
	}
	return unsafe.Pointer(((*noxrender.Image)(sp.Icon)).C())
}

func nox_xxx_spellIconHighlight_424AB0(ind int) unsafe.Pointer {
	sp := noxServer.Spells.DefByInd(spell.ID(ind))
	if sp == nil {
		return nil
	}
	return unsafe.Pointer(((*noxrender.Image)(sp.IconEnabled)).C())
}

func serverSetAllBeastScrolls(p *Player, enable bool) {
	lvl := 0
	if enable {
		lvl = 1
	}
	legacy.Nox_xxx_playerResetProtectionCRC_56F7D0(p.Prot4640, 0)
	for i := 1; i < len(p.BeastScrollLvl); i++ {
		p.BeastScrollLvl[i] = uint32(lvl)
		legacy.Nox_xxx_playerAwardSpellProtectionCRC_56FCE0(p.Prot4640, i, lvl)
	}
	legacy.Nox_xxx_playerApplyProtectionCRC_56FD50(p.Prot4640, unsafe.Pointer(&p.BeastScrollLvl[0]), len(p.BeastScrollLvl))
}

func serverSetAllSpells(p *Player, enable bool, max int) {
	lvl := 0
	if enable {
		lvl = max
		if max <= 0 {
			lvl = 3
		}
	}
	legacy.Nox_xxx_playerResetProtectionCRC_56F7D0(p.Prot4636, 0)
	// set max level for all possible spells
	// the engine will automatically allow only ones that have WIS_USE, CON_USE or COMMON_USE set
	for i := 1; i < len(p.SpellLvl); i++ {
		p.SpellLvl[i] = uint32(lvl)
		legacy.Nox_xxx_playerAwardSpellProtectionCRC_56FCE0(p.Prot4636, i, lvl)
	}
	if !enable && noxflags.HasGame(noxflags.GameModeQuest) {
		u := p.PlayerUnit
		// grant default spells for Quest when disabling the cheat
		switch p.PlayerClass() {
		case player.Wizard:
			legacy.Nox_xxx_spellGrantToPlayer_4FB550(u, spell.SPELL_FIREBALL, 1, 1, 1)
		case player.Conjurer:
			legacy.Nox_xxx_spellGrantToPlayer_4FB550(u, spell.SPELL_CHARM, 1, 1, 1)
			legacy.Nox_xxx_spellGrantToPlayer_4FB550(u, spell.SPELL_LESSER_HEAL, 1, 1, 1)
		}
	}
	legacy.Nox_xxx_playerApplyProtectionCRC_56FD50(p.Prot4636, unsafe.Pointer(&p.SpellLvl[0]), len(p.SpellLvl))
}

func serverSetSpell(p *Player, sp spell.ID, lvl int) {
	if sp == spell.SPELL_INVALID {
		return
	}
	legacy.Nox_xxx_playerResetProtectionCRC_56F7D0(p.Prot4636, 0)
	for i := 1; i < len(p.SpellLvl); i++ {
		cur := int(p.SpellLvl[i])
		if sp == spell.ID(i) {
			p.SpellLvl[i] = uint32(lvl)
			cur = lvl
		}
		legacy.Nox_xxx_playerAwardSpellProtectionCRC_56FCE0(p.Prot4636, i, cur)
	}
	legacy.Nox_xxx_playerApplyProtectionCRC_56FD50(p.Prot4636, unsafe.Pointer(&p.SpellLvl[0]), len(p.SpellLvl))
}

func serverSetAllWarriorAbilities(p *Player, enable bool, max int) {
	if p.PlayerClass() != player.Warrior {
		return
	}
	lvl := 0
	if enable {
		lvl = max
		if max <= 0 {
			lvl = 5
		}
	}
	for i := 1; i < 6; i++ {
		p.SpellLvl[i] = uint32(lvl)
		legacy.Nox_xxx_playerAwardSpellProtectionCRC_56FCE0(p.Prot4636, i, lvl)
	}
}

func nox_xxx_spellBookReact_4FCB70() {
	spellBookReact4FCB70(
		nox_xxx_spellCastByBook_4FCB80,
		noxServer.spells.duration.spellCastByPlayer,
	)
}

// SpellAccept4FD400 supplies the remaining root-owned cast callbacks to the
// native-width server model of GAME.EXE 004FD400. Its signed-dword result is
// deliberately not normalized to a Go bool.
func (s *Server) SpellAccept4FD400(
	spellID spell.ID,
	a2, obj3, obj4 *server.Object,
	sa *server.SpellAcceptArg,
	lvl int32,
) int32 {
	return s.Server.SpellAccept4FD400(
		spellID, a2, obj3, obj4, sa, lvl,
		server.SpellAcceptRuntime4FD400{
			CaptureMagic: func(id spell.ID, obj *server.Object) int32 {
				if s.gameCaptureMagic4FDC10(id, obj) {
					return 1
				}
				return 0
			},
			Audio: func(id sound.ID, obj *server.Object) {
				s.Audio.EventObj(id, obj, 0, 0)
			},
			Instant:    s.spellAcceptInstant4FD400,
			Duration:   s.spellAcceptDuration4FD400,
			PlasmaTime: func() float64 { return s.Balance.Float("PlasmaSearchTime") },
		},
	)
}

// Nox_xxx_spellAccept4FD400 is the compatibility surface used by migrated Go
// callers. The legacy C export calls SpellAccept4FD400 directly so that the
// original signed-dword return value is preserved across the ABI.
func (s *Server) Nox_xxx_spellAccept4FD400(
	spellID spell.ID,
	a2, obj3, obj4 *server.Object,
	sa *server.SpellAcceptArg,
	lvl int,
) bool {
	return s.SpellAccept4FD400(spellID, a2, obj3, obj4, sa, int32(lvl)) != 0
}

func (s *Server) spellAcceptInstant4FD400(
	spellID spell.ID,
	a2, obj3, obj4 *server.Object,
	sa *server.SpellAcceptArg,
	lvl int32,
) int32 {
	var fnc func(spell.ID, *server.Object, *server.Object, *server.Object, *server.SpellAcceptArg, int) int
	switch spellID {
	case spell.SPELL_ANCHOR:
		fnc = castAnchor
	case spell.SPELL_ARACHNAPHOBIA:
		fnc = legacy.Nox_xxx_spellArachna_52DC80
	case spell.SPELL_BLIND:
		fnc = castBlind
	case spell.SPELL_BURN:
		fnc = legacy.Nox_xxx_castBurn_52C3E0
	case spell.SPELL_CLEANSING_FLAME, spell.SPELL_CLEANSING_MANA_FLAME:
		fnc = legacy.Nox_xxx_spellCastCleansingFlame_52D5C0
	case spell.SPELL_CONFUSE:
		fnc = legacy.Nox_xxx_castConfuse_52C1E0
	case spell.SPELL_COUNTERSPELL:
		fnc = nox_xxx_castCounterSpell_52BBB0
	case spell.SPELL_CURE_POISON:
		fnc = legacy.Nox_xxx_castCurePoison_52CDB0
	case spell.SPELL_DEATH:
		fnc = castDeath
	case spell.SPELL_DEATH_RAY:
		fnc = castDeathRay
	case spell.SPELL_DETECT_MAGIC:
		fnc = castDetectMagic
	case spell.SPELL_DETONATE_GLYPHS:
		fnc = castDetonateGlyphs
	case spell.SPELL_EARTHQUAKE:
		fnc = legacy.Nox_xxx_castEquake_52DE40
	case spell.SPELL_FEAR:
		fnc = castFear
	case spell.SPELL_FIREBALL:
		fnc = legacy.Nox_xxx_castFireball_52C790
	case spell.SPELL_FIST:
		fnc = legacy.Nox_xxx_castFist_52D3C0
	case spell.SPELL_FREEZE:
		fnc = castFreeze
	case spell.SPELL_FUMBLE:
		fnc = legacy.Nox_xxx_castFumble_52C060
	case spell.SPELL_GLYPH:
		fnc = castGlyph
	case spell.SPELL_HASTE:
		fnc = castHaste
	case spell.SPELL_INFRAVISION:
		fnc = castInfravision
	case spell.SPELL_INVERSION:
		fnc = legacy.Sub_52BEB0
	case spell.SPELL_INVISIBILITY:
		fnc = castInvisibility
	case spell.SPELL_INVULNERABILITY:
		fnc = castInvulnerability
	case spell.SPELL_LESSER_HEAL:
		fnc = legacy.Sub_52DD50
	case spell.SPELL_LIGHT:
		fnc = castLight
	case spell.SPELL_LOCK:
		fnc = legacy.Nox_xxx_castLock_52CE90
	case spell.SPELL_MARK:
		fnc = legacy.Sub_52CA80
	case spell.SPELL_MARK_1, spell.SPELL_MARK_2, spell.SPELL_MARK_3, spell.SPELL_MARK_4:
		fnc = legacy.Sub_52CBD0
	case spell.SPELL_MAGIC_MISSILE:
		fnc = s.spells.missiles.Cast
	case spell.SPELL_METEOR:
		fnc = legacy.Nox_xxx_castMeteor_52D9D0
	case spell.SPELL_METEOR_SHOWER:
		fnc = legacy.Nox_xxx_castMeteorShower_52D8A0
	case spell.SPELL_NULLIFY:
		fnc = castNullify
	case spell.SPELL_PIXIE_SWARM:
		fnc = legacy.Nox_xxx_castPixies_540440
	case spell.SPELL_POISON:
		fnc = legacy.Nox_xxx_castPoison_52C720
	case spell.SPELL_PROTECTION_FROM_ELECTRICITY:
		fnc = castProtectElectricity
	case spell.SPELL_PROTECTION_FROM_FIRE:
		fnc = castProtectFire
	case spell.SPELL_PROTECTION_FROM_POISON:
		fnc = castProtectPoison
	case spell.SPELL_PULL:
		fnc = legacy.Nox_xxx_castPull_52BFA0
	case spell.SPELL_PUSH:
		fnc = legacy.Nox_xxx_castPush_52C000
	case spell.SPELL_RESTORE_HEALTH, spell.SPELL_WINK:
		fnc = legacy.Nox_xxx_castSpellWinkORrestoreHealth_52BF20
	case spell.SPELL_RESTORE_MANA:
		fnc = legacy.Sub_52BF50
	case spell.SPELL_RUN:
		fnc = castRun
	case spell.SPELL_SHOCK:
		fnc = legacy.Nox_xxx_useShock_52C5A0
	case spell.SPELL_SLOW:
		fnc = castSlow
	case spell.SPELL_STUN:
		fnc = legacy.Nox_xxx_castStun_52C2C0
	case spell.SPELL_TELEKINESIS:
		fnc = legacy.Nox_xxx_castTelekinesis_52D330
	case spell.SPELL_TOXIC_CLOUD:
		fnc = legacy.Nox_xxx_castToxicCloud_52DB60
	case spell.SPELL_TRIGGER_GLYPH:
		fnc = legacy.Sub_52CCD0
	case spell.SPELL_VAMPIRISM:
		fnc = castVampirism
	case spell.SPELL_VILLAIN:
		fnc = castVillain
	case spell.ID(6), spell.ID(18), spell.ID(57), spell.ID(63):
		// GAME.EXE 0052BBA0, 0052BF00, 0052CA70 and 0052D190 are
		// one-instruction success callbacks.
		return 1
	default:
		// The decoded selector never routes another ID here. Keep the
		// original default-success behavior if the table is extended.
		return 1
	}
	return int32(fnc(spellID, a2, obj3, obj4, sa, int(lvl)))
}

func spellAcceptBool4FD400(ok bool) int32 {
	if ok {
		return 1
	}
	return 0
}

func (s *Server) spellAcceptDuration4FD400(
	spellID spell.ID,
	a2, obj3, obj4 *server.Object,
	sa *server.SpellAcceptArg,
	lvl int32,
	timeout uint32,
) int32 {
	level := int(lvl)
	var create, update, destroy unsafe.Pointer
	switch spellID {
	case spell.SPELL_BLINK:
		create, update = legacy.Get_nox_xxx_spellBlink2_530310(), legacy.Get_nox_xxx_spellBlink1_530380()
	case spell.SPELL_CHANNEL_LIFE:
		update = legacy.Get_sub_52F460()
	case spell.SPELL_CHARM:
		create, update, destroy = legacy.Get_nox_xxx_charmCreature1_5011F0(), legacy.Get_nox_xxx_charmCreatureFinish_5013E0(), legacy.Get_nox_xxx_charmCreature2_501690()
	case spell.SPELL_TURN_UNDEAD:
		create, update, destroy = legacy.Get_nox_xxx_spellTurnUndeadCreate_531310(), legacy.Get_nox_xxx_spellTurnUndeadUpdate_531410(), legacy.Get_nox_xxx_spellTurnUndeadDelete_531420()
	case spell.SPELL_DRAIN_MANA:
		update = legacy.Get_nox_xxx_spellDrainMana_52E210()
	case spell.SPELL_LIGHTNING:
		create, update, destroy = legacy.Get_nox_xxx_spellEnergyBoltStop_52E820(), legacy.Get_nox_xxx_spellEnergyBoltTick_52E850(), legacy.Get_nullsub_29()
	case spell.SPELL_FIREWALK:
		update = legacy.Get_nox_xxx_firewalkTick_52ED40()
	case spell.SPELL_FORCE_OF_NATURE:
		create, update, destroy = legacy.Get_sub_52EF30(), legacy.Get_sub_52EFD0(), legacy.Get_sub_52F1D0()
	case spell.SPELL_GREATER_HEAL:
		create, update = legacy.Get_sub_52F220(), legacy.Get_sub_52F2E0()
	case spell.SPELL_CHAIN_LIGHTNING:
		create, update, destroy = legacy.Get_nox_xxx_onStartLightning_52F820(), legacy.Get_nox_xxx_onFrameLightning_52F8A0(), legacy.Get_sub_530100()
	case spell.SPELL_SHIELD:
		create, update, destroy = legacy.Get_nox_xxx_castShield1_52F5A0(), legacy.Get_sub_52F650(), legacy.Get_sub_52F670()
	case spell.SPELL_MOONGLOW:
		create, destroy = legacy.Get_nox_xxx_spellCreateMoonglow_531A00(), legacy.Get_sub_531AF0()
	case spell.SPELL_MANA_BOMB:
		create, update, destroy = legacy.Get_nox_xxx_manaBomb_530F90(), legacy.Get_nox_xxx_manaBombBoom_5310C0(), legacy.Get_sub_531290()
	case spell.SPELL_PLASMA:
		create, update, destroy = legacy.Get_nox_xxx_plasmaSmth_531580(), legacy.Get_nox_xxx_plasmaShot_531600(), legacy.Get_sub_5319E0()
	case spell.SPELL_OVAL_SHIELD:
		create, update, destroy = legacy.Get_sub_531490(), legacy.Get_sub_5314F0(), legacy.Get_sub_531560()
	case spell.SPELL_SUMMON_BAT,
		spell.SPELL_SUMMON_BLACK_BEAR,
		spell.SPELL_SUMMON_BEAR,
		spell.SPELL_SUMMON_BEHOLDER,
		spell.SPELL_SUMMON_CARNIVOROUS_PLANT,
		spell.SPELL_SUMMON_ALBINO_SPIDER,
		spell.SPELL_SUMMON_SMALL_ALBINO_SPIDER,
		spell.SPELL_SUMMON_EVIL_CHERUB,
		spell.SPELL_SUMMON_EMBER_DEMON,
		spell.SPELL_SUMMON_GHOST,
		spell.SPELL_SUMMON_GIANT_LEECH,
		spell.SPELL_SUMMON_IMP,
		spell.SPELL_SUMMON_MECHANICAL_FLYER,
		spell.SPELL_SUMMON_MECHANICAL_GOLEM,
		spell.SPELL_SUMMON_MIMIC,
		spell.SPELL_SUMMON_OGRE,
		spell.SPELL_SUMMON_OGRE_BRUTE,
		spell.SPELL_SUMMON_OGRE_WARLORD,
		spell.SPELL_SUMMON_SCORPION,
		spell.SPELL_SUMMON_SHADE,
		spell.SPELL_SUMMON_SKELETON,
		spell.SPELL_SUMMON_SKELETON_LORD,
		spell.SPELL_SUMMON_SPIDER,
		spell.SPELL_SUMMON_SMALL_SPIDER,
		spell.SPELL_SUMMON_SPITTING_SPIDER,
		spell.SPELL_SUMMON_STONE_GOLEM,
		spell.SPELL_SUMMON_TROLL,
		spell.SPELL_SUMMON_URCHIN,
		spell.SPELL_SUMMON_WASP,
		spell.SPELL_SUMMON_WILLOWISP,
		spell.SPELL_SUMMON_WOLF,
		spell.SPELL_SUMMON_BLACK_WOLF,
		spell.SPELL_SUMMON_WHITE_WOLF,
		spell.SPELL_SUMMON_ZOMBIE,
		spell.SPELL_SUMMON_VILE_ZOMBIE,
		spell.SPELL_SUMMON_DEMON,
		spell.SPELL_SUMMON_LICH,
		spell.SPELL_SUMMON_DRYAD,
		spell.SPELL_SUMMON_URCHIN_SHAMAN:
		create, update, destroy = legacy.Get_nox_xxx_summonStart_500DA0(), legacy.Get_nox_xxx_summonFinish_5010D0(), legacy.Get_nox_xxx_summonCancel_5011C0()
	case spell.SPELL_SWAP:
		create, update = legacy.Get_sub_530CA0(), legacy.Get_sub_530D30()
	case spell.SPELL_TAG:
		create, update, destroy = legacy.Get_nox_xxx_spellTagCreature_530160(), legacy.Get_sub_530250(), legacy.Get_sub_530270()
	case spell.SPELL_TELEPORT_OTHER_TO_MARK_1, spell.SPELL_TELEPORT_OTHER_TO_MARK_2, spell.SPELL_TELEPORT_OTHER_TO_MARK_3, spell.SPELL_TELEPORT_OTHER_TO_MARK_4,
		spell.SPELL_TELEPORT_TO_MARK_1, spell.SPELL_TELEPORT_TO_MARK_2, spell.SPELL_TELEPORT_TO_MARK_3, spell.SPELL_TELEPORT_TO_MARK_4:
		create, update = legacy.Get_sub_5305D0(), legacy.Get_sub_530650()
	case spell.SPELL_TELEPORT_POP:
		create, update = legacy.Get_nox_xxx_castTele_530820(), legacy.Get_sub_530880()
	case spell.SPELL_TELEPORT_TO_TARGET:
		create, update = legacy.Get_sub_530A30_spell_execdur(), legacy.Get_nox_xxx_castTTT_530B70()
	case spell.SPELL_WALL:
		create, update, destroy = legacy.Get_nox_xxx_spellWallCreate_4FFA90(), legacy.Get_nox_xxx_spellWallUpdate_500070(), legacy.Get_nox_xxx_spellWallDestroy_500080()
	default:
		return 1
	}
	return spellAcceptBool4FD400(s.spells.duration.New(spellID, a2, obj3, obj4, sa, level, create, update, destroy, timeout))
}

func nox_xxx_castSpellByUser_4FDD20(spellID int32, caster *server.Object, arg *server.SpellAcceptArg) int32 {
	return noxServer.CastSpellByUser4FDD20(spell.ID(spellID), caster, arg)
}

// CastSpellByUser4FDD20 supplies the remaining root-owned callbacks to the
// native-width server model of GAME.EXE 004FDD20. Power and acceptance results
// are narrowed to and returned as the original signed dwords, respectively.
func (s *Server) CastSpellByUser4FDD20(
	spellID spell.ID,
	caster *server.Object,
	arg *server.SpellAcceptArg,
) int32 {
	return s.Server.CastSpellByUser4FDD20(spellID, caster, arg, server.CastSpellByUserRuntime4FDD20{
		SpellGetPower: func(id spell.ID, obj *server.Object) int32 {
			return int32(legacy.Nox_xxx_spellGetPower_4FE7B0(id, obj))
		},
		DisableEnchant: func(obj *server.Object, enchant server.EnchantID) {
			asObjectS(obj).DisableEnchant(enchant)
		},
		CancelDuration: func(id spell.ID, obj *server.Object) {
			s.Spells.Dur.CancelFor(id, obj)
		},
		CreateProjectile: func(source, target *server.Object, id spell.ID) {
			s.CreateSpellProjectile4FDDA0(source, target, id)
		},
		SpellAccept: s.SpellAccept4FD400,
	})
}

// CreateSpellProjectile4FDDA0 supplies the remaining root-owned callbacks to
// the native-width server model of GAME.EXE 004FDDA0.
func (s *Server) CreateSpellProjectile4FDDA0(
	source, target *server.Object,
	spellID spell.ID,
) *server.Object {
	return s.Server.CreateSpellProjectile4FDDA0(
		source, target, spellID,
		server.CreateSpellProjectileRuntime4FDDA0{
			SpellGetPower: func(id spell.ID, object *server.Object) int32 {
				return int32(legacy.Nox_xxx_spellGetPower_4FE7B0(id, object))
			},
			CreateAt: func(object, owner *server.Object, position types.Pointf, _ int32) {
				s.CreateObjectAt(object, owner, position)
			},
			ApplyEnchant: func(object *server.Object, enchant server.EnchantID, duration int16, power uint8) {
				legacy.Nox_xxx_buffApplyTo_4FF380(object, enchant, int(duration), int(power))
			},
		},
	)
}

// CollisionEnchant4FDF90 supplies the remaining root-owned enchant removal
// callback to the native-width server model of GAME.EXE 004FDF90.
func (s *Server) CollisionEnchant4FDF90(source, target *server.Object) {
	s.Server.CollisionEnchant4FDF90(source, target, server.CollisionEnchantRuntime4FDF90{
		DisableEnchant: func(obj *server.Object, enchant server.EnchantID) {
			asObjectS(obj).DisableEnchant(enchant)
		},
	})
}

// RandomSpell4FE060 routes the exact fixed-width GAME.EXE 004FE060 ABI to
// the native server spell registry and logic RNG.
func (s *Server) RandomSpell4FE060(firstMask, secondMask uint32) int32 {
	return s.Server.RandomSpell4FE060(firstMask, secondMask)
}

// castSpellAtLevel is the compatibility route for script APIs that explicitly
// supply a level. GAME.EXE 004FDD20 itself never accepts such an argument and
// is exposed separately through CastSpellByUser4FDD20 above.
func (s *Server) castSpellAtLevel(spellInd spell.ID, lvl int, u *server.Object, a3 *server.SpellAcceptArg) bool {
	if s.Spells.HasFlags(spellInd, things.SpellOffensive) {
		asObjectS(u).DisableEnchant(server.ENCHANT_INVISIBLE)
		asObjectS(u).DisableEnchant(server.ENCHANT_INVULNERABLE)
		s.Spells.Dur.CancelFor(spell.SPELL_OVAL_SHIELD, u)
	}
	if !s.Spells.HasFlags(spellInd, things.SpellTargeted) || u == a3.Obj {
		return s.Nox_xxx_spellAccept4FD400(spellInd, u, u, u, a3, lvl)
	}
	s.CreateSpellProjectile4FDDA0(u, a3.Obj, spellInd)
	return true
}

func (s *Server) castSpellBy(spellInd spell.ID, lvl int, caster *server.Object, targ server.Obj, targPos types.Pointf) bool {
	sa, freeArg := alloc.New(server.SpellAcceptArg{})
	defer freeArg()
	sa.Obj = server.ToObject(targ)
	sa.Pos = targPos
	return s.castSpellByUserAtLevel(spellInd, lvl, caster, sa)
}

func (s *Server) castSpellByUserAtLevel(spellInd spell.ID, lvl int, u *server.Object, sa *server.SpellAcceptArg) bool {
	if lvl < 0 {
		lvl = legacy.Nox_xxx_spellGetPower_4FE7B0(spellInd, u)
	}
	return s.castSpellAtLevel(spellInd, lvl, u, sa)
}
