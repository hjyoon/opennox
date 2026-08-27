package opennox

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"hash"
	"image"
	"image/png"
	"io"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	"github.com/opennox/libs/datapath"
	"github.com/opennox/libs/ifs"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v2"

	"github.com/opennox/libs/client/keybind"
	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/log"
	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/platform"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

var (
	e2eLog = log.New("E2E")

	e2ePlay     = os.Getenv("NOX_E2E")
	e2eRecord   = os.Getenv("NOX_E2E_RECORD")
	e2eSlow     = os.Getenv("NOX_E2E_SLOW")
	e2eOverride = os.Getenv("NOX_E2E_OVERRIDE") == "true"
	e2eFailFast = os.Getenv("NOX_E2E_FAILFAST") != "false"
)

const e2eDefaultDelay = 15 * time.Millisecond

var e2e struct {
	recording bool
	path      string
	p         *platformE2E
	onInput   []func(ev seat.InputEvent)

	slow       time.Duration
	real       seat.Seat
	realMouse  image.Point
	realEnable bool

	done      chan<- struct{}
	steps     []e2eStep
	input     []seat.InputEvent
	recorded  []e2eRecordedEvent
	err       error
	checkSave *e2eCheckSave

	shopMerchant          *server.Object
	shopSession           *server.TradeSession
	monster               *server.Object
	monsterWorldTarget    *server.Object
	monsterWorldTargetHP  uint16
	groundItem            *server.Object
	groundItemTypeID      string
	groundItemBefore      int
	groundItemDropped     *server.Object
	engageItem            *server.Object
	engageItemTypeID      string
	engageModifier        *server.ModifierEff
	engageOwner           *server.Object
	engageOwnerMask       uint32
	engageOwnerMaskBefore uint32
	deadPlayer            *server.Object
	reloadPlayer          *server.Object
	reloadPos             types.Pointf
}

func e2eError(err error) {
	if e2eFailFast {
		panic(err)
	}
	e2eLog.Println(err)
	e2e.err = err
}

type e2eStep struct {
	name        string
	time        time.Duration
	fnc         func()
	ready       func() bool
	waited      time.Duration
	waitTimeout time.Duration
}

type e2eScenario struct {
	steps []e2eStep
	done  chan struct{}
}

func (sc *e2eScenario) Exec() {
	sc.done = make(chan struct{})
	e2eJobs <- sc
	<-sc.done
	sc.steps = nil
}

func (sc *e2eScenario) add(dt time.Duration, name string, fnc func()) {
	var last time.Duration
	if n := len(sc.steps); n != 0 {
		last = sc.steps[n-1].time
	}
	sc.steps = append(sc.steps, e2eStep{name: name, time: last + dt, fnc: fnc})
}

func (sc *e2eScenario) addWhen(dt time.Duration, name string, timeout time.Duration, ready func() bool, fnc func()) {
	var last time.Duration
	if n := len(sc.steps); n != 0 {
		last = sc.steps[n-1].time
	}
	sc.steps = append(sc.steps, e2eStep{
		name:        name,
		time:        last + dt,
		fnc:         fnc,
		ready:       ready,
		waitTimeout: timeout,
	})
}

func (sc *e2eScenario) Slow(dt time.Duration) {
	sc.add(0, "", func() {
		e2e.slow = dt
	})
}

func (sc *e2eScenario) Wait(dt time.Duration, name string) {
	if dt == 0 && name == "" {
		return
	}
	sc.add(dt, name, nil)
}

func (sc *e2eScenario) Input(dt time.Duration, name string, evs ...seat.InputEvent) {
	sc.add(dt, name, func() {
		e2eQueueInput(evs...)
	})
}

func (sc *e2eScenario) Quit(dt time.Duration) {
	sc.Input(dt, "", seat.WindowClosed)
	sc.Input(1, "", seat.WindowClosed)
	sc.add(1, "", func() {
		if e2e.err != nil {
			panic(e2e.err)
		}
	})
}

func (sc *e2eScenario) Move(x, y int, name string) {
	sc.Input(0, name, &seat.MouseMoveEvent{Pos: image.Point{X: x, Y: y}, Relative: false})
}

func (sc *e2eScenario) Click(pos image.Point, btn seat.MouseButton, name string) {
	sc.Input(0, name,
		&seat.MouseMoveEvent{Pos: pos, Relative: false},
		&seat.MouseButtonEvent{Button: btn, Pressed: true},
	)
	sc.Input(1, "", &seat.MouseButtonEvent{Button: btn, Pressed: false})
}

func (sc *e2eScenario) ClickSlow(pos image.Point, btn seat.MouseButton, name string) {
	sc.Input(0, name, &seat.MouseMoveEvent{Pos: pos, Relative: false})
	sc.Input(1, "", &seat.MouseButtonEvent{Button: btn, Pressed: true})
	sc.Input(1, "", &seat.MouseButtonEvent{Button: btn, Pressed: false})
}

func (sc *e2eScenario) Key(key keybind.Key, name string) {
	sc.Input(0, name, &seat.KeyboardEvent{Key: key, Pressed: true})
	sc.Input(1, "", &seat.KeyboardEvent{Key: key, Pressed: false})
}

func (sc *e2eScenario) ClickLeft(x, y int, name string) {
	sc.Click(image.Point{X: x, Y: y}, seat.MouseButtonLeft, name)
}

func (sc *e2eScenario) ClickSlowLeft(x, y int, name string) {
	sc.ClickSlow(image.Point{X: x, Y: y}, seat.MouseButtonLeft, name)
}

func e2eAngToPos(ang float64, dist int) image.Point {
	sz := image.Point{X: 1024, Y: 768}
	rad := (0.5 - ang) * math.Pi
	return image.Point{
		X: sz.X/2 + int(math.Cos(rad)*float64(dist)),
		Y: sz.Y/2 - int(math.Sin(rad)*float64(dist)),
	}
}

func (sc *e2eScenario) runStart(ang float64, dist int, name string) {
	sc.add(0, name, func() {
		pos := e2eAngToPos(ang, dist)
		e2eQueueInput(
			&seat.MouseMoveEvent{Pos: pos, Relative: false},
			&seat.MouseButtonEvent{Button: seat.MouseButtonRight, Pressed: true},
		)
	})
}

func (sc *e2eScenario) runDir(ang float64, dist int, name string) {
	sc.add(0, name, func() {
		pos := e2eAngToPos(ang, dist)
		e2eQueueInput(&seat.MouseMoveEvent{Pos: pos, Relative: false})
	})
}

func (sc *e2eScenario) runEnd(dt time.Duration) {
	sc.Input(dt, "", &seat.MouseButtonEvent{Button: seat.MouseButtonRight, Pressed: false})
	sc.Wait(5, "")
}

func (sc *e2eScenario) runFor(ang float64, dist int, dt time.Duration, name string) {
	sc.runStart(ang, dist, name)
	sc.runEnd(dt)
}

const (
	e2eWalkDist = 50
	e2eRunDist  = 200
)

func (sc *e2eScenario) WalkFor(ang float64, dt time.Duration, name string) {
	sc.runFor(ang, e2eWalkDist, dt, name)
}

func (sc *e2eScenario) WalkStart(ang float64, dt time.Duration, name string) {
	sc.runStart(ang, e2eWalkDist, name)
	sc.Wait(dt, "")
}

func (sc *e2eScenario) WalkDir(ang float64, dt time.Duration, name string) {
	sc.runDir(ang, e2eWalkDist, name)
	sc.Wait(dt, "")
}

func (sc *e2eScenario) WalkEnd() {
	sc.runEnd(0)
}

func (sc *e2eScenario) RunFor(ang float64, dt time.Duration, name string) {
	sc.runFor(ang, e2eRunDist, dt, name)
}

func (sc *e2eScenario) RunStart(ang float64, dt time.Duration, name string) {
	sc.runStart(ang, e2eRunDist, name)
	sc.Wait(dt, "")
}

func (sc *e2eScenario) RunDir(ang float64, dt time.Duration, name string) {
	sc.runDir(ang, e2eRunDist, name)
	sc.Wait(dt, "")
}

func (sc *e2eScenario) RunEnd() {
	sc.runEnd(0)
}

func (sc *e2eScenario) Melee(ang float64, name string) {
	sc.add(0, name, func() {
		pos := e2eAngToPos(ang, 20)
		e2eQueueInput(
			&seat.MouseMoveEvent{Pos: pos, Relative: false},
			&seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true},
		)
	})
	sc.Input(1, "", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false})
}

func (sc *e2eScenario) SpawnMonster(typeID string, offset image.Point, name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return noxServer.Players.HostUnit() != nil
	}, func() {
		if e2e.monster != nil {
			e2eError(fmt.Errorf("monster fixture is already active: %p", e2e.monster))
			return
		}
		typ := noxServer.Types.ByID(typeID)
		if typ == nil {
			e2eError(fmt.Errorf("unknown monster fixture type %q", typeID))
			return
		}
		if !typ.Class().Has(object.ClassMonster) {
			e2eError(fmt.Errorf("monster fixture type %q has class %v", typeID, typ.Class()))
			return
		}
		player := noxServer.Players.HostUnit()
		monster := noxServer.NewObjectByTypeID(typeID)
		if monster == nil {
			e2eError(fmt.Errorf("cannot create monster fixture %q", typeID))
			return
		}
		pos := player.Pos().Add(types.Ptf(float32(offset.X), float32(offset.Y)))
		noxServer.CreateObjectAt(monster, nil, pos)
		noxServer.ObjectsAddPending()
		if monster.UpdateData == nil || monster.Flags().Has(object.FlagDestroyed) {
			e2eError(fmt.Errorf("monster fixture %q was not initialized: update=%p flags=%v", typeID, monster.UpdateData, monster.Flags()))
			return
		}
		e2e.monster = monster
		e2e.monsterWorldTarget = nil
		e2e.monsterWorldTargetHP = 0
		if targetType := noxServer.Types.ByID("AirshipCaptain"); targetType != nil {
			bestDistance := math.MaxFloat64
			for candidate := noxServer.Objs.First(); candidate != nil; candidate = candidate.Next() {
				if int(candidate.TypeInd) != targetType.Ind() || candidate.HealthData == nil ||
					candidate.HealthData.Cur == 0 || candidate.Flags().HasAny(object.FlagDead|object.FlagDestroyed) {
					continue
				}
				delta := candidate.PosVec.Sub(monster.PosVec)
				distance := float64(delta.X*delta.X + delta.Y*delta.Y)
				if distance < bestDistance {
					bestDistance = distance
					e2e.monsterWorldTarget = candidate
					e2e.monsterWorldTargetHP = candidate.HealthData.Cur
				}
			}
		}
		update := monster.UpdateDataMonster()
		update.PreferredEnemy = player
		direction := server.DirFromVec(player.PosVec.Sub(monster.PosVec))
		monster.Direction1 = direction
		monster.Direction2 = direction
		actions := make([]string, 0, int(update.AIStackInd)+1)
		for i := 0; i <= int(update.AIStackInd) && i < len(update.AIStack); i++ {
			actions = append(actions, update.AIStack[i].Type().String())
		}
		var healthCur, healthMax, healthBase uint16
		if monster.HealthData != nil {
			healthCur = monster.HealthData.Cur
			healthMax = monster.HealthData.Max
			healthBase = monster.HealthData.Field2
		}
		var defStatus object.MonsterStatus
		var meleeRange, missileRange float32
		if update.MonsterDef != nil {
			defStatus = update.MonsterDef.StatusFlags92
			meleeRange = update.MonsterDef.MeleeAttackRange112
			missileRange = update.MonsterDef.MissileAttackRange212
		}
		e2eLog.Printf("MONSTER FIXTURE: type=%s object=%p netcode=%d player_pos=(%.3f,%.3f) monster_pos=(%.3f,%.3f)",
			typeID, monster, monster.NetCode, player.PosVec.X, player.PosVec.Y, monster.PosVec.X, monster.PosVec.Y)
		if player.UpdateData != nil && player.Class().Has(object.ClassPlayer) {
			playerUpdate := player.UpdateDataPlayer()
			playerInfo := playerUpdate.Player
			var armorEquip, weaponEquip, playerStatus uint32
			if playerInfo != nil {
				armorEquip = playerInfo.ArmorEquip
				weaponEquip = playerInfo.WeaponEquip
				playerStatus = playerInfo.Field3680
			}
			items := make([]string, 0)
			for item := player.InvFirstItem; item != nil; item = item.InvNextItem {
				itemType := fmt.Sprintf("#%d", item.TypeInd)
				if typ := noxServer.Types.ByInd(int(item.TypeInd)); typ != nil {
					itemType = typ.ID()
				}
				defend := 0
				if item.InitData != nil && item.Class().HasAny(object.ClassWeapon|object.ClassArmor|object.ClassWand) {
					for _, modifier := range item.InitDataModifier().Modifiers {
						if modifier != nil && modifier.Defend76.Fnc != nil {
							defend++
						}
					}
				}
				items = append(items, fmt.Sprintf("%s(flags=%#x,class=%#x,defend=%d)",
					itemType, uint32(item.ObjFlags), uint32(item.ObjClass), defend))
			}
			var playerHealthCur, playerHealthMax uint16
			if player.HealthData != nil {
				playerHealthCur, playerHealthMax = player.HealthData.Cur, player.HealthData.Max
			}
			e2eLog.Printf("PLAYER COMBAT STATE: object=%p flags=%#x state=%d field75=%#x field76=%#x status=%#x health=%d/%d armor_value=%g buffs=%#x weapon=%#x armor=%#x inventory=[%s]",
				player, uint32(player.ObjFlags), playerUpdate.State, playerUpdate.Field75, playerUpdate.Field76,
				playerStatus, playerHealthCur, playerHealthMax, math.Float32frombits(playerUpdate.Field57), player.Buffs,
				weaponEquip, armorEquip, strings.Join(items, ","))
		}
		e2eLog.Printf("MONSTER FIXTURE STATE: frame=%d flags=%#x class=%#x subclass=%#x stack=%d[%s] field137=%d status=%#x def_status=%#x aggression=%g sight=%g flee=%g retreat=%g speed=%g health=%d/%d base=%d melee_range=%g missile_range=%g current=%p preferred=%p seen=%d buffs=%#x weapon=%#x armor=%#x",
			noxServer.Frame(), uint32(monster.ObjFlags), uint32(monster.ObjClass), uint32(monster.ObjSubClass),
			update.AIStackInd, strings.Join(actions, ","), update.Field137, uint32(update.StatusFlags), uint32(defStatus),
			update.Aggression, update.SightRange, update.FleeRange, update.RetreatLevel, monster.SpeedBase,
			healthCur, healthMax, healthBase, meleeRange, missileRange, update.CurrentEnemy, update.PreferredEnemy,
			update.Field282_1, monster.Buffs, update.WeaponEquipFlags, update.ArmorEquipFlags)
		if target := e2e.monsterWorldTarget; target != nil {
			delta := target.PosVec.Sub(monster.PosVec)
			e2eLog.Printf("MONSTER WORLD TARGET: type=AirshipCaptain object=%p health=%d/%d distance=%.3f",
				target, target.HealthData.Cur, target.HealthData.Max,
				math.Hypot(float64(delta.X), float64(delta.Y)))
		}
	})
}

func (sc *e2eScenario) AssertMonsterEncounter(name string) {
	sc.add(0, name, func() {
		player := noxServer.Players.HostUnit()
		monster := e2e.monster
		if player == nil || monster == nil {
			e2eError(fmt.Errorf("monster encounter fixture is unavailable: player=%p monster=%p", player, monster))
			return
		}
		if monster.Flags().Has(object.FlagDestroyed) {
			e2eError(fmt.Errorf("monster encounter fixture was destroyed before assertion"))
			return
		}
		update := monster.UpdateDataMonster()
		engaged := update.CurrentEnemy == player || update.PreferredEnemy == player
		for i := 0; !engaged && i < int(update.Field282_1) && i < len(update.SeenEnemies); i++ {
			engaged = update.SeenEnemies[i] == player
		}
		if !engaged {
			e2eError(fmt.Errorf("monster did not acquire player: current=%p preferred=%p seen=%d player=%p", update.CurrentEnemy, update.PreferredEnemy, update.Field282_1, player))
			return
		}
		wireCode := noxServer.GetUnitNetCode(monster)
		if wireCode <= 0 || wireCode > int(^uint16(0)) {
			e2eError(fmt.Errorf("monster wire code = %#x", wireCode))
			return
		}
		drawable := noxClient.Objs.ByNetCode(uint16(wireCode))
		if drawable == nil || !drawable.Class().Has(object.ClassMonster) {
			e2eError(fmt.Errorf("monster client drawable = %p for wire code %#x", drawable, wireCode))
			return
		}
		delta := monster.Pos().Sub(player.Pos())
		distance := math.Hypot(float64(delta.X), float64(delta.Y))
		e2eLog.Printf("MONSTER ENCOUNTER: object=%p drawable=%p netcode=%d current=%p preferred=%p seen=%d distance=%.3f",
			monster, drawable, wireCode, update.CurrentEnemy, update.PreferredEnemy, update.Field282_1, distance)
	})
}

func (sc *e2eScenario) WaitMonsterDead(name string) {
	sc.addWhen(0, name, 2400, func() bool {
		monster := e2e.monster
		return monster != nil && monster.HealthData != nil && monster.HealthData.Cur == 0 &&
			monster.Flags().HasAny(object.FlagDead|object.FlagDestroyed)
	}, func() {
		player, playerUpdate := e2eHostPlayerUnit()
		monster := e2e.monster
		if monster == nil || monster.HealthData == nil {
			e2eError(fmt.Errorf("dead monster fixture is unavailable"))
			return
		}
		if player == nil || playerUpdate == nil || player.HealthData == nil || player.HealthData.Cur == 0 ||
			player.Flags().HasAny(object.FlagDead|object.FlagDestroyed) {
			e2eError(fmt.Errorf("player did not survive monster kill: player=%p update=%p", player, playerUpdate))
			return
		}
		wireCode := noxServer.GetUnitNetCode(monster)
		drawable := noxClient.Objs.ByNetCode(uint16(wireCode))
		e2eLog.Printf("MONSTER DEAD: object=%p drawable=%p netcode=%d frame=%d flags=%#x health=%d/%d player_health=%d/%d player_state=%d",
			monster, drawable, wireCode, noxServer.Frame(), uint32(monster.Flags()),
			monster.HealthData.Cur, monster.HealthData.Max, player.HealthData.Cur, player.HealthData.Max, playerUpdate.State)
	})
}

func e2eHostPlayerUnit() (*server.Object, *server.PlayerUpdateData) {
	unit := noxServer.Players.HostUnit()
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassPlayer) {
		return nil, nil
	}
	return unit, unit.UpdateDataPlayer()
}

func (sc *e2eScenario) WaitPlayerDead(name string) {
	sc.addWhen(0, name, 2400, func() bool {
		unit, update := e2eHostPlayerUnit()
		return unit != nil && update != nil && unit.HealthData != nil &&
			unit.HealthData.Cur == 0 && unit.Flags().Has(object.FlagDead) &&
			update.State == server.PlayerState4
	}, func() {
		unit, update := e2eHostPlayerUnit()
		if unit == nil || update == nil || unit.HealthData == nil {
			e2eError(fmt.Errorf("dead host player is unavailable"))
			return
		}
		drawable := noxClient.ClientPlayerUnit()
		wireCode := noxServer.GetUnitNetCode(unit)
		if drawable == nil || wireCode <= 0 || int(drawable.NetCode32) != wireCode {
			e2eError(fmt.Errorf("dead player client binding = drawable:%p client-code:%d server-code:%d", drawable, func() uint32 {
				if drawable == nil {
					return 0
				}
				return drawable.NetCode32
			}(), wireCode))
			return
		}
		e2e.deadPlayer = unit
		e2eLog.Printf("PLAYER DEAD: object=%p drawable=%p netcode=%d frame=%d state=%d flags=%#x health=%d/%d pos=(%.3f,%.3f)",
			unit, drawable, wireCode, noxServer.Frame(), update.State, uint32(unit.Flags()),
			unit.HealthData.Cur, unit.HealthData.Max, unit.PosVec.X, unit.PosVec.Y)
	})
}

func (sc *e2eScenario) AssertMonsterWorldDamage(name string) {
	sc.add(0, name, func() {
		attacker := e2e.monster
		target := e2e.monsterWorldTarget
		if attacker == nil || attacker.UpdateData == nil || target == nil || target.UpdateData == nil || target.HealthData == nil {
			e2eError(fmt.Errorf("monster world-damage fixture is unavailable: attacker=%p target=%p", attacker, target))
			return
		}
		before := e2e.monsterWorldTargetHP
		after := target.HealthData.Cur
		if before <= after {
			e2eError(fmt.Errorf("AirshipCaptain health did not decrease: before=%d after=%d", before, after))
			return
		}
		delta := before - after
		if delta%3 != 0 {
			e2eError(fmt.Errorf("AirshipCaptain damage = %d, want a positive multiple of Spider BITE damage 3", delta))
			return
		}
		update := target.UpdateDataMonster()
		if target.Obj130 != attacker || target.Field131 != uint32(object.DamageBite) || target.Frame134 == 0 {
			e2eError(fmt.Errorf("AirshipCaptain attribution = source:%p type:%d frame:%d, want Spider/%d/nonzero",
				target.Obj130, target.Field131, target.Frame134, object.DamageBite))
			return
		}
		// MonStatusInjured is consumed by the target monster's AI update, so it
		// is deliberately asserted by the immediate unit test instead of here.
		// Field546/Field547 are the persistent hit latch available after the
		// Spider has also completed the asynchronous player-kill sequence.
		if update.Field546 != uint32(object.DamageBite) || update.Field547 != 2 {
			e2eError(fmt.Errorf("AirshipCaptain hit state = status:%#x type:%d latch:%d",
				uint32(update.StatusFlags), update.Field546, update.Field547))
			return
		}
		// The attacker's Field130 combat timestamp is likewise cleared by the
		// AI after one second. Its exact immediate value and mutation order are
		// covered by TestDefaultDamageWorld4E0B30SpiderBitesAirshipCaptain.
		e2eLog.Printf("MONSTER WORLD DAMAGE: attacker=Spider(%p) target=AirshipCaptain(%p) health=%d->%d damage=%d hits=%d type=%d frame=%d status=%#x",
			attacker, target, before, after, delta, delta/3, target.Field131, target.Frame134, uint32(update.StatusFlags))
	})
}

func (sc *e2eScenario) WaitDeathScreen(name string) {
	sc.addWhen(0, name, 2400, func() bool {
		return e2e.deadPlayer != nil && legacy.Get_dword_5d4594_831260() != 0 &&
			legacy.Get_dword_5d4594_831220() == 0 &&
			legacy.Get_nox_gameDisableMapDraw_5d4594_2650672() != 0
	}, func() {
		e2eLog.Printf("PLAYER DEATH SCREEN: briefing=%d chapter=%d map_draw_disabled=%d",
			legacy.Get_dword_5d4594_831260(), legacy.Get_dword_5d4594_831220(),
			legacy.Get_nox_gameDisableMapDraw_5d4594_2650672())
	})
}

func (sc *e2eScenario) ClickSaveLoad(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		win := dword_5d4594_1082856
		if win == nil || win.GetFlags().IsHidden() {
			return false
		}
		load := win.ChildByID(502)
		return load != nil && !load.GetFlags().IsHidden() && load.GetFlags().IsEnabled()
	}, func() {
		win := dword_5d4594_1082856
		if win == nil {
			e2eError(fmt.Errorf("save/load window is unavailable"))
			return
		}
		load := win.ChildByID(502)
		if load == nil || load.GetFlags().IsHidden() || !load.GetFlags().IsEnabled() {
			e2eError(fmt.Errorf("save/load load control is unavailable"))
			return
		}
		size := load.Size()
		pos := load.GlobalPos().Add(image.Pt(size.X/2, size.Y/2))
		selected := int32(-1)
		if list := dword_5d4594_1082864; list != nil && list.WidgetData != nil {
			selected = *(*int32)(unsafe.Add(list.WidgetData, 48))
		}
		e2eLog.Printf("SAVE/LOAD CLICK: load point=%v window=%v size=%v selected=%d", pos, win.GlobalPos(), win.Size(), selected)
		e2eQueueInput(&seat.MouseMoveEvent{Pos: pos, Relative: false})
	})
	sc.Input(1, "", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true})
	sc.Input(1, "", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false})
}

func (sc *e2eScenario) ClickDialogYes(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		dialog := nox_gui_curDialog_830224
		if dialog != nil && !dialog.GetFlags().IsHidden() {
			yes := dialog.ChildByID(guiDialogYesID)
			if yes != nil && !yes.GetFlags().IsHidden() && yes.GetFlags().IsEnabled() {
				return true
			}
		}
		unit, update := e2eHostPlayerUnit()
		return unit != nil && update != nil && unit.HealthData != nil && unit.HealthData.Cur > 0 &&
			!unit.Flags().HasAny(object.FlagDead|object.FlagDestroyed) &&
			update.State != server.PlayerState3 && update.State != server.PlayerState4
	}, func() {
		dialog := nox_gui_curDialog_830224
		if dialog == nil || dialog.GetFlags().IsHidden() {
			e2eLog.Printf("CONFIRMATION: autosave reload proceeded without a dialog")
			return
		}
		yes := dialog.ChildByID(guiDialogYesID)
		if yes == nil || yes.GetFlags().IsHidden() || !yes.GetFlags().IsEnabled() {
			e2eError(fmt.Errorf("confirmation dialog yes control is unavailable"))
			return
		}
		size := yes.Size()
		pos := yes.GlobalPos().Add(image.Pt(size.X/2, size.Y/2))
		e2eLog.Printf("CONFIRMATION CLICK: yes point=%v dialog=%v size=%v", pos, dialog.GlobalPos(), dialog.Size())
		e2eQueueInput(&seat.MouseMoveEvent{Pos: pos, Relative: false})
	})
	sc.Input(1, "", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true})
	sc.Input(1, "", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false})
}

func (sc *e2eScenario) WaitPlayerReloaded(name string) {
	sc.addWhen(0, name, 3600, func() bool {
		unit, update := e2eHostPlayerUnit()
		return unit != nil && update != nil && unit.HealthData != nil && unit.HealthData.Cur > 0 &&
			!unit.Flags().HasAny(object.FlagDead|object.FlagDestroyed) &&
			update.State != server.PlayerState3 && update.State != server.PlayerState4 &&
			noxClient.ClientPlayerUnit() != nil && nox_client_isConnected()
	}, func() {
		unit, update := e2eHostPlayerUnit()
		if unit == nil || update == nil || unit.HealthData == nil {
			e2eError(fmt.Errorf("reloaded host player is unavailable"))
			return
		}
		drawable := noxClient.ClientPlayerUnit()
		wireCode := noxServer.GetUnitNetCode(unit)
		if drawable == nil || wireCode <= 0 || int(drawable.NetCode32) != wireCode {
			e2eError(fmt.Errorf("reloaded player client binding = drawable:%p server-code:%d", drawable, wireCode))
			return
		}
		e2e.reloadPlayer = unit
		e2e.reloadPos = unit.PosVec
		e2eLog.Printf("PLAYER RELOADED: previous=%p object=%p drawable=%p netcode=%d frame=%d state=%d flags=%#x health=%d/%d pos=(%.3f,%.3f)",
			e2e.deadPlayer, unit, drawable, wireCode, noxServer.Frame(), update.State, uint32(unit.Flags()),
			unit.HealthData.Cur, unit.HealthData.Max, unit.PosVec.X, unit.PosVec.Y)
	})
}

func (sc *e2eScenario) AssertPlayerMovedAfterReload(name string) {
	sc.add(0, name, func() {
		unit, update := e2eHostPlayerUnit()
		if unit == nil || update == nil || e2e.reloadPlayer == nil {
			e2eError(fmt.Errorf("reloaded player movement fixture is unavailable"))
			return
		}
		if unit.HealthData == nil || unit.HealthData.Cur == 0 || unit.Flags().HasAny(object.FlagDead|object.FlagDestroyed) {
			e2eError(fmt.Errorf("reloaded player cannot be controlled: flags=%#x health=%v", uint32(unit.Flags()), unit.HealthData))
			return
		}
		delta := unit.PosVec.Sub(e2e.reloadPos)
		distance := math.Hypot(float64(delta.X), float64(delta.Y))
		if distance < 5 {
			e2eError(fmt.Errorf("reloaded player moved %.3f units, want at least 5", distance))
			return
		}
		e2eLog.Printf("PLAYER RELOAD CONTROL: object=%p state=%d from=(%.3f,%.3f) to=(%.3f,%.3f) distance=%.3f",
			unit, update.State, e2e.reloadPos.X, e2e.reloadPos.Y, unit.PosVec.X, unit.PosVec.Y, distance)
	})
}

func e2eInventoryItemCount(typeID string) (int, error) {
	typ := noxServer.Types.ByID(typeID)
	if typ == nil {
		return 0, fmt.Errorf("unknown inventory fixture type %q", typeID)
	}
	unit := noxServer.Players.HostUnit()
	if unit == nil {
		return 0, fmt.Errorf("host unit is unavailable for inventory fixture %q", typeID)
	}
	count := 0
	for item := unit.InvFirstItem; item != nil; item = item.InvNextItem {
		if int(item.TypeInd) == typ.Ind() {
			count++
		}
	}
	return count, nil
}

func e2eInventoryItem(typeID string) (*server.Object, int, error) {
	typ := noxServer.Types.ByID(typeID)
	if typ == nil {
		return nil, 0, fmt.Errorf("unknown inventory fixture type %q", typeID)
	}
	unit := noxServer.Players.HostUnit()
	if unit == nil {
		return nil, typ.Ind(), fmt.Errorf("host unit is unavailable for inventory fixture %q", typeID)
	}
	for item := unit.InvFirstItem; item != nil; item = item.InvNextItem {
		if int(item.TypeInd) == typ.Ind() {
			return item, typ.Ind(), nil
		}
	}
	return nil, typ.Ind(), fmt.Errorf("inventory item %q is unavailable", typeID)
}

func (sc *e2eScenario) GrantInventoryItems(typeID string, count int, name string) {
	sc.add(0, name, func() {
		if count <= 0 {
			e2eError(fmt.Errorf("inventory fixture count for %q must be positive, got %d", typeID, count))
			return
		}
		before, err := e2eInventoryItemCount(typeID)
		if err != nil {
			e2eError(err)
			return
		}
		unit := noxServer.Players.HostUnit()
		for i := 0; i < count; i++ {
			if item := legacy.Nox_xxx_playerRespawnItem_4EF750(unit, typeID, nil, 1, 0); item == nil {
				e2eError(fmt.Errorf("failed to grant inventory fixture %q at index %d", typeID, i))
				return
			}
		}
		after, err := e2eInventoryItemCount(typeID)
		if err != nil {
			e2eError(err)
			return
		}
		if want := before + count; after != want {
			e2eError(fmt.Errorf("inventory fixture %q count = %d, want %d", typeID, after, want))
			return
		}
		e2eLog.Printf("INVENTORY FIXTURE: item=%s before=%d granted=%d after=%d", typeID, before, count, after)
	})
}

func (sc *e2eScenario) GrantEngageItem(typeID, modifierName string, mask uint32, name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return noxServer.Players.HostUnit() != nil
	}, func() {
		if e2e.engageItem != nil {
			e2eError(fmt.Errorf("engage-item fixture is already active: %p", e2e.engageItem))
			return
		}
		if mask == 0 {
			e2eError(fmt.Errorf("engage-item fixture mask must be nonzero"))
			return
		}
		beforeCount, err := e2eInventoryItemCount(typeID)
		if err != nil {
			e2eError(err)
			return
		}
		if beforeCount != 0 {
			e2eError(fmt.Errorf("engage-item fixture type %q is not unique: inventory count = %d", typeID, beforeCount))
			return
		}
		modifierID := noxServer.Modif.Nox_xxx_modifGetIdByName413290(modifierName)
		modifier := noxServer.Modif.Nox_xxx_modifGetDescById413330(modifierID)
		if modifier == nil || modifier.Name() != modifierName || modifier.Engage112 == nil {
			e2eError(fmt.Errorf("engage modifier %q is unavailable: id=%d modifier=%p callback=%p", modifierName, modifierID, modifier, func() unsafe.Pointer {
				if modifier == nil {
					return nil
				}
				return modifier.Engage112
			}()))
			return
		}
		owner := noxServer.Players.HostUnit()
		beforeMask := owner.Field110
		if beforeMask&mask != 0 {
			e2eError(fmt.Errorf("engage owner mask %#x already contains fixture mask %#x", beforeMask, mask))
			return
		}
		attrs := &server.ModifierInitData{Modifiers: [4]*server.ModifierEff{nil, nil, modifier, nil}}
		item := legacy.Nox_xxx_playerRespawnItem_4EF750(owner, typeID, attrs, 1, 0)
		if item == nil || !owner.HasItem(item) || item.InvHolder != owner || item.Flags().Has(object.FlagEquipped) {
			e2eError(fmt.Errorf("engage item %q was not placed unequipped: item=%p holder=%p flags=%v", typeID, item, func() *server.Object {
				if item == nil {
					return nil
				}
				return item.InvHolder
			}(), func() object.Flags {
				if item == nil {
					return 0
				}
				return item.Flags()
			}()))
			return
		}
		data := item.InitDataModifier()
		if data == nil || data.Modifiers[2] != modifier || data.Modifiers[3] != nil {
			e2eError(fmt.Errorf("engage item %q modifier state = %p/%p/%p, want slot2 %p and nil slot3", typeID, data, func() *server.ModifierEff {
				if data == nil {
					return nil
				}
				return data.Modifiers[2]
			}(), func() *server.ModifierEff {
				if data == nil {
					return nil
				}
				return data.Modifiers[3]
			}(), modifier))
			return
		}
		e2e.engageItem = item
		e2e.engageItemTypeID = typeID
		e2e.engageModifier = modifier
		e2e.engageOwner = owner
		e2e.engageOwnerMask = mask
		e2e.engageOwnerMaskBefore = beforeMask
		e2eLog.Printf("ENGAGE ITEM GRANTED: item=%s object=%p modifier=%s modifier_object=%p callback=%p owner=%p mask_before=%#x mask_expected=%#x",
			typeID, item, modifierName, modifier, modifier.Engage112, owner, beforeMask, beforeMask|mask)
	})
}

func (sc *e2eScenario) AssertEngageItemEquipped(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return e2e.engageItem != nil && e2e.engageItem.Flags().Has(object.FlagEquipped)
	}, func() {
		item := e2e.engageItem
		owner := e2e.engageOwner
		if item == nil || owner == nil || !owner.HasItem(item) || item.InvHolder != owner || !item.Flags().Has(object.FlagEquipped) {
			e2eError(fmt.Errorf("engage item was not equipped: item=%p owner=%p holder=%p flags=%v", item, owner, func() *server.Object {
				if item == nil {
					return nil
				}
				return item.InvHolder
			}(), func() object.Flags {
				if item == nil {
					return 0
				}
				return item.Flags()
			}()))
			return
		}
		data := item.InitDataModifier()
		if data == nil || data.Modifiers[2] != e2e.engageModifier || data.Modifiers[3] != nil {
			e2eError(fmt.Errorf("equipped engage item modifier identity changed: data=%p slot2=%p slot3=%p want=%p", data, func() *server.ModifierEff {
				if data == nil {
					return nil
				}
				return data.Modifiers[2]
			}(), func() *server.ModifierEff {
				if data == nil {
					return nil
				}
				return data.Modifiers[3]
			}(), e2e.engageModifier))
			return
		}
		update := owner.UpdateDataPlayer()
		if update == nil || update.EquippedWeapon != item {
			e2eError(fmt.Errorf("equipped weapon pointer = %p, want engage item %p", func() *server.Object {
				if update == nil {
					return nil
				}
				return update.EquippedWeapon
			}(), item))
			return
		}
		wantMask := e2e.engageOwnerMaskBefore | e2e.engageOwnerMask
		if owner.Field110 != wantMask {
			e2eError(fmt.Errorf("engage owner mask = %#x, want %#x after native callback", owner.Field110, wantMask))
			return
		}
		e2eLog.Printf("ENGAGE ITEM EQUIPPED: item=%s object=%p modifier=%s modifier_object=%p callback=%p owner=%p equipped_weapon=%p mask=%#x",
			e2e.engageItemTypeID, item, e2e.engageModifier.Name(), e2e.engageModifier, e2e.engageModifier.Engage112, owner, update.EquippedWeapon, owner.Field110)
	})
}

func (sc *e2eScenario) AssertEngageItemDequipped(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return e2e.engageItem != nil && !e2e.engageItem.Flags().Has(object.FlagEquipped)
	}, func() {
		item := e2e.engageItem
		owner := e2e.engageOwner
		if item == nil || owner == nil || !owner.HasItem(item) || item.InvHolder != owner || item.Flags().Has(object.FlagEquipped) {
			e2eError(fmt.Errorf("engage item was not dequipped: item=%p owner=%p holder=%p flags=%v", item, owner, func() *server.Object {
				if item == nil {
					return nil
				}
				return item.InvHolder
			}(), func() object.Flags {
				if item == nil {
					return 0
				}
				return item.Flags()
			}()))
			return
		}
		data := item.InitDataModifier()
		if data == nil || data.Modifiers[2] != e2e.engageModifier || data.Modifiers[3] != nil {
			e2eError(fmt.Errorf("dequipped engage item modifier identity changed: data=%p slot2=%p slot3=%p want=%p", data, func() *server.ModifierEff {
				if data == nil {
					return nil
				}
				return data.Modifiers[2]
			}(), func() *server.ModifierEff {
				if data == nil {
					return nil
				}
				return data.Modifiers[3]
			}(), e2e.engageModifier))
			return
		}
		update := owner.UpdateDataPlayer()
		if update == nil || update.EquippedWeapon == item {
			e2eError(fmt.Errorf("equipped weapon pointer still references dequipped item %p: update=%p equipped=%p", item, update, func() *server.Object {
				if update == nil {
					return nil
				}
				return update.EquippedWeapon
			}()))
			return
		}
		if owner.Field110 != e2e.engageOwnerMaskBefore {
			e2eError(fmt.Errorf("disengage owner mask = %#x, want original %#x after native callback", owner.Field110, e2e.engageOwnerMaskBefore))
			return
		}
		e2eLog.Printf("ENGAGE ITEM DEQUIPPED: item=%s object=%p modifier=%s modifier_object=%p callback=%p owner=%p equipped_weapon=%p mask=%#x",
			e2e.engageItemTypeID, item, e2e.engageModifier.Name(), e2e.engageModifier, e2e.engageModifier.Disengage116, owner, update.EquippedWeapon, owner.Field110)
	})
}

func (sc *e2eScenario) SpawnGroundItem(typeID string, offset image.Point, name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return noxServer.Players.HostUnit() != nil
	}, func() {
		if e2e.groundItem != nil {
			e2eError(fmt.Errorf("ground-item fixture is already active: %p", e2e.groundItem))
			return
		}
		typ := noxServer.Types.ByID(typeID)
		if typ == nil {
			e2eError(fmt.Errorf("unknown ground-item fixture type %q", typeID))
			return
		}
		before, err := e2eInventoryItemCount(typeID)
		if err != nil {
			e2eError(err)
			return
		}
		player := noxServer.Players.HostUnit()
		item := noxServer.NewObjectByTypeID(typeID)
		if item == nil || item.Pickup.Ptr == nil {
			e2eError(fmt.Errorf("cannot create pickup-enabled ground item %q: item=%p", typeID, item))
			return
		}
		pos := player.Pos().Add(types.Ptf(float32(offset.X), float32(offset.Y)))
		noxServer.CreateObjectAt(item, nil, pos)
		noxServer.ObjectsAddPending()
		wireCode := noxServer.GetUnitNetCode(item)
		if wireCode <= 0 || wireCode > int(^uint16(0)) || item.InvHolder != nil ||
			!item.Flags().Has(object.FlagActive) || item.Flags().Has(object.FlagDestroyed) {
			e2eError(fmt.Errorf("ground item %q was not initialized: item=%p wire=%#x holder=%p flags=%v", typeID, item, wireCode, item.InvHolder, item.Flags()))
			return
		}
		e2e.groundItem = item
		e2e.groundItemTypeID = typeID
		e2e.groundItemBefore = before
		e2e.groundItemDropped = nil
		e2eLog.Printf("GROUND ITEM SPAWNED: item=%s object=%p netcode=%d wire=%#x before=%d player_pos=(%.3f,%.3f) item_pos=(%.3f,%.3f)",
			typeID, item, item.NetCode, wireCode, before, player.PosVec.X, player.PosVec.Y, item.PosVec.X, item.PosVec.Y)
	})
}

func (sc *e2eScenario) PickupGroundItem(name string) {
	sc.addWhen(0, name+" visible", 1200, func() bool {
		item := e2e.groundItem
		if item == nil {
			return false
		}
		wireCode := noxServer.GetUnitNetCode(item)
		return wireCode > 0 && wireCode <= int(^uint16(0)) && noxClient.Objs.ByNetCode(uint16(wireCode)) != nil
	}, func() {
		item := e2e.groundItem
		wireCode := noxServer.GetUnitNetCode(item)
		drawable := noxClient.Objs.ByNetCode(uint16(wireCode))
		pos := noxClient.Viewport().ToScreenPos(drawable.Pos())
		if !pos.In(noxClient.Viewport().Screen) {
			e2eError(fmt.Errorf("ground item %q is outside the viewport: world=%v screen=%v viewport=%v", e2e.groundItemTypeID, drawable.Pos(), pos, noxClient.Viewport().Screen))
			return
		}
		e2eLog.Printf("GROUND ITEM MOUSE: item=%s object=%p drawable=%p wire=%#x world=%v screen=%v", e2e.groundItemTypeID, item, drawable, wireCode, drawable.Pos(), pos)
		e2eQueueInput(&seat.MouseMoveEvent{Pos: pos, Relative: false})
	})
	sc.addWhen(1, name+" cursor", 600, func() bool {
		cursor := noxClient.Nox_client_getCursorType()
		return cursor == gui.CursorPickup || cursor == gui.CursorCaution
	}, func() {
		e2eLog.Printf("GROUND ITEM CURSOR: item=%s cursor=%d", e2e.groundItemTypeID, noxClient.Nox_client_getCursorType())
		e2eQueueInput(&seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true})
	})
	sc.Input(1, "", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false})
}

func (sc *e2eScenario) AssertGroundItemPicked(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		item := e2e.groundItem
		player := noxServer.Players.HostUnit()
		if item == nil || player == nil || !player.HasItem(item) || item.InvHolder != player || item.ObjOwner != player || item.Flags().Has(object.FlagActive) {
			return false
		}
		count, err := e2eInventoryItemCount(e2e.groundItemTypeID)
		if err != nil || count != e2e.groundItemBefore+1 {
			return false
		}
		typ := noxServer.Types.ByID(e2e.groundItemTypeID)
		found, clientCount, _, _ := legacy.Nox_client_inventoryItemState(uint32(typ.Ind()))
		return found && clientCount == uint32(count)
	}, func() {
		item := e2e.groundItem
		count, _ := e2eInventoryItemCount(e2e.groundItemTypeID)
		typ := noxServer.Types.ByID(e2e.groundItemTypeID)
		_, clientCount, _, _ := legacy.Nox_client_inventoryItemState(uint32(typ.Ind()))
		e2eLog.Printf("GROUND ITEM PICKED: item=%s object=%p netcode=%d holder=%p owner=%p active=%t server_count=%d client_count=%d",
			e2e.groundItemTypeID, item, item.NetCode, item.InvHolder, item.ObjOwner, item.Flags().Has(object.FlagActive), count, clientCount)
	})
}

func (sc *e2eScenario) DragInventoryItemOut(typeID, name string) {
	sc.addWhen(0, name+" item visible", 1200, func() bool {
		typ := noxServer.Types.ByID(typeID)
		if typ == nil {
			return false
		}
		found, column, row, _ := legacy.Nox_client_inventoryItemLocation(uint32(typ.Ind()))
		if !found {
			return false
		}
		offset := legacy.Nox_client_inventoryAnimationOffset()
		pos := image.Pt(314+50*column+25, 13+50*row-offset+25)
		return pos.X >= 314 && pos.X < 514 && pos.Y >= 13 && pos.Y < 213
	}, func() {
		typ := noxServer.Types.ByID(typeID)
		_, column, row, netCode := legacy.Nox_client_inventoryItemLocation(uint32(typ.Ind()))
		offset := legacy.Nox_client_inventoryAnimationOffset()
		pos := image.Pt(314+50*column+25, 13+50*row-offset+25)
		e2eLog.Printf("INVENTORY DROP MOUSE: item=%s column=%d row=%d netcode=%d offset=%d start=%v", typeID, column, row, netCode, offset, pos)
		e2eQueueInput(&seat.MouseMoveEvent{Pos: pos, Relative: false})
	})
	sc.Input(2, "", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true})
	sc.Input(4, "", &seat.MouseMoveEvent{Pos: image.Pt(100, 400), Relative: false})
	sc.addWhen(1, name+" dragging", 600, legacy.Nox_client_inventoryHasDragged, func() {
		e2eLog.Printf("INVENTORY DROP DRAGGING: item=%s destination=%v", typeID, image.Pt(100, 400))
	})
	sc.Input(1, "", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false})
}

func (sc *e2eScenario) AssertGroundItemDropped(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		item := e2e.groundItem
		player := noxServer.Players.HostUnit()
		if item == nil || player == nil || player.HasItem(item) || item.InvHolder != nil ||
			!item.Flags().Has(object.FlagActive) || item.Flags().Has(object.FlagDestroyed) {
			return false
		}
		count, err := e2eInventoryItemCount(e2e.groundItemTypeID)
		if err != nil || count != e2e.groundItemBefore {
			return false
		}
		typ := noxServer.Types.ByID(e2e.groundItemTypeID)
		found, clientCount, _, _ := legacy.Nox_client_inventoryItemState(uint32(typ.Ind()))
		if clientCount != uint32(count) || found != (count != 0) {
			return false
		}
		wireCode := noxServer.GetUnitNetCode(item)
		if wireCode <= 0 || wireCode > int(^uint16(0)) {
			return false
		}
		drawable := noxClient.Objs.ByNetCode(uint16(wireCode))
		return drawable != nil && drawable.Class().Has(object.ClassPickup)
	}, func() {
		item := e2e.groundItem
		player := noxServer.Players.HostUnit()
		wireCode := noxServer.GetUnitNetCode(item)
		drawable := noxClient.Objs.ByNetCode(uint16(wireCode))
		count, _ := e2eInventoryItemCount(e2e.groundItemTypeID)
		delta := item.PosVec.Sub(player.PosVec)
		distance := math.Hypot(float64(delta.X), float64(delta.Y))
		if distance > 200 {
			e2eError(fmt.Errorf("dropped ground item %q is %.3f units from player", e2e.groundItemTypeID, distance))
			return
		}
		e2e.groundItemDropped = item
		e2eLog.Printf("GROUND ITEM DROPPED: item=%s object=%p drawable=%p netcode=%d wire=%#x holder=%p active=%t server_count=%d distance=%.3f pos=(%.3f,%.3f)",
			e2e.groundItemTypeID, item, drawable, item.NetCode, wireCode, item.InvHolder, item.Flags().Has(object.FlagActive), count, distance, item.PosVec.X, item.PosVec.Y)
	})
}

func (sc *e2eScenario) AssertInventoryItemCount(typeID string, want int, name string) {
	sc.add(0, name, func() {
		got, err := e2eInventoryItemCount(typeID)
		if err != nil {
			e2eError(err)
			return
		}
		if got != want {
			e2eError(fmt.Errorf("inventory %q count = %d, want %d", typeID, got, want))
			return
		}
		typ := noxServer.Types.ByID(typeID)
		found, clientCount, _, _ := legacy.Nox_client_inventoryItemState(uint32(typ.Ind()))
		if clientCount != uint32(want) || found != (want != 0) {
			e2eError(fmt.Errorf("client inventory %q = found:%t count:%d, want found:%t count:%d", typeID, found, clientCount, want != 0, want))
			return
		}
		e2eLog.Printf("INVENTORY COUNT: item=%s server=%d client=%d", typeID, got, clientCount)
	})
}

func (sc *e2eScenario) ClickInventoryItem(typeID, name string) {
	sc.add(0, name, func() {
		typ := noxServer.Types.ByID(typeID)
		if typ == nil {
			e2eError(fmt.Errorf("unknown client inventory type %q", typeID))
			return
		}
		found, column, row, netCode := legacy.Nox_client_inventoryItemLocation(uint32(typ.Ind()))
		if !found {
			e2eError(fmt.Errorf("client inventory item %q has no visible cell", typeID))
			return
		}
		offset := legacy.Nox_client_inventoryAnimationOffset()
		pos := image.Point{
			X: 314 + 50*column + 25,
			Y: 13 + 50*row - offset + 25,
		}
		if pos.X < 314 || pos.X >= 514 || pos.Y < 13 || pos.Y >= 213 {
			e2eError(fmt.Errorf("client inventory item %q cell column:%d row:%d is outside the visible tray at offset %d (point %v)", typeID, column, row, offset, pos))
			return
		}
		e2eLog.Printf("INVENTORY CLICK: item=%s column=%d row=%d netcode=%d offset=%d point=%v cursor_mode=%d shop_mode=%d", typeID, column, row, netCode, offset, pos, legacy.Sub_4675B0(), legacy.Sub_479590())
		e2eQueueInput(
			&seat.MouseMoveEvent{Pos: pos, Relative: false},
			&seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true},
		)
	})
	sc.Input(1, "", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false})
}

func (sc *e2eScenario) ClickItemAmountAccept(offset image.Point, name string) {
	sc.add(0, name, func() {
		dialog := legacy.Get_nox_gui_itemAmount_dialog_1319228()
		if dialog == nil || dialog.GetFlags().IsHidden() {
			e2eError(fmt.Errorf("item amount dialog is not active"))
			return
		}
		accept := dialog.ChildByID(3604)
		if accept == nil {
			e2eError(fmt.Errorf("item amount accept control is unavailable"))
			return
		}
		size := accept.Size()
		pos := accept.GlobalPos().Add(image.Pt(size.X/2, size.Y/2)).Add(offset)
		e2eLog.Printf("ITEM AMOUNT CLICK: accept point=%v offset=%v dialog=%v size=%v", pos, offset, dialog.GlobalPos(), dialog.Size())
		e2eQueueInput(
			&seat.MouseMoveEvent{Pos: pos, Relative: false},
			&seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true},
		)
	})
	sc.Input(1, "", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false})
}

func (sc *e2eScenario) ClickNPCDialogDone(name string) {
	sc.add(0, name, func() {
		dialog := legacy.Get_dword_5d4594_1123524()
		if dialog == nil || dialog.GetFlags().IsHidden() {
			e2eError(fmt.Errorf("NPC dialog is not active"))
			return
		}
		done := dialog.ChildByID(3906)
		if done == nil || done.GetFlags().IsHidden() || !done.GetFlags().IsEnabled() {
			e2eError(fmt.Errorf("NPC dialog done control is unavailable"))
			return
		}
		size := done.Size()
		pos := done.GlobalPos().Add(image.Pt(size.X/2, size.Y/2))
		e2eLog.Printf("NPC DIALOG CLICK: done point=%v dialog=%v size=%v", pos, dialog.GlobalPos(), dialog.Size())
		e2eQueueInput(&seat.MouseMoveEvent{Pos: pos, Relative: false})
	})
	sc.Input(1, "", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true})
	sc.Input(1, "", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false})
}

func (sc *e2eScenario) DamageInventoryItem(typeID string, health int, name string) {
	sc.add(0, name, func() {
		item, _, err := e2eInventoryItem(typeID)
		if err != nil {
			e2eError(err)
			return
		}
		if item.HealthData == nil || item.HealthData.Max == 0 {
			e2eError(fmt.Errorf("inventory item %q has no durability", typeID))
			return
		}
		if health <= 0 || health >= int(item.HealthData.Max) || health > int(^uint16(0)) {
			e2eError(fmt.Errorf("damage health for %q must be in 1..%d, got %d", typeID, item.HealthData.Max-1, health))
			return
		}
		before := item.HealthData.Cur
		legacy.Nox_xxx_unitSetHP_4E4560(item, uint16(health))
		packet := server.BuildShopItemHealthPacket4D87A0(item)
		player := noxServer.Players.HostUnit().UpdateDataPlayer().Player
		noxServer.NetSendPacketXxx1(player.Index(), packet[:], nil, 0)
		e2eLog.Printf("INVENTORY DAMAGE FIXTURE: item=%s netcode=%d before=%d after=%d max=%d worth=%d", typeID, item.NetCode, before, item.HealthData.Cur, item.HealthData.Max, item.Worth)
	})
}

func (sc *e2eScenario) AssertInventoryItemHealth(typeID string, health int, full bool, name string) {
	sc.add(0, name, func() {
		item, typeInd, err := e2eInventoryItem(typeID)
		if err != nil {
			e2eError(err)
			return
		}
		if item.HealthData == nil || item.HealthData.Max == 0 {
			e2eError(fmt.Errorf("inventory item %q has no durability", typeID))
			return
		}
		want := uint16(health)
		if full {
			want = item.HealthData.Max
		}
		found, clientCount, clientCurrent, clientMaximum := legacy.Nox_client_inventoryItemState(uint32(typeInd))
		if item.HealthData.Cur != want || !found || clientCount == 0 || clientCurrent != want || clientMaximum != item.HealthData.Max {
			e2eError(fmt.Errorf("inventory health %q = server:%d/%d client:%d/%d found:%t count:%d, want %d/%d", typeID, item.HealthData.Cur, item.HealthData.Max, clientCurrent, clientMaximum, found, clientCount, want, item.HealthData.Max))
			return
		}
		e2eLog.Printf("INVENTORY HEALTH: item=%s server=%d/%d client=%d/%d", typeID, item.HealthData.Cur, item.HealthData.Max, clientCurrent, clientMaximum)
	})
}

func (sc *e2eScenario) SetPlayerGold(gold int, name string) {
	sc.add(0, name, func() {
		if gold < 0 || uint64(gold) > uint64(^uint32(0)) {
			e2eError(fmt.Errorf("player gold must fit uint32, got %d", gold))
			return
		}
		unit := noxServer.Players.HostUnit()
		if unit == nil {
			e2eError(fmt.Errorf("host unit is unavailable for gold fixture"))
			return
		}
		player := unit.UpdateDataPlayer().Player
		if player == nil {
			e2eError(fmt.Errorf("host player is unavailable for gold fixture"))
			return
		}
		before := player.GoldVal
		player.GoldVal = uint32(gold)
		legacy.Nox_xxx_protectGoldDelta_56F920(player.ProtPlayerGold, int32(player.GoldVal-before))
		packet := server.BuildShopGoldReportPacket4D8870(player.GoldVal)
		noxServer.NetSendPacketXxx0(player.Index(), packet[:], nil, 1)
		e2eLog.Printf("PLAYER GOLD FIXTURE: before=%d after=%d", before, player.GoldVal)
	})
}

func (sc *e2eScenario) AssertPlayerGold(gold int, name string) {
	sc.add(0, name, func() {
		unit := noxServer.Players.HostUnit()
		if unit == nil || unit.UpdateDataPlayer().Player == nil {
			e2eError(fmt.Errorf("host player is unavailable for gold assertion"))
			return
		}
		serverGold := unit.UpdateDataPlayer().Player.GoldVal
		clientGold := legacy.Nox_client_gold_4674A0()
		if serverGold != uint32(gold) || clientGold != uint32(gold) {
			e2eError(fmt.Errorf("player gold = server:%d client:%d, want %d", serverGold, clientGold, gold))
			return
		}
		e2eLog.Printf("PLAYER GOLD: server=%d client=%d", serverGold, clientGold)
	})
}

func (sc *e2eScenario) AssertItemAmount(amount, maxAmount, price int, name string) {
	sc.add(0, name, func() {
		active, gotAmount, gotMax := legacy.Nox_gui_itemAmountState()
		if !active || gotAmount != uint32(amount) || gotMax != uint32(maxAmount) {
			e2eError(fmt.Errorf("item amount state = active:%t amount:%d max:%d, want active:true amount:%d max:%d", active, gotAmount, gotMax, amount, maxAmount))
			return
		}
		priceEnabled, gotPrice := legacy.Nox_gui_itemAmountPrice()
		if price > 0 && (!priceEnabled || gotPrice != uint32(price)) {
			e2eError(fmt.Errorf("item amount price = enabled:%t price:%d, want enabled:true price:%d", priceEnabled, gotPrice, price))
			return
		}
		e2eLog.Printf("ITEM AMOUNT: active=true amount=%d max=%d price_enabled=%t unit_price=%d", gotAmount, gotMax, priceEnabled, gotPrice)
	})
}

func (sc *e2eScenario) AssertItemAmountClosed(name string) {
	sc.add(0, name, func() {
		active, amount, maxAmount := legacy.Nox_gui_itemAmountState()
		if active {
			e2eError(fmt.Errorf("item amount state remained open: amount=%d max=%d", amount, maxAmount))
			return
		}
		e2eLog.Printf("ITEM AMOUNT: active=false")
	})
}

func (sc *e2eScenario) OpenShopFixture(typeID string, count, price int, name string) {
	sc.add(0, name, func() {
		if count <= 0 || count > 32 {
			e2eError(fmt.Errorf("shop fixture count must be in 1..32, got %d", count))
			return
		}
		if price <= 0 {
			e2eError(fmt.Errorf("shop fixture price must be positive, got %d", price))
			return
		}
		gold := uint64(count) * uint64(price)
		if gold > uint64(^uint32(0)) {
			e2eError(fmt.Errorf("shop fixture total price overflows uint32: count=%d price=%d", count, price))
			return
		}
		itemType := noxServer.Types.ByID(typeID)
		shopType := noxServer.Types.ByID("Shopkeeper")
		if itemType == nil || shopType == nil {
			e2eError(fmt.Errorf("shop fixture types are unavailable: shop=%t item=%q:%t", shopType != nil, typeID, itemType != nil))
			return
		}

		var reportGold [5]byte
		reportGold[0] = byte(netmsg.MSG_REPORT_GOLD)
		binary.LittleEndian.PutUint32(reportGold[1:], uint32(gold))
		if got := legacy.Nox_xxx_netOnPacketRecvCli_48EA70_switch(server.HostPlayerIndex, netmsg.MSG_REPORT_GOLD, reportGold[:]); got != len(reportGold) {
			e2eError(fmt.Errorf("shop fixture gold packet consumed %d bytes, want %d", got, len(reportGold)))
			return
		}

		var start [86]byte
		start[0] = 0xC9
		start[1] = 0x0D
		binary.LittleEndian.PutUint16(start[2:4], uint16(shopType.Ind()))
		for i, r := range "E2E Merchant" {
			binary.LittleEndian.PutUint16(start[4+2*i:6+2*i], uint16(r))
		}
		if got := legacy.Nox_xxx_netOnPacketRecvCli_48EA70_switch(server.HostPlayerIndex, netmsg.MSG_TRADE, start[:]); got != len(start) {
			e2eError(fmt.Errorf("shop fixture start packet consumed %d bytes, want %d", got, len(start)))
			return
		}

		for i := 0; i < count; i++ {
			var item [18]byte
			item[0] = 0xC9
			item[1] = 0x08
			binary.LittleEndian.PutUint16(item[2:4], uint16(itemType.Ind()))
			binary.LittleEndian.PutUint16(item[4:6], uint16(0x700+i))
			binary.LittleEndian.PutUint32(item[6:10], uint32(price))
			for j := 14; j < 18; j++ {
				item[j] = 0xFF
			}
			if got := legacy.Nox_xxx_netOnPacketRecvCli_48EA70_switch(server.HostPlayerIndex, netmsg.MSG_TRADE, item[:]); got != len(item) {
				e2eError(fmt.Errorf("shop fixture item packet %d consumed %d bytes, want %d", i, got, len(item)))
				return
			}
		}
		e2eLog.Printf("SHOP FIXTURE: item=%s count=%d price=%d", typeID, count, price)
	})
}

func (sc *e2eScenario) CloseShopFixture(name string) {
	sc.add(0, name, func() {
		packet := [...]byte{byte(netmsg.MSG_TRADE), 0x02}
		if got := legacy.Nox_xxx_netOnPacketRecvCli_48EA70_switch(server.HostPlayerIndex, netmsg.MSG_TRADE, packet[:]); got != len(packet) {
			e2eError(fmt.Errorf("shop fixture close packet consumed %d bytes, want %d", got, len(packet)))
			return
		}
		e2eLog.Printf("SHOP FIXTURE: closed")
	})
}

func (sc *e2eScenario) OpenServerShopFixture(typeID string, count int, name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return noxServer.Players.HostUnit() != nil
	}, func() {
		if count <= 0 || count > 32 {
			e2eError(fmt.Errorf("server shop fixture count must be in 1..32, got %d", count))
			return
		}
		itemType := noxServer.Types.ByID(typeID)
		if itemType == nil {
			e2eError(fmt.Errorf("server shop fixture item type %q is unavailable", typeID))
			return
		}
		clientType := noxClient.Things.TypeByID(typeID)
		if clientType == nil {
			e2eError(fmt.Errorf("server shop fixture client item type %q is unavailable", typeID))
			return
		}
		if clientType.Index() != itemType.Ind() {
			e2eError(fmt.Errorf("server shop fixture item type index differs: server=%d client=%d", itemType.Ind(), clientType.Index()))
			return
		}
		e2eLog.Printf("SERVER SHOP ITEM TYPE: id=%s index=%d class=%v", typeID, itemType.Ind(), itemType.Class())
		player := noxServer.Players.HostUnit()
		merchant := noxServer.NewObjectByTypeID("Shopkeeper")
		if merchant == nil {
			e2eError(fmt.Errorf("server shop fixture cannot create Shopkeeper"))
			return
		}
		// This fixture verifies the trade protocol and UI, not autonomous NPC
		// behavior. Keep its synthetic merchant static like a scripted map shop.
		merchant.ObjFlags |= object.FlagNoUpdate
		idata := merchant.InitDataShopkeeper()
		idata.Count = 1
		idata.Items[0] = server.ShopkeeperItemDefinition{
			TypeInd: uint32(itemType.Ind()),
			Count:   uint8(count),
		}
		idata.BuyMultiplier = 1
		idata.SellMultiplier = 1
		noxServer.CreateObjectAt(merchant, nil, player.Pos())
		noxServer.ObjectsAddPending()
		wireCode := noxServer.GetUnitNetCode(merchant)
		if wireCode <= 0 || wireCode > int(^uint16(0)) {
			e2eError(fmt.Errorf("server shop fixture wire code = %#x", wireCode))
			return
		}
		update := merchant.UpdateDataMonster()
		head := update.AIStackHead()
		e2eLog.Printf("SERVER SHOP MERCHANT AI: flags=%v subclass=%v stack=%d action=%v aggression=%g status=%v enemy=%p health=%d/%d",
			merchant.Flags(), merchant.SubClass().AsMonster(), update.AIStackInd, head.Type(), update.Aggression,
			update.StatusFlags, update.CurrentEnemy, merchant.HealthData.Cur, merchant.HealthData.Max)
		packet := [...]byte{byte(netmsg.MSG_TRADE), 0x15, 0, 0}
		binary.LittleEndian.PutUint16(packet[2:4], uint16(wireCode))
		if got := nox_xxx_netClientSend2_4E53C0(server.HostPlayerIndex, packet[:], nil, 1); got != 1 {
			e2eError(fmt.Errorf("server shop fixture client send = %d, want 1", got))
			return
		}
		e2e.shopMerchant = merchant
		e2e.shopSession = nil
		e2eLog.Printf("SERVER SHOP FIXTURE: merchant=%p netcode=%d wire=%#x item=%s count=%d", merchant, merchant.NetCode, wireCode, typeID, count)
	})
}

func (sc *e2eScenario) AssertServerShop(active bool, typeID string, count int, name string) {
	sc.add(0, name, func() {
		player := noxServer.Players.HostUnit()
		if player == nil {
			e2eError(fmt.Errorf("server shop assertion has no host player unit"))
			return
		}
		session := player.UpdateDataPlayer().Trade70
		if active {
			if session == nil || !noxServer.Server.IsTradeSessionNative(session) {
				e2eError(fmt.Errorf("server shop session = %p native=%t, want active native session", session, noxServer.Server.IsTradeSessionNative(session)))
				return
			}
			if session.Field0 != 1 || session.Field8 != player || session.Field12 != e2e.shopMerchant || session.Field16 != 1 {
				e2eError(fmt.Errorf("server shop session fields = active:%d player:%p merchant:%p kind:%d", session.Field0, session.Field8, session.Field12, session.Field16))
				return
			}
			gotCount := 0
			var cost uint32
			for item := session.Field20; item != nil; item = item.Field8 {
				if item.Item0 == nil {
					e2eError(fmt.Errorf("server shop item %d has nil object", gotCount))
					return
				}
				typ := item.Item0.ObjectTypeC()
				if typ == nil {
					e2eError(fmt.Errorf("server shop item %d has unknown type index %d", gotCount, item.Item0.TypeInd))
					return
				}
				if typeID != "" && typ.ID() != typeID {
					e2eError(fmt.Errorf("server shop item %d type = %q, want %q", gotCount, typ.ID(), typeID))
					return
				}
				if item.Cost4 == 0 {
					e2eError(fmt.Errorf("server shop item %d has zero cost", gotCount))
					return
				}
				cost = item.Cost4
				gotCount++
			}
			if count != 0 && gotCount != count {
				e2eError(fmt.Errorf("server shop item count = %d, want %d", gotCount, count))
				return
			}
			if count != 0 && e2e.shopMerchant != nil {
				definitionCount := 0
				idata := e2e.shopMerchant.InitDataShopkeeper()
				for i := 0; i < int(idata.Count); i++ {
					if typ := noxServer.Types.ByInd(int(idata.Items[i].TypeInd)); typ != nil && (typeID == "" || typ.ID() == typeID) {
						definitionCount += int(idata.Items[i].Count)
					}
				}
				if definitionCount != count {
					e2eError(fmt.Errorf("server shop definition count = %d, want %d", definitionCount, count))
					return
				}
			}
			e2e.shopSession = session
			e2eLog.Printf("SERVER SHOP: active=true session=%p merchant=%p item=%s count=%d cost=%d", session, session.Field12, typeID, gotCount, cost)
			return
		}
		if session != nil {
			e2eError(fmt.Errorf("server shop session remained active: %p", session))
			return
		}
		if e2e.shopSession != nil && noxServer.Server.IsTradeSessionNative(e2e.shopSession) {
			e2eError(fmt.Errorf("server shop session remained allocated: %p", e2e.shopSession))
			return
		}
		e2eLog.Printf("SERVER SHOP: active=false released=true")
	})
}

func (sc *e2eScenario) AssertShop(active bool, mode, count int, name string) {
	sc.add(0, name, func() {
		gotActive, gotMode, gotCount := legacy.Nox_gui_shopState()
		if gotActive != active || gotMode != uint32(mode) || gotCount != uint32(count) {
			e2eError(fmt.Errorf("shop state = active:%t mode:%d count:%d, want active:%t mode:%d count:%d", gotActive, gotMode, gotCount, active, mode, count))
			return
		}
		e2eLog.Printf("SHOP: active=%t mode=%d count=%d", gotActive, gotMode, gotCount)
	})
}

func imageDiff(pix1, pix2 []byte) []byte {
	out := make([]byte, len(pix1))
	for i := range out {
		if i >= len(pix2) {
			out[i] = pix1[i]
			if i%4 == 3 { // keep unmatched pixels visible
				out[i] = 0xff
			}
			continue
		}
		dp := int16(pix1[i]) - int16(pix2[i])
		if dp < 0 {
			dp = -dp
		}
		dp *= 10
		if dp > 0xff {
			dp = 0xff
		}
		if i%4 == 3 { // alpha
			dp = 0xff - dp
		}
		out[i] = byte(dp)
	}
	return out
}

type e2eCheckSave struct {
	Name   string
	Hashes map[string]string
}

func (sc *e2eScenario) Save(name string, hashes map[string]string) {
	sc.add(0, name, func() {
		e2e.checkSave = &e2eCheckSave{Name: name, Hashes: hashes}
	})
}

func (sc *e2eScenario) Screen(name string) {
	sc.add(0, name, func() {
		var serverNetCode uint32
		var playerStatus uint32
		var playerPhase byte
		var serverPos types.Pointf
		var serverHealthCur, serverHealthMax uint16
		var dialogActive bool
		if unit := noxServer.Players.HostUnit(); unit != nil {
			serverNetCode = unit.NetCode
			serverPos = unit.PosVec
			if health := unit.HealthData; health != nil {
				serverHealthCur = health.Cur
				serverHealthMax = health.Max
			}
			update := unit.UpdateDataPlayer()
			dialogActive = update.DialogWith != nil
			if player := unit.ControllingPlayer(); player != nil {
				playerStatus = player.Field3680
				playerPhase = player.Field3676
			}
		}
		itemAmountActive, itemAmount, itemAmountMax := legacy.Nox_gui_itemAmountState()
		e2eLog.Printf("SCREEN: %s connected=%t player_netcode=%d server_netcode=%d player_phase=%d player_status=%#x server_pos=(%.3f,%.3f) server_health=%d/%d dialog_active=%t drawables=%d player_drawable=%t inventory_state=%d inventory_offset=%d inventory_dragged=%t item_amount_active=%t item_amount=%d item_amount_max=%d", name, nox_client_isConnected(), legacy.ClientPlayerNetCode(), serverNetCode, playerPhase, playerStatus, serverPos.X, serverPos.Y, serverHealthCur, serverHealthMax, dialogActive, noxClient.Objs.Count, noxClient.ClientPlayerUnit() != nil, legacy.Nox_client_inventoryAnimationState(), legacy.Nox_client_inventoryAnimationOffset(), legacy.Nox_client_inventoryHasDragged(), itemAmountActive, itemAmount, itemAmountMax)
		fname := strings.ReplaceAll(strings.ToLower(name), " ", "_")
		fname = filepath.Join(e2e.path, "testdata", fname)
		if err := os.MkdirAll(filepath.Dir(fname), 0755); err != nil {
			panic(err)
		}
		img := noxClient.r.CopyPixBuffer()
		var ibuf bytes.Buffer
		if err := png.Encode(&ibuf, img); err != nil {
			panic(err)
		}
		if e2eOverride {
			if err := os.WriteFile(fname+".png", ibuf.Bytes(), 0644); err != nil {
				panic(err)
			}
			return
		}
		gotName := fname + "_got.png"
		diffName := fname + "_diff.png"
		if _, err := os.Stat(gotName); err == nil {
			if err = os.Remove(gotName); err != nil {
				e2eLog.Println(err)
			}
		}
		if _, err := os.Stat(diffName); err == nil {
			if err = os.Remove(diffName); err != nil {
				e2eLog.Println(err)
			}
		}
		if data, err := os.ReadFile(fname + ".png"); err == nil {
			exp, err := png.Decode(bytes.NewReader(data))
			if err != nil {
				panic(err)
			}
			var edata []byte
			switch exp := exp.(type) {
			case *image.RGBA:
				edata = exp.Pix
			case *image.NRGBA:
				edata = exp.Pix
			default:
				panic(exp)
			}
			if !bytes.Equal(img.Pix, edata) {
				if err := os.WriteFile(gotName, ibuf.Bytes(), 0644); err != nil {
					panic(err)
				}

				diff := imageDiff(img.Pix, edata)
				ibuf.Reset()
				img.Pix = diff
				if err := png.Encode(&ibuf, img); err != nil {
					panic(err)
				}
				if err := os.WriteFile(diffName, ibuf.Bytes(), 0644); err != nil {
					panic(err)
				}
				e2eError(fmt.Errorf("screen %q differs from %s", name, fname+".png"))
			}
		} else if os.IsNotExist(err) {
			if err := os.WriteFile(fname+".png", ibuf.Bytes(), 0644); err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	})
}

type e2eFileYML struct {
	Steps []e2eStepYML `yaml:"steps"`
}

type e2eStepYML struct {
	Action   string        `yaml:"action"`
	Time     uint64        `yaml:"dt,omitempty"`
	Dur      time.Duration `yaml:"dur,omitempty"`
	Name     string        `yaml:"name,omitempty"`
	X        int           `yaml:"x,omitempty"`
	Y        int           `yaml:"y,omitempty"`
	Ang      float64       `yaml:"ang,omitempty"`
	Slot     int           `yaml:"slot,omitempty"`
	Item     string        `yaml:"item,omitempty"`
	Modifier string        `yaml:"modifier,omitempty"`
	Mask     uint32        `yaml:"mask,omitempty"`
	Count    int           `yaml:"count,omitempty"`
	Amount   int           `yaml:"amount,omitempty"`
	Max      int           `yaml:"max,omitempty"`
	Price    int           `yaml:"price,omitempty"`
	Gold     int           `yaml:"gold,omitempty"`
	Health   int           `yaml:"health,omitempty"`
	Full     bool          `yaml:"full,omitempty"`
	Mode     int           `yaml:"mode,omitempty"`
	Active   bool          `yaml:"active,omitempty"`
	Event    *e2eStepRaw   `yaml:"ev,omitempty"`
}

func (sc *e2eScenario) Load(path string) {
	e2eLog.Printf("LOAD: %s", path)
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	var list e2eFileYML
	err = yaml.Unmarshal(data, &list)
	if err != nil {
		panic(err)
	}
	for _, l := range list.Steps {
		dt := time.Duration(l.Time)
		if l.Dur != 0 {
			dt = l.Dur
		}
		switch l.Action {
		case "quit":
			sc.Quit(dt)
		case "slow":
			sc.Slow(dt)
		case "wait":
			sc.Wait(dt, l.Name)
		case "move":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.Move(l.X, l.Y, l.Name)
		case "click":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.ClickLeft(l.X, l.Y, l.Name)
		case "click-inventory-item":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.ClickInventoryItem(l.Item, l.Name)
		case "click-item-amount-accept":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.ClickItemAmountAccept(image.Pt(l.X, l.Y), l.Name)
		case "click-npc-dialog-done":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.ClickNPCDialogDone(l.Name)
		case "interact":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.ClickSlowLeft(l.X, l.Y, l.Name)
		case "walk":
			sc.WalkFor(l.Ang, dt, l.Name)
		case "walk-start":
			sc.WalkStart(l.Ang, dt, l.Name)
		case "walk-dir":
			sc.WalkDir(l.Ang, dt, l.Name)
		case "walk-stop":
			sc.WalkEnd()
		case "run":
			sc.RunFor(l.Ang, dt, l.Name)
		case "run-start":
			sc.RunStart(l.Ang, dt, l.Name)
		case "run-dir":
			sc.RunDir(l.Ang, dt, l.Name)
		case "run-stop":
			sc.RunEnd()
		case "screen":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.Screen(l.Name)
		case "esc":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.Key(keybind.KeyEsc, l.Name)
		case "inventory":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.Key(keybind.KeyI, l.Name)
		case "jump":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.Key(keybind.KeySpace, l.Name)
		case "melee":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.Melee(l.Ang, l.Name)
		case "spawn-monster":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.SpawnMonster(l.Item, image.Pt(l.X, l.Y), l.Name)
		case "assert-monster-encounter":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertMonsterEncounter(l.Name)
		case "wait-monster-dead":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.WaitMonsterDead(l.Name)
		case "wait-player-dead":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.WaitPlayerDead(l.Name)
		case "assert-monster-world-damage":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertMonsterWorldDamage(l.Name)
		case "wait-death-screen":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.WaitDeathScreen(l.Name)
		case "click-save-load":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.ClickSaveLoad(l.Name)
		case "click-dialog-yes":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.ClickDialogYes(l.Name)
		case "wait-player-reloaded":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.WaitPlayerReloaded(l.Name)
		case "assert-player-moved-after-reload":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertPlayerMovedAfterReload(l.Name)
		case "grant-item":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.GrantInventoryItems(l.Item, l.Count, l.Name)
		case "grant-engage-item":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.GrantEngageItem(l.Item, l.Modifier, l.Mask, l.Name)
		case "assert-engage-item-equipped":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertEngageItemEquipped(l.Name)
		case "assert-engage-item-dequipped":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertEngageItemDequipped(l.Name)
		case "spawn-ground-item":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.SpawnGroundItem(l.Item, image.Pt(l.X, l.Y), l.Name)
		case "pickup-ground-item":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.PickupGroundItem(l.Name)
		case "assert-ground-item-picked":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertGroundItemPicked(l.Name)
		case "drag-inventory-item-out":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.DragInventoryItemOut(l.Item, l.Name)
		case "assert-ground-item-dropped":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertGroundItemDropped(l.Name)
		case "assert-inventory-count":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertInventoryItemCount(l.Item, l.Count, l.Name)
		case "damage-inventory-item":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.DamageInventoryItem(l.Item, l.Health, l.Name)
		case "assert-inventory-health":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertInventoryItemHealth(l.Item, l.Health, l.Full, l.Name)
		case "set-player-gold":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.SetPlayerGold(l.Gold, l.Name)
		case "assert-player-gold":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertPlayerGold(l.Gold, l.Name)
		case "assert-item-amount":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertItemAmount(l.Amount, l.Max, l.Price, l.Name)
		case "assert-item-amount-closed":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertItemAmountClosed(l.Name)
		case "open-shop-fixture":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.OpenShopFixture(l.Item, l.Count, l.Price, l.Name)
		case "close-shop-fixture":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.CloseShopFixture(l.Name)
		case "open-server-shop-fixture":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.OpenServerShopFixture(l.Item, l.Count, l.Name)
		case "assert-server-shop":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertServerShop(l.Active, l.Item, l.Count, l.Name)
		case "assert-shop":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertShop(l.Active, l.Mode, l.Count, l.Name)
		case "cast":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			switch l.Slot {
			default:
				fallthrough
			case 1:
				sc.Key(keybind.KeyA, l.Name)
			case 2:
				sc.Key(keybind.KeyS, l.Name)
			case 3:
				sc.Key(keybind.KeyD, l.Name)
			case 4:
				sc.Key(keybind.KeyF, l.Name)
			case 5:
				sc.Key(keybind.KeyG, l.Name)
			}
		case "raw":
			ev := l.Event
			switch ev.Type {
			case "save":
				sc.Save(ev.SaveName, ev.Hashes)
			case "move":
				sc.Input(dt, "", &seat.MouseMoveEvent{
					Relative: ev.Relative, Pos: ev.Pos, Rel: ev.Rel,
				})
			case "button":
				sc.Input(dt, "", &seat.MouseButtonEvent{
					Pressed: ev.Pressed, Button: ev.Button,
				})
			case "wheel":
				sc.Input(dt, "", &seat.MouseWheelEvent{
					Wheel: ev.Wheel,
				})
			case "key":
				sc.Input(dt, "", &seat.KeyboardEvent{
					Pressed: ev.Pressed, Key: ev.Key,
				})
			case "text_edit":
				sc.Input(dt, "", &seat.TextEditEvent{
					Text: ev.Text,
				})
			case "text_input":
				sc.Input(dt, "", &seat.TextInputEvent{
					Text: ev.Text,
				})
			case "closed":
				sc.Input(dt, "", seat.WindowClosed)
			default:
				panic("unsupported type: " + ev.Type)
			}
		default:
			panic("unsupported type: " + l.Action)
		}
	}
}

var (
	e2eJobs = make(chan *e2eScenario)
)

func e2eAbsPath(s string) string {
	if filepath.IsAbs(s) {
		return s
	}
	if _, err := os.Stat(s); err != nil {
		s = filepath.Join(filepath.Dir(os.Args[0]), s)
	}
	p, err := filepath.Abs(s)
	if err != nil {
		panic(err)
	}
	return p
}

func e2eInit() {
	opennoxDir := filepath.Dir(os.Args[0])
	e2e.path = filepath.Join(opennoxDir, "e2e")
	fname := filepath.Join(e2e.path, "e2e.yaml")
	if s := e2eRecord; s != "" {
		if filepath.Ext(s) == "" {
			s = filepath.Join(s, "e2e.yaml")
		}
		s = e2eAbsPath(s)
		e2e.recording = true
		fname = s
		e2e.path = s
	} else if s = e2ePlay; s != "" && s != "true" {
		s = e2eAbsPath(s)
		fname = s
		e2e.path = filepath.Dir(s)
	}
	if s := e2eSlow; s != "" {
		dt, err := time.ParseDuration(s)
		if err != nil {
			panic(err)
		}
		e2e.slow = dt
	}

	e2eLog.Println("WARNING: starting in e2e test mode")
	e2e.p = newPlayformE2E()
	platform.Set(e2e.p)
	if e2e.recording {
		e2eLog.Printf("RECORD: %s", fname)
		if e2e.slow == 0 {
			e2e.slow = e2eDefaultDelay
		}
		return
	}

	go testInit(fname)
	sc, ok := <-e2eJobs
	if !ok {
		panic("cannot init e2e")
	}
	e2eQueue(sc)
}

type e2eStepRaw struct {
	Type string `yaml:"type"`

	Relative bool         `yaml:"rel,omitempty"`
	Pos      image.Point  `yaml:"pos,omitempty"`
	Rel      types.Pointf `yaml:"pos_rel,omitempty"`

	Button  seat.MouseButton `yaml:"button,omitempty"`
	Pressed bool             `yaml:"pressed,omitempty"`
	Key     keybind.Key      `yaml:"key,omitempty"`

	Wheel int `yaml:"wheel,omitempty"`

	Text string `yaml:"text,omitempty"`

	SaveName string            `yaml:"savename,omitempty"`
	Hashes   map[string]string `yaml:"hashes,omitempty"`
}

func e2eSaveRecording() {
	var list e2eFileYML
	var last time.Duration
	for _, r := range e2e.recorded {
		dt := r.Time - last
		last = r.Time
		if r.Save != nil {
			list.Steps = append(list.Steps, e2eStepYML{
				Action: "raw",
				Time:   uint64(dt),
				Event: &e2eStepRaw{
					Type:     "save",
					SaveName: r.Save.Name,
					Hashes:   r.Save.Hash,
				},
			})
		} else if r.Input != nil {
			s := e2eStepYML{Action: "raw", Time: uint64(dt)}
			switch ev := r.Input.(type) {
			case *seat.MouseMoveEvent:
				s.Event = &e2eStepRaw{
					Type:     "move",
					Relative: ev.Relative,
					Pos:      ev.Pos,
					Rel:      ev.Rel,
				}
			case *seat.MouseButtonEvent:
				s.Event = &e2eStepRaw{
					Type:    "button",
					Pressed: ev.Pressed,
					Button:  ev.Button,
				}
			case *seat.MouseWheelEvent:
				s.Event = &e2eStepRaw{
					Type:  "wheel",
					Wheel: ev.Wheel,
				}
			case *seat.KeyboardEvent:
				s.Event = &e2eStepRaw{
					Type:    "key",
					Pressed: ev.Pressed,
					Key:     ev.Key,
				}
			case *seat.TextEditEvent:
				s.Event = &e2eStepRaw{
					Type: "text_edit",
					Text: ev.Text,
				}
			case *seat.TextInputEvent:
				s.Event = &e2eStepRaw{
					Type: "text_input",
					Text: ev.Text,
				}
			case seat.WindowEvent:
				switch ev {
				case seat.WindowClosed:
					s.Event = &e2eStepRaw{
						Type: "closed",
					}
				default:
					e2eLog.Printf("SKIPPED: %T", ev)
				}
			default:
				e2eLog.Printf("SKIPPED: %T", ev)
			}
			if s.Event != nil {
				list.Steps = append(list.Steps, s)
			}
		}
	}
	f, err := os.Create(e2e.path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	enc := yaml.NewEncoder(f)
	if err = enc.Encode(list); err != nil {
		panic(err)
	}
	if err = f.Close(); err != nil {
		panic(err)
	}
	e2eLog.Printf("RECORDED: %d events", len(list.Steps))
}

func e2eStop() {
	if !e2e.recording {
		return
	}
	e2eSaveRecording()
}

func e2eDone() {
	close(e2eJobs)
}

func testInit(fname string) {
	defer e2eDone()
	var sc e2eScenario
	sc.Load(fname)
	sc.Exec()
}

func e2eQueue(sc *e2eScenario) {
	var last time.Duration
	if n := len(e2e.steps); n == 0 {
		last = e2e.p.ticks()
	} else {
		last = e2e.steps[n-1].time
	}
	for _, st := range sc.steps {
		st.time += last
		e2e.steps = append(e2e.steps, st)
	}
	sc.steps = nil
	e2e.done = sc.done
}

func e2eQueueInput(evs ...seat.InputEvent) {
	e2e.input = append(e2e.input, evs...)
}

func e2eRun() {
	defer e2e.p.tick(1)
	if e2e.slow != 0 {
		time.Sleep(e2e.slow)
	}
	if e2e.recording {
		return
	}
	if len(e2e.steps) == 0 {
		if e2e.done != nil {
			close(e2e.done)
			e2e.done = nil
			e2eLog.Println("DONE")
			if sc, ok := <-e2eJobs; ok {
				e2eQueue(sc)
			} else {
				e2e.realEnable = true
				if e2e.slow == 0 {
					e2e.slow = e2eDefaultDelay
				}
				e2eLog.Println("SCRIPT COMPLETE")
			}
		}
		return
	}
	t := e2e.p.Ticks()
	n := 0
	for i := range e2e.steps {
		s := &e2e.steps[i]
		if t < s.time {
			break
		}
		if s.ready != nil && !s.ready() {
			s.waited++
			if s.waited >= s.waitTimeout {
				e2eError(fmt.Errorf("timed out after %d ticks waiting for %s", s.waitTimeout, s.name))
				n++
				continue
			}
			for j := i; j < len(e2e.steps); j++ {
				e2e.steps[j].time++
			}
			break
		}
		n++
		if s.name != "" {
			e2eLog.Println("STATE:", s.name)
		}
		if s.fnc != nil {
			s.fnc()
		}
	}
	e2e.steps = e2e.steps[n:]
}

type e2eRecordedEvent struct {
	Time  time.Duration
	Input seat.InputEvent
	Save  *e2eSave
}

type e2eSave struct {
	Name string            `json:"name"`
	Hash map[string]string `json:"hash"`
}

func e2eOnSave(name string) {
	if e2e.recording {
		t := platform.Ticks()
		path := datapath.Save(name)
		hash := e2eHashDir(path)
		e2e.recorded = append(e2e.recorded, e2eRecordedEvent{
			Time: t - 1, Save: &e2eSave{Name: name, Hash: hash},
		})
	} else if s := e2e.checkSave; s != nil {
		defer func() {
			e2e.checkSave = nil
		}()
		path := datapath.Save(name)
		got := e2eHashDir(path)
		if !maps.Equal(got, s.Hashes) {
			err := fmt.Errorf("unexpected save data:\ngot: %+v\nvs\nexp: %+v", got, s.Hashes)
			e2eError(err)
		}
	}
}

func e2eRealInput(ev seat.InputEvent) {
	t := platform.Ticks()
	if e2e.recording {
		if ev == seat.WindowClosed {
			e2eSaveRecording()
		}
		e2e.recorded = append(e2e.recorded, e2eRecordedEvent{
			Time: t - 1, Input: ev,
		})
		e2eQueueInput(ev)
		return
	}
	switch ev := ev.(type) {
	case *seat.MouseMoveEvent:
		if !ev.Relative {
			e2e.realMouse = ev.Pos
		}
	case *seat.MouseButtonEvent:
		e2eLog.Printf("input(%v,%d): %#v @ %v", t, uint64(t), ev, e2e.realMouse)
	}
	if e2e.realEnable {
		e2eQueueInput(ev)
		return
	}
	switch ev := ev.(type) {
	case seat.WindowEvent:
		switch ev {
		case seat.WindowClosed:
			e2eQueueInput(ev)
			e2e.realEnable = true
			e2e.steps = nil
		}
	}
}

func e2eInputTick() {
	for _, ev := range e2e.input {
		for _, fnc := range e2e.onInput {
			fnc(ev)
		}
	}
	e2e.input = e2e.input[:0]
}

const e2eInputConf = `
---
MousePickup = Left
MOUSE_BUTTON_RIGHT = MoveForward
MOUSE_BUTTON_LEFT = Action
SPACE = Jump
MOUSE_BUTTON_MID = Jump
I = ToggleInventory
Q = ToggleInventory
B = ToggleBook
TAB = ToggleMap
1 = MapZoomOut
2 = MapZoomIn
A = InvokeSlot1
S = InvokeSlot2
D = InvokeSlot3
F = InvokeSlot4
G = InvokeSlot5
MOUSE_WHEEL_UP = PreviousSpellSet
W = PreviousSpellSet
MOUSE_WHEEL_DOWN = NextSpellSet
E = NextSpellSet
R = SelectSpellSet
LEFT_SHIFT = InvertSpellTarget
RIGHT_SHIFT = InvertSpellTarget
T = PlaceTrapBomber
V = SwapWeapons
X = QuickHealth
C = QuickMana
Z = QuickCurePoison
ENTER = Chat
BACKSPACE = TeamChat
F1 = ToggleConsole
ESC = ToggleQuitMenu
HOME = ToggleServerMenu
F9 = ToggleRank
F10 = ToggleNetstat
F11 = ToggleGUI
F2 = AutoSave
F4 = AutoLoad
J = Taunt
K = Point
L = Laugh
PAGEUP = IncreaseWindowSize
PAGEDOWN = DecreaseWindowSize
INS = IncreaseGamma
DEL = DecreaseGamma
F12 = ScreenShot
---
`

func e2eHash() hash.Hash {
	h, err := blake2b.New256(nil)
	if err != nil {
		panic(err)
	}
	return h
}

func e2eHashDir(dir string) map[string]string {
	hashes := make(map[string]string)
	err := ifs.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		h := e2eHash()
		f, err := ifs.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(h, f)
		if err != nil {
			return err
		}
		name := strings.TrimPrefix(path, dir)
		name = strings.TrimPrefix(name, string(filepath.Separator))
		name = strings.ReplaceAll(name, string(filepath.Separator), "/")
		hashes[name] = hex.EncodeToString(h.Sum(nil))
		return nil
	})
	if err != nil {
		panic(err)
	}
	return hashes
}
