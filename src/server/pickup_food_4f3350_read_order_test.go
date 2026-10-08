package server

import (
	"math"
	"testing"
)

// GAME.EXE 004F33C0 loads subclass once, after DefaultPickup and the first
// sound sentinel. The back edge 004F33D9 returns to 004F33C3, not to that
// load. Material remains a live WORD read at 004F33CB on every subclass miss.
// Read-boundary mutations below are diagnostic fixtures, not stock gameplay
// or a claim that the real sound table has callbacks.
func TestPickupFood4F3350SoundScanCachesSubclassOnceAndReadsLiveMaterial(t *testing.T) {
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
		{name: "flesh", material: 1, wantSound: 834, wantClassReads: 1, wantMatReads: 1},
		{name: "apple", subClass: 2, wantSound: 836, wantClassReads: 1, wantMatReads: 1},
		{name: "jug", subClass: 4, wantSound: 832, wantClassReads: 1, wantMatReads: 2},
		{name: "mushroom-high-bits", subClass: 0x80000080, wantSound: 838, wantClassReads: 1, wantMatReads: 3},
		{name: "clear-after-class-load", subClass: 2, changeClass: true, wantSound: 836, wantClassReads: 1, wantMatReads: 1},
		{name: "set-after-class-load", postClass: 2, changeClass: true, wantClassReads: 1, wantMatReads: 4},
		{name: "change-after-row-zero-material", subClass: 2, postClass: 4, changeClass: true, afterMaterial: true, wantSound: 836, wantClassReads: 1, wantMatReads: 1},
		{name: "live-material-after-class", subClass: 2, postMaterial: 1, changeMaterial: true, wantSound: 834, wantClassReads: 1, wantMatReads: 1},
		{name: "default-zero", subClass: 2, zeroDefault: true},
		{name: "first-zero-sentinel", subClass: 2, zeroSentinel: true},
		{name: "post-default-class", defaultClass: 0x80, changeDefault: true, wantSound: 838, wantClassReads: 1, wantMatReads: 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := newPickupFoodTestWorld4F3350()
			w.playerState, w.defaultValue = 1, math.MinInt32
			owner := &pickupFoodTestObject4F3350{name: "owner-a"}
			item := &pickupFoodTestObject4F3350{name: "food-a", subClass: tc.subClass, materialLow: tc.material}
			if tc.zeroDefault {
				w.defaultValue = 0
			}
			if tc.zeroSentinel {
				w.rules[0].sound = 0
			}
			w.afterDefault = func(*pickupFoodTestWorld4F3350) {
				if tc.changeDefault {
					item.subClass = tc.defaultClass
				}
			}
			hooks := w.hooks()
			originalClass, originalMaterial := hooks.loadSubClass, hooks.loadMaterialLow
			classReads, materialReads, audioCalls := 0, 0, 0
			hooks.loadSubClass = func(food *pickupFoodTestObject4F3350) uint32 {
				value := originalClass(food)
				classReads++
				if classReads == 1 {
					if tc.changeClass && !tc.afterMaterial {
						food.subClass = tc.postClass
					}
					if tc.changeMaterial {
						food.materialLow = tc.postMaterial
					}
				}
				return value
			}
			hooks.loadMaterialLow = func(food *pickupFoodTestObject4F3350) uint16 {
				value := originalMaterial(food)
				materialReads++
				if materialReads == 1 && tc.changeClass && tc.afterMaterial {
					food.subClass = tc.postClass
				}
				return value
			}
			hooks.audio = func(id uint32, audioOwner *pickupFoodTestObject4F3350, kind int32, code uint32) {
				audioCalls++
				if id != tc.wantSound || audioOwner != owner || kind != 0 || code != 0 {
					t.Errorf("audio = %d/%p/%d/%d, want %d on cached owner %p", id, audioOwner, kind, code, tc.wantSound, owner)
				}
			}
			if result := pickupFood4F3350(owner, item, math.MinInt32, math.MaxInt32, hooks); result != w.defaultValue {
				t.Errorf("result = %d, want full DefaultPickup result %d", result, w.defaultValue)
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
