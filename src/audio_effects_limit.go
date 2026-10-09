package opennox

import (
	"strings"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/client/audio/ail"
)

// One AUD event may submit several simultaneous samples. Retain its identity
// outside Sample.UserData, which the live FX mixer needs as a uint32 gain.
type nativeAudioEffectEvent struct {
	id     sound.ID
	bank   *nativeAudioBank
	volume uint32
	voices []ail.Sample
}

func (s *nativeAudioEffectsState) hasPlayableEventSamplesLocked(samples []string) bool {
	for _, name := range samples {
		entry := s.bank.entries[strings.ToLower(name)]
		if entry != nil && len(entry.data) != 0 && entry.flags&8 != 0 && entry.blockSize != 0 {
			return true
		}
	}
	return false
}

func (s *nativeAudioEffectsState) pruneAudioEventsLocked() {
	current := make(map[ail.Sample]bool, len(s.voices))
	for _, voice := range s.voices {
		current[voice] = true
	}
	active := s.events[:0]
	for _, event := range s.events {
		if event.bank != s.bank {
			continue // A closed/reopened driver must not retain the old bank.
		}
		voices := event.voices[:0]
		for _, voice := range event.voices {
			if current[voice] && voice.Status() == 4 {
				voices = append(voices, voice)
			}
		}
		event.voices = voices
		if len(voices) != 0 {
			active = append(active, event)
		}
	}
	s.events = active
}

// GAME.EXE 00451BE0 orders each sound's events by relative volume. Within
// one tenth of the definition volume, it keeps the already-active event;
// behavior bit 0x10 instead prefers the new event. field14 is def+56's
// maximum event count; zero means no per-sound limit, not silence.
func (s *nativeAudioEffectsState) audioEventInsertionLocked(id sound.ID, def *nativeSoundDef, volume uint32) int {
	for i, event := range s.events {
		if event.id != id {
			continue
		}
		delta := int64(event.volume) - int64(volume)
		if delta < 0 {
			delta = -delta
		}
		if delta < int64(def.volume/10) {
			if def.behavior&0x10 != 0 {
				return i
			}
		} else if event.volume < volume {
			return i
		}
	}
	return len(s.events)
}

func (s *nativeAudioEffectsState) admitAudioEventLocked(id sound.ID, def *nativeSoundDef, volume uint32) bool {
	if def.field14 == 0 {
		return true
	}
	count, last := 0, -1
	for i, event := range s.events {
		if event.id == id {
			count++
			last = i
		}
	}
	if count < int(def.field14) {
		return true
	}
	if s.audioEventInsertionLocked(id, def, volume) > last {
		return false // Original list tail would be this new event itself.
	}
	for _, voice := range s.events[last].voices {
		voice.Init() // Stop only the displaced event's owned voices/buffers.
	}
	copy(s.events[last:], s.events[last+1:])
	s.events[len(s.events)-1] = nil
	s.events = s.events[:len(s.events)-1]
	return true
}

func (s *nativeAudioEffectsState) trackAudioEventVoiceLocked(event *nativeAudioEffectEvent) {
	// takeVoiceLocked leaves next immediately after the selected voice.
	voice := s.voices[(s.next+len(s.voices)-1)%len(s.voices)]
	remove := func(voices []ail.Sample) []ail.Sample {
		kept := voices[:0]
		for _, old := range voices {
			if old != voice {
				kept = append(kept, old)
			}
		}
		return kept
	}
	for _, old := range s.events {
		old.voices = remove(old.voices) // Global voice-pool stealing.
	}
	event.voices = append(remove(event.voices), voice)
}
