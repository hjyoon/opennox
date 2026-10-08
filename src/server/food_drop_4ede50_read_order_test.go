package server

import (
	"math"
	"testing"
)

// GAME.EXE 004EDEBD loads subclass once, after the first sound sentinel and
// after DefaultDrop/decay. The back edge 004EDED6 returns to 004EDEC0, not to
// that load. Material remains a live WORD read at 004EDEC8 on every subclass
// miss. Read-boundary mutations below expose that distinction; they are not
// stock gameplay or a claim that the real sound table has callbacks.
func TestFoodDrop4EDE50SoundScanCachesSubclassOnceAndReadsLiveMaterial(t *testing.T) {
	for _, tc := range []struct {
		name           string
		subClass       uint32
		material       uint16
		defaultClass   uint32
		changeDefault  bool
		postClass      uint32
		changeClass    bool
		postMaterial   uint16
		changeMaterial bool
		afterMaterial  bool
		zeroDefault    bool
		zeroSentinel   bool
		wantSound      uint32
		wantClassReads int
		wantMatReads   int
	}{
		{name: "no-match", wantClassReads: 1, wantMatReads: 4},
		{name: "flesh", material: 1, wantSound: 835, wantClassReads: 1, wantMatReads: 1},
		{name: "apple", subClass: 2, wantSound: 837, wantClassReads: 1, wantMatReads: 1},
		{name: "jug", subClass: 4, wantSound: 833, wantClassReads: 1, wantMatReads: 2},
		{name: "mushroom-high-bits", subClass: 0x80000080, wantSound: 839, wantClassReads: 1, wantMatReads: 3},
		{name: "clear-after-class-load", subClass: 2, changeClass: true, wantSound: 837, wantClassReads: 1, wantMatReads: 1},
		{name: "set-after-class-load", postClass: 2, changeClass: true, wantClassReads: 1, wantMatReads: 4},
		{name: "change-after-row-zero-material", subClass: 2, postClass: 4, changeClass: true, afterMaterial: true, wantSound: 837, wantClassReads: 1, wantMatReads: 1},
		{name: "live-material-after-class", subClass: 2, postMaterial: 1, changeMaterial: true, wantSound: 835, wantClassReads: 1, wantMatReads: 1},
		{name: "default-zero", subClass: 2, zeroDefault: true},
		{name: "first-zero-sentinel", subClass: 2, zeroSentinel: true},
		{name: "post-default-class", defaultClass: 0x80, changeDefault: true, wantSound: 839, wantClassReads: 1, wantMatReads: 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := newFoodDropTestWorld4EDE50()
			w.gameFlag, w.defaultValue = 1, math.MinInt32
			w.subClass["food-a"], w.flagsLow["food-a"] = tc.subClass, tc.material
			if tc.zeroDefault {
				w.defaultValue = 0
			}
			if tc.zeroSentinel {
				w.rules[0].sound = 0
			}
			w.afterDefault = func(*foodDropTestWorld4EDE50) {
				if tc.changeDefault {
					w.subClass["food-a"] = tc.defaultClass
				}
			}
			hooks := w.hooks()
			originalClass, originalMaterial := hooks.loadSubClass, hooks.loadFlagsLow
			classReads, materialReads, audioCalls := 0, 0, 0
			hooks.loadSubClass = func(food string) uint32 {
				value := originalClass(food)
				classReads++
				if classReads == 1 {
					if tc.changeClass && !tc.afterMaterial {
						w.subClass[food] = tc.postClass
					}
					if tc.changeMaterial {
						w.flagsLow[food] = tc.postMaterial
					}
				}
				return value
			}
			hooks.loadFlagsLow = func(food string) uint16 {
				value := originalMaterial(food)
				materialReads++
				if materialReads == 1 && tc.changeClass && tc.afterMaterial {
					w.subClass[food] = tc.postClass
				}
				return value
			}
			hooks.audio = func(id uint32, owner string, kind int32, code uint32) {
				audioCalls++
				if id != tc.wantSound || owner != "owner-a" || kind != 0 || code != 0 {
					t.Errorf("audio = %d/%s/%d/%d, want %d on cached owner", id, owner, kind, code, tc.wantSound)
				}
			}
			if result := foodDrop4EDE50(hooks); result != w.defaultValue {
				t.Errorf("result = %d, want full DefaultDrop result %d", result, w.defaultValue)
			}
			wantAudioCalls := 0
			if tc.wantSound != 0 {
				wantAudioCalls = 1
			}
			if classReads != tc.wantClassReads || materialReads != tc.wantMatReads || audioCalls != wantAudioCalls {
				t.Fatalf("reads class/material and audio = %d/%d/%d, want %d/%d/%d; trace=%v", classReads, materialReads, audioCalls, tc.wantClassReads, tc.wantMatReads, wantAudioCalls, w.events)
			}
		})
	}
}
