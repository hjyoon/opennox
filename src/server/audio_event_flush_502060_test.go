package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

const (
	audioFlushListener502060 = uint64(0x100000001)
	audioFlushEventA502060   = uint64(0x300000003)
	audioFlushEventB502060   = uint64(0x400000004)
	audioFlushEventC502060   = uint64(0x500000005)
	audioFlushEventD502060   = uint64(0x600000006)
	audioFlushEventE502060   = uint64(0x700000007)
)

type audioEventFlushWorld502060 struct {
	bitmap        [32]uint32
	heads         map[uint64]uint64
	limits        map[uint64]int32
	percentages   map[uint64]int32
	next          map[uint64]uint64
	trace         []string
	bitmapLoads   []int32
	descriptorIDs []int32
	sends         []uint64
	onSend        func(uint64)
}

func audioFlushDescriptor502060(soundID int32) uint64 {
	return uint64(0x200000002) + uint64(uint32(soundID))
}

func newAudioEventFlushWorld502060() *audioEventFlushWorld502060 {
	return &audioEventFlushWorld502060{
		heads:       make(map[uint64]uint64),
		limits:      make(map[uint64]int32),
		percentages: make(map[uint64]int32),
		next:        make(map[uint64]uint64),
	}
}

func (w *audioEventFlushWorld502060) hooks() audioEventFlushHooks502060[uint64, uint64, uint64] {
	return audioEventFlushHooks502060[uint64, uint64, uint64]{
		loadBitmapWord: func(word int32) uint32 {
			w.trace = append(w.trace, fmt.Sprintf("bitmap:%d", word))
			w.bitmapLoads = append(w.bitmapLoads, word)
			return w.bitmap[word]
		},
		descriptor: func(soundID int32) uint64 {
			w.trace = append(w.trace, fmt.Sprintf("descriptor:%d", soundID))
			w.descriptorIDs = append(w.descriptorIDs, soundID)
			return audioFlushDescriptor502060(soundID)
		},
		loadHead: func(descriptor uint64) uint64 {
			w.trace = append(w.trace, fmt.Sprintf("head:%#x", descriptor))
			return w.heads[descriptor]
		},
		loadLimit: func(descriptor uint64) int32 {
			w.trace = append(w.trace, fmt.Sprintf("limit:%#x", descriptor))
			return w.limits[descriptor]
		},
		loadPercentage: func(event uint64) int32 {
			w.trace = append(w.trace, fmt.Sprintf("percentage:%#x", event))
			return w.percentages[event]
		},
		send: func(listener, event uint64, percentage int32) bool {
			w.trace = append(w.trace, fmt.Sprintf("send:%#x:%#x:%d", listener, event, percentage))
			w.sends = append(w.sends, event)
			if w.onSend != nil {
				w.onSend(event)
			}
			return event != audioFlushEventA502060
		},
		loadNext: func(event uint64) uint64 {
			w.trace = append(w.trace, fmt.Sprintf("next:%#x", event))
			return w.next[event]
		},
	}
}

func (w *audioEventFlushWorld502060) flush() {
	flushAudioEvents502060(audioFlushListener502060, w.hooks())
}

func TestFlushAudioEvents502060AccessOrderAndSignedLimits(t *testing.T) {
	w := newAudioEventFlushWorld502060()
	w.bitmap[0] = 0x0f
	descriptor0 := audioFlushDescriptor502060(0)
	descriptor1 := audioFlushDescriptor502060(1)
	descriptor2 := audioFlushDescriptor502060(2)
	descriptor3 := audioFlushDescriptor502060(3)
	w.limits[descriptor0] = 99
	w.heads[descriptor1] = audioFlushEventA502060
	w.limits[descriptor1] = 0
	w.heads[descriptor2] = audioFlushEventB502060
	w.limits[descriptor2] = math.MinInt32
	w.heads[descriptor3] = audioFlushEventC502060
	w.limits[descriptor3] = 2
	w.percentages[audioFlushEventC502060] = -7
	w.percentages[audioFlushEventD502060] = 8
	w.next[audioFlushEventC502060] = audioFlushEventD502060
	w.next[audioFlushEventD502060] = audioFlushEventE502060

	w.flush()

	want := []string{
		"bitmap:0",
		"descriptor:0",
		fmt.Sprintf("head:%#x", descriptor0),
		fmt.Sprintf("limit:%#x", descriptor0),
		"descriptor:1",
		fmt.Sprintf("head:%#x", descriptor1),
		fmt.Sprintf("limit:%#x", descriptor1),
		"descriptor:2",
		fmt.Sprintf("head:%#x", descriptor2),
		fmt.Sprintf("limit:%#x", descriptor2),
		"descriptor:3",
		fmt.Sprintf("head:%#x", descriptor3),
		fmt.Sprintf("limit:%#x", descriptor3),
		fmt.Sprintf("percentage:%#x", audioFlushEventC502060),
		fmt.Sprintf("send:%#x:%#x:-7", audioFlushListener502060, audioFlushEventC502060),
		fmt.Sprintf("next:%#x", audioFlushEventC502060),
		fmt.Sprintf("percentage:%#x", audioFlushEventD502060),
		fmt.Sprintf("send:%#x:%#x:8", audioFlushListener502060, audioFlushEventD502060),
		fmt.Sprintf("next:%#x", audioFlushEventD502060),
	}
	for word := 1; word < 32; word++ {
		want = append(want, fmt.Sprintf("bitmap:%d", word))
	}
	if !reflect.DeepEqual(w.trace, want) {
		t.Fatalf("trace =\n%v\nwant\n%v", w.trace, want)
	}
	if !reflect.DeepEqual(w.sends, []uint64{audioFlushEventC502060, audioFlushEventD502060}) {
		t.Fatalf("sends = %#x, want events C and D", w.sends)
	}
}

func TestFlushAudioEvents502060CacheAndReloadBoundaries(t *testing.T) {
	w := newAudioEventFlushWorld502060()
	descriptor0 := audioFlushDescriptor502060(0)
	descriptor33 := audioFlushDescriptor502060(33)
	w.bitmap[0] = 1
	w.heads[descriptor0] = audioFlushEventA502060
	w.limits[descriptor0] = 2
	w.percentages[audioFlushEventA502060] = 11
	w.percentages[audioFlushEventB502060] = 12
	w.next[audioFlushEventB502060] = audioFlushEventE502060
	w.heads[descriptor33] = audioFlushEventC502060
	w.limits[descriptor33] = 1
	w.percentages[audioFlushEventC502060] = 13
	w.onSend = func(event uint64) {
		if event != audioFlushEventA502060 {
			return
		}
		w.bitmap[0] |= uint32(1) << 31
		w.bitmap[1] |= uint32(1) << 1
		w.limits[descriptor0] = math.MaxInt32
		w.next[audioFlushEventA502060] = audioFlushEventB502060
	}

	w.flush()

	if !reflect.DeepEqual(w.sends, []uint64{
		audioFlushEventA502060,
		audioFlushEventB502060,
		audioFlushEventC502060,
	}) {
		t.Fatalf("sends = %#x, want live-next A/B and future-word C", w.sends)
	}
	if !reflect.DeepEqual(w.descriptorIDs, []int32{0, 33}) {
		t.Fatalf("descriptor IDs = %v, want cached-word 0 then live future-word 33", w.descriptorIDs)
	}
	wantWords := make([]int32, 32)
	for i := range wantWords {
		wantWords[i] = int32(i)
	}
	if !reflect.DeepEqual(w.bitmapLoads, wantWords) {
		t.Fatalf("bitmap loads = %v, want each of 32 words once", w.bitmapLoads)
	}
}

func TestFlushAudioEvents502060VisitsHighestSound(t *testing.T) {
	w := newAudioEventFlushWorld502060()
	descriptor := audioFlushDescriptor502060(1023)
	w.bitmap[31] = uint32(1) << 31
	w.heads[descriptor] = audioFlushEventE502060
	w.limits[descriptor] = 1
	w.percentages[audioFlushEventE502060] = math.MaxInt32

	w.flush()

	if !reflect.DeepEqual(w.descriptorIDs, []int32{1023}) ||
		!reflect.DeepEqual(w.sends, []uint64{audioFlushEventE502060}) {
		t.Fatalf("descriptor IDs/sends = %v/%#x, want sound 1023/event E", w.descriptorIDs, w.sends)
	}
}
