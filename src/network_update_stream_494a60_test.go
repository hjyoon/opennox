package opennox

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/client"
)

func highAddressUpdateStreamDrawable494A60(t *testing.T) *client.Drawable {
	t.Helper()
	dr := new(client.Drawable)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= uintptr(math.MaxUint32) {
		t.Skipf("allocator returned a low address: %p", dr)
	}
	return dr
}

func updateStreamTestHooks494A60(
	entries *[updateStreamAliasCount494A60]updateStreamAliasEntry494A60,
	frame uint32,
) updateStreamHooks494A60 {
	return updateStreamHooks494A60{
		frame:       func() uint32 { return frame },
		lookupAlias: func(ind byte) updateStreamAliasEntry494A60 { return entries[ind] },
		reserveAlias: func(code, typeID uint16, now, deadline uint32) (byte, bool) {
			return reserveUpdateStreamAlias494A60(
				code, typeID, now, deadline,
				func(ind byte) updateStreamAliasEntry494A60 { return entries[ind] },
				func(ind byte, entry updateStreamAliasEntry494A60) { entries[ind] = entry },
			)
		},
		sendAlias:    func(byte, updateStreamAliasEntry494A60) {},
		create:       func(int, uint16, int, int) *client.Drawable { return nil },
		updateCamera: func(int, int) {},
		addToIndex2D: func(*client.Drawable) {},
	}
}

func TestHandleUpdateStreamNative494A60PreservesHighAddressPlayer(t *testing.T) {
	const (
		code   = uint16(0x1234)
		typeID = uint16(0x5678)
		frame  = uint32(77)
	)
	entries := new([updateStreamAliasCount494A60]updateStreamAliasEntry494A60)
	hooks := updateStreamTestHooks494A60(entries, frame)
	dr := highAddressUpdateStreamDrawable494A60(t)
	dr.ObjClass = object.ClassComplex | object.ClassPlayer
	var creates int
	hooks.create = func(gotType int, gotCode uint16, x, y int) *client.Drawable {
		creates++
		if gotType != int(typeID) || gotCode != code {
			t.Fatalf("create %d type/code = %#x/%#x, want %#x/%#x", creates, gotType, gotCode, typeID, code)
		}
		wantX, wantY := 1000, 1200
		if creates == 2 {
			wantX, wantY = 1005, 1197
		}
		if x != wantX || y != wantY {
			t.Fatalf("create %d position = (%d,%d), want (%d,%d)", creates, x, y, wantX, wantY)
		}
		dr.PosVec.X, dr.PosVec.Y = x, y
		return dr
	}
	var cameraX, cameraY int
	hooks.updateCamera = func(x, y int) { cameraX, cameraY = x, y }
	var sent []struct {
		alias byte
		entry updateStreamAliasEntry494A60
	}
	hooks.sendAlias = func(alias byte, entry updateStreamAliasEntry494A60) {
		sent = append(sent, struct {
			alias byte
			entry updateStreamAliasEntry494A60
		}{alias: alias, entry: entry})
	}
	hooks.addToIndex2D = func(*client.Drawable) { t.Fatal("complex drawable was re-added as simple") }

	packet := []byte{
		byte(netmsg.MSG_UPDATE_STREAM),
		0xff, 0x34, 0x12, 0x78, 0x56,
		0xe8, 0x03, 0xb0, 0x04,
		0xd0, 9, 2,
		0x34, 5, 0xfd, 0xa7, 11, 4,
		0, 0, 0,
	}
	if got := handleUpdateStreamNative494A60(packet, hooks); got != len(packet) {
		t.Fatalf("consumed bytes = %d, want %d", got, len(packet))
	}
	if creates != 2 {
		t.Fatalf("create calls = %d, want 2", creates)
	}
	if cameraX != 1000 || cameraY != 1200 {
		t.Fatalf("camera = (%d,%d), want (1000,1200)", cameraX, cameraY)
	}
	if len(sent) != 1 || sent[0].alias != 0x34 || sent[0].entry != (updateStreamAliasEntry494A60{
		code: code, typeID: typeID, deadline: math.MaxUint32,
	}) {
		t.Fatalf("alias notifications = %+v", sent)
	}
	if entries[0x34] != sent[0].entry {
		t.Fatalf("stored alias = %+v, want %+v", entries[0x34], sent[0].entry)
	}
	if dr.PosVec.X != 1005 || dr.PosVec.Y != 1197 || dr.Field_72 != frame ||
		dr.AnimDir != 2 || dr.AnimFrameSlave != 11 || dr.AnimInd != 4 || dr.AnimStart != frame {
		t.Fatalf("final player drawable = pos %v, tick %d, dir %d, frame %d, anim %d, start %d",
			dr.PosVec, dr.Field_72, dr.AnimDir, dr.AnimFrameSlave, dr.AnimInd, dr.AnimStart)
	}
}

func TestHandleUpdateStreamNative494A60FullComplexRecordWithoutFrame(t *testing.T) {
	entries := new([updateStreamAliasCount494A60]updateStreamAliasEntry494A60)
	entries[2] = updateStreamAliasEntry494A60{code: 0x2345, typeID: 0x3456, deadline: math.MaxUint32}
	hooks := updateStreamTestHooks494A60(entries, 91)
	dr := highAddressUpdateStreamDrawable494A60(t)
	dr.ObjClass = object.ClassComplex
	dr.AnimFrameSlave = 99
	hooks.create = func(typeID int, code uint16, x, y int) *client.Drawable {
		if typeID != 0x3456 || code != 0x2345 || x != 300 || y != 400 {
			t.Fatalf("create = type %#x, code %#x, pos (%d,%d)", typeID, code, x, y)
		}
		dr.PosVec.X, dr.PosVec.Y = x, y
		return dr
	}

	packet := []byte{
		byte(netmsg.MSG_UPDATE_STREAM),
		1, 100, 0, 200, 0, 0x10, 1,
		0, 2, 0x2c, 0x01, 0x90, 0x01, 0x65,
		0, 0, 0,
	}
	if got := handleUpdateStreamNative494A60(packet, hooks); got != len(packet) {
		t.Fatalf("consumed bytes = %d, want %d", got, len(packet))
	}
	if dr.PosVec.X != 300 || dr.PosVec.Y != 400 || dr.Field_72 != 91 ||
		dr.AnimDir != 7 || dr.AnimFrameSlave != 99 || dr.AnimInd != 5 || dr.AnimStart != 91 {
		t.Fatalf("complex drawable = pos %v, tick %d, dir %d, frame %d, anim %d, start %d",
			dr.PosVec, dr.Field_72, dr.AnimDir, dr.AnimFrameSlave, dr.AnimInd, dr.AnimStart)
	}
}

func TestHandleUpdateStreamNative494A60SimpleDeltaRecord(t *testing.T) {
	entries := new([updateStreamAliasCount494A60]updateStreamAliasEntry494A60)
	entries[3] = updateStreamAliasEntry494A60{code: 0x3456, typeID: 0x4567, deadline: math.MaxUint32}
	hooks := updateStreamTestHooks494A60(entries, 123)
	dr := highAddressUpdateStreamDrawable494A60(t)
	dr.ObjClass = object.ClassSimple
	hooks.create = func(typeID int, code uint16, x, y int) *client.Drawable {
		if typeID != 0x4567 || code != 0x3456 || x != 105 || y != 194 {
			t.Fatalf("create = type %#x, code %#x, pos (%d,%d)", typeID, code, x, y)
		}
		dr.PosVec.X, dr.PosVec.Y = x, y
		return dr
	}
	var indexed *client.Drawable
	hooks.addToIndex2D = func(got *client.Drawable) { indexed = got }

	packet := []byte{
		byte(netmsg.MSG_UPDATE_STREAM),
		1, 100, 0, 200, 0, 0x10, 1,
		3, 5, 0xfa,
		0, 0, 0,
	}
	if got := handleUpdateStreamNative494A60(packet, hooks); got != len(packet) {
		t.Fatalf("consumed bytes = %d, want %d", got, len(packet))
	}
	if indexed != dr || dr.Field_72 != 123 {
		t.Fatalf("simple drawable index/tick = %p/%d, want %p/123", indexed, dr.Field_72, dr)
	}
}

func TestHandleUpdateStreamNative494A60RejectsTruncatedPackets(t *testing.T) {
	packet := []byte{
		byte(netmsg.MSG_UPDATE_STREAM),
		0xff, 0x34, 0x12, 0x78, 0x56,
		0xe8, 0x03, 0xb0, 0x04,
		0xd0, 9, 2,
		0x34, 5, 0xfd, 0xa7, 11, 4,
		0, 0, 0,
	}
	for n := 0; n < len(packet); n++ {
		entries := new([updateStreamAliasCount494A60]updateStreamAliasEntry494A60)
		hooks := updateStreamTestHooks494A60(entries, 77)
		dr := &client.Drawable{ObjClass: object.ClassComplex | object.ClassPlayer}
		hooks.create = func(_ int, _ uint16, x, y int) *client.Drawable {
			dr.PosVec.X, dr.PosVec.Y = x, y
			return dr
		}
		if got := handleUpdateStreamNative494A60(packet[:n], hooks); got != 0 {
			t.Fatalf("prefix of %d bytes consumed %d, want 0", n, got)
		}
	}
}

func TestHandleUpdateStreamNative494A60StopsAtInvalidPosition(t *testing.T) {
	entries := new([updateStreamAliasCount494A60]updateStreamAliasEntry494A60)
	entries[2] = updateStreamAliasEntry494A60{code: 2, typeID: 3, deadline: math.MaxUint32}
	hooks := updateStreamTestHooks494A60(entries, 10)
	creates := 0
	hooks.create = func(int, uint16, int, int) *client.Drawable { creates++; return &client.Drawable{} }
	packet := []byte{
		byte(netmsg.MSG_UPDATE_STREAM),
		1, 0x70, 0x17, 100, 0, 0x10, 1,
		2, 1, 0,
		0, 0, 0,
	}
	if got := handleUpdateStreamNative494A60(packet, hooks); got != 11 {
		t.Fatalf("consumed bytes = %d, want invalid-record prefix 11", got)
	}
	if creates != 0 {
		t.Fatalf("create calls = %d, want 0", creates)
	}
}

func TestReserveUpdateStreamAlias494A60UsesMatchExpiryAndFullSentinel(t *testing.T) {
	entries := new([updateStreamAliasCount494A60]updateStreamAliasEntry494A60)
	lookup := func(ind byte) updateStreamAliasEntry494A60 { return entries[ind] }
	store := func(ind byte, entry updateStreamAliasEntry494A60) { entries[ind] = entry }

	entries[0x34] = updateStreamAliasEntry494A60{code: 0x1234, typeID: 7, deadline: 100}
	if ind, ok := reserveUpdateStreamAlias494A60(0x1234, 7, 50, 200, lookup, store); !ok || ind != 0x34 {
		t.Fatalf("matching alias = %#x/%v, want 0x34/true", ind, ok)
	}
	if entries[0x34].deadline != 200 {
		t.Fatalf("matching alias deadline = %d, want 200", entries[0x34].deadline)
	}

	entries[1] = updateStreamAliasEntry494A60{code: 9, typeID: 9, deadline: 49}
	if ind, ok := reserveUpdateStreamAlias494A60(0xff00, 8, 50, 300, lookup, store); !ok || ind != 1 {
		t.Fatalf("expired alias = %#x/%v, want 1/true", ind, ok)
	}

	for i := 1; i < 0xff; i++ {
		entries[i] = updateStreamAliasEntry494A60{code: uint16(i + 1), typeID: 1, deadline: math.MaxUint32}
	}
	if ind, ok := reserveUpdateStreamAlias494A60(1, 2, 50, 400, lookup, store); ok || ind != 0xff {
		t.Fatalf("full alias table = %#x/%v, want 0xff/false", ind, ok)
	}
}
