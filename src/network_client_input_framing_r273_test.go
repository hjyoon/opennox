package opennox

import (
	"bytes"
	"image"
	"math"
	"reflect"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet"
	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/player"
	"github.com/opennox/libs/strman"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Keep the reference wire table independent of CtrlCode.DataSize and of the
// producer/receiver implementations. Unknown low bytes retain zero payload.
func inputPayloadWidthR273(code player.CtrlCode) int {
	switch byte(code) {
	case 1, 2, 3, 4, 5, 28:
		return 1
	case 29, 30:
		return 4
	default:
		return 0
	}
}

func inputReferenceFramesR273(events []noxCtrlEvent) []byte {
	var out, body []byte
	flush := func() {
		if len(body) != 0 {
			out = append(out, 0x3f, byte(len(body)))
			out = append(out, body...)
			body = body[:0]
		}
	}
	for _, event := range events {
		if !event.Active {
			continue
		}
		n := inputPayloadWidthR273(event.Code)
		if len(body)+4+n > 255 {
			flush()
		}
		body = append(body, byte(event.Code), 0, 0, 0)
		for i := 0; i < n; i++ {
			body = append(body, byte(event.Data>>uint(8*i)))
		}
	}
	flush()
	return out
}

func inputFrameEventsR273(t *testing.T, wire []byte) [][]server.PlayerCtrl {
	t.Helper()
	var frames [][]server.PlayerCtrl
	for len(wire) != 0 {
		if len(wire) < 2 || wire[0] != 0x3f || int(wire[1]) > len(wire)-2 || wire[1] == 0 {
			t.Fatalf("invalid produced input frame: %x", wire)
		}
		n := int(wire[1])
		body := wire[2 : 2+n]
		wire = wire[2+n:]
		var events []server.PlayerCtrl
		for len(body) != 0 {
			if len(body) < 4 || body[1] != 0 || body[2] != 0 || body[3] != 0 {
				t.Fatalf("record was split or its code DWORD changed: %x", body)
			}
			code := player.CtrlCode(body[0])
			width := inputPayloadWidthR273(code)
			if len(body) < 4+width {
				t.Fatalf("payload split at frame boundary: %x", body)
			}
			var value uint32
			for i := 0; i < width; i++ {
				value |= uint32(body[4+i]) << uint(8*i)
			}
			events = append(events, server.PlayerCtrl{Code: code, Data: value, Active: true})
			body = body[4+width:]
		}
		frames = append(frames, events)
	}
	return frames
}

func inputHostModeR273(t *testing.T, enabled bool) {
	t.Helper()
	old := noxflags.HasGame(noxflags.GameHost)
	set := func(value bool) {
		if value {
			noxflags.SetGame(noxflags.GameHost)
		} else {
			noxflags.UnsetGame(noxflags.GameHost)
		}
	}
	set(enabled)
	t.Cleanup(func() { set(old) })
}

func inputClientFixtureR273(t *testing.T) *Client {
	t.Helper()
	list := netlist.New()
	list.Init()
	t.Cleanup(list.Free)
	return &Client{srv: &Server{Server: &server.Server{NetList: list}}, ctrl: new(CtrlEventHandler)}
}

func inputRepeatEventsR273(count int, code player.CtrlCode) []noxCtrlEvent {
	events := make([]noxCtrlEvent, count)
	for i := range events {
		events[i] = noxCtrlEvent{tick: uint32(i), PlayerCtrl: server.PlayerCtrl{
			Code: code, Data: 0x89abcd00 + uint32(i), Active: true,
		}}
	}
	return events
}

func inputProduceAndQueueR273(t *testing.T, c *Client, events []noxCtrlEvent) []byte {
	t.Helper()
	*c.ctrl = CtrlEventHandler{}
	copy(c.ctrl.bufA[:], events)
	c.ctrl.indA = len(events)
	before := *c.ctrl
	before.netBuf = nil
	c.ctrl.writeToNetBuffer()
	after := *c.ctrl
	after.netBuf = nil
	if !reflect.DeepEqual(before, after) {
		t.Fatal("non-host producer changed control-ring state")
	}
	want := inputReferenceFramesR273(events)
	if !bytes.Equal(c.ctrl.netBuf, want) {
		t.Fatalf("producer framing = %x, want %x", c.ctrl.netBuf, want)
	}
	prepared := bytes.Clone(c.ctrl.netBuf)
	if !c.clientSendInput(server.HostPlayerIndex) {
		t.Fatal("complete prepared input did not fit an empty message list")
	}
	list := c.srv.NetList.ByInd(server.HostPlayerIndex, netlist.Kind0)
	wantCount := 0
	if len(want) != 0 {
		wantCount = 1
	}
	if list.Count() != wantCount || list.Size() != len(want) || !bytes.Equal(c.ctrl.netBuf, prepared) {
		t.Fatalf("sender must queue one complete copied message: count=%d/%d size=%d/%d", list.Count(), wantCount, list.Size(), len(want))
	}
	return want
}

func TestClientInputFramingMatrixR273(t *testing.T) {
	inputHostModeR273(t, false)
	c := inputClientFixtureR273(t)
	for code := 0; code < 256; code++ {
		for count := 0; count <= 128; count++ {
			events := inputRepeatEventsR273(count, player.CtrlCode(code))
			want := inputProduceAndQueueR273(t, c, events)
			got := c.srv.NetList.CopyPacketsA(server.HostPlayerIndex, netlist.Kind0)
			if !bytes.Equal(got, want) {
				t.Fatalf("code=%d count=%d: queued wire=%x want=%x", code, count, got, want)
			}
			inputFrameEventsR273(t, got)
		}
	}
}

func TestClientInputFramingBoundariesAndFilteringR273(t *testing.T) {
	inputHostModeR273(t, false)
	c := inputClientFixtureR273(t)
	for _, spec := range []struct {
		count int
		code  player.CtrlCode
		body  int
		parts []int
	}{{50, 1, 250, []int{250}}, {51, 1, 255, []int{255}}, {32, 29, 256, []int{248, 8}}, {64, 6, 256, []int{252, 4}}} {
		want := inputProduceAndQueueR273(t, c, inputRepeatEventsR273(spec.count, spec.code))
		got := c.srv.NetList.CopyPacketsA(server.HostPlayerIndex, netlist.Kind0)
		var sizes []int
		for remaining := got; len(remaining) != 0; remaining = remaining[2+int(remaining[1]):] {
			sizes = append(sizes, int(remaining[1]))
		}
		if !bytes.Equal(got, want) || !slices.Equal(sizes, spec.parts) || len(got) != spec.body+2*len(spec.parts) {
			t.Fatalf("body=%d count=%d code=%d: parts=%v/%v bytes=%d", spec.body, spec.count, spec.code, sizes, spec.parts, len(got))
		}
	}
	for _, allInactive := range []bool{false, true} {
		events := inputRepeatEventsR273(128, 6)
		codes := []player.CtrlCode{0, 1, 29, 2, 30, 4, 6, 255}
		for i := range events {
			events[i].Code = codes[i%len(codes)]
			events[i].Active = !allInactive && i%3 != 1
		}
		want := inputProduceAndQueueR273(t, c, events)
		got := c.srv.NetList.CopyPacketsA(server.HostPlayerIndex, netlist.Kind0)
		if !bytes.Equal(got, want) || allInactive && len(got) != 0 {
			t.Fatalf("inactive/mixed filtering all_inactive=%t wire=%x want=%x", allInactive, got, want)
		}
		inputFrameEventsR273(t, got)
	}
}

func TestClientInputFramingHostRingAndTickCutoffR273(t *testing.T) {
	inputHostModeR273(t, true)
	c := inputClientFixtureR273(t)
	ctrl := c.ctrl
	ctrl.indB, ctrl.indA = 125, 3
	var selected []noxCtrlEvent
	for i, slot := range []int{125, 126, 127, 0, 1, 2} {
		event := inputRepeatEventsR273(1, []player.CtrlCode{1, 29, 6}[i%3])[0]
		event.Data = 0x12345600 + uint32(i)
		event.tick = 0
		event.Active = i != 1
		ctrl.bufA[slot] = event
		selected = append(selected, event)
	}
	before := *ctrl
	before.netBuf = nil
	before.indB = before.indA
	ctrl.writeToNetBuffer()
	after := *ctrl
	after.netBuf = nil
	if !reflect.DeepEqual(before, after) || !bytes.Equal(ctrl.netBuf, inputReferenceFramesR273(selected)) || !c.clientSendInput(server.HostPlayerIndex) {
		t.Fatal("host ring traversal, inactive filtering, or indB=indA changed")
	}
	if got := c.srv.NetList.CopyPacketsA(server.HostPlayerIndex, netlist.Kind0); !bytes.Equal(got, inputReferenceFramesR273(selected)) {
		t.Fatalf("wrapped host wire=%x", got)
	}

	// Use the real platform clock. The final tick value remains beyond its
	// 50-tick cutoff; no clock replacement or wait is needed by this fixture.
	ctrl.indB, ctrl.indA = 126, 2
	ctrl.bufA[126] = inputRepeatEventsR273(1, 29)[0]
	ctrl.bufA[126].tick = 0
	ctrl.bufA[127] = inputRepeatEventsR273(1, 1)[0]
	ctrl.bufA[127].tick = math.MaxUint32
	ctrl.bufA[127].Active = false // Tick gate precedes the Active gate.
	ctrl.bufA[0] = inputRepeatEventsR273(1, 6)[0]
	ctrl.bufA[0].tick = 0
	ctrl.bufA[1] = ctrl.bufA[0]
	selected = []noxCtrlEvent{ctrl.bufA[126]}
	before = *ctrl
	before.netBuf = nil
	before.indB = before.indA
	ctrl.writeToNetBuffer()
	after = *ctrl
	after.netBuf = nil
	if !reflect.DeepEqual(before, after) || !bytes.Equal(ctrl.netBuf, inputReferenceFramesR273(selected)) {
		t.Fatal("host future-event cutoff or deliberate indB=indA behavior changed")
	}
	ctrl.writeToNetBuffer()
	if len(ctrl.netBuf) != 0 || !c.clientSendInput(server.HostPlayerIndex) || c.srv.NetList.ByInd(server.HostPlayerIndex, netlist.Kind0).Count() != 0 {
		t.Fatal("an exhausted/equal-index host ring queued an empty header or replayed skipped events")
	}

	ctrl.indB, ctrl.indA = 5, 8
	ctrl.bufA[5] = inputRepeatEventsR273(1, 29)[0]
	ctrl.bufA[5].tick = math.MaxUint32
	ctrl.bufA[6] = inputRepeatEventsR273(1, 6)[0]
	ctrl.bufA[6].tick = 0
	ctrl.bufA[7] = ctrl.bufA[6]
	before = *ctrl
	before.netBuf = nil
	before.indB = before.indA
	ctrl.writeToNetBuffer()
	after = *ctrl
	after.netBuf = nil
	if !reflect.DeepEqual(before, after) || len(ctrl.netBuf) != 0 ||
		!c.clientSendInput(server.HostPlayerIndex) || c.srv.NetList.ByInd(server.HostPlayerIndex, netlist.Kind0).Count() != 0 {
		t.Fatal("future FIRST host event must leave no header while still setting indB=indA")
	}
}

// The receiver's 128-entry storage policy is intentionally NOT changed by
// framing: Append rejects a complete frame when stored+frameCount >= 128.
// Dedup marks older movement events inactive but does not free storage slots.
func inputAdmissionReferenceR273(seed []server.PlayerCtrl, frames [][]server.PlayerCtrl, ready bool) []server.PlayerCtrl {
	stored := slices.Clone(seed)
	if ready {
		for _, frame := range frames {
			if len(stored)+len(frame) >= 128 {
				continue
			}
			stored = append(stored, frame...)
			seen := make(map[player.CtrlCode]bool)
			for i := len(stored) - 1; i >= 0; i-- {
				event := &stored[i]
				if event.Active && (event.Code == 2 || event.Code == 4 || event.Code == 5) {
					if seen[event.Code] {
						event.Active = false
					} else {
						seen[event.Code] = true
					}
				}
			}
		}
	}
	var active []server.PlayerCtrl
	for _, event := range stored {
		if event.Active {
			active = append(active, event)
		}
	}
	return active
}

func TestClientInputFramingActualServerFIFOAndAdmissionR273(t *testing.T) {
	inputHostModeR273(t, false)
	base := server.New(nil, nil, strman.New())
	t.Cleanup(base.Close)
	s := &Server{Server: base}
	s.NetList.Init()
	t.Cleanup(s.NetList.Free)
	pl := s.Players.ResetInd(server.HostPlayerIndex)
	unit, freeUnit := alloc.New(server.Object{})
	t.Cleanup(freeUnit)
	update, freeUpdate := alloc.New(server.PlayerUpdateData{})
	t.Cleanup(freeUpdate)
	*unit = server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	*update = server.PlayerUpdateData{Player: pl}
	pl.PlayerUnit = unit
	t.Cleanup(func() { pl.PlayerUnit = nil; unit.UpdateData = nil; update.Player = nil })
	if unit.UpdateDataPlayer() != update || update.Player != pl || pl.PlayerUnit != unit ||
		unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 || uintptr(unsafe.Pointer(update)) <= math.MaxUint32 || uintptr(unsafe.Pointer(pl)) <= math.MaxUint32) {
		t.Fatal("actual raw receiver requires initialized native-width player/unit/update links")
	}
	replay := noxflags.HasEngine(noxflags.EngineReplayWrite)
	noxflags.UnsetEngine(noxflags.EngineReplayWrite)
	t.Cleanup(func() {
		if replay {
			noxflags.SetEngine(noxflags.EngineReplayWrite)
		}
	})
	oldSprite := legacy.Get_dword_5d4594_1096640()
	legacy.Set_dword_5d4594_1096640(nil)
	t.Cleanup(func() { legacy.Set_dword_5d4594_1096640(oldSprite) })
	keepalive, err := netmsg.Append(nil, &noxnet.MsgKeepAlive{})
	if err != nil {
		t.Fatal(err)
	}
	c := &Client{srv: s, ctrl: new(CtrlEventHandler)}
	queue := s.Players.Control.Player(server.HostPlayerIndex)
	for _, spec := range []struct {
		count, seed int
		code        player.CtrlCode
		ready       bool
	}{{0, 0, 6, true}, {128, 0, 6, true}, {128, 0, 1, true}, {128, 0, 29, true},
		{128, 1, 6, true}, {128, 100, 29, true}, {1, 127, 6, true}, {16, 126, 6, true},
		{128, 0, 2, true}, {128, 1, 4, true}, {128, 0, 5, true}, {128, 3, 29, false},
		{127, 0, 6, true}, {127, 0, 1, true}, {127, 0, 29, true}, {128, 0, 0xff, true}} {
		queue.Reset()
		var seed []server.PlayerCtrl
		for i := 0; i < spec.seed; i++ {
			seed = append(seed, server.PlayerCtrl{Code: player.CCJump, Data: uint32(i), Active: true})
		}
		queue.Append(seed)
		pl.Field3680 = 0
		if spec.ready {
			pl.Field3680 = 0x10
		}
		pl.SetCursorPos(image.Pt(1, 2))
		pl.Frame3596 = 1
		s.SetFrame(777)
		events := inputRepeatEventsR273(spec.count, spec.code)
		if spec.code == 0xff {
			// The tag selects a fixture whose actual records interleave
			// repeated forward/left/right inputs with all payload widths.
			// Movement duplicates deliberately span frame boundaries.
			codes := []player.CtrlCode{2, 29, 4, 6, 5, 30, 2, 4, 5, 1}
			for i := range events {
				events[i].Code = codes[i%len(codes)]
			}
		}
		wire := inputProduceAndQueueR273(t, c, events)
		frames := inputFrameEventsR273(t, wire)
		want := inputAdmissionReferenceR273(seed, frames, spec.ready)
		c.netPrevMouse = image.Point{}
		destination := image.Pt(0x1234, 0xabcd)
		if !c.clientSendInputMouse(server.HostPlayerIndex, destination) || !s.NetList.AddToMsgListCli(server.HostPlayerIndex, netlist.Kind0, keepalive) {
			t.Fatal("ordinary following Mouse/KeepAlive could not be queued")
		}
		suffix := []byte{byte(netmsg.MSG_MOUSE), 0x34, 0x12, 0xcd, 0xab, byte(netmsg.MSG_KEEP_ALIVE)}
		wantWire := append(bytes.Clone(wire), suffix...)
		var drained []byte
		s.NetList.HandlePacketsA(server.HostPlayerIndex, netlist.Kind0, func(data []byte) {
			drained = bytes.Clone(data)
			if !s.onPacketRaw(server.HostPlayerIndex, data) {
				t.Fatalf("actual server rejected producer wire: %x", data)
			}
		})
		var got []server.PlayerCtrl
		for event := queue.First(); event != nil; event = queue.Next() {
			got = append(got, *event)
		}
		if !bytes.Equal(drained, wantWire) || !slices.Equal(got, want) || pl.CursorVec != destination || pl.Frame3596 != s.Frame() || s.NetList.ByInd(server.HostPlayerIndex, netlist.Kind0).Count() != 0 {
			t.Fatalf("count=%d seed=%d code=%d ready=%t: active=%v want=%v cursor=%v frame=%d FIFO=%t", spec.count, spec.seed, spec.code, spec.ready, got, want, pl.CursorVec, pl.Frame3596, bytes.Equal(drained, wantWire))
		}
	}
}

func TestClientInputFramingQueueCapacityAtomicR273(t *testing.T) {
	inputHostModeR273(t, false)
	c := inputClientFixtureR273(t)
	*c.ctrl = CtrlEventHandler{indA: 1}
	c.ctrl.bufA[0] = inputRepeatEventsR273(1, 6)[0]
	c.ctrl.writeToNetBuffer()
	prepared := bytes.Clone(c.ctrl.netBuf)
	if !bytes.Equal(prepared, inputReferenceFramesR273(c.ctrl.bufA[:1])) {
		t.Fatal("capacity fixture did not use the actual producer")
	}
	for _, spec := range []struct {
		nodes, bytes int
		accept       bool
	}{{511, 511, true}, {512, 512, false}, {1, 2048 - len(prepared), true},
		{1, 2049 - len(prepared), false}, {1, 2047, false}, {1, 2048, false}} {
		c.srv.NetList.ResetByInd(server.HostPlayerIndex, netlist.Kind0)
		var existing []byte
		for i := 0; i < spec.nodes; i++ {
			n := 1
			if spec.nodes == 1 {
				n = spec.bytes
			}
			// Each filler byte is a complete ordinary KEEP_ALIVE message.
			message := bytes.Repeat([]byte{byte(netmsg.MSG_KEEP_ALIVE)}, n)
			if !c.srv.NetList.AddToMsgListCli(server.HostPlayerIndex, netlist.Kind0, message) {
				t.Fatalf("cannot prepare legitimate queue node %d/%d", i, spec.nodes)
			}
			existing = append(existing, message...)
		}
		before := *c.ctrl
		if got := c.clientSendInput(server.HostPlayerIndex); got != spec.accept {
			t.Fatalf("nodes=%d bytes=%d accepted=%t, want %t", spec.nodes, spec.bytes, got, spec.accept)
		}
		if !reflect.DeepEqual(before, *c.ctrl) || !bytes.Equal(c.ctrl.netBuf, prepared) ||
			cap(before.netBuf) != cap(c.ctrl.netBuf) || unsafe.SliceData(before.netBuf) != unsafe.SliceData(c.ctrl.netBuf) {
			t.Fatal("send success/failure changed the prepared producer state")
		}
		want := bytes.Clone(existing)
		wantNodes := spec.nodes
		if spec.accept {
			want = append(want, prepared...)
			wantNodes++
		}
		list := c.srv.NetList.ByInd(server.HostPlayerIndex, netlist.Kind0)
		if list.Count() != wantNodes || list.Size() != len(want) {
			t.Fatalf("nodes=%d bytes=%d left orphan header/partial data: queued=%d/%d want=%d/%d", spec.nodes, spec.bytes, list.Count(), list.Size(), wantNodes, len(want))
		}
		if got := c.srv.NetList.CopyPacketsA(server.HostPlayerIndex, netlist.Kind0); !bytes.Equal(got, want) {
			t.Fatalf("capacity failure modified existing queued bytes: got=%x want=%x", got, want)
		}
	}

	// Removing nodes directly does not rewind NetList's byte arena. Exercise
	// the allocation failure after its count/size checks have both succeeded.
	c.srv.NetList.ResetByInd(server.HostPlayerIndex, netlist.Kind0)
	filler := bytes.Repeat([]byte{byte(netmsg.MSG_KEEP_ALIVE)}, 2048)
	if !c.srv.NetList.AddToMsgListCli(server.HostPlayerIndex, netlist.Kind0, filler) {
		t.Fatal("cannot fill the complete queue byte arena")
	}
	list := c.srv.NetList.ByInd(server.HostPlayerIndex, netlist.Kind0)
	if got := list.Get(); !bytes.Equal(got, filler) || list.Count() != 0 || list.Size() != 0 {
		t.Fatal("direct node drain did not preserve the full arena fixture")
	}
	before := *c.ctrl
	if c.clientSendInput(server.HostPlayerIndex) || list.Count() != 0 || list.Size() != 0 || len(list.Get()) != 0 ||
		!reflect.DeepEqual(before, *c.ctrl) || !bytes.Equal(c.ctrl.netBuf, prepared) ||
		cap(before.netBuf) != cap(c.ctrl.netBuf) || unsafe.SliceData(before.netBuf) != unsafe.SliceData(c.ctrl.netBuf) {
		t.Fatal("exhausted byte arena queued an orphan header or modified producer state")
	}
	c.srv.NetList.ResetByInd(server.HostPlayerIndex, netlist.Kind0)
	if !c.clientSendInput(server.HostPlayerIndex) || list.Count() != 1 || list.Size() != len(prepared) ||
		!reflect.DeepEqual(before, *c.ctrl) || !bytes.Equal(c.ctrl.netBuf, prepared) {
		t.Fatal("ResetByInd did not recover the arena for one complete input message")
	}
	if got := c.srv.NetList.CopyPacketsA(server.HostPlayerIndex, netlist.Kind0); !bytes.Equal(got, prepared) {
		t.Fatal("arena recovery changed the prepared complete input bytes")
	}
}

func TestClientInputFramingCopiedOwnershipR273(t *testing.T) {
	inputHostModeR273(t, false)
	c := inputClientFixtureR273(t)
	events := inputRepeatEventsR273(128, 29)
	want := inputProduceAndQueueR273(t, c, events)
	// Regenerate into the producer's reusable backing allocation. The queued
	// message must retain the previous complete bytes, not borrow that slice.
	for i := range c.ctrl.bufA {
		c.ctrl.bufA[i].Data ^= 0xffffffff
	}
	c.ctrl.writeToNetBuffer()
	if bytes.Equal(c.ctrl.netBuf, want) {
		t.Fatal("copy-ownership fixture did not actually change producer bytes")
	}
	if got := c.srv.NetList.CopyPacketsA(server.HostPlayerIndex, netlist.Kind0); !bytes.Equal(got, want) {
		t.Fatal("queued complete input borrowed the producer's reusable netBuf")
	}
}

// Deliberately reach the sender before asserting producer framing. This
// independently exposes the baseline sender's orphan-header failure even
// when its old producer still returns an unframed control body.
func TestClientInputFramingQueueFailureAtomicR273(t *testing.T) {
	inputHostModeR273(t, false)
	c := inputClientFixtureR273(t)
	events := inputRepeatEventsR273(1, 6)
	copy(c.ctrl.bufA[:], events)
	c.ctrl.indA = len(events)
	c.ctrl.writeToNetBuffer()
	wantMessage := inputReferenceFramesR273(events)
	prepared := bytes.Clone(c.ctrl.netBuf)
	for _, spec := range []struct {
		nodes, bytes int
		accept       bool
	}{{511, 511, true}, {512, 512, false}, {1, 2044, false}, {1, 2048 - len(wantMessage), true}} {
		c.srv.NetList.ResetByInd(server.HostPlayerIndex, netlist.Kind0)
		var existing []byte
		for i := 0; i < spec.nodes; i++ {
			n := 1
			if spec.nodes == 1 {
				n = spec.bytes
			}
			message := bytes.Repeat([]byte{byte(netmsg.MSG_KEEP_ALIVE)}, n)
			if !c.srv.NetList.AddToMsgListCli(server.HostPlayerIndex, netlist.Kind0, message) {
				t.Fatalf("cannot prepare sender-only queue node %d/%d", i, spec.nodes)
			}
			existing = append(existing, message...)
		}
		before := *c.ctrl
		accepted := c.clientSendInput(server.HostPlayerIndex)
		list := c.srv.NetList.ByInd(server.HostPlayerIndex, netlist.Kind0)
		gotCount, gotSize := list.Count(), list.Size()
		got := c.srv.NetList.CopyPacketsA(server.HostPlayerIndex, netlist.Kind0)
		want := bytes.Clone(existing)
		wantCount := spec.nodes
		if spec.accept {
			want = append(want, wantMessage...)
			wantCount++
		}
		if accepted != spec.accept || gotCount != wantCount || gotSize != len(want) || !bytes.Equal(got, want) ||
			!reflect.DeepEqual(before, *c.ctrl) || !bytes.Equal(c.ctrl.netBuf, prepared) ||
			cap(before.netBuf) != cap(c.ctrl.netBuf) || unsafe.SliceData(before.netBuf) != unsafe.SliceData(c.ctrl.netBuf) {
			// Report the actual partial tail as evidence instead of failing
			// solely on the sender's return before inspecting its queue.
			queuedTail := got[min(len(existing), len(got)):]
			t.Errorf("sender nodes=%d bytes=%d accepted=%t/%t queued_count=%d/%d queued_size=%d/%d queued_tail=%x complete_expected=%x", spec.nodes, spec.bytes, accepted, spec.accept, gotCount, wantCount, gotSize, len(want), queuedTail, wantMessage)
		}
	}
}
