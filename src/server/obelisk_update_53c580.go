package server

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/player"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
)

const (
	obeliskUpdateRangeSquared53C580 = float64(2500)
	obeliskUpdateMaxMana53C580      = int32(50)
	obeliskUpdateSyncStep53C580     = int32(8)
	obeliskUpdateWandMask53C580     = object.WeaponClass(0x047f0000)
)

// ObeliskUpdateRuntime53C580 supplies the operations whose storage or packet
// format remains owned by the legacy runtime. All Object, Player, modifier,
// and use-data identities stay native-width on the Go side of the boundary.
type ObeliskUpdateRuntime53C580 struct {
	ReplenishmentEffect unsafe.Pointer
	ProtectMana         func(uint32, int16)
	ProtectPlayerHPMana func(uint32, uint16)
	ReportCharges       func(uint8, *Object, uint8, uint8)
}

type obeliskUpdateNativeDeps53C580 struct {
	playerUnits         func() []*Object
	mapTraceVision      func(*Object, *Object) bool
	frame               func() uint32
	tickRate            func() uint32
	questMode           bool
	onlineMode          bool
	balanceFloat        func(string) float32
	questManaMultiplier func(player.Class) float32
	replenishmentEffect unsafe.Pointer
	protectMana         func(uint32, int16)
	protectPlayerHPMana func(uint32, uint16)
	reportCharges       func(uint8, *Object, uint8, uint8)
	needSync            func(*Object)
	audioManaRecharge   func(*Object)
}

// obeliskRechargeItem53C520 is the native-width specialization of GAME.EXE
// 0053C520. WandUseData is fixed-width, but its owning Object pointer is not.
func obeliskRechargeItem53C520(item *Object, amount int32) bool {
	if item == nil || item.UseData.Ptr == nil {
		return false
	}
	data := item.UseData.AsWand()
	progress := int32(data.Progress)
	if progress >= 100 {
		return false
	}
	progress += amount
	if progress < 100 {
		data.Progress = uint32(progress)
	} else {
		data.Progress = 100
	}
	charge := int32(data.Progress) * int32(data.MaxCharge) / 100
	if charge == int32(data.Charge) {
		return false
	}
	data.Charge = uint8(charge)
	return true
}

// obeliskRechargeRate53C940 preserves the Oblivion Staff balance override and
// the ReplenishmentEffect lookup in modifier slots two and three.
func obeliskRechargeRate53C940(item *Object, deps obeliskUpdateNativeDeps53C580) int32 {
	if item == nil {
		return 0
	}
	if item.ObjClass.Has(object.ClassWand) &&
		item.ObjSubClass.AsWeapon().Has(object.WeaponStaffOblivionOrb) {
		return playerUnitInitFloatToInt4EFE80(deps.balanceFloat("OblivionStaffRechargeRate"))
	}
	attrs := (*ModifierInitData)(item.InitData)
	if attrs == nil {
		return 0
	}
	for i := 2; i < 4; i++ {
		effect := attrs.Modifiers[i]
		if effect != nil && effect.Attack40.Fnc == deps.replenishmentEffect {
			return effect.Attack40.Val
		}
	}
	return 0
}

func obeliskQuestManaAmount53C580(pl *Player, deps obeliskUpdateNativeDeps53C580) int16 {
	amount := int16(1)
	if pl == nil {
		return amount
	}
	switch class := pl.PlayerClass(); class {
	case player.Warrior, player.Wizard, player.Conjurer:
		amount = int16(playerUnitInitFloatToInt4EFE80(deps.questManaMultiplier(class)))
	}
	return amount
}

// obeliskUpdateNative53C580 preserves GAME.EXE 0053C580 while reading every
// widened pointer through the canonical Go layouts. An eligible nearby player
// suppresses passive regeneration even when neither the wand nor player mana
// needs charging, matching the original no_player_nearby transition.
func obeliskUpdateNative53C580(obelisk *Object, deps obeliskUpdateNativeDeps53C580) int32 {
	if obelisk == nil || obelisk.UpdateData == nil {
		return 0
	}
	data := (*ObeliskUpdateData)(obelisk.UpdateData)
	noPlayerNearby := true
	for _, unit := range deps.playerUnits() {
		if unit == nil || unit.ObjFlags.Has(object.FlagDead) {
			continue
		}
		update := (*PlayerUpdateData)(unit.UpdateData)
		if obelisk.TeamVal.Has() && !obelisk.TeamVal.SameAs(&unit.TeamVal) {
			continue
		}
		dx := float64(obelisk.PosVec.X - unit.PosVec.X)
		dy := float64(obelisk.PosVec.Y - unit.PosVec.Y)
		if dx*dx+dy*dy >= obeliskUpdateRangeSquared53C580 || !deps.mapTraceVision(obelisk, unit) {
			continue
		}

		noPlayerNearby = false
		consumed := false
		if data.Mana >= 1 {
			weapon := update.EquippedWeapon
			rechargeRate := obeliskRechargeRate53C940(weapon, deps)
			if deps.onlineMode && !deps.questMode {
				rechargeRate = 1
			}
			if rechargeRate != 0 && weapon != nil &&
				weapon.ObjClass.Has(object.ClassWand) &&
				weapon.ObjSubClass.AsWeapon().HasAny(obeliskUpdateWandMask53C580) {
				wand := weapon.UseData.AsWand()
				if int32(wand.Progress) < 100 {
					if deps.questMode {
						if deps.frame()%(deps.tickRate()>>1) == 0 && deps.audioManaRecharge != nil {
							deps.audioManaRecharge(obelisk)
						}
					} else {
						consumed = true
						data.Mana--
					}
					if obeliskRechargeItem53C520(weapon, rechargeRate) && deps.reportCharges != nil {
						deps.reportCharges(update.Player.PlayerInd, weapon, wand.Charge, wand.MaxCharge)
					}
				}
			}
		}

		if data.Mana >= 1 && update.ManaCur < update.ManaMax && data.Mana > 0 {
			amount := int16(1)
			if deps.questMode {
				amount = obeliskQuestManaAmount53C580(update.Player, deps)
			}
			update.ManaCur = uint16(int32(update.ManaCur) + int32(amount))
			if deps.protectMana != nil {
				deps.protectMana(update.Player.ProtUnitManaCur, amount)
			}
			if update.ManaCur > update.ManaMax {
				update.ManaCur = update.ManaMax
				if deps.protectPlayerHPMana != nil {
					deps.protectPlayerHPMana(update.Player.ProtUnitManaCur, update.ManaMax)
				}
			}
			if deps.questMode {
				if deps.frame()%(deps.tickRate()>>1) == 0 && deps.audioManaRecharge != nil {
					deps.audioManaRecharge(obelisk)
				}
			} else {
				data.Mana--
				consumed = true
			}
		}

		if consumed {
			if data.Mana%obeliskUpdateSyncStep53C580 == 0 && deps.needSync != nil {
				deps.needSync(obelisk)
			}
			if deps.frame()-obelisk.Field34 > deps.tickRate()>>1 {
				if deps.audioManaRecharge != nil {
					deps.audioManaRecharge(obelisk)
				}
				obelisk.Field34 = deps.frame()
			}
		}
	}

	if !noPlayerNearby {
		return 0
	}
	halfRate := deps.tickRate() >> 1
	result := int32(deps.frame() / halfRate)
	if deps.frame()%halfRate == 0 {
		result = data.Mana
		if data.Mana < obeliskUpdateMaxMana53C580 {
			if result%obeliskUpdateSyncStep53C580 == 0 && deps.needSync != nil {
				deps.needSync(obelisk)
			}
			data.Mana++
		}
	}
	return result
}

func obeliskUpdateServerDeps53C580(
	s *Server,
	runtime ObeliskUpdateRuntime53C580,
) obeliskUpdateNativeDeps53C580 {
	return obeliskUpdateNativeDeps53C580{
		playerUnits:    s.Players.ListUnits,
		mapTraceVision: s.MapTraceVision,
		frame:          s.Frame,
		tickRate:       s.TickRate,
		questMode:      noxflags.HasGame(noxflags.GameModeQuest),
		onlineMode:     noxflags.HasGame(noxflags.GameOnline),
		balanceFloat: func(name string) float32 {
			return float32(s.Balance.Float(name))
		},
		questManaMultiplier: func(class player.Class) float32 {
			stats := s.Players.ClassStatsMult(class)
			if stats == nil {
				return 1
			}
			return stats.Mana
		},
		replenishmentEffect: runtime.ReplenishmentEffect,
		protectMana:         runtime.ProtectMana,
		protectPlayerHPMana: runtime.ProtectPlayerHPMana,
		reportCharges:       runtime.ReportCharges,
		needSync: func(obj *Object) {
			obj.NeedSync()
		},
		audioManaRecharge: func(obj *Object) {
			s.Audio.EventObj(sound.SoundManaRecharge, obj, 0, 0)
		},
	}
}

// ObeliskUpdate53C580 binds the original update callback to native-width
// Object, PlayerUpdateData, Player, modifier, and wand-use-data records.
func (s *Server) ObeliskUpdate53C580(obelisk *Object, runtime ObeliskUpdateRuntime53C580) int32 {
	return obeliskUpdateNative53C580(obelisk, obeliskUpdateServerDeps53C580(s, runtime))
}
