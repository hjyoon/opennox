package legacy

import (
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/server"
)

type mapRestoreObjectsHooks4DBA30 struct {
	playerUnit     func() *server.Object
	moveToMarker   func(playerUnit, marker *server.Object)
	delayedDelete  func(*server.Object)
	resolveMonster func(*server.Object)
	needSync       func(*server.Object)
	addUpdatable   func(*server.Object)
	elevatorLink   func(*server.Object) *server.Object
}

// mapRestoreObjects4DBA30 restores all post-transfer object links and update
// scheduling. Every successor is loaded before callbacks can unlink or queue
// the current object, matching GAME.EXE's PE32 traversal order.
func mapRestoreObjects4DBA30(first *server.Object, saveLocationType uint32, h mapRestoreObjectsHooks4DBA30) {
	for obj := first; obj != nil; {
		next := obj.ObjNext
		if uint32(obj.TypeInd) == saveLocationType {
			playerUnit := h.playerUnit()
			h.moveToMarker(playerUnit, obj)
			playerUnit.ScriptIDVal = obj.ScriptIDVal
			obj.ScriptIDVal = 0
			h.delayedDelete(obj)
		} else {
			class := obj.Class()
			switch {
			case class.Has(object.ClassMonster):
				h.resolveMonster(obj)
			case class.Has(object.ClassElevator):
				if obj.UpdateDataElevator().Field_4 != 0 {
					h.needSync(obj)
				}
			case class.Has(object.ClassElevatorShaft):
				if elevator := h.elevatorLink(obj); elevator != nil && elevator.UpdateDataElevator().Field_4 != 0 {
					h.needSync(obj)
				}
			case class.Has(object.ClassDoor):
				update := obj.UpdateDataDoor()
				if update.CurrentDirection != update.TargetDirection {
					h.addUpdatable(obj)
				}
			}
		}
		obj = next
	}
}

type mapRestoreCleanupHooks4DBA30 struct {
	isOfflineMigratingMonster func(*server.Object) bool
	isCoopPlayerPixie         func(*server.Object) bool
	delayedDelete             func(*server.Object)
}

func mapRestoreCleanup4DBA30(first, firstMissile *server.Object, glyphType uint32, h mapRestoreCleanupHooks4DBA30) {
	for obj := first; obj != nil; {
		next := obj.ObjNext
		if int32(obj.ObjFlags) >= 0 && h.isOfflineMigratingMonster(obj) {
			if obj.Class().Has(object.ClassMonster) && uint32(obj.ObjSubClass)&0x2000 != 0 {
				for item := obj.InvFirstItem; item != nil; {
					nextItem := item.InvNextItem
					if uint32(item.TypeInd) == glyphType {
						h.delayedDelete(item)
					}
					item = nextItem
				}
			}
			h.delayedDelete(obj)
		}
		obj = next
	}
	for obj := firstMissile; obj != nil; {
		next := obj.ObjNext
		if int32(obj.ObjFlags) >= 0 && h.isCoopPlayerPixie(obj) {
			h.delayedDelete(obj)
		}
		obj = next
	}
}

type mapRestoreOwnedHooks4DBA30 struct {
	reportAcquire func(byte, *server.Object)
	monitor       func(byte, *server.Object)
	markMinimap   func(byte, *server.Object)
}

func mapRestoreOwned4DBA30(playerInd byte, first *server.Object, h mapRestoreOwnedHooks4DBA30) {
	for owned := first; owned != nil; owned = owned.Field128 {
		if !owned.Class().Has(object.ClassMonster) {
			continue
		}
		if owned.UpdateData != nil && owned.UpdateDataMonster().StatusFlags.Has(object.MonStatusSummoned) {
			h.reportAcquire(playerInd, owned)
			h.markMinimap(playerInd, owned)
			continue
		}
		if owned.SubClass().AsMonster().Has(object.MonsterMonitor) {
			h.monitor(playerInd, owned)
			h.markMinimap(playerInd, owned)
		}
	}
}

// Sub_4DBA30 restores GAME.EXE 004DBA30 without placing Player, Object,
// MonsterUpdateData, or owned-list links in PE32 integer locals.
func Sub_4DBA30(switchToSolo bool) {
	outer := GetServer()
	srv := outer.S()
	pl := srv.Players.ByInd(ntype.PlayerInd(server.HostPlayerIndex))
	if pl == nil || pl.PlayerUnit == nil {
		return
	}
	playerUnit := pl.PlayerUnit
	saveLocationType := memmap.PtrUint32(0x5D4594, 1563128)
	glyphType := memmap.PtrUint32(0x5D4594, 1563132)
	if *saveLocationType == 0 {
		*saveLocationType = uint32(srv.Types.IndByID("SaveGameLocation"))
		*glyphType = uint32(srv.Types.IndByID("Glyph"))
	}
	if switchToSolo {
		mapRestoreObjects4DBA30(srv.Objs.First(), *saveLocationType, mapRestoreObjectsHooks4DBA30{
			playerUnit: func() *server.Object {
				return pl.PlayerUnit
			},
			moveToMarker: func(playerUnit, marker *server.Object) {
				Nox_xxx_unitMove_4E7010(playerUnit, marker.Pos())
			},
			delayedDelete:  outer.DelayedDelete,
			resolveMonster: monsterResolveXferRefs528DB0,
			needSync: func(obj *server.Object) {
				obj.NeedSync()
			},
			addUpdatable: srv.Objs.AddToUpdatable,
			elevatorLink: (*server.Object).ElevatorLink,
		})
		outer.NoxScriptC().ActResolveObjs()
		resolvePendingOwnsRuntime516FC0()
		if Get_dword_5d4594_1563096() != 0 {
			mapRestoreCleanup4DBA30(srv.Objs.First(), srv.Objs.MissileList, *glyphType, mapRestoreCleanupHooks4DBA30{
				isOfflineMigratingMonster: objectIsUnit_4E5B50,
				isCoopPlayerPixie:         objectIsCoopPlayerPixieRuntime_4E5B80,
				delayedDelete:             outer.DelayedDelete,
			})
		}
	}
	mapRestoreOwned4DBA30(pl.PlayerInd, playerUnit.Field129, mapRestoreOwnedHooks4DBA30{
		reportAcquire: func(playerInd byte, obj *server.Object) {
			Nox_xxx_netReportAcquireCreature_4D91A0(int(playerInd), obj)
		},
		monitor: func(playerInd byte, obj *server.Object) {
			netMonitorCreatureNative4D9250(srv, playerInd, obj)
		},
		markMinimap: func(playerInd byte, obj *server.Object) {
			srv.Players.Nox_xxx_netMarkMinimapObject_417190(ntype.PlayerInd(playerInd), obj, 1)
		},
	})
	Nox_xxx_gameSetSwitchSolo_4DB220(0)
	Set_dword_5d4594_1563096(0)
	Nox_ticks_reset_416D40()
}
