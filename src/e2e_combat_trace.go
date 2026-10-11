package opennox

import (
	"encoding/json"
	"math"
	"os"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

// Opt-in evidence from executed frames, not a combat fixture. No input,
// health, clock, action, inventory, allegiance, or RNG is supplied by this
// observer. Its 240-frame recording limit does not change scenario timeouts.
var e2eCombatTraceEnabled = os.Getenv("NOX_E2E_COMBAT_TRACE") == "true"

type e2eCombatTraceUnit struct {
	Pointer, Update, HealthPointer, DamageSource uintptr
	Type                                         uint16
	Class, Subclass, Flags, Buffs                uint32
	Health                                       [3]uint16
	PositionBits, PreviousBits, VelocityBits     [2]uint32
	Direction                                    [2]uint16
	AttackStart, DamageFrame                     uint32
	State, Animation                             uint8
	AnimationFlags                               [3]uint8
	StateDeadline, AttackDeadline, Status        uint32
	Stamina                                      uint8
	Player, Weapon, Cursor, Current, Preferred   uintptr
	Equipment, Strength, ArmorBits               uint32
	StackIndex                                   int8
	Stack                                        [24]uint32
	HeadArgs                                     [4]uintptr
	Seen                                         uint8
}

func e2eCombatTraceSnapshot(unit *server.Object) (e2eCombatTraceUnit, bool) {
	if unit == nil || !unit.Class().HasAny(object.ClassPlayer|object.ClassMonster) {
		return e2eCombatTraceUnit{}, false
	}
	state := e2eCombatTraceUnit{
		Pointer: uintptr(unsafe.Pointer(unit)), Update: uintptr(unit.UpdateData),
		HealthPointer: uintptr(unsafe.Pointer(unit.HealthData)), DamageSource: uintptr(unsafe.Pointer(unit.Obj130)),
		Type: unit.TypeInd, Class: uint32(unit.ObjClass), Subclass: uint32(unit.ObjSubClass),
		Flags: uint32(unit.ObjFlags), Buffs: unit.Buffs,
		PositionBits: [2]uint32{math.Float32bits(unit.PosVec.X), math.Float32bits(unit.PosVec.Y)},
		PreviousBits: [2]uint32{math.Float32bits(unit.PrevPos.X), math.Float32bits(unit.PrevPos.Y)},
		VelocityBits: [2]uint32{math.Float32bits(unit.VelVec.X), math.Float32bits(unit.VelVec.Y)},
		Direction:    [2]uint16{uint16(unit.Direction1), uint16(unit.Direction2)},
		AttackStart:  unit.Field34, DamageFrame: unit.Frame134, StackIndex: -1,
	}
	if unit.HealthData != nil {
		state.Health = [3]uint16{unit.HealthData.Cur, unit.HealthData.Max, unit.HealthData.Field2}
	}
	if unit.UpdateData == nil {
		return state, true
	}
	if unit.Class().Has(object.ClassPlayer) {
		update := unit.UpdateDataPlayer()
		state.State, state.Animation, state.Stamina = uint8(update.State), update.Field59_0, update.Stamina
		state.StateDeadline, state.ArmorBits = update.Field0, update.Field57
		state.Player, state.Weapon, state.Cursor = uintptr(unsafe.Pointer(update.Player)), uintptr(unsafe.Pointer(update.EquippedWeapon)), uintptr(unsafe.Pointer(update.CursorObj))
		if update.Player != nil {
			state.Equipment, state.Strength = update.Player.WeaponEquip, update.Player.Info().Field2239()
		}
	} else {
		update := unit.UpdateDataMonster()
		state.Animation = update.Field120_1
		state.AnimationFlags = [3]uint8{update.Field120_0, update.Field120_2, update.Field120_3}
		state.AttackDeadline, state.Status, state.Stamina = update.Field128, uint32(update.StatusFlags), update.Stamina
		state.Current, state.Preferred = uintptr(unsafe.Pointer(update.CurrentEnemy)), uintptr(unsafe.Pointer(update.PreferredEnemy))
		state.Equipment, state.StackIndex, state.Seen = update.WeaponEquipFlags, update.AIStackInd, update.Field282_1
		if state.StackIndex >= 0 && int(state.StackIndex) < len(update.AIStack) {
			for i := 0; i <= int(state.StackIndex); i++ {
				state.Stack[i] = update.AIStack[i].Action
			}
			state.HeadArgs = update.AIStack[state.StackIndex].Args
		}
	}
	return state, true
}

var e2eCombatTrace struct {
	player, monster *server.Object
	start, last     uint32
	started, ended  bool
}

func e2eCombatTraceTick() {
	if !e2eCombatTraceEnabled || e2ePlay == "" || noxServer == nil || e2e.monster == nil {
		return
	}
	player, monster := noxServer.Players.HostUnit(), e2e.monster
	frame := noxServer.Frame()
	if !e2eCombatTrace.started || e2eCombatTrace.player != player || e2eCombatTrace.monster != monster {
		e2eCombatTrace.player, e2eCombatTrace.monster = player, monster
		e2eCombatTrace.start, e2eCombatTrace.last = frame, frame-1
		e2eCombatTrace.started, e2eCombatTrace.ended = true, false
		if monster.UpdateData != nil && monster.Class().Has(object.ClassMonster) {
			e2eLog.Printf("COMBAT TRACE DEFINITION: monster=%p def=%+v", monster, monster.UpdateDataMonster().MonsterDef)
		}
		if player != nil && player.UpdateData != nil && player.Class().Has(object.ClassPlayer) {
			weapon := player.UpdateDataPlayer().EquippedWeapon
			if weapon != nil {
				e2eLog.Printf("COMBAT TRACE WEAPON: weapon=%p definition=%+v", weapon, noxServer.Modif.Nox_xxx_getProjectileClassById413250(int(weapon.TypeInd)))
			}
		}
	}
	if e2eCombatTrace.ended || e2eCombatTrace.last == frame {
		return
	}
	if frame-e2eCombatTrace.start >= 240 {
		e2eCombatTrace.ended = true
		e2eLog.Printf("COMBAT TRACE END: frame=%d first=%d limit=240; scenario input and timeout unchanged", frame, e2eCombatTrace.start)
		return
	}
	e2eCombatTrace.last = frame
	playerState, playerOK := e2eCombatTraceSnapshot(player)
	monsterState, monsterOK := e2eCombatTraceSnapshot(monster)
	row := struct {
		Frame, FPS          uint32
		Logic, Other        int
		PlayerOK, MonsterOK bool
		Player, Monster     e2eCombatTraceUnit
	}{frame, noxServer.TickRate(), noxServer.Rand.Logic.Index(), noxServer.Rand.Other.Index(), playerOK, monsterOK, playerState, monsterState}
	data, err := json.Marshal(row)
	if err != nil {
		e2eLog.Printf("COMBAT TRACE ENCODE: %v", err)
		return
	}
	e2eLog.Printf("COMBAT TRACE FRAME: %s", data)
}
