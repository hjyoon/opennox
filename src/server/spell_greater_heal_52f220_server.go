package server

import (
	"math"

	"github.com/opennox/libs/player"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
)

type SpellGreaterHealRuntime52F220 struct {
	AdjustHP func(*Object, int32)
}

// SpellGreaterHealCreate52F220 binds the original create callback to native
// duration fields, target search, and the already-native ray packet path.
//
//go:noinline
func (s *Server) SpellGreaterHealCreate52F220(record *DurSpell, runtime SpellGreaterHealRuntime52F220) int32 {
	return spellGreaterHealCreate52F220(record, greaterHealCreateHooks52F220[*DurSpell, *Object]{
		loadMode:   func(record *DurSpell) uint32 { return record.Flag20 },
		loadCaster: func(record *DurSpell) *Object { return record.Caster16 },
		loadSpell:  func(record *DurSpell) uint32 { return record.Spell },
		loadAim:    func(record *DurSpell) *types.Pointf { return &record.Pos2 },
		spellFlags: func(id uint32) uint32 { return uint32(s.Spells.Flags(spell.ID(id))) },
		searchTarget: func(aim *types.Pointf, caster *Object, flags uint32, distance float32, mode int, self *Object) *Object {
			return s.Nox_xxx_spellFlySearchTarget(aim, caster, things.SpellFlags(flags), distance, mode, self)
		},
		adjustHP:    runtime.AdjustHP,
		storeTarget: func(record *DurSpell, target *Object) { record.Target48 = target },
		isEnemy:     s.IsEnemyTo,
		startRay:    s.NetStartDurationRaySpell,
		noTarget: func(caster *Object) {
			s.NetPriMsgToPlayer(caster, strman.ID("ExecDur.c:GreaterHealNoTarget"), 0)
		},
	})
}

type SpellGreaterHealRuntime52F2E0 struct {
	AdjustHP func(*Object, int32)
	ManaSub  func(*Object, int32)
}

// SpellGreaterHealUpdate52F2E0 binds GAME.EXE 0052F2E0 to native pointers
// and uses the original float32 accumulator in Field72.
//
//go:noinline
func (s *Server) SpellGreaterHealUpdate52F2E0(record *DurSpell, runtime SpellGreaterHealRuntime52F2E0) int32 {
	return spellGreaterHealUpdate52F2E0(record, greaterHealUpdateHooks52F2E0[*DurSpell, *Object]{
		loadTarget: func(record *DurSpell) *Object { return record.Target48 },
		loadFlags:  func(target *Object) uint32 { return uint32(target.ObjFlags) },
		loadCaster: func(record *DurSpell) *Object { return record.Caster16 },
		testBuff:   func(caster *Object, buff int32) int32 { return caster.UnitBuffTest4FF350(buff) },
		canInteract: func(caster, target *Object) bool {
			return s.CanInteract(caster, target, 0)
		},
		oldMana:   UnitGetOldMana4EEC80,
		loadClass: func(caster *Object) uint8 { return uint8(caster.ObjClass) },
		positionDelta: func(caster *Object, record *DurSpell) int32 {
			return s.PositionDelta4FEA70(caster, &record.Pos)
		},
		wasDamaged: func(caster *Object) bool {
			return caster.HealthData != nil && s.Frame()-caster.Frame134 <= 1
		},
		maxHP:        UnitGetMaxHP4EE7A0,
		getHP:        UnitGetHP4EE780,
		loadFraction: func(record *DurSpell) float32 { return math.Float32frombits(uint32(record.Field72)) },
		loadLevel:    func(record *DurSpell) uint32 { return record.Level },
		coefficient: func(level uint32) float32 {
			return memmap.Float32(0x587000, 260360+uintptr(4*level))
		},
		playerClass: func(caster *Object) uint8 {
			return uint8(caster.UpdateDataPlayer().Player.PlayerClass())
		},
		classHealth: func(class uint8) float32 {
			switch player.Class(class) {
			case player.Warrior:
				return s.Players.Mult.Warrior.Health
			case player.Conjurer:
				return s.Players.Mult.Conjurer.Health
			case player.Wizard:
				return s.Players.Mult.Wizard.Health
			default:
				return 1
			}
		},
		storeFraction: func(record *DurSpell, fraction float32) {
			record.Field72 = int32(math.Float32bits(fraction))
		},
		adjustHP: runtime.AdjustHP,
		manaSub:  runtime.ManaSub,
	})
}
