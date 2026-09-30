//go:build !server

package opennox

import (
	"fmt"
	"os"

	"github.com/timshannon/go-openal/openal"
)

func (sc *e2eScenario) AssertOpenALPlayback(dialog bool, name string) {
	sc.add(0, name, func() {
		if os.Getenv("NOX_E2E_AUDIO") != "openal" || nox_enable_audio == 0 || audioDev == 0 {
			e2eError(fmt.Errorf("OpenAL playback assertion requires NOX_E2E_AUDIO=openal and enabled audio"))
			return
		}
		stats := audioDev.PlaybackStats()
		if !stats.Native || stats.SampleSources != nativeAudioVoiceCount || stats.MusicStreamsOpened == 0 ||
			stats.SampleBuffersQueued == 0 || stats.SampleBuffersProcessed == 0 ||
			stats.StreamBuffersQueued == 0 || stats.StreamBuffersProcessed == 0 {
			e2eError(fmt.Errorf("OpenAL music/effect playback has not reached the hardware: %+v", stats))
			return
		}
		if dialog && (stats.DialogStreamsOpened == 0 || stats.DialogBuffersQueued == 0 || stats.DialogBuffersProcessed == 0) {
			e2eError(fmt.Errorf("no non-empty dialog stream played on the OpenAL device: %+v", stats))
			return
		}
		if stats.SampleBuffersProcessed > stats.SampleBuffersQueued || stats.StreamBuffersProcessed > stats.StreamBuffersQueued ||
			stats.DialogBuffersProcessed > stats.DialogBuffersQueued {
			e2eError(fmt.Errorf("OpenAL playback counters exceed queued buffers: %+v", stats))
			return
		}
		if err := openal.Err(); err != nil {
			e2eError(fmt.Errorf("OpenAL playback error: %w", err))
			return
		}
		e2eLog.Printf("OPENAL PLAYBACK VERIFIED: driver=%#x dialog=%t stats=%+v", audioDev, dialog, stats)
	})
}
