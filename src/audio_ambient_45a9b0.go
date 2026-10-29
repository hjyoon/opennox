package opennox

import (
	"image"
	"math"
	"strings"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/client/audio/ail"
	"github.com/opennox/opennox/v1/legacy/timer"
)

func init() {
	legacy.UpdateDrawableAudioLoop45A9B0 = func(source, listener *client.Drawable) {
		if source == nil {
			return
		}
		if listener == nil || noxClient == nil || noxClient.Viewport() == nil {
			nativeAudioFX.stopDrawableLoop(source)
			return
		}
		vp := noxClient.Viewport()
		nativeAudioFX.updateDrawableLoop(source, listener, vp.Screen.Min.X, vp.World.Min.X, nox_win_width)
	}
}

// Original45A9B0 uses DWORD integer coordinates, wrapping IMUL/LEA and the
// CRT's truncating int64 conversion after FSQRT. A negative wrapped square
// yields integer-indefinite INT64_MIN, whose observed low DWORD is zero.
func ambientParameters45A9B0(source, listener image.Point, flags, flags70 uint32, screenX, worldX, width, radius int32) (volume, pan int32) {
	if flags&0x1000000 == 0 || flags70&12 != 0 {
		return
	}
	dx, dy := int32(listener.X)-int32(source.X), int32(listener.Y)-int32(source.Y)
	if dx >= radius || dy >= radius || radius <= 0 {
		return
	}
	square := dx*dx + dy*dy + 1
	distance := int32(0)
	if square >= 0 {
		distance = int32(math.Sqrt(float64(square)))
	}
	if distance >= radius {
		return
	}
	volume = 100 * (radius - distance) / radius
	if volume < 0 {
		volume = 0
	} else if volume > 100 {
		volume = 100
	}
	if half := width / 2; half != 0 {
		pan = 50 * (int32(source.X) - worldX - screenX) / half
	}
	return
}

// These owners live outside C-owned Drawable and its obsolete PE32 channel
// handle. Their voices participate in the same bounded FX pool and event
// limits as UI/network effects, and are detached before drawable memory reuse.
type nativeAmbientLoop struct {
	id          sound.ID
	bank        *nativeAudioBank
	event       *nativeAudioEffectEvent
	remaining   []string
	cycles      uint32
	until       uint64
	volume, pan int
	exhausted   bool
	format      *nativeAudioBankEntry
}

func (s *nativeAudioEffectsState) ambientRandomLocked(max int) int {
	if max <= 1 {
		return 0
	}
	if noxServer != nil && noxServer.Rand.Other != nil {
		return noxServer.Rand.Other.Int(0, max-1)
	}
	n := int(s.sequence % uint32(max))
	s.sequence++
	return n
}

func (s *nativeAudioEffectsState) ambientDelayLocked(def *nativeSoundDef) uint64 {
	if def.behavior&8 == 0 {
		return 0
	}
	delay := def.delayMin
	if def.delayMax > delay {
		delay += uint32(s.ambientRandomLocked(int(uint64(def.delayMax) - uint64(delay) + 1)))
	}
	if delay <= 33 {
		return 0
	}
	return uint64(delay)
}

func (s *nativeAudioEffectsState) nextAmbientSampleLocked(loop *nativeAmbientLoop, def *nativeSoundDef) *nativeAudioBankEntry {
	if len(loop.remaining) == 0 {
		if loop.cycles != 0 && (def.behavior&1 == 0 || def.field15 != 0 && loop.cycles >= uint32(def.field15)) {
			loop.exhausted = true
			return nil
		}
		loop.remaining = append(loop.remaining[:0], def.sampleIDs...)
		if def.behavior&4 == 0 && len(loop.remaining) > 1 {
			pick := 0
			if def.behavior&2 != 0 {
				pick = s.ambientRandomLocked(len(loop.remaining))
			}
			loop.remaining = []string{loop.remaining[pick]}
		}
		loop.cycles++
	}
	for len(loop.remaining) != 0 {
		pick := 0
		if def.behavior&2 != 0 {
			pick = s.ambientRandomLocked(len(loop.remaining))
		}
		name := loop.remaining[pick]
		loop.remaining = append(loop.remaining[:pick], loop.remaining[pick+1:]...)
		entry := s.bank.entries[strings.ToLower(name)]
		if entry != nil && len(entry.data) != 0 && entry.flags&8 != 0 && entry.blockSize != 0 {
			return entry
		}
	}
	return nil
}

func (s *nativeAudioEffectsState) stopAmbientEventLocked(loop *nativeAmbientLoop) {
	if loop.event == nil {
		return
	}
	for i, e := range s.events {
		if e == loop.event {
			for _, voice := range e.voices {
				voice.Init()
			}
			copy(s.events[i:], s.events[i+1:])
			s.events[len(s.events)-1] = nil
			s.events = s.events[:len(s.events)-1]
			break
		}
	}
	loop.event = nil
}

func (s *nativeAudioEffectsState) stopDrawableLoop(source *client.Drawable) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if loop := s.ambient[source]; loop != nil {
		s.stopAmbientEventLocked(loop)
		delete(s.ambient, source)
	}
}

func (s *nativeAudioEffectsState) updateDrawableLoop(source, listener *client.Drawable, screenX, worldX, width int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := sound.ID(source.AudioLoop)
	if int(id) <= 0 || int(id) >= len(s.defs) {
		s.stopAmbientOwnerLocked(source)
		return
	}
	def := &s.defs[id]
	volume, pan := ambientParameters45A9B0(source.PosVec, listener.PosVec, uint32(source.ObjFlags), source.Flags70Val, int32(screenX), int32(worldX), int32(width), int32(def.maxDist))
	if volume == 0 || !def.enabled || s.bank == nil || len(s.voices) == 0 || legacy.Sub_453070() == 0 || !s.hasPlayableEventSamplesLocked(def.sampleIDs) {
		s.stopAmbientOwnerLocked(source)
		return
	}
	loop := s.ambient[source]
	if loop != nil && (loop.id != id || loop.bank != s.bank) {
		s.stopAmbientOwnerLocked(source)
		loop = nil
	}
	if loop == nil {
		loop = &nativeAmbientLoop{id: id, bank: s.bank, until: timer.PlatformTicks() + s.ambientDelayLocked(def)}
		if s.ambient == nil {
			s.ambient = make(map[*client.Drawable]*nativeAmbientLoop)
		}
		s.ambient[source] = loop
	}
	loop.volume, loop.pan = int(volume), int(pan)
	s.pruneAudioEventsLocked()
	s.serviceAmbientLoopLocked(loop, def)
}

func (s *nativeAudioEffectsState) stopAmbientOwnerLocked(source *client.Drawable) {
	if loop := s.ambient[source]; loop != nil {
		s.stopAmbientEventLocked(loop)
		delete(s.ambient, source)
	}
}

func (s *nativeAudioEffectsState) serviceAmbientLoopsLocked() {
	s.pruneAudioEventsLocked()
	for source, loop := range s.ambient {
		if loop.bank != s.bank || legacy.Sub_453070() == 0 {
			s.stopAmbientOwnerLocked(source)
			continue
		}
		s.serviceAmbientLoopLocked(loop, &s.defs[loop.id])
	}
}

func (s *nativeAudioEffectsState) serviceAmbientLoopLocked(loop *nativeAmbientLoop, def *nativeSoundDef) {
	gain := uint32(uint64(163*loop.volume) * uint64(def.volume) >> 14)
	if loop.event != nil {
		event := loop.event
		owned := false
		for _, current := range s.events {
			if current == event {
				owned = true
				break
			}
		}
		if !owned {
			event.voices = nil
		}
		if len(event.voices) != 0 {
			event.volume = gain
			for _, voice := range event.voices {
				voice.SetUserData(gain)
				voice.SetPan(nativeAudioSamplePan(loop.pan))
				setNativeAudioVoiceVolume(voice, gain)
				// Continuous sequences feed the second AIL buffer ahead of completion;
				// a moving emitter changes gain/pan without restarting either buffer.
				if def.behavior&8 == 0 && !loop.exhausted {
					if ready := voice.BufferReady(); ready >= 0 && ready < 2 {
						entry := s.nextAmbientSampleLocked(loop, def)
						if entry != nil {
							if entry.rate == loop.format.rate && entry.flags&1 == loop.format.flags&1 && entry.blockSize == loop.format.blockSize {
								voice.LoadBuffer(uint32(ready), entry.data)
							} else {
								loop.remaining = append([]string{entry.name}, loop.remaining...)
							}
						}
					}
				}
			}
			return
		}
		// A completed or stolen voice is no longer owned by this emitter.
		loop.event = nil
		loop.until = timer.PlatformTicks() + s.ambientDelayLocked(def)
	}
	if loop.exhausted || timer.PlatformTicks() < loop.until {
		return
	}
	// Apply the original per-sound limit first: a louder replacement can free
	// its old voice even when the pool is full. Other background loops yield
	// a saturated pool to active gameplay effects.
	if !s.admitAudioEventLocked(loop.id, def, gain) {
		return
	}
	idle := false
	for _, voice := range s.voices {
		if voice.Status() != 4 {
			idle = true
			break
		}
	}
	if !idle {
		return
	}
	entry := s.nextAmbientSampleLocked(loop, def)
	if entry == nil {
		return
	}
	event := &nativeAudioEffectEvent{id: loop.id, bank: s.bank, volume: gain}
	if !s.playSamplePannedLocked(loop.id, def, entry.name, loop.volume, loop.pan) {
		return
	}
	s.trackAudioEventVoiceLocked(event)
	index := s.audioEventInsertionLocked(loop.id, def, gain)
	s.events = append(s.events, nil)
	copy(s.events[index+1:], s.events[index:])
	s.events[index] = event
	loop.event, loop.format = event, entry
	s.serviceAmbientLoopLocked(loop, def)
}

func setNativeAudioVoiceVolume(voice ail.Sample, gain uint32) {
	fx := (*timer.TimerGroup)(legacy.Get_dword_587000_127004())
	mixed := uint64(gain) * uint64(fx.Timers[0].Current>>16) / VolumeMax
	volume := int(uint64(127) * mixed >> 14)
	if volume > 127 {
		volume = 127
	}
	voice.SetVolume(volume)
}
