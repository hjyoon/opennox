package opennox

import (
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
)

// UI and network sounds share the live FX gate, original per-event gain,
// sample selection, and native voice pool. The signed wire pan is applied
// before the first buffer is submitted, not after playback has started.
func (s *nativeAudioEffectsState) playPanned(id sound.ID, requestedVolume, pan int) bool {
	ind := int(id)
	if ind <= 0 || ind >= len(s.defs) {
		return false
	}
	if requestedVolume < 0 {
		requestedVolume = 0
	} else if requestedVolume > 100 {
		requestedVolume = 100
	}
	// Like the original effect allocator, admission follows the live FX
	// enabled flag. The configuration scalar is only a startup value.
	if requestedVolume == 0 || legacy.Sub_453070() == 0 {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	def := &s.defs[ind]
	if !def.enabled || len(def.sampleIDs) == 0 || s.bank == nil || len(s.voices) == 0 {
		return false
	}

	samples := def.sampleIDs
	if def.behavior&4 == 0 {
		choice := 0
		if def.behavior&2 != 0 && len(samples) > 1 {
			if noxServer != nil && noxServer.Rand.Other != nil {
				choice = noxServer.Rand.Other.Int(0, len(samples)-1)
			} else {
				choice = int(s.sequence % uint32(len(samples)))
				s.sequence++
			}
		}
		samples = samples[choice : choice+1]
	}
	if !s.hasPlayableEventSamplesLocked(samples) {
		return false
	}
	volume := uint32((uint64(163*requestedVolume) * uint64(def.volume)) >> 14)
	s.pruneAudioEventsLocked()
	if !s.admitAudioEventLocked(id, def, volume) {
		return false
	}
	event := &nativeAudioEffectEvent{id: id, bank: s.bank, volume: volume}
	played := false
	for _, sample := range samples {
		if s.playSamplePannedLocked(id, def, sample, requestedVolume, pan) {
			s.trackAudioEventVoiceLocked(event)
			played = true
		}
	}
	if played {
		index := s.audioEventInsertionLocked(id, def, volume)
		s.events = append(s.events, nil)
		copy(s.events[index+1:], s.events[index:])
		s.events[index] = event
	}
	return played
}
