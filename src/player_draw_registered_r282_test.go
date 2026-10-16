//go:build !server

package opennox

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/opennox/libs/noximage"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

const (
	r282PlayerDrawCaseEnv = "OPENNOX_R282_PLAYER_DRAW_CASE"
	r282PlayerDrawMarker  = "R282_REGISTERED_PLAYER_DRAW "
)

type r282PlayerDrawCase struct {
	id      string
	alpha   byte
	enabled bool
	dir     byte
}

var r282PlayerDrawCases = [...]r282PlayerDrawCase{
	{"stale-0-direction-0", 0, false, 0},
	{"stale-28-direction-0", 28, false, 0},
	{"stale-128-direction-0", 128, false, 0},
	{"enabled-128-direction-0", 128, true, 0},
	{"stale-0-direction-5", 0, false, 5},
	{"stale-28-direction-5", 28, false, 5},
	{"stale-128-direction-5", 128, false, 5},
	{"enabled-128-direction-5", 128, true, 5},
}

type r282PlayerDrawReceipt struct {
	Case                    string `json:"case"`
	PID                     int    `json:"pid"`
	ParentPID               int    `json:"parent_pid"`
	NativeWordBits          int    `json:"native_word_bits"`
	RegisteredCalls         int    `json:"registered_calls"`
	FixedSlaveFrame         uint32 `json:"fixed_slave_frame"`
	AlphaBefore             byte   `json:"alpha_before"`
	AlphaEnabledBefore      bool   `json:"alpha_enabled_before"`
	AlphaAfter              byte   `json:"alpha_after"`
	AlphaEnabledAfter       bool   `json:"alpha_enabled_after"`
	BodyWord                uint16 `json:"body_word"`
	ArmorWord               uint16 `json:"armor_word"`
	WeaponWord              uint16 `json:"weapon_word"`
	LastAnchorWord          uint16 `json:"last_anchor_word"`
	SameFrameAndOrderPassed bool   `json:"same_frame_and_order_passed"`
	WholeSoftwareRaster     bool   `json:"whole_software_raster_passed"`
	OriginalExecuted        bool   `json:"original_executed"`
	PhysicalGPUCertified    bool   `json:"physical_gpu_certified"`
	PersonalOpacityFixed    bool   `json:"personal_opacity_fixed"`
}

// The registered PlayerDraw -> native C drawObject -> real type-4 RLE raster
// path is invoked once in each fresh test-binary child. Inherited enabled alpha
// is an explicit input, not a proposed production reset. Fresh processes also
// isolate drawObject's previous_drawable/previous_z and chat/Ghost C statics;
// restoring Go globals alone would not restore all those native caches.
//
// This is an independent colored fixture, not stock artwork, an original binary
// oracle, a personal Host Load reproduction, or a window/GPU/compositor test.
func TestRegisteredPlayerDrawRasterR282(t *testing.T) {
	if id := os.Getenv(r282PlayerDrawCaseEnv); id != "" {
		for _, tc := range r282PlayerDrawCases {
			if id == tc.id {
				r282RunRegisteredPlayerDrawChild(t, tc)
				return
			}
		}
		t.Fatalf("unknown explicit child case %q", id)
	}

	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range r282PlayerDrawCases {
		t.Run(tc.id, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, executable,
				"-test.run=^TestRegisteredPlayerDrawRasterR282$", "-test.count=1", "-test.v")
			// Do not synthesize game-data/config flags or pass a golden. The only
			// added environment input is the literal, allowlisted case identifier.
			for _, v := range os.Environ() {
				if !strings.HasPrefix(v, r282PlayerDrawCaseEnv+"=") {
					cmd.Env = append(cmd.Env, v)
				}
			}
			cmd.Env = append(cmd.Env, r282PlayerDrawCaseEnv+"="+tc.id)
			out, runErr := cmd.CombinedOutput()
			if ctx.Err() != nil {
				t.Fatalf("fresh registered PlayerDraw child exceeded 30 seconds: %v\n%s", ctx.Err(), out)
			}
			if runErr != nil {
				t.Fatalf("fresh registered PlayerDraw child failed (panic/crash is not a pass): %v\n%s", runErr, out)
			}
			if cmd.ProcessState == nil || !cmd.ProcessState.Exited() || cmd.ProcessState.ExitCode() != 0 {
				t.Fatalf("fresh registered PlayerDraw child did not settle at exit 0:\n%s", out)
			}
			if strings.Contains(string(out), "--- SKIP:") {
				t.Fatalf("fresh registered PlayerDraw child skipped instead of executing:\n%s", out)
			}
			var receipts []r282PlayerDrawReceipt
			for _, line := range strings.Split(string(out), "\n") {
				if !strings.HasPrefix(line, r282PlayerDrawMarker) {
					continue
				}
				var receipt r282PlayerDrawReceipt
				if err := json.Unmarshal([]byte(strings.TrimPrefix(line, r282PlayerDrawMarker)), &receipt); err != nil {
					t.Fatalf("invalid child receipt: %v\n%s", err, out)
				}
				receipts = append(receipts, receipt)
			}
			if len(receipts) != 1 {
				t.Fatalf("expected one successful callback receipt, got %d:\n%s", len(receipts), out)
			}
			got := receipts[0]
			body, anchor := r282ExpectedPlayerDrawWords(tc)
			if got.Case != tc.id || got.PID != cmd.Process.Pid || got.PID == os.Getpid() || got.ParentPID != os.Getpid() ||
				got.NativeWordBits != int(unsafe.Sizeof(uintptr(0))*8) || got.RegisteredCalls != 1 || got.FixedSlaveFrame != 1 ||
				got.AlphaBefore != tc.alpha || got.AlphaEnabledBefore != tc.enabled || got.AlphaAfter != tc.alpha || got.AlphaEnabledAfter ||
				got.BodyWord != body || got.ArmorWord != 0x001e || got.WeaponWord != 0x7800 || got.LastAnchorWord != anchor ||
				!got.SameFrameAndOrderPassed || !got.WholeSoftwareRaster || got.OriginalExecuted || got.PhysicalGPUCertified || got.PersonalOpacityFixed {
				t.Fatalf("unexpected fresh-child receipt: %+v\n%s", got, out)
			}
			encoded, err := json.Marshal(got)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("%s%s", r282PlayerDrawMarker, encoded)
		})
	}
}

// Literal RGB555 words are independent of SplitColor, ColorNRGBA, the production
// selector and blend functions. Full-white multiply maps source 31 to 30; the
// explicit enabled-128 body control maps its green component to 15 over black.
func r282ExpectedPlayerDrawWords(tc r282PlayerDrawCase) (body, anchor uint16) {
	body = 0x03c0
	if tc.enabled {
		body = 0x01e0
	}
	anchor = 0x001e // direction 0: body, weapon, armor
	if tc.dir == 5 {
		anchor = 0x7800 // direction 5: body, armor, weapon
	}
	return body, anchor
}

func r282RunRegisteredPlayerDrawChild(t *testing.T, tc r282PlayerDrawCase) {
	t.Helper()
	t.Chdir(t.TempDir())
	r282WritePlayerRLEBag(t)
	handles.Init()
	t.Cleanup(handles.Release)
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	cl := client.NewClient(nil, nil, s)
	t.Cleanup(cl.Close)
	t.Cleanup(func() { alloc.FreePtr(cl.Viewport().C()) })
	cl.Objs.Init(4)
	t.Cleanup(cl.Objs.Free)
	r := cl.Render()
	t.Cleanup(r.Part.Free)
	draw, freeDraw := noxrender.NewRenderData()
	t.Cleanup(freeDraw)
	r.SetData(draw)
	pix := noximage.NewImage16(image.Rect(0, 0, 64, 64))
	r.SetPixBuffer(pix)
	if pix.Stride != 64 || len(pix.Pix) != 64*64 {
		t.Fatalf("unexpected independent raster dimensions: stride=%d words=%d", pix.Stride, len(pix.Pix))
	}
	if err := r.GetBag().ReadVideoBag(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(r.GetBag().Free)
	c := &Client{Client: cl, srv: &Server{Server: s}, r: NewNoxRender(r)}
	c.tiles.lightsOutBuf, _ = alloc.Make([]uint32{}, 6)
	t.Cleanup(func() { alloc.FreeSlice(c.tiles.lightsOutBuf) })
	for i := range c.tiles.lightsOutBuf {
		c.tiles.lightsOutBuf[i] = 255
	}
	for x := range c.tiles.nox_arr2_853BC0 {
		for y := range c.tiles.nox_arr2_853BC0[x] {
			c.tiles.nox_arr2_853BC0[x][y] = noxrender.RGB{R: 255 << 16, G: 255 << 16, B: 255 << 16}
		}
	}
	oldClient, oldServer := noxClient, noxServer
	noxClient, noxServer = c, c.srv
	t.Cleanup(func() { noxClient, noxServer = oldClient, oldServer })
	previousFlags := noxflags.GetGame()
	noxflags.ResetGame()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(previousFlags)
	})
	connected := memmap.PtrUint32(0x5D4594, 815764)
	previousConnected := *connected
	*connected = 0
	t.Cleanup(func() { *connected = previousConnected })
	previousFade := legacy.Get_nox_client_fadeObjects_80836()
	legacy.Set_nox_client_fadeObjects_80836(0)
	t.Cleanup(func() { legacy.Set_nox_client_fadeObjects_80836(previousFade) })
	previousWindowClip := legacy.Get_dword_5d4594_3799452()
	legacy.Set_dword_5d4594_3799452(0)
	t.Cleanup(func() { legacy.Set_dword_5d4594_3799452(previousWindowClip) })

	// Use the real native-width server player registry; no player-info service
	// is stubbed. ResetInd initializes its C-owned record explicitly.
	pl := s.Players.ResetInd(0)
	pl.NetCodeVal, pl.Field3680 = 10, 0x10
	pl.ArmorEquip, pl.WeaponEquip = 1, 4
	pl.Colors = server.PlayerUnitColors{Skin: 0x7fff, Hair: 0x7fff, Mustache: 0x7fff, Goatee: 0x7fff, Beard: 0x7fff, UnkColor: 0x7fff}
	previousPlayer := legacy.Get_dword_8531A0_2576()
	legacy.Set_dword_8531A0_2576(pl)
	t.Cleanup(func() { legacy.Set_dword_8531A0_2576(previousPlayer) })
	previousNetCode := legacy.ClientPlayerNetCode()
	legacy.ClientSetPlayerNetCode(10)
	t.Cleanup(func() { legacy.ClientSetPlayerNetCode(previousNetCode) })

	dd, freeDD := alloc.New(client.PlayerDrawData{})
	t.Cleanup(freeDD)
	dd.Size = uint32(unsafe.Sizeof(*dd))
	frames, freeFrames := alloc.Make([]noxrender.ImageHandle{}, 9)
	t.Cleanup(freeFrames)
	for i := range frames {
		img := r.GetBag().ImageByIndex(i)
		if img == nil || img.Type() != 4 {
			t.Fatalf("fixture image %d is not legal type-4 RLE", i)
		}
		frames[i] = img.C()
	}
	panim := &dd.Anim[1]
	panim.Base = client.AnimationVector{Size: uint32(unsafe.Sizeof(panim.Base)), Cnt40: 3, Kind: client.AnimSlave}
	var equipment [3]*client.PlayerEquipAnimation
	for layer := range equipment {
		eani, freeEani := alloc.New(client.PlayerEquipAnimation{})
		t.Cleanup(freeEani)
		eani.Size = uint32(unsafe.Sizeof(*eani))
		for _, dir := range []int{0, 1, 2, 3, 5, 6, 7, 8} {
			eani.Frames[dir] = &frames[layer*3]
		}
		equipment[layer] = eani
	}
	panim.Naked, panim.Armor[0], panim.Weapon[2] = equipment[0], equipment[1], equipment[2]
	dr := cl.Objs.New()
	if dr == nil {
		t.Fatal("real drawable pool exhausted")
	}
	// Do not overwrite the real pool's clientHandle: DrawObjectName's GetExt
	// lookup precedes its empty-name return even though no text is rendered.
	dr.ObjClass, dr.ObjFlags = object.ClassPlayer, 0x01000004
	dr.TypeIDVal, dr.NetCode32 = 0x7ffffffe, 10
	dr.PosVec = image.Pt(20, 20)
	dr.Field_8, dr.Field_9, dr.Field_5 = 20, 20, 1
	dr.DrawData, dr.DrawFuncPtr = unsafe.Pointer(dd), legacy.Get_nox_thing_player_draw()
	dr.AnimInd, dr.AnimDir, dr.AnimFrameSlave = 1, tc.dir, 1
	cl.Objs.Count, cl.Objs.List1 = 1, dr
	if !client.IsRegisteredDrawFunc(dr.DrawFuncPtr) || s.Players.ByID(10) != pl || dr.Client() != cl {
		t.Fatal("fixture did not establish actual callback/player/drawable ownership")
	}
	previousLocal := *memmap.PtrPtr(0x852978, 8)
	c.setClientPlayerUnit(dr)
	t.Cleanup(func() { *memmap.PtrPtr(0x852978, 8) = previousLocal })
	vp := cl.Viewport()
	vp.Screen, vp.World, vp.Size = pix.Rect, pix.Rect, pix.Size()
	for _, ptr := range []unsafe.Pointer{dr.C(), pl.C(), unsafe.Pointer(dd), unsafe.Pointer(&frames[0]), unsafe.Pointer(equipment[0]), vp.C(), draw.C()} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= 0xffffffff {
			t.Fatal("fixture requires native-width C-owned records above 4 GiB")
		}
	}
	for _, handle := range frames {
		if handle == nil || unsafe.Sizeof(uintptr(0)) == 8 && uintptr(handle) <= 0xffffffff {
			t.Fatal("fixture requires actual native-width image handles")
		}
	}
	clear(pix.Pix)
	draw.SetClip(true)
	draw.SetClipRect(pix.Rect)
	draw.SetClipRect2(pix.Rect)
	draw.SetAlpha(tc.alpha)
	draw.SetAlphaEnabled(tc.enabled)
	s.SetFrame(200)
	beforeDD, beforeDR, beforePlayer, beforeVP := *dd, *dr, *pl, *vp
	beforeFrames := slices.Clone(frames)
	beforeEquipment := [3]client.PlayerEquipAnimation{*equipment[0], *equipment[1], *equipment[2]}
	beforeAlpha, beforeEnabled := draw.Alpha(), draw.IsAlphaEnabled()
	logic, other := s.Rand.Logic.Index(), s.Rand.Other.Index()

	// Exactly one registered C bridge call: no pre-selection, alternate Go
	// DrawPlayer invocation, hook, replay, Random case, or expected-frame probe.
	got := dr.CallDraw(vp)
	if cl.Objs.Count != 1 || cl.Objs.List1 != dr {
		t.Fatal("registered player draw changed the real drawable pool")
	}
	if got != 1 {
		t.Fatalf("registered PlayerDraw return=%d, want 1", got)
	}
	if *dd != beforeDD || !slices.Equal(frames, beforeFrames) || *pl != beforePlayer || *vp != beforeVP ||
		s.Frame() != 200 || s.Rand.Logic.Index() != logic || s.Rand.Other.Index() != other {
		t.Fatal("registered PlayerDraw changed shared animation/frames/player/viewport/frame/RNG inputs")
	}
	for layer, eani := range equipment {
		if *eani != beforeEquipment[layer] {
			t.Fatalf("registered PlayerDraw changed equipment animation %d", layer)
		}
	}
	last := frames[4] // armor frame 1, not the production selector's result
	if tc.dir == 5 {
		last = frames[7] // weapon frame 1
	}
	wantDR := beforeDR
	wantDR.Field_0, wantDR.Field_1, wantDR.Field_2 = 1<<16, 10|(2<<16), unsafe.Pointer(last)
	if *dr != wantDR {
		t.Fatal("registered PlayerDraw changed drawable state beyond actual last-image bookkeeping")
	}
	if draw.Alpha() != tc.alpha || draw.IsAlphaEnabled() || draw.Multiply14() || draw.Colorize17() {
		t.Fatalf("actual draw cleanup: alpha=%d enabled=%t multiply=%t colorize=%t", draw.Alpha(), draw.IsAlphaEnabled(), draw.Multiply14(), draw.Colorize17())
	}
	body, anchor := r282ExpectedPlayerDrawWords(tc)
	want := make([]uint16, len(pix.Pix))
	for y := 20; y < 22; y++ {
		for x := 20; x < 22; x++ {
			want[y*64+x] = body
			want[y*64+x+3] = 0x001e // armor's blue frame 1
			want[y*64+x+6] = 0x7800 // weapon's red frame 1
		}
		want[y*64+29] = anchor
	}
	if !slices.Equal(pix.Pix, want) {
		for i, word := range want {
			if pix.Pix[i] != word {
				t.Fatalf("actual player RLE raster (%d,%d)=%04x want=%04x alpha=%d enabled=%t direction=%d", i%64, i/64, pix.Pix[i], word, tc.alpha, tc.enabled, tc.dir)
			}
		}
	}
	receipt := r282PlayerDrawReceipt{
		Case: tc.id, PID: os.Getpid(), ParentPID: os.Getppid(), NativeWordBits: int(unsafe.Sizeof(uintptr(0)) * 8),
		RegisteredCalls: 1, FixedSlaveFrame: 1,
		AlphaBefore: beforeAlpha, AlphaEnabledBefore: beforeEnabled, AlphaAfter: draw.Alpha(), AlphaEnabledAfter: draw.IsAlphaEnabled(),
		BodyWord: pix.Pix[20*64+20], ArmorWord: pix.Pix[20*64+23], WeaponWord: pix.Pix[20*64+26], LastAnchorWord: pix.Pix[20*64+29],
		SameFrameAndOrderPassed: true, WholeSoftwareRaster: true,
	}
	encoded, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%s%s\n", r282PlayerDrawMarker, encoded)
}

// Three literal colored frames for each body/armor/weapon layer. Every image
// has private two-pixel columns plus one common right-hand anchor. All records
// are real type-4 RLE (op1 skips, op2 literal words), not the type-2 path that
// bypasses renderer alpha. No asset or expected output comes from GAME.EXE.
func r282WritePlayerRLEBag(t *testing.T) {
	t.Helper()
	colors := [3][3]uint16{
		{0x7c00, 0x03e0, 0x001f}, // body: red, green, blue
		{0x03e0, 0x001f, 0x7c00}, // armor: green, blue, red
		{0x001f, 0x7c00, 0x03e0}, // weapon: blue, red, green
	}
	var payloads [9][]byte
	var data []byte
	for layer, row := range colors {
		for frame, word := range row {
			payload := make([]byte, 17)
			binary.LittleEndian.PutUint32(payload, 10)
			binary.LittleEndian.PutUint32(payload[4:], 2)
			for y := 0; y < 2; y++ {
				if skip := 3 * layer; skip != 0 {
					payload = append(payload, 1, byte(skip))
				}
				payload = append(payload, 2, 2)
				payload = binary.LittleEndian.AppendUint16(payload, word)
				payload = binary.LittleEndian.AppendUint16(payload, word)
				payload = append(payload, 1, byte(7-3*layer), 2, 1)
				payload = binary.LittleEndian.AppendUint16(payload, word)
			}
			payloads[layer*3+frame] = payload
			data = append(data, payload...)
		}
	}
	idx := make([]byte, 24)
	binary.LittleEndian.PutUint32(idx, 0xfaedbcea)
	binary.LittleEndian.PutUint32(idx[8:], 1)
	for _, word := range []uint32{0, uint32(len(data)), 9} {
		idx = binary.LittleEndian.AppendUint32(idx, word)
	}
	for i, payload := range payloads {
		name := fmt.Sprintf("player-rle-%d", i)
		idx = append(idx, byte(len(name)))
		idx = append(idx, name...)
		idx = append(idx, 4)
		idx = binary.LittleEndian.AppendUint32(idx, uint32(len(payload)))
		idx = binary.LittleEndian.AppendUint32(idx, uint32(len(payload)))
	}
	for name, payload := range map[string][]byte{"video.bag": data, "video.idx": idx} {
		if err := os.WriteFile(name, payload, 0600); err != nil {
			t.Fatal(err)
		}
	}
}
