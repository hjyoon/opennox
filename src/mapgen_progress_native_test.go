//go:build !server

package opennox

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"testing"
	"unsafe"

	noxcolor "github.com/opennox/libs/color"
	"github.com/opennox/libs/datapath"
	"github.com/opennox/libs/noximage"
	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/client/render"
	"github.com/opennox/opennox/v1/client/seat/headless"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
	"golang.org/x/image/font/gofont/goregular"
)

func mapGenActual201(t *testing.T, mode string) {
	t.Helper()
	const variable = "OPENNOX_TEST_MAPGEN_PROGRESS_R201"
	if raw := os.Getenv(variable); raw != "" {
		lang, err := strconv.Atoi(raw)
		if err != nil {
			t.Fatal(err)
		}
		mapGenActualChild201(t, mode, lang)
		return
	}
	for _, lang := range []int{1, 8} {
		cmd := exec.Command(os.Args[0], "-test.run=^"+t.Name()+"$", "-test.count=1", "-test.v")
		cmd.Env = append(os.Environ(), variable+"="+strconv.Itoa(lang))
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("actual %s/lang%d map generation: %v\n%s", mode, lang, err, output)
		}
		if !bytes.Contains(output, []byte("MAPGEN ACTUAL PASS: "+mode)) {
			t.Fatalf("missing actual progress receipt: %s", output)
		}
		t.Log(string(output))
	}
}

func TestMapGenProgressActualHeadlessLocal4A9D00(t *testing.T)  { mapGenActual201(t, "local") }
func TestMapGenProgressActualHeadlessPacket48EA70(t *testing.T) { mapGenActual201(t, "packet") }

func mapGenActualChild201(t *testing.T, mode string, lang int) {
	legacy.InitBlobData()
	t.Chdir(t.TempDir())
	generatorBag4BC720(t)
	for _, name := range []string{"default", "large", "number", "small"} {
		if err := os.WriteFile(name+".ttf", goregular.TTF, 0600); err != nil {
			t.Fatal(err)
		}
	}
	oldData := datapath.Data()
	datapath.SetData(".")
	t.Cleanup(func() { datapath.SetData(oldData) })
	handles.Init()
	t.Cleanup(handles.Release)
	c, _, d := actualVideoRender200(t)
	labels := []string{"Generating", "Assembling", "Populating"}
	texts := []string{"Generating a map", "Assembling a map", "Populating a map"}
	if lang == 8 {
		texts = []string{"생성중 ⚔", "조립중 ★", "배치중 ☀"}
	}
	entries := make([]strman.Entry, 3)
	for i, key := range labels {
		entries[i] = strman.Entry{ID: strman.ID("guigen.c:" + key), Vals: []strman.Variant{{Str: texts[i]}}}
	}
	j, err := json.Marshal(struct {
		Lang    int            `json:"lang"`
		Entries []strman.Entry `json:"entries"`
	}{lang, entries})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("progress.csf.json", j, 0600); err != nil {
		t.Fatal(err)
	}
	if err := c.Server.Strings().ReadJSON(filepath.Join(".", "progress.csf.json")); err != nil {
		t.Fatal(err)
	}
	if err := c.r.GetFonts().Load(lang); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(c.r.GetFonts().Free)
	if err := c.r.GetBag().ReadVideoBag(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(c.r.GetBag().Free)
	pix := noximage.NewImage16(image.Rect(0, 0, 640, 480))
	c.r.SetPixBuffer(pix)
	seat := headless.New(pix.Size())
	t.Cleanup(func() { _ = seat.Close() })
	window, err := render.New(seat)
	if err != nil {
		t.Fatal(err)
	}
	c.Win = window
	t.Cleanup(window.Close)
	oldPix := noxPixBuffer.img
	noxPixBuffer.img = pix
	t.Cleanup(func() { noxPixBuffer.img = oldPix })
	legacy.Set_nox_win_width(640)
	legacy.Set_nox_win_height(480)
	oldRender := *legacy.VideoModeRenderDataC
	legacy.Set_nox_draw_curDrawData_3799572(d)
	t.Cleanup(func() { *legacy.VideoModeRenderDataC = oldRender })
	*legacy.Get_nox_color_white_2523948_ptr() = 0x7fff
	*memmap.PtrUint32(0x852978, 4) = 0x03e0
	*memmap.PtrUint32(0x5D4594, 815764) = 1
	var sounds [][2]int
	oldSound := legacy.ClientPlaySoundSpecial
	legacy.ClientPlaySoundSpecial = func(id sound.ID, volume int) { sounds = append(sounds, [2]int{int(id), volume}) }
	t.Cleanup(func() { legacy.ClientPlaySoundSpecial = oldSound })
	var images [36]*noxrender.Image
	calls := 0
	oldLoader := legacy.Nox_xxx_gLoadImg
	legacy.Nox_xxx_gLoadImg = func(name string) *noxrender.Image {
		want := fmt.Sprintf("Compass%d", calls+1)
		if calls >= 4 {
			want = fmt.Sprintf("CompassMainArrow%d", calls-3)
		}
		if calls >= 36 || name != want {
			t.Fatalf("compass source name=%q call=%d", name, calls)
		}
		img := c.r.GetBag().ImageByIndex(calls % len(generatorColors4BC720))
		images[calls] = img
		calls++
		return img
	}
	if legacy.Nox_xxx_compassGenStrings_4A9C80() != 1 || calls != 36 {
		t.Fatal("actual compass initialization")
	}
	legacy.Nox_xxx_gLoadImg = oldLoader
	*memmap.PtrUint32(0x5D4594, 1309664) = 0xa55aa55a
	packedBefore := bytes.Clone(memmap.Slice(0x5D4594, 1309516)[:152])
	packet, freePacket := alloc.Make([]byte{}, 35)
	t.Cleanup(func() { freePacket() })
	message := packet[16:19]
	beforeFrame := memmap.Uint32(0x973A20, 496)
	call := func(op byte, sequence uint16) {
		message[0] = op
		binary.LittleEndian.PutUint16(message[1:], sequence)
		before := bytes.Clone(packet)
		if mode == "packet" {
			if got := legacy.Nox_xxx_netOnPacketRecvCli_48EA70_switch(server.HostPlayerIndex, netmsg.Op(op), message); got != 3 {
				t.Fatalf("actual dispatcher bytes=%d", got)
			}
		} else if got := legacy.MapGenClientTextC4A9D00(unsafe.Pointer(&message[0])); got != int16(sequence) {
			t.Fatalf("existing short API=%d", got)
		}
		if !bytes.Equal(packet, before) {
			t.Fatal("actual C entry changed packet/guards")
		}
	}
	reset := func() {
		*d = noxrender.RenderData{}
		d.SetRect3(pix.Rect)
		d.SetClip(true)
		d.SetClipRect(pix.Rect)
		d.SetClipRect2(image.Rect(0, 0, 639, 479))
		d.SetSelectColor(0)
		c.r.ClearScreen(d.BgColor())
	}
	frames := 0
	for stage, op := range []byte{0x9b, 0x9c, 0x9d} {
		for frame := 0; frame < 32; frame++ {
			sequence := uint16(stage*32 + frame + 1)
			*memmap.PtrUint32(0x5D4594, 1309668) = 0xffffffff
			*memmap.PtrUint32(0x5D4594, 1309672) = uint32(frame)
			reset()
			c.r.DrawImageAt(images[frame%4], image.Pt(160, 120))
			c.r.DrawImageAt(images[frame+4], image.Pt(160, 120))
			font := c.r.GetFonts().DefaultFont()
			width := c.r.GetStringSizeWrapped(font, texts[stage], 0).X
			height := c.r.FontHeight(font)
			d.SetTextColor(noxcolor.RGBA5551(0x7fff))
			d.SetColor(noxcolor.RGBA5551(0x03e0))
			c.r.DrawStringHL(font, texts[stage], image.Pt((640-width)/2, 240-(2*height+70)))
			want := slices.Clone(pix.Pix)
			reset()
			audioBefore := len(sounds)
			call(op, sequence)
			if len(sounds) != audioBefore+1 || sounds[len(sounds)-1] != [2]int{897, 50} {
				t.Fatal("original progress sound differs")
			}
			ptr := *memmap.PtrPtr(0x5D4594, 1309660)
			if ptr == nil || (unsafe.Sizeof(uintptr(0)) > 4 && uintptr(ptr) <= 0xffffffff) {
				t.Fatal("actual native UTF-16 pointer missing")
			}
			if !slices.Equal(pix.Pix, want) {
				t.Fatalf("actual original-layout raster mode=%s lang=%d op=%x frame=%d", mode, lang, op, frame)
			}
			if memmap.Uint32(0x5D4594, 1309668) != uint32(sequence) || memmap.Uint32(0x5D4594, 1309672) != uint32((frame+1)%32) {
				t.Fatal("original sequence/frame cycle")
			}
			if !bytes.Equal(memmap.Slice(0x5D4594, 1309516)[:152], packedBefore) {
				t.Fatal("native text pointer corrupted packed compass/neighbor")
			}
			// The actual Go presentation adapter already uploaded this frame.
			presentParticleRegistration4B01B0(t, pix, window, seat, want, 0)
			drawn := slices.Clone(pix.Pix)
			screenCount := memmap.Uint32(0x973A20, 496)
			call(op, sequence)
			if len(sounds) != audioBefore+1 || memmap.Uint32(0x973A20, 496) != screenCount || memmap.Uint32(0x5D4594, 1309672) != uint32((frame+1)%32) || !slices.Equal(pix.Pix, drawn) {
				t.Fatal("duplicate packet repeated sound/draw/frame/present")
			}
			frames++
		}
	}
	if memmap.Uint32(0x973A20, 496) != beforeFrame+uint32(frames) || c.Server.Frame() != 100 || c.Server.Rand.Logic.Index() != 93 || c.Server.Rand.Other.Index() != 17 {
		t.Fatal("actual progress changed game clock/RNG or presentation count")
	}
	t.Logf("MAPGEN ACTUAL PASS: %s/lang%d; %d real localized native CGo/raster/headless frames, duplicates, packed neighbors, three-byte packet and sound requests verified", mode, lang, frames)
}
