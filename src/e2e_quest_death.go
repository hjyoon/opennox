package opennox

import (
	"fmt"
	"image"
	"math"
	"reflect"
	"unsafe"

	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

// Observation only. Resolve the declared native DWORD, never a PE32 byte
// offset; none of the death fixture's statistics or life counts are written.
func e2eQuestDeathPlayerDWORD(player *server.Player, name string) (uint32, error) {
	if player == nil {
		return 0, fmt.Errorf("Quest death observation has no Player")
	}
	field := reflect.ValueOf(player).Elem().FieldByName(name)
	if !field.IsValid() || field.Kind() != reflect.Uint32 {
		return 0, fmt.Errorf("Quest death observation: %q is not a native DWORD", name)
	}
	return uint32(field.Uint()), nil
}

type e2eQuestDeathState struct {
	lives, deaths, gold, blocker         uint32
	stage, generators, monsters, secrets uint32
	health, maximum                      uint16
	state                                server.PlayerState
	flags                                object.Flags
	scoreVisible                         bool
}

func e2eReadQuestDeathState(unit *server.Object) (e2eQuestDeathState, error) {
	if unit == nil || !e2eObjectInWorld(unit) || unit.UpdateData == nil || unit.HealthData == nil || unit.ControllingPlayer() == nil {
		return e2eQuestDeathState{}, fmt.Errorf("Quest death observation lost native unit %p", unit)
	}
	update, player := unit.UpdateDataPlayer(), unit.ControllingPlayer()
	state := e2eQuestDeathState{
		lives: update.ExtraLives, gold: player.GoldVal, blocker: update.Field137,
		health: unit.HealthData.Cur, maximum: unit.HealthData.Max, state: update.State, flags: unit.Flags(),
	}
	for _, member := range []struct {
		name  string
		value *uint32
	}{
		{"field4660", &state.deaths}, {"field4688", &state.stage}, {"field4668", &state.generators},
		{"field4664", &state.monsters}, {"field4672", &state.secrets},
	} {
		value, err := e2eQuestDeathPlayerDWORD(player, member.name)
		if err != nil {
			return e2eQuestDeathState{}, err
		}
		*member.value = value
	}
	if score := noxClient.GUI.ChildByID(10700); score != nil {
		state.scoreVisible = !score.GetFlags().IsHidden()
	}
	return state, nil
}

func e2eQuestDeathTransition(before, after e2eQuestDeathState) error {
	if after.health != 0 || !after.flags.Has(object.FlagDead) || after.flags.Has(object.FlagDestroyed) || after.state != server.PlayerState4 {
		return fmt.Errorf("Quest death did not reach normal dead state: %+v", after)
	}
	if before.lives != 0 {
		if after.lives != before.lives-1 || after.deaths != before.deaths+1 || after.gold != before.gold || after.blocker != 0 || after.scoreVisible {
			return fmt.Errorf("Quest positive-life death changed the wrong branch: before=%+v after=%+v", before, after)
		}
	} else if after.lives != 2 || after.deaths != 0 || after.generators != 0 || after.monsters != 0 || after.secrets != 0 || after.stage != 5 || after.blocker == 0 || !after.scoreVisible || after.gold > before.gold {
		return fmt.Errorf("Quest zero-life statistics/reset/penalty branch is incomplete: before=%+v after=%+v", before, after)
	}
	return nil
}

type e2eQuestDeathFixture struct {
	combat                   e2eQuestCombatFixture
	before, dead             e2eQuestDeathState
	cycle                    int
	keepaliveFrame           uint32
	keepalivePressed         bool
	incoming, clientIncoming bool
	lethalSeen, lethalMinion bool
	lethalSource             *server.Object
	lethalType, lethalFrame  uint32
}

func (f *e2eQuestDeathFixture) hasNaturalDamageEvidence() bool {
	return f.incoming && f.clientIncoming && f.lethalSeen
}

func (f *e2eQuestDeathFixture) prepare() {
	f.combat = e2eQuestCombatFixture{typeID: "Necromancer"}
	f.combat.prepare() // The only fixture is placement beside a stock minion.
	if e2e.err != nil {
		return
	}
	var err error
	f.before, err = e2eReadQuestDeathState(f.combat.unit)
	stockLives := float32(noxServer.Balance.Float("QuestGameStartingExtraLives"))
	if err != nil || stockLives != 2 || f.before.lives != uint32(2-f.cycle) || f.before.blocker != 0 || f.before.scoreVisible {
		e2eError(fmt.Errorf("Quest death cycle %d starting state: %+v stock-balance=%g error=%v", f.cycle+1, f.before, stockLives, err))
		return
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, pointer := range []unsafe.Pointer{f.combat.unit.CObj(), f.combat.target.CObj(), f.combat.unit.UpdateData, unsafe.Pointer(f.combat.unit.ControllingPlayer()), unsafe.Pointer(f.combat.unit.HealthData)} {
			if uintptr(pointer) <= math.MaxUint32 {
				e2eError(fmt.Errorf("Quest death requires high native pointer, got %p", pointer))
				return
			}
		}
	}
	f.keepaliveFrame, f.keepalivePressed = noxServer.Frame(), false
	f.incoming, f.clientIncoming = false, false
	f.lethalSeen, f.lethalMinion = false, false
	f.lethalSource, f.lethalType, f.lethalFrame = nil, 0, 0
	e2eLog.Printf("QUEST PLAYER DEATH PREPARED: cycle=%d frame=%d player=%p monster=%p hp=%d/%d lives=%d deaths=%d gold=%d injected=placement-only", f.cycle+1, f.keepaliveFrame, f.combat.unit, f.combat.target, f.before.health, f.before.maximum, f.before.lives, f.before.deaths, f.before.gold)
}

func (f *e2eQuestDeathFixture) naturallyDead() bool {
	unit, target := f.combat.unit, f.combat.target
	state, err := e2eReadQuestDeathState(unit)
	if err != nil {
		e2eError(err)
		return true
	}
	f.combat.ticks++
	if e2eObjectInWorld(target) && state.health < f.before.health && e2eQuestCombatAttributedTo(unit.Obj130, target) {
		f.incoming = true
	}
	if drawable := noxClient.ClientPlayerUnit(); drawable != nil {
		if delta, ok := legacy.HealthChangeForDrawable(drawable.NetCode32); ok && delta < 0 {
			f.clientIncoming = true
		}
	}
	if !f.lethalSeen && state.health == 0 && state.flags.Has(object.FlagDead) {
		// Observe attribution when death begins. The original projectile can
		// leave the world during the death animation; never dereference it
		// after waiting for the ordinary input-ready dead state.
		f.lethalSeen = true
		f.lethalMinion = e2eQuestCombatAttributedTo(unit.Obj130, target)
		f.lethalSource, f.lethalType, f.lethalFrame = unit.Obj130, unit.Field131, unit.Frame134
		e2eLog.Printf("QUEST PLAYER LETHAL DAMAGE: cycle=%d frame=%d state=%d source=%p type=%d damage-frame=%d minion=%p attributed=%t source-in-world=%t", f.cycle+1, noxServer.Frame(), state.state, f.lethalSource, f.lethalType, f.lethalFrame, target, f.lethalMinion, e2eObjectInWorld(f.lethalSource))
	}
	now := noxServer.Frame()
	if f.keepalivePressed {
		e2eQueueInput(&seat.MouseButtonEvent{Button: seat.MouseButtonRight, Pressed: false})
		f.keepalivePressed = false
	} else if state.health != 0 && !state.flags.Has(object.FlagDead) && now-f.keepaliveFrame >= 300 {
		// Normal one-frame movement keeps the long natural fight active. Do
		// not change observer flags, timers, AI, HP or damage callbacks.
		offset := 16
		if f.combat.ticks/300&1 != 0 {
			offset = -offset
		}
		pos := noxClient.Viewport().ToScreenPos(image.Pt(int(unit.PosVec.X)+offset, int(unit.PosVec.Y)))
		e2eQueueInput(&seat.MouseMoveEvent{Pos: pos}, &seat.MouseButtonEvent{Button: seat.MouseButtonRight, Pressed: true})
		f.keepaliveFrame, f.keepalivePressed = now, true
	}
	if f.combat.ticks%300 == 0 {
		f.combat.logState()
	}
	if state.health != 0 || !state.flags.Has(object.FlagDead) || state.state != server.PlayerState4 {
		return false
	}
	// Prove observed minion damage and a natural death, not exclusive lethal
	// attribution. Other stock attackers remain active, and the lethal
	// projectile may already be removed before the next observation tick.
	if !f.hasNaturalDamageEvidence() {
		e2eError(fmt.Errorf("Quest natural death lacks observed minion/client damage: cycle=%d incoming=%t client=%t lethal-seen=%t", f.cycle+1, f.incoming, f.clientIncoming, f.lethalSeen))
		return true
	}
	if err := e2eQuestDeathTransition(f.before, state); err != nil {
		e2eError(err)
		return true
	}
	f.dead = state
	e2eLog.Printf("QUEST PLAYER NATURALLY DEAD: cycle=%d frame=%d hp=%d->0 lives=%d->%d deaths=%d->%d gold=%d->%d state=%d lethal-source=%p lethal-type=%d lethal-frame=%d exclusive-minion-attribution=unverified blocker=%d score-visible=%t client-damage=true", f.cycle+1, now, f.before.health, f.before.lives, state.lives, f.before.deaths, state.deaths, f.before.gold, state.gold, state.state, f.lethalSource, f.lethalType, f.lethalFrame, state.blocker, state.scoreVisible)
	return true
}

func (f *e2eQuestDeathFixture) respawnMouse() {
	var pos image.Point
	if f.before.lives != 0 {
		pos = noxClient.Viewport().ToScreenPos(image.Pt(int(f.combat.unit.PosVec.X), int(f.combat.unit.PosVec.Y)))
	} else {
		root := noxClient.GUI.ChildByID(10700)
		if root == nil || root.GetFlags().IsHidden() {
			e2eError(fmt.Errorf("Quest game-over window unavailable"))
			return
		}
		button := root.ChildByID(10702)
		if button == nil || button.GetFlags().IsHidden() || !button.GetFlags().IsEnabled() {
			e2eError(fmt.Errorf("Quest game-over Continue button unavailable"))
			return
		}
		pos = button.GlobalPos().Add(button.Size().Div(2))
	}
	e2eQueueInput(&seat.MouseMoveEvent{Pos: pos}, &seat.MouseButtonEvent{Button: seat.MouseButtonRight, Pressed: false})
	e2eLog.Printf("QUEST PLAYER RESPAWN INPUT: cycle=%d point=%v game-over=%t", f.cycle+1, pos, f.before.lives == 0)
}

func (f *e2eQuestDeathFixture) respawned() bool {
	state, err := e2eReadQuestPlayer()
	if err != nil || state.validate(5, "g_forest") != nil {
		return false
	}
	if noxServer.Players.HostUnit() != f.combat.unit {
		e2eError(fmt.Errorf("Quest respawn replaced the player identity"))
		return true
	}
	after, err := e2eReadQuestDeathState(f.combat.unit)
	if err != nil {
		e2eError(err)
		return true
	}
	if after.lives != f.dead.lives || after.deaths != f.dead.deaths || after.gold != f.dead.gold || after.blocker != 0 || after.scoreVisible || after.health != after.maximum {
		e2eError(fmt.Errorf("Quest real-input respawn state incomplete: dead=%+v respawn=%+v", f.dead, after))
		return true
	}
	e2eLog.Printf("QUEST PLAYER RESPAWNED: cycle=%d frame=%d player=%p hp=%d/%d lives=%d deaths=%d gold=%d pos=%v blocker=0 score-hidden=true", f.cycle+1, noxServer.Frame(), f.combat.unit, after.health, after.maximum, after.lives, after.deaths, after.gold, state.position)
	f.cycle++
	return true
}

// Three natural deaths cover the stock 2->1->0 extra-life decrements followed by
// the zero-life statistics/penalty path and real GGOver Continue input.
func (sc *e2eScenario) CheckQuestPlayerDeaths(name string) {
	f := &e2eQuestDeathFixture{}
	for cycle := 1; cycle <= 3; cycle++ {
		label := fmt.Sprintf("%s cycle %d", name, cycle)
		sc.add(0, label+" place only player beside stock Necromancer", f.prepare)
		sc.addWhen(0, label+" wait for natural lethal damage", 15000, f.naturallyDead, func() {})
		sc.Screen(label + " dead")
		sc.Wait(20, label+" finish dead-state input delay")
		sc.add(0, label+" locate ordinary respawn input", f.respawnMouse)
		sc.Input(1, "", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true})
		sc.Input(1, "", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false})
		sc.addWhen(0, label+" wait for real-input respawn", 1200, f.respawned, func() {})
		sc.Screen(label + " respawned")
	}
}
