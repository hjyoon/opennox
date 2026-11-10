package legacy

/*
#include "GAME4.h"
#include "GAME4_1.h"
*/
import "C"

import (
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

type game4WaypointRuntime struct {
	delete    func(*server.Object)
	move      func(*server.Object, types.Pointf)
	newObject func(string) *server.Object
	create    func(*server.Object, *server.Object, types.Pointf)
	buffed    func(*server.Object) bool
	confused  func(*server.Object) server.Dir16
}

var game4WaypointServices = func() game4WaypointRuntime {
	return game4WaypointRuntime{
		delete:    func(o *server.Object) { GetServer().DelayedDelete(o) },
		move:      Nox_xxx_unitMove_4E7010,
		newObject: func(name string) *server.Object { return GetServer().S().NewObjectByTypeID(name) },
		create:    func(o, owner *server.Object, pos types.Pointf) { GetServer().CreateObjectAt(o, owner, pos) },
		buffed:    func(o *server.Object) bool { return C.nox_xxx_testUnitBuffs_4FF350(asObjectC(o), 3) != 0 },
		confused:  func(o *server.Object) server.Dir16 { return GetServer().S().PlayerConfusedDirection4F7A40(o) },
	}
}

//export sub_4F7950
func sub_4F7950(unit *C.nox_object_t) {
	o := asObjectS(unit)
	if o == nil || o.UpdateData == nil {
		return
	}
	u := (*server.PlayerUpdateData)(o.UpdateData)
	r := game4WaypointServices()
	for i := range u.CustomWaypoints {
		if wp := u.CustomWaypoints[i]; wp != nil {
			r.delete(wp)
		}
		u.CustomWaypoints[i] = nil
	}
	u.CustomWaypointRead, u.CustomWaypointWrite = 0, 0
}

//export nox_xxx_playerSetCustomWP_4F79A0
func nox_xxx_playerSetCustomWP_4F79A0(unit *C.nox_object_t, x, y C.float) {
	o := asObjectS(unit)
	if o == nil || o.UpdateData == nil {
		return
	}
	u := (*server.PlayerUpdateData)(o.UpdateData)
	if u.Player == nil || u.Player.Field3680&3 != 0 || int(u.CustomWaypointWrite) >= len(u.CustomWaypoints) {
		return
	}
	i := u.CustomWaypointWrite
	r := game4WaypointServices()
	pos := types.Pointf{X: float32(x), Y: float32(y)}
	if wp := u.CustomWaypoints[i]; wp != nil {
		r.move(wp, pos)
	} else {
		wp = r.newObject("PlayerWaypoint")
		u.CustomWaypoints[i] = wp
		r.create(u.CustomWaypoints[i], o, pos)
	}
}

//export sub_4F9A80
func sub_4F9A80(unit *C.nox_object_t) C.int {
	o := asObjectS(unit)
	if o == nil || o.UpdateData == nil {
		return 0
	}
	u := (*server.PlayerUpdateData)(o.UpdateData)
	if int(u.CustomWaypointRead) < len(u.CustomWaypoints) && u.CustomWaypoints[u.CustomWaypointRead] != nil {
		return 1
	}
	return 0
}

//export sub_4F9AB0
func sub_4F9AB0(unit *C.nox_object_t) C.int {
	o := asObjectS(unit)
	if o == nil || o.UpdateData == nil {
		return 0
	}
	u := (*server.PlayerUpdateData)(o.UpdateData)
	if int(u.CustomWaypointRead) >= len(u.CustomWaypoints) {
		return 0
	}
	wp := u.CustomWaypoints[u.CustomWaypointRead]
	if wp == nil {
		return 0
	}
	r := game4WaypointServices()
	d := types.Pointf{X: wp.PosVec.X - o.PosVec.X, Y: wp.PosVec.Y - o.PosVec.Y}
	if float64(d.Y)*float64(d.Y)+float64(d.X*d.X) >= 100 {
		o.Direction2 = server.DirFromVec(d)
		if r.buffed(o) {
			o.Direction2 = r.confused(o)
		}
		i := 8 * int(int16(o.Direction2))
		o.ForceVec.X += memmap.Float32(0x587000, uintptr(194136+i)) * o.SpeedCur
		fy := memmap.Float32(0x587000, uintptr(194140+i)) * o.SpeedCur
		o.ForceVec.Y = float32(float64(o.ForceVec.Y) + float64(fy))
		return 1
	}
	r.delete(wp)
	// The callback may change the read cursor; the original reloads it here.
	if int(u.CustomWaypointRead) < len(u.CustomWaypoints) {
		u.CustomWaypoints[u.CustomWaypointRead] = nil
	}
	return 0
}

//export nox_xxx_playerCanMove_4F9BC0
func nox_xxx_playerCanMove_4F9BC0(unit *C.nox_object_t) C.int {
	return C.int(server.PlayerCanMove4F9BC0(asObjectS(unit)))
}

//export nox_xxx_playerCanAttack_4F9C40
func nox_xxx_playerCanAttack_4F9C40(unit *C.nox_object_t) C.int {
	return C.int(server.PlayerCanAttack4F9C40(asObjectS(unit)))
}

//export sub_4FA280
func sub_4FA280(mask C.int) C.int {
	for bit := uint(2); bit < 27; bit++ {
		if uint32(mask)&(1<<bit) != 0 {
			return C.int(int32(memmap.Uint32(0x587000, uintptr(215824+4*bit))))
		}
	}
	return 0
}
