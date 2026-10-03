package opennox

import (
	"fmt"
	"image"
	"math"
	"unsafe"

	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

type e2ePlayerItemEnchantmentReportOutcome struct {
	current                uint32
	cached, client         byte
	equipped, held, weapon bool
}

func (o e2ePlayerItemEnchantmentReportOutcome) validate(baseline, effect uint32, equipped bool) error {
	want := baseline
	if equipped {
		want |= effect
	}
	if effect == 0 || baseline&effect != 0 || o.current != want || o.cached != byte(want) ||
		o.client != byte(want) || !o.held || o.equipped != equipped || o.weapon != equipped {
		return fmt.Errorf("item enchantment report: server=%08x want=%08x cache=%02x client=%02x equipped=%t/%t weapon=%t held=%t",
			o.current, want, o.cached, o.client, o.equipped, equipped, o.weapon, o.held)
	}
	return nil
}

type e2ePlayerItemEnchantmentReportFixture struct {
	unit, item       *server.Object
	baseline, effect uint32
	tooltip          string
}

func (f *e2ePlayerItemEnchantmentReportFixture) read() (e2ePlayerItemEnchantmentReportOutcome, error) {
	if f.unit == nil || f.item == nil || noxServer.Players.HostUnit() != f.unit ||
		f.unit.UpdateData == nil || f.unit.ControllingPlayer() == nil {
		return e2ePlayerItemEnchantmentReportOutcome{}, fmt.Errorf("item enchantment report lost its live player/item binding")
	}
	return e2ePlayerItemEnchantmentReportOutcome{
		current: f.unit.Field110, cached: f.unit.ControllingPlayer().Field2172,
		client:   memmap.Uint8(0x5D4594, 1062536), // The real opcode 91 client receiver, never an injected packet.
		equipped: f.item.Flags().Has(object.FlagEquipped), held: f.item.InvHolder == f.unit,
		weapon: f.unit.UpdateDataPlayer().EquippedWeapon == f.item,
	}, nil
}

func (f *e2ePlayerItemEnchantmentReportFixture) prepare(typeID string) {
	f.unit, f.item = e2e.engageOwner, e2e.engageItem
	f.baseline, f.effect = e2e.engageOwnerMaskBefore, e2e.engageOwnerMask
	count, err := e2eInventoryItemCount(typeID)
	if err != nil || count != 1 || typeID != "Sword" || e2e.engageItemTypeID != typeID ||
		f.baseline != 0 || f.effect != 1 || memmap.Uint32(0x5D4594, 1062540) != 0 {
		e2eError(fmt.Errorf("item enchantment GUI needs one stock FireProtect1 Sword and no pre-existing icons: count=%d baseline=%08x effect=%08x error=%v",
			count, f.baseline, f.effect, err))
		return
	}
	state, err := f.read()
	if err == nil {
		err = state.validate(f.baseline, f.effect, false)
	}
	if err != nil {
		e2eError(err)
		return
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(f.unit), f.unit.UpdateData, unsafe.Pointer(f.item), unsafe.Pointer(f.unit.ControllingPlayer())} {
			if uintptr(ptr) <= math.MaxUint32 {
				e2eError(fmt.Errorf("item enchantment fixture pointer %p does not exceed 4 GiB", ptr))
				return
			}
		}
	}
	f.tooltip = noxClient.Strings().GetStringInFile("modifier.db:FireProtectItemEnchantDesc", `C:\NoxPost\src\common\Object\Modifier.c`)
	if f.tooltip == "" {
		e2eError(fmt.Errorf("stock fire protection tooltip is empty"))
		return
	}
	e2eLog.Printf("ITEM ENCHANTMENT REPORT BASELINE: unit=%p update=%p item=%p player=%p server=%08x cache=%02x client=%02x tooltip=%q",
		f.unit, f.unit.UpdateData, f.item, f.unit.ControllingPlayer(), state.current, state.cached, state.client, f.tooltip)
}

func (f *e2ePlayerItemEnchantmentReportFixture) observe(equipped bool, name string) bool {
	state, err := f.read()
	if err == nil {
		err = state.validate(f.baseline, f.effect, equipped)
	}
	if err != nil {
		if noxServer.Frame()%30 == 0 {
			e2eLog.Printf("ITEM ENCHANTMENT REPORT WAIT: %s: %v", name, err)
		}
		return false
	}
	if equipped {
		// Only inspect the cache populated by normal inventory rendering. Do not
		// call the icon loader/draw path to manufacture a successful observation.
		handle := *memmap.PtrPtr(0x587000, 27380) // FireProtect: row 2, image slot +12.
		if memmap.Uint32(0x5D4594, 251624) == 0 || handle == nil ||
			noxClient.r.Bag.AsImage(noxrender.ImageHandle(handle)) == nil ||
			(unsafe.Sizeof(uintptr(0)) == 8 && uintptr(handle) <= math.MaxUint32) {
			return false
		}
		e2eLog.Printf("ITEM ENCHANTMENT ICON: %s native_handle=%p loaded=%d", name, handle, memmap.Uint32(0x5D4594, 251624))
	}
	e2eLog.Printf("ITEM ENCHANTMENT REPORT PASS: %s frame=%d server=%08x cache=%02x client=%02x equipped=%t held=%t weapon=%t",
		name, noxServer.Frame(), state.current, state.cached, state.client, state.equipped, state.held, state.weapon)
	return true
}

func (f *e2ePlayerItemEnchantmentReportFixture) hover() {
	win := legacy.InventoryWindow()
	if win == nil || legacy.Nox_client_inventoryAnimationState() != 2 {
		e2eError(fmt.Errorf("inventory is not open for item enchantment hover"))
		return
	}
	pos := win.GlobalPos().Add(image.Pt(25, 244))
	e2eQueueInput(&seat.MouseMoveEvent{Pos: pos, Relative: false})
	e2eLog.Printf("ITEM ENCHANTMENT HOVER INPUT: mouse=%v", pos)
}

func (f *e2ePlayerItemEnchantmentReportFixture) observeTooltip(equipped bool, name string) bool {
	state, err := f.read()
	if err != nil || state.validate(f.baseline, f.effect, equipped) != nil || memmap.Uint32(0x5D4594, 1062540) != 0 {
		return false
	}
	want := ""
	if equipped {
		want = f.tooltip
	}
	if got := e2eObjectTooltipText(); got != want || (equipped && got == "") {
		return false
	}
	e2eLog.Printf("ITEM ENCHANTMENT TOOLTIP PASS: %s equipped=%t text=%q", name, equipped, e2eObjectTooltipText())
	return true
}

// CheckPlayerItemEnchantmentReport observes two actual mouse equip/dequip
// cycles, the reliable report/cache/client value, native render image cache,
// and normal GUI hover text. Only the prior stock item grant is a fixture.
// It does not set masks, send reports, call equipment callbacks or draw icons.
func (sc *e2eScenario) CheckPlayerItemEnchantmentReport(typeID, name string) {
	f := new(e2ePlayerItemEnchantmentReportFixture)
	sc.addWhen(0, name+" prepare", 1200, func() bool {
		return noxServer.Players.HostUnit() != nil && noxClient.ClientPlayerUnit() != nil &&
			nox_client_isConnected() && legacy.Nox_client_inventoryAnimationState() == 2 && e2e.engageItem != nil
	}, func() { f.prepare(typeID) })
	for cycle := 1; cycle <= 2; cycle++ {
		for _, equipped := range []bool{true, false} {
			stage := fmt.Sprintf("%s cycle %d dequip", name, cycle)
			if equipped {
				stage = fmt.Sprintf("%s cycle %d equip", name, cycle)
			}
			sc.ClickInventoryItem(typeID, stage+" actual inventory input")
			sc.addWhen(1, stage+" report", 120, func() bool { return f.observe(equipped, stage+" report") }, func() {})
			sc.Wait(12, stage+" settle")
			sc.addWhen(0, stage+" stable report", 120, func() bool { return f.observe(equipped, stage+" stable report") }, func() {})
			sc.add(0, stage+" hover input", func() { f.hover() })
			sc.addWhen(1, stage+" tooltip", 120, func() bool { return f.observeTooltip(equipped, stage+" tooltip") }, func() {})
		}
	}
}
