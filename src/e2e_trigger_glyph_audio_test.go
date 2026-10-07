package opennox

import "testing"

func TestE2ETriggerGlyphCastAudioState(t *testing.T) {
	for _, tc := range []struct {
		name, mode string
		destroyed  bool
		want       bool
	}{
		{"queued_script_after_death", "player-script", true, true},
		{"queued_script_before_death", "player-script", false, false},
		{"immediate_npc_before_death", "npc-animated", false, true},
		{"immediate_npc_after_death", "npc-animated", true, false},
		{"empty_mode_live", "", false, false},
		{"empty_mode_destroyed", "", true, false},
		{"unknown_mode_live", "player-script-typo", false, false},
		{"unknown_mode_destroyed", "player-script-typo", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := e2eTriggerGlyphCastAudioState(tc.mode, tc.destroyed); got != tc.want {
				t.Fatalf("mode=%q destroyed=%t: got=%t want=%t", tc.mode, tc.destroyed, got, tc.want)
			}
		})
	}
}
