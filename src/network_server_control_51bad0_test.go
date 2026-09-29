package opennox

import (
	"encoding/binary"
	"reflect"
	"testing"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

type networkControlTrace51BAD0 struct {
	waypointUnit *server.Object
	waypoint     types.Pointf
	voteAction   string
	voteKind     int
	voteUnit     *server.Object
	voteName     []uint16
	gauntletUnit *server.Object
	gauntletData *server.PlayerUpdateData
	gauntlet     [server.NetworkGauntletPacketSize51BAD0]byte
}

func networkControlTestRuntime51BAD0(trace *networkControlTrace51BAD0, quest bool) networkControlRuntime51BAD0 {
	return networkControlRuntime51BAD0{
		questMode: quest,
		setWaypoint: func(unit *server.Object, point types.Pointf) {
			trace.waypointUnit = unit
			trace.waypoint = point
		},
		voteStart: func(kind int, unit *server.Object, name []uint16) {
			trace.voteAction = "start"
			trace.voteKind = kind
			trace.voteUnit = unit
			trace.voteName = append([]uint16(nil), name...)
		},
		voteCancel: func(kind int, unit *server.Object, name []uint16) {
			trace.voteAction = "cancel"
			trace.voteKind = kind
			trace.voteUnit = unit
			trace.voteName = append([]uint16(nil), name...)
		},
		gauntlet: func(packet *[server.NetworkGauntletPacketSize51BAD0]byte, unit *server.Object, update *server.PlayerUpdateData) int32 {
			trace.gauntlet = *packet
			trace.gauntletUnit = unit
			trace.gauntletData = update
			return int32(server.NetworkGauntletPacketSize51BAD0)
		},
	}
}

func TestDispatchNetworkControlNative51BAD0WaypointAndGauntlet(t *testing.T) {
	unit := &server.Object{}
	update := &server.PlayerUpdateData{}
	trace := &networkControlTrace51BAD0{}
	runtime := networkControlTestRuntime51BAD0(trace, false)

	waypoint := []byte{byte(netmsg.MSG_PLAYER_SET_WAYPOINT), 0xaa, 0xbb, 0x34, 0x12, 0xcd, 0xab}
	if n, handled, valid := dispatchNetworkControlNative51BAD0(netmsg.MSG_PLAYER_SET_WAYPOINT, waypoint, unit, update, runtime); n != len(waypoint) || !handled || !valid {
		t.Fatalf("waypoint dispatch = (%d,%t,%t), want (%d,true,true)", n, handled, valid, len(waypoint))
	}
	if trace.waypointUnit != unit || trace.waypoint != (types.Pointf{X: 0x1234, Y: 0xabcd}) {
		t.Fatalf("waypoint call = (%p,%v), want (%p,{4660 43981})", trace.waypointUnit, trace.waypoint, unit)
	}

	gauntlet := []byte{byte(netmsg.MSG_GAUNTLET), 27, 0xfe}
	if n, handled, valid := dispatchNetworkControlNative51BAD0(netmsg.MSG_GAUNTLET, gauntlet, unit, update, runtime); n != server.NetworkGauntletPacketSize51BAD0 || !handled || !valid {
		t.Fatalf("gauntlet dispatch = (%d,%t,%t), want (%d,true,true)", n, handled, valid, server.NetworkGauntletPacketSize51BAD0)
	}
	if trace.gauntlet != ([2]byte{byte(netmsg.MSG_GAUNTLET), 27}) || trace.gauntletUnit != unit || trace.gauntletData != update {
		t.Fatalf("gauntlet call = (%v,%p,%p), want ([240 27],%p,%p)", trace.gauntlet, trace.gauntletUnit, trace.gauntletData, unit, update)
	}
}

func TestDispatchNetworkControlNative51BAD0Votes(t *testing.T) {
	unit := &server.Object{}
	packet := make([]byte, networkVoteNamedPacketSize51BAD0)
	packet[0] = byte(netmsg.MSG_VOTE)
	wantName := make([]uint16, networkVoteNameUnits51BAD0+1)
	for i := 0; i < networkVoteNameUnits51BAD0; i++ {
		wantName[i] = uint16(0x4100 + i)
		binary.LittleEndian.PutUint16(packet[2+i*2:], wantName[i])
	}

	for _, tc := range []struct {
		name    string
		subtype byte
		quest   bool
		action  string
		kind    int
	}{
		{"start kick", 0, false, "start", 0},
		{"start quest kick", 0, true, "start", 3},
		{"start ban", 1, false, "start", 1},
		{"cancel kick", 2, false, "cancel", 0},
		{"cancel quest kick", 2, true, "cancel", 3},
		{"cancel ban", 3, false, "cancel", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			trace := &networkControlTrace51BAD0{}
			packet[1] = tc.subtype
			n, handled, valid := dispatchNetworkControlNative51BAD0(netmsg.MSG_VOTE, packet, unit, nil, networkControlTestRuntime51BAD0(trace, tc.quest))
			if n != len(packet) || !handled || !valid {
				t.Fatalf("dispatch = (%d,%t,%t), want (%d,true,true)", n, handled, valid, len(packet))
			}
			if trace.voteAction != tc.action || trace.voteKind != tc.kind || trace.voteUnit != unit || !reflect.DeepEqual(trace.voteName, wantName) {
				t.Fatalf("vote call = (%q,%d,%p,%#v), want (%q,%d,%p,%#v)", trace.voteAction, trace.voteKind, trace.voteUnit, trace.voteName, tc.action, tc.kind, unit, wantName)
			}
		})
	}

	for _, tc := range []struct {
		subtype byte
		action  string
	}{
		{4, "start"},
		{5, "cancel"},
	} {
		trace := &networkControlTrace51BAD0{}
		control := []byte{byte(netmsg.MSG_VOTE), tc.subtype}
		n, handled, valid := dispatchNetworkControlNative51BAD0(netmsg.MSG_VOTE, control, unit, nil, networkControlTestRuntime51BAD0(trace, false))
		if n != len(control) || !handled || !valid || trace.voteAction != tc.action || trace.voteKind != 2 || trace.voteName != nil {
			t.Fatalf("subtype %d dispatch/call = (%d,%t,%t,%q,%d,%#v)", tc.subtype, n, handled, valid, trace.voteAction, trace.voteKind, trace.voteName)
		}
	}
}

func TestDispatchNetworkControlNative51BAD0RejectsMalformedWithoutFallback(t *testing.T) {
	unit := &server.Object{}
	update := &server.PlayerUpdateData{}
	for _, tc := range []struct {
		name string
		op   netmsg.Op
		data []byte
	}{
		{"empty waypoint", netmsg.MSG_PLAYER_SET_WAYPOINT, nil},
		{"short waypoint", netmsg.MSG_PLAYER_SET_WAYPOINT, make([]byte, networkPlayerSetWaypointPacketSize51BAD0-1)},
		{"empty vote", netmsg.MSG_VOTE, nil},
		{"short named vote", netmsg.MSG_VOTE, []byte{byte(netmsg.MSG_VOTE), 0}},
		{"invalid vote", netmsg.MSG_VOTE, []byte{byte(netmsg.MSG_VOTE), 6}},
		{"empty gauntlet", netmsg.MSG_GAUNTLET, nil},
		{"short gauntlet", netmsg.MSG_GAUNTLET, []byte{byte(netmsg.MSG_GAUNTLET)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			trace := &networkControlTrace51BAD0{}
			n, handled, valid := dispatchNetworkControlNative51BAD0(tc.op, tc.data, unit, update, networkControlTestRuntime51BAD0(trace, false))
			if n != 0 || !handled || valid {
				t.Fatalf("dispatch = (%d,%t,%t), want (0,true,false)", n, handled, valid)
			}
			if !reflect.DeepEqual(trace, &networkControlTrace51BAD0{}) {
				t.Fatalf("malformed packet invoked runtime: %#v", trace)
			}
		})
	}

	trace := &networkControlTrace51BAD0{}
	runtime := networkControlTestRuntime51BAD0(trace, false)
	runtime.gauntlet = func(*[server.NetworkGauntletPacketSize51BAD0]byte, *server.Object, *server.PlayerUpdateData) int32 {
		return -1
	}
	if n, handled, valid := dispatchNetworkControlNative51BAD0(netmsg.MSG_GAUNTLET, []byte{byte(netmsg.MSG_GAUNTLET), 0xff}, unit, update, runtime); n != 0 || !handled || valid {
		t.Fatalf("invalid gauntlet dispatch = (%d,%t,%t), want (0,true,false)", n, handled, valid)
	}

	if n, handled, valid := dispatchNetworkControlNative51BAD0(netmsg.MSG_KEEP_ALIVE, []byte{byte(netmsg.MSG_KEEP_ALIVE)}, unit, update, networkControlTestRuntime51BAD0(trace, false)); n != 0 || handled || valid {
		t.Fatalf("unrelated dispatch = (%d,%t,%t), want (0,false,false)", n, handled, valid)
	}
}
