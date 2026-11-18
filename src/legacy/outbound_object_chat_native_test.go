package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/outbound_object_chat_original_v1.json.gz
var outboundObjectChatOriginal []byte

type outboundObjectChatEvent struct {
	Call   string
	Args   []uint32
	Packet string
	Result uint32
}

type outboundObjectChatCase struct {
	Input struct {
		Name            string
		Root            string
		Unit            uint32
		Code            uint32
		Recipient       int
		SendResult      int32 `json:"send_result"`
		Source          uint32
		SourceClass     uint32 `json:"source_class"`
		X, Y            uint32
		Duration        uint16
		Command         string
		DirectRecipient uint8 `json:"direct_recipient"`
		Recipients      []uint8
		QueueResults    []int32 `json:"queue_results"`
	}
	Observed struct {
		Kind   string
		Result uint32
		Events []outboundObjectChatEvent
	}
}

func loadOutboundObjectChatCases(t *testing.T) []outboundObjectChatCase {
	t.Helper()
	if got := fmt.Sprintf("%x", sha256.Sum256(outboundObjectChatOriginal)); got != "b053f0eb2baf73a81c7b948aa0bedda84381659e4f4105f873dcb371a89db189" {
		t.Fatal("original packet fixture changed", got)
	}
	z, err := gzip.NewReader(bytes.NewReader(outboundObjectChatOriginal))
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	raw, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		OriginalSHA256 string `json:"original_sha256"`
		Rows           []outboundObjectChatCase
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.OriginalSHA256 != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(fixture.Rows) != 92 {
		t.Fatal("original GAME.EXE provenance or case count changed")
	}
	return fixture.Rows
}

type outboundObjectChatOwners struct {
	Unit, Source server.Object
	Peers        [3]server.Object
	Updates      [4]server.PlayerUpdateData
	Players      [4]server.Player
	Command      [512]byte
}

func testOutboundObjectChatNative(t *testing.T, tc outboundObjectChatCase, route, placement int) {
	t.Helper()
	const unitID, firstID, sourceID = 0x21101000, 0x21102000, 0x21105000
	bytesAt := int(unsafe.Sizeof(outboundObjectChatOwners{})) + placement + 64
	buf, free := alloc.MakeHigh([]byte{}, bytesAt)
	defer free()
	for i := range buf {
		buf[i] = byte(159 + i*13)
	}
	frame := (*outboundObjectChatOwners)(unsafe.Pointer(&buf[placement+32]))
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(frame)) <= math.MaxUint32 {
		t.Fatal("foreign fixture address did not exceed 32 bits")
	}
	*frame = outboundObjectChatOwners{}
	c := tc.Input
	frame.Unit.PosVec = types.Ptf(math.Float32frombits(c.X), math.Float32frombits(c.Y))
	frame.Source.ObjClass = object.Class(c.SourceClass)
	units := map[*server.Object]uint32{nil: 0, &frame.Unit: unitID, &frame.Source: sourceID}
	var peers []*server.Object
	for i, index := range c.Recipients {
		unit := &frame.Peers[i]
		unit.ObjClass = object.ClassPlayer
		unit.UpdateData = unsafe.Pointer(&frame.Updates[i])
		frame.Updates[i].Player = &frame.Players[i]
		frame.Players[i].PlayerInd = index
		units[unit] = firstID + uint32(i)*0x1000
		peers = append(peers, unit)
	}
	var source *server.Object
	if c.Source != 0 {
		source = &frame.Source
		source.UpdateData = unsafe.Pointer(&frame.Updates[3])
		frame.Updates[3].Player = &frame.Players[3]
		frame.Players[3].PlayerInd = c.DirectRecipient
	}
	command, err := hex.DecodeString(c.Command)
	if err != nil || len(command) >= len(frame.Command) {
		t.Fatal("bad command fixture", err)
	}
	copy(frame.Command[:], command)
	before := bytes.Clone(buf)
	pointerID := func(unit *server.Object) uint32 {
		id, ok := units[unit]
		if !ok {
			t.Fatalf("service saw truncated or unknown pointer %p", unit)
		}
		return id
	}
	events := make([]outboundObjectChatEvent, 0)
	nextPeer, queued := 0, 0
	old := outboundObjectChatRuntime
	defer func() { outboundObjectChatRuntime = old }()
	outboundObjectChatRuntime = func() server.OutboundObjectChatRuntime {
		return server.OutboundObjectChatRuntime{
			UnitCode: func(unit *server.Object) uint32 {
				events = append(events, outboundObjectChatEvent{Call: "code", Args: []uint32{pointerID(unit)}, Result: c.Code})
				return c.Code
			},
			Send: func(index int, packet []byte, related *server.Object, remove int) int {
				events = append(events, outboundObjectChatEvent{Call: "send", Args: []uint32{uint32(index), uint32(len(packet)), pointerID(related), uint32(remove)}, Packet: hex.EncodeToString(packet), Result: uint32(c.SendResult)})
				return int(c.SendResult)
			},
			FirstUnit: func() *server.Object {
				var unit *server.Object
				if len(peers) != 0 {
					unit = peers[0]
				}
				events = append(events, outboundObjectChatEvent{Call: "first", Args: []uint32{}, Result: pointerID(unit)})
				return unit
			},
			NextUnit: func(current *server.Object) *server.Object {
				if nextPeer >= len(peers) || peers[nextPeer] != current {
					t.Fatal("incorrect player traversal")
				}
				nextPeer++
				var unit *server.Object
				if nextPeer < len(peers) {
					unit = peers[nextPeer]
				}
				events = append(events, outboundObjectChatEvent{Call: "next", Args: []uint32{pointerID(current)}, Result: pointerID(unit)})
				return unit
			},
			Queue: func(index uint8, kind netlist.Kind, packet []byte) int32 {
				result := c.QueueResults[queued]
				queued++
				events = append(events, outboundObjectChatEvent{Call: "queue", Args: []uint32{uint32(index), uint32(kind), uint32(len(packet))}, Packet: hex.EncodeToString(packet), Result: uint32(result)})
				return result
			},
		}
	}
	entry := map[string]int{"out_of_sight": 0, "in_shadows": 1, "monster_command": 2, "clear_chat": 3}[c.Root]
	var unit *server.Object
	if c.Unit != 0 || entry == 2 {
		unit = &frame.Unit
	}
	actual := callOutboundObjectChatC(entry, route, c.Recipient, unit, source, &frame.Command[0], c.Duration)
	expected := uintptr(tc.Observed.Result)
	if tc.Observed.Kind == "pointer" {
		expected = uintptr(unsafe.Pointer(source))
	}
	if !reflect.DeepEqual(events, tc.Observed.Events) || actual != expected {
		t.Fatalf("got return %#x and calls %+v, want %#x and %+v", actual, events, expected, tc.Observed.Events)
	}
	if !bytes.Equal(buf, before) {
		t.Fatal("native owner or guard bytes changed")
	}
}

func TestOutboundObjectChatNativeOriginal(t *testing.T) {
	cases := loadOutboundObjectChatCases(t)
	for route := range 2 {
		for _, placement := range []int{0x100, 0x200, 0x300} {
			t.Run(fmt.Sprintf("route%d/place%x", route, placement), func(t *testing.T) {
				for _, tc := range cases {
					t.Run(tc.Input.Name, func(t *testing.T) {
						testOutboundObjectChatNative(t, tc, route, placement)
					})
				}
			})
		}
	}
}

func TestOutboundObjectChatNativeQueues(t *testing.T) {
	s := server.NewHigh(nil, nil, strman.New())
	defer s.Close()
	s.NetList.Init()
	defer s.NetList.Free()
	old := GetServer
	GetServer = func() Server { return &chatLegacyServer528AC0{srv: s} }
	defer func() { GetServer = old }()
	unit, freeUnit := alloc.NewHigh(server.Object{})
	defer freeUnit()
	unit.NetCode = 0x4321
	unit.PosVec = types.Ptf(-1.8, 2.8)
	for _, index := range []ntype.PlayerInd{2, 7, 31} {
		player := s.Players.ByIndRaw(index)
		player.Active = 1
		player.PlayerInd = byte(index)
		if index == 7 {
			continue // no unit in this active slot
		}
		object, freeObject := alloc.NewHigh(server.Object{})
		defer freeObject()
		update, freeUpdate := alloc.NewHigh(server.PlayerUpdateData{})
		defer freeUpdate()
		object.ObjClass = 4
		object.UpdateData = unsafe.Pointer(update)
		update.Player = player
		player.PlayerUnit = object
	}
	command, freeCommand := alloc.CloneSliceHigh([]byte{'i', 'd', 'l', 'e', 0})
	defer freeCommand()
	runtime.GC()
	if result := callOutboundObjectChatC(2, 1, 0, unit, nil, &command[0], 0x1234); result != 0 {
		t.Fatal("broadcast result", result)
	}
	if result := callOutboundObjectChatC(3, 1, 0, nil, nil, nil, 0); result != 0 {
		t.Fatal("clear chat result", result)
	}
	want := []byte{0xa8, 0x21, 0x43, 8, 0xff, 0xff, 2, 0, 5, 0x34, 0x12, 'i', 'd', 'l', 'e', 0, 0xca, 0xad, 0xde}
	for _, index := range []ntype.PlayerInd{2, 31} {
		if got := s.NetList.CopyPacketsA(index, netlist.Kind1); !bytes.Equal(got, want) {
			t.Fatalf("player %d queue = %x, want %x", index, got, want)
		}
	}
	if got := s.NetList.CopyPacketsA(7, netlist.Kind1); len(got) != 0 {
		t.Fatalf("unitless player queue = %x", got)
	}
}
