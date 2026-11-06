package opennox

import (
	"fmt"
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

// This is a prepared actual-product C API diagnostic. Initial injury,
// cooldown and synchronous early-gate fixtures are explicit preparation.
// The world clock, stock descriptor, armor lookup and HP adjustment are real.
// It does not establish naturally earned campaign progression.
// Stock Regeneration1 at 7200, 30 FPS and 150 max HP has a 1440-frame
// interval. Allow its next natural boundary without changing those inputs.
const e2eRawRegenerationWaitFrames = 3600

type e2eRawRegenerationFixture struct {
	host, sword, leather, robe  *server.Object
	modifier                    *server.ModifierEff
	modifierBefore              server.ModifierEff
	health                      *server.HealthData
	maximum, savedHP            uint16
	fps, interval, robeInterval uint32
	cases, calls, restores      int
}

type e2eRawRegenerationCase struct {
	name, item, gate string
	phase            int // -1: any frame; 0: exact stock interval; 1: noninterval
	positive         bool
}

var e2eRawRegenerationCases = []e2eRawRegenerationCase{
	{name: "nil-item", item: "nil", phase: -1},
	{name: "nil-holder", item: "sword", gate: "holder", phase: -1},
	{name: "nil-health", item: "sword", gate: "health", phase: -1},
	{name: "cooldown", item: "sword", gate: "cooldown", phase: -1},
	{name: "dead-mask-20", item: "sword", gate: "dead20", phase: -1},
	{name: "dead-mask-8000", item: "sword", gate: "dead8000", phase: -1},
	{name: "full-health", item: "sword", gate: "full", phase: -1},
	{name: "sword-stock-on-interval", item: "sword", phase: 0, positive: true},
	{name: "sword-stock-off-interval", item: "sword", phase: 1},
	{name: "leather-stock-on-interval", item: "leather", phase: 0, positive: true},
	{name: "robe-stock-on-interval", item: "robe", phase: 0, positive: true},
	{name: "robe-stock-off-interval", item: "robe", phase: 1},
}

func (f *e2eRawRegenerationFixture) item(name string) *server.Object {
	switch name {
	case "sword":
		return f.sword
	case "leather":
		return f.leather
	case "robe":
		return f.robe
	default:
		return nil
	}
}

func (f *e2eRawRegenerationFixture) validate() error {
	if f.host != noxServer.Players.HostUnit() || f.host == nil || f.health == nil || f.host.HealthData != f.health || f.host.UpdateData == nil ||
		f.health.Max != f.maximum || f.host.Flags().HasAny(object.FlagDead|object.FlagDestroyed) ||
		f.sword == nil || !f.sword.Flags().Has(object.FlagEquipped) || f.host.UpdateDataPlayer().EquippedWeapon != f.sword ||
		f.modifier == nil || *f.modifier != f.modifierBefore {
		return fmt.Errorf("live host, equipped stock sword, health identity or immutable Regeneration1 changed")
	}
	for _, item := range []*server.Object{f.sword, f.leather, f.robe} {
		if item == nil || !f.host.HasItem(item) || item.InvHolder != f.host || item.InitData == nil ||
			item.InitDataModifier().Modifiers != [4]*server.ModifierEff{nil, nil, f.modifier, nil} {
			return fmt.Errorf("actual stock construction, inventory or original descriptor slots changed: %p", item)
		}
		for _, pointer := range []unsafe.Pointer{unsafe.Pointer(f.host), unsafe.Pointer(f.health), f.host.UpdateData,
			unsafe.Pointer(item), item.InitData, item.UpdateData, unsafe.Pointer(f.modifier)} {
			if pointer == nil || unsafe.Sizeof(uintptr(0)) != 8 || uintptr(pointer) <= math.MaxUint32 {
				return fmt.Errorf("macOS ARM64 product requires real native data above4GiB: %p", pointer)
			}
		}
	}
	facts, ok := legacy.RegenerationRawActualServices4E01D0(f.sword)
	if !ok || facts.FPS != f.fps || facts.RegisteredUpdate != f.modifier.Update100.Fnc || facts.RawEntry == nil || facts.RawEntry == facts.RegisteredUpdate {
		return fmt.Errorf("ordinary live frame/FPS/armor/HP services, nil observers and distinct raw C entry required")
	}
	return nil
}

func (f *e2eRawRegenerationFixture) prepare() error {
	f.host, f.sword, f.modifier = noxServer.Players.HostUnit(), e2e.engageItem, e2e.engageModifier
	if f.host == nil || f.host.HealthData == nil || f.host.HealthData.Cur == 0 || f.host.HealthData.Max < 3 ||
		f.host != e2e.engageOwner || f.sword == nil || f.modifier == nil || f.modifier.Name() != "Regeneration1" ||
		f.modifier.Update100.Fnc == nil || f.modifier.Update100.Val <= 0 || f.host.Field110&0x20 == 0 {
		return fmt.Errorf("actual stock Regeneration1 grant and real UI equipment required; disabled descriptors are not substituted")
	}
	if noxflags.HasGame(noxflags.GameFlag(0x04000000)) {
		return fmt.Errorf("ordinary HP adjustment is disabled by the existing game flag")
	}
	facts, ok := legacy.RegenerationRawActualServices4E01D0(f.sword)
	if !ok || facts.FPS == 0 || facts.RegisteredUpdate != f.modifier.Update100.Fnc {
		return fmt.Errorf("stock registered callback and actual default raw-C services required")
	}
	f.modifierBefore, f.health = *f.modifier, f.host.HealthData
	f.maximum, f.savedHP = f.health.Max, f.health.Cur
	if f.savedHP != f.maximum {
		return fmt.Errorf("fresh arena fixture must start at full living HP")
	}
	f.fps = facts.FPS
	rate := uint32(f.modifier.Update100.Val)
	f.interval = rate * f.fps / uint32(f.maximum)
	f.robeInterval = (rate / 3) * f.fps / uint32(f.maximum)
	if f.interval <= 1 || f.robeInterval <= 1 || f.interval > e2eRawRegenerationWaitFrames || f.robeInterval > e2eRawRegenerationWaitFrames {
		return fmt.Errorf("stock rates cannot provide bounded on/off interval controls: rate=%d FPS=%d MaxHP=%d intervals=%d/%d", rate, f.fps, f.maximum, f.interval, f.robeInterval)
	}
	attrs := &server.ModifierInitData{Modifiers: [4]*server.ModifierEff{nil, nil, f.modifier, nil}}
	f.leather = legacy.Nox_xxx_playerRespawnItem_4EF750(f.host, "LeatherArmor", attrs, 1, 0)
	f.robe = legacy.Nox_xxx_playerRespawnItem_4EF750(f.host, "WizardRobe", attrs, 1, 0)
	if f.sword.Class().Has(object.ClassArmor) || f.leather == nil || f.robe == nil || !f.leather.Class().Has(object.ClassArmor) || !f.robe.Class().Has(object.ClassArmor) {
		return fmt.Errorf("actual Sword/LeatherArmor/WizardRobe constructor classes required")
	}
	leather, leatherOK := legacy.RegenerationRawActualServices4E01D0(f.leather)
	robe, robeOK := legacy.RegenerationRawActualServices4E01D0(f.robe)
	if !leatherOK || !robeOK || leather.ArmorFlags != 0x8000 || robe.ArmorFlags != 0x4000 {
		return fmt.Errorf("actual original armor table flags changed: leather=%#x robe=%#x", leather.ArmorFlags, robe.ArmorFlags)
	}
	if err := f.validate(); err != nil {
		return err
	}
	e2eLog.Printf("RAW REGEN C PREPARED: host=%p health=%p sword=%p leather=%p robe=%p modifier=%p stock=%s rate=%d max=%d fps=%d intervals=%d/%d raw=%p registered=%p stock-active=true original-descriptor=unchanged clocks=live services=default observers=nil earned-campaign-credit=false", f.host, f.health, f.sword, f.leather, f.robe, f.modifier, f.modifier.Name(), rate, f.maximum, f.fps, f.interval, f.robeInterval, facts.RawEntry, facts.RegisteredUpdate)
	return nil
}

func (f *e2eRawRegenerationFixture) ready(row e2eRawRegenerationCase) bool {
	if row.phase < 0 {
		return true
	}
	interval := f.interval
	if row.item == "robe" {
		interval = f.robeInterval
	}
	if interval <= 1 {
		return false
	}
	on := noxServer.Frame()%interval == 0
	return (row.phase == 0 && on) || (row.phase == 1 && !on)
}

func (f *e2eRawRegenerationFixture) run(row e2eRawRegenerationCase) error {
	if err := f.validate(); err != nil {
		return err
	}
	item, effect := f.item(row.item), f.modifier
	frame := noxServer.Frame()
	originalHP, originalFlags, originalFrame := f.health.Cur, f.host.Flags(), f.host.Frame134
	var holder *server.Object
	if item != nil {
		holder = item.InvHolder
	}
	defer func() {
		f.host.HealthData = f.health
		if item != nil {
			item.InvHolder = holder
		}
		f.host.SetFlags(originalFlags)
		f.host.Frame134 = originalFrame
		legacy.Nox_xxx_unitSetHP_4E4560(f.host, originalHP)
		if f.host.HealthData != f.health || f.health.Cur != originalHP || f.host.Flags() != originalFlags || f.host.Frame134 != originalFrame ||
			(item != nil && item.InvHolder != holder) {
			e2eError(fmt.Errorf("case=%s fixture pointer/HP/flags/cooldown restoration failed", row.name))
			return
		}
		f.restores++
	}()
	before := f.maximum - 2
	if row.gate == "full" {
		before = f.maximum
	}
	legacy.Nox_xxx_unitSetHP_4E4560(f.host, before)
	if f.health.Cur != before {
		return fmt.Errorf("real HP setter did not prepare initial injury")
	}
	f.host.Frame134 = frame - f.fps - 1
	if row.gate == "cooldown" {
		f.host.Frame134 = frame
	}
	if row.gate == "dead20" {
		f.host.SetFlags(f.host.Flags() | object.Flags(0x20))
	}
	if row.gate == "dead8000" {
		f.host.SetFlags(f.host.Flags() | object.Flags(0x8000))
	}
	preparedFlags, preparedFrame := f.host.Flags(), f.host.Frame134
	if row.gate == "holder" {
		item.InvHolder = nil
	}
	if row.gate == "health" {
		f.host.HealthData = nil
	}
	if row.item == "nil" {
		effect = nil
	}
	healthPointerBefore := f.host.HealthData
	var holderBefore *server.Object
	if item != nil {
		holderBefore = item.InvHolder
	}
	healthBefore := *f.health
	attrsBefore := [3]server.ModifierInitData{*f.sword.InitDataModifier(), *f.leather.InitDataModifier(), *f.robe.InitDataModifier()}
	itemHealth := [3]*server.HealthData{f.sword.HealthData, f.leather.HealthData, f.robe.HealthData}
	var itemHealthBefore [3]server.HealthData
	for index, hp := range itemHealth {
		if hp != nil {
			itemHealthBefore[index] = *hp
		}
	}
	if !legacy.RegenerationRawActualDefaultC4E01D0(effect, item) {
		return fmt.Errorf("actual raw public C transport or ordinary services unavailable")
	}
	f.calls++
	want := healthBefore
	if row.positive {
		want.Cur++
	}
	if *f.health != want || f.host.HealthData != healthPointerBefore || (item != nil && item.InvHolder != holderBefore) ||
		f.host.Flags() != preparedFlags || f.host.Frame134 != preparedFrame ||
		noxServer.Frame() != frame || noxServer.TickRate() != f.fps || *f.modifier != f.modifierBefore {
		return fmt.Errorf("case=%s direct C health/whole-health guards/flags/cooldown/clock/descriptor mismatch: HP=%d->%d want=%d frame=%d", row.name, before, f.health.Cur, want.Cur, frame)
	}
	for index, obj := range []*server.Object{f.sword, f.leather, f.robe} {
		if *obj.InitDataModifier() != attrsBefore[index] || obj.HealthData != itemHealth[index] ||
			(itemHealth[index] != nil && *itemHealth[index] != itemHealthBefore[index]) {
			return fmt.Errorf("case=%s item attributes or unrelated complete item health changed", row.name)
		}
	}
	f.cases++
	e2eLog.Printf("RAW REGEN C PASS: case=%s item=%s frame=%d fps=%d hp=%d->%d/%d positive=%t actual-public-two-pointer-C=true hp-adjust=real armor-table=real clocks=live observers=nil whole-health-guards=unchanged descriptor=unchanged prepared-inputs=true earned-campaign-credit=false", row.name, row.item, frame, f.fps, before, f.health.Cur, f.maximum, row.positive)
	return nil
}

func (sc *e2eScenario) CheckRawRegenerationNative(name string) {
	f := &e2eRawRegenerationFixture{}
	sc.addWhen(0, name+" prepare actual stock items", e2eRawRegenerationWaitFrames, func() bool {
		return noxServer.Players.HostUnit() != nil && e2e.engageItem != nil
	}, func() {
		if err := f.prepare(); err != nil {
			e2eError(err)
		}
	})
	for _, row := range e2eRawRegenerationCases {
		row := row
		sc.addWhen(0, name+" "+row.name, e2eRawRegenerationWaitFrames, func() bool { return f.ready(row) }, func() {
			if err := f.run(row); err != nil {
				e2eError(err)
			}
		})
	}
	sc.Wait(2, name+" publish restored fixture HP")
	sc.Screen(name + " actual raw regeneration complete")
	sc.add(0, name+" finalize raw C diagnostic", func() {
		if err := f.validate(); err != nil {
			e2eError(err)
			return
		}
		if f.cases != 12 || f.calls != 12 || f.restores != 12 || f.health.Cur != f.savedHP {
			e2eError(fmt.Errorf("all12 direct C conditions and exact fixture HP/cooldown/flags restoration required"))
			return
		}
		noxServer.DelayedDelete(f.leather)
		noxServer.DelayedDelete(f.robe)
		e2eLog.Printf("RAW REGEN C VERIFIED: cases=%d calls=%d positive=3 negative=9 stock-Regeneration1=active actual-stock-items=true actual-mouse-equipment=true actual-public-two-pointer-C=true native-data-above4GiB=true services=default observers=nil original-descriptor=unchanged fixture-restored=true prepared-inputs=true earned-campaign-credit=false", f.cases, f.calls)
	})
}
