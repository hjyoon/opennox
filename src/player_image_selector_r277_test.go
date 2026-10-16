//go:build !server

package opennox

import (
	"fmt"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/prand"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Test the current Go player selector with the shared original frame helper's
// signed-return contract. Original player-owner identity/remap equivalence is
// unproven; these remaps/clamps are regression controls, not a binary oracle.
// C-owned opaque image tokens are never decoded or rendered. Raw animation
// zero/chat-bubble ownership and personal Host Load opacity are outside scope.

var playerSelectorDirsR277 = [...]byte{0, 1, 2, 3, 5, 6, 7, 8}

type playerSelectorFixtureR277 struct {
	c           *Client
	dr          *client.Drawable
	dd          *client.PlayerDrawData
	equip       []client.PlayerEquipAnimation
	images      []noxrender.ImageHandle
	tokens      []uint32
	nakedFrames [server.PlayerAnimCnt][9][]noxrender.ImageHandle
}

type playerSelectorSnapshotR277 struct {
	draw   client.Drawable
	data   client.PlayerDrawData
	equip  []client.PlayerEquipAnimation
	images []noxrender.ImageHandle
	tokens []uint32
	frame  uint32
}

func newPlayerSelectorFixtureR277(t *testing.T, selected int, count, delay uint16, kind client.AnimKind, seed int) *playerSelectorFixtureR277 {
	t.Helper()
	if selected < 0 || selected >= server.PlayerAnimCnt || count > 255 || delay > 255 {
		t.Fatal("fixture metadata must use a legal player animation and U8 count/delay")
	}
	s := &server.Server{}
	s.Rand.Other, s.Rand.Logic = prand.New(seed), prand.New(233)
	f := &playerSelectorFixtureR277{c: &Client{Client: &client.Client{Server: s}, srv: &Server{Server: s}}}
	var free func()
	f.dr, free = alloc.New(client.Drawable{})
	t.Cleanup(free)
	f.dd, free = alloc.New(client.PlayerDrawData{})
	t.Cleanup(free)
	f.equip, free = alloc.Make([]client.PlayerEquipAnimation{}, server.PlayerAnimCnt*3)
	t.Cleanup(free)
	f.dd.Size = uint32(unsafe.Sizeof(*f.dd))
	f.dr.DrawData = unsafe.Pointer(f.dd)
	for i := range f.equip {
		f.equip[i].Size = uint32(unsafe.Sizeof(f.equip[i]))
	}
	// Every animation has typed base metadata. Each nonempty animation has
	// independent base, Naked, Armor[0], and Weapon[0] arrays in all eight legal
	// directions; absent equipment slots and unused direction 4 remain nil.
	total := 0
	for ai := range f.dd.Anim {
		a := &f.dd.Anim[ai]
		a.Base.Size = uint32(unsafe.Sizeof(a.Base))
		a.Base.Cnt40, a.Base.Val42, a.Base.Kind = 3, 1, client.AnimLoop
		if ai == selected {
			a.Base.Cnt40, a.Base.Val42, a.Base.Kind = count, delay, kind
		}
		total += int(a.Base.Cnt40) * len(playerSelectorDirsR277) * 4
	}
	f.images, free = alloc.Make([]noxrender.ImageHandle{}, total)
	t.Cleanup(free)
	f.tokens, free = alloc.Make([]uint32{}, total)
	t.Cleanup(free)
	pos := 0
	for ai := range f.dd.Anim {
		a := &f.dd.Anim[ai]
		n := int(a.Base.Cnt40)
		if n == 0 {
			// An empty selected animation intentionally has no Naked record.
			continue
		}
		a.Naked, a.Armor[0], a.Weapon[0] = &f.equip[ai*3], &f.equip[ai*3+1], &f.equip[ai*3+2]
		for layer := 0; layer < 4; layer++ {
			for _, dir := range playerSelectorDirsR277 {
				arr := f.images[pos : pos+n]
				for fi := range arr {
					f.tokens[pos+fi] = uint32(pos + fi + 1)
					arr[fi] = noxrender.ImageHandle(unsafe.Pointer(&f.tokens[pos+fi]))
				}
				if layer == 0 {
					a.Base.Frames[dir] = &arr[0]
				} else {
					f.equip[ai*3+layer-1].Frames[dir] = &arr[0]
					if layer == 1 {
						f.nakedFrames[ai][dir] = arr
					}
				}
				pos += n
			}
		}
	}
	if pos != total {
		t.Fatalf("typed native fixture initialization covered %d/%d handles", pos, total)
	}
	return f
}

func (f *playerSelectorFixtureR277) snapshot() playerSelectorSnapshotR277 {
	return playerSelectorSnapshotR277{
		draw: *f.dr, data: *f.dd, equip: slices.Clone(f.equip),
		images: slices.Clone(f.images), tokens: slices.Clone(f.tokens), frame: f.c.Server.Frame(),
	}
}

func (f *playerSelectorFixtureR277) checkSelection(t *testing.T, mapped, frame int, otherOracle *prand.Rand) {
	t.Helper()
	before := f.snapshot()
	wantOther, beforeLogic := *f.c.Server.Rand.Other, *f.c.Server.Rand.Logic
	if otherOracle != nil {
		wantOther = *otherOracle
	}
	img, panim, got := f.c.sub_4B8FA0(f.dr)
	if frame < 0 {
		if img != nil || panim != nil || got != -1 {
			t.Errorf("negative/empty selection returned img=%p panim=%p frame=%d, want nil/nil/-1", img, panim, got)
		}
	} else {
		wantImage := f.nakedFrames[mapped][before.draw.AnimDir][frame]
		if img != wantImage || panim != &f.dd.Anim[mapped] || got != frame {
			t.Errorf("raw=%d dir=%d: img=%p panim=%p frame=%d, want exact Naked img=%p panim=%p frame=%d", before.draw.AnimInd, before.draw.AnimDir, img, panim, got, wantImage, &f.dd.Anim[mapped], frame)
		}
	}
	if *f.dr != before.draw || *f.dd != before.data || !slices.Equal(f.equip, before.equip) || !slices.Equal(f.images, before.images) || !slices.Equal(f.tokens, before.tokens) || f.c.Server.Frame() != before.frame {
		t.Error("player selection changed drawable/raw AnimInd, typed draw/equip data, frame arrays, handle tokens, or server frame")
	}
	if *f.c.Server.Rand.Other != wantOther || *f.c.Server.Rand.Logic != beforeLogic {
		t.Errorf("unexpected RNG state: Other index=%d Logic index=%d; want Other index=%d and unchanged Logic", f.c.Server.Rand.Other.Index(), f.c.Server.Rand.Logic.Index(), wantOther.Index())
	}
}

func TestPlayerImageSelectorDirectionsAndRemapsR277(t *testing.T) {
	for _, spec := range []struct {
		name             string
		raw, buffs       uint32
		mapped           int
		count, delay     uint16
		kind             client.AnimKind
		now, start, code uint32
		want             int
	}{
		{"raw-one", 1, 0, 1, 3, 1, client.AnimOneShot, 102, 100, 0, 1},
		{"raw-four-plain", 4, 0, 4, 3, 1, client.AnimOneShot, 104, 100, 0, 2},
		{"raw-four-sneak", 4, uint32(1) << uint32(server.ENCHANT_SNEAK), 53, 3, 1, client.AnimOneShot, 100, 100, 0, 0},
		{"raise-shield-stays-forty", 40, 0, 40, 3, 0, client.AnimOneShot, 3, 0, 0, 2},
		{"count-one-maximum-delay", 1, 0, 1, 1, 255, client.AnimOneShot, 100, 0, 0, 0},
		{"U8-maximum-count-delay", 1, 0, 1, 255, 255, client.AnimOneShot, 512, 0, 0, 2},
		{"elapsed-wrap-before-divide", 1, 0, 1, 255, 1, client.AnimOneShot, 0x10, 0xfffffff0, 0, 16},
		{"loop-sum-wrap-before-divide", 1, 0, 1, 7, 0, client.AnimLoop, 0xfffffff0, 0, 0x20, 2},
		{"unsigned-divide-before-signed-result", 1, 0, 1, 3, 1, client.AnimOneShot, 0xffffffff, 0, 0, 2},
		{"loop-maximum-delay", 1, 0, 1, 3, 255, client.AnimLoop, 0xffffffff, 0, 0, 0},
		{"default-kind-zero-frame", 1, 0, 1, 3, 255, client.AnimLoopAndFade, 0xffffffff, 0, 0, 0},
	} {
		for _, dir := range playerSelectorDirsR277 {
			t.Run(fmt.Sprintf("%s-dir-%d", spec.name, dir), func(t *testing.T) {
				f := newPlayerSelectorFixtureR277(t, spec.mapped, spec.count, spec.delay, spec.kind, 17)
				f.dr.AnimInd, f.dr.AnimDir, f.dr.Buffs = spec.raw, dir, spec.buffs
				f.dr.AnimStart, f.dr.NetCode32 = spec.start, spec.code
				f.c.Server.SetFrame(spec.now)
				f.checkSelection(t, spec.mapped, spec.want, nil)
			})
		}
	}
}

func TestPlayerImageSelectorSignedFramesR277(t *testing.T) {
	// The existing Go player's negative guard must receive the shared helper's
	// int32 result. This does not certify the unrecognized original player owner.
	for _, spec := range []struct {
		name        string
		kind        client.AnimKind
		now, start  uint32
		code, slave uint32
	}{
		{"one-high-bit", client.AnimOneShot, 0x80000000, 0, 0, 0},
		{"one-minus-one", client.AnimOneShot, 0xffffffff, 0, 0, 0},
		{"one-future-start", client.AnimOneShot, 0, 1, 0, 0},
		{"loop-high-bit", client.AnimLoop, 0x80000000, 0, 0, 0},
		{"loop-add-enters-high-bit", client.AnimLoop, 0x7fffffff, 0, 1, 0},
		{"loop-minus-one", client.AnimLoop, 0xffffffff, 0, 0, 0},
		{"slave-minimum", client.AnimSlave, 100, 0, 0, 0x80000000},
		{"slave-minus-two", client.AnimSlave, 100, 0, 0, 0xfffffffe},
		{"slave-minus-one", client.AnimSlave, 100, 0, 0, 0xffffffff},
	} {
		t.Run(spec.name, func(t *testing.T) {
			f := newPlayerSelectorFixtureR277(t, 1, 3, 0, spec.kind, 17)
			f.dr.AnimInd, f.dr.AnimDir = 1, 8
			f.dr.AnimStart, f.dr.NetCode32, f.dr.AnimFrameSlave = spec.start, spec.code, spec.slave
			f.c.Server.SetFrame(spec.now)
			f.checkSelection(t, -1, -1, nil)
		})
	}
}

func TestPlayerImageSelectorSlaveClampR277(t *testing.T) {
	// The last-frame clamp belongs to the player selector, not to the shared
	// Slave helper and not automatically to generic-vector/NPC consumers.
	for _, spec := range []struct {
		slave uint32
		want  int
	}{{0, 0}, {1, 1}, {2, 2}, {3, 2}, {255, 2}, {0x7fffffff, 2}} {
		t.Run(fmt.Sprintf("slave-%08x", spec.slave), func(t *testing.T) {
			f := newPlayerSelectorFixtureR277(t, 1, 3, 255, client.AnimSlave, 17)
			f.dr.AnimInd, f.dr.AnimDir, f.dr.AnimFrameSlave = 1, 7, spec.slave
			f.c.Server.SetFrame(0xffffffff)
			f.checkSelection(t, 1, spec.want, nil)
		})
	}
}

func TestPlayerImageSelectorZeroCountR277(t *testing.T) {
	for _, kind := range []client.AnimKind{client.AnimOneShot, client.AnimLoop, client.AnimRandom, client.AnimSlave} {
		t.Run(fmt.Sprintf("kind-%d", kind), func(t *testing.T) {
			f := newPlayerSelectorFixtureR277(t, 1, 0, 255, kind, 17)
			f.dr.AnimInd, f.dr.AnimDir, f.dr.AnimFrameSlave = 1, 8, 0xffffffff
			f.c.Server.SetFrame(0xffffffff)
			if f.dd.Anim[1].Naked != nil {
				t.Fatal("empty-animation fixture accidentally supplied a Naked record")
			}
			f.checkSelection(t, -1, -1, nil)
		})
	}
}

func TestPlayerImageSelectorRandomR277(t *testing.T) {
	for _, count := range []uint16{1, 3, 255} {
		for _, seed := range []int{0, 17, 233} {
			t.Run(fmt.Sprintf("count-%d-seed-%d", count, seed), func(t *testing.T) {
				f := newPlayerSelectorFixtureR277(t, 53, count, 255, client.AnimRandom, seed)
				f.dr.AnimInd, f.dr.AnimDir = 4, 6
				f.dr.Buffs = uint32(1) << uint32(server.ENCHANT_SNEAK)
				f.dr.AnimStart, f.dr.NetCode32, f.dr.AnimFrameSlave = 0xffffffff, 0x80000000, 0xffffffff
				f.c.Server.SetFrame(0xffffffff)
				oracle := prand.New(seed)
				want := oracle.Int(0, int(count)-1) // Independent instance, exactly one draw.
				f.checkSelection(t, 53, want, oracle)
			})
		}
	}
}
