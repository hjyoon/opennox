//go:build !server

package opennox

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/opennox/libs/datapath"
	"github.com/opennox/libs/noxfont"
	"github.com/opennox/libs/noximage"
	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

const r284ChatEnv = "OPENNOX_R284_LINKED_CHAT_CASE"
const r284ChatMarker = "R284_LINKED_CHAT "

var r284ChatCases = [...]struct {
	id, name        string
	player, present bool
	column          int
}{
	{"ASCII", "A", true, true, 0},
	{"Korean", "한", true, true, 2},
	{"empty-name", "", true, true, -1},
	{"missing-player", "A", true, false, -1},
	{"non-player", "A", false, true, -1},
}

// Invoke the linked production C chat renderer, its actual player lookup and
// UTF-16/text raster callbacks. Generated bitmap fonts make the observed name
// pixels independent of host fonts or artwork. Fresh processes own all native
// chat/global/font state. This is software rendering, not a visible GUI, original
// binary oracle, full game session, or Host Load opacity reproduction.
func TestLinkedChatPlayerNameRasterR284(t *testing.T) {
	if id := os.Getenv(r284ChatEnv); id != "" {
		for i := range r284ChatCases {
			if r284ChatCases[i].id == id {
				r284LinkedChatChild(t, i)
				return
			}
		}
		t.Fatalf("unknown explicit child case %q", id)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range r284ChatCases {
		t.Run(tc.id, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, executable, "-test.run=^TestLinkedChatPlayerNameRasterR284$", "-test.count=1", "-test.v")
			for _, value := range os.Environ() {
				if !strings.HasPrefix(value, r284ChatEnv+"=") {
					cmd.Env = append(cmd.Env, value)
				}
			}
			cmd.Env = append(cmd.Env, r284ChatEnv+"="+tc.id)
			output, err := cmd.CombinedOutput()
			if ctx.Err() != nil || err != nil || cmd.ProcessState == nil || !cmd.ProcessState.Exited() || cmd.ProcessState.ExitCode() != 0 {
				t.Fatalf("linked native chat child did not pass and settle: %v, deadline=%v\n%s", err, ctx.Err(), output)
			}
			if strings.Contains(string(output), "--- SKIP:") || bytes.Count(output, []byte(r284ChatMarker)) != 1 {
				t.Fatalf("linked native chat child lacked one non-skipped completion: %s", output)
			}
			for _, line := range strings.Split(string(output), "\n") {
				if !strings.HasPrefix(line, r284ChatMarker) {
					continue
				}
				var receipt struct {
					Case             string `json:"case"`
					PID              int    `json:"pid"`
					Parent           int    `json:"parent"`
					LinkedCalls      int    `json:"linked_calls"`
					BandPassed       bool   `json:"name_band_passed"`
					InputsPreserved  bool   `json:"inputs_preserved"`
					OriginalExecuted bool   `json:"original_executed"`
					OpacityFixed     bool   `json:"personal_opacity_fixed"`
				}
				if err := json.Unmarshal([]byte(strings.TrimPrefix(line, r284ChatMarker)), &receipt); err != nil {
					t.Fatal(err)
				}
				if receipt.Case != tc.id || receipt.PID != cmd.Process.Pid || receipt.PID == os.Getpid() || receipt.Parent != os.Getpid() || receipt.LinkedCalls != 1 || !receipt.BandPassed || !receipt.InputsPreserved || receipt.OriginalExecuted || receipt.OpacityFixed {
					t.Fatalf("linked native chat receipt differs: %+v", receipt)
				}
			}
		})
	}
}

func r284LinkedChatChild(t *testing.T, index int) {
	t.Helper()
	tc := r284ChatCases[index]
	legacy.InitBlobData()
	t.Chdir(t.TempDir())
	r284WriteChatFonts(t)
	oldData := datapath.Data()
	datapath.SetData(".")
	t.Cleanup(func() { datapath.SetData(oldData) })
	handles.Init()
	t.Cleanup(handles.Release)
	c, _, draw := actualVideoRender200(t)
	if err := c.r.GetFonts().Load(1); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(c.r.GetFonts().Free)
	pix := noximage.NewImage16(image.Rect(0, 0, 256, 256))
	c.r.SetPixBuffer(pix)
	draw.SetRect3(pix.Rect)
	draw.SetClip(true)
	draw.SetClipRect(pix.Rect)
	draw.SetClipRect2(image.Rect(0, 0, 255, 255))
	draw.SetAlphaEnabled(false)
	legacy.Set_nox_draw_curDrawData_3799572(draw)
	legacy.Set_nox_win_width(256)
	legacy.Set_nox_win_height(256)
	legacy.Set_nox_color_white_2523948(0x7fff)
	*memmap.PtrUint32(0x852978, 4) = 0
	*memmap.PtrUint32(0x85B3FC, 956) = 0
	nox_client_renderGUI_80828 = false
	noxflags.ResetGame()
	noxflags.ResetEngine()
	viewport := c.Viewport()
	viewport.Screen, viewport.World, viewport.Size = pix.Rect, pix.Rect, pix.Size()
	c.Objs.Init(2)
	t.Cleanup(c.Objs.Free)
	dr := c.Objs.New()
	if dr == nil {
		t.Fatal("real drawable allocation failed")
	}
	dr.NetCode32, dr.PosVec = 10, image.Pt(100, 200)
	dr.ObjClass = object.ClassMonster
	if tc.player {
		dr.ObjClass = object.ClassPlayer
	}
	c.Objs.Count, c.Objs.List1 = 1, dr
	pl := c.Server.Players.ResetInd(0)
	pl.NetCodeVal = 10
	pl.SetName(tc.name)
	if !tc.present {
		pl.NetCodeVal = 11
	}
	playerBytes := unsafe.Slice((*byte)(pl.C()), int(unsafe.Sizeof(*pl)))
	if unsafe.Offsetof(server.Player{}.NameFinal) != 4704 {
		// This is native C-owned memory, never a Go-scanned Player value.
		// Keep the old location a valid short UTF-16 decoy, not a crash probe.
		copy(playerBytes[4704:4712], []byte{'Z', 0, 0, 0, 0, 0, 0, 0})
	}
	beforePlayer := bytes.Clone(playerBytes)
	beforeDR, beforeViewport := *dr, *viewport
	beforeFrame := c.Server.Frame()
	logic, other := c.Server.Rand.Logic.Index(), c.Server.Rand.Other.Index()
	pool := alloc.NewClass("linked-chat-r284", legacy.ChatBubbleSize(), 2)
	legacy.Set_nox_alloc_chat_1197364(pool.UPtr())
	t.Cleanup(func() {
		legacy.ResetChatBubbles()
		legacy.Set_nox_alloc_chat_1197364(nil)
		pool.Free()
	})
	packet := make([]byte, 11)
	binary.LittleEndian.PutUint16(packet[1:], 10)
	packet[8] = 8
	binary.LittleEndian.PutUint16(packet[9:], 90)
	legacy.Nox_xxx_createTextBubble_48D880(packet, "m")
	chat := legacy.Nox_xxx_netCode2ChatBubble_48D850(10)
	if chat == 0 || c.Objs.ByNetCodeDynamic(10) != dr || c.r.FontHeight(nil) != 4 {
		t.Fatal("real chat/drawable/font ownership not established")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(pl.C()) <= 0xffffffff || uintptr(dr.C()) <= 0xffffffff || uintptr(chat) <= 0xffffffff) {
		t.Fatal("native C-owned high addresses required")
	}
	clear(pix.Pix)
	legacy.Sub_437180() // The existing linked C renderer, exactly once.
	if !bytes.Equal(playerBytes, beforePlayer) || *dr != beforeDR || *viewport != beforeViewport || c.Server.Frame() != beforeFrame || c.Server.Rand.Logic.Index() != logic || c.Server.Rand.Other.Index() != other || legacy.Nox_xxx_netCode2ChatBubble_48D850(10) != chat {
		t.Fatal("linked chat rendering changed player/drawable/viewport/frame/RNG/list identity")
	}
	// A one-pixel-wide body glyph gives the original layout x99/y132; the
	// second outline pass shifts text to98/131, then the name starts at98/126.
	// Foreground masks are literal columns, not rendered reference strings.
	for y := 126; y < 130; y++ {
		for x := 98; x < 102; x++ {
			var want uint16
			if tc.column >= 0 && x == 98+tc.column {
				want = 0x7fff
			}
			if got := pix.Pix[y*pix.Stride+x]; got != want {
				t.Fatalf("native name-band pixel(%d,%d)=%04x want=%04x case=%s (legacy byte4704 contains Z)", x, y, got, want, tc.id)
			}
		}
	}
	receipt, err := json.Marshal(map[string]any{"case": tc.id, "pid": os.Getpid(), "parent": os.Getppid(), "linked_calls": 1, "name_band_passed": true, "inputs_preserved": true, "original_executed": false, "personal_opacity_fixed": false})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(r284ChatMarker + string(receipt))
}

func r284WriteChatFonts(t *testing.T) {
	t.Helper()
	f := &noxfont.Font{}
	for _, glyph := range []struct {
		char  uint16
		mask  byte
		width int
	}{{'A', 0x80, 3}, {'Z', 0x40, 3}, {'m', 0x80, 1}, {'한', 0x20, 3}} {
		f.Ranges = append(f.Ranges, noxfont.Range{StartChar: glyph.char, EndChar: glyph.char, Glyphs: []*noxfont.Bitmap{{Pix: []byte{glyph.mask, glyph.mask, glyph.mask, glyph.mask}, Stride: 1, Rect: image.Rect(0, 0, glyph.width, 4)}}})
	}
	data, err := f.Encode()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"default", "large", "number", "small"} {
		if err := os.WriteFile(name+noxfont.Ext, data, 0600); err != nil {
			t.Fatal(err)
		}
	}
}
