package opennox

import (
	"strings"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/timer"
)

// nativeAudioSamplePan preserves GAME.EXE 00452FA0's signed, truncated
// -50..50 to 0..16384 conversion followed by 0043F060's 0..127 sample scale.
// A centered UI effect retains the previous native sample pan of 63.
func nativeAudioSamplePan(pan int) int {
	if pan < -50 {
		pan = -50
	} else if pan > 50 {
		pan = 50
	}
	return (127 * (pan*8192/50 + 8192)) >> 14
}

func (s *nativeAudioEffectsState) playSamplePannedLocked(id sound.ID, def *nativeSoundDef, sample string, requestedVolume, pan int) bool {
	entry := s.bank.entries[strings.ToLower(sample)]
	if entry == nil || len(entry.data) == 0 {
		if audioEffectsDebug {
			audioEffectsLog.Printf("sound %s references missing sample %q", id, sample)
		}
		return false
	}
	if entry.flags&8 == 0 || entry.blockSize == 0 {
		if audioEffectsDebug {
			audioEffectsLog.Printf("sound %s sample %q is not supported ADPCM", id, sample)
		}
		return false
	}

	voice := s.takeVoiceLocked()
	if voice == 0 {
		return false
	}
	voice.Init()
	format := int32(5)
	if entry.flags&1 != 0 {
		format = 7
	}
	voice.SetType(format, 0)
	voice.SetADPCMBlockSize(entry.blockSize)
	voice.SetPlaybackRate(int(entry.rate))
	voice.SetPan(nativeAudioSamplePan(pan))

	// The original effect service updates the live FX group before mixing
	// each event, then converts that mixed volume to the sample's 0..127 scale.
	// The saved configuration scalar is only the group's startup value.
	fx := (*timer.TimerGroup)(legacy.Get_dword_587000_127004())
	fx.Update()
	eventVolume := uint32((uint64(163*requestedVolume) * uint64(def.volume)) >> 14)
	mixedVolume := uint64(eventVolume) * uint64(fx.Timers[0].Current>>16) / VolumeMax
	volume := int((uint64(127) * mixedVolume) >> 14)
	if volume > 127 {
		volume = 127
	}
	voice.SetVolume(volume)
	ready := voice.BufferReady()
	if ready < 0 {
		return false
	}
	// Retain the original event gain for live mixing, even after a zero gain.
	voice.SetUserData(eventVolume)
	voice.LoadBuffer(uint32(ready), entry.data)
	if audioEffectsDebug {
		audioEffectsLog.Printf("playing %s via %s (%d bytes, %d Hz)", id, entry.name, len(entry.data), entry.rate)
	}
	return true
}
