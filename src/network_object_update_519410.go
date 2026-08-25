package opennox

import (
	"encoding/binary"
	"math"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/server"
)

const objectSpecialMask527E50 = uint32(0x0fff0000)

type objectSpecialHooks527E50 struct {
	reportAnimation func()
	reportHealth    func()
	reportHidden    func()
	reportXStatus   func()
	reportHeight    func()
	reportEnchant   func()
	reportTeamBase  func()
	reportNPC       func()
}

// consumeObjectSpecialFlags527E50 preserves the ordered tests, callbacks and
// post-callback bit clears of GAME.EXE 00527E50.
func consumeObjectSpecialFlags527E50(state *uint32, hooks objectSpecialHooks527E50) bool {
	if state == nil || *state&objectSpecialMask527E50 == 0 {
		return false
	}
	steps := [...]struct {
		bit  uint32
		call func()
	}{
		{0x00010000, hooks.reportAnimation},
		{0x00020000, hooks.reportHealth},
		{0x00040000, hooks.reportHidden},
		{0x00080000, hooks.reportXStatus},
		{0x00400000, hooks.reportHeight},
		{0x00800000, hooks.reportEnchant},
		{0x02000000, hooks.reportTeamBase},
		{0x04000000, hooks.reportNPC},
	}
	for _, step := range steps {
		if *state&step.bit == 0 {
			continue
		}
		if step.call != nil {
			step.call()
		}
		*state &^= step.bit
	}
	return true
}

func (s *Server) netUpdateObjectSpecialNative527E50(recipient, obj *server.Object) bool {
	if recipient == nil || obj == nil {
		return true
	}
	pl := recipient.ControllingPlayer()
	if pl == nil {
		return true
	}
	ind := pl.PlayerIndex()
	if ind < 0 || int(ind) >= len(obj.Field140) {
		return true
	}
	return consumeObjectSpecialFlags527E50(&obj.Field140[ind], objectSpecialHooks527E50{
		reportAnimation: func() { s.Nox_xxx_netReportAnimFrame_4D81F0(int(ind), obj) },
		reportHealth: func() {
			if !s.IsEnemyTo(recipient, obj) {
				s.CurrentHPReport4D8620(int32(ind), obj)
			}
		},
		reportHidden:   func() { s.netReportObjectHiddenNative4D8FD0(ind, obj) },
		reportXStatus:  func() { s.netReportXStatusNative4D8230(ind, obj) },
		reportHeight:   func() { s.netReportUnitHeightNative4D9020(ind, obj) },
		reportEnchant:  func() { s.netReportEnchant_4D8F90(ind, obj) },
		reportTeamBase: func() { s.netReportTeamBaseNative4D92D0(ind, obj) },
		reportNPC:      func() { s.netSendReportNPCNative4D93A0(ind, obj) },
	})
}

func (s *Server) netReportXStatusNative4D8230(ind ntype.PlayerInd, obj *server.Object) {
	var packet [7]byte
	packet[0] = byte(netmsg.MSG_REPORT_X_STATUS)
	binary.LittleEndian.PutUint16(packet[1:], uint16(s.GetUnitNetCode(obj)))
	binary.LittleEndian.PutUint32(packet[3:], obj.Field5)
	s.NetSendPacketXxx0(int(ind), packet[:], nil, 1)
}

func (s *Server) netReportObjectHiddenNative4D8FD0(ind ntype.PlayerInd, obj *server.Object) {
	if !obj.Class().Has(object.ClassVisibleEnable) {
		return
	}
	op := netmsg.MSG_DISABLE_OBJECT
	if obj.Flags().Has(object.FlagEnabled) {
		op = netmsg.MSG_ENABLE_OBJECT
	}
	var packet [3]byte
	packet[0] = byte(op)
	binary.LittleEndian.PutUint16(packet[1:], uint16(s.GetUnitNetCode(obj)))
	s.NetSendPacketXxx0(int(ind), packet[:], nil, 1)
}

func truncObjectHeight4D9020(v float32) byte {
	return byte(int64(v))
}

func (s *Server) netReportUnitHeightNative4D9020(ind ntype.PlayerInd, obj *server.Object) {
	if byte(obj.Field5)&0x20 != 0 {
		packet := [6]byte{
			159,
			byte(obj.NetCode), byte(obj.NetCode >> 8),
			truncObjectHeight4D9020(obj.ZVal),
			truncObjectHeight4D9020(obj.Field27),
			truncObjectHeight4D9020(math.Float32frombits(obj.Field29)),
		}
		s.NetList.AddToMsgListCli(ind, netlist.Kind1, packet[:])
		return
	}
	z := obj.ZVal
	op := netmsg.MSG_REPORT_Z_PLUS
	if z < 0 {
		op = netmsg.MSG_REPORT_Z_MINUS
		z = -z
	}
	var packet [4]byte
	packet[0] = byte(op)
	binary.LittleEndian.PutUint16(packet[1:], uint16(s.GetUnitNetCode(obj)))
	packet[3] = truncObjectHeight4D9020(z)
	s.NetList.AddToMsgListCli(ind, netlist.Kind1, packet[:])
}

func (s *Server) netReportTeamBaseNative4D92D0(ind ntype.PlayerInd, obj *server.Object) {
	const eligible = object.ClassFlag | object.ClassWeapon | object.ClassArmor | object.ClassWand
	if !obj.Class().HasAny(eligible) && int(obj.TypeInd) != s.Types.IndByID("TeamBase") {
		return
	}
	var packet [7]byte
	packet[0] = byte(netmsg.MSG_REPORT_MODIFIER)
	binary.LittleEndian.PutUint16(packet[1:], uint16(s.GetUnitNetCode(obj)))
	if obj.InitData != nil {
		for i, mod := range obj.InitDataModifier().Modifiers {
			packet[i+3] = 0xff
			if mod != nil {
				packet[i+3] = byte(mod.Index())
			}
		}
	} else {
		for i := 3; i < len(packet); i++ {
			packet[i] = 0xff
		}
	}
	s.NetSendPacketXxx0(int(ind), packet[:], nil, 1)
}

func (s *Server) netSendReportNPCNative4D93A0(ind ntype.PlayerInd, obj *server.Object) {
	if obj.UpdateData == nil || !obj.Class().Has(object.ClassMonster) {
		return
	}
	var packet [21]byte
	packet[0] = byte(netmsg.MSG_REPORT_NPC)
	code := uint16(s.GetUnitNetCode(obj))
	if obj.Poison540 != 0 {
		code |= 0x8000
	}
	binary.LittleEndian.PutUint16(packet[1:], code)
	for i, col := range obj.UpdateDataMonster().Color {
		copy(packet[3+3*i:], []byte{col.R, col.G, col.B})
	}
	s.NetSendPacketXxx1(int(ind), packet[:], nil, 1)
}

func objectCoord519410(v float32) uint16 {
	return uint16(int32(math.RoundToEven(float64(v))))
}

func (s *Server) simpleObjectPacketNative5188A0(obj *server.Object) [9]byte {
	var packet [9]byte
	packet[0] = byte(netmsg.MSG_SIMPLE_OBJ)
	binary.LittleEndian.PutUint16(packet[1:], uint16(s.GetUnitNetCode(obj)))
	binary.LittleEndian.PutUint16(packet[3:], obj.TypeInd)
	binary.LittleEndian.PutUint16(packet[5:], objectCoord519410(obj.PosVec.X))
	binary.LittleEndian.PutUint16(packet[7:], objectCoord519410(obj.PosVec.Y))
	return packet
}

func (s *Server) phantomObjectPacketNative5187E0(obj *server.Object) [11]byte {
	var packet [11]byte
	packet[0] = byte(netmsg.MSG_COMPLEX_OBJ)
	binary.LittleEndian.PutUint16(packet[1:], uint16(s.GetUnitNetCode(obj)))
	binary.LittleEndian.PutUint16(packet[3:], obj.TypeInd)
	binary.LittleEndian.PutUint16(packet[5:], objectCoord519410(obj.PosVec.X))
	binary.LittleEndian.PutUint16(packet[7:], objectCoord519410(obj.PosVec.Y))
	packet[9] = playerObjectDirection518C30(obj.Direction1) << 4
	packet[10] = 0xff
	if obj.Class().Has(object.ClassMissile) && uint32(obj.SubClass())&0x30 != 0 {
		packet[10] = byte(uint16(obj.Direction1) >> 3)
	}
	return packet
}

func (s *Server) monsterAnimationNative533790(obj *server.Object) byte {
	ud := obj.UpdateDataMonster()
	animation := byte(8)
	if ind := int(ud.AIStackInd); ind >= 0 && ind < len(ud.AIStack) {
		switch ud.AIStack[ind].Action {
		case 7, 8, 0x0a, 0x0d, 0x1d, 0x24, 0x25:
			animation = byte((uint32(ud.StatusFlags)&0x4000 | 0x30000) >> 14)
		case 9:
			animation = 12
		case 0x10:
			animation = 1
		case 0x11:
			animation = 3
		case 0x12, 0x13, 0x14:
			animation = 7
		case 0x15, 0x17:
			animation = 5
		case 0x16:
			animation = 6
		case 0x18:
			animation = 13
		case 0x1e:
			animation = 9
		case 0x1f:
			animation = 10
		case 0x21, 0x23:
			animation = 14
		case 0x22:
			animation = 15
		}
	}
	if animation == 8 && s.IsMimic(obj) && uint32(ud.StatusFlags)&0x40000 != 0 {
		return 0
	}
	if animation == 8 && s.IsPlant(obj) && ud.CurrentEnemy == nil {
		return 14
	}
	if animation == 9 && s.IsZombie(obj) && uint32(ud.StatusFlags)&0x80000 != 0 {
		return 15
	}
	return animation
}

func (s *Server) complexObjectPacketNative518960(obj *server.Object) [11]byte {
	packet := s.phantomObjectPacketNative5187E0(obj)
	ud := obj.UpdateDataMonster()
	packet[9] |= s.monsterAnimationNative533790(obj) & 0x0f
	packet[10] = ud.Field120_1
	if uint32(obj.SubClass())&0x10 != 0 && ud.Field523_2 != 0 && s.Rand.Logic.IntClamp(0, 10) >= 8 {
		packet[9] = packet[9]&0xf0 | 14
		_, count := s.PlayerAnimFrames(50)
		if count > 0 {
			packet[10] = byte(s.Rand.Logic.IntClamp(0, count))
		}
	}
	return packet
}

func playerMapTracksObjectNative519410(pl *server.Player, obj *server.Object) bool {
	if pl == nil || obj == nil || pl.Field4580 == nil {
		return false
	}
	first := pl.Field4580
	for it := first; it != nil; it = it.Field8 {
		if it.Field4 == obj {
			return true
		}
		if it.Field8 == first {
			break
		}
	}
	return false
}

func playerTrackedObjectCountNative519710(pl *server.Player) int {
	if pl == nil || pl.Field4580 == nil {
		return 0
	}
	first := pl.Field4580
	count := 0
	for it := first; it != nil; it = it.Field8 {
		count++
		if it.Field8 == first {
			break
		}
	}
	return count
}

func netTrackedObjectRefreshDueNative519710(frame, last uint32, tracked int) bool {
	if tracked == 0 {
		return false
	}
	return tracked > 60 || frame-last > uint32(60/tracked)
}

func (s *Server) netTrackedObjectRefreshDueNative519710(update *server.PlayerUpdateData) bool {
	if update == nil {
		return false
	}
	return netTrackedObjectRefreshDueNative519710(
		s.Frame(),
		update.Field67,
		playerTrackedObjectCountNative519710(update.Player),
	)
}

func (s *Server) netFriendAddRemoveNative4D97A0(ind ntype.PlayerInd, obj *server.Object, add bool) {
	var packet [3]byte
	packet[0] = byte(netmsg.MSG_OBJECT_FRIEND_REMOVE)
	if add {
		packet[0] = byte(netmsg.MSG_OBJECT_FRIEND_ADD)
	}
	binary.LittleEndian.PutUint16(packet[1:], uint16(s.GetUnitNetCode(obj)))
	s.NetSendPacketXxx1(int(ind), packet[:], nil, 1)
}

// netSendObjects2PlayerNative519410 restores the architecture-independent
// visibility and packet-selection core of GAME.EXE 00519410. Type-specific
// immobile draw-frame updates remain with their map-loaded client drawables;
// their common special-state reports are still consumed below.
func (s *Server) netSendObjects2PlayerNative519410(recipient, obj *server.Object) bool {
	if recipient == nil || obj == nil || recipient.Flags().Has(object.FlagDestroyed) || obj.Class().Has(object.ClassClientPredict) {
		return false
	}
	pl := recipient.ControllingPlayer()
	if pl == nil {
		return false
	}
	ind := pl.PlayerIndex()
	if ind < 0 || ind >= 32 {
		return false
	}
	bit := uint32(1) << ind
	if obj != recipient && obj.Field35&bit != 0 {
		s.netFriendAddRemoveNative4D97A0(ind, obj, obj.Field36&bit != 0)
		obj.Field35 &^= bit
	}
	if obj.Field38&bit == 0 && !obj.Class().Has(object.ClassMissile) {
		return false
	}
	if obj.Flags().Has(object.FlagOwnerVisible) && !obj.HasOwner(recipient) && !noxflags.HasGame(noxflags.GameModeCoopTeam) {
		return false
	}
	visible := obj.Class().HasAny(object.ClassClientPersist|object.ClassImmobile) ||
		playerMapTracksObjectNative519410(pl, obj) ||
		s.MapTraceRayAt(pl.Pos3632Vec, obj.PosVec, nil, nil, 69)
	if !visible {
		if obj.Field37&bit != 0 {
			if obj.Class().HasAny(object.ClassMonster | object.ClassPlayer) {
				s.Nox_xxx_netObjectOutOfSight_528A60(int(ind), obj)
			} else {
				s.Nox_xxx_netObjectInShadows_528A90(int(ind), obj)
			}
			obj.Field37 &^= bit
			obj.Field38 |= bit
		}
		return false
	}
	if obj.Field37&bit != 0 && byte(obj.Field5)&0x20 != 0 {
		return false
	}
	if obj.Field37&bit == 0 && obj.Field140[ind]&0x0fff != 0 {
		obj.Field140[ind] |= (obj.Field140[ind] & 0x0fff) << 16
	}

	var sent bool
	switch {
	case obj.Class().Has(object.ClassImmobile):
		// The client loaded these drawables from the same map. Only their
		// portable special-state reports are required at this boundary.
	case obj.Class().Has(object.ClassComplex) && obj.Class().Has(object.ClassMonster):
		packet := s.complexObjectPacketNative518960(obj)
		sent = nox_netlist_addToMsgListSrv(ind, packet[:])
	case obj.Class().Has(object.ClassComplex) && obj.Class().Has(object.ClassPlayer):
		sent = s.netPlayerObjectSendNative518C30(recipient, obj, true)
	case obj.Class().Has(object.ClassComplex):
		packet := s.phantomObjectPacketNative5187E0(obj)
		sent = nox_netlist_addToMsgListSrv(ind, packet[:])
	case obj.Class().Has(object.ClassSimple):
		packet := s.simpleObjectPacketNative5188A0(obj)
		sent = nox_netlist_addToMsgListSrv(ind, packet[:])
	default:
		obj.Field38 = 0
	}
	s.netUpdateObjectSpecialNative527E50(recipient, obj)
	if sent {
		obj.Field37 |= bit
		obj.Field38 &^= bit
	}
	return sent
}
