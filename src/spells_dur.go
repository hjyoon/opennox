package opennox

import (
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

type spellsDuration struct {
	s *Server
	*server.SpellsDuration
	moonglowVisuals       map[*server.DurSpell]*server.Object
	forceOfNatureCharges  map[*server.DurSpell]*server.Object
	manaBombCharges       map[*server.DurSpell]*server.Object
	chainLightningWeapons map[*server.DurSpell]*server.Object
	durationRayTargets    map[*server.DurSpell]*server.Object
	// A projectile may collide and disappear before the next E2E poll.
	forceOfNatureLaunches uint64
}

func (sp *spellsDuration) Init(s *Server) {
	sp.s = s
	sp.SpellsDuration = &s.Server.Spells.Dur
}

func (sp *spellsDuration) Free() {
	sp.moonglowVisuals = nil
	sp.forceOfNatureCharges = nil
	sp.manaBombCharges = nil
	sp.chainLightningWeapons = nil
	sp.durationRayTargets = nil
	sp.forceOfNatureLaunches = 0
}

func (sp *spellsDuration) destroyDurSpell(spl *server.DurSpell) {
	sp.SpellsDuration.SpellDurationDestroy4FEDA0(spl, server.SpellDurationDestroyRuntime4FEDA0{
		CallDestroy: sp.callDestroy4FEDA0,
		SetPlayerState: func(unit *server.Object, state server.PlayerState) {
			_ = nox_xxx_playerSetState_4FA020(unit, state)
		},
	})
	delete(sp.durationRayTargets, spl)
}

func (sp *spellsDuration) callDestroy4FEDA0(callback unsafe.Pointer, record *server.DurSpell) {
	if callback == legacy.Get_nox_xxx_charmCreature2_501690() {
		_ = legacy.CharmCancelNative501690(record)
		return
	}
	if callback == legacy.Get_sub_52F670() {
		legacy.SpellShieldDestroyNative52F670(record)
		return
	}
	if callback == legacy.Get_nox_xxx_spellTurnUndeadDelete_531420() {
		server.SpellTurnUndeadDestroy531420(record, sp.turnUndeadRuntime531310())
		return
	}
	if callback == legacy.Get_sub_530270() {
		sp.s.S().SpellTagDestroy530270(record)
		return
	}
	if callback == legacy.Get_sub_531560() {
		sp.s.S().SpellOvalShieldDestroy531560(record, ovalShieldRuntime531490())
		return
	}
	if callback == legacy.Get_sub_531AF0() {
		server.SpellMoonglowDestroy531AF0(record, sp.moonglowRuntime531A00())
		// The PE32 field dies with the duration record even when its target
		// vanished first. Drop the Go sidecar entry in that case as well.
		delete(sp.moonglowVisuals, record)
		return
	}
	if callback == legacy.Get_sub_52F1D0() {
		server.SpellForceOfNatureDestroy52F1D0(record, sp.forceOfNatureRuntime52EF30())
		return
	}
	if callback == legacy.Get_sub_531290() {
		server.SpellManaBombDestroy531290(record, sp.manaBombRuntime530F90())
		return
	}
	if callback == legacy.Get_sub_530100() {
		server.SpellChainLightningDestroy530100(record, sp.chainLightningRuntime52F820())
		return
	}
	traceCDurationCall("destroy", callback, record)
	ccall.CallVoidPtr(callback, record.C())
}

func (sp *spellsDuration) process4FEEF0() {
	sp.SpellsDuration.SpellDurationProcess4FEEF0(server.SpellDurationProcessRuntime4FEEF0{
		Destroy:    sp.destroyDurSpell,
		CallUpdate: sp.callUpdate4FEEF0,
	})
}

func (sp *spellsDuration) callUpdate4FEEF0(callback unsafe.Pointer, record *server.DurSpell) int32 {
	if callback == legacy.Get_nox_xxx_charmCreatureFinish_5013E0() {
		return legacy.CharmFinishNative5013E0(record)
	}
	if callback == legacy.Get_sub_52F650() {
		return legacy.SpellShieldUpdateNative52F650(record)
	}
	if callback == legacy.Get_nox_xxx_spellBlink1_530380() {
		return server.SpellBlinkUpdate530380(record, sp.blinkRuntime530310())
	}
	if callback == legacy.Get_sub_530D30() {
		return server.SpellSwapUpdate530D30(record, sp.swapRuntime530CA0())
	}
	if callback == legacy.Get_nox_xxx_castTTT_530B70() {
		return server.SpellTeleportToTargetUpdate530B70(record, sp.teleportToTargetRuntime530A30())
	}
	if callback == legacy.Get_sub_530880() {
		return server.SpellTeleportPopUpdate530880(record, sp.teleportPopRuntime530820())
	}
	if callback == legacy.Get_nox_xxx_spellTurnUndeadUpdate_531410() {
		return server.SpellTurnUndeadUpdate531410(record)
	}
	if callback == legacy.Get_sub_530250() {
		return server.SpellTagUpdate530250(record)
	}
	if callback == legacy.Get_sub_5314F0() {
		return sp.s.S().SpellOvalShieldUpdate5314F0(record)
	}
	if callback == legacy.Get_sub_52F460() {
		return sp.s.S().SpellChannelLifeUpdate52F460(record, server.SpellChannelLifeRuntime52F460{
			AddMana: func(target *server.Object, amount int16) {
				legacy.Nox_xxx_playerManaAdd_4EEB80(target, int(amount))
			},
			ClearDamage: func(target *server.Object, amount int32) {
				legacy.Nox_xxx_unitDamageClear_4EE5E0(target, int(amount))
			},
			Coefficient: func(index uint32) float64 {
				return sp.s.Balance.FloatInd("ChannelLifeCoeff", int(int32(index)))
			},
		})
	}
	if callback == legacy.Get_nox_xxx_firewalkTick_52ED40() {
		return sp.s.S().SpellFirewalkUpdate52ED40(record, server.SpellFirewalkRuntime52ED40{
			SpawnFlame: func(kind int, position types.Pointf) {
				flame := sp.s.S().NewObjectByTypeID([3]string{"SmallFlame", "MediumFlame", "Flame"}[kind])
				if flame == nil {
					return
				}
				sp.s.CreateObjectAt(flame, nil, position)
				sp.s.Audio.EventPos(sound.ID(46), position, 0, 0)
				sp.s.S().DecaySetTime511660(flame, 25*sp.s.S().TickRate())
			},
		})
	}
	if callback == legacy.Get_sub_52F2E0() {
		return sp.s.S().SpellGreaterHealUpdate52F2E0(record, server.SpellGreaterHealRuntime52F2E0{
			AdjustHP: func(target *server.Object, amount int32) {
				legacy.Nox_xxx_unitAdjustHP_4EE460(target, int(amount))
			},
			ManaSub: func(caster *server.Object, amount int32) {
				legacy.Nox_xxx_playerManaSub_4EEBF0(caster, int(amount))
			},
		})
	}
	if callback == legacy.Get_sub_52EFD0() {
		return server.SpellForceOfNatureUpdate52EFD0(record, sp.forceOfNatureRuntime52EF30())
	}
	if callback == legacy.Get_nox_xxx_manaBombBoom_5310C0() {
		return server.SpellManaBombUpdate5310C0(record, sp.manaBombRuntime530F90())
	}
	if callback == legacy.Get_nox_xxx_onFrameLightning_52F8A0() {
		return server.SpellChainLightningUpdate52F8A0(record, sp.chainLightningRuntime52F820())
	}
	if callback == legacy.Get_nox_xxx_spellEnergyBoltTick_52E850() {
		return server.SpellEnergyBoltUpdate52E850(record, sp.energyBoltRuntime52E820())
	}
	if callback == legacy.Get_nox_xxx_spellDrainMana_52E210() {
		return server.SpellDrainManaUpdate52E210(record, sp.drainManaRuntime52E210())
	}
	traceCDurationCall("update", callback, record)
	return int32(ccall.CallIntPtr(callback, record.C()))
}

func (sp *spellsDuration) callCreate4FEBA0(callback unsafe.Pointer, record *server.DurSpell) int32 {
	if callback == legacy.Get_nox_xxx_charmCreature1_5011F0() {
		return legacy.CharmStartNative5011F0(record)
	}
	if callback == legacy.Get_nox_xxx_castShield1_52F5A0() {
		return legacy.SpellShieldCreateNative52F5A0(record)
	}
	if callback == legacy.Get_nox_xxx_spellBlink2_530310() {
		return server.SpellBlinkCreate530310(record, sp.blinkRuntime530310())
	}
	if callback == legacy.Get_sub_530CA0() {
		return server.SpellSwapCreate530CA0(record, sp.swapRuntime530CA0())
	}
	if callback == legacy.Get_sub_530A30_spell_execdur() {
		return server.SpellTeleportToTargetCreate530A30(record, sp.teleportToTargetRuntime530A30())
	}
	if callback == legacy.Get_nox_xxx_castTele_530820() {
		return server.SpellTeleportPopCreate530820(record, sp.teleportPopRuntime530820())
	}
	if callback == legacy.Get_nox_xxx_spellTurnUndeadCreate_531310() {
		return server.SpellTurnUndeadCreate531310(record, sp.turnUndeadRuntime531310())
	}
	if callback == legacy.Get_sub_52F220() {
		return sp.s.S().SpellGreaterHealCreate52F220(record, server.SpellGreaterHealRuntime52F220{
			AdjustHP: func(target *server.Object, amount int32) {
				legacy.Nox_xxx_unitAdjustHP_4EE460(target, int(amount))
			},
		})
	}
	if callback == legacy.Get_sub_52EF30() {
		return server.SpellForceOfNatureCreate52EF30(record, sp.forceOfNatureRuntime52EF30())
	}
	if callback == legacy.Get_nox_xxx_spellTagCreature_530160() {
		return sp.s.S().SpellTagCreate530160(record)
	}
	if callback == legacy.Get_sub_531490() {
		return sp.s.S().SpellOvalShieldCreate531490(record, ovalShieldRuntime531490())
	}
	if callback == legacy.Get_nox_xxx_spellCreateMoonglow_531A00() {
		return server.SpellMoonglowCreate531A00(record, sp.moonglowRuntime531A00())
	}
	if callback == legacy.Get_nox_xxx_manaBomb_530F90() {
		return server.SpellManaBombCreate530F90(record, sp.manaBombRuntime530F90())
	}
	if callback == legacy.Get_nox_xxx_onStartLightning_52F820() {
		return server.SpellChainLightningCreate52F820(record, sp.chainLightningRuntime52F820())
	}
	if callback == legacy.Get_nox_xxx_spellEnergyBoltStop_52E820() {
		return server.SpellEnergyBoltCreate52E820(record, sp.energyBoltRuntime52E820())
	}
	traceCDurationCall("create", callback, record)
	return int32(ccall.CallIntPtr(callback, record.C()))
}

func (sp *spellsDuration) blinkRuntime530310() server.SpellBlinkRuntime530310 {
	world := sp.s.S()
	return server.SpellBlinkRuntime530310{
		QuestMode: func() bool { return noxflags.HasGame(noxflags.GameModeQuest) },
		CoopMode:  func() bool { return noxflags.HasGame(noxflags.GameModeCoop) },
		Frame:     sp.s.Frame,
		TickRate:  world.TickRate,
		TeleportDelay: func(levelIndex uint32) float32 {
			return float32(sp.s.Balance.FloatInd("TeleportDelay", int(int32(levelIndex))))
		},
		Waypoint:        world.Nox_xxx_waypoint_579F00,
		PlayerStart:     sp.s.nox_xxx_mapFindPlayerStart_4F7AB0,
		RandomReachable: world.RandomReachablePointAround,
		NewObject:       world.NewObjectByTypeID,
		CreateAt: func(object, owner *server.Object, point types.Pointf) {
			sp.s.CreateObjectAt(object, owner, point)
		},
		SendPointFX: world.Nox_xxx_netSendPointFx_522FF0,
		CastSound: func(id spell.ID) sound.ID {
			return world.Spells.DefByInd(id).GetCastSound()
		},
		Audio: func(id sound.ID, obj *server.Object, kind int, code uint32) {
			sp.s.Audio.EventObj(id, obj, kind, code)
		},
		Teleport:    legacy.TeleportToMB4E7190,
		Attribution: legacy.Sub_4E7540,
	}
}

func (sp *spellsDuration) swapRuntime530CA0() server.SpellSwapRuntime530CA0 {
	world := sp.s.S()
	return server.SpellSwapRuntime530CA0{
		CoopMode: func() bool { return noxflags.HasGame(noxflags.GameModeCoop) },
		Frame:    sp.s.Frame,
		TeleportDelay: func(levelIndex uint32) float32 {
			return float32(sp.s.Balance.FloatInd("TeleportDelay", int(int32(levelIndex))))
		},
		CanInteract: func(caster, target *server.Object) bool {
			return world.CanInteract(caster, target, 0)
		},
		InformNoLOS: func(caster *server.Object) {
			world.NetPriMsgToPlayer(caster, "ExecDur.c:NeedClearLOSForSwap", 0)
		},
		SendPointFX: world.Nox_xxx_netSendPointFx_522FF0,
		CastSound: func(id spell.ID) sound.ID {
			return world.Spells.DefByInd(id).GetCastSound()
		},
		Audio: func(id sound.ID, obj *server.Object, kind int, code uint32) {
			sp.s.Audio.EventObj(id, obj, kind, code)
		},
		Teleport:    legacy.TeleportToMB4E7190,
		Attribution: legacy.Sub_4E7540,
	}
}

func (sp *spellsDuration) teleportToTargetRuntime530A30() server.SpellTeleportToTargetRuntime530A30 {
	world := sp.s.S()
	return server.SpellTeleportToTargetRuntime530A30{
		CoopMode: func() bool { return noxflags.HasGame(noxflags.GameModeCoop) },
		Frame:    sp.s.Frame,
		TickRate: world.TickRate,
		TeleportDelay: func(levelIndex uint32) float32 {
			return float32(sp.s.Balance.FloatInd("TeleportDelay", int(int32(levelIndex))))
		},
		TileBlocksTeleport: legacy.MapTileBlocksTeleport411A90,
		TraceRay9:          world.MapTraceRay9,
		SendUnseenTarget: func(target *server.Object) {
			message := world.Strings().GetStringInFile("UnseenTarget", "C:\\NoxPost\\src\\Server\\Magic\\Spell\\ExecDur.c")
			legacy.Nox_xxx_netSendLineMessage_4D9EB0(target, message)
		},
		InformNoLOS: func(caster *server.Object) {
			if caster.UpdateData == nil || caster.UpdateDataPlayer().Player == nil {
				return
			}
			world.NetInformTextMsg(caster.UpdateDataPlayer().Player.PlayerIndex(), 0, 2)
		},
		NewObject: world.NewObjectByTypeID,
		CreateAt: func(object, owner *server.Object, point types.Pointf) {
			sp.s.CreateObjectAt(object, owner, point)
		},
		SendPointFX: world.Nox_xxx_netSendPointFx_522FF0,
		CastSound: func(id spell.ID) sound.ID {
			return world.Spells.DefByInd(id).GetCastSound()
		},
		Audio: func(id sound.ID, obj *server.Object, kind int, code uint32) {
			sp.s.Audio.EventObj(id, obj, kind, code)
		},
		Teleport:    legacy.TeleportToMB4E7190,
		Attribution: legacy.Sub_4E7540,
	}
}

func (sp *spellsDuration) teleportPopRuntime530820() server.SpellTeleportPopRuntime530820 {
	world := sp.s.S()
	return server.SpellTeleportPopRuntime530820{
		CoopMode: func() bool { return noxflags.HasGame(noxflags.GameModeCoop) },
		Frame:    sp.s.Frame,
		TickRate: world.TickRate,
		TeleportDelay: func(levelIndex uint32) float32 {
			return float32(sp.s.Balance.FloatInd("TeleportDelay", int(int32(levelIndex))))
		},
		RandomIndex: func() int { return world.Rand.Logic.IntClamp(0, 3) },
		NewObject:   world.NewObjectByTypeID,
		CreateAt: func(object, owner *server.Object, point types.Pointf) {
			sp.s.CreateObjectAt(object, owner, point)
		},
		SendPointFX: world.Nox_xxx_netSendPointFx_522FF0,
		CastSound: func(id spell.ID) sound.ID {
			return world.Spells.DefByInd(id).GetCastSound()
		},
		Audio: func(id sound.ID, obj *server.Object, kind int, code uint32) {
			sp.s.Audio.EventObj(id, obj, kind, code)
		},
		Teleport:      legacy.TeleportToMB4E7190,
		DelayedDelete: sp.s.DelayedDelete,
		Attribution:   legacy.Sub_4E7540,
	}
}

func (sp *spellsDuration) turnUndeadRuntime531310() server.SpellTurnUndeadRuntime531310 {
	world := sp.s.S()
	return server.SpellTurnUndeadRuntime531310{
		KillPoints: func(levelIndex uint32) float32 {
			return float32(sp.s.Balance.FloatInd("TurnUndeadKillPoints", int(int32(levelIndex))))
		},
		NewObject: world.NewObjectByTypeID,
		CreateAt: func(object, owner *server.Object, point types.Pointf) {
			sp.s.CreateObjectAt(object, owner, point)
		},
		SendPointFX: func(effect netmsg.Op, point types.Pointf) {
			world.Nox_xxx_netSendPointFx_522FF0(effect, point)
		},
		FirstObject:   world.Objs.First,
		TypeInd:       world.Types.IndByID,
		DelayedDelete: sp.s.DelayedDelete,
	}
}

func (sp *spellsDuration) forceOfNatureRuntime52EF30() server.SpellForceOfNatureRuntime52EF30 {
	return server.SpellForceOfNatureRuntime52EF30{
		NewObject: sp.s.S().NewObjectByTypeID,
		CreateAt: func(object, owner *server.Object, point types.Pointf) {
			sp.s.CreateObjectAt(object, owner, point)
			if int(object.TypeInd) == sp.s.S().Types.DeathBallID() {
				sp.forceOfNatureLaunches++
			}
		},
		DelayedDelete: sp.s.DelayedDelete,
		LoadCharge: func(record *server.DurSpell) *server.Object {
			return sp.forceOfNatureCharges[record]
		},
		StoreCharge: func(record *server.DurSpell, charge *server.Object) {
			if charge == nil {
				delete(sp.forceOfNatureCharges, record)
				return
			}
			if sp.forceOfNatureCharges == nil {
				sp.forceOfNatureCharges = make(map[*server.DurSpell]*server.Object)
			}
			sp.forceOfNatureCharges[record] = charge
		},
		CurrentFrame: sp.s.Frame,
		TraceRay:     sp.s.S().MapTraceRay,
		SetPlayerState: func(unit *server.Object, state server.PlayerState) {
			_ = nox_xxx_playerSetState_4FA020(unit, state)
		},
		EventObj: func(unit *server.Object) {
			sp.s.Audio.EventObj(sound.ID(38), unit, 0, 0)
		},
	}
}

func (sp *spellsDuration) moonglowRuntime531A00() server.SpellMoonglowRuntime531A00 {
	return server.SpellMoonglowRuntime531A00{
		EnchantmentDuration: func() float32 {
			return float32(sp.s.Balance.Float("MoonglowEnchantmentDuration"))
		},
		NewObject: sp.s.S().NewObjectByTypeID,
		CreateAt: func(visual, target *server.Object, point types.Pointf) {
			sp.s.CreateObjectAt(visual, target, point)
		},
		ApplyBuff: func(target *server.Object, buff server.EnchantID, duration int16, power int8) {
			legacy.Nox_xxx_buffApplyTo_4FF380(target, buff, int(duration), int(power))
		},
		BuffOff:       legacy.Nox_xxx_spellBuffOff_4FF5B0,
		DelayedDelete: sp.s.DelayedDelete,
		LoadVisual: func(record *server.DurSpell) *server.Object {
			return sp.moonglowVisuals[record]
		},
		StoreVisual: func(record *server.DurSpell, visual *server.Object) {
			if visual == nil {
				delete(sp.moonglowVisuals, record)
				return
			}
			if sp.moonglowVisuals == nil {
				sp.moonglowVisuals = make(map[*server.DurSpell]*server.Object)
			}
			sp.moonglowVisuals[record] = visual
		},
	}
}

func (sp *spellsDuration) manaBombRuntime530F90() server.SpellManaBombRuntime530F90 {
	world := sp.s.S()
	return server.SpellManaBombRuntime530F90{
		Frame:    sp.s.Frame,
		TickRate: world.TickRate,
		Balance: func(key string) float32 {
			return float32(sp.s.Balance.Float(key))
		},
		BalanceLevel: func(key string, level uint32) float32 {
			return float32(sp.s.Balance.FloatInd(key, int(int32(level))))
		},
		NewObject: world.NewObjectByTypeID,
		CreateAt: func(charge, owner *server.Object, point types.Pointf) {
			sp.s.CreateObjectAt(charge, owner, point)
		},
		LoadCharge: func(record *server.DurSpell) *server.Object {
			return sp.manaBombCharges[record]
		},
		StoreCharge: func(record *server.DurSpell, charge *server.Object) {
			if charge == nil {
				delete(sp.manaBombCharges, record)
				return
			}
			if sp.manaBombCharges == nil {
				sp.manaBombCharges = make(map[*server.DurSpell]*server.Object)
			}
			sp.manaBombCharges[record] = charge
		},
		DelayedDelete: sp.s.DelayedDelete,
		ApplyBuff: func(caster *server.Object, buff server.EnchantID, duration int16, power int8) {
			legacy.Nox_xxx_buffApplyTo_4FF380(caster, buff, int(duration), int(power))
		},
		BuffOff: legacy.Nox_xxx_spellBuffOff_4FF5B0,
		DamageAround: func(pos types.Pointf, outer, inner float32, damage int, caster *server.Object) {
			sp.s.Nox_xxx_mapDamageUnitsAround(pos, outer, inner, damage, object.DamageType(15), caster, nil, true)
		},
		Earthquake: world.Nox_xxx_earthquakeSend_4D9110,
		PointFX:    world.Nox_xxx_netSendPointFx_522FF0,
		Audio: func(caster *server.Object) {
			sp.s.Audio.EventObj(sound.ID(81), caster, 0, 0)
		},
		ManaSub: func(caster *server.Object, amount int32) {
			legacy.Nox_xxx_playerManaSub_4EEBF0(caster, int(amount))
		},
	}
}

func (sp *spellsDuration) chainLightningRuntime52F820() server.SpellChainLightningRuntime52F820 {
	world := sp.s.S()
	return server.SpellChainLightningRuntime52F820{
		Frame:    sp.s.Frame,
		TickRate: world.TickRate,
		Balance: func(key string) float32 {
			return float32(sp.s.Balance.Float(key))
		},
		SpellLevel: func(id uint32) uint32 {
			return memmap.Uint32(0x587000, 260380+uintptr(4*id))
		},
		ObjectsInCircle: world.EachChainLightningObject52F8A0,
		CanInteract: func(source, target *server.Object) bool {
			return world.CanInteract(source, target, 0)
		},
		IsEnemy:       world.IsEnemyTo,
		TraceRay:      world.MapTraceRay,
		PositionDelta: world.PositionDelta4FEA70,
		CancelSpell: func(id int32, caster *server.Object) {
			world.Spells.Dur.SpellCancelDurSpell4FEB10(id, caster)
		},
		FreeDuration:    world.Spells.Dur.FreeRecursive,
		NewLightningSub: world.Spells.Dur.NewLightningSub,
		StartRay:        world.NetStartDurationRaySpell,
		StopRay:         world.NetStopRaySpell,
		PointFX: func(code uint8, pos types.Pointf) {
			world.Nox_xxx_netSendPointFx_522FF0(netmsg.Op(code), pos)
		},
		Damage: func(target, caster *server.Object, amount int32) {
			target.CallDamage(caster, nil, int(amount), object.DamageAirborneElectric)
		},
		Audio: func(id uint16, target *server.Object) {
			sp.s.Audio.EventObj(sound.ID(id), target, 0, 0)
		},
		CastSound: func() uint16 {
			return uint16(world.Spells.DefByInd(spell.SPELL_CHAIN_LIGHTNING).GetCastSound())
		},
		SetPlayerState: func(caster *server.Object, state server.PlayerState) {
			_ = nox_xxx_playerSetState_4FA020(caster, state)
		},
		ManaSub: func(caster *server.Object, amount int32) {
			legacy.Nox_xxx_playerManaSub_4EEBF0(caster, int(amount))
		},
		ReportCharges: func(caster, wand *server.Object, charge, maxCharge uint8) {
			legacy.Nox_xxx_netReportCharges_4D82B0(caster.UpdateDataPlayer().Player.PlayerInd, wand, charge, maxCharge)
		},
		LoadWeapon: func(record *server.DurSpell) *server.Object {
			return sp.chainLightningWeapons[record]
		},
		StoreWeapon: func(record *server.DurSpell, wand *server.Object) {
			if wand == nil {
				delete(sp.chainLightningWeapons, record)
				return
			}
			if sp.chainLightningWeapons == nil {
				sp.chainLightningWeapons = make(map[*server.DurSpell]*server.Object)
			}
			sp.chainLightningWeapons[record] = wand
		},
	}
}

func (sp *spellsDuration) energyBoltRuntime52E820() server.SpellEnergyBoltRuntime52E820 {
	world := sp.s.S()
	return server.SpellEnergyBoltRuntime52E820{
		Frame:    sp.s.Frame,
		TickRate: world.TickRate,
		Balance: func(key string) float32 {
			return float32(sp.s.Balance.Float(key))
		},
		BalanceLevel: func(key string, level uint32) float32 {
			return float32(sp.s.Balance.FloatInd(key, int(int32(level))))
		},
		ObjectsInCircle: world.EachChainLightningObject52F8A0,
		CanInteract: func(source, target *server.Object) bool {
			return world.CanInteract(source, target, 0)
		},
		IsEnemy: world.IsEnemyTo,
		InFront: func(source, target *server.Object) bool {
			return legacy.Nox_server_testTwoPointsAndDirection_4E6E50(source.PosVec, int16(source.Direction1), target.PosVec)&1 != 0
		},
		PositionDelta: world.PositionDelta4FEA70,
		CancelSpell: func(id int32, caster *server.Object) {
			world.Spells.Dur.SpellCancelDurSpell4FEB10(id, caster)
		},
		StartRay: world.NetStartDurationRaySpell,
		StopRay:  world.NetStopRaySpell,
		PointFX: func(code uint8, pos types.Pointf) {
			world.Nox_xxx_netSendPointFx_522FF0(netmsg.Op(code), pos)
		},
		Damage: func(target, caster *server.Object, amount int32) {
			target.CallDamage(caster, nil, int(amount), object.DamageAirborneElectric)
		},
		Audio: func(id uint16, target *server.Object) {
			sp.s.Audio.EventObj(sound.ID(id), target, 0, 0)
		},
		CastSound: func() uint16 {
			return uint16(world.Spells.DefByInd(spell.SPELL_LIGHTNING).GetCastSound())
		},
		SetPlayerState: func(caster *server.Object, state server.PlayerState) {
			_ = nox_xxx_playerSetState_4FA020(caster, state)
		},
		ManaSub: func(caster *server.Object, amount int32) {
			legacy.Nox_xxx_playerManaSub_4EEBF0(caster, int(amount))
		},
		LoadRayTarget: func(record *server.DurSpell) *server.Object {
			return sp.durationRayTargets[record]
		},
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

func (sp *spellsDuration) drainManaRuntime52E210() server.SpellDrainManaRuntime52E210 {
	world := sp.s.S()
	return server.SpellDrainManaRuntime52E210{
		Frame:    sp.s.Frame,
		TickRate: world.TickRate,
		Balance: func(key string) float32 {
			return float32(sp.s.Balance.Float(key))
		},
		BalanceLevel: func(key string, level uint32) float32 {
			return float32(sp.s.Balance.FloatInd(key, int(int32(level))))
		},
		ObjectsInCircle: world.EachChainLightningObject52F8A0,
		IsEnemy:         world.IsEnemyTo,
		SameTeam:        server.UnitsHaveSameTeam4EC520,
		TraceRay:        world.MapTraceRay,
		PositionDelta:   world.PositionDelta4FEA70,
		QuestMode: func() bool {
			return noxflags.HasGame(noxflags.GameModeQuest)
		},
		QuestManaScale: func(caster *server.Object) float32 {
			return world.Players.ClassStatsMult(caster.UpdateDataPlayer().Player.PlayerClass()).Mana
		},
		AddMana: func(caster *server.Object, amount int16) {
			legacy.Nox_xxx_playerManaAdd_4EEB80(caster, int(amount))
		},
		SubMana: func(target *server.Object, amount int32) {
			legacy.Nox_xxx_playerManaSub_4EEBF0(target, int(amount))
		},
		StartRay: world.NetStartDurationRaySpell,
		StopRay:  world.NetStopRaySpell,
		Audio: func(id uint16, target *server.Object) {
			sp.s.Audio.EventObj(sound.ID(id), target, 0, 0)
		},
		LoadRayTarget: func(record *server.DurSpell) *server.Object {
			return sp.durationRayTargets[record]
		},
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

func ovalShieldRuntime531490() server.SpellOvalShieldRuntime531490 {
	return server.SpellOvalShieldRuntime531490{
		ApplyBuff: func(target *server.Object, buff int32, duration int16, power int8) {
			legacy.Nox_xxx_buffApplyTo_4FF380(target, server.EnchantID(buff), int(duration), int(power))
		},
		BuffOff: func(target *server.Object, buff int32) {
			legacy.Nox_xxx_spellBuffOff_4FF5B0(target, server.EnchantID(buff))
		},
	}
}

func (sp *spellsDuration) New(spellID spell.ID, u1, u2, u3 *server.Object, sa *server.SpellAcceptArg, lvl int, create, update, destroy unsafe.Pointer, dt uint32) bool {
	return sp.SpellsDuration.SpellDurationCreate4FEBA0(
		int32(spellID),
		u1,
		u2,
		u3,
		sa,
		int32(lvl),
		create,
		update,
		destroy,
		dt,
		server.SpellDurationCreateRuntime4FEBA0{
			DestroySpell: sp.destroyDurSpell,
			CallCreate:   sp.callCreate4FEBA0,
			AudioEvent: func(id sound.ID, object *server.Object, kind int, code uint32) {
				sp.s.Audio.EventObj(id, object, kind, code)
			},
		},
	) != 0
}

// SpellDurationCreate4FEBA0 is the legacy C ABI bridge for GAME.EXE
// 004FEBA0. Pointer arguments remain native-width, while duration is converted
// from its signed C dword by bit pattern before the original wrapping frame
// addition is performed by the server model.
func (s *Server) SpellDurationCreate4FEBA0(
	spellID int32,
	second, third, fourth *server.Object,
	arg *server.SpellAcceptArg,
	level int32,
	create, update, destroy unsafe.Pointer,
	duration int32,
) int32 {
	return spellAcceptBool4FD400(s.spells.duration.New(
		spell.ID(spellID),
		second,
		third,
		fourth,
		arg,
		int(level),
		create,
		update,
		destroy,
		uint32(duration),
	))
}
