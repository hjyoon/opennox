package legacy

import (
	"bytes"
	"encoding/binary"
	"math"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestQuestStageBriefing450980CEntryFullPointerAndSignedScalars(t *testing.T) {
	packet, free := alloc.Make([]byte{0xf0, 0x0d}, 69)
	t.Cleanup(free)
	if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(&packet[0])) <= math.MaxUint32 {
		t.Fatalf("packet=%p", &packet[0])
	}
	old := questStageBriefingCall450980
	t.Cleanup(func() { questStageBriefingCall450980 = old })
	var seen []*byte
	var shows []int32
	questStageBriefingCall450980 = func(p *byte, show int32) int32 {
		seen = append(seen, p)
		shows = append(shows, show)
		return math.MinInt32 + 1
	}
	for _, show := range []int32{0, 256, math.MinInt32, math.MaxInt32} {
		if got := questStageBriefingCEntry450980(&packet[0], show); got != math.MinInt32+1 {
			t.Fatalf("return=%d", got)
		}
	}
	if got := questStageBriefingCEntry450980(nil, -1); got != math.MinInt32+1 {
		t.Fatalf("nil return=%d", got)
	}
	if !reflect.DeepEqual(seen, []*byte{&packet[0], &packet[0], &packet[0], &packet[0], nil}) || !reflect.DeepEqual(shows, []int32{0, 256, math.MinInt32, math.MaxInt32, -1}) {
		t.Fatalf("seen=%v shows=%v", seen, shows)
	}
}

func TestQuestStageBriefing450980NativePacketCStorageAndFallback(t *testing.T) {
	image, freeImage := alloc.Malloc(8)
	text, freeText := alloc.Malloc(8)
	t.Cleanup(freeImage)
	t.Cleanup(freeText)
	imageSlot, textSlot := memmap.PtrPtr(0x5D4594, 832460), memmap.PtrPtr(0x5D4594, 832464)
	stageSlot, stateSlot := memmap.PtrUint32(0x5D4594, 832468), questStartBriefingStatePtr450A30()
	oldImage, oldText, oldStage, oldState := *imageSlot, *textSlot, *stageSlot, *stateSlot
	oldCall, oldLoader := questStageBriefingCall450980, Nox_xxx_gLoadImg
	t.Cleanup(func() {
		*imageSlot, *textSlot, *stageSlot, *stateSlot = oldImage, oldText, oldStage, oldState
		questStageBriefingCall450980, Nox_xxx_gLoadImg = oldCall, oldLoader
	})
	for _, tc := range []struct {
		key                       string
		flags                     uint8
		show                      int32
		nilImage, nilText, mutate bool
	}{
		{key: "Description", flags: 2, show: math.MinInt32},
		{flags: 1},
		{key: "Description", flags: 0xff, show: 256, nilImage: true, nilText: true},
		{key: strings.Repeat("x", 80), flags: 0x80, show: -1},
		{key: "Description", flags: 0, show: 1, mutate: true},
	} {
		packet, free := alloc.Make([]byte{0xf0, 0x0d}, max(69, 38+len(tc.key)))
		t.Cleanup(free)
		copy(packet[5:], "StageImage")
		copy(packet[37:], tc.key)
		binary.LittleEndian.PutUint16(packet[2:], 0x8001)
		packet[4] = tc.flags
		before := bytes.Clone(packet)
		for _, p := range []unsafe.Pointer{unsafe.Pointer(&packet[0]), image, text} {
			if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(p) <= math.MaxUint32 {
				t.Fatalf("pointer=%p", p)
			}
		}
		h := questStageBriefingNativeHooks450980()
		var trace []string
		h.resetParticles = func() {
			trace = append(trace, "reset")
			if *stateSlot != 0 {
				t.Fatalf("initial state=%d", *stateSlot)
			}
		}
		h.hideBook = func(v int32) int32 {
			trace = append(trace, "hide")
			if v != 1 {
				t.Fatalf("hide=%d", v)
			}
			return -1
		}
		h.prepare = func() { trace = append(trace, "prepare") }
		h.loadImage = func(name *byte) unsafe.Pointer {
			trace = append(trace, "image")
			if name != &packet[5] || alloc.GoString(name) != "StageImage" {
				t.Fatalf("image-key=%p", name)
			}
			if tc.nilImage {
				return nil
			}
			return image
		}
		h.loadText = func(key *byte, source string, line int32) unsafe.Pointer {
			trace = append(trace, "text")
			if key != &packet[37] || alloc.GoString(key) != tc.key || source != questStageBriefingSource450980 || line != 1714 {
				t.Fatalf("text-key=%p source=%q line=%d", key, source, line)
			}
			if tc.nilText {
				return nil
			}
			return text
		}
		Nox_xxx_gLoadImg = func(name string) *noxrender.Image {
			trace = append(trace, "fallback")
			if !tc.nilImage || name != "WarriorChapterBegin8" {
				t.Fatalf("fallback=%q", name)
			}
			return nil
		}
		originalSetStage := h.setStage
		h.setStage = func(v uint32) {
			trace = append(trace, "stage")
			originalSetStage(v)
			if tc.mutate {
				packet[4] = 2
			} // Must be observed by the subsequent flag read.
		}
		h.lock = func(screen, mode int32, flags int8) int32 {
			trace = append(trace, "lock")
			if screen != 254 || mode != 1 || flags != 2 || *stageSlot != 0x8001 {
				t.Fatalf("lock=%d/%d/%d stage=%#x", screen, mode, flags, *stageSlot)
			}
			return math.MinInt32 + 1
		}
		questStageBriefingCall450980 = func(p *byte, show int32) int32 { return questStageBriefing450980(p, show, h) }
		*stateSlot = math.MaxUint32
		got := questStageBriefingCEntry450980(&packet[0], tc.show)
		wantReturn, wantState := int32(0), uint32(0)
		if tc.show != 0 {
			wantReturn = math.MinInt32 + 1
		}
		if tc.flags&2 != 0 || tc.mutate {
			wantState = 1
		}
		wantImage, wantText := image, text
		if tc.nilImage {
			wantImage = nil
		}
		if tc.nilText {
			wantText = nil
		}
		if tc.key == "" {
			wantText = memmap.Ptr(0x5D4594 + 832544)
		}
		if tc.mutate {
			before[4] = 2
		}
		if got != wantReturn || *imageSlot != wantImage || *textSlot != wantText || *stateSlot != wantState || *stageSlot != 0x8001 || !bytes.Equal(packet, before) {
			t.Fatalf("return=%d image=%p text=%p state=%d stage=%#x packet=%x", got, *imageSlot, *textSlot, *stateSlot, *stageSlot, packet)
		}
		wantTrace := []string{"reset", "hide", "prepare", "image"}
		if tc.nilImage {
			wantTrace = append(wantTrace, "fallback")
		}
		if tc.key != "" {
			wantTrace = append(wantTrace, "text")
		}
		wantTrace = append(wantTrace, "stage")
		if tc.show != 0 {
			wantTrace = append(wantTrace, "lock")
		}
		if !reflect.DeepEqual(trace, wantTrace) {
			t.Fatalf("trace=%v want=%v", trace, wantTrace)
		}
	}
}
