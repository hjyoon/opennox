package opennox

import (
	"fmt"
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

// The stock script kills the preplaced Apprentice2 in ApprenticeStart, then
// moves that corpse to HorvathApprentice five frames after the real Done input.
// Only player placement/open robe-room access are fixture inputs. NPC health,
// death, position, dialog callbacks, timer and client state are never supplied.
func (sc *e2eScenario) CheckWoundedApprentice(name string) {
	var host, wounded, corpse *server.Object
	var original, woundedPos, corpsePos types.Pointf
	var woundedHP, corpseHP uint16
	var opened, closed uint32
	identities := func() bool {
		a, b := 0, 0
		for obj := noxServer.Objs.First(); obj != nil; obj = obj.Next() {
			if obj.EqualID("HorvathApprentice") {
				a++
				if obj != wounded {
					return false
				}
			}
			if obj.EqualID("Apprentice2") {
				b++
				if obj != corpse {
					return false
				}
			}
		}
		return a == 1 && b == 1
	}
	sc.addWhen(0, name+" stock fixture", 1200, func() bool {
		return noxServer.Players.HostUnit() != nil && noxClient.ClientPlayerUnit() != nil && legacy.Get_dword_5d4594_1548524() == 0
	}, func() {
		if e2eMapBaseName(legacy.Nox_xxx_mapGetMapName_409B40()) != "wiz01a" {
			e2eError(fmt.Errorf("WoundedApprentice requires stock Wiz01a"))
			return
		}
		host = noxServer.Players.HostUnit()
		ns := noxServer.noxScriptP()
		wounded, corpse = ns.asObj(ns.Object("HorvathApprentice")), ns.asObj(ns.Object("Apprentice2"))
		for _, unit := range []*server.Object{wounded, corpse} {
			if unit != nil {
				e2eLog.Printf("WOUNDED APPRENTICE INITIAL: unit=%p type=%s world=%t health=%+v update=%p flags=%#x pos=%v", unit, unit.ObjectTypeC().ID(), e2eObjectInWorld(unit), unit.HealthData, unit.UpdateData, uint32(unit.Flags()), unit.PosVec)
			}
			// WoundedApprentice is the original scripted talking prop, with
			// zero max health. Only the ordinary NPC copy must start alive.
			if unit == nil || !e2eObjectInWorld(unit) || unit.HealthData == nil || unit.UpdateData == nil || unit.Flags().HasAny(object.FlagDead|object.FlagDestroyed) || unit == corpse && unit.HealthData.Cur == 0 {
				e2eError(fmt.Errorf("Wiz01a preplaced apprentice record is invalid: %p", unit))
				return
			}
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unit.CObj()) <= math.MaxUint32 {
				e2eError(fmt.Errorf("Wiz01a apprentice is not native-width: %p", unit))
				return
			}
		}
		if wounded == corpse || wounded.ObjectTypeC().ID() != "WoundedApprentice" || corpse.ObjectTypeC().ID() != "NPC" || !identities() {
			e2eError(fmt.Errorf("Wiz01a apprentice identities/type records differ"))
			return
		}
		original, woundedPos, corpsePos = host.PosVec, wounded.PosVec, corpse.PosVec
		woundedHP, corpseHP = wounded.HealthData.Cur, corpse.HealthData.Cur
		asObjectS(host).SetPos(woundedPos.Add(types.Ptf(0, -40)))
		// The stock room trigger registers this dialog when either robe
		// door is disabled/absent. If present, supply access through the
		// normal on/off API. Dispatch that original trigger, not death.
		door := ns.Object("RobeDoor1")
		_, room := noxServer.S().NoxScriptVM.FuncByName("PlayerInRobeRoom")
		if room < 0 {
			e2eError(fmt.Errorf("Wiz01a stock robe-room entry is missing"))
			return
		}
		if door != nil {
			door.Enable(false)
		}
		if err := noxServer.S().NoxScriptVM.CallByIndex(room, host, host); err != nil {
			e2eError(err)
			return
		}
		_, start := noxServer.S().NoxScriptVM.FuncByName("ApprenticeStart")
		_, end := noxServer.S().NoxScriptVM.FuncByName("ApprenticeEnd")
		ud := wounded.UpdateDataMonster()
		if start < 0 || end < 0 || ud.DialogStartFunc != int32(start) || ud.DialogEndFunc != int32(end) || corpse.HealthData.Cur != corpseHP {
			e2eError(fmt.Errorf("stock robe-room trigger failed to register apprentice dialog: %d/%d", ud.DialogStartFunc, ud.DialogEndFunc))
			return
		}
		e2eLog.Printf("WOUNDED APPRENTICE STOCK: wounded=%p type=%s HP=%d pos=%v copy=%p type=%s HP=%d pos=%v damage=%p start=%d end=%d preplaced=true fixture=player-position,open-robe-room", wounded, wounded.ObjectTypeC().ID(), woundedHP, woundedPos, corpse, corpse.ObjectTypeC().ID(), corpseHP, corpsePos, corpse.Damage, start, end)
	})
	sc.add(20, name+" stock StartDialog", func() {
		opened = noxServer.Frame()
		ns := noxServer.noxScriptP()
		ns.StartDialog(ns.toObj(wounded), ns.GetHost())
	})
	sc.addWhen(0, name+" stock dialog death", 300, func() bool {
		root := legacy.Get_dword_5d4594_1123524()
		return root != nil && !root.GetFlags().IsHidden() && host.UpdateDataPlayer().DialogWith == wounded
	}, func() {
		if corpse.HealthData.Cur != 0 || !corpse.Flags().Has(object.FlagDead) || corpse.PosVec != corpsePos || wounded.HealthData.Cur != woundedHP || wounded.PosVec != woundedPos || !identities() {
			e2eError(fmt.Errorf("ApprenticeStart did not kill the preplaced copy before dialog close: HP=%d->%d flags=%#x position=%v", corpseHP, corpse.HealthData.Cur, uint32(corpse.Flags()), corpse.PosVec))
			return
		}
		e2eLog.Printf("WOUNDED APPRENTICE DEATH PASS: stock=ApprenticeStart source=nil weapon=nil raw=500 type=BLADE HP=%d->0 flags=%#x frame=%d copy-position=unchanged talking-prop=unchanged", corpseHP, uint32(corpse.Flags()), noxServer.Frame())
	})
	sc.ClickNPCDialogDone(name + " real Done input")
	sc.addWhen(0, name+" actual close", 300, func() bool {
		root := legacy.Get_dword_5d4594_1123524()
		return host.UpdateDataPlayer().DialogWith == nil && (root == nil || root.GetFlags().IsHidden())
	}, func() { closed = noxServer.Frame() })
	sc.addWhen(0, name+" stock corpse replacement", 300, func() bool {
		return corpse.PosVec.Sub(woundedPos).Len() < .5 && wounded.PosVec.Sub(types.Ptf(4279, 5285)).Len() < .5
	}, func() {
		ud := wounded.UpdateDataMonster()
		if corpse.HealthData.Cur != 0 || !corpse.Flags().Has(object.FlagDead) || wounded.HealthData.Cur != woundedHP || ud.DialogStartFunc != -1 || ud.DialogEndFunc != -1 || !identities() {
			e2eError(fmt.Errorf("stock KillApprentice did not preserve dead copy/canceled dialog"))
			return
		}
		e2eLog.Printf("WOUNDED APPRENTICE REPLACEMENT PASS: original-timer=5 copy=%p HP=0 dead=true pos=%v wounded=%p hidden-pos=%v dialog=-1/-1 identities=unchanged frame=%d elapsed-open=%d elapsed-close=%d", corpse, corpse.PosVec, wounded, wounded.PosVec, noxServer.Frame(), noxServer.Frame()-opened, noxServer.Frame()-closed)
	})
	sc.Wait(60, name+" ordinary death animation")
	sc.addWhen(0, name+" client corpse", 300, func() bool {
		drawable := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(corpse)))
		// Complex-object updates transmit dying/dead animation 9/10,
		// not the server's FlagDead bit. NPCDraw maps those to 1/2.
		return drawable != nil && (drawable.AnimInd == 9 || drawable.AnimInd == 10) && math.Abs(float64(drawable.Pos().X)-float64(woundedPos.X)) <= 1 && math.Abs(float64(drawable.Pos().Y)-float64(woundedPos.Y)) <= 1
	}, func() {
		drawable := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(corpse)))
		if !identities() || corpse.HealthData.Cur != 0 || !corpse.Flags().Has(object.FlagDead) {
			e2eError(fmt.Errorf("apprentice corpse changed after ordinary updates"))
			return
		}
		path, err := e2eWriteMagicFrame("", noxClient.r.CopyPixBuffer())
		if err != nil {
			e2eError(err)
			return
		}
		e2eLog.Printf("WOUNDED APPRENTICE CLIENT PASS: drawable=%p flags=%#x pos=%v death-animation=%d copy=%p HP=0 dead=true new-NPC-created=false path=%s", drawable, uint32(drawable.Flags()), drawable.Pos(), drawable.AnimInd, corpse, path)
		asObjectS(host).SetPos(original)
	})
}
