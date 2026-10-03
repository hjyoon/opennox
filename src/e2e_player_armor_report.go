package opennox

import (
	"fmt"
	"math"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

type e2ePlayerArmorReportOutcome struct {
	current, cached, client uint32
	equipped, held          bool
}

func (o e2ePlayerArmorReportOutcome) validate(baseline uint32, equipped bool) error {
	finite := func(bits uint32) bool { return bits&0x7f800000 != 0x7f800000 }
	if !finite(baseline) || !finite(o.current) || !o.held || o.equipped != equipped ||
		o.current != o.cached || o.current != o.client {
		return fmt.Errorf("armor report: baseline=%08x current=%08x cache=%08x client=%08x equipped=%t want=%t held=%t",
			baseline, o.current, o.cached, o.client, o.equipped, equipped, o.held)
	}
	if equipped {
		if math.Float32frombits(o.current) <= math.Float32frombits(baseline) {
			return fmt.Errorf("equipped armor did not increase: %08x->%08x", baseline, o.current)
		}
	} else if o.current != baseline {
		return fmt.Errorf("dequipped armor did not restore baseline: %08x->%08x", baseline, o.current)
	}
	return nil
}

type e2ePlayerArmorReportFixture struct {
	unit, item *server.Object
	baseline   uint32
}

func (f *e2ePlayerArmorReportFixture) read() (e2ePlayerArmorReportOutcome, error) {
	if f.unit == nil || f.item == nil || noxServer.Players.HostUnit() != f.unit || f.unit.UpdateData == nil {
		return e2ePlayerArmorReportOutcome{}, fmt.Errorf("armor report lost its live player/item binding")
	}
	update := f.unit.UpdateDataPlayer()
	return e2ePlayerArmorReportOutcome{
		current: update.Field57, cached: update.Field58,
		client:   memmap.Uint32(0x5D4594, 1062548), // Actual opcode 73 receiver's value, not a supplied packet.
		equipped: f.item.Flags().Has(object.FlagEquipped), held: f.item.InvHolder == f.unit,
	}, nil
}

func (f *e2ePlayerArmorReportFixture) prepare(typeID string) {
	f.unit = noxServer.Players.HostUnit()
	var err error
	f.item, _, err = e2eInventoryItem(typeID)
	if err != nil || f.unit == nil || f.unit.UpdateData == nil || f.unit.ControllingPlayer() == nil {
		e2eError(fmt.Errorf("armor report requires live player and inventory %q: %v", typeID, err))
		return
	}
	count, err := e2eInventoryItemCount(typeID)
	if err != nil || count != 1 || !f.item.Class().Has(object.ClassArmor) {
		e2eError(fmt.Errorf("armor report requires exactly one stock armor %q: count=%d error=%v", typeID, count, err))
		return
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(f.unit), f.unit.UpdateData, unsafe.Pointer(f.item), unsafe.Pointer(f.unit.ControllingPlayer())} {
			if uintptr(ptr) <= math.MaxUint32 {
				e2eError(fmt.Errorf("armor report fixture pointer %p does not exceed 4 GiB", ptr))
				return
			}
		}
	}
	f.baseline = f.unit.UpdateDataPlayer().Field57
	state, err := f.read()
	if err == nil {
		err = state.validate(f.baseline, false)
	}
	if err != nil {
		e2eError(err)
		return
	}
	e2eLog.Printf("ARMOR REPORT BASELINE: item=%s unit=%p update=%p item_pointer=%p player=%p bits=%08x value=%g cache=%08x client=%08x",
		typeID, f.unit, f.unit.UpdateData, f.item, f.unit.ControllingPlayer(), f.baseline, math.Float32frombits(f.baseline), state.cached, state.client)
}

func (f *e2ePlayerArmorReportFixture) observe(equipped bool, name string) bool {
	state, err := f.read()
	if err == nil {
		err = state.validate(f.baseline, equipped)
	}
	if err != nil {
		if noxServer.Frame()%30 == 0 {
			e2eLog.Printf("ARMOR REPORT WAIT: %s: %v", name, err)
		}
		return false
	}
	e2eLog.Printf("ARMOR REPORT PASS: %s frame=%d baseline=%08x current=%08x value=%g cache=%08x client=%08x equipped=%t held=%t",
		name, noxServer.Frame(), f.baseline, state.current, math.Float32frombits(state.current), state.cached, state.client, state.equipped, state.held)
	return true
}

// CheckPlayerArmorReport uses actual inventory mouse input for two complete
// equip/dequip cycles. It only observes live armor/cache/client data; no armor
// setters, direct equipment calls, or injected network reports are used.
func (sc *e2eScenario) CheckPlayerArmorReport(typeID, name string) {
	f := new(e2ePlayerArmorReportFixture)
	sc.addWhen(0, name+" prepare", 1200, func() bool {
		if noxServer.Players.HostUnit() == nil || noxClient.ClientPlayerUnit() == nil || !nox_client_isConnected() ||
			legacy.Nox_client_inventoryAnimationState() != 2 {
			return false
		}
		_, _, err := e2eInventoryItem(typeID)
		return err == nil
	}, func() { f.prepare(typeID) })
	for cycle := 1; cycle <= 2; cycle++ {
		prefix := fmt.Sprintf("%s cycle %d", name, cycle)
		for _, equipped := range []bool{true, false} {
			stage := prefix + " dequip"
			if equipped {
				stage = prefix + " equip"
			}
			sc.ClickInventoryItem(typeID, stage+" actual inventory input")
			sc.addWhen(1, stage+" report", 120, func() bool { return f.observe(equipped, stage+" report") }, func() {})
			sc.Wait(12, stage+" settle")
			sc.addWhen(0, stage+" stable report", 120, func() bool { return f.observe(equipped, stage+" stable report") }, func() {})
		}
	}
}
