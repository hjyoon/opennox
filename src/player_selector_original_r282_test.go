//go:build !server

package opennox

import (
	"encoding/binary"
	"testing"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

// These literal row/frame results were observed by executing the preserved PE32
// selector 4B8FA0, frame helper 4BC5D0 and chat wrapper 48D830 in Unicorn 2.1.4,
// not computed using Go selection/frame arithmetic. GAME.EXE SHA256:
// 0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a
// Original report SHA256:
// 729b67f7c1964376c116c7c18a93ea13704097c681e4965ec76d3a063de9fda4
// Retained report: original-selector-observations-v4-host/report.json under
// /private/tmp/opennox-selector-run-r282.Up3RyQ; driver SHA256:
// 9d2031f067460f2ad03beb22c3d57d0fae585934556661951f85af6bbb399ee9
//
// The original's four external services have declared synthetic returns, not
// emulated implementations. This Go comparison projects 50/56 original cases:
// four nonempty Random cases and two actual deletion requests are excluded.
// It compares consumer-visible image/row/frame, not PE32 addresses or physical
// C output slots (Go returns a tuple). Original nonzero lookup scalars all map
// to one real native chat bubble: their zero/nonzero meaning, not exact pointer
// bits, is compared. This is neither full PlayerDraw nor personal opacity proof.
func TestPlayerSelectorOriginalObservationsR282(t *testing.T) {
	if legacy.Get_nox_alloc_chat_1197364() != nil {
		t.Fatal("chat pool must be uninitialized for the owned fixture")
	}
	handles.Init()
	t.Cleanup(handles.Release)
	pool := alloc.NewClass("player-selector-original-r282", legacy.ChatBubbleSize(), 1)
	legacy.Set_nox_alloc_chat_1197364(pool.UPtr())
	t.Cleanup(func() {
		legacy.ResetChatBubbles()
		legacy.Set_nox_alloc_chat_1197364(nil)
		pool.Free()
	})

	cases := []struct {
		name                      string
		raw, row                  int
		dir                       byte
		count, delay              uint16
		kind                      client.AnimKind
		frame, start, code, slave uint32
		chat, sneak, nullImage    bool
		wantFrame                 int
	}{
		{"direction-0", 1, 1, 0, 3, 0, 0, 0x00000001, 0x00000000, 0x00000000, 0x00000001, false, false, false, 1},
		{"direction-1", 1, 1, 1, 3, 0, 0, 0x00000001, 0x00000000, 0x00000000, 0x00000001, false, false, false, 1},
		{"direction-2", 1, 1, 2, 3, 0, 0, 0x00000001, 0x00000000, 0x00000000, 0x00000001, false, false, false, 1},
		{"direction-3", 1, 1, 3, 3, 0, 0, 0x00000001, 0x00000000, 0x00000000, 0x00000001, false, false, false, 1},
		{"direction-5", 1, 1, 5, 3, 0, 0, 0x00000001, 0x00000000, 0x00000000, 0x00000001, false, false, false, 1},
		{"direction-6", 1, 1, 6, 3, 0, 0, 0x00000001, 0x00000000, 0x00000000, 0x00000001, false, false, false, 1},
		{"direction-7", 1, 1, 7, 3, 0, 0, 0x00000001, 0x00000000, 0x00000000, 0x00000001, false, false, false, 1},
		{"direction-8", 1, 1, 8, 3, 0, 0, 0x00000001, 0x00000000, 0x00000000, 0x00000001, false, false, false, 1},
		{"raw-idle-lookup-zero", 0, 0, 0, 3, 0, 0, 0x00000001, 0x00000000, 0x00000000, 0x00000001, false, false, false, 1},
		{"raw-idle-lookup-one", 0, 19, 0, 3, 0, 0, 0x00000001, 0x00000000, 0x00000000, 0x00000001, true, false, false, 1},
		{"raw-idle-lookup-high-bit", 0, 19, 0, 3, 0, 0, 0x00000001, 0x00000000, 0x00000000, 0x00000001, true, false, false, 1},
		{"raw-idle-lookup-all-bits", 0, 19, 0, 3, 0, 0, 0x00000001, 0x00000000, 0x00000000, 0x00000001, true, false, false, 1},
		{"raw-walk-enchant-absent", 4, 4, 0, 3, 0, 0, 0x00000001, 0x00000000, 0x00000000, 0x00000001, false, false, false, 1},
		{"raw-walk-enchant-present", 4, 53, 0, 3, 0, 0, 0x00000001, 0x00000000, 0x00000000, 0x00000001, false, true, false, 1},
		{"last-animation-row-54", 54, 54, 8, 3, 0, 5, 0x00000001, 0x00000000, 0x00000000, 0x00000001, false, false, false, 1},
		{"zero-count-both", 1, 1, 0, 0, 0, 4, 0x00000001, 0x00000000, 0x00000000, 0x00000001, false, false, false, -1},
		{"zero-count-row-only", 1, 1, 0, 0, 0, 0, 0x00000001, 0x00000000, 0x00000000, 0x00000001, false, false, false, -1},
		{"zero-count-frame-only", 1, 1, 0, 0, 0, 0, 0x00000001, 0x00000000, 0x00000000, 0x00000001, false, false, false, -1},
		{"zero-count-no-outputs", 1, 1, 0, 0, 0, 0, 0x00000001, 0x00000000, 0x00000000, 0x00000001, false, false, false, -1},
		{"zero-count-after-walk-remap", 4, 53, 0, 0, 0, 0, 0x00000001, 0x00000000, 0x00000000, 0x00000001, false, true, false, -1},
		{"one-frame-delay-255", 1, 1, 0, 1, 255, 0, 0x00000100, 0x00000000, 0x00000000, 0x00000001, false, false, false, 0},
		{"count-255-delay-255", 1, 1, 0, 255, 255, 0, 0x00000100, 0x00000000, 0x00000000, 0x00000001, false, false, false, 1},
		{"oneshot-exact-last", 1, 1, 0, 3, 0, 0, 0x00000002, 0x00000000, 0x00000000, 0x00000001, false, false, false, 2},
		{"oneshot-positive-overshoot", 1, 1, 0, 3, 0, 0, 0x00000003, 0x00000000, 0x00000000, 0x00000001, false, false, false, 2},
		{"oneshot-subtraction-wrap", 1, 1, 0, 3, 1, 0, 0x00000005, 0xfffffff0, 0x00000000, 0x00000001, false, false, false, 2},
		{"oneshot-positive-quotient-max", 1, 1, 0, 3, 0, 0, 0x7fffffff, 0x00000000, 0x00000000, 0x00000001, false, false, false, 2},
		{"oneshot-negative-quotient-min", 1, 1, 0, 3, 0, 0, 0x80000000, 0x00000000, 0x00000000, 0x00000001, false, false, false, -1},
		{"oneshot-negative-quotient-all-bits", 1, 1, 0, 3, 0, 0, 0xffffffff, 0x00000000, 0x00000000, 0x00000001, false, false, false, -1},
		{"loop-before-wrap", 1, 1, 0, 3, 0, 2, 0x00000002, 0x00000000, 0x00000000, 0x00000001, false, false, false, 2},
		{"loop-positive-modulo", 1, 1, 0, 3, 0, 2, 0x00000007, 0x00000000, 0x00000000, 0x00000001, false, false, false, 1},
		{"loop-frame-plus-netcode-wrap", 1, 1, 0, 3, 1, 2, 0xfffffff0, 0x00000000, 0x00000030, 0x00000001, false, false, false, 1},
		{"loop-full-addition-wrap", 1, 1, 0, 3, 0, 2, 0xffffffff, 0x00000000, 0x00000001, 0x00000001, false, false, false, 0},
		{"loop-delay-255", 1, 1, 0, 3, 255, 2, 0x00000100, 0x00000000, 0x00000000, 0x00000001, false, false, false, 1},
		{"loop-positive-quotient-max", 1, 1, 0, 3, 0, 2, 0x7fffffff, 0x00000000, 0x00000000, 0x00000001, false, false, false, 1},
		{"loop-negative-quotient-min", 1, 1, 0, 3, 0, 2, 0x80000000, 0x00000000, 0x00000000, 0x00000001, false, false, false, -1},
		{"loop-negative-quotient-all-bits", 1, 1, 0, 3, 0, 2, 0xffffffff, 0x00000000, 0x00000000, 0x00000001, false, false, false, -1},
		{"slave-zero", 1, 1, 0, 3, 0, 5, 0x00000001, 0x00000000, 0x00000000, 0x00000000, false, false, false, 0},
		{"slave-exact-last", 1, 1, 0, 3, 0, 5, 0x00000001, 0x00000000, 0x00000000, 0x00000002, false, false, false, 2},
		{"slave-positive-overshoot", 1, 1, 0, 3, 0, 5, 0x00000001, 0x00000000, 0x00000000, 0x00000003, false, false, false, 2},
		{"slave-positive-max", 1, 1, 0, 3, 0, 5, 0x00000001, 0x00000000, 0x00000000, 0x7fffffff, false, false, false, 2},
		{"slave-negative-min", 1, 1, 0, 3, 0, 5, 0x00000001, 0x00000000, 0x00000000, 0x80000000, false, false, false, -1},
		{"slave-negative-all-bits", 1, 1, 0, 3, 0, 5, 0x00000001, 0x00000000, 0x00000000, 0xffffffff, false, false, false, -1},
		{"remove-before-expiry", 1, 1, 0, 3, 0, 1, 0x00000002, 0x00000000, 0x00000000, 0x00000001, false, false, false, 2},
		{"remove-negative-quotient-min", 1, 1, 0, 3, 0, 1, 0x80000000, 0x00000000, 0x00000000, 0x00000001, false, false, false, -1},
		{"remove-negative-quotient-all-bits", 1, 1, 0, 3, 0, 1, 0xffffffff, 0x00000000, 0x00000000, 0x00000001, false, false, false, -1},
		{"kind-three-original-default", 1, 1, 0, 3, 0, 3, 0xffffffff, 0x00000000, 0x00000000, 0x00000001, false, false, false, 0},
		{"selected-row-output-only", 1, 1, 0, 3, 0, 5, 0x00000001, 0x00000000, 0x00000000, 0x00000001, false, false, false, 1},
		{"selected-frame-output-only", 1, 1, 0, 3, 0, 5, 0x00000001, 0x00000000, 0x00000000, 0x00000001, false, false, false, 1},
		{"selected-no-output-pointers", 1, 1, 0, 3, 0, 5, 0x00000001, 0x00000000, 0x00000000, 0x00000001, false, false, false, 1},
		{"selected-null-image-not-early-return", 1, 1, 0, 3, 0, 5, 0x00000001, 0x00000000, 0x00000000, 0x00000001, false, false, true, 1},
	}
	if len(cases) != 50 {
		t.Fatal("the reviewed original observation projection must have 50 cases")
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newPlayerSelectorFixtureR277(t, tc.row, tc.count, tc.delay, tc.kind, 17)
			f.dr.AnimInd, f.dr.AnimDir = uint32(tc.raw), tc.dir
			f.dr.AnimStart, f.dr.NetCode32, f.dr.AnimFrameSlave = tc.start, tc.code, tc.slave
			if tc.sneak {
				f.dr.Buffs = uint32(1) << uint32(server.ENCHANT_SNEAK)
			}
			if tc.nullImage {
				f.nakedFrames[tc.row][tc.dir][tc.wantFrame] = nil
			}
			f.c.Server.SetFrame(tc.frame)
			oldClient, oldServer := noxClient, noxServer
			noxClient, noxServer = f.c, f.c.srv
			t.Cleanup(func() { noxClient, noxServer = oldClient, oldServer })
			legacy.ResetChatBubbles()
			t.Cleanup(legacy.ResetChatBubbles)
			if tc.chat {
				if tc.code > 0xffff {
					t.Fatal("native chat packet fixture only admits uint16 netcodes")
				}
				var packet [11]byte
				binary.LittleEndian.PutUint16(packet[1:3], uint16(tc.code))
				binary.LittleEndian.PutUint16(packet[9:11], 90)
				legacy.Nox_xxx_createTextBubble_48D880(packet[:], "original-selector")
			}
			beforeChat := legacy.Nox_xxx_netCode2ChatBubble_48D850(int(tc.code))
			if (beforeChat != 0) != tc.chat {
				t.Fatal("actual native chat list did not establish the declared zero/nonzero input")
			}
			// The existing typed fixture supplies distinct base/Naked/equipment
			// handles in all eight legal directions and checks all input bytes
			// plus both RNG states; no production function predicts the frame.
			f.checkSelection(t, tc.row, tc.wantFrame, nil)
			if legacy.Nox_xxx_netCode2ChatBubble_48D850(int(tc.code)) != beforeChat {
				t.Fatal("selection changed the native chat list")
			}
		})
	}
}
