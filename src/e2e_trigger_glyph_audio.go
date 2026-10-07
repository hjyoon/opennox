package opennox

// e2eTriggerGlyphCastAudioState distinguishes the fixture's queued script
// callback from the NPC's immediate callback without changing game audio.
func e2eTriggerGlyphCastAudioState(mode string, destroyed bool) bool {
	switch mode {
	case "player-script":
		return destroyed
	case "npc-animated":
		return !destroyed
	default:
		return false
	}
}
