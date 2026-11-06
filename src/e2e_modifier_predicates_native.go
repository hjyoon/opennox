package opennox

import (
	"context"
	"fmt"
	"math"
	"unsafe"

	"github.com/opennox/libs/console"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Playback-only diagnostic: real asset parsing, stock item construction,
// inventory placement, mouse equipment messages and the four public C APIs.
// Strength/slot negatives use private copies of the parsed descriptors.
// No original modifier, game algorithm, parent service or result is replaced.
var e2eModifierNativeRun func([]string) bool

func (sc *e2eScenario) CheckModifierPredicatesNative(mode, name string) {
	sc.addWhen(0, name, 1200, func() bool { return noxServer.Players.HostUnit() != nil }, func() {
		if e2eModifierNativeRun == nil || !e2eModifierNativeRun([]string{mode}) {
			e2eError(fmt.Errorf("modifier native C diagnostic mode %q is unavailable", mode))
		}
	})
}

func init() {
	if e2ePlay == "" {
		return
	}
	var host, item, target *server.Object
	var inversion, grip *server.ModifierEff
	var hostHP, targetHP uint16
	var calls, cases int
	finished := false
	fail := func(message string) { e2eError(fmt.Errorf("modifier native C diagnostic: %s", message)) }
	check := func(label string, inv, gr *server.ModifierEff, want [4]int32) bool {
		if host == nil || item == nil || target == nil || !host.HasItem(item) || item.InvHolder != host || target.ObjOwner != host || target.FindOwnerChainPlayer() != host || host.HealthData.Cur != hostHP || target.HealthData.Cur != targetHP {
			fail("live stock inventory, default owner chain and unchanged health required")
			return false
		}
		for _, p := range []unsafe.Pointer{unsafe.Pointer(host), host.UpdateData, unsafe.Pointer(item), item.InitData, unsafe.Pointer(target), target.UpdateData, unsafe.Pointer(inv), unsafe.Pointer(gr)} {
			if p == nil || (unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32) {
				fail("complete real objects/modifiers and native addresses above4GiB required")
				return false
			}
		}
		var got [4]int32
		for entry := range got {
			effect := inv
			if entry == 1 {
				effect = gr
			}
			scratch := [3]int32{0x13579bdf, 37, 0x2468ace0}
			var ok bool
			got[entry], ok = legacy.ModifierDefendActualDefaultC4E03D0(entry, host, item, target, effect, &scratch[1])
			if !ok || got[entry] != want[entry] || scratch[0] != 0x13579bdf || scratch[2] != 0x2468ace0 || (entry < 2 && scratch[1] != want[entry]) || (entry >= 2 && scratch[1] != 37) {
				fail(fmt.Sprintf("case=%s actual C entry=%d return=%d scratch=%v want=%d", label, entry, got[entry], scratch, want[entry]))
				return false
			}
			calls++
		}
		cases++
		e2eLog.Printf("MODIFIER NATIVE C PASS: case=%s results=%v host=%p item=%p attrs=%p target=%p owner=%p inv=%p grip=%p equipped=%t actual-public-C=true actual-parent-default=true health=unchanged original-assets=unchanged earned-campaign-credit=false", label, got, host, item, item.InitData, target, target.ObjOwner, inv, gr, item.Flags().Has(object.FlagEquipped))
		return true
	}
	handler := func(args []string) bool {
		if len(args) != 1 || finished {
			return false
		}
		switch args[0] {
		case "start":
			if host != nil {
				fail("one prepared diagnostic per product run")
				return true
			}
			host = noxServer.Players.HostUnit()
			if host == nil || host.HealthData == nil || host.HealthData.Cur == 0 || host.UpdateData == nil {
				fail("living hosted player required")
				return true
			}
			functions, ok := legacy.ModifierDefendActualCallbacks4E03D0()
			if !ok {
				fail("native public C signatures and default services required")
				return true
			}
			for mod := noxServer.Modif.ListByType(2); mod != nil; mod = mod.Next() {
				if mod.DefendCollide88.Val < 1 {
					continue
				}
				if inversion == nil && mod.DefendCollide88.Fnc == functions[0] {
					inversion = mod
				}
				if grip == nil && mod.DefendCollide88.Fnc == functions[1] {
					grip = mod
				}
			}
			if inversion == nil || grip == nil {
				fail("both stock parsed positive Inversion/Grip descriptors required")
				return true
			}
			count, err := e2eInventoryItemCount("Sword")
			if err != nil || count != 0 {
				fail("unique new Sword fixture required")
				return true
			}
			attrs := &server.ModifierInitData{Modifiers: [4]*server.ModifierEff{nil, nil, inversion, grip}}
			item = legacy.Nox_xxx_playerRespawnItem_4EF750(host, "Sword", attrs, 1, 0)
			if item == nil || item.InitData == nil || item.InitDataModifier().Modifiers != attrs.Modifiers || item.Flags().Has(object.FlagEquipped) {
				fail("actual respawn/init/modifier/placement services must preserve both stock slots")
				return true
			}
			target = noxServer.NewObjectByTypeID("Troll")
			if target == nil || target.HealthData == nil || target.UpdateData == nil {
				fail("actual stock Troll constructor required")
				return true
			}
			noxServer.CreateObjectAt(target, host, host.PosVec.Add(types.Ptf(32, 0)))
			noxServer.ObjectsAddPending()
			hostHP, targetHP = host.HealthData.Cur, target.HealthData.Cur
			e2eLog.Printf("MODIFIER NATIVE C PREPARED: inv=%s/%d/%d grip=%s/%d/%d stock-descriptors=true constructor-and-placement=actual stock-strengths=unchanged", inversion.Name(), inversion.Index(), inversion.DefendCollide88.Val, grip.Name(), grip.Index(), grip.DefendCollide88.Val)
			check("loose-stock", inversion, grip, [4]int32{1, 0, 1, 0})
		case "equipped":
			if item == nil || !item.Flags().Has(object.FlagEquipped) || host.UpdateDataPlayer().EquippedWeapon != item {
				fail("actual mouse equipment message must equip the prepared Sword")
				return true
			}
			data := item.InitDataModifier()
			original := data.Modifiers
			if !check("equipped-stock", inversion, grip, [4]int32{1, 0, 1, 1}) {
				return true
			}
			inv, freeInv := alloc.New(*inversion)
			gr, freeGrip := alloc.New(*grip)
			defer func() {
				data.Modifiers = original
				freeGrip()
				freeInv()
			}()
			for _, row := range []struct {
				name     string
				strength int32
				want     [4]int32
			}{
				{"negative-copy", -1, [4]int32{0, 1, 0, 0}},
				{"zero-copy", 0, [4]int32{0, 1, 0, 0}},
				{"maximum-copy", math.MaxInt32, [4]int32{1, 0, 1, 1}},
			} {
				inv.DefendCollide88.Val, gr.DefendCollide88.Val = row.strength, row.strength
				data.Modifiers = [4]*server.ModifierEff{nil, nil, inv, gr}
				if !check(row.name, inv, gr, row.want) {
					return true
				}
			}
			data.Modifiers = [4]*server.ModifierEff{inversion, grip, nil, nil}
			if !check("ineligible-slots", inversion, grip, [4]int32{1, 0, 0, 0}) {
				return true
			}
			data.Modifiers = [4]*server.ModifierEff{nil, nil, grip, inversion}
			check("reversed-eligible-slots", inversion, grip, [4]int32{1, 0, 1, 1})
		case "dequipped":
			if item == nil || item.Flags().Has(object.FlagEquipped) || host.UpdateDataPlayer().EquippedWeapon == item {
				fail("actual mouse equipment message must dequip the prepared Sword")
				return true
			}
			if !check("dequipped-stock", inversion, grip, [4]int32{1, 0, 1, 0}) || cases != 8 || calls != 32 {
				fail("all eight cases and32actual C calls must complete")
				return true
			}
			path, err := e2eWriteMagicFrame("", noxClient.r.CopyPixBuffer())
			if err != nil {
				e2eError(err)
				return true
			}
			e2eLog.Printf("MODIFIER NATIVE C VERIFIED: cases=%d calls=%d actual-stock-constructors=true parsed-original-modifiers=true actual-mouse-equip-dequip=true parent-services=default source-objects=live health=unchanged path=%s earned-campaign-credit=false", cases, calls, path)
			noxServer.DelayedDelete(target)
			noxServer.DelayedDelete(item)
			finished = true
		default:
			return false
		}
		return true
	}
	e2eModifierNativeRun = handler
	noxConsole.Register(&console.Command{Token: "e2e-modifier-native", HelpID: "nohelp", Flags: console.Server | console.NoHelp, Func: func(_ context.Context, _ *console.Console, args []string) bool { return handler(args) }})
}
