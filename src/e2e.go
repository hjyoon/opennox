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

	"github.com/opennox/libs/common"
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
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/gui"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
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
	shopMerchantWireCode  uint16
	shopSession           *server.TradeSession
	fieldGuideID          int
	fieldGuideCreature    string
	monster               *server.Object
	monsterPlayerHP       uint16
	monsterShield         *server.Object
	monsterShieldHP       uint16
	monsterShieldCarry    uint32
	monsterWorldTarget    *server.Object
	monsterWorldTargetHP  uint16
	groundItem            *server.Object
	groundItemTypeID      string
	groundItemPickupName  string
	groundItemPickupPtr   unsafe.Pointer
	groundItemOwned       bool
	groundItemBefore      int
	groundItemWireCode    uint16
	groundItemLivesBefore uint32
	groundItemDropped     *server.Object
	groundItemDropChecks  uint32
	engageItem            *server.Object
	engageItemTypeID      string
	engageModifier        *server.ModifierEff
	engageOwner           *server.Object
	engageOwnerMask       uint32
	engageOwnerMaskBefore uint32
	deadPlayer            *server.Object
	reloadPlayer          *server.Object
	reloadPos             types.Pointf
	lavaPlayer            *server.Object
	lavaOriginalPos       types.Pointf
	lavaPos               types.Pointf
	lavaHealthBefore      uint16
	lavaFrameBefore       uint32
	lavaGroundItem        *server.Object
	lavaGroundOriginalPos types.Pointf
	lavaGroundHealth      uint16
	lavaGroundFrame       uint32
	poisonPlayer          *server.Object
	poisonHealthBefore    uint16
	poisonFrameBefore     uint32
	ovalShieldPlayer      *server.Object
	ovalShieldRecord      *server.DurSpell
	ovalShieldFrameBefore uint32
	channelLifePlayer     *server.Object
	channelLifeRecord     *server.DurSpell
	channelLifeFrame      uint32
	channelLifeHP         uint16
	firewalkPlayer        *server.Object
	firewalkRecord        *server.DurSpell
	firewalkFrame         uint32
	greaterHealPlayer     *server.Object
	greaterHealRecord     *server.DurSpell
	greaterHealFrame      uint32
	greaterHealHP         uint16
	greaterHealMana       uint16
	forceOfNaturePlayer   *server.Object
	forceOfNatureRecord   *server.DurSpell
	forceOfNatureCharge   *server.Object
	forceOfNatureFrame    uint32
	forceOfNatureLaunches uint64
	manaBombPlayer        *server.Object
	manaBombRecord        *server.DurSpell
	manaBombCharge        *server.Object
	manaBombFrame         uint32
	manaBombMass          uint32
	manaBombPower         int32
	chainLightningPlayer  *server.Object
	chainLightningTarget  *server.Object
	chainLightningRecord  *server.DurSpell
	chainLightningFrame   uint32
	chainLightningHealth  uint16
	energyBoltRecord      *server.DurSpell
	energyBoltFrame       uint32
	energyBoltHealth      uint16
	drainManaRecord       *server.DurSpell
	drainManaFrame        uint32
	drainManaBefore       uint16
	durationRayDrawSource uint16
	durationRayDrawTarget uint16
	durationRayDrawFrame  uint32
	durationRayDrawables  [6]*client.Drawable
	turnUndeadRecord      *server.DurSpell
	turnUndeadFrame       uint32
	blinkPlayer           *server.Object
	blinkRecord           *server.DurSpell
	blinkFrame            uint32
	blinkOrigin           types.Pointf
	swapCaster            *server.Object
	swapTarget            *server.Object
	swapRecord            *server.DurSpell
	swapFrame             uint32
	swapCasterOrigin      types.Pointf
	swapTargetOrigin      types.Pointf
	teleportTargetPlayer  *server.Object
	teleportTargetRecord  *server.DurSpell
	teleportTargetOrigin  types.Pointf
	teleportTargetPos     types.Pointf
	teleportTargetFrame   uint32
	teleportPopPlayer     *server.Object
	teleportPopRecord     *server.DurSpell
	teleportPopMarker     *server.Object
	teleportPopOrigin     types.Pointf
	teleportPopMarkerPos  types.Pointf
	teleportPopFrame      uint32
	teleportMarkPlayer    *server.Object
	teleportMarkRecord    *server.DurSpell
	teleportMarkOrigin    types.Pointf
	teleportMarkPos       types.Pointf
	teleportMarkFrame     uint32
	moonglowPlayer        *server.Object
	moonglowRecord        *server.DurSpell
	moonglowVisual        *server.Object
	moonglowFrameBefore   uint32
	smokeBlastBaseline    map[*client.Drawable]struct{}
	smokeBlastPos         image.Point
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
	steps                 []e2eStep
	done                  chan struct{}
	wizard1UrchinsBefore  int
	wizard1UrchinHPBefore int
	wizard1LightningStart uint32
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

func (sc *e2eScenario) AssertLastSpellSlot(slot int, name string) {
	sc.add(0, name, func() {
		got := memmap.Int32(0x587000, 133484)
		if got != int32(slot-1) {
			e2eError(fmt.Errorf("last spell slot: got %d, want %d", got, slot-1))
			return
		}
		e2eLog.Printf("LAST SPELL SLOT: %d", got)
	})
}

func (sc *e2eScenario) AssertSpellSetRow(row int, name string) {
	sc.add(0, name, func() {
		got := legacy.Nox_xxx_buttonsGetSelectedRow_45E180()
		if got != row {
			e2eError(fmt.Errorf("spell set row: got %d, want %d", got, row))
			return
		}
		e2eLog.Printf("SPELL SET ROW: %d", got)
	})
}

func (sc *e2eScenario) AssertQuickbarExpanded(active bool, name string) {
	sc.add(0, name, func() {
		got := memmap.Int32(0x5D4594, 1049476) == 1
		if got != active {
			e2eError(fmt.Errorf("expanded quickbar: got %t, want %t", got, active))
			return
		}
		e2eLog.Printf("EXPANDED QUICKBAR: %t", got)
	})
}

func (sc *e2eScenario) SetQuickbarSpell(spell, slot int, name string) {
	sc.add(0, name, func() {
		if spell <= 0 || slot < 0 || slot >= 5 {
			e2eError(fmt.Errorf("quickbar spell/slot: invalid %d/%d", spell, slot))
			return
		}
		legacy.Nox_xxx_quickBarSetSpell(spell, slot)
		e2eLog.Printf("QUICKBAR SPELL: spell=%d slot=%d", spell, slot)
	})
}

func (sc *e2eScenario) AssertQuickbarSpell(spell, slot int, name string) {
	sc.add(0, name, func() {
		got := legacy.Nox_xxx_quickBarSpell(slot)
		if got != spell {
			e2eError(fmt.Errorf("quickbar spell at slot %d: got %d, want %d", slot, got, spell))
			return
		}
		e2eLog.Printf("QUICKBAR SPELL: slot=%d spell=%d", slot, got)
	})
}

func e2eExitWithDestination() *server.Object {
	for obj := noxServer.Objs.First(); obj != nil; obj = obj.Next() {
		if obj.Xfer != legacy.Get_nox_xxx_XFerExit_4F4B90() || obj.CollideData == nil {
			continue
		}
		data := exitCollideData4DB600(unsafe.Pointer(obj))
		if data.DestinationX != 0 || data.DestinationY != 0 {
			return obj
		}
	}
	return nil
}

func (sc *e2eScenario) AssertNativeExitSaveLocation(name string) {
	sc.add(0, name, func() {
		player := noxServer.Players.HostUnit()
		if player == nil {
			e2eError(fmt.Errorf("exit save location: host player is missing"))
			return
		}
		exit := e2eExitWithDestination()
		if exit == nil {
			e2eError(fmt.Errorf("exit save location: map %q has no loaded exit with a destination", legacy.Nox_xxx_mapGetMapName_409B40()))
			return
		}
		before := make(map[*server.Object]struct{})
		for obj := noxServer.Objs.First(); obj != nil; obj = obj.Next() {
			if typ := obj.ObjectTypeC(); typ != nil && typ.ID() == "SaveGameLocation" {
				before[obj] = struct{}{}
			}
		}
		if !nox_xxx_saveMakePlayerLocation_4DB600(unsafe.Pointer(exit)) {
			e2eError(fmt.Errorf("exit save location: creation failed for exit %p", exit))
			return
		}
		data := exitCollideData4DB600(unsafe.Pointer(exit))
		want := types.Pointf{X: data.DestinationX, Y: data.DestinationY}
		for obj := noxServer.Objs.First(); obj != nil; obj = obj.Next() {
			if typ := obj.ObjectTypeC(); typ != nil && typ.ID() == "SaveGameLocation" {
				if _, existed := before[obj]; existed {
					continue
				}
				if obj.PosVec != want || obj.ScriptIDVal != player.ScriptIDVal {
					e2eError(fmt.Errorf("exit save location: got pos=%v script=%d, want pos=%v script=%d", obj.PosVec, obj.ScriptIDVal, want, player.ScriptIDVal))
					return
				}
				e2eLog.Printf("EXIT SAVE LOCATION: map=%q exit=%p collide=%p saved=%p destination=%v pointers=native", legacy.Nox_xxx_mapGetMapName_409B40(), exit, exit.CollideData, obj, want)
				return
			}
		}
		e2eError(fmt.Errorf("exit save location: no new SaveGameLocation object was linked"))
	})
}

func (sc *e2eScenario) AssertNativeExitCoopSave(name string) {
	sc.add(0, name, func() {
		exit := e2eExitWithDestination()
		if exit == nil {
			e2eError(fmt.Errorf("exit coop save: map %q has no loaded exit with a destination", noxServer.getServerMap()))
			return
		}
		previousExit := dword_5d4594_1563084
		dword_5d4594_1563084 = unsafe.Pointer(exit)
		defer func() { dword_5d4594_1563084 = previousExit }()
		wasSaveFlag := noxflags.HasGame(noxflags.GameFlag28)
		noxflags.SetGame(noxflags.GameFlag28)
		defer func() {
			if !wasSaveFlag {
				noxflags.UnsetGame(noxflags.GameFlag28)
			}
		}()
		if !saveCoopGame(common.SaveTmp) {
			e2eError(fmt.Errorf("exit coop save: saveCoopGame failed for exit %p", exit))
			return
		}
		mapName := noxServer.getServerMap()
		for _, path := range []string{
			datapath.Save(common.SaveTmp, mapName, mapName+".map"),
			datapath.Save(common.SaveTmp, common.PlayerFile),
		} {
			info, err := ifs.Stat(path)
			if err != nil || info.Size() == 0 {
				e2eError(fmt.Errorf("exit coop save: invalid output %q: %v", path, err))
				return
			}
		}
		e2eLog.Printf("EXIT COOP SAVE: map=%q exit=%p destination=(%.3f,%.3f)", mapName, exit,
			exitCollideData4DB600(unsafe.Pointer(exit)).DestinationX,
			exitCollideData4DB600(unsafe.Pointer(exit)).DestinationY)
	})
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

func e2eMapBaseName(name string) string {
	name = filepath.Base(strings.TrimSpace(name))
	return strings.ToLower(strings.TrimSuffix(name, filepath.Ext(name)))
}

func (sc *e2eScenario) SwitchMap(mapName, name string) {
	sc.add(0, name, func() {
		if e2eMapBaseName(mapName) == "" {
			e2eError(fmt.Errorf("empty E2E map name"))
			return
		}
		e2e.lavaPlayer = nil
		e2e.lavaHealthBefore = 0
		fileName := mapName
		if filepath.Ext(fileName) == "" {
			fileName += ".map"
		}
		e2eLog.Printf("MAP SWITCH: current=%q requested=%q", legacy.Nox_xxx_mapGetMapName_409B40(), fileName)
		noxServer.SwitchMap(fileName)
	})
}

func (sc *e2eScenario) WaitMap(mapName, name string) {
	want := e2eMapBaseName(mapName)
	sc.addWhen(0, name, 2400, func() bool {
		return legacy.Get_dword_5d4594_1548524() == 0 &&
			e2eMapBaseName(legacy.Nox_xxx_mapGetMapName_409B40()) == want &&
			noxServer.Players.HostUnit() != nil && noxClient.ClientPlayerUnit() != nil && nox_client_isConnected()
	}, func() {
		player := noxServer.Players.HostUnit()
		e2eLog.Printf("MAP READY: map=%q frame=%d player=%p drawable=%p pos=(%.3f,%.3f)",
			legacy.Nox_xxx_mapGetMapName_409B40(), noxServer.Frame(), player,
			noxClient.ClientPlayerUnit(), player.PosVec.X, player.PosVec.Y)
	})
}

func (sc *e2eScenario) WaitForcedCoopAutosave(name string) {
	sc.add(0, name+" resume", func() {
		if !nox_xxx_gameGet_4DB1B0() {
			return
		}
		// A forced E2E map switch does not generate client input while the
		// cooperative map-entry autosave delay is active. Release that input
		// pause and let the normal server save path finish and clear its gate.
		sub_413980(0)
		sub_413A00(0)
		e2eLog.Printf("FORCED MAP AUTOSAVE: resumed frame=%d", noxServer.Frame())
	})
	sc.addWhen(0, name, 1200, func() bool {
		return !nox_xxx_gameGet_4DB1B0()
	}, func() {
		e2eLog.Printf("FORCED MAP AUTOSAVE: completed frame=%d", noxServer.Frame())
	})
}

func (sc *e2eScenario) CallNoxScriptFunction(function, name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return noxServer.Players.HostUnit() != nil && legacy.Get_dword_5d4594_1548524() == 0
	}, func() {
		if function == "" {
			e2eError(fmt.Errorf("empty E2E NoxScript function name"))
			return
		}
		_, index := noxServer.S().NoxScriptVM.FuncByName(function)
		if index < 0 {
			e2eError(fmt.Errorf("NoxScript function %q not found on map %q", function, legacy.Nox_xxx_mapGetMapName_409B40()))
			return
		}
		player := noxServer.Players.HostUnit()
		if err := noxServer.S().NoxScriptVM.CallByIndex(index, player, player); err != nil {
			e2eError(fmt.Errorf("NoxScript function %q failed: %w", function, err))
			return
		}
		e2eLog.Printf("NOXSCRIPT FUNCTION: map=%q function=%q frame=%d", legacy.Nox_xxx_mapGetMapName_409B40(), function, noxServer.Frame())
	})
}

func wizard1UrchinStats() (int, int, *server.Object) {
	var horvath *server.Object
	var urchins, health int
	for obj := noxServer.Objs.First(); obj != nil; obj = obj.Next() {
		if obj.EqualID("Horvath") {
			horvath = obj
		}
		typ := obj.ObjectTypeC()
		if typ == nil || typ.ID() != "Urchin" || obj.Flags().HasAny(object.FlagDead|object.FlagDestroyed) {
			continue
		}
		urchins++
		cur, _ := obj.Health()
		health += cur
	}
	return urchins, health, horvath
}

func (sc *e2eScenario) CaptureWizard1Urchins(name string) {
	sc.add(0, name, func() {
		count, health, horvath := wizard1UrchinStats()
		if horvath == nil || count < 12 {
			e2eError(fmt.Errorf("WIZARD1 setup incomplete: Horvath=%p Urchins=%d", horvath, count))
			return
		}
		sc.wizard1UrchinsBefore, sc.wizard1UrchinHPBefore = count, health
		e2eLog.Printf("WIZARD1 LIGHTNING BASELINE: frame=%d Urchins=%d HP=%d", noxServer.Frame(), count, health)
	})
}

func (sc *e2eScenario) EnterWizard1UrchinSetupTrigger(name string) {
	sc.add(0, name, func() {
		player := noxServer.Players.HostUnit()
		if player == nil {
			e2eError(fmt.Errorf("Wizard 1 player is missing"))
			return
		}
		_, setupIndex := noxServer.S().NoxScriptVM.FuncByName("UrchinSetup")
		if setupIndex < 0 {
			e2eError(fmt.Errorf("WIZARD1 UrchinSetup NoxScript function is missing"))
			return
		}
		var trigger *server.Object
		for obj := noxServer.Objs.First(); obj != nil; obj = obj.Next() {
			if !obj.Class().Has(object.ClassTrigger) || !obj.Flags().Has(object.FlagEnabled) ||
				obj.UpdateDataTrigger().ScriptCollide.Func != int32(setupIndex) {
				continue
			}
			trigger = obj
			break
		}
		if trigger == nil {
			e2eError(fmt.Errorf("WIZARD1 UrchinSetup collision trigger is missing"))
			return
		}
		asObjectS(player).SetPos(trigger.PosVec)
		e2eLog.Printf("WIZARD1 PLAYER ENTERED SETUP TRIGGER: frame=%d id=%d pos=%v", noxServer.Frame(), trigger.ScriptID(), player.PosVec)
	})
}

func (sc *e2eScenario) AssertWizard1UrchinsSpawned(name string) {
	sc.add(0, name, func() {
		count, health, _ := wizard1UrchinStats()
		if count-sc.wizard1UrchinsBefore < 12 || health-sc.wizard1UrchinHPBefore < 80 {
			e2eError(fmt.Errorf("WIZARD1 UrchinSetup trigger: Urchins=%d->%d HP=%d->%d frame=%d",
				sc.wizard1UrchinsBefore, count, sc.wizard1UrchinHPBefore, health, noxServer.Frame()))
			return
		}
		e2eLog.Printf("WIZARD1 URCHINS SPAWNED: frame=%d Urchins=%d->%d HP=%d->%d",
			noxServer.Frame(), sc.wizard1UrchinsBefore, count, sc.wizard1UrchinHPBefore, health)
		sc.wizard1UrchinsBefore, sc.wizard1UrchinHPBefore = count, health
	})
}

func (sc *e2eScenario) AssertWizard1LightningKills(name string) {
	sc.add(0, name, func() {
		count, health, horvath := wizard1UrchinStats()
		if horvath == nil || horvath.Flags().HasAny(object.FlagDead|object.FlagDestroyed) ||
			sc.wizard1UrchinsBefore-count < 5 || sc.wizard1UrchinHPBefore-health < 40 {
			e2eError(fmt.Errorf("WIZARD1 lightning encounter: Horvath=%p Urchins=%d->%d HP=%d->%d frame=%d",
				horvath, sc.wizard1UrchinsBefore, count, sc.wizard1UrchinHPBefore, health, noxServer.Frame()))
			return
		}
		e2eLog.Printf("WIZARD1 LIGHTNING KILLS: frame=%d Urchins=%d->%d HP=%d->%d",
			noxServer.Frame(), sc.wizard1UrchinsBefore, count, sc.wizard1UrchinHPBefore, health)
	})
}

func (sc *e2eScenario) WaitWizard1MultipleLightningHits(name string) {
	hits := make(map[int]uint32)
	sc.addWhen(0, name, 1200, func() bool {
		_, _, horvath := wizard1UrchinStats()
		if horvath == nil {
			return false
		}
		for obj := noxServer.Objs.First(); obj != nil; obj = obj.Next() {
			typ := obj.ObjectTypeC()
			if typ == nil || typ.ID() != "Urchin" || obj.Obj130 != horvath ||
				obj.Frame134 < sc.wizard1LightningStart {
				continue
			}
			hitType := object.DamageType(obj.Field131)
			if hitType != object.DamageElectric && hitType != object.DamageAirborneElectric {
				continue
			}
			hits[obj.ScriptID()] = obj.Frame134
		}
		return len(hits) >= 5
	}, func() {
		e2eLog.Printf("WIZARD1 HORVATH LIGHTNING HITS: frame=%d distinct Urchins=%d", noxServer.Frame(), len(hits))
	})
}

func (sc *e2eScenario) MoveWizard1PlayerNearHorvath(name string) {
	sc.add(0, name, func() {
		player := noxServer.Players.HostUnit()
		if player == nil {
			e2eError(fmt.Errorf("Wizard 1 player is missing"))
			return
		}
		asObjectS(player).SetPos(types.Ptf(2365, 3500))
		sc.wizard1LightningStart = noxServer.Frame()
		e2eLog.Printf("WIZARD1 PLAYER MOVED: frame=%d pos=%v", noxServer.Frame(), player.PosVec)
	})
}

func e2eDoorTileCoordinate4F4CB0(component int32, position float32) int32 {
	offset := component / 2
	scaled := (float64(offset) + float64(position)) * float64(math.Float32frombits(0x3d321643))
	return int32(int64(math.Trunc(scaled)))
}

func (sc *e2eScenario) AssertDoorXferLoaded(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return noxServer.Players.HostUnit() != nil && legacy.Get_dword_5d4594_1548524() == 0
	}, func() {
		xfer := legacy.Get_nox_xxx_XFerDoor_4F4CB0()
		var count int
		var sample *server.Object
		for obj := noxServer.Objs.First(); obj != nil; obj = obj.Next() {
			if obj.Xfer != xfer {
				continue
			}
			count++
			if !obj.Class().Has(object.ClassDoor) || obj.UpdateData == nil {
				e2eError(fmt.Errorf("DoorXfer object is not a native Door: object=%p class=%#x update=%p", obj, uint32(obj.Class()), obj.UpdateData))
				return
			}
			if unsafe.Sizeof(uintptr(0)) == 8 &&
				(uintptr(obj.CObj()) <= math.MaxUint32 || uintptr(obj.UpdateData) <= math.MaxUint32) {
				e2eError(fmt.Errorf("DoorXfer object used a low native address: object=%p update=%p", obj, obj.UpdateData))
				return
			}
			update := obj.UpdateDataDoor()
			if update.CurrentDirection < 0 || update.CurrentDirection >= 32 ||
				update.TargetDirection < 0 || update.TargetDirection >= 32 ||
				update.SyncedDirection < 0 || update.SyncedDirection >= 32 {
				e2eError(fmt.Errorf("DoorXfer directions are outside 0..31: object=%p current=%d target=%d synced=%d",
					obj, update.CurrentDirection, update.TargetDirection, update.SyncedDirection))
				return
			}
			if got := int32(update.FractionalDir) * 32 / 256; got != update.CurrentDirection {
				e2eError(fmt.Errorf("DoorXfer fractional direction mismatch: object=%p fractional=%d current=%d derived=%d",
					obj, update.FractionalDir, update.CurrentDirection, got))
				return
			}
			wantTileX := e2eDoorTileCoordinate4F4CB0(server.DoorDirectionX(update.TargetDirection), obj.PosVec.X)
			wantTileY := e2eDoorTileCoordinate4F4CB0(server.DoorDirectionY(update.TargetDirection), obj.PosVec.Y)
			if update.TileX != wantTileX || update.TileY != wantTileY {
				e2eError(fmt.Errorf("DoorXfer tile mismatch: object=%p tile=(%d,%d) want=(%d,%d) pos=(%.3f,%.3f) target=%d",
					obj, update.TileX, update.TileY, wantTileX, wantTileY, obj.PosVec.X, obj.PosVec.Y, update.TargetDirection))
				return
			}
			if sample == nil {
				sample = obj
			}
		}
		if count == 0 {
			e2eError(fmt.Errorf("map %q contains no object bound to DoorXfer", legacy.Nox_xxx_mapGetMapName_409B40()))
			return
		}
		update := sample.UpdateDataDoor()
		e2eLog.Printf("DOOR XFER LOADED: map=%q count=%d callback=%p object=%p update=%p direction=%d/%d/%d fractional=%d tile=(%d,%d) pointers=native",
			legacy.Nox_xxx_mapGetMapName_409B40(), count, xfer, sample, sample.UpdateData,
			update.CurrentDirection, update.TargetDirection, update.SyncedDirection, update.FractionalDir,
			update.TileX, update.TileY)
	})
}

func (sc *e2eScenario) AssertTriggerXferLoaded(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return noxServer.Players.HostUnit() != nil && legacy.Get_dword_5d4594_1548524() == 0
	}, func() {
		xfer := legacy.Get_nox_xxx_UnitTriggerXfer_4F4E50()
		var count int
		var sample *server.Object
		var scriptDataCount int
		for obj := noxServer.Objs.First(); obj != nil; obj = obj.Next() {
			if obj.Xfer != xfer {
				continue
			}
			count++
			if !obj.Class().Has(object.ClassTrigger) || obj.UpdateData == nil {
				e2eError(fmt.Errorf("TriggerXfer object is not a native Trigger: object=%p class=%#x update=%p", obj, uint32(obj.Class()), obj.UpdateData))
				return
			}
			if obj.Shape.Box.W > 60 || obj.Shape.Box.H > 60 {
				e2eError(fmt.Errorf("TriggerXfer box escaped load clamp: object=%p size=(%.3f,%.3f)", obj, obj.Shape.Box.W, obj.Shape.Box.H))
				return
			}
			wantBox := obj.Shape.Box
			wantBox.Calc()
			if obj.Shape.Box.LeftTop != wantBox.LeftTop ||
				obj.Shape.Box.LeftBottom != wantBox.LeftBottom ||
				obj.Shape.Box.RightTop != wantBox.RightTop ||
				obj.Shape.Box.RightBottom != wantBox.RightBottom {
				e2eError(fmt.Errorf("TriggerXfer box was not recalculated after load: object=%p size=(%.3f,%.3f)", obj, obj.Shape.Box.W, obj.Shape.Box.H))
				return
			}
			if unsafe.Sizeof(uintptr(0)) == 8 &&
				(uintptr(obj.CObj()) <= math.MaxUint32 || uintptr(obj.UpdateData) <= math.MaxUint32) {
				e2eError(fmt.Errorf("TriggerXfer object used a low native address: object=%p update=%p", obj, obj.UpdateData))
				return
			}
			if obj.Field189 != nil {
				scriptDataCount++
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(obj.Field189) <= math.MaxUint32 {
					e2eError(fmt.Errorf("TriggerXfer script data used a low native address: object=%p script=%p", obj, obj.Field189))
					return
				}
			}
			if sample == nil {
				sample = obj
			}
		}
		if count == 0 {
			e2eError(fmt.Errorf("map %q contains no object bound to TriggerXfer", legacy.Nox_xxx_mapGetMapName_409B40()))
			return
		}
		update := sample.UpdateDataTrigger()
		e2eLog.Printf("TRIGGER XFER LOADED: map=%q count=%d scripts=%d callback=%p object=%p update=%p script=%p size=(%.3f,%.3f) flags=%#x state=%d/%d class=%#x/%#x team=%d/%d colors=%v pointers=native",
			legacy.Nox_xxx_mapGetMapName_409B40(), count, scriptDataCount, xfer, sample, sample.UpdateData, sample.Field189,
			sample.Shape.Box.W, sample.Shape.Box.H, update.Flags, update.State, update.Field9,
			update.ClassInclude, update.ClassExclude, update.TeamInclude, update.TeamExclude, update.Colors)
	})
}

func (sc *e2eScenario) AssertHoleXferLoaded(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return noxServer.Players.HostUnit() != nil && legacy.Get_dword_5d4594_1548524() == 0
	}, func() {
		xfer := legacy.Get_nox_xxx_XFerHole_4F51D0()
		var count int
		var sample *server.Object
		var scriptDataCount int
		var destinationCount int
		for obj := noxServer.Objs.First(); obj != nil; obj = obj.Next() {
			if obj.Xfer != xfer {
				continue
			}
			count++
			if !obj.Class().Has(object.ClassHole) || obj.CollideData == nil {
				e2eError(fmt.Errorf("HoleXfer object is not a native Hole: object=%p class=%#x collide=%p", obj, uint32(obj.Class()), obj.CollideData))
				return
			}
			if unsafe.Sizeof(uintptr(0)) == 8 &&
				(uintptr(obj.CObj()) <= math.MaxUint32 || uintptr(obj.CollideData) <= math.MaxUint32) {
				e2eError(fmt.Errorf("HoleXfer object used a low native address: object=%p collide=%p", obj, obj.CollideData))
				return
			}
			if obj.Field189 != nil {
				scriptDataCount++
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(obj.Field189) <= math.MaxUint32 {
					e2eError(fmt.Errorf("HoleXfer script data used a low native address: object=%p script=%p", obj, obj.Field189))
					return
				}
			}
			data := (*server.HoleCollideData)(obj.CollideData)
			if data.DestinationX != 0 || data.DestinationY != 0 {
				destinationCount++
			}
			if sample == nil {
				sample = obj
			}
		}
		if count == 0 {
			e2eError(fmt.Errorf("map %q contains no object bound to HoleXfer", legacy.Nox_xxx_mapGetMapName_409B40()))
			return
		}
		if destinationCount == 0 {
			e2eError(fmt.Errorf("map %q HoleXfer objects contain no destination payload", legacy.Nox_xxx_mapGetMapName_409B40()))
			return
		}
		data := (*server.HoleCollideData)(sample.CollideData)
		e2eLog.Printf("HOLE XFER LOADED: map=%q count=%d destinations=%d scripts=%d callback=%p object=%p collide=%p script=%p callback_data=%#x/%d destination=(%d,%d) extent=%d netcode=%d reserved=%d field24=%#x pointers=native",
			legacy.Nox_xxx_mapGetMapName_409B40(), count, destinationCount, scriptDataCount, xfer, sample, sample.CollideData, sample.Field189,
			data.Script.Flags, data.Script.Func, data.DestinationX, data.DestinationY,
			data.DestinationExtent, data.DestinationNetCode, data.Reserved22, data.Field24)
	})
}

func (sc *e2eScenario) AssertTransporterXferLoaded(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return noxServer.Players.HostUnit() != nil && legacy.Get_dword_5d4594_1548524() == 0
	}, func() {
		xfer := legacy.Get_nox_xxx_XFerTransporter_4F5300()
		var count int
		var linkedCount int
		var sample *server.Object
		var sampleTarget *server.Object
		for obj := noxServer.Objs.First(); obj != nil; obj = obj.Next() {
			if obj.Xfer != xfer {
				continue
			}
			count++
			if !obj.Class().Has(object.ClassTransporter) || obj.UpdateData == nil {
				e2eError(fmt.Errorf("TransporterXfer object is not a native Transporter: object=%p class=%#x update=%p", obj, uint32(obj.Class()), obj.UpdateData))
				return
			}
			if unsafe.Sizeof(uintptr(0)) == 8 &&
				(uintptr(obj.CObj()) <= math.MaxUint32 || uintptr(obj.UpdateData) <= math.MaxUint32) {
				e2eError(fmt.Errorf("TransporterXfer object used a low native address: object=%p update=%p", obj, obj.UpdateData))
				return
			}
			data := obj.UpdateDataTransporter()
			if data.TargetPE32 != 0 {
				e2eError(fmt.Errorf("TransporterXfer retained a PE32 target pointer: object=%p target_pe32=%#x extent=%d", obj, data.TargetPE32, data.TargetExtent))
				return
			}
			target := obj.TransporterTargetFor(data)
			if target != nil {
				linkedCount++
				if !target.Class().Has(object.ClassTransporter) || target.Extent != data.TargetExtent {
					e2eError(fmt.Errorf("TransporterXfer native target is inconsistent: object=%p target=%p class=%#x extent=%d/%d", obj, target, uint32(target.Class()), target.Extent, data.TargetExtent))
					return
				}
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(target.CObj()) <= math.MaxUint32 {
					e2eError(fmt.Errorf("TransporterXfer target used a low native address: object=%p target=%p", obj, target))
					return
				}
			} else if data.TargetExtent != 0 {
				e2eError(fmt.Errorf("TransporterXfer extent has no native target: object=%p extent=%d", obj, data.TargetExtent))
				return
			}
			if sample == nil || (sampleTarget == nil && target != nil) {
				sample = obj
				sampleTarget = target
			}
		}
		if count == 0 {
			e2eError(fmt.Errorf("map %q contains no object bound to TransporterXfer", legacy.Nox_xxx_mapGetMapName_409B40()))
			return
		}
		if linkedCount == 0 {
			e2eError(fmt.Errorf("map %q TransporterXfer objects contain no native target links", legacy.Nox_xxx_mapGetMapName_409B40()))
			return
		}
		data := sample.UpdateDataTransporter()
		e2eLog.Printf("TRANSPORTER XFER LOADED: map=%q count=%d linked=%d callback=%p object=%p update=%p target=%p target_pe32=%#x extent=%d pointers=native",
			legacy.Nox_xxx_mapGetMapName_409B40(), count, linkedCount, xfer, sample, sample.UpdateData, sampleTarget,
			data.TargetPE32, data.TargetExtent)
	})
}

func (sc *e2eScenario) AssertElevatorXferLoaded(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return noxServer.Players.HostUnit() != nil && legacy.Get_dword_5d4594_1548524() == 0
	}, func() {
		xfer := legacy.Get_nox_xxx_XFerElevator_4F53D0()
		shaftXfer := legacy.Get_nox_xxx_XFerElevatorShaft_4F54A0()
		var count int
		var linkedCount int
		var sample *server.Object
		var sampleShaft *server.Object
		for obj := noxServer.Objs.First(); obj != nil; obj = obj.Next() {
			if obj.Xfer != xfer {
				continue
			}
			count++
			if !obj.Class().Has(object.ClassElevator) || obj.UpdateData == nil {
				e2eError(fmt.Errorf("ElevatorXfer object is not a native Elevator: object=%p class=%#x update=%p", obj, uint32(obj.Class()), obj.UpdateData))
				return
			}
			if unsafe.Sizeof(uintptr(0)) == 8 &&
				(uintptr(obj.CObj()) <= math.MaxUint32 || uintptr(obj.UpdateData) <= math.MaxUint32) {
				e2eError(fmt.Errorf("ElevatorXfer object used a low native address: object=%p update=%p", obj, obj.UpdateData))
				return
			}
			data := obj.UpdateDataElevator()
			if data.Field_1 != 0 {
				e2eError(fmt.Errorf("ElevatorXfer retained a PE32 shaft pointer: object=%p shaft_pe32=%#x extent=%d", obj, data.Field_1, data.Field_2))
				return
			}
			shaft := obj.ElevatorLinkFor(obj.UpdateData)
			if shaft != nil {
				linkedCount++
				if shaft.Xfer != shaftXfer {
					e2eError(fmt.Errorf("ElevatorShaftXfer object has the wrong callback: object=%p shaft=%p callback=%p want=%p",
						obj, shaft, shaft.Xfer, shaftXfer))
					return
				}
				if !shaft.Class().Has(object.ClassElevatorShaft) || shaft.UpdateData == nil || shaft.Extent != data.Field_2 {
					e2eError(fmt.Errorf("ElevatorXfer native shaft is inconsistent: object=%p shaft=%p class=%#x update=%p extent=%d/%d",
						obj, shaft, uint32(shaft.Class()), shaft.UpdateData, shaft.Extent, data.Field_2))
					return
				}
				if shaft.ElevatorLinkFor(shaft.UpdateData) != obj {
					e2eError(fmt.Errorf("ElevatorXfer shaft link is not reciprocal: object=%p shaft=%p back=%p",
						obj, shaft, shaft.ElevatorLinkFor(shaft.UpdateData)))
					return
				}
				shaftData := shaft.UpdateDataElevatorShaft()
				if shaftData.Field_1 != 0 {
					e2eError(fmt.Errorf("ElevatorShaftXfer retained a PE32 elevator pointer: object=%p shaft=%p elevator_pe32=%#x",
						obj, shaft, shaftData.Field_1))
					return
				}
				if unsafe.Sizeof(uintptr(0)) == 8 &&
					(uintptr(shaft.CObj()) <= math.MaxUint32 || uintptr(shaft.UpdateData) <= math.MaxUint32) {
					e2eError(fmt.Errorf("ElevatorXfer shaft used a low native address: object=%p shaft=%p update=%p", obj, shaft, shaft.UpdateData))
					return
				}
			} else if data.Field_2 != 0 {
				e2eError(fmt.Errorf("ElevatorXfer extent has no native shaft: object=%p extent=%d", obj, data.Field_2))
				return
			}
			if sample == nil || (sampleShaft == nil && shaft != nil) {
				sample = obj
				sampleShaft = shaft
			}
		}
		if count == 0 {
			e2eError(fmt.Errorf("map %q contains no object bound to ElevatorXfer", legacy.Nox_xxx_mapGetMapName_409B40()))
			return
		}
		if linkedCount == 0 {
			e2eError(fmt.Errorf("map %q ElevatorXfer objects contain no native shaft links", legacy.Nox_xxx_mapGetMapName_409B40()))
			return
		}
		data := sample.UpdateDataElevator()
		shaftData := sampleShaft.UpdateDataElevatorShaft()
		e2eLog.Printf("ELEVATOR XFER LOADED: map=%q count=%d linked=%d callback=%p shaft_callback=%p elevator=%p update=%p shaft=%p shaft_update=%p link_pe32=%#x shaft_link_pe32=%#x extent=%d pointers=native",
			legacy.Nox_xxx_mapGetMapName_409B40(), count, linkedCount, xfer, shaftXfer, sample, sample.UpdateData,
			sampleShaft, sampleShaft.UpdateData, data.Field_1, shaftData.Field_1, data.Field_2)
	})
}

func e2eFindLavaTile() (types.Pointf, bool) {
	// GAME.EXE 00411160 accepts only the interior 128x128 tile grid. Sampling
	// every half-cell visits both halves of the diamond floor representation.
	const (
		minimum = float32(92)
		maximum = float32(5750)
		step    = float32(23)
	)
	for y := minimum; y <= maximum; y += step {
		for x := minimum; x <= maximum; x += step {
			pos := types.Ptf(x, y)
			if legacy.Nox_xxx_tileNFromPoint_411160(pos) == 6 {
				return pos, true
			}
		}
	}
	return types.Pointf{}, false
}

func (sc *e2eScenario) PlacePlayerOnLava(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		player := noxServer.Players.HostUnit()
		return player != nil && player.HealthData != nil && player.HealthData.Cur != 0 &&
			!player.Flags().HasAny(object.FlagDead|object.FlagDestroyed) &&
			!player.HasEnchant(server.ENCHANT_INVULNERABLE)
	}, func() {
		player := noxServer.Players.HostUnit()
		pos, ok := e2eFindLavaTile()
		if !ok {
			e2eError(fmt.Errorf("map %q contains no tile index 6", legacy.Nox_xxx_mapGetMapName_409B40()))
			return
		}
		if player.Damage == nil {
			e2eError(fmt.Errorf("host player has no damage callback"))
			return
		}
		e2e.lavaPlayer = player
		e2e.lavaOriginalPos = player.PosVec
		e2e.lavaPos = pos
		e2e.lavaHealthBefore = player.HealthData.Cur
		e2e.lavaFrameBefore = noxServer.Frame()
		flagsBefore := player.ObjFlags
		queuedBefore := player.Field116
		asObjectS(player).SetPos(pos)
		player.NewPos = pos
		player.PrevPos = pos
		player.VelVec = types.Pointf{}
		player.ForceVec = types.Pointf{}
		player.Pos24 = types.Pointf{}
		// The scripted teleport itself has no movement input. Queue the player as
		// the ordinary movement scheduler does before 005118A0, so this fixture
		// exercises the real 005118A0 -> 00548630 -> 00548740 collision dispatch
		// instead of calling the damage callback directly.
		legacy.Nox_xxx_unitHasCollideOrUpdateFn_537610(player)
		tileName := ""
		if tiles := legacy.Get_nox_tile_defs_arr(); len(tiles) > 6 {
			tileName = tiles[6].Name()
		}
		e2eLog.Printf("LAVA CONTACT ARMED: map=%q tile=6/%q pos=(%.3f,%.3f) player=%p update=%p collide=%p damage=%p health=%d/%d frame=%d flags=%#x->%#x buffs=%#x active=%d queue=%#x->%#x pending=%d tiles=%d/%d",
			legacy.Nox_xxx_mapGetMapName_409B40(), tileName, pos.X, pos.Y, player, player.Update, player.Collide, player.Damage,
			player.HealthData.Cur, player.HealthData.Max, e2e.lavaFrameBefore, uint32(flagsBefore), uint32(player.ObjFlags),
			player.Buffs, player.IsUpdatable, queuedBefore, player.Field116, legacy.Get_dword_5d4594_2488604(),
			legacy.Nox_xxx_tileNFromPoint_411160(player.PosVec), legacy.Nox_xxx_tileNFromPoint_411160(player.NewPos))
	})
}

func (sc *e2eScenario) AssertPlayerLavaDamage(name string) {
	sc.add(0, name, func() {
		player := e2e.lavaPlayer
		if player == nil || player.HealthData == nil {
			e2eError(fmt.Errorf("LAVA player fixture is unavailable: player=%p", player))
			return
		}
		after := player.HealthData.Cur
		frame := noxServer.Frame()
		posBeforeRestore := player.PosVec
		newPosBeforeRestore := player.NewPos
		flagsBeforeRestore := player.ObjFlags
		queuedBeforeRestore := player.Field116
		pendingBeforeRestore := legacy.Get_dword_5d4594_2488604()
		tilePosBeforeRestore := legacy.Nox_xxx_tileNFromPoint_411160(posBeforeRestore)
		tileNewBeforeRestore := legacy.Nox_xxx_tileNFromPoint_411160(newPosBeforeRestore)
		update := player.UpdateDataPlayer()
		markerBeforeRestore := update.Field75
		markerStateBeforeRestore := update.Field76
		damageTypeBeforeRestore := player.Field131
		playerStateBeforeRestore := uint32(0)
		if update.Player != nil {
			playerStateBeforeRestore = update.Player.Field3680
		}
		asObjectS(player).SetPos(e2e.lavaOriginalPos)
		player.NewPos = e2e.lavaOriginalPos
		player.PrevPos = e2e.lavaOriginalPos
		player.VelVec = types.Pointf{}
		player.ForceVec = types.Pointf{}
		player.Pos24 = types.Pointf{}
		if after >= e2e.lavaHealthBefore {
			e2eError(fmt.Errorf("LAVA did not reduce player health: before=%d after=%d frames=%d->%d pos=%v new=%v tiles=%d/%d flags=%#x active=%d queue=%#x pending=%d marker=%#x/%#x type=%d player-state=%#x",
				e2e.lavaHealthBefore, after, e2e.lavaFrameBefore, frame, posBeforeRestore, newPosBeforeRestore,
				tilePosBeforeRestore, tileNewBeforeRestore, uint32(flagsBeforeRestore), player.IsUpdatable,
				queuedBeforeRestore, pendingBeforeRestore, markerBeforeRestore, markerStateBeforeRestore,
				damageTypeBeforeRestore, playerStateBeforeRestore))
			return
		}
		if after == 0 || player.Flags().HasAny(object.FlagDead|object.FlagDestroyed) {
			e2eError(fmt.Errorf("LAVA fixture killed player: health=%d flags=%#x", after, uint32(player.Flags())))
			return
		}
		if update.Field76 != 2 || update.Field75 != math.Float32bits(float32(object.DamageLava)) ||
			player.Obj130 != nil || player.Field131 != uint32(object.DamageLava) || player.Pos132 != (types.Pointf{}) {
			e2eError(fmt.Errorf("LAVA metadata = marker:%#x/%#x source:%p type:%d hit-pos:%v",
				update.Field75, update.Field76, player.Obj130, player.Field131, player.Pos132))
			return
		}
		e2eLog.Printf("LAVA DAMAGE: player=%p health=%d->%d damage=%d frames=%d->%d type=%d restored=(%.3f,%.3f)",
			player, e2e.lavaHealthBefore, after, e2e.lavaHealthBefore-after,
			e2e.lavaFrameBefore, frame, player.Field131, e2e.lavaOriginalPos.X, e2e.lavaOriginalPos.Y)
	})
}

func (sc *e2eScenario) ArmPlayerPoison(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		player := noxServer.Players.HostUnit()
		return player != nil && player.HealthData != nil && player.HealthData.Cur > 0 &&
			!player.Flags().HasAny(object.FlagDead|object.FlagDestroyed) &&
			!player.HasEnchant(server.ENCHANT_INVULNERABLE)
	}, func() {
		player := noxServer.Players.HostUnit()
		if player.Damage == nil || player.Poison540 != 0 {
			e2eError(fmt.Errorf("player poison fixture is unavailable: damage=%p poison=%d", player.Damage, player.Poison540))
			return
		}
		e2e.poisonPlayer = player
		e2e.poisonHealthBefore = player.HealthData.Cur
		e2e.poisonFrameBefore = noxServer.Frame()
		// A value of 2 reaches the normal poison-tick loop after its 60-frame
		// grace period, then deals one point every 64 frames.
		noxServer.S().SetPoison4EEA90(player, 2)
		e2eLog.Printf("PLAYER POISON ARMED: player=%p health=%d frame=%d poison=%d damage=%p",
			player, e2e.poisonHealthBefore, e2e.poisonFrameBefore, player.Poison540, player.Damage)
	})
}

func (sc *e2eScenario) AssertPlayerPoisonDamage(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		player := e2e.poisonPlayer
		return player != nil && player.HealthData != nil && player.HealthData.Cur < e2e.poisonHealthBefore
	}, func() {
		player := e2e.poisonPlayer
		after := player.HealthData.Cur
		noxServer.S().SetPoison4EEA90(player, 0)
		update := player.UpdateDataPlayer()
		if after != e2e.poisonHealthBefore-1 || player.Flags().HasAny(object.FlagDead|object.FlagDestroyed) ||
			update.Field76 != 2 || update.Field75 != math.Float32bits(float32(object.DamagePoison)) ||
			player.Obj130 != nil || player.Field131 != uint32(object.DamagePoison) || player.Pos132 != (types.Pointf{}) {
			e2eError(fmt.Errorf("POISON tick state: health=%d->%d flags=%#x marker=%#x/%#x source=%p type=%d hit-pos=%v",
				e2e.poisonHealthBefore, after, uint32(player.Flags()), update.Field75, update.Field76,
				player.Obj130, player.Field131, player.Pos132))
			return
		}
		e2eLog.Printf("PLAYER POISON DAMAGE: player=%p health=%d->%d frames=%d->%d type=%d poison=%d",
			player, e2e.poisonHealthBefore, after, e2e.poisonFrameBefore, noxServer.Frame(), player.Field131, player.Poison540)
	})
}

func (sc *e2eScenario) ArmOvalShield(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		player := noxServer.Players.HostUnit()
		return player != nil && player.HealthData != nil && player.HealthData.Cur > 0 &&
			!player.Flags().HasAny(object.FlagDead|object.FlagDestroyed)
	}, func() {
		player := noxServer.Players.HostUnit()
		if player.HasEnchant(server.ENCHANT_REFLECTIVE_SHIELD) {
			e2eError(fmt.Errorf("OVAL SHIELD fixture already has the reflective-shield buff"))
			return
		}
		arg := &server.SpellAcceptArg{Obj: player, Pos: player.PosVec}
		if !noxServer.spells.duration.New(spell.SPELL_OVAL_SHIELD, player, player, player, arg, 1,
			legacy.Get_sub_531490(), legacy.Get_sub_5314F0(), legacy.Get_sub_531560(), 600) {
			e2eError(fmt.Errorf("OVAL SHIELD duration creation failed for player %p", player))
			return
		}
		record := noxServer.Spells.Dur.List
		if record == nil || record.Spell != uint32(spell.SPELL_OVAL_SHIELD) ||
			record.Caster16 != player || record.Target48 != player ||
			record.Create != legacy.Get_sub_531490() || record.Update != legacy.Get_sub_5314F0() ||
			record.Destroy != legacy.Get_sub_531560() ||
			!player.HasEnchant(server.ENCHANT_REFLECTIVE_SHIELD) {
			e2eError(fmt.Errorf("OVAL SHIELD creation state: record=%p player=%p buff=%t", record, player,
				player.HasEnchant(server.ENCHANT_REFLECTIVE_SHIELD)))
			return
		}
		// The Linux SIGSEGV interpreted the old PE32 Target48 offset as this
		// binary32 coordinate, then dereferenced 0x3fdcccdc. The target stays
		// valid at the native-width field while real game ticks call Update.
		record.Pos.X = math.Float32frombits(0x3fdccccc)
		e2e.ovalShieldPlayer = player
		e2e.ovalShieldRecord = record
		e2e.ovalShieldFrameBefore = noxServer.Frame()
		e2eLog.Printf("OVAL SHIELD ARMED: record=%p target=%p update=%p pos-bits=%#x frame=%d buff=%#x",
			record, record.Target48, record.Update, math.Float32bits(record.Pos.X), e2e.ovalShieldFrameBefore, player.Buffs)
	})
}

func (sc *e2eScenario) AssertOvalShieldUpdate(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return e2e.ovalShieldRecord != nil && noxServer.Frame() >= e2e.ovalShieldFrameBefore+3
	}, func() {
		record, player := e2e.ovalShieldRecord, e2e.ovalShieldPlayer
		found := false
		for cur := noxServer.Spells.Dur.List; cur != nil; cur = cur.Next {
			if cur == record {
				found = true
				break
			}
		}
		if !found {
			e2eError(fmt.Errorf("OVAL SHIELD duration disappeared before update: record=%p frames=%d->%d",
				record, e2e.ovalShieldFrameBefore, noxServer.Frame()))
			return
		}
		if record.Target48 != player || record.Update != legacy.Get_sub_5314F0() ||
			math.Float32bits(record.Pos.X) != 0x3fdccccc || record.Flags88&1 != 0 ||
			!player.HasEnchant(server.ENCHANT_REFLECTIVE_SHIELD) {
			e2eError(fmt.Errorf("OVAL SHIELD update state: found=%t record=%p target=%p flags=%#x pos-bits=%#x buff=%t",
				found, record, record.Target48, record.Flags88, math.Float32bits(record.Pos.X),
				player.HasEnchant(server.ENCHANT_REFLECTIVE_SHIELD)))
			return
		}
		e2eLog.Printf("OVAL SHIELD UPDATED: record=%p target=%p frames=%d->%d pos-bits=%#x buff=%#x",
			record, player, e2e.ovalShieldFrameBefore, noxServer.Frame(), math.Float32bits(record.Pos.X), player.Buffs)
	})
}

func (sc *e2eScenario) AssertShieldDamage(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		player := noxServer.Players.HostUnit()
		return player != nil && player.HealthData != nil && player.HealthData.Cur > 4 &&
			!player.Flags().HasAny(object.FlagDead|object.FlagDestroyed) &&
			!player.HasEnchant(server.ENCHANT_INVULNERABLE)
	}, func() {
		player := noxServer.Players.HostUnit()
		if player.HasEnchant(server.ENCHANT_SHIELD) {
			e2eError(fmt.Errorf("SHIELD damage fixture already has the Shield buff"))
			return
		}
		arg := &server.SpellAcceptArg{Obj: player, Pos: player.PosVec}
		if !noxServer.spells.duration.New(spell.SPELL_SHIELD, player, player, player, arg, 1,
			legacy.Get_nox_xxx_castShield1_52F5A0(), legacy.Get_sub_52F650(), legacy.Get_sub_52F670(), 600) {
			e2eError(fmt.Errorf("SHIELD duration creation failed for player %p", player))
			return
		}
		var record *server.DurSpell
		for cur := noxServer.Spells.Dur.List; cur != nil; cur = cur.Next {
			if cur.Spell == uint32(spell.SPELL_SHIELD) && cur.Target48 == player {
				record = cur
				break
			}
		}
		shieldHealth := int32(0)
		if record != nil {
			shieldHealth = record.Field72
		}
		if record == nil || shieldHealth <= 2 || !player.HasEnchant(server.ENCHANT_SHIELD) {
			e2eError(fmt.Errorf("SHIELD creation state: record=%p health=%d player=%p buff=%t",
				record, shieldHealth, player, player.HasEnchant(server.ENCHANT_SHIELD)))
			return
		}
		if unsafe.Sizeof(uintptr(0)) == 8 &&
			(uintptr(player.CObj()) <= math.MaxUint32 || uintptr(record.C()) <= math.MaxUint32) {
			e2eError(fmt.Errorf("SHIELD fixture used a low native address: player=%p record=%p", player, record))
			return
		}

		beforeHP, beforeShield := player.HealthData.Cur, record.Field72
		if !player.CallDamage(nil, nil, 4, object.DamageLava) {
			e2eError(fmt.Errorf("SHIELD source-less LAVA damage returned false: player=%p record=%p", player, record))
			return
		}
		afterHP, afterShield := player.HealthData.Cur, record.Field72
		if afterHP != beforeHP-2 || afterShield != beforeShield-2 ||
			!player.HasEnchant(server.ENCHANT_SHIELD) || player.Obj130 != nil ||
			player.Field131 != uint32(object.DamageLava) || player.Pos132 != (types.Pointf{}) {
			e2eError(fmt.Errorf("SHIELD damage state: hp=%d->%d shield=%d->%d buff=%t source=%p type=%d hit-pos=%v",
				beforeHP, afterHP, beforeShield, afterShield, player.HasEnchant(server.ENCHANT_SHIELD),
				player.Obj130, player.Field131, player.Pos132))
			return
		}
		e2eLog.Printf("SHIELD DAMAGE: player=%p record=%p hp=%d->%d shield=%d->%d type=%d pointers=native",
			player, record, beforeHP, afterHP, beforeShield, afterShield, player.Field131)
	})
}

func (sc *e2eScenario) ArmChannelLife(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		player := noxServer.Players.HostUnit()
		return player != nil && player.HealthData != nil && player.HealthData.Cur > 15 &&
			player.Class().Has(object.ClassPlayer) && player.UpdateDataPlayer() != nil &&
			player.UpdateDataPlayer().ManaMax > 0 &&
			!player.Flags().HasAny(object.FlagDead|object.FlagDestroyed)
	}, func() {
		player := noxServer.Players.HostUnit()
		update := player.UpdateDataPlayer()
		update.ManaPrev = update.ManaCur
		update.ManaCur = 0
		arg := &server.SpellAcceptArg{Obj: player, Pos: player.PosVec}
		if !noxServer.spells.duration.New(spell.SPELL_CHANNEL_LIFE, player, player, player, arg, 1,
			nil, legacy.Get_sub_52F460(), nil, 600) {
			e2eError(fmt.Errorf("CHANNEL LIFE duration creation failed for player %p", player))
			return
		}
		record := noxServer.Spells.Dur.List
		if record == nil || record.Spell != uint32(spell.SPELL_CHANNEL_LIFE) ||
			record.Caster16 != player || record.Target48 != player ||
			record.Flag20 != 0 || record.Update != legacy.Get_sub_52F460() {
			e2eError(fmt.Errorf("CHANNEL LIFE creation state: record=%p player=%p", record, player))
			return
		}
		record.Pos.X = math.Float32frombits(0x3fdccccc)
		e2e.channelLifePlayer = player
		e2e.channelLifeRecord = record
		e2e.channelLifeFrame = noxServer.Frame()
		e2e.channelLifeHP = player.HealthData.Cur
		e2eLog.Printf("CHANNEL LIFE ARMED: record=%p target=%p caster=%p frame=%d hp=%d mana=%d/%d pos-bits=%#x",
			record, player, player, e2e.channelLifeFrame, e2e.channelLifeHP,
			update.ManaCur, update.ManaMax, math.Float32bits(record.Pos.X))
	})
}

func (sc *e2eScenario) AssertChannelLifeUpdate(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return e2e.channelLifeRecord != nil && noxServer.Frame() >= e2e.channelLifeFrame+10
	}, func() {
		record, player := e2e.channelLifeRecord, e2e.channelLifePlayer
		found := false
		for cur := noxServer.Spells.Dur.List; cur != nil; cur = cur.Next {
			if cur == record {
				found = true
				break
			}
		}
		update := player.UpdateDataPlayer()
		if !found || record.Target48 != player || record.Caster16 != player ||
			record.Update != legacy.Get_sub_52F460() || math.Float32bits(record.Pos.X) != 0x3fdccccc ||
			player.HealthData.Cur >= e2e.channelLifeHP || update.ManaCur == 0 {
			e2eError(fmt.Errorf("CHANNEL LIFE update state: found=%t record=%p hp=%d->%d mana=%d/%d pos-bits=%#x",
				found, record, e2e.channelLifeHP, player.HealthData.Cur,
				update.ManaCur, update.ManaMax, math.Float32bits(record.Pos.X)))
			return
		}
		e2eLog.Printf("CHANNEL LIFE UPDATED: record=%p target=%p frames=%d->%d hp=%d->%d mana=%d/%d fraction=%g",
			record, player, e2e.channelLifeFrame, noxServer.Frame(), e2e.channelLifeHP,
			player.HealthData.Cur, update.ManaCur, update.ManaMax,
			math.Float32frombits(uint32(record.Field72)))
		noxServer.Spells.Dur.CancelSpell(record)
	})
}

func (sc *e2eScenario) ArmFirewalk(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		player := noxServer.Players.HostUnit()
		return player != nil && player.Class().Has(object.ClassPlayer) &&
			!player.Flags().HasAny(object.FlagDead|object.FlagDestroyed)
	}, func() {
		player := noxServer.Players.HostUnit()
		arg := &server.SpellAcceptArg{Obj: player, Pos: player.PosVec}
		if !noxServer.spells.duration.New(spell.SPELL_FIREWALK, player, player, player, arg, 4,
			nil, legacy.Get_nox_xxx_firewalkTick_52ED40(), nil, 600) {
			e2eError(fmt.Errorf("FIREWALK duration creation failed for player %p", player))
			return
		}
		record := noxServer.Spells.Dur.List
		if record == nil || record.Spell != uint32(spell.SPELL_FIREWALK) ||
			record.Target48 != player || record.Update != legacy.Get_nox_xxx_firewalkTick_52ED40() ||
			record.Frame60 != record.Frame64 {
			e2eError(fmt.Errorf("FIREWALK creation state: record=%p player=%p", record, player))
			return
		}
		record.Pos.X = math.Float32frombits(0x3fdccccc)
		e2e.firewalkPlayer = player
		e2e.firewalkRecord = record
		e2e.firewalkFrame = noxServer.Frame()
		e2eLog.Printf("FIREWALK ARMED: record=%p target=%p frame=%d pos-bits=%#x",
			record, player, e2e.firewalkFrame, math.Float32bits(record.Pos.X))
	})
}

func (sc *e2eScenario) AssertFirewalkUpdate(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return e2e.firewalkRecord != nil && noxServer.Frame() >= e2e.firewalkFrame+4
	}, func() {
		record, player := e2e.firewalkRecord, e2e.firewalkPlayer
		found := false
		for cur := noxServer.Spells.Dur.List; cur != nil; cur = cur.Next {
			if cur == record {
				found = true
				break
			}
		}
		if !found || record.Target48 != player || record.Frame64 != record.Frame60+1 ||
			record.Update != legacy.Get_nox_xxx_firewalkTick_52ED40() ||
			math.Float32bits(record.Pos.X) != 0x3fdccccc ||
			math.Float32frombits(uint32(record.Field72)) != player.PosVec.X ||
			math.Float32frombits(uint32(record.Field76)) != player.PosVec.Y {
			e2eError(fmt.Errorf("FIREWALK update state: found=%t record=%p target=%p frame=%d/%d pos-bits=%#x previous=(%g,%g) player=%v",
				found, record, record.Target48, record.Frame60, record.Frame64,
				math.Float32bits(record.Pos.X), math.Float32frombits(uint32(record.Field72)),
				math.Float32frombits(uint32(record.Field76)), player.PosVec))
			return
		}
		e2eLog.Printf("FIREWALK UPDATED: record=%p target=%p frames=%d->%d previous=(%g,%g)",
			record, player, e2e.firewalkFrame, noxServer.Frame(),
			math.Float32frombits(uint32(record.Field72)), math.Float32frombits(uint32(record.Field76)))
		noxServer.Spells.Dur.CancelSpell(record)
	})
}

func (sc *e2eScenario) ArmGreaterHeal(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		player := noxServer.Players.HostUnit()
		return player != nil && player.Class().Has(object.ClassPlayer) &&
			player.HealthData != nil && player.HealthData.Max > 15 &&
			player.UpdateDataPlayer() != nil && player.UpdateDataPlayer().ManaMax > 0 &&
			!player.Flags().HasAny(object.FlagDead|object.FlagDestroyed)
	}, func() {
		player := noxServer.Players.HostUnit()
		player.HealthData.Cur = player.HealthData.Max - 15
		update := player.UpdateDataPlayer()
		update.ManaCur = update.ManaMax
		update.ManaPrev = update.ManaCur
		arg := &server.SpellAcceptArg{Obj: player, Pos: player.PosVec}
		if !noxServer.spells.duration.New(spell.SPELL_GREATER_HEAL, player, player, player, arg, 1,
			legacy.Get_sub_52F220(), legacy.Get_sub_52F2E0(), nil, 600) {
			e2eError(fmt.Errorf("GREATER HEAL duration creation failed for player %p", player))
			return
		}
		record := noxServer.Spells.Dur.List
		if record == nil || record.Spell != uint32(spell.SPELL_GREATER_HEAL) ||
			record.Caster16 != player || record.Target48 != player ||
			record.Create != legacy.Get_sub_52F220() || record.Update != legacy.Get_sub_52F2E0() {
			e2eError(fmt.Errorf("GREATER HEAL creation state: record=%p player=%p", record, player))
			return
		}
		record.Pos.X = math.Float32frombits(0x3fdccccc)
		e2e.greaterHealPlayer = player
		e2e.greaterHealRecord = record
		e2e.greaterHealFrame = noxServer.Frame()
		e2e.greaterHealHP = player.HealthData.Cur
		e2e.greaterHealMana = update.ManaCur
		e2eLog.Printf("GREATER HEAL ARMED: record=%p target=%p frame=%d hp=%d/%d mana=%d/%d pos-bits=%#x",
			record, player, e2e.greaterHealFrame, e2e.greaterHealHP, player.HealthData.Max,
			e2e.greaterHealMana, update.ManaMax, math.Float32bits(record.Pos.X))
	})
}

func (sc *e2eScenario) AssertGreaterHealUpdate(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return e2e.greaterHealRecord != nil && noxServer.Frame() >= e2e.greaterHealFrame+3
	}, func() {
		record, player := e2e.greaterHealRecord, e2e.greaterHealPlayer
		found := false
		for cur := noxServer.Spells.Dur.List; cur != nil; cur = cur.Next {
			if cur == record {
				found = true
				break
			}
		}
		update := player.UpdateDataPlayer()
		if !found || record.Target48 != player || record.Caster16 != player ||
			record.Update != legacy.Get_sub_52F2E0() || math.Float32bits(record.Pos.X) != 0x3fdccccc ||
			update.ManaCur >= e2e.greaterHealMana || player.HealthData.Cur < e2e.greaterHealHP {
			e2eError(fmt.Errorf("GREATER HEAL update state: found=%t record=%p hp=%d->%d mana=%d->%d pos-bits=%#x",
				found, record, e2e.greaterHealHP, player.HealthData.Cur,
				e2e.greaterHealMana, update.ManaCur, math.Float32bits(record.Pos.X)))
			return
		}
		e2eLog.Printf("GREATER HEAL UPDATED: record=%p target=%p frames=%d->%d hp=%d->%d mana=%d->%d fraction=%g",
			record, player, e2e.greaterHealFrame, noxServer.Frame(), e2e.greaterHealHP,
			player.HealthData.Cur, e2e.greaterHealMana, update.ManaCur,
			math.Float32frombits(uint32(record.Field72)))
		noxServer.Spells.Dur.CancelSpell(record)
	})
}

func (sc *e2eScenario) ArmForceOfNature(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		player := noxServer.Players.HostUnit()
		return player != nil && player.Class().Has(object.ClassPlayer) &&
			!player.Flags().HasAny(object.FlagDead|object.FlagDestroyed)
	}, func() {
		player := noxServer.Players.HostUnit()
		arg := &server.SpellAcceptArg{Obj: player, Pos: player.PosVec}
		if !noxServer.spells.duration.New(spell.SPELL_FORCE_OF_NATURE, player, player, player, arg, 1,
			legacy.Get_sub_52EF30(), legacy.Get_sub_52EFD0(), legacy.Get_sub_52F1D0(), 12) {
			e2eError(fmt.Errorf("FORCE OF NATURE duration creation failed for player %p", player))
			return
		}
		record := noxServer.Spells.Dur.List
		charge := noxServer.spells.duration.forceOfNatureCharges[record]
		if record == nil || record.Spell != uint32(spell.SPELL_FORCE_OF_NATURE) ||
			record.Caster16 != player || charge == nil || record.Update != legacy.Get_sub_52EFD0() {
			e2eError(fmt.Errorf("FORCE OF NATURE creation state: record=%p caster=%p charge=%p", record, player, charge))
			return
		}
		record.Field76 = uintptr(0x3fdccccc)
		e2e.forceOfNaturePlayer = player
		e2e.forceOfNatureRecord = record
		e2e.forceOfNatureCharge = charge
		e2e.forceOfNatureFrame = noxServer.Frame()
		e2e.forceOfNatureLaunches = noxServer.spells.duration.forceOfNatureLaunches
		e2eLog.Printf("FORCE OF NATURE ARMED: record=%p caster=%p charge=%p frame=%d field76=%#x",
			record, player, charge, e2e.forceOfNatureFrame, record.Field76)
	})
}

func (sc *e2eScenario) AssertForceOfNatureChargeRemoved(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return e2e.forceOfNatureRecord != nil && noxServer.Frame() >= e2e.forceOfNatureFrame+6
	}, func() {
		record := e2e.forceOfNatureRecord
		found := false
		for cur := noxServer.Spells.Dur.List; cur != nil; cur = cur.Next {
			if cur == record {
				found = true
				break
			}
		}
		if !found || noxServer.spells.duration.forceOfNatureCharges[record] != nil ||
			record.Field76 != uintptr(0x3fdccccc) || record.Caster16 != e2e.forceOfNaturePlayer {
			e2eError(fmt.Errorf("FORCE OF NATURE charge state: found=%t charge=%p field76=%#x caster=%p",
				found, noxServer.spells.duration.forceOfNatureCharges[record], record.Field76, record.Caster16))
			return
		}
		e2eLog.Printf("FORCE OF NATURE CHARGE REMOVED: record=%p charge=%p frame=%d field76=%#x",
			record, e2e.forceOfNatureCharge, noxServer.Frame(), record.Field76)
	})
}

func (sc *e2eScenario) AssertForceOfNatureLaunched(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return e2e.forceOfNatureRecord != nil && noxServer.Frame() >= e2e.forceOfNatureFrame+12
	}, func() {
		count := noxServer.spells.duration.forceOfNatureLaunches
		if count != e2e.forceOfNatureLaunches+1 {
			e2eError(fmt.Errorf("FORCE OF NATURE launch count = %d, want %d at frame %d",
				count, e2e.forceOfNatureLaunches+1, noxServer.Frame()))
			return
		}
		e2eLog.Printf("FORCE OF NATURE LAUNCHED: caster=%p frame=%d count=%d",
			e2e.forceOfNaturePlayer, noxServer.Frame(), count)
	})
}

func (sc *e2eScenario) AssertForceOfNatureCompleted(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return e2e.forceOfNatureRecord != nil && noxServer.Frame() >= e2e.forceOfNatureFrame+14
	}, func() {
		for cur := noxServer.Spells.Dur.List; cur != nil; cur = cur.Next {
			if cur == e2e.forceOfNatureRecord {
				e2eError(fmt.Errorf("FORCE OF NATURE duration still linked at frame %d", noxServer.Frame()))
				return
			}
		}
		if noxServer.spells.duration.forceOfNatureCharges[e2e.forceOfNatureRecord] != nil {
			e2eError(fmt.Errorf("FORCE OF NATURE charge retained after destroy: %p", e2e.forceOfNatureRecord))
			return
		}
		e2eLog.Printf("FORCE OF NATURE COMPLETED: record=%p caster=%p frame=%d",
			e2e.forceOfNatureRecord, e2e.forceOfNaturePlayer, noxServer.Frame())
	})
}

func (sc *e2eScenario) ArmManaBomb(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		player := noxServer.Players.HostUnit()
		return player != nil && player.Class().Has(object.ClassPlayer) &&
			player.UpdateDataPlayer() != nil && player.UpdateDataPlayer().ManaMax >= 20 &&
			!player.Flags().HasAny(object.FlagDead|object.FlagDestroyed)
	}, func() {
		player := noxServer.Players.HostUnit()
		update := player.UpdateDataPlayer()
		update.ManaCur = update.ManaMax
		update.ManaPrev = update.ManaCur
		mass := math.Float32bits(player.Mass)
		arg := &server.SpellAcceptArg{Obj: player, Pos: player.PosVec}
		if !noxServer.spells.duration.New(spell.SPELL_MANA_BOMB, player, player, player, arg, 1,
			legacy.Get_nox_xxx_manaBomb_530F90(), legacy.Get_nox_xxx_manaBombBoom_5310C0(), legacy.Get_sub_531290(), 24) {
			e2eError(fmt.Errorf("MANA BOMB duration creation failed for player %p", player))
			return
		}
		record := noxServer.Spells.Dur.List
		charge := noxServer.spells.duration.manaBombCharges[record]
		if record == nil || record.Spell != uint32(spell.SPELL_MANA_BOMB) ||
			record.Caster16 != player || charge == nil || record.Update != legacy.Get_nox_xxx_manaBombBoom_5310C0() {
			e2eError(fmt.Errorf("MANA BOMB creation state: record=%p caster=%p charge=%p", record, player, charge))
			return
		}
		record.Field76 = uintptr(0x3fdccccc)
		e2e.manaBombPlayer = player
		e2e.manaBombRecord = record
		e2e.manaBombCharge = charge
		e2e.manaBombFrame = noxServer.Frame()
		e2e.manaBombMass = mass
		e2e.manaBombPower = record.Field72
		e2eLog.Printf("MANA BOMB ARMED: record=%p caster=%p charge=%p frame=%d power=%d field76=%#x",
			record, player, charge, e2e.manaBombFrame, record.Field72, record.Field76)
	})
}

func (sc *e2eScenario) AssertManaBombUpdateAndCancel(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return e2e.manaBombRecord != nil && noxServer.Frame() >= e2e.manaBombFrame+3
	}, func() {
		record, player := e2e.manaBombRecord, e2e.manaBombPlayer
		found := false
		for cur := noxServer.Spells.Dur.List; cur != nil; cur = cur.Next {
			if cur == record {
				found = true
				break
			}
		}
		if !found || record.Caster16 != player || record.Field76 != uintptr(0x3fdccccc) ||
			noxServer.spells.duration.manaBombCharges[record] != e2e.manaBombCharge ||
			record.Field72 <= e2e.manaBombPower || math.Float32bits(player.Mass) != 1203982323 {
			e2eError(fmt.Errorf("MANA BOMB update state: found=%t record=%p caster=%p charge=%p power=%d->%d mass=%#x field76=%#x",
				found, record, record.Caster16, noxServer.spells.duration.manaBombCharges[record],
				e2e.manaBombPower, record.Field72, math.Float32bits(player.Mass), record.Field76))
			return
		}
		e2eLog.Printf("MANA BOMB UPDATED: record=%p frame=%d power=%d->%d charge=%p",
			record, noxServer.Frame(), e2e.manaBombPower, record.Field72, e2e.manaBombCharge)
		noxServer.Spells.Dur.CancelSpell(record)
	})
}

func (sc *e2eScenario) AssertManaBombCompleted(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return e2e.manaBombRecord != nil && noxServer.Frame() >= e2e.manaBombFrame+5
	}, func() {
		for cur := noxServer.Spells.Dur.List; cur != nil; cur = cur.Next {
			if cur == e2e.manaBombRecord {
				e2eError(fmt.Errorf("MANA BOMB duration still linked at frame %d", noxServer.Frame()))
				return
			}
		}
		if noxServer.spells.duration.manaBombCharges[e2e.manaBombRecord] != nil ||
			math.Float32bits(e2e.manaBombPlayer.Mass) != e2e.manaBombMass {
			e2eError(fmt.Errorf("MANA BOMB cleanup state: charge=%p mass=%#x want=%#x",
				noxServer.spells.duration.manaBombCharges[e2e.manaBombRecord],
				math.Float32bits(e2e.manaBombPlayer.Mass), e2e.manaBombMass))
			return
		}
		e2eLog.Printf("MANA BOMB COMPLETED: record=%p charge=%p frame=%d mass=%#x",
			e2e.manaBombRecord, e2e.manaBombCharge, noxServer.Frame(), e2e.manaBombMass)
	})
}

func (sc *e2eScenario) ArmChainLightning(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		player := noxServer.Players.HostUnit()
		return player != nil && player.Class().Has(object.ClassPlayer) &&
			player.UpdateDataPlayer() != nil && player.UpdateDataPlayer().ManaMax >= 10 &&
			e2e.monster != nil && !e2e.monster.Flags().HasAny(object.FlagDead|object.FlagDestroyed)
	}, func() {
		player, target := noxServer.Players.HostUnit(), e2e.monster
		update := player.UpdateDataPlayer()
		update.ManaCur = update.ManaMax
		update.ManaPrev = update.ManaCur
		update.CursorObj = target
		arg := &server.SpellAcceptArg{Obj: player, Pos: player.PosVec}
		if !noxServer.spells.duration.New(spell.SPELL_CHAIN_LIGHTNING, player, player, player, arg, 1,
			legacy.Get_nox_xxx_onStartLightning_52F820(), legacy.Get_nox_xxx_onFrameLightning_52F8A0(),
			legacy.Get_sub_530100(), 30) {
			e2eError(fmt.Errorf("CHAIN LIGHTNING duration creation failed for player %p", player))
			return
		}
		record := noxServer.Spells.Dur.List
		if record == nil || record.Spell != uint32(spell.SPELL_CHAIN_LIGHTNING) ||
			record.Caster16 != player || record.Update != legacy.Get_nox_xxx_onFrameLightning_52F8A0() {
			e2eError(fmt.Errorf("CHAIN LIGHTNING creation state: record=%p caster=%p target=%p", record, player, target))
			return
		}
		e2e.chainLightningPlayer = player
		e2e.chainLightningTarget = target
		e2e.chainLightningRecord = record
		e2e.chainLightningFrame = noxServer.Frame()
		e2e.chainLightningHealth = target.HealthData.Cur
		e2eLog.Printf("CHAIN LIGHTNING ARMED: record=%p caster=%p target=%p frame=%d", record, player, target, e2e.chainLightningFrame)
	})
}

func (sc *e2eScenario) AssertChainLightningUpdateAndCancel(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return e2e.chainLightningRecord != nil && noxServer.Frame() >= e2e.chainLightningFrame+3
	}, func() {
		record := e2e.chainLightningRecord
		found, rayFound := false, false
		for cur := noxServer.Spells.Dur.List; cur != nil; cur = cur.Next {
			if cur == record {
				found = true
				break
			}
		}
		for ray := record.Sub108; ray != nil; ray = ray.Next {
			if ray.Target48 == e2e.chainLightningTarget {
				rayFound = true
				break
			}
		}
		if !found || !rayFound || record.Caster16 != e2e.chainLightningPlayer {
			e2eError(fmt.Errorf("CHAIN LIGHTNING update: linked=%t ray=%t caster=%p target=%p frame=%d",
				found, rayFound, record.Caster16, e2e.chainLightningTarget, noxServer.Frame()))
			return
		}
		if health := e2e.chainLightningTarget.HealthData.Cur; health >= e2e.chainLightningHealth {
			e2eError(fmt.Errorf("CHAIN LIGHTNING target health did not decrease: before=%d after=%d",
				e2e.chainLightningHealth, health))
			return
		}
		e2eLog.Printf("CHAIN LIGHTNING UPDATED: record=%p ray=%p target=%p frame=%d",
			record, record.Sub108, e2e.chainLightningTarget, noxServer.Frame())
		noxServer.Spells.Dur.CancelSpell(record)
	})
}

func (sc *e2eScenario) AssertChainLightningCompleted(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return e2e.chainLightningRecord != nil && noxServer.Frame() >= e2e.chainLightningFrame+5
	}, func() {
		for cur := noxServer.Spells.Dur.List; cur != nil; cur = cur.Next {
			if cur == e2e.chainLightningRecord {
				e2eError(fmt.Errorf("CHAIN LIGHTNING duration still linked at frame %d", noxServer.Frame()))
				return
			}
		}
		if wand := noxServer.spells.duration.chainLightningWeapons[e2e.chainLightningRecord]; wand != nil {
			e2eError(fmt.Errorf("CHAIN LIGHTNING wand sidecar retained %p", wand))
			return
		}
		for _, ray := range noxClient.fxDurationRays {
			if ray.drawable != nil {
				e2eError(fmt.Errorf("CHAIN LIGHTNING client ray retained: drawable=%p source=%#x target=%#x",
					ray.drawable, ray.source, ray.target))
				return
			}
		}
		e2eLog.Printf("CHAIN LIGHTNING COMPLETED: record=%p frame=%d", e2e.chainLightningRecord, noxServer.Frame())
	})
}

func (sc *e2eScenario) ArmEnergyBolt(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		player, target := noxServer.Players.HostUnit(), e2e.monster
		return player != nil && target != nil && target.HealthData != nil &&
			!target.Flags().HasAny(object.FlagDead|object.FlagDestroyed)
	}, func() {
		player, target := noxServer.Players.HostUnit(), e2e.monster
		update := player.UpdateDataPlayer()
		update.ManaCur = update.ManaMax
		update.ManaPrev = update.ManaCur
		update.CursorObj = target
		direction := server.DirFromVec(target.PosVec.Sub(player.PosVec))
		player.Direction1, player.Direction2 = direction, direction
		rt := noxServer.spells.duration.energyBoltRuntime52E820()
		e2eLog.Printf("ENERGY BOLT FIXTURE: enemy=%t front=%t interact=%t cursor=%p range=%g player=%v target=%v",
			rt.IsEnemy(player, target), rt.InFront(player, target), rt.CanInteract(player, target),
			update.CursorObj, rt.Balance("LightningRange"), player.PosVec, target.PosVec)
		arg := &server.SpellAcceptArg{Obj: target, Pos: target.PosVec}
		if !noxServer.spells.duration.New(spell.SPELL_LIGHTNING, player, player, player, arg, 2,
			legacy.Get_nox_xxx_spellEnergyBoltStop_52E820(),
			legacy.Get_nox_xxx_spellEnergyBoltTick_52E850(), legacy.Get_nullsub_29(), 30) {
			e2eError(fmt.Errorf("ENERGY BOLT duration creation failed for player %p", player))
			return
		}
		record := noxServer.Spells.Dur.List
		if record == nil || record.Spell != uint32(spell.SPELL_LIGHTNING) ||
			record.Update != legacy.Get_nox_xxx_spellEnergyBoltTick_52E850() || record.Caster16 != player {
			e2eError(fmt.Errorf("ENERGY BOLT creation state: record=%p player=%p", record, player))
			return
		}
		e2e.energyBoltRecord = record
		e2e.energyBoltFrame = noxServer.Frame()
		e2e.energyBoltHealth = target.HealthData.Cur
		e2eLog.Printf("ENERGY BOLT ARMED: record=%p target=%p initial=%p frame=%d expiry=%d health=%d", record,
			target, record.Target48, e2e.energyBoltFrame, record.Frame68, e2e.energyBoltHealth)
	})
}

func (sc *e2eScenario) AssertEnergyBoltUpdateAndCancel(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return e2e.energyBoltRecord != nil && noxServer.Frame() >= e2e.energyBoltFrame+3
	}, func() {
		record, target := e2e.energyBoltRecord, e2e.monster
		linked := false
		for cur := noxServer.Spells.Dur.List; cur != nil; cur = cur.Next {
			if cur == record {
				linked = true
				break
			}
		}
		player := noxServer.Players.HostUnit()
		rt := noxServer.spells.duration.energyBoltRuntime52E820()
		e2eLog.Printf("ENERGY BOLT CHECK: spell=%d caster=%p flags=%#x expiry=%d cursor=%p enemy=%t front=%t interact=%t target_flags=%#x player=%v target=%v",
			record.Spell, record.Caster16, record.Flag20, record.Frame68,
			player.UpdateDataPlayer().CursorObj, rt.IsEnemy(player, target), rt.InFront(player, target),
			rt.CanInteract(player, target), uint32(target.ObjFlags), player.PosVec, target.PosVec)
		if !linked || record.Target48 != target || noxServer.spells.duration.durationRayTargets[record] != target ||
			target.HealthData.Cur >= e2e.energyBoltHealth {
			e2eError(fmt.Errorf("ENERGY BOLT update: linked=%t target=%p ray=%p health=%d before=%d",
				linked, record.Target48, noxServer.spells.duration.durationRayTargets[record],
				target.HealthData.Cur, e2e.energyBoltHealth))
			return
		}
		e2eLog.Printf("ENERGY BOLT UPDATED: record=%p target=%p frame=%d health=%d", record, target,
			noxServer.Frame(), target.HealthData.Cur)
		noxServer.Spells.Dur.CancelSpell(record)
	})
}

func (sc *e2eScenario) AssertEnergyBoltCompleted(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return e2e.energyBoltRecord != nil && noxServer.Frame() >= e2e.energyBoltFrame+5
	}, func() {
		for cur := noxServer.Spells.Dur.List; cur != nil; cur = cur.Next {
			if cur == e2e.energyBoltRecord {
				e2eError(fmt.Errorf("ENERGY BOLT duration still linked at frame %d", noxServer.Frame()))
				return
			}
		}
		if ray := noxServer.spells.duration.durationRayTargets[e2e.energyBoltRecord]; ray != nil {
			e2eError(fmt.Errorf("ENERGY BOLT ray sidecar retained %p", ray))
			return
		}
		e2eLog.Printf("ENERGY BOLT COMPLETED: record=%p frame=%d", e2e.energyBoltRecord, noxServer.Frame())
	})
}

func (sc *e2eScenario) ArmDrainMana(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		player, target := noxServer.Players.HostUnit(), e2e.monster
		return player != nil && target != nil && !target.Flags().HasAny(object.FlagDead|object.FlagDestroyed)
	}, func() {
		player, target := noxServer.Players.HostUnit(), e2e.monster
		update := player.UpdateDataPlayer()
		update.ManaCur = update.ManaMax - 10
		update.ManaPrev = update.ManaCur
		target.UpdateDataMonster().StatusFlags |= 0x20
		arg := &server.SpellAcceptArg{Obj: target, Pos: target.PosVec}
		if !noxServer.spells.duration.New(spell.SPELL_DRAIN_MANA, player, player, player, arg, 2,
			nil, legacy.Get_nox_xxx_spellDrainMana_52E210(), nil, 30) {
			e2eError(fmt.Errorf("DRAIN MANA duration creation failed for player %p", player))
			return
		}
		record := noxServer.Spells.Dur.List
		if record == nil || record.Spell != uint32(spell.SPELL_DRAIN_MANA) ||
			record.Update != legacy.Get_nox_xxx_spellDrainMana_52E210() || record.Caster16 != player {
			e2eError(fmt.Errorf("DRAIN MANA creation state: record=%p player=%p", record, player))
			return
		}
		e2e.drainManaRecord = record
		e2e.drainManaFrame = noxServer.Frame()
		e2e.drainManaBefore = update.ManaCur
		e2eLog.Printf("DRAIN MANA ARMED: record=%p target=%p frame=%d mana=%d", record, target,
			e2e.drainManaFrame, e2e.drainManaBefore)
	})
}

func (sc *e2eScenario) AssertDrainManaUpdateAndCancel(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return e2e.drainManaRecord != nil && noxServer.Frame() >= e2e.drainManaFrame+3
	}, func() {
		record, target := e2e.drainManaRecord, e2e.monster
		linked := false
		for cur := noxServer.Spells.Dur.List; cur != nil; cur = cur.Next {
			if cur == record {
				linked = true
				break
			}
		}
		mana := noxServer.Players.HostUnit().UpdateDataPlayer().ManaCur
		if !linked || record.Target48 != target || noxServer.spells.duration.durationRayTargets[record] != target ||
			mana <= e2e.drainManaBefore {
			e2eError(fmt.Errorf("DRAIN MANA update: linked=%t target=%p ray=%p mana=%d before=%d",
				linked, record.Target48, noxServer.spells.duration.durationRayTargets[record],
				mana, e2e.drainManaBefore))
			return
		}
		e2eLog.Printf("DRAIN MANA UPDATED: record=%p target=%p frame=%d mana=%d", record, target,
			noxServer.Frame(), mana)
		noxServer.Spells.Dur.CancelSpell(record)
	})
}

func (sc *e2eScenario) AssertDrainManaCompleted(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return e2e.drainManaRecord != nil && noxServer.Frame() >= e2e.drainManaFrame+5
	}, func() {
		for cur := noxServer.Spells.Dur.List; cur != nil; cur = cur.Next {
			if cur == e2e.drainManaRecord {
				e2eError(fmt.Errorf("DRAIN MANA duration still linked at frame %d", noxServer.Frame()))
				return
			}
		}
		if ray := noxServer.spells.duration.durationRayTargets[e2e.drainManaRecord]; ray != nil {
			e2eError(fmt.Errorf("DRAIN MANA ray sidecar retained %p", ray))
			return
		}
		e2eLog.Printf("DRAIN MANA COMPLETED: record=%p frame=%d", e2e.drainManaRecord, noxServer.Frame())
	})
}

func (sc *e2eScenario) ArmDurationRayDraws(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		for _, ray := range noxClient.fxDurationRays {
			if ray.kind == 3 && ray.drawable != nil {
				return true
			}
		}
		return false
	}, func() {
		for _, ray := range noxClient.fxDurationRays {
			if ray.kind == 3 && ray.drawable != nil {
				e2e.durationRayDrawSource, e2e.durationRayDrawTarget = ray.source, ray.target
				break
			}
		}
		for index, kind := range [...]byte{1, 2, 4, 5, 6, 7} {
			packet := []byte{0x9e, kind, 0,
				byte(e2e.durationRayDrawSource), byte(e2e.durationRayDrawSource >> 8),
				byte(e2e.durationRayDrawTarget), byte(e2e.durationRayDrawTarget >> 8)}
			if got := noxClient.handleDurationRayPacketNative48EA70(packet); got != len(packet) {
				e2eError(fmt.Errorf("DURATION RAY kind %d consumed %d bytes", kind, got))
				return
			}
			for _, ray := range noxClient.fxDurationRays {
				if ray.kind == kind && ray.source == e2e.durationRayDrawSource && ray.target == e2e.durationRayDrawTarget {
					e2e.durationRayDrawables[index] = ray.drawable
					break
				}
			}
			if e2e.durationRayDrawables[index] == nil {
				e2eError(fmt.Errorf("DURATION RAY kind %d did not spawn", kind))
				return
			}
			e2eLog.Printf("DURATION RAY kind=%d drawable=%p update=%p draw=%p", kind,
				e2e.durationRayDrawables[index], e2e.durationRayDrawables[index].ClientUpdateFuncPtr,
				e2e.durationRayDrawables[index].DrawFuncPtr)
		}
		e2e.durationRayDrawFrame = noxServer.Frame()
		e2eLog.Printf("DURATION RAYS ARMED: source=%#x target=%#x frame=%d", e2e.durationRayDrawSource,
			e2e.durationRayDrawTarget, e2e.durationRayDrawFrame)
	})
}

func (sc *e2eScenario) AssertDurationRayDrawsAndCleanup(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return e2e.durationRayDrawFrame != 0 && noxServer.Frame() >= e2e.durationRayDrawFrame+3
	}, func() {
		for index, kind := range [...]byte{1, 2, 4, 5, 6, 7} {
			found := false
			for _, ray := range noxClient.fxDurationRays {
				if ray.kind == kind && ray.drawable == e2e.durationRayDrawables[index] {
					found = true
					break
				}
			}
			if !found {
				e2eError(fmt.Errorf("DURATION RAY kind %d disappeared before draw check", kind))
				return
			}
			stop := []byte{0x9e, kind + 7, 0,
				byte(e2e.durationRayDrawTarget), byte(e2e.durationRayDrawTarget >> 8),
				byte(e2e.durationRayDrawSource), byte(e2e.durationRayDrawSource >> 8)}
			if got := noxClient.handleDurationRayPacketNative48EA70(stop); got != len(stop) {
				e2eError(fmt.Errorf("DURATION RAY stop kind %d consumed %d bytes", kind, got))
				return
			}
		}
		for _, ray := range noxClient.fxDurationRays {
			if ray.kind != 3 && ray.drawable != nil {
				e2eError(fmt.Errorf("DURATION RAY cleanup retained kind %d drawable=%p", ray.kind, ray.drawable))
				return
			}
		}
		e2eLog.Printf("DURATION RAYS DRAWN AND REMOVED: frame=%d", noxServer.Frame())
	})
}

func (sc *e2eScenario) ArmTurnUndead(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		player := noxServer.Players.HostUnit()
		return player != nil && player.Class().Has(object.ClassPlayer) &&
			!player.Flags().HasAny(object.FlagDead|object.FlagDestroyed)
	}, func() {
		player := noxServer.Players.HostUnit()
		arg := &server.SpellAcceptArg{Obj: player, Pos: player.PosVec}
		if !noxServer.spells.duration.New(spell.SPELL_TURN_UNDEAD, player, player, player, arg, 1,
			legacy.Get_nox_xxx_spellTurnUndeadCreate_531310(),
			legacy.Get_nox_xxx_spellTurnUndeadUpdate_531410(),
			legacy.Get_nox_xxx_spellTurnUndeadDelete_531420(), 70) {
			e2eError(fmt.Errorf("TURN UNDEAD duration creation failed for player %p", player))
			return
		}
		record := noxServer.Spells.Dur.List
		if record == nil {
			e2eError(fmt.Errorf("TURN UNDEAD duration not linked for player %p", player))
			return
		}
		typeInd := uint16(noxServer.Types.IndByID("UndeadKiller"))
		var pending int
		for obj := noxServer.Objs.Pending; obj != nil; obj = obj.ObjNext {
			if obj.TypeInd == typeInd && obj.CollideData != nil &&
				(*server.UndeadKillerCollideData)(obj.CollideData).Spell == record {
				pending++
			}
		}
		if record.Spell != uint32(spell.SPELL_TURN_UNDEAD) ||
			record.Caster16 != player || record.Field72 <= 0 || pending != 43 ||
			record.Update != legacy.Get_nox_xxx_spellTurnUndeadUpdate_531410() {
			e2eError(fmt.Errorf("TURN UNDEAD creation state: record=%p player=%p budget=%d pending=%d",
				record, player, record.Field72, pending))
			return
		}
		record.Pos.X = math.Float32frombits(0x3fdccccc)
		e2e.turnUndeadRecord = record
		e2e.turnUndeadFrame = noxServer.Frame()
		e2eLog.Printf("TURN UNDEAD ARMED: record=%p player=%p budget=%d pending=%d frame=%d",
			record, player, record.Field72, pending, e2e.turnUndeadFrame)
	})
}

func (sc *e2eScenario) ArmBlink(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		player := noxServer.Players.HostUnit()
		return player != nil && player.Class().Has(object.ClassPlayer) &&
			!player.Flags().HasAny(object.FlagDead|object.FlagDestroyed)
	}, func() {
		player := noxServer.Players.HostUnit()
		arg := &server.SpellAcceptArg{Obj: player, Pos: player.PosVec}
		if !noxServer.spells.duration.New(spell.SPELL_BLINK, player, player, player, arg, 1,
			legacy.Get_nox_xxx_spellBlink2_530310(), legacy.Get_nox_xxx_spellBlink1_530380(), nil, 60) {
			e2eError(fmt.Errorf("BLINK duration creation failed for player %p", player))
			return
		}
		record := noxServer.Spells.Dur.List
		if record == nil || record.Spell != uint32(spell.SPELL_BLINK) ||
			record.Caster16 != player || record.Target48 != player ||
			record.Create != legacy.Get_nox_xxx_spellBlink2_530310() ||
			record.Update != legacy.Get_nox_xxx_spellBlink1_530380() {
			e2eError(fmt.Errorf("BLINK creation state: record=%p player=%p", record, player))
			return
		}
		// E2E actions run between game updates. Defer the callback by one
		// full tick so Frame68-1 matches the next server update frame.
		record.Frame68 = noxServer.Frame() + 2
		e2e.blinkPlayer = player
		e2e.blinkRecord = record
		e2e.blinkFrame = noxServer.Frame()
		e2e.blinkOrigin = player.PosVec
		e2eLog.Printf("BLINK ARMED: record=%p target=%p frame=%d trigger=%d origin=%v",
			record, player, e2e.blinkFrame, record.Frame68-1, player.PosVec)
	})
}

func (sc *e2eScenario) AssertBlinkCompleted(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return e2e.blinkRecord != nil && noxServer.Frame() >= e2e.blinkFrame+3
	}, func() {
		for cur := noxServer.Spells.Dur.List; cur != nil; cur = cur.Next {
			if cur == e2e.blinkRecord {
				e2eError(fmt.Errorf("BLINK duration still linked at frame %d", noxServer.Frame()))
				return
			}
		}
		wakeType := uint16(noxServer.Types.IndByID("TeleportWake"))
		var wake *server.Object
		for obj := noxServer.Objs.First(); obj != nil; obj = obj.ObjNext {
			if obj.TypeInd == wakeType && obj.ObjOwner == e2e.blinkPlayer && obj.CollideData != nil &&
				!obj.Flags().Has(object.FlagDestroyed) {
				wake = obj
				break
			}
		}
		if wake == nil {
			e2eError(fmt.Errorf("BLINK teleport wake missing at frame %d", noxServer.Frame()))
			return
		}
		destination := (*server.TeleportWakeCollideData)(wake.CollideData).Destination
		if destination == e2e.blinkOrigin || wake.PosVec != e2e.blinkOrigin ||
			e2e.blinkPlayer.PosVec != destination ||
			wake.Field34 <= noxServer.Frame() {
			e2eError(fmt.Errorf("BLINK result: wake=%p origin=%v/%v destination=%v player=%v expires=%d frame=%d",
				wake, wake.PosVec, e2e.blinkOrigin, destination, e2e.blinkPlayer.PosVec,
				wake.Field34, noxServer.Frame()))
			return
		}
		e2eLog.Printf("BLINK COMPLETED: record=%p player=%p wake=%p origin=%v destination=%v frame=%d",
			e2e.blinkRecord, e2e.blinkPlayer, wake, e2e.blinkOrigin, destination, noxServer.Frame())
	})
}

func (sc *e2eScenario) ArmSwap(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		player, target := noxServer.Players.HostUnit(), e2e.monster
		return player != nil && target != nil &&
			!player.Flags().HasAny(object.FlagDead|object.FlagDestroyed) &&
			!target.Flags().HasAny(object.FlagDead|object.FlagDestroyed)
	}, func() {
		caster, target := noxServer.Players.HostUnit(), e2e.monster
		arg := &server.SpellAcceptArg{Obj: target, Pos: target.PosVec}
		if !noxServer.spells.duration.New(spell.SPELL_SWAP, caster, caster, caster, arg, 1,
			legacy.Get_sub_530CA0(), legacy.Get_sub_530D30(), nil, 60) {
			e2eError(fmt.Errorf("SWAP duration creation failed for caster %p and target %p", caster, target))
			return
		}
		record := noxServer.Spells.Dur.List
		if record == nil || record.Spell != uint32(spell.SPELL_SWAP) ||
			record.Caster16 != caster || record.Target48 != target ||
			record.Create != legacy.Get_sub_530CA0() || record.Update != legacy.Get_sub_530D30() {
			e2eError(fmt.Errorf("SWAP creation state: record=%p caster=%p target=%p", record, caster, target))
			return
		}
		// E2E actions run between updates. The glyph flag skips range/LOS so
		// this checks the native teleport path independent of map geometry.
		// The old PE32 update treated the first word of Pos as Target48 and
		// dereferenced this sentinel at +16 (the reported 0x3fdcccdc fault).
		record.Pos.X = math.Float32frombits(0x3fdccccc)
		record.Flag20 = 1
		record.Frame68 = noxServer.Frame() + 2
		e2e.swapCaster, e2e.swapTarget = caster, target
		e2e.swapRecord, e2e.swapFrame = record, noxServer.Frame()
		e2e.swapCasterOrigin, e2e.swapTargetOrigin = caster.PosVec, target.PosVec
		e2eLog.Printf("SWAP ARMED: record=%p caster=%p target=%p frame=%d trigger=%d positions=%v/%v",
			record, caster, target, e2e.swapFrame, record.Frame68-1,
			e2e.swapCasterOrigin, e2e.swapTargetOrigin)
	})
}

func (sc *e2eScenario) AssertSwapCompleted(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return e2e.swapRecord != nil && noxServer.Frame() >= e2e.swapFrame+3
	}, func() {
		for cur := noxServer.Spells.Dur.List; cur != nil; cur = cur.Next {
			if cur == e2e.swapRecord {
				e2eError(fmt.Errorf("SWAP duration still linked at frame %d", noxServer.Frame()))
				return
			}
		}
		caster, target := e2e.swapCaster, e2e.swapTarget
		casterDelta := caster.PosVec.Sub(e2e.swapTargetOrigin)
		targetDelta := target.PosVec.Sub(e2e.swapCasterOrigin)
		casterDistance := math.Hypot(float64(casterDelta.X), float64(casterDelta.Y))
		targetDistance := math.Hypot(float64(targetDelta.X), float64(targetDelta.Y))
		// Spider AI can move between arming and the scheduled tick. The
		// original positions are 48 units apart; require both to be near
		// the other's origin and to have reversed their X ordering.
		if casterDistance >= 40 || targetDistance >= 40 ||
			e2e.swapCasterOrigin.X >= e2e.swapTargetOrigin.X || caster.PosVec.X <= target.PosVec.X {
			e2eError(fmt.Errorf("SWAP positions: caster=%v target=%v, origins=%v/%v, distances=%.2f/%.2f",
				caster.PosVec, target.PosVec, e2e.swapCasterOrigin, e2e.swapTargetOrigin,
				casterDistance, targetDistance))
			return
		}
		e2eLog.Printf("SWAP COMPLETED: record=%p frame=%d caster=%v target=%v distances=%.2f/%.2f",
			e2e.swapRecord, noxServer.Frame(), caster.PosVec, target.PosVec,
			casterDistance, targetDistance)
	})
}

func (sc *e2eScenario) ArmTeleportToTarget(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		player := noxServer.Players.HostUnit()
		return player != nil && !player.Flags().HasAny(object.FlagDead|object.FlagDestroyed)
	}, func() {
		player := noxServer.Players.HostUnit()
		origin := player.PosVec
		destination := types.Ptf(origin.X+12, origin.Y)
		arg := &server.SpellAcceptArg{Obj: player, Pos: destination}
		if !noxServer.spells.duration.New(spell.SPELL_TELEPORT_TO_TARGET, player, player, player, arg, 1,
			legacy.Get_sub_530A30_spell_execdur(), legacy.Get_nox_xxx_castTTT_530B70(), nil, 60) {
			e2eError(fmt.Errorf("TELEPORT TO TARGET duration creation failed: player=%p origin=%v destination=%v", player, origin, destination))
			return
		}
		record := noxServer.Spells.Dur.List
		if record == nil || record.Spell != uint32(spell.SPELL_TELEPORT_TO_TARGET) ||
			record.Caster16 != player || record.Target48 != player || record.Pos2 != destination ||
			record.Create != legacy.Get_sub_530A30_spell_execdur() ||
			record.Update != legacy.Get_nox_xxx_castTTT_530B70() {
			e2eError(fmt.Errorf("TELEPORT TO TARGET creation state: record=%p player=%p", record, player))
			return
		}
		// The old PE32 update reads this float as Target48 and faults at +16.
		record.Pos.X = math.Float32frombits(0x3fdccccc)
		record.Frame68 = noxServer.Frame() + 2
		e2e.teleportTargetPlayer, e2e.teleportTargetRecord = player, record
		e2e.teleportTargetOrigin, e2e.teleportTargetPos = origin, destination
		e2e.teleportTargetFrame = noxServer.Frame()
		e2eLog.Printf("TELEPORT TO TARGET ARMED: record=%p player=%p origin=%v destination=%v frame=%d",
			record, player, origin, destination, e2e.teleportTargetFrame)
	})
}

func (sc *e2eScenario) AssertTeleportToTargetCompleted(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return e2e.teleportTargetRecord != nil && noxServer.Frame() >= e2e.teleportTargetFrame+3
	}, func() {
		for record := noxServer.Spells.Dur.List; record != nil; record = record.Next {
			if record == e2e.teleportTargetRecord {
				e2eError(fmt.Errorf("TELEPORT TO TARGET duration still linked at frame %d", noxServer.Frame()))
				return
			}
		}
		wakeType := uint16(noxServer.Types.IndByID("TeleportWake"))
		var wake *server.Object
		for obj := noxServer.Objs.First(); obj != nil; obj = obj.ObjNext {
			if obj.TypeInd == wakeType && obj.ObjOwner == e2e.teleportTargetPlayer && obj.CollideData != nil &&
				!obj.Flags().Has(object.FlagDestroyed) {
				wake = obj
				break
			}
		}
		if wake == nil || wake.PosVec != e2e.teleportTargetOrigin ||
			(*server.TeleportWakeCollideData)(wake.CollideData).Destination != e2e.teleportTargetPos ||
			e2e.teleportTargetPlayer.PosVec != e2e.teleportTargetPos || wake.Field34 <= noxServer.Frame() {
			e2eError(fmt.Errorf("TELEPORT TO TARGET result: wake=%p player=%v origin=%v destination=%v frame=%d",
				wake, e2e.teleportTargetPlayer.PosVec, e2e.teleportTargetOrigin, e2e.teleportTargetPos, noxServer.Frame()))
			return
		}
		e2eLog.Printf("TELEPORT TO TARGET COMPLETED: record=%p player=%p wake=%p destination=%v frame=%d",
			e2e.teleportTargetRecord, e2e.teleportTargetPlayer, wake, e2e.teleportTargetPos, noxServer.Frame())
	})
}

func (sc *e2eScenario) ArmTeleportPop(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		player := noxServer.Players.HostUnit()
		return player != nil && !player.Flags().HasAny(object.FlagDead|object.FlagDestroyed)
	}, func() {
		player := noxServer.Players.HostUnit()
		marker := noxServer.NewObjectByTypeID("TeleportGlyph1")
		if marker == nil {
			e2eError(fmt.Errorf("TELEPORT POP cannot create TeleportGlyph1 fixture"))
			return
		}
		markerPos := player.PosVec.Add(types.Ptf(48, 0))
		noxServer.CreateObjectAt(marker, player, markerPos)
		noxServer.ObjectsAddPending()
		data := player.UpdateDataPlayer()
		if data.Field29[0] != nil {
			e2eError(fmt.Errorf("TELEPORT POP marker slot already occupied: %p", data.Field29[0]))
			return
		}
		data.Field29[0] = marker
		data.Field39 = data.Field39&^uint32(0xff) | 1
		origin := player.PosVec
		arg := &server.SpellAcceptArg{Obj: player, Pos: marker.PosVec}
		if !noxServer.spells.duration.New(spell.SPELL_TELEPORT_POP, player, player, player, arg, 1,
			legacy.Get_nox_xxx_castTele_530820(), legacy.Get_sub_530880(), nil, 60) {
			e2eError(fmt.Errorf("TELEPORT POP duration creation failed: player=%p marker=%p", player, marker))
			return
		}
		record := noxServer.Spells.Dur.List
		if record == nil || record.Spell != uint32(spell.SPELL_TELEPORT_POP) ||
			record.Caster16 != player || record.Target48 != player ||
			record.Create != legacy.Get_nox_xxx_castTele_530820() || record.Update != legacy.Get_sub_530880() {
			e2eError(fmt.Errorf("TELEPORT POP creation state: record=%p player=%p", record, player))
			return
		}
		// The old PE32 update reads this float as Target48 and faults at +16.
		record.Pos.X = math.Float32frombits(0x3fdccccc)
		record.Frame68 = noxServer.Frame() + 2
		e2e.teleportPopPlayer, e2e.teleportPopRecord, e2e.teleportPopMarker = player, record, marker
		e2e.teleportPopOrigin, e2e.teleportPopMarkerPos, e2e.teleportPopFrame = origin, marker.PosVec, noxServer.Frame()
		e2eLog.Printf("TELEPORT POP ARMED: record=%p player=%p marker=%p origin=%v marker_pos=%v frame=%d",
			record, player, marker, origin, marker.PosVec, e2e.teleportPopFrame)
	})
}

func (sc *e2eScenario) AssertTeleportPopCompleted(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return e2e.teleportPopRecord != nil && noxServer.Frame() >= e2e.teleportPopFrame+3
	}, func() {
		for record := noxServer.Spells.Dur.List; record != nil; record = record.Next {
			if record == e2e.teleportPopRecord {
				e2eError(fmt.Errorf("TELEPORT POP duration still linked at frame %d", noxServer.Frame()))
				return
			}
		}
		player := e2e.teleportPopPlayer
		data := player.UpdateDataPlayer()
		if data.Field29[0] != nil || byte(data.Field39) != 0 {
			e2eError(fmt.Errorf("TELEPORT POP marker not consumed: marker=%p charge=%d", data.Field29[0], byte(data.Field39)))
			return
		}
		wakeType := uint16(noxServer.Types.IndByID("TeleportWake"))
		var wake *server.Object
		for obj := noxServer.Objs.First(); obj != nil; obj = obj.ObjNext {
			if obj.TypeInd == wakeType && obj.ObjOwner == player && obj.CollideData != nil &&
				obj.PosVec == e2e.teleportPopOrigin && !obj.Flags().Has(object.FlagDestroyed) {
				wake = obj
				break
			}
		}
		markerDelta := player.PosVec.Sub(e2e.teleportPopMarkerPos)
		markerDistance := math.Hypot(float64(markerDelta.X), float64(markerDelta.Y))
		if wake == nil || (*server.TeleportWakeCollideData)(wake.CollideData).Destination != e2e.teleportPopMarkerPos ||
			player.PosVec.X <= e2e.teleportPopOrigin.X+15 || markerDistance >= 40 || wake.Field34 <= noxServer.Frame() {
			e2eError(fmt.Errorf("TELEPORT POP result: wake=%p player_pos=%v origin=%v marker_pos=%v distance=%.2f frame=%d",
				wake, player.PosVec, e2e.teleportPopOrigin, e2e.teleportPopMarkerPos, markerDistance, noxServer.Frame()))
			return
		}
		e2eLog.Printf("TELEPORT POP COMPLETED: record=%p player=%p marker=%p wake=%p destination=%v frame=%d",
			e2e.teleportPopRecord, player, e2e.teleportPopMarker, wake, player.PosVec, noxServer.Frame())
	})
}

func (sc *e2eScenario) ArmTeleportToMark(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		player := noxServer.Players.HostUnit()
		return player != nil && !player.Flags().HasAny(object.FlagDead|object.FlagDestroyed)
	}, func() {
		player := noxServer.Players.HostUnit()
		marker := noxServer.NewObjectByTypeID("TeleportGlyph2")
		if marker == nil {
			e2eError(fmt.Errorf("TELEPORT TO MARK cannot create TeleportGlyph2 fixture"))
			return
		}
		markerPos := player.PosVec.Add(types.Ptf(48, 0))
		noxServer.CreateObjectAt(marker, player, markerPos)
		noxServer.ObjectsAddPending()
		data := player.UpdateDataPlayer()
		if data.Field29[1] != nil {
			e2eError(fmt.Errorf("TELEPORT TO MARK slot already occupied: %p", data.Field29[1]))
			return
		}
		data.Field29[1] = marker
		data.Field39 = data.Field39&^(uint32(0xff)<<8) | 1<<8
		origin := player.PosVec
		arg := &server.SpellAcceptArg{Obj: player, Pos: marker.PosVec}
		if !noxServer.spells.duration.New(spell.SPELL_TELEPORT_TO_MARK_2, player, player, player, arg, 1,
			legacy.Get_sub_5305D0(), legacy.Get_sub_530650(), nil, 60) {
			e2eError(fmt.Errorf("TELEPORT TO MARK duration creation failed: player=%p marker=%p", player, marker))
			return
		}
		record := noxServer.Spells.Dur.List
		if record == nil || record.Spell != uint32(spell.SPELL_TELEPORT_TO_MARK_2) ||
			record.Obj12 != player || record.Caster16 != player || record.Target48 != player ||
			record.Create != legacy.Get_sub_5305D0() || record.Update != legacy.Get_sub_530650() {
			e2eError(fmt.Errorf("TELEPORT TO MARK creation state: record=%p player=%p", record, player))
			return
		}
		// The PE32 update would interpret this float as a native target pointer.
		record.Pos.X = math.Float32frombits(0x3fdccccc)
		record.Frame68 = noxServer.Frame() + 2
		e2e.teleportMarkPlayer, e2e.teleportMarkRecord = player, record
		e2e.teleportMarkOrigin, e2e.teleportMarkPos, e2e.teleportMarkFrame = origin, marker.PosVec, noxServer.Frame()
		e2eLog.Printf("TELEPORT TO MARK ARMED: record=%p player=%p origin=%v marker_pos=%v frame=%d",
			record, player, origin, marker.PosVec, e2e.teleportMarkFrame)
	})
}

func (sc *e2eScenario) AssertTeleportToMarkCompleted(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return e2e.teleportMarkRecord != nil && noxServer.Frame() >= e2e.teleportMarkFrame+3
	}, func() {
		for record := noxServer.Spells.Dur.List; record != nil; record = record.Next {
			if record == e2e.teleportMarkRecord {
				e2eError(fmt.Errorf("TELEPORT TO MARK duration still linked at frame %d", noxServer.Frame()))
				return
			}
		}
		player := e2e.teleportMarkPlayer
		data := player.UpdateDataPlayer()
		if data.Field29[1] != nil || byte(data.Field39>>8) != 0 {
			e2eError(fmt.Errorf("TELEPORT TO MARK glyph not consumed: marker=%p charge=%d", data.Field29[1], byte(data.Field39>>8)))
			return
		}
		wakeType := uint16(noxServer.Types.IndByID("TeleportWake"))
		var wake *server.Object
		var wakePositions []types.Pointf
		for obj := noxServer.Objs.First(); obj != nil; obj = obj.ObjNext {
			if obj.TypeInd == wakeType && obj.ObjOwner == player && obj.CollideData != nil &&
				!obj.Flags().Has(object.FlagDestroyed) {
				wakePositions = append(wakePositions, obj.PosVec)
				delta := obj.PosVec.Sub(e2e.teleportMarkOrigin)
				if math.Hypot(float64(delta.X), float64(delta.Y)) < 8 &&
					(*server.TeleportWakeCollideData)(obj.CollideData).Destination == e2e.teleportMarkPos {
					wake = obj
					break
				}
			}
		}
		markerDelta := player.PosVec.Sub(e2e.teleportMarkPos)
		markerDistance := math.Hypot(float64(markerDelta.X), float64(markerDelta.Y))
		if wake == nil || (*server.TeleportWakeCollideData)(wake.CollideData).Destination != e2e.teleportMarkPos ||
			player.PosVec.X <= e2e.teleportMarkOrigin.X+15 || markerDistance >= 40 || wake.Field34 <= noxServer.Frame() {
			e2eError(fmt.Errorf("TELEPORT TO MARK result: wake=%p wake_positions=%v player_pos=%v origin=%v marker_pos=%v distance=%.2f frame=%d",
				wake, wakePositions, player.PosVec, e2e.teleportMarkOrigin, e2e.teleportMarkPos, markerDistance, noxServer.Frame()))
			return
		}
		e2eLog.Printf("TELEPORT TO MARK COMPLETED: record=%p player=%p wake=%p destination=%v frame=%d",
			e2e.teleportMarkRecord, player, wake, player.PosVec, noxServer.Frame())
	})
}

func (sc *e2eScenario) AssertTurnUndeadActiveAndCancel(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return e2e.turnUndeadRecord != nil && noxServer.Frame() >= e2e.turnUndeadFrame+3
	}, func() {
		record := e2e.turnUndeadRecord
		found := false
		for cur := noxServer.Spells.Dur.List; cur != nil; cur = cur.Next {
			if cur == record {
				found = true
				break
			}
		}
		typeInd := uint16(noxServer.Types.IndByID("UndeadKiller"))
		var active int
		for obj := noxServer.Objs.First(); obj != nil; obj = obj.ObjNext {
			if obj.TypeInd == typeInd && obj.CollideData != nil &&
				(*server.UndeadKillerCollideData)(obj.CollideData).Spell == record &&
				!obj.Flags().Has(object.FlagDestroyed) {
				active++
			}
		}
		if !found || active == 0 || math.Float32bits(record.Pos.X) != 0x3fdccccc {
			e2eError(fmt.Errorf("TURN UNDEAD active state: found=%t active=%d pos-bits=%#x",
				found, active, math.Float32bits(record.Pos.X)))
			return
		}
		e2eLog.Printf("TURN UNDEAD ACTIVE: record=%p active=%d frame=%d", record, active, noxServer.Frame())
		noxServer.Spells.Dur.CancelSpell(record)
	})
}

func (sc *e2eScenario) AssertTurnUndeadCompleted(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return e2e.turnUndeadRecord != nil && noxServer.Frame() >= e2e.turnUndeadFrame+6
	}, func() {
		record := e2e.turnUndeadRecord
		for cur := noxServer.Spells.Dur.List; cur != nil; cur = cur.Next {
			if cur == record {
				e2eError(fmt.Errorf("TURN UNDEAD duration still linked at frame %d", noxServer.Frame()))
				return
			}
		}
		typeInd := uint16(noxServer.Types.IndByID("UndeadKiller"))
		for obj := noxServer.Objs.First(); obj != nil; obj = obj.ObjNext {
			if obj.TypeInd == typeInd && obj.CollideData != nil &&
				(*server.UndeadKillerCollideData)(obj.CollideData).Spell == record &&
				!obj.Flags().Has(object.FlagDestroyed) {
				e2eError(fmt.Errorf("TURN UNDEAD projectile still active: %p at frame %d", obj, noxServer.Frame()))
				return
			}
		}
		e2eLog.Printf("TURN UNDEAD COMPLETED: record=%p frame=%d", record, noxServer.Frame())
	})
}

func (sc *e2eScenario) ArmMoonglow(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		player := noxServer.Players.HostUnit()
		return player != nil && player.HealthData != nil && player.HealthData.Cur > 0 &&
			!player.Flags().HasAny(object.FlagDead|object.FlagDestroyed)
	}, func() {
		player := noxServer.Players.HostUnit()
		if player.HasEnchant(server.ENCHANT_MOONGLOW) {
			e2eError(fmt.Errorf("MOONGLOW fixture already has the buff"))
			return
		}
		arg := &server.SpellAcceptArg{Obj: player, Pos: player.PosVec}
		if !noxServer.spells.duration.New(spell.SPELL_MOONGLOW, player, player, player, arg, 1,
			legacy.Get_nox_xxx_spellCreateMoonglow_531A00(), nil, legacy.Get_sub_531AF0(), 600) {
			e2eError(fmt.Errorf("MOONGLOW duration creation failed for player %p", player))
			return
		}
		record := noxServer.Spells.Dur.List
		visual := noxServer.spells.duration.moonglowVisuals[record]
		if record == nil || record.Spell != uint32(spell.SPELL_MOONGLOW) ||
			record.Target48 != player || record.Create != legacy.Get_nox_xxx_spellCreateMoonglow_531A00() ||
			record.Destroy != legacy.Get_sub_531AF0() || visual == nil ||
			!player.HasEnchant(server.ENCHANT_MOONGLOW) {
			e2eError(fmt.Errorf("MOONGLOW creation state: record=%p player=%p visual=%p buff=%t",
				record, player, visual, player.HasEnchant(server.ENCHANT_MOONGLOW)))
			return
		}
		record.Pos.X = math.Float32frombits(0x3fdccccc)
		e2e.moonglowPlayer = player
		e2e.moonglowRecord = record
		e2e.moonglowVisual = visual
		e2e.moonglowFrameBefore = noxServer.Frame()
		e2eLog.Printf("MOONGLOW ARMED: record=%p target=%p visual=%p frame=%d",
			record, player, visual, e2e.moonglowFrameBefore)
	})
}

func (sc *e2eScenario) AssertMoonglowAndCancel(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return e2e.moonglowRecord != nil && noxServer.Frame() >= e2e.moonglowFrameBefore+3
	}, func() {
		record, player := e2e.moonglowRecord, e2e.moonglowPlayer
		found := false
		for cur := noxServer.Spells.Dur.List; cur != nil; cur = cur.Next {
			if cur == record {
				found = true
				break
			}
		}
		if !found || noxServer.spells.duration.moonglowVisuals[record] != e2e.moonglowVisual ||
			!player.HasEnchant(server.ENCHANT_MOONGLOW) || math.Float32bits(record.Pos.X) != 0x3fdccccc {
			e2eError(fmt.Errorf("MOONGLOW live state: found=%t visual=%p buff=%t pos-bits=%#x",
				found, noxServer.spells.duration.moonglowVisuals[record],
				player.HasEnchant(server.ENCHANT_MOONGLOW), math.Float32bits(record.Pos.X)))
			return
		}
		noxServer.Spells.Dur.CancelSpell(record)
		e2e.moonglowFrameBefore = noxServer.Frame()
		e2eLog.Printf("MOONGLOW CANCELLED: record=%p frame=%d", record, e2e.moonglowFrameBefore)
	})
}

func (sc *e2eScenario) AssertMoonglowDestroyed(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return e2e.moonglowRecord != nil && noxServer.Frame() >= e2e.moonglowFrameBefore+3
	}, func() {
		record := e2e.moonglowRecord
		found := false
		for cur := noxServer.Spells.Dur.List; cur != nil; cur = cur.Next {
			if cur == record {
				found = true
				break
			}
		}
		if found || noxServer.spells.duration.moonglowVisuals[record] != nil ||
			e2e.moonglowPlayer.HasEnchant(server.ENCHANT_MOONGLOW) {
			e2eError(fmt.Errorf("MOONGLOW destroy state: found=%t visual=%p buff=%t",
				found, noxServer.spells.duration.moonglowVisuals[record],
				e2e.moonglowPlayer.HasEnchant(server.ENCHANT_MOONGLOW)))
			return
		}
		e2eLog.Printf("MOONGLOW DESTROYED: record=%p player=%p frame=%d",
			record, e2e.moonglowPlayer, noxServer.Frame())
	})
}

func (sc *e2eScenario) PlaceGroundItemOnLava(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		item := e2e.groundItem
		return item != nil && item.HealthData != nil && item.HealthData.Cur != 0 &&
			item.Flags().Has(object.FlagActive) && !item.Flags().HasAny(object.FlagDead|object.FlagDestroyed)
	}, func() {
		item := e2e.groundItem
		pos, ok := e2eFindLavaTile()
		if !ok {
			e2eError(fmt.Errorf("map %q contains no tile index 6", legacy.Nox_xxx_mapGetMapName_409B40()))
			return
		}
		if item.Damage == nil {
			e2eError(fmt.Errorf("ground item %q has no damage callback", e2e.groundItemTypeID))
			return
		}
		e2e.lavaGroundItem = item
		e2e.lavaGroundOriginalPos = item.PosVec
		e2e.lavaGroundHealth = item.HealthData.Cur
		e2e.lavaGroundFrame = noxServer.Frame()
		queuedBefore := item.Field116
		asObjectS(item).SetPos(pos)
		item.NewPos = pos
		item.PrevPos = pos
		item.VelVec = types.Pointf{}
		item.ForceVec = types.Pointf{}
		item.Pos24 = types.Pointf{}
		e2eLog.Printf("GROUND ITEM LAVA CONTACT ARMED: item=%s object=%p class=%#x damage=%p health=%d/%d frame=%d pos=(%.3f,%.3f) queue=%#x pending=%d tile=%d",
			e2e.groundItemTypeID, item, uint32(item.ObjClass), item.Damage, item.HealthData.Cur,
			item.HealthData.Max, e2e.lavaGroundFrame, pos.X, pos.Y, queuedBefore,
			legacy.Get_dword_5d4594_2488604(), legacy.Nox_xxx_tileNFromPoint_411160(item.PosVec))
		// Weapons have neither Update nor Collide callbacks, so the movement
		// scheduler does not admit them to its pending-object queue. Exercise
		// the exact production hit allocator and 00548740 dispatcher directly;
		// this is the path used once any collidable health object reaches lava.
		legacy.Nox_xxx_allocHitArray_5486D0()
		legacy.Nox_xxx_collSysAddCollision_548630(item, 6, types.Pointf{})
		legacy.Nox_xxx_collide_548740()
	})
}

func (sc *e2eScenario) AssertGroundItemLavaDamage(name string) {
	sc.add(0, name, func() {
		item := e2e.lavaGroundItem
		if item == nil || item.HealthData == nil {
			e2eError(fmt.Errorf("LAVA ground-item fixture is unavailable: item=%p", item))
			return
		}
		after := item.HealthData.Cur
		frame := noxServer.Frame()
		posBeforeRestore := item.PosVec
		newPosBeforeRestore := item.NewPos
		flagsBeforeRestore := item.ObjFlags
		queuedBeforeRestore := item.Field116
		pendingBeforeRestore := legacy.Get_dword_5d4594_2488604()
		asObjectS(item).SetPos(e2e.lavaGroundOriginalPos)
		item.NewPos = e2e.lavaGroundOriginalPos
		item.PrevPos = e2e.lavaGroundOriginalPos
		item.VelVec = types.Pointf{}
		item.ForceVec = types.Pointf{}
		item.Pos24 = types.Pointf{}
		if after >= e2e.lavaGroundHealth {
			e2eError(fmt.Errorf("LAVA did not reduce ground item health: item=%s before=%d after=%d frames=%d->%d pos=%v new=%v flags=%#x queue=%#x pending=%d",
				e2e.groundItemTypeID, e2e.lavaGroundHealth, after, e2e.lavaGroundFrame, frame,
				posBeforeRestore, newPosBeforeRestore, uint32(flagsBeforeRestore), queuedBeforeRestore, pendingBeforeRestore))
			return
		}
		if after == 0 || item.Flags().HasAny(object.FlagDead|object.FlagDestroyed) {
			e2eError(fmt.Errorf("LAVA fixture destroyed ground item %q: health=%d flags=%#x", e2e.groundItemTypeID, after, uint32(item.Flags())))
			return
		}
		if item.Obj130 != nil || item.Field131 != uint32(object.DamageLava) || item.Pos132 != (types.Pointf{}) {
			e2eError(fmt.Errorf("LAVA ground-item metadata = source:%p type:%d hit-pos:%v", item.Obj130, item.Field131, item.Pos132))
			return
		}
		e2eLog.Printf("GROUND ITEM LAVA DAMAGE: item=%s object=%p damage_callback=%p health=%d->%d damage=%d frames=%d->%d type=%d restored=(%.3f,%.3f)",
			e2e.groundItemTypeID, item, item.Damage, e2e.lavaGroundHealth, after,
			e2e.lavaGroundHealth-after, e2e.lavaGroundFrame, frame, item.Field131,
			e2e.lavaGroundOriginalPos.X, e2e.lavaGroundOriginalPos.Y)
	})
}

func e2eNewSmokeBlastDrawables() (smokes, puffs []*client.Drawable) {
	if noxClient == nil || e2e.smokeBlastBaseline == nil {
		return nil, nil
	}
	smokeType := uint32(noxClient.Things.IndByID("Smoke"))
	puffType := uint32(noxClient.Things.IndByID("Puff"))
	for dr := noxClient.Objs.FirstList1(); dr != nil; dr = dr.Next() {
		if _, ok := e2e.smokeBlastBaseline[dr]; ok {
			continue
		}
		switch dr.TypeIDVal {
		case smokeType:
			smokes = append(smokes, dr)
		case puffType:
			puffs = append(puffs, dr)
		}
	}
	return smokes, puffs
}

func e2eAssertSmokeBlastDrawables() error {
	smokes, puffs := e2eNewSmokeBlastDrawables()
	if len(smokes) != 1 || len(puffs) != 6 {
		return fmt.Errorf("smoke-blast drawables = Smoke:%d Puff:%d, want 1/6", len(smokes), len(puffs))
	}
	for _, dr := range append(smokes, puffs...) {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= uintptr(^uint32(0)) {
			return fmt.Errorf("smoke-blast drawable used a low address: %p", dr)
		}
		if !dr.Flags().Has(object.FlagActive) {
			return fmt.Errorf("smoke-blast drawable is inactive: %p flags=%#x", dr, uint32(dr.Flags()))
		}
	}
	if dr := smokes[0]; dr.PosVec != e2e.smokeBlastPos || dr.ZVal != 20 {
		return fmt.Errorf("Smoke drawable = pos:%v Z:%d, want pos:%v Z:20", dr.PosVec, dr.ZVal, e2e.smokeBlastPos)
	}
	for _, dr := range puffs {
		delta := dr.PosVec.Sub(e2e.smokeBlastPos)
		if delta.X < -15 || delta.X > 15 || delta.Y < -15 || delta.Y > 15 || dr.ZVal < 5 || dr.ZVal > 25 {
			return fmt.Errorf("Puff drawable = pos:%v delta:%v Z:%d, want offsets -15..15 and Z 5..25", dr.PosVec, delta, dr.ZVal)
		}
	}
	e2eLog.Printf("SMOKE BLAST DECODED: Smoke=%p Puff=%d pos=%v Z=%d pointers=native", smokes[0], len(puffs), e2e.smokeBlastPos, smokes[0].ZVal)
	return nil
}

func (sc *e2eScenario) SmokeBlast(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return nox_client_isConnected() && noxServer.Players.HostUnit() != nil
	}, func() {
		smokeType := noxClient.Things.IndByID("Smoke")
		puffType := noxClient.Things.IndByID("Puff")
		if smokeType == 0 || puffType == 0 {
			e2eError(fmt.Errorf("smoke-blast client types are unavailable: Smoke=%d Puff=%d", smokeType, puffType))
			return
		}
		e2e.smokeBlastBaseline = make(map[*client.Drawable]struct{}, noxClient.Objs.Count)
		for dr := noxClient.Objs.FirstList1(); dr != nil; dr = dr.Next() {
			e2e.smokeBlastBaseline[dr] = struct{}{}
		}
		pos := noxServer.Players.HostUnit().Pos()
		e2e.smokeBlastPos = image.Pt(int(pos.X), int(pos.Y))
		if e2e.smokeBlastPos.X < math.MinInt16 || e2e.smokeBlastPos.X > math.MaxInt16 ||
			e2e.smokeBlastPos.Y < math.MinInt16 || e2e.smokeBlastPos.Y > math.MaxInt16 {
			e2eError(fmt.Errorf("smoke-blast position is outside packet range: %v", e2e.smokeBlastPos))
			return
		}
		var packet [5]byte
		packet[0] = byte(netmsg.MSG_FX_SMOKE_BLAST)
		binary.LittleEndian.PutUint16(packet[1:3], uint16(int16(e2e.smokeBlastPos.X)))
		binary.LittleEndian.PutUint16(packet[3:5], uint16(int16(e2e.smokeBlastPos.Y)))
		if got := noxClient.nox_xxx_netOnPacketRecvCli48EA70(server.HostPlayerIndex, packet[:]); got != 1 {
			e2eError(fmt.Errorf("smoke-blast production packet loop returned %d, want 1", got))
			return
		}
		if err := e2eAssertSmokeBlastDrawables(); err != nil {
			e2eError(err)
			return
		}
		e2eLog.Printf("SMOKE BLAST PACKET: opcode=%#x bytes=%x baseline=%d", packet[0], packet, len(e2e.smokeBlastBaseline))
		e2e.smokeBlastBaseline = nil
	})
}

func (sc *e2eScenario) DamagePoof(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return nox_client_isConnected() && noxServer.Players.HostUnit() != nil
	}, func() {
		typeInd := uint32(noxClient.Things.IndByID("DamagePoof"))
		if typeInd == 0 {
			e2eError(fmt.Errorf("damage-poof client type is unavailable"))
			return
		}
		baseline := make(map[*client.Drawable]struct{}, noxClient.Objs.Count)
		for dr := noxClient.Objs.FirstList1(); dr != nil; dr = dr.Next() {
			baseline[dr] = struct{}{}
		}
		pos := noxServer.Players.HostUnit().Pos()
		packetPos := image.Pt(int(pos.X), int(pos.Y))
		if packetPos.X < math.MinInt16 || packetPos.X > math.MaxInt16 ||
			packetPos.Y < math.MinInt16 || packetPos.Y > math.MaxInt16-2 {
			e2eError(fmt.Errorf("damage-poof position is outside packet range: %v", packetPos))
			return
		}
		var packet [5]byte
		packet[0] = byte(netmsg.MSG_FX_DAMAGE_POOF)
		binary.LittleEndian.PutUint16(packet[1:3], uint16(int16(packetPos.X)))
		binary.LittleEndian.PutUint16(packet[3:5], uint16(int16(packetPos.Y)))
		if got := noxClient.nox_xxx_netOnPacketRecvCli48EA70(server.HostPlayerIndex, packet[:]); got != 1 {
			e2eError(fmt.Errorf("damage-poof production packet loop returned %d, want 1", got))
			return
		}
		var created []*client.Drawable
		for dr := noxClient.Objs.FirstList1(); dr != nil; dr = dr.Next() {
			if _, ok := baseline[dr]; !ok && dr.TypeIDVal == typeInd {
				created = append(created, dr)
			}
		}
		if len(created) != 1 {
			e2eError(fmt.Errorf("new DamagePoof drawables = %d, want 1", len(created)))
			return
		}
		dr := created[0]
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= uintptr(^uint32(0)) {
			e2eError(fmt.Errorf("damage-poof drawable used a low address: %p", dr))
			return
		}
		wantPos := packetPos.Add(image.Pt(0, 2))
		if dr.PosVec != wantPos || !dr.Flags().Has(object.FlagActive) {
			e2eError(fmt.Errorf("DamagePoof drawable = %p pos:%v flags:%#x, want pos:%v active", dr, dr.PosVec, uint32(dr.Flags()), wantPos))
			return
		}
		e2eLog.Printf("DAMAGE POOF DECODED: drawable=%p pos=%v opcode=%#x pointers=native", dr, dr.PosVec, packet[0])
	})
}

func (sc *e2eScenario) ManaBombCancel(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return nox_client_isConnected() && noxServer.Players.HostUnit() != nil
	}, func() {
		typeInd := uint32(noxClient.Things.IndByID("CyanSpark"))
		if typeInd == 0 {
			e2eError(fmt.Errorf("mana-bomb-cancel client type CyanSpark is unavailable"))
			return
		}
		radius := manaBombCancelFloatToInt48EA70(float32(noxServer.Balance.Float("ManaBombOutRadius")))
		if radius <= 0 {
			e2eError(fmt.Errorf("mana-bomb-cancel radius = %d, want positive", radius))
			return
		}
		baseline := make(map[*client.Drawable]struct{}, noxClient.Objs.Count)
		for dr := noxClient.Objs.FirstList1(); dr != nil; dr = dr.Next() {
			baseline[dr] = struct{}{}
		}
		pos := noxServer.Players.HostUnit().Pos()
		packetPos := image.Pt(int(pos.X), int(pos.Y))
		if packetPos.X < math.MinInt16 || packetPos.X > math.MaxInt16 ||
			packetPos.Y < math.MinInt16 || packetPos.Y > math.MaxInt16 {
			e2eError(fmt.Errorf("mana-bomb-cancel position is outside packet range: %v", packetPos))
			return
		}
		var packet [5]byte
		packet[0] = byte(netmsg.MSG_FX_MANA_BOMB_CANCEL)
		binary.LittleEndian.PutUint16(packet[1:3], uint16(int16(packetPos.X)))
		binary.LittleEndian.PutUint16(packet[3:5], uint16(int16(packetPos.Y)))
		if got := noxClient.nox_xxx_netOnPacketRecvCli48EA70(server.HostPlayerIndex, packet[:]); got != 1 {
			e2eError(fmt.Errorf("mana-bomb-cancel production packet loop returned %d, want 1", got))
			return
		}
		var created []*client.Drawable
		for dr := noxClient.Objs.FirstList1(); dr != nil; dr = dr.Next() {
			if _, ok := baseline[dr]; !ok && dr.TypeIDVal == typeInd {
				created = append(created, dr)
			}
		}
		if len(created) != manaBombCancelSparkCount48EA70 {
			e2eError(fmt.Errorf("new mana-bomb-cancel CyanSpark drawables = %d, want %d", len(created), manaBombCancelSparkCount48EA70))
			return
		}
		for i, dr := range created {
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= uintptr(^uint32(0)) {
				e2eError(fmt.Errorf("mana-bomb-cancel spark %d used a low address: %p", i, dr))
				return
			}
			if !dr.Flags().Has(object.FlagActive) {
				e2eError(fmt.Errorf("mana-bomb-cancel spark %d is inactive: %p flags=%#x", i, dr, uint32(dr.Flags())))
				return
			}
			delta := dr.PosVec.Sub(packetPos)
			if delta.X < -int(radius) || delta.X > int(radius) || delta.Y < -int(radius) || delta.Y > int(radius) {
				e2eError(fmt.Errorf("mana-bomb-cancel spark %d position = %v delta:%v radius:%d", i, dr.PosVec, delta, radius))
				return
			}
			effect := dr.UnionEffect()
			if effect.Field_108 != uint32(dr.PosVec.X)<<12 || effect.Field_109 != uint32(dr.PosVec.Y)<<12 || effect.Field_110 != 0 {
				e2eError(fmt.Errorf("mana-bomb-cancel spark %d fixed state = (%#x, %#x, %#x)", i,
					effect.Field_108, effect.Field_109, effect.Field_110))
				return
			}
			duration := effect.Field_112 - effect.Field_111
			if duration < 30 || duration > 40 || dr.Field_74_4 != 0 || dr.ZVal != 0 || dr.VelZ < 4 || dr.VelZ > 10 {
				e2eError(fmt.Errorf("mana-bomb-cancel spark %d effect = duration:%d angle:%d Z:%d VelZ:%d", i,
					duration, dr.Field_74_4, dr.ZVal, dr.VelZ))
				return
			}
		}
		e2eLog.Printf("MANA BOMB CANCEL DECODED: CyanSpark=%d first=%p pos=%v radius=%d opcode=%#x pointers=native",
			len(created), created[0], packetPos, radius, packet[0])
	})
}

func e2eStockObjectDeath54E010(typeID, handler string) (*server.Object, *server.CreateSpawnObjectDeathData54E010, error) {
	typ := noxServer.Types.ByID(typeID)
	if typ == nil {
		return nil, nil, fmt.Errorf("stock object-death type %q is unavailable", typeID)
	}
	death, dataSize, ok := server.ObjectDeathHandler(handler)
	if !ok || death == nil {
		return nil, nil, fmt.Errorf("stock object-death handler %q is unavailable", handler)
	}
	wantSize := unsafe.Sizeof(server.CreateSpawnObjectDeathData54E010{})
	if dataSize != wantSize {
		return nil, nil, fmt.Errorf("stock object-death handler %q data size = %d, want %d", handler, dataSize, wantSize)
	}
	if typ.Death != death || typ.DeathData == nil {
		return nil, nil, fmt.Errorf("thing.bin type %q death contract = callback:%p data:%p, want %s/%p", typeID, typ.Death, typ.DeathData, handler, death)
	}
	obj := noxServer.NewObjectByTypeID(typeID)
	if obj == nil {
		return nil, nil, fmt.Errorf("cannot create stock object-death type %q", typeID)
	}
	if obj.Death != death || obj.DeathData != typ.DeathData {
		return nil, nil, fmt.Errorf("stock object %q copied death contract = callback:%p data:%p, want %p/%p", typeID, obj.Death, obj.DeathData, death, typ.DeathData)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(obj.CObj()) <= math.MaxUint32 {
		return nil, nil, fmt.Errorf("stock object %q used a low address: %p", typeID, obj.CObj())
	}
	return obj, (*server.CreateSpawnObjectDeathData54E010)(obj.DeathData), nil
}

func e2eObjectDeathTypeID54E010(data *server.CreateSpawnObjectDeathData54E010) string {
	end := bytes.IndexByte(data.TypeID[:], 0)
	if end < 0 {
		end = len(data.TypeID)
	}
	return string(data.TypeID[:end])
}

func (sc *e2eScenario) ObjectDeathSpawns(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return nox_client_isConnected() && noxServer.Players.HostUnit() != nil
	}, func() {
		player := noxServer.Players.HostUnit()
		chest, chestData, err := e2eStockObjectDeath54E010("Chest1", "SpawnObjectDie")
		if err != nil {
			e2eError(err)
			return
		}
		if got := e2eObjectDeathTypeID54E010(chestData); got != "NULL" || chestData.Sound == 0 {
			e2eError(fmt.Errorf("Chest1 parsed death data = type:%q sound:%d, want NULL/nonzero", got, chestData.Sound))
			return
		}
		chestPos := player.Pos().Add(types.Ptf(48, 0))
		noxServer.CreateObjectAt(chest, nil, chestPos)
		noxServer.ObjectsAddPending()
		if chest.Collide == nil || !chest.Flags().Has(object.FlagActive) || chest.Flags().HasAny(object.FlagDead|object.FlagDestroyed) {
			e2eError(fmt.Errorf("Chest1 fixture is not collision-ready: object=%p collide=%p flags=%#x", chest, chest.Collide, uint32(chest.Flags())))
			return
		}
		collision := [2]float32{3.5, -8.25}
		collisionPtr := unsafe.Pointer(&collision[0])
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(collisionPtr) <= math.MaxUint32 {
			e2eError(fmt.Errorf("Chest1 collision record used a low address: %p", collisionPtr))
			return
		}
		chest.CallCollide(int(uintptr(player.CObj())), int(uintptr(collisionPtr)))
		if collision != [2]float32{3.5, -8.25} {
			e2eError(fmt.Errorf("Chest1 collision record changed to %v", collision))
			return
		}
		if !chest.Flags().Has(object.FlagDead) || chest.Flags().Has(object.FlagDestroyed) || chest.DeathData != unsafe.Pointer(chestData) {
			e2eError(fmt.Errorf("SpawnObjectDie result = flags:%#x data:%p, want DEAD without DESTROYED and data %p", uint32(chest.Flags()), chest.DeathData, chestData))
			return
		}

		crate, crateData, err := e2eStockObjectDeath54E010("Crate1", "CreateObjectDie")
		if err != nil {
			e2eError(err)
			return
		}
		spawnedType := e2eObjectDeathTypeID54E010(crateData)
		if spawnedType != "CrateBreaking1" || crateData.Sound == 0 {
			e2eError(fmt.Errorf("Crate1 parsed death data = type:%q sound:%d, want CrateBreaking1/nonzero", spawnedType, crateData.Sound))
			return
		}
		cratePos := player.Pos().Add(types.Ptf(-48, 0))
		noxServer.CreateObjectAt(crate, nil, cratePos)
		noxServer.ObjectsAddPending()
		baseline := make(map[*server.Object]struct{})
		for _, obj := range noxServer.S().Objs.AllObjects() {
			baseline[obj] = struct{}{}
		}
		ccall.CallVoidPtr(crate.Death, crate.CObj())
		noxServer.ObjectsAddPending()
		if !crate.Flags().Has(object.FlagDestroyed) || crate.Flags().Has(object.FlagDead) || crate.DeathData != unsafe.Pointer(crateData) {
			e2eError(fmt.Errorf("CreateObjectDie result = flags:%#x data:%p, want DESTROYED without DEAD and data %p", uint32(crate.Flags()), crate.DeathData, crateData))
			return
		}
		var spawned []*server.Object
		for _, obj := range noxServer.S().Objs.AllObjects() {
			if _, ok := baseline[obj]; ok {
				continue
			}
			typ := obj.ObjectTypeC()
			if typ != nil && typ.ID() == spawnedType && obj.Pos() == cratePos {
				spawned = append(spawned, obj)
			}
		}
		if len(spawned) != 1 || !spawned[0].Flags().Has(object.FlagActive) {
			e2eError(fmt.Errorf("CreateObjectDie spawned %q objects = %d, want one active object at %v", spawnedType, len(spawned), cratePos))
			return
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(spawned[0].CObj()) <= math.MaxUint32 {
			e2eError(fmt.Errorf("CreateObjectDie spawned object used a low address: %p", spawned[0]))
			return
		}
		e2eLog.Printf("OBJECT DEATH SPAWNS: chest=%p chest_data=%p collision=%p flags=%#x crate=%p crate_data=%p spawned=%p/%s flags=%#x pointers=native",
			chest, chestData, collisionPtr, uint32(chest.Flags()), crate, crateData, spawned[0], spawnedType, uint32(spawned[0].Flags()))
	})
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
		e2e.monsterPlayerHP = 0
		if player.HealthData != nil {
			e2e.monsterPlayerHP = player.HealthData.Cur
		}
		e2e.monsterShield = nil
		e2e.monsterShieldHP = 0
		e2e.monsterShieldCarry = 0
		for item := player.InvFirstItem; item != nil; item = item.InvNextItem {
			if item.Flags().Has(object.FlagEquipped) && uint32(item.ObjSubClass)&2 != 0 && item.HealthData != nil {
				e2e.monsterShield = item
				e2e.monsterShieldHP = item.HealthData.Cur
				if item.UpdateData != nil {
					e2e.monsterShieldCarry = item.UpdateDataWeaponArmor().Field0
				}
				break
			}
		}
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

// SelectPlacedMonster binds the encounter fixture to a monster loaded from the
// campaign map and moves that object next to the host player. Unlike
// SpawnMonster, this preserves the map object's original init data, AI stack,
// script links, and per-map property overrides so headless tests exercise the
// same activation state as a normal campaign approach.
func (sc *e2eScenario) SelectPlacedMonster(typeID string, offset image.Point, name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return noxServer.Players.HostUnit() != nil
	}, func() {
		if e2e.monster != nil {
			e2eError(fmt.Errorf("monster fixture is already active: %p", e2e.monster))
			return
		}
		typ := noxServer.Types.ByID(typeID)
		if typ == nil {
			e2eError(fmt.Errorf("unknown placed monster type %q", typeID))
			return
		}
		if !typ.Class().Has(object.ClassMonster) {
			e2eError(fmt.Errorf("placed monster type %q has class %v", typeID, typ.Class()))
			return
		}

		player := noxServer.Players.HostUnit()
		var monster *server.Object
		bestDistance := math.MaxFloat64
		count := 0
		for candidate := noxServer.Objs.First(); candidate != nil; candidate = candidate.Next() {
			if int(candidate.TypeInd) != typ.Ind() || candidate.UpdateData == nil ||
				candidate.HealthData == nil || candidate.HealthData.Cur == 0 ||
				candidate.Flags().HasAny(object.FlagDead|object.FlagDestroyed) {
				continue
			}
			count++
			delta := candidate.PosVec.Sub(player.PosVec)
			distance := float64(delta.X*delta.X + delta.Y*delta.Y)
			if distance < bestDistance {
				bestDistance = distance
				monster = candidate
			}
		}
		if monster == nil {
			e2eError(fmt.Errorf("no live placed monster of type %q (objects=%d)", typeID, count))
			return
		}

		e2e.monster = monster
		e2e.monsterWorldTarget = nil
		e2e.monsterWorldTargetHP = 0
		originalMonsterPos := monster.PosVec
		pos := player.PosVec.Add(types.Ptf(float32(offset.X), float32(offset.Y)))
		asObjectS(monster).SetPos(pos)
		monster.NewPos = pos
		monster.PrevPos = pos
		monster.VelVec = types.Pointf{}
		monster.ForceVec = types.Pointf{}
		monster.Pos24 = types.Pointf{}

		update := monster.UpdateDataMonster()
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
		e2eLog.Printf("PLACED MONSTER: type=%s count=%d object=%p netcode=%d original_monster_pos=(%.3f,%.3f) player_pos=(%.3f,%.3f) monster_pos=(%.3f,%.3f) original_distance=%.3f can_see=%t",
			typeID, count, monster, monster.NetCode, originalMonsterPos.X, originalMonsterPos.Y,
			player.PosVec.X, player.PosVec.Y, monster.PosVec.X, monster.PosVec.Y, math.Sqrt(bestDistance), monster.CanSee(player))
		e2eLog.Printf("PLACED MONSTER STATE: frame=%d flags=%#x class=%#x subclass=%#x stack=%d[%s] field137=%d status=%#x def_status=%#x aggression=%g sight=%g flee=%g retreat=%g speed=%g health=%d/%d base=%d melee_range=%g missile_range=%g current=%p preferred=%p seen=%d buffs=%#x weapon=%#x armor=%#x",
			noxServer.Frame(), uint32(monster.ObjFlags), uint32(monster.ObjClass), uint32(monster.ObjSubClass),
			update.AIStackInd, strings.Join(actions, ","), update.Field137, uint32(update.StatusFlags), uint32(defStatus),
			update.Aggression, update.SightRange, update.FleeRange, update.RetreatLevel, monster.SpeedBase,
			healthCur, healthMax, healthBase, meleeRange, missileRange, update.CurrentEnemy, update.PreferredEnemy,
			update.Field282_1, monster.Buffs, update.WeaponEquipFlags, update.ArmorEquipFlags)
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

func (sc *e2eScenario) AssertMonsterShieldWear(name string) {
	sc.add(0, name, func() {
		shield := e2e.monsterShield
		if shield == nil || shield.HealthData == nil || shield.UpdateData == nil {
			e2eError(fmt.Errorf("monster shield-wear fixture is unavailable: shield=%p", shield))
			return
		}
		beforeHP := e2e.monsterShieldHP
		afterHP := shield.HealthData.Cur
		beforeCarry := math.Float32frombits(e2e.monsterShieldCarry)
		afterCarry := math.Float32frombits(shield.UpdateDataWeaponArmor().Field0)
		player := noxServer.Players.HostUnit()
		monster := e2e.monster
		var playerHP uint16
		var playerState server.PlayerState
		var playerDir server.Dir16
		var front int
		var playerPos, monsterPos, monsterPrevPos types.Pointf
		if player != nil {
			playerPos = player.PosVec
			playerDir = player.Direction1
			if player.HealthData != nil {
				playerHP = player.HealthData.Cur
			}
			if player.UpdateData != nil {
				playerState = player.UpdateDataPlayer().State
			}
			if monster != nil {
				monsterPos = monster.PosVec
				monsterPrevPos = monster.PrevPos
				front = legacy.Nox_server_testTwoPointsAndDirection_4E6E50(
					player.PosVec, int16(player.Direction1), monster.PrevPos,
				)
			}
		}
		if afterHP >= beforeHP && afterCarry <= beforeCarry && !shield.Flags().Has(object.FlagDestroyed) {
			e2eError(fmt.Errorf("monster attacks did not wear equipped shield: health=%d/%d carry=%g/%g player_health=%d state=%d direction=%d front=%d player_pos=%v monster_pos=%v monster_prev=%v",
				afterHP, beforeHP, afterCarry, beforeCarry, playerHP, playerState, playerDir, front,
				playerPos, monsterPos, monsterPrevPos))
			return
		}
		if int(playerHP)+10 < int(e2e.monsterPlayerHP) {
			e2eError(fmt.Errorf("monster wore shield but dealt unblocked player damage: health=%d->%d shield=%d->%d",
				e2e.monsterPlayerHP, playerHP, beforeHP, afterHP))
			return
		}
		e2eLog.Printf("MONSTER SHIELD WEAR: shield=%p health=%d->%d carry=%g->%g flags=%#x player_health=%d->%d state=%d direction=%d front=%d",
			shield, beforeHP, afterHP, beforeCarry, afterCarry, uint32(shield.Flags()),
			e2e.monsterPlayerHP, playerHP, playerState, playerDir, front)
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

func (sc *e2eScenario) KillZombieForRaise(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		player := noxServer.Players.HostUnit()
		monster := e2e.monster
		return player != nil && monster != nil && monster.UpdateData != nil && monster.HealthData != nil &&
			monster.HealthData.Cur != 0 && !monster.Flags().HasAny(object.FlagDead|object.FlagDestroyed)
	}, func() {
		player := noxServer.Players.HostUnit()
		monster := e2e.monster
		if !noxServer.S().IsZombie(monster) {
			e2eError(fmt.Errorf("zombie raise fixture has type %d", monster.TypeInd))
			return
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(monster.CObj()) <= math.MaxUint32 {
			e2eError(fmt.Errorf("zombie raise fixture used a low address: %p", monster))
			return
		}
		before := monster.HealthData.Cur
		update := monster.UpdateDataMonster()
		update.CurrentEnemy = player
		update.PreferredEnemy = player
		legacy.Nox_xxx_unitDamageClear_4EE5E0(monster, int(before))
		head := update.AIStackHead()
		if monster.HealthData.Cur != 0 || !monster.Flags().Has(object.FlagDead) ||
			head == nil || head.Type() != ai.ACTION_DYING {
			e2eError(fmt.Errorf("zombie death dispatch = health:%d flags:%#x stack:%#v",
				monster.HealthData.Cur, uint32(monster.Flags()), update.GetAIStack()))
			return
		}
		e2eLog.Printf("ZOMBIE DEATH DISPATCHED: object=%p frame=%d health=%d->%d flags=%#x action=%s pointers=native",
			monster, noxServer.Frame(), before, monster.HealthData.Cur, uint32(monster.Flags()), head.Type())
	})
}

func (sc *e2eScenario) ArmZombieRaise(name string) {
	sc.addWhen(0, name, 2400, func() bool {
		monster := e2e.monster
		if monster == nil || monster.UpdateData == nil || monster.HealthData == nil {
			return false
		}
		head := monster.UpdateDataMonster().AIStackHead()
		return head != nil && head.Type() == ai.ACTION_DEAD &&
			monster.Flags().Has(object.FlagShort) && !monster.Flags().Has(object.FlagAllowOverlap)
	}, func() {
		player := noxServer.Players.HostUnit()
		monster := e2e.monster
		update := monster.UpdateDataMonster()
		if player == nil || monster.HealthData.Cur != 0 || !monster.Flags().Has(object.FlagDead) ||
			!monster.Flags().Has(object.FlagShort) || monster.Flags().Has(object.FlagAllowOverlap) ||
			update.StatusFlags.HasAny(object.MonStatusOnFire|object.MonStatusStayDead) {
			e2eError(fmt.Errorf("zombie dead state = player:%p health:%d flags:%#x duration:%d status:%#x",
				player, monster.HealthData.Cur, uint32(monster.Flags()), update.Field123, uint32(update.StatusFlags)))
			return
		}
		originalDuration := update.Field123
		frame := noxServer.Frame()
		update.CurrentEnemy = player
		update.Field123 = 0
		update.Field137 = frame - 1
		e2eLog.Printf("ZOMBIE RAISE ARMED: object=%p frame=%d original_duration=%d dead_frame=%d enemy=%p flags=%#x",
			monster, frame, originalDuration, update.Field137, update.CurrentEnemy, uint32(monster.Flags()))
	})
}

func (sc *e2eScenario) WaitZombieRaised(name string) {
	blocked := object.FlagAllowOverlap | object.FlagShort | object.FlagNoCollide | object.FlagDead
	sc.addWhen(0, name, 1200, func() bool {
		monster := e2e.monster
		if monster == nil || monster.UpdateData == nil || monster.HealthData == nil {
			return false
		}
		head := monster.UpdateDataMonster().AIStackHead()
		return head != nil && head.Type() == ai.ACTION_GET_UP && monster.HealthData.Cur == monster.HealthData.Max &&
			!monster.Flags().HasAny(blocked)
	}, func() {
		monster := e2e.monster
		update := monster.UpdateDataMonster()
		if !noxServer.S().IsZombie(monster) || update.AIStackInd != 1 ||
			update.AIStack[0].Type() != ai.DEPENDENCY_UNINTERRUPTABLE ||
			update.AIStack[1].Type() != ai.ACTION_GET_UP {
			e2eError(fmt.Errorf("zombie raise stack = index:%d stack:%#v", update.AIStackInd, update.GetAIStack()))
			return
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(monster.CObj()) <= math.MaxUint32 {
			e2eError(fmt.Errorf("raised zombie used a low address: %p", monster))
			return
		}
		e2eLog.Printf("ZOMBIE RAISED: object=%p frame=%d health=%d/%d flags=%#x stack=%s,%s pointers=native",
			monster, noxServer.Frame(), monster.HealthData.Cur, monster.HealthData.Max, uint32(monster.Flags()),
			update.AIStack[0].Type(), update.AIStack[1].Type())
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

func (sc *e2eScenario) SpawnGroundItem(typeID, pickupHandler, expectedHandler string, ownedByPlayer bool, amount int, offset image.Point, name string) {
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
		if item == nil {
			e2eError(fmt.Errorf("cannot create ground item %q", typeID))
			return
		}
		if amount != 0 {
			if amount < 0 || uint64(amount) > uint64(^uint32(0)) {
				e2eError(fmt.Errorf("ground-item fixture amount must fit uint32, got %d", amount))
				return
			}
			switch typeID {
			case "Gold", "QuestGoldPile", "QuestGoldChest":
				data := item.InitDataGold()
				if data == nil {
					e2eError(fmt.Errorf("gold ground item %q has no init data", typeID))
					return
				}
				data.Amount = uint32(amount)
			default:
				e2eError(fmt.Errorf("ground-item fixture amount is unsupported for %q", typeID))
				return
			}
		}
		if pickupHandler != "" {
			handler, ok := server.ObjectPickupHandler(pickupHandler)
			if !ok || handler.Ptr == nil {
				e2eError(fmt.Errorf("unknown or nil pickup handler %q for ground item %q", pickupHandler, typeID))
				return
			}
			item.Pickup = handler
		}
		if item.Pickup.Ptr == nil {
			e2eError(fmt.Errorf("cannot create pickup-enabled ground item %q: item=%p", typeID, item))
			return
		}
		if expectedHandler != "" {
			handler, ok := server.ObjectPickupHandler(expectedHandler)
			if !ok || handler.Ptr == nil {
				e2eError(fmt.Errorf("unknown or nil expected pickup handler %q for ground item %q", expectedHandler, typeID))
				return
			}
			if item.Pickup.Ptr != handler.Ptr {
				e2eError(fmt.Errorf("ground item %q pickup callback = %p, want stock %s callback %p", typeID, item.Pickup.Ptr, expectedHandler, handler.Ptr))
				return
			}
		}
		pos := player.Pos().Add(types.Ptf(float32(offset.X), float32(offset.Y)))
		noxServer.CreateObjectAt(item, nil, pos)
		noxServer.ObjectsAddPending()
		if ownedByPlayer {
			noxServer.S().ObjSetOwner(player, item)
		}
		wireCode := noxServer.GetUnitNetCode(item)
		if wireCode <= 0 || wireCode > int(^uint16(0)) || item.InvHolder != nil ||
			!item.Flags().Has(object.FlagActive) || item.Flags().Has(object.FlagDestroyed) ||
			(ownedByPlayer && (item.ObjOwner != player || !item.HasOwner(player))) {
			e2eError(fmt.Errorf("ground item %q was not initialized: item=%p wire=%#x holder=%p owner=%p flags=%v", typeID, item, wireCode, item.InvHolder, item.ObjOwner, item.Flags()))
			return
		}
		e2e.groundItem = item
		e2e.groundItemTypeID = typeID
		e2e.groundItemPickupName = pickupHandler
		if expectedHandler != "" {
			e2e.groundItemPickupName = expectedHandler
		}
		e2e.groundItemPickupPtr = item.Pickup.Ptr
		e2e.groundItemOwned = ownedByPlayer
		e2e.groundItemBefore = before
		e2e.groundItemWireCode = uint16(wireCode)
		e2e.groundItemLivesBefore = player.UpdateDataPlayer().ExtraLives
		e2e.groundItemDropped = nil
		e2e.groundItemDropChecks = 0
		e2e.lavaGroundItem = nil
		e2e.lavaGroundHealth = 0
		e2eLog.Printf("GROUND ITEM SPAWNED: item=%s pickup=%s callback=%p object=%p owner=%p owned=%t amount=%d netcode=%d wire=%#x before=%d player_pos=(%.3f,%.3f) item_pos=(%.3f,%.3f)",
			typeID, e2e.groundItemPickupName, item.Pickup.Ptr, item, item.ObjOwner, ownedByPlayer, amount, item.NetCode, wireCode, before, player.PosVec.X, player.PosVec.Y, item.PosVec.X, item.PosVec.Y)
	})
}

func (sc *e2eScenario) AssertGroundItemConsumedExtraLife(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		player := noxServer.Players.HostUnit()
		if player == nil || e2e.groundItem == nil || e2e.groundItemWireCode == 0 {
			return false
		}
		update := player.UpdateDataPlayer()
		if update.ExtraLives != e2e.groundItemLivesBefore+1 {
			return false
		}
		count, err := e2eInventoryItemCount(e2e.groundItemTypeID)
		if err != nil || count != e2e.groundItemBefore {
			return false
		}
		return noxServer.S().ObjectFromNetCode4ECCB0(uint32(e2e.groundItemWireCode)) == nil &&
			noxClient.Objs.ByNetCode(e2e.groundItemWireCode) == nil
	}, func() {
		player := noxServer.Players.HostUnit()
		if player == nil {
			e2eError(fmt.Errorf("host player disappeared while asserting consumed extra life"))
			return
		}
		update := player.UpdateDataPlayer()
		wantLives := e2e.groundItemLivesBefore + 1
		count, err := e2eInventoryItemCount(e2e.groundItemTypeID)
		if err != nil {
			e2eError(err)
			return
		}
		serverObject := noxServer.S().ObjectFromNetCode4ECCB0(uint32(e2e.groundItemWireCode))
		clientObject := noxClient.Objs.ByNetCode(e2e.groundItemWireCode)
		if update.ExtraLives != wantLives || count != e2e.groundItemBefore || serverObject != nil || clientObject != nil {
			e2eError(fmt.Errorf("consumed extra-life item state = lives:%d inventory:%d server:%p client:%p, want lives:%d inventory:%d and no objects",
				update.ExtraLives, count, serverObject, clientObject, wantLives, e2e.groundItemBefore))
			return
		}
		e2eLog.Printf("GROUND ITEM CONSUMED: item=%s pickup=%s callback=%p object=%p wire=%#x extra_lives=%d->%d server_inventory=%d server_object=%p client_drawable=%p",
			e2e.groundItemTypeID, e2e.groundItemPickupName, e2e.groundItemPickupPtr, e2e.groundItem,
			e2e.groundItemWireCode, e2e.groundItemLivesBefore, update.ExtraLives, count, serverObject, clientObject)
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
		if item == nil || player == nil || item.Pickup.Ptr != e2e.groundItemPickupPtr || !player.HasItem(item) || item.InvHolder != player || item.ObjOwner != player || item.Flags().Has(object.FlagActive) {
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
		e2eLog.Printf("GROUND ITEM PICKED: item=%s pickup=%s callback=%p object=%p netcode=%d holder=%p owner=%p active=%t server_count=%d client_count=%d",
			e2e.groundItemTypeID, e2e.groundItemPickupName, item.Pickup.Ptr, item, item.NetCode, item.InvHolder, item.ObjOwner, item.Flags().Has(object.FlagActive), count, clientCount)
	})
}

func (sc *e2eScenario) DropGroundTrapByGameEx(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		player := noxServer.Players.HostUnit()
		return player != nil && e2e.groundItem != nil && player.HasItem(e2e.groundItem)
	}, func() {
		player := noxServer.Players.HostUnit()
		item := e2e.groundItem
		e2eLog.Printf("TRAP GAMEEX DROP: owner=%p item=%p class=%#x player_pos=%v", player, item, uint32(item.ObjClass), player.UpdateDataPlayer().Player.Pos3632())
		legacy.PlayerDropATrap(player)
	})
}

func (sc *e2eScenario) DragInventoryItemOut(typeID string, destination image.Point, name string) {
	if destination == (image.Point{}) {
		destination = image.Pt(100, 400)
	}
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
	sc.Input(4, "", &seat.MouseMoveEvent{Pos: destination, Relative: false})
	sc.addWhen(1, name+" dragging", 600, legacy.Nox_client_inventoryHasDragged, func() {
		e2eLog.Printf("INVENTORY DROP DRAGGING: item=%s destination=%v", typeID, destination)
	})
	sc.Input(1, "", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false})
}

func (sc *e2eScenario) AssertGroundItemDropped(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		item := e2e.groundItem
		player := noxServer.Players.HostUnit()
		e2e.groundItemDropChecks++
		if e2e.groundItemDropChecks == 1 || e2e.groundItemDropChecks%300 == 0 {
			var (
				count       = -1
				clientFound bool
				clientCount uint32
				wireCode    int
				drawable    bool
				serverType  uint16
				clientType  uint32
				serverClass object.Class
				clientClass object.Class
				pickupMatch bool
			)
			if item != nil {
				serverType = item.TypeInd
				serverClass = item.Class()
				pickupMatch = item.Pickup.Ptr == e2e.groundItemPickupPtr
				count, _ = e2eInventoryItemCount(e2e.groundItemTypeID)
				if typ := noxServer.Types.ByID(e2e.groundItemTypeID); typ != nil {
					clientFound, clientCount, _, _ = legacy.Nox_client_inventoryItemState(uint32(typ.Ind()))
				}
				wireCode = noxServer.GetUnitNetCode(item)
				if wireCode > 0 && wireCode <= int(^uint16(0)) {
					if dr := noxClient.Objs.ByNetCode(uint16(wireCode)); dr != nil {
						drawable = true
						clientType = dr.TypeIDVal
						clientClass = dr.Class()
					}
				}
			}
			e2eLog.Printf("GROUND ITEM DROP WAIT: item=%p player=%p in_inventory=%t holder=%p owner=%p active=%t destroyed=%t pickup_match=%t server_type=%d client_type=%d server_class=%v client_class=%v server_count=%d client_found=%t client_count=%d wire=%#x drawable=%t checks=%d",
				item, player, item != nil && player != nil && player.HasItem(item), func() *server.Object {
					if item == nil {
						return nil
					}
					return item.InvHolder
				}(), func() *server.Object {
					if item == nil {
						return nil
					}
					return item.ObjOwner
				}(), item != nil && item.Flags().Has(object.FlagActive), item != nil && item.Flags().Has(object.FlagDestroyed), pickupMatch, serverType, clientType, serverClass, clientClass, count, clientFound, clientCount, wireCode, drawable, e2e.groundItemDropChecks)
		}
		if item == nil || player == nil || player.HasItem(item) || item.InvHolder != nil ||
			item.Pickup.Ptr != e2e.groundItemPickupPtr || !item.Flags().Has(object.FlagActive) ||
			item.Flags().Has(object.FlagDestroyed) || (e2e.groundItemOwned && item.ObjOwner != player) {
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
		return drawable != nil && drawable.TypeIDVal == uint32(item.TypeInd)
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
		e2eLog.Printf("GROUND ITEM DROPPED: item=%s pickup=%s callback=%p object=%p drawable=%p server_type=%d client_type=%d server_class=%v client_class=%v netcode=%d wire=%#x holder=%p owner=%p active=%t server_count=%d distance=%.3f pos=(%.3f,%.3f)",
			e2e.groundItemTypeID, e2e.groundItemPickupName, item.Pickup.Ptr, item, drawable, item.TypeInd, drawable.TypeIDVal, item.Class(), drawable.Class(), item.NetCode, wireCode, item.InvHolder, item.ObjOwner, item.Flags().Has(object.FlagActive), count, distance, item.PosVec.X, item.PosVec.Y)
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

func (sc *e2eScenario) AssertServerInventoryItemCount(typeID string, want int, name string) {
	sc.add(0, name, func() {
		got, err := e2eInventoryItemCount(typeID)
		if err != nil {
			e2eError(err)
			return
		}
		if got != want {
			e2eError(fmt.Errorf("server inventory %q count = %d, want %d", typeID, got, want))
			return
		}
		e2eLog.Printf("SERVER INVENTORY COUNT: item=%s count=%d", typeID, got)
	})
}

func (sc *e2eScenario) AssertClientInventoryItemCount(typeID string, want int, name string) {
	sc.add(0, name, func() {
		typ := noxServer.Types.ByID(typeID)
		if typ == nil {
			e2eError(fmt.Errorf("unknown client inventory type %q", typeID))
			return
		}
		found, got, _, _ := legacy.Nox_client_inventoryItemState(uint32(typ.Ind()))
		if got != uint32(want) || found != (want != 0) {
			e2eError(fmt.Errorf("client inventory %q = found:%t count:%d, want found:%t count:%d", typeID, found, got, want != 0, want))
			return
		}
		e2eLog.Printf("CLIENT INVENTORY COUNT: item=%s count=%d", typeID, got)
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
		pos := player.Pos()
		pos.X += 40
		noxServer.CreateObjectAt(merchant, nil, pos)
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
		e2e.shopMerchant = merchant
		e2e.shopMerchantWireCode = uint16(wireCode)
		e2e.shopSession = nil
		e2eLog.Printf("SERVER SHOP FIXTURE: merchant=%p netcode=%d wire=%#x item=%s count=%d player_pos=%v merchant_pos=%v", merchant, merchant.NetCode, wireCode, typeID, count, player.Pos(), merchant.Pos())
	})
	sc.addWhen(0, name+" visible", 1200, func() bool {
		return e2e.shopMerchant != nil && e2e.shopMerchantWireCode != 0 &&
			noxClient.Objs.ByNetCode(e2e.shopMerchantWireCode) != nil
	}, func() {
		drawable := noxClient.Objs.ByNetCode(e2e.shopMerchantWireCode)
		pos := noxClient.Viewport().ToScreenPos(drawable.Pos())
		if !pos.In(noxClient.Viewport().Screen) {
			e2eError(fmt.Errorf("server shop fixture merchant is outside the viewport: world=%v screen=%v viewport=%v", drawable.Pos(), pos, noxClient.Viewport().Screen))
			return
		}
		e2eLog.Printf("SERVER SHOP MOUSE: merchant=%p drawable=%p wire=%#x world=%v screen=%v captured=%p focused=%p",
			e2e.shopMerchant, drawable, e2e.shopMerchantWireCode, drawable.Pos(), pos, noxClient.GUI.Captured(), noxClient.GUI.Focused())
		e2eQueueInput(&seat.MouseMoveEvent{Pos: pos, Relative: false})
	})
	sc.addWhen(1, name+" cursor", 600, func() bool {
		return noxClient.Nox_client_getCursorType() == gui.CursorShop
	}, func() {
		e2eLog.Printf("SERVER SHOP CURSOR: merchant=%p wire=%#x cursor=%d captured=%p focused=%p",
			e2e.shopMerchant, e2e.shopMerchantWireCode, noxClient.Nox_client_getCursorType(), noxClient.GUI.Captured(), noxClient.GUI.Focused())
		e2eQueueInput(&seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true})
	})
	sc.Input(1, "", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false})
}

func (sc *e2eScenario) AssertMapShopkeeperFieldGuide(id, creature, name string, count int) {
	sc.addWhen(0, name, 1200, func() bool {
		return noxServer.Objs.GetObjectByID(id) != nil
	}, func() {
		merchant := noxServer.Objs.GetObjectByID(id)
		if merchant == nil || !merchant.Class().Has(object.ClassMonster) ||
			!merchant.SubClass().AsMonster().Has(object.MonsterShopkeeper) || merchant.InitData == nil {
			e2eError(fmt.Errorf("map shopkeeper %q is invalid: object=%p", id, merchant))
			return
		}
		idata := merchant.InitDataShopkeeper()
		if count != 0 && int(idata.Count) != count {
			e2eError(fmt.Errorf("map shopkeeper %q definition count = %d, want %d", id, idata.Count, count))
			return
		}
		fieldGuide := noxServer.Types.ByID("FieldGuide")
		if fieldGuide == nil {
			e2eError(fmt.Errorf("FieldGuide object type is unavailable"))
			return
		}
		found := 0
		for i := 0; i < int(idata.Count) && i < len(idata.Items); i++ {
			def := &idata.Items[i]
			if def.TypeInd != uint32(fieldGuide.Ind()) {
				continue
			}
			param := noxServer.Types.ByInd(int(def.Param))
			if param == nil || param.ID() != creature {
				got := ""
				if param != nil {
					got = param.ID()
				}
				e2eError(fmt.Errorf("map shopkeeper %q FieldGuide parameter = %q (%d), want %q", id, got, def.Param, creature))
				return
			}
			if def.Count == 0 {
				e2eError(fmt.Errorf("map shopkeeper %q has an empty FieldGuide definition", id))
				return
			}
			found++
		}
		if found != 1 {
			e2eError(fmt.Errorf("map shopkeeper %q FieldGuide definition count = %d, want 1", id, found))
			return
		}
		e2eLog.Printf("MAP SHOPKEEPER FIELD GUIDE: id=%q merchant=%p definitions=%d creature=%q", id, merchant, idata.Count, creature)
	})
}

func (sc *e2eScenario) OpenMapShopkeeper(id, name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return noxServer.Players.HostUnit() != nil && noxServer.Objs.GetObjectByID(id) != nil
	}, func() {
		player := noxServer.Players.HostUnit()
		merchant := noxServer.Objs.GetObjectByID(id)
		if merchant == nil || !merchant.Class().Has(object.ClassMonster) ||
			!merchant.SubClass().AsMonster().Has(object.MonsterShopkeeper) || merchant.InitData == nil {
			e2eError(fmt.Errorf("map shopkeeper %q is invalid: object=%p", id, merchant))
			return
		}
		wireCode := noxServer.GetUnitNetCode(merchant)
		if wireCode <= 0 || wireCode > int(^uint16(0)) {
			e2eError(fmt.Errorf("map shopkeeper %q wire code = %#x", id, wireCode))
			return
		}
		pos := merchant.Pos()
		pos.X -= 48
		asObjectS(player).SetPos(pos)
		e2e.shopMerchant = merchant
		e2e.shopMerchantWireCode = uint16(wireCode)
		e2e.shopSession = nil
		e2eLog.Printf("MAP SHOPKEEPER: id=%q merchant=%p wire=%#x player_pos=%v merchant_pos=%v", id, merchant, wireCode, player.Pos(), merchant.Pos())
	})
	sc.addWhen(0, name+" visible", 1200, func() bool {
		if e2e.shopMerchant == nil || e2e.shopMerchantWireCode == 0 {
			return false
		}
		drawable := noxClient.Objs.ByNetCode(e2e.shopMerchantWireCode)
		return drawable != nil && noxClient.Viewport().ToScreenPos(drawable.Pos()).In(noxClient.Viewport().Screen)
	}, func() {
		drawable := noxClient.Objs.ByNetCode(e2e.shopMerchantWireCode)
		pos := noxClient.Viewport().ToScreenPos(drawable.Pos())
		e2eLog.Printf("MAP SHOPKEEPER TARGETING: merchant=%p drawable=%p wire=%#x world=%v screen=%v",
			e2e.shopMerchant, drawable, e2e.shopMerchantWireCode, drawable.Pos(), pos)
		e2eQueueInput(&seat.MouseMoveEvent{Pos: pos, Relative: false})
	})
	sc.add(2, name+" target", func() {
		target := legacy.Nox_xxx_clientGetSpriteAtCursor_476F90()
		if target == nil {
			e2eError(fmt.Errorf("map shopkeeper %q is not the client target at mouse=%v", id, noxClient.Inp.GetMousePos()))
			return
		}
		if target.NetCode32 != uint32(e2e.shopMerchantWireCode) {
			e2eError(fmt.Errorf("map shopkeeper %q client target code = %#x, want %#x", id, target.NetCode32, e2e.shopMerchantWireCode))
			return
		}
		serverPlayer := noxServer.Players.HostUnit()
		if serverPlayer == nil {
			e2eError(fmt.Errorf("map shopkeeper %q has no host player", id))
			return
		}
		serverUpdate := serverPlayer.UpdateDataPlayer()
		resolved := noxServer.S().ObjectFromNetCode4ECCB0(uint32(e2e.shopMerchantWireCode))
		if resolved != e2e.shopMerchant {
			e2eError(fmt.Errorf("map shopkeeper %q wire code resolved to %p, want %p", id, resolved, e2e.shopMerchant))
			return
		}
		if dialogState := legacy.Sub_47A260(); dialogState != 0 || nox_xxx_gameGet_4DB1B0() || serverUpdate.DialogWith != nil || serverUpdate.Trade70 != nil {
			e2eError(fmt.Errorf("map shopkeeper %q request gates are active: client dialog=%d blocked=%t server dialog=%p trade=%p",
				id, dialogState, nox_xxx_gameGet_4DB1B0(), serverUpdate.DialogWith, serverUpdate.Trade70))
			return
		}
		// This is the same client request issued by the action handler after a
		// shop cursor click. The synthetic server-shop scenario separately covers
		// cursor selection; this map regression must exercise Mystic's real wire
		// code, server object, shop definitions, and trade session.
		legacy.Nox_xxx_clientTrade_42E850(target)
		e2eLog.Printf("MAP SHOPKEEPER REQUEST: id=%q target=%p wire=%#x", id, target, target.NetCode32)
	})
}

func (sc *e2eScenario) AcquireFieldGuideFixture(creature, name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return noxServer.Players.HostUnit() != nil
	}, func() {
		guide := server.RewardFieldGuideID4F0D20(creature)
		if guide <= 0 || guide >= 41 {
			e2eError(fmt.Errorf("field-guide fixture creature %q has invalid guide ID %d", creature, guide))
			return
		}
		player := noxServer.Players.HostUnit()
		update := player.UpdateDataPlayer()
		if update == nil || update.Player == nil {
			e2eError(fmt.Errorf("field-guide fixture player data is unavailable: unit=%p update=%p", player, update))
			return
		}
		// Regular multiplayer grants every guide during player setup. Clear only
		// this fixture's guide so the real collision/use/award path can run.
		initialLevel := update.Player.BeastScrollLvl[guide]
		update.Player.BeastScrollLvl[guide] = 0
		item := noxServer.NewObjectByTypeID("FieldGuide")
		if item == nil {
			e2eError(fmt.Errorf("field-guide fixture cannot create FieldGuide"))
			return
		}
		if item.Use.Ptr != legacy.Get_sub_53F930() || item.UseData.Ptr == nil {
			e2eError(fmt.Errorf("field-guide fixture callbacks are unavailable: item=%p use=%p data=%p", item, item.Use.Ptr, item.UseData.Ptr))
			return
		}
		item.UseDataFieldGuide().SetCreature(creature)
		noxServer.CreateObjectAt(item, nil, player.Pos())
		noxServer.ObjectsAddPending()
		if !item.Flags().Has(object.FlagActive) || item.Flags().Has(object.FlagDestroyed) {
			e2eError(fmt.Errorf("field-guide fixture item was not activated: item=%p flags=%v", item, item.Flags()))
			return
		}

		e2e.fieldGuideID = guide
		e2e.fieldGuideCreature = creature
		// A listen server does not loop its host-targeted reliable report back
		// through the client queue. Capture that report before it enters the
		// reliable stream: manually injecting a duplicate while leaving the
		// original unacknowledged would block later sequenced packets, including
		// the shop dialog packets this scenario is intended to verify.
		var report []byte
		reportCount := 0
		func() {
			origSend := noxServer.Server.NetSendPacketXxx
			noxServer.Server.NetSendPacketXxx = func(recipient int, buf []byte, related *server.Object, removeIfDisconnected, sequenceEnabled int) int {
				if recipient == server.HostPlayerIndex && len(buf) == 3 && netmsg.Op(buf[0]) == netmsg.MSG_REPORT_GUIDE_AWARD {
					report = append(report[:0], buf...)
					reportCount++
					return 1
				}
				return origSend(recipient, buf, related, removeIfDisconnected, sequenceEnabled)
			}
			defer func() {
				noxServer.Server.NetSendPacketXxx = origSend
			}()
			noxServer.SignCollide4EAB40(item, player, nil)
		}()
		if level := update.Player.BeastScrollLvl[guide]; level != 1 {
			e2eError(fmt.Errorf("field-guide fixture %q server level = %d, want 1", creature, level))
			return
		}
		if reportCount != 1 || len(report) != 3 || int(report[1]) != guide || report[2] != 1 {
			e2eError(fmt.Errorf("field-guide fixture report = %v (count %d), want [%d %d 1]", report, reportCount, netmsg.MSG_REPORT_GUIDE_AWARD, guide))
			return
		}
		// Feed the captured server packet to the normal C decoder so this still
		// exercises MSG_REPORT_GUIDE_AWARD and the 45D140 reward handler.
		if got := legacy.Nox_xxx_netOnPacketRecvCli_48EA70_switch(server.HostPlayerIndex, netmsg.MSG_REPORT_GUIDE_AWARD, report); got != len(report) {
			e2eError(fmt.Errorf("field-guide reward packet consumed %d bytes, want %d", got, len(report)))
			return
		}
		e2eLog.Printf("FIELD GUIDE ACQUIRED: creature=%s guide=%d item=%p player=%p initial_level=%d", creature, guide, item, player, initialLevel)
	})
}

func (sc *e2eScenario) AssertFieldGuideReward(creature, name string) {
	sc.add(0, name, func() {
		guide := server.RewardFieldGuideID4F0D20(creature)
		if guide != e2e.fieldGuideID || creature != e2e.fieldGuideCreature {
			e2eError(fmt.Errorf("field-guide reward target = %q/%d, acquired %q/%d", creature, guide, e2e.fieldGuideCreature, e2e.fieldGuideID))
			return
		}
		level, guideMode, page, found := legacy.Nox_client_guideRewardState45D140(guide)
		if level != 1 || !guideMode || !found {
			e2eError(fmt.Errorf("field-guide client reward = level:%d guide-mode:%t page:%d found:%t, want level 1 active sorted page", level, guideMode, page, found))
			return
		}
		e2eLog.Printf("FIELD GUIDE REWARD: creature=%s guide=%d level=%d page=%d guide_mode=%t", creature, guide, level, page, guideMode)
	})
}

func (sc *e2eScenario) AssertFieldGuideRewardOpen(creature, name string) {
	sc.add(0, name, func() {
		guide := server.RewardFieldGuideID4F0D20(creature)
		if guide != e2e.fieldGuideID || creature != e2e.fieldGuideCreature {
			e2eError(fmt.Errorf("field-guide reward target = %q/%d, acquired %q/%d", creature, guide, e2e.fieldGuideCreature, e2e.fieldGuideID))
			return
		}
		level, guideMode, page, found := legacy.Nox_client_guideRewardState45D140(guide)
		if level != 1 || !guideMode {
			e2eError(fmt.Errorf("field-guide client reward = level:%d guide-mode:%t page:%d found:%t, want level 1 with guide open", level, guideMode, page, found))
			return
		}
		e2eLog.Printf("FIELD GUIDE REWARD OPEN: creature=%s guide=%d level=%d page=%d guide_mode=%t found=%t", creature, guide, level, page, guideMode, found)
	})
}

func (sc *e2eScenario) CloseFieldGuideReward(name string) {
	sc.add(0, name, func() {
		legacy.Nox_client_toggleSpellbook_45AC70()
		e2eLog.Printf("FIELD GUIDE REWARD: closed")
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
	Row      int           `yaml:"row,omitempty"`
	Slot     int           `yaml:"slot,omitempty"`
	Spell    int           `yaml:"spell,omitempty"`
	Item     string        `yaml:"item,omitempty"`
	Creature string        `yaml:"creature,omitempty"`
	Handler  string        `yaml:"handler,omitempty"`
	Expected string        `yaml:"expect-handler,omitempty"`
	Owned    bool          `yaml:"owned-by-player,omitempty"`
	Modifier string        `yaml:"modifier,omitempty"`
	Mask     uint32        `yaml:"mask,omitempty"`
	Count    int           `yaml:"count,omitempty"`
	Amount   int           `yaml:"amount,omitempty"`
	Max      int           `yaml:"max,omitempty"`
	Price    int           `yaml:"price,omitempty"`
	Gold     int           `yaml:"gold,omitempty"`
	Health   int           `yaml:"health,omitempty"`
	Map      string        `yaml:"map,omitempty"`
	Function string        `yaml:"function,omitempty"`
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
		case "switch-map":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.SwitchMap(l.Map, l.Name)
		case "wait-map":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.WaitMap(l.Map, l.Name)
		case "wait-forced-coop-autosave":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.WaitForcedCoopAutosave(l.Name)
		case "call-noxscript-function":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.CallNoxScriptFunction(l.Function, l.Name)
		case "capture-wizard1-urchins":
			sc.CaptureWizard1Urchins(l.Name)
		case "enter-wizard1-urchin-setup-trigger":
			sc.EnterWizard1UrchinSetupTrigger(l.Name)
		case "assert-wizard1-urchins-spawned":
			sc.AssertWizard1UrchinsSpawned(l.Name)
		case "assert-wizard1-lightning-kills":
			sc.AssertWizard1LightningKills(l.Name)
		case "wait-wizard1-multiple-lightning-hits":
			sc.WaitWizard1MultipleLightningHits(l.Name)
		case "move-wizard1-player-near-horvath":
			sc.MoveWizard1PlayerNearHorvath(l.Name)
		case "assert-door-xfer-loaded":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertDoorXferLoaded(l.Name)
		case "assert-trigger-xfer-loaded":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertTriggerXferLoaded(l.Name)
		case "assert-hole-xfer-loaded":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertHoleXferLoaded(l.Name)
		case "assert-transporter-xfer-loaded":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertTransporterXferLoaded(l.Name)
		case "assert-elevator-xfer-loaded":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertElevatorXferLoaded(l.Name)
		case "place-player-on-lava":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.PlacePlayerOnLava(l.Name)
		case "assert-player-lava-damage":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertPlayerLavaDamage(l.Name)
		case "arm-player-poison":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.ArmPlayerPoison(l.Name)
		case "assert-player-poison-damage":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertPlayerPoisonDamage(l.Name)
		case "arm-oval-shield":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.ArmOvalShield(l.Name)
		case "assert-oval-shield-update":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertOvalShieldUpdate(l.Name)
		case "assert-shield-damage":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertShieldDamage(l.Name)
		case "arm-channel-life":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.ArmChannelLife(l.Name)
		case "assert-channel-life-update":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertChannelLifeUpdate(l.Name)
		case "arm-firewalk":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.ArmFirewalk(l.Name)
		case "assert-firewalk-update":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertFirewalkUpdate(l.Name)
		case "arm-greater-heal":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.ArmGreaterHeal(l.Name)
		case "assert-greater-heal-update":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertGreaterHealUpdate(l.Name)
		case "arm-force-of-nature":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.ArmForceOfNature(l.Name)
		case "assert-force-of-nature-charge-removed":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertForceOfNatureChargeRemoved(l.Name)
		case "assert-force-of-nature-launched":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertForceOfNatureLaunched(l.Name)
		case "assert-force-of-nature-completed":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertForceOfNatureCompleted(l.Name)
		case "arm-mana-bomb":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.ArmManaBomb(l.Name)
		case "assert-mana-bomb-update-and-cancel":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertManaBombUpdateAndCancel(l.Name)
		case "assert-mana-bomb-completed":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertManaBombCompleted(l.Name)
		case "arm-chain-lightning":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.ArmChainLightning(l.Name)
		case "arm-duration-ray-draws":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.ArmDurationRayDraws(l.Name)
		case "assert-duration-ray-draws-and-cleanup":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertDurationRayDrawsAndCleanup(l.Name)
		case "assert-chain-lightning-update-and-cancel":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertChainLightningUpdateAndCancel(l.Name)
		case "assert-chain-lightning-completed":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertChainLightningCompleted(l.Name)
		case "arm-energy-bolt":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.ArmEnergyBolt(l.Name)
		case "assert-energy-bolt-update-and-cancel":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertEnergyBoltUpdateAndCancel(l.Name)
		case "assert-energy-bolt-completed":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertEnergyBoltCompleted(l.Name)
		case "arm-drain-mana":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.ArmDrainMana(l.Name)
		case "assert-drain-mana-update-and-cancel":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertDrainManaUpdateAndCancel(l.Name)
		case "assert-drain-mana-completed":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertDrainManaCompleted(l.Name)
		case "arm-turn-undead":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.ArmTurnUndead(l.Name)
		case "assert-turn-undead-active-and-cancel":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertTurnUndeadActiveAndCancel(l.Name)
		case "assert-turn-undead-completed":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertTurnUndeadCompleted(l.Name)
		case "arm-blink":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.ArmBlink(l.Name)
		case "assert-blink-completed":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertBlinkCompleted(l.Name)
		case "arm-swap":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.ArmSwap(l.Name)
		case "assert-swap-completed":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertSwapCompleted(l.Name)
		case "arm-teleport-to-target":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.ArmTeleportToTarget(l.Name)
		case "assert-teleport-to-target-completed":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertTeleportToTargetCompleted(l.Name)
		case "arm-teleport-pop":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.ArmTeleportPop(l.Name)
		case "assert-teleport-pop-completed":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertTeleportPopCompleted(l.Name)
		case "arm-teleport-to-mark":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.ArmTeleportToMark(l.Name)
		case "assert-teleport-to-mark-completed":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertTeleportToMarkCompleted(l.Name)
		case "arm-moonglow":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.ArmMoonglow(l.Name)
		case "assert-moonglow-and-cancel":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertMoonglowAndCancel(l.Name)
		case "assert-moonglow-destroyed":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertMoonglowDestroyed(l.Name)
		case "place-ground-item-on-lava":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.PlaceGroundItemOnLava(l.Name)
		case "assert-ground-item-lava-damage":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertGroundItemLavaDamage(l.Name)
		case "smoke-blast":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.SmokeBlast(l.Name)
		case "damage-poof":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.DamagePoof(l.Name)
		case "mana-bomb-cancel":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.ManaBombCancel(l.Name)
		case "object-death-spawns":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.ObjectDeathSpawns(l.Name)
		case "spawn-monster":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.SpawnMonster(l.Item, image.Pt(l.X, l.Y), l.Name)
		case "select-placed-monster":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.SelectPlacedMonster(l.Item, image.Pt(l.X, l.Y), l.Name)
		case "assert-monster-encounter":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertMonsterEncounter(l.Name)
		case "assert-monster-shield-wear":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertMonsterShieldWear(l.Name)
		case "wait-monster-dead":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.WaitMonsterDead(l.Name)
		case "kill-zombie-for-raise":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.KillZombieForRaise(l.Name)
		case "arm-zombie-raise":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.ArmZombieRaise(l.Name)
		case "wait-zombie-raised":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.WaitZombieRaised(l.Name)
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
			sc.SpawnGroundItem(l.Item, l.Handler, l.Expected, l.Owned, l.Amount, image.Pt(l.X, l.Y), l.Name)
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
		case "drop-ground-trap-by-gameex":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.DropGroundTrapByGameEx(l.Name)
		case "assert-ground-item-consumed-extra-life":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertGroundItemConsumedExtraLife(l.Name)
		case "drag-inventory-item-out":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.DragInventoryItemOut(l.Item, image.Pt(l.X, l.Y), l.Name)
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
		case "assert-server-inventory-count":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertServerInventoryItemCount(l.Item, l.Count, l.Name)
		case "assert-client-inventory-count":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertClientInventoryItemCount(l.Item, l.Count, l.Name)
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
		case "assert-map-shopkeeper-field-guide":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertMapShopkeeperFieldGuide(l.Item, l.Creature, l.Name, l.Count)
		case "open-map-shopkeeper":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.OpenMapShopkeeper(l.Item, l.Name)
		case "acquire-field-guide-fixture":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AcquireFieldGuideFixture(l.Item, l.Name)
		case "assert-field-guide-reward":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertFieldGuideReward(l.Item, l.Name)
		case "assert-field-guide-reward-open":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertFieldGuideRewardOpen(l.Item, l.Name)
		case "close-field-guide-reward":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.CloseFieldGuideReward(l.Name)
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
		case "spell-set-next":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.Key(keybind.KeyE, l.Name)
		case "spell-set-prev":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.Key(keybind.KeyW, l.Name)
		case "spell-set-select":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.Key(keybind.KeyR, l.Name)
		case "assert-spell-set-row":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertSpellSetRow(l.Row, l.Name)
		case "assert-quickbar-expanded":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertQuickbarExpanded(l.Active, l.Name)
		case "toggle-expanded-quickbar":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.add(0, l.Name, legacy.Nox_xxx_quickBarToggle_460920)
		case "set-quickbar-spell":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.SetQuickbarSpell(l.Spell, l.Slot, l.Name)
		case "assert-quickbar-spell":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertQuickbarSpell(l.Spell, l.Slot, l.Name)
		case "assert-native-exit-save-location":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertNativeExitSaveLocation(l.Name)
		case "assert-native-exit-coop-save":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertNativeExitCoopSave(l.Name)
		case "assert-last-spell-slot":
			if dt != 0 {
				sc.Wait(dt, "")
			}
			sc.AssertLastSpellSlot(l.Slot, l.Name)
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
