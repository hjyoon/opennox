//go:build !server

package opennox

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// REA's original 4BC5D0 instructions use unsigned DIV, then signed JL.
// These literal expectations preserve the quotient/return bits as int32 before
// converting to a native Go int. Count/delay stay in the ordinary U8 asset
// domain; malformed signed-short metadata and division faults are not tested.
// This is not a whole original-player-selector or opacity oracle.

func frameFixtureR277(t *testing.T) (*Client, *client.Drawable, *client.AnimationVector) {
	t.Helper()
	s := &server.Server{}
	c := &Client{Client: &client.Client{Server: s}, srv: &Server{Server: s}}
	dr, freeDraw := alloc.New(client.Drawable{})
	ani, freeAnim := alloc.New(client.AnimationVector{})
	t.Cleanup(freeDraw)
	t.Cleanup(freeAnim)
	ani.Size = uint32(unsafe.Sizeof(*ani))
	return c, dr, ani
}

func TestAnimationFrameSignedQuotientsR277(t *testing.T) {
	for _, spec := range []struct {
		name         string
		kind         client.AnimKind
		count, delay uint16
		frame, value uint32 // value is AnimStart for OneShot, NetCode32 for Loop.
		want         int64
	}{
		{"one-first", client.AnimOneShot, 3, 1, 100, 100, 0},
		{"one-next", client.AnimOneShot, 3, 1, 102, 100, 1},
		{"one-last", client.AnimOneShot, 3, 1, 104, 100, 2},
		{"one-expired-clamp", client.AnimOneShot, 3, 1, 106, 100, 2},
		{"one-high-positive-clamp", client.AnimOneShot, 3, 0, 0x7fffffff, 0, 2},
		{"one-high-bit-negative", client.AnimOneShot, 3, 0, 0x80000000, 0, -2147483648},
		{"one-high-bit-negative-next", client.AnimOneShot, 255, 0, 0x80000001, 0, -2147483647},
		{"one-minus-one-return", client.AnimOneShot, 3, 0, 0xffffffff, 0, -1},
		{"one-start-in-future", client.AnimOneShot, 3, 0, 0, 1, -1},
		{"one-elapsed-wrap-before-divide", client.AnimOneShot, 255, 1, 0x10, 0xfffffff0, 16},
		{"one-unsigned-divide-clears-sign", client.AnimOneShot, 3, 1, 0xffffffff, 0, 2},
		{"one-U8-maximum-delay", client.AnimOneShot, 255, 255, 512, 0, 2},
		{"one-count-one-negative-not-clamp", client.AnimOneShot, 1, 0, 0x80000000, 0, -2147483648},
		{"loop-before-modulo", client.AnimLoop, 3, 1, 0, 5, 2},
		{"loop-modulo", client.AnimLoop, 3, 1, 1, 5, 0},
		{"loop-positive-signed-modulo", client.AnimLoop, 3, 0, 0x7fffffff, 0, 1},
		{"loop-high-bit-negative-no-modulo", client.AnimLoop, 3, 0, 0x80000000, 0, -2147483648},
		{"loop-minus-one-no-modulo", client.AnimLoop, 3, 0, 0xffffffff, 0, -1},
		{"loop-add-wrap-before-divide", client.AnimLoop, 7, 0, 0xfffffff0, 0x20, 2},
		{"loop-add-enters-negative", client.AnimLoop, 3, 0, 0x7fffffff, 1, -2147483648},
		{"loop-add-wraps-to-zero", client.AnimLoop, 3, 0, 0xffffffff, 1, 0},
		{"loop-unsigned-divide-clears-sign", client.AnimLoop, 3, 1, 0xffffffff, 0, 1},
		{"loop-U8-maximum-delay", client.AnimLoop, 3, 255, 0xffffffff, 0, 0},
	} {
		t.Run(spec.name, func(t *testing.T) {
			c, dr, ani := frameFixtureR277(t)
			c.Server.SetFrame(spec.frame)
			ani.Kind, ani.Cnt40, ani.Val42 = spec.kind, spec.count, spec.delay
			if spec.kind == client.AnimLoop {
				dr.NetCode32 = spec.value
			} else {
				dr.AnimStart = spec.value
			}
			beforeDraw, beforeAnim := *dr, *ani
			got := c.getAnimFrameInd(dr, ani)
			if int64(got) != spec.want {
				t.Fatalf("kind=%d count=%d delay=%d frame=%#x value=%#x: frame=%d (%#x), want=%d", spec.kind, spec.count, spec.delay, spec.frame, spec.value, got, uint64(got), spec.want)
			}
			if *dr != beforeDraw || *ani != beforeAnim || c.Server.Frame() != spec.frame {
				t.Fatal("pure frame selection changed drawable, animation metadata, or server frame")
			}
		})
	}
}

func TestAnimationFrameSlaveSignedReturnR277(t *testing.T) {
	// Preserve the raw EAX bits for existing signed negative-frame consumers.
	for _, spec := range []struct {
		slave uint32
		want  int64
	}{{0, 0}, {1, 1}, {255, 255}, {0x7fffffff, 2147483647},
		{0x80000000, -2147483648}, {0xfffffffe, -2}, {0xffffffff, -1}} {
		t.Run(fmt.Sprintf("slave-%08x", spec.slave), func(t *testing.T) {
			c, dr, ani := frameFixtureR277(t)
			ani.Kind, ani.Cnt40, ani.Val42 = client.AnimSlave, 3, 255
			dr.AnimFrameSlave = spec.slave
			beforeDraw, beforeAnim := *dr, *ani
			got := c.getAnimFrameInd(dr, ani)
			if int64(got) != spec.want || *dr != beforeDraw || *ani != beforeAnim {
				t.Fatalf("Slave bits=%#x returned=%d, want=%d, or changed input state", spec.slave, got, spec.want)
			}
		})
	}
}

func frameRemoveFixtureR277(t *testing.T) (*Client, *client.Drawable, *client.AnimationVector) {
	t.Helper()
	// Real pool/list ownership permits the unchanged baseline to delete a
	// negative-quotient input safely, producing an observable wrong lifecycle
	// result instead of faulting on a fake/stack drawable. No images/assets.
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	cl := client.NewClient(nil, nil, s)
	t.Cleanup(cl.Close)
	t.Cleanup(func() { alloc.FreePtr(cl.Viewport().C()) })
	t.Cleanup(cl.Render().Part.Free)
	cl.Objs.Init(4)
	t.Cleanup(cl.Objs.Free)
	c := &Client{Client: cl, srv: &Server{Server: s}}
	dr := cl.Objs.New()
	if dr == nil {
		t.Fatal("real isolated drawable pool allocation failed")
	}
	cl.Objs.List1, cl.Objs.Count = dr, 1
	ani, freeAnim := alloc.New(client.AnimationVector{})
	t.Cleanup(freeAnim)
	ani.Size = uint32(unsafe.Sizeof(*ani))
	ani.Kind, ani.Cnt40 = client.AnimOneShotRemove, 3
	return c, dr, ani
}

func TestAnimationFrameRemoveSignedBoundaryR277(t *testing.T) {
	for _, spec := range []struct {
		name         string
		delay        uint16
		frame, start uint32
		want         int64
		alive        bool
	}{
		{"first", 0, 0, 0, 0, true},
		{"last", 0, 2, 0, 2, true},
		{"ordinary-expiry", 0, 3, 0, -1, false},
		{"high-positive-expiry", 0, 0x7fffffff, 0, -1, false},
		{"negative-minimum-not-expired", 0, 0x80000000, 0, -2147483648, true},
		{"negative-minus-one-not-expired", 0, 0xffffffff, 0, -1, true},
		{"future-start-not-expired", 0, 0, 1, -1, true},
		{"U8-delay-before-expiry", 255, 767, 0, 2, true},
		{"U8-delay-expiry", 255, 768, 0, -1, false},
	} {
		t.Run(spec.name, func(t *testing.T) {
			c, dr, ani := frameRemoveFixtureR277(t)
			c.Server.SetFrame(spec.frame)
			dr.AnimStart, ani.Val42 = spec.start, spec.delay
			beforeDraw, beforeAnim := *dr, *ani
			got := c.getAnimFrameInd(dr, ani)
			alive := c.Objs.Count == 1 && c.Objs.List1 == dr
			if int64(got) != spec.want || alive != spec.alive || c.Objs.Count != boolToFrameCountR277(spec.alive) {
				t.Fatalf("Remove frame=%#x start=%#x delay=%d: result=%d count=%d alive=%t, want=%d/%t", spec.frame, spec.start, spec.delay, got, c.Objs.Count, alive, spec.want, spec.alive)
			}
			if !spec.alive && c.Objs.List1 != nil {
				t.Fatal("expired drawable still linked")
			}
			// Never dereference a record after it was returned to its real pool.
			if spec.alive && *dr != beforeDraw || *ani != beforeAnim || c.Server.Frame() != spec.frame {
				t.Fatal("unexpected live drawable/metadata/frame mutation")
			}
		})
	}
}

func boolToFrameCountR277(alive bool) int {
	if alive {
		return 1
	}
	return 0
}

func TestAnimationFrameDefaultAndZeroCountControlsR277(t *testing.T) {
	for _, spec := range []struct {
		kind  client.AnimKind
		count uint16
		want  int64
	}{{client.AnimLoopAndFade, 3, 0}, {client.AnimKind(6), 3, 0},
		{client.AnimKind(0xffffffff), 3, 0}, {client.AnimOneShot, 0, -1}} {
		t.Run(fmt.Sprintf("kind-%d-count-%d", spec.kind, spec.count), func(t *testing.T) {
			c, dr, ani := frameFixtureR277(t)
			c.Server.SetFrame(0xffffffff)
			ani.Kind, ani.Cnt40, ani.Val42 = spec.kind, spec.count, 255
			beforeDraw, beforeAnim := *dr, *ani
			if got := c.getAnimFrameInd(dr, ani); int64(got) != spec.want || *dr != beforeDraw || *ani != beforeAnim {
				t.Fatalf("default/count control got=%d, want=%d, or input state changed", got, spec.want)
			}
		})
	}
}
