package opennox

import (
	"reflect"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/server"
)

func TestConsumeObjectSpecialFlags527E50OrderAndClears(t *testing.T) {
	state := uint32(0x80000000 | 0x04000000 | 0x00800000 | 0x00020000 | 0x000000ab)
	var calls []string
	got := consumeObjectSpecialFlags527E50(&state, objectSpecialHooks527E50{
		reportAnimation: func() { calls = append(calls, "animation") },
		reportHealth:    func() { calls = append(calls, "health") },
		reportHidden:    func() { calls = append(calls, "hidden") },
		reportXStatus:   func() { calls = append(calls, "status") },
		reportHeight:    func() { calls = append(calls, "height") },
		reportEnchant:   func() { calls = append(calls, "enchant") },
		reportTeamBase:  func() { calls = append(calls, "team") },
		reportNPC:       func() { calls = append(calls, "npc") },
	})
	if !got {
		t.Fatal("pending special state was not consumed")
	}
	if want := []string{"health", "enchant", "npc"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("callback order = %v, want %v", calls, want)
	}
	if state != 0x800000ab {
		t.Fatalf("remaining state = %#08x, want %#08x", state, uint32(0x800000ab))
	}
	if consumeObjectSpecialFlags527E50(&state, objectSpecialHooks527E50{}) {
		t.Fatal("unrelated state was reported as consumed")
	}
}

func TestObjectPacketsNative519410UseNamedFields(t *testing.T) {
	obj := &server.Object{
		TypeInd:    0x1234,
		ObjClass:   object.ClassSimple,
		NetCode:    0x2345,
		PosVec:     types.Ptf(10.5, 11.5),
		Direction1: 0,
	}
	s := &Server{Server: &server.Server{}}
	if got, want := s.simpleObjectPacketNative5188A0(obj), [9]byte{47, 0x45, 0x23, 0x34, 0x12, 10, 0, 12, 0}; got != want {
		t.Fatalf("simple packet = % x, want % x", got, want)
	}
	obj.ObjClass = object.ClassComplex
	if got, want := s.phantomObjectPacketNative5187E0(obj), [11]byte{48, 0x45, 0x23, 0x34, 0x12, 10, 0, 12, 0, 0x40, 0xff}; got != want {
		t.Fatalf("phantom packet = % x, want % x", got, want)
	}
}

func TestComplexObjectVisualStateNative519410(t *testing.T) {
	data := []byte{48, 2, 0, 4, 0, 10, 0, 20, 0, 0x75, 9}
	state, ok := decodeComplexObjectVisualState519410(data)
	if !ok {
		t.Fatal("valid complex-object packet was rejected")
	}
	if state.Code != 2 || state.TypeID != 4 || state.X != 10 || state.Y != 20 || state.Direction != 8 || state.Animation != 5 || state.Frame != 9 {
		t.Fatalf("decoded state = %+v", state)
	}
	dr := &client.Drawable{AnimFrameSlave: 3, AnimInd: 4, AnimStart: 5}
	applyComplexObjectVisualState519410(dr, state, 100)
	if dr.Field_72 != 100 || dr.AnimFrameSlave != 9 || dr.Field_78 != 3 || dr.AnimDir != 8 || dr.AnimInd != 5 || dr.AnimStart != 100 {
		t.Fatalf("drawable state was not applied through native fields: %+v", dr)
	}
	if _, ok := decodeSimpleObjectVisualState519410(data[:8]); ok {
		t.Fatal("short simple-object packet was accepted")
	}
	if _, ok := decodeComplexObjectVisualState519410(data[:10]); ok {
		t.Fatal("short complex-object packet was accepted")
	}
}

func TestObjectEnchantVisualStateNative48EA70(t *testing.T) {
	state, ok := decodeObjectEnchantVisualState48EA70([]byte{90, 0x34, 0x92, 0x00, 0x80, 0x00, 0x00})
	if !ok || state.Code != 0x9234 || state.Buffs != 0x8000 {
		t.Fatalf("decoded enchant state = %+v, ok=%t", state, ok)
	}
	if _, ok := decodeObjectEnchantVisualState48EA70(make([]byte, 6)); ok {
		t.Fatal("short enchant packet was accepted")
	}

	dr := &client.Drawable{Buffs: 1 << server.ENCHANT_LIGHT, LightIntensity: 200}
	if !applyObjectEnchantVisualState48EA70(dr, 0, nil, 0, 37.5) {
		t.Fatal("removed light enchant did not restore the thing intensity")
	}
	if dr.Buffs != 0 || dr.LightIntensity != 37.5 {
		t.Fatalf("drawable enchant/light state = %#x/%v", dr.Buffs, dr.LightIntensity)
	}
	dr.Buffs = 1 << server.ENCHANT_LIGHT
	dr.LightIntensity = 200
	if applyObjectEnchantVisualState48EA70(dr, 0, dr, 8, 10) || dr.LightIntensity != 200 {
		t.Fatal("local item-light override did not preserve enchanted intensity")
	}
}

func TestPlayerMapTracksObjectNative519410CircularList(t *testing.T) {
	want := &server.Object{}
	first := &server.MinimapItem{Field4: &server.Object{}}
	second := &server.MinimapItem{Field4: want}
	first.Field8 = second
	second.Field8 = first
	pl := &server.Player{Field4580: first}
	if !playerMapTracksObjectNative519410(pl, want) {
		t.Fatal("tracked object was not found")
	}
	if playerMapTracksObjectNative519410(pl, &server.Object{}) {
		t.Fatal("untracked object was found")
	}
	if got := playerTrackedObjectCountNative519710(pl); got != 2 {
		t.Fatalf("tracked count = %d, want 2", got)
	}
	if playerTrackedObjectCountNative519710(&server.Player{}) != 0 {
		t.Fatal("empty tracked list has a nonzero count")
	}
}

func TestNetTrackedObjectRefreshDueNative519710(t *testing.T) {
	if netTrackedObjectRefreshDueNative519710(100, 80, 0) {
		t.Fatal("zero tracked objects requested a refresh")
	}
	if !netTrackedObjectRefreshDueNative519710(100, 80, 4) {
		t.Fatal("elapsed refresh window was not detected")
	}
	if netTrackedObjectRefreshDueNative519710(95, 80, 4) {
		t.Fatal("the strict GAME.EXE time comparison became inclusive")
	}
	if !netTrackedObjectRefreshDueNative519710(1, 1, 61) {
		t.Fatal("more than sixty tracked objects did not force a refresh")
	}
}
