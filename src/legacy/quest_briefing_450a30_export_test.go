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

func TestQuestStartBriefing450A30CEntryDelegation(t *testing.T) {
	packet, freePacket := alloc.Make([]byte{0xf0, 0x0e}, 69)
	t.Cleanup(freePacket)
	if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(&packet[0])) <= math.MaxUint32 {
		t.Fatalf("packet=%p, want above 4 GiB", &packet[0])
	}
	old := questStartBriefingCall450A30
	t.Cleanup(func() { questStartBriefingCall450A30 = old })
	var packets []*byte
	var shows []int32
	questStartBriefingCall450A30 = func(packet *byte, show int32) int32 {
		packets, shows = append(packets, packet), append(shows, show)
		return math.MinInt32 + 1
	}
	for _, show := range []int32{0, 256, math.MinInt32, math.MaxInt32} {
		if got := questStartBriefingCEntry450A30(&packet[0], show); got != math.MinInt32+1 {
			t.Fatalf("return=%d", got)
		}
	}
	if got := questStartBriefingCEntry450A30(nil, -1); got != math.MinInt32+1 {
		t.Fatalf("nil return=%d", got)
	}
	if !reflect.DeepEqual(packets, []*byte{&packet[0], &packet[0], &packet[0], &packet[0], nil}) || !reflect.DeepEqual(shows, []int32{0, 256, math.MinInt32, math.MaxInt32, -1}) {
		t.Fatalf("packets=%v shows=%v", packets, shows)
	}
}

func TestQuestStartBriefing450A30NativePacketAndCStorage(t *testing.T) {
	image, freeImage := alloc.Malloc(8)
	text, freeText := alloc.Malloc(8)
	t.Cleanup(freeImage)
	t.Cleanup(freeText)
	imageSlot, textSlot := memmap.PtrPtr(0x5D4594, 832460), memmap.PtrPtr(0x5D4594, 832464)
	stageSlot := (*uint32)(memmap.Ptr(0x5D4594 + 832468))
	stateSlot := questStartBriefingStatePtr450A30()
	oldImage, oldText, oldStage, oldState := *imageSlot, *textSlot, *stageSlot, *stateSlot
	oldCall, oldLoader := questStartBriefingCall450A30, Nox_xxx_gLoadImg
	t.Cleanup(func() {
		*imageSlot, *textSlot, *stageSlot, *stateSlot = oldImage, oldText, oldStage, oldState
		questStartBriefingCall450A30, Nox_xxx_gLoadImg = oldCall, oldLoader
	})
	for _, tc := range []struct {
		key               string
		nilImage, nilText bool
		show              int32
	}{
		{key: "Description", show: math.MinInt32},
		{show: 0},
		{key: "Description", nilImage: true, nilText: true, show: 256},
		// A valid NUL-terminated allocation may exceed the wire field. Do not
		// replace PE32 strlen with a 32-byte cap inside the original body.
		{key: strings.Repeat("x", 80), show: -1},
	} {
		packet, freePacket := alloc.Make([]byte{0xf0, 0x0e}, max(69, 37+len(tc.key)+1))
		t.Cleanup(freePacket)
		copy(packet[5:], "StartImage")
		copy(packet[37:], tc.key)
		binary.LittleEndian.PutUint16(packet[2:], 0x8001)
		packet[4] = 0xff // This body must not use subtype-0D's bit-2 state rule.
		before := bytes.Clone(packet)
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(&packet[0]), image, text} {
			if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(ptr) <= math.MaxUint32 {
				t.Fatalf("pointer=%p, want above 4 GiB", ptr)
			}
		}
		h := questStartBriefingNativeHooks450A30()
		var trace []string
		h.resetParticles = func() {
			trace = append(trace, "reset")
			if *stateSlot != 0 {
				t.Fatalf("state=%#x before reset", *stateSlot)
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
			if name != &packet[5] || alloc.GoString(name) != "StartImage" {
				t.Fatalf("image name=%p want=%p", name, &packet[5])
			}
			if tc.nilImage {
				return nil
			}
			return image
		}
		h.loadText = func(key *byte, source string, line int32) unsafe.Pointer {
			trace = append(trace, "text")
			if key != &packet[37] || alloc.GoString(key) != tc.key || source != questStartBriefingSource450A30 || line != 1756 {
				t.Fatalf("text key=%p want=%p source=%q line=%d", key, &packet[37], source, line)
			}
			if tc.nilText {
				return nil
			}
			return text
		}
		Nox_xxx_gLoadImg = func(name string) *noxrender.Image {
			trace = append(trace, "fallback-image")
			if !tc.nilImage || name != "WarriorChapterBegin8" {
				t.Fatalf("unexpected fallback=%q", name)
			}
			return nil
		}
		h.lock = func(screen, mode int32, flags int8) int32 {
			trace = append(trace, "lock")
			if screen != 254 || mode != 1 || flags != 4 || *stageSlot != 0x8001 {
				t.Fatalf("lock=%d/%d/%d stage=%#x", screen, mode, flags, *stageSlot)
			}
			return math.MinInt32 + 1
		}
		questStartBriefingCall450A30 = func(packet *byte, show int32) int32 { return questStartBriefing450A30(packet, show, h) }
		*stateSlot = 1
		got := questStartBriefingCEntry450A30(&packet[0], tc.show)
		wantReturn := int32(0)
		if tc.show != 0 {
			wantReturn = math.MinInt32 + 1
		}
		wantImage, wantText := image, text
		if tc.nilImage {
			wantImage = nil
		}
		if tc.nilText {
			wantText = nil
		}
		if tc.key == "" {
			wantText = memmap.Ptr(0x5D4594 + 832548)
		}
		if got != wantReturn || *imageSlot != wantImage || *textSlot != wantText || *stateSlot != 0 || *stageSlot != 0x8001 || !bytes.Equal(packet, before) {
			t.Fatalf("return=%d image=%p text=%p state=%#x stage=%#x packet=%x", got, *imageSlot, *textSlot, *stateSlot, *stageSlot, packet)
		}
		wantTrace := []string{"reset", "hide", "prepare", "image"}
		if tc.nilImage {
			wantTrace = append(wantTrace, "fallback-image")
		}
		if tc.key != "" {
			wantTrace = append(wantTrace, "text")
		}
		if tc.show != 0 {
			wantTrace = append(wantTrace, "lock")
		}
		if !reflect.DeepEqual(trace, wantTrace) {
			t.Fatalf("trace=%v want=%v", trace, wantTrace)
		}
	}
}
