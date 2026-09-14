package opennox

import (
	"testing"

	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func TestSpellsDurationMoonglowDestroyDispatchUsesNativeRecord(t *testing.T) {
	sp := &spellsDuration{s: &Server{Server: &server.Server{}}}
	record := &server.DurSpell{}
	visual := &server.Object{}
	runtime := sp.moonglowRuntime531A00()
	runtime.StoreVisual(record, visual)
	if got := runtime.LoadVisual(record); got != visual {
		t.Fatalf("stored visual = %p, want %p", got, visual)
	}
	// A nil target makes the destroy callback a no-op. Dispatching to the old
	// C callback would interpret the native-width record at PE32 offsets.
	sp.callDestroy4FEDA0(legacy.Get_sub_531AF0(), record)
	if got := runtime.LoadVisual(record); got != nil {
		t.Fatalf("visual after nil-target destroy = %p, want nil", got)
	}
	runtime.StoreVisual(record, visual)
	sp.Free()
	if got := runtime.LoadVisual(record); got != nil {
		t.Fatalf("visual after duration cleanup = %p, want nil", got)
	}
}
