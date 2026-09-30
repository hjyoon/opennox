package server

import "testing"

func TestPlayerMinimapTrackingNativePointers(t *testing.T) {
	want := &Object{}
	first := &MinimapItem{Field4: &Object{}}
	second := &MinimapItem{Field4: want}
	third := &MinimapItem{Field4: &Object{}}
	first.Field8 = second
	second.Field8 = third
	third.Field8 = first

	pl := &Player{Field4580: first}
	if !pl.MinimapTracks(want) {
		t.Fatal("tracked object was not found")
	}
	if pl.MinimapTracks(&Object{}) {
		t.Fatal("untracked object was found")
	}
	if got := pl.MinimapTrackCount(); got != 3 {
		t.Fatalf("tracked object count = %d, want 3", got)
	}
}

func TestPlayerMinimapTrackingGuardsAndBrokenTail(t *testing.T) {
	var nilPlayer *Player
	if nilPlayer.MinimapTracks(&Object{}) || nilPlayer.MinimapTrackCount() != 0 {
		t.Fatal("nil player reported tracked objects")
	}
	if (&Player{}).MinimapTracks(nil) {
		t.Fatal("nil object was reported as tracked")
	}

	want := &Object{}
	first := &MinimapItem{Field4: want}
	second := &MinimapItem{Field4: &Object{}}
	first.Field8 = second
	pl := &Player{Field4580: first}
	if !pl.MinimapTracks(want) {
		t.Fatal("object in nil-terminated list was not found")
	}
	if got := pl.MinimapTrackCount(); got != 2 {
		t.Fatalf("nil-terminated tracked object count = %d, want 2", got)
	}
}
