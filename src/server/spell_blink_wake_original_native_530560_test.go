package server

import "testing"

func TestSpellTeleportWakeOriginalNative530560(t *testing.T) {
	fixture := blinkOriginalFixtureRead530310(t)
	count := 0
	for _, row := range fixture.Rows {
		if row.Input.Kind != "wake" {
			continue
		}
		count++
		t.Run(row.Input.Name, func(t *testing.T) {
			blinkOriginalCompare530310(t, row.Result, blinkOriginalRun530310(t, row.Input, SpellTeleportCreateWake530560))
		})
	}
	if count != 11 {
		t.Fatal("incomplete original wake helper inputs", count)
	}
}
