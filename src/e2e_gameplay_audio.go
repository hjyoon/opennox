//go:build !server

package opennox

import (
	"fmt"
	"os"

	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/timer"
	"github.com/timshannon/go-openal/openal"
)

var e2eGameplayAudioCheckpoint struct {
	phase     string
	queued    uint64
	processed uint64
}

// CheckGameplayAudio observes an actual, unmuted OpenAL device and loaded
// stock definitions. Mode 0 begins a phase; mode 1 requires new queued and
// consumed FX buffers. The network sound IDs in NOX_DEBUG_AUDIO logs identify
// the gameplay effects within the same phase. No sound or result is injected.
func (sc *e2eScenario) CheckGameplayAudio(mode int, phase, name string) {
	sc.add(0, name, func() {
		if os.Getenv("NOX_E2E_AUDIO") != "openal" || nox_enable_audio == 0 || audioDev == 0 {
			e2eError(fmt.Errorf("gameplay audio audit requires enabled native OpenAL audio"))
			return
		}
		fx := (*timer.TimerGroup)(legacy.Get_dword_587000_127004())
		if legacy.Sub_453070() == 0 || fx == nil || fx.Timers[0].Current>>16 == 0 {
			e2eError(fmt.Errorf("gameplay audio audit requires live, unmuted FX"))
			return
		}
		stats := audioDev.PlaybackStats()
		if !stats.Native || stats.SampleSources != nativeAudioVoiceCount {
			e2eError(fmt.Errorf("gameplay audio audit requires the real sixteen-voice FX pool: %+v", stats))
			return
		}
		if phase == "" || mode < 0 || mode > 1 {
			e2eError(fmt.Errorf("invalid gameplay audio phase=%q mode=%d", phase, mode))
			return
		}
		if mode == 0 {
			report, err := auditNativeAudioDefinitions(&nativeAudioFX)
			if err != nil {
				e2eError(err)
				return
			}
			e2eGameplayAudioCheckpoint.phase = phase
			e2eGameplayAudioCheckpoint.queued = stats.SampleBuffersQueued
			e2eGameplayAudioCheckpoint.processed = stats.SampleBuffersProcessed
			e2eLog.Printf("GAMEPLAY FX CHECKPOINT: phase=%s bank=%d definitions=%d references=%d silent=%d queued=%d processed=%d",
				phase, report.bankSamples, report.enabledDefinitions, report.sampleReferences, report.emptyDefinitions, stats.SampleBuffersQueued, stats.SampleBuffersProcessed)
		} else {
			if e2eGameplayAudioCheckpoint.phase != phase || stats.SampleBuffersQueued <= e2eGameplayAudioCheckpoint.queued || stats.SampleBuffersProcessed <= e2eGameplayAudioCheckpoint.processed {
				e2eError(fmt.Errorf("gameplay FX phase=%s has no newly queued and consumed buffers: checkpoint=%+v actual=%+v", phase, e2eGameplayAudioCheckpoint, stats))
				return
			}
			e2eLog.Printf("GAMEPLAY FX PLAYBACK VERIFIED: phase=%s queued=%d processed=%d delta-queued=%d delta-processed=%d",
				phase, stats.SampleBuffersQueued, stats.SampleBuffersProcessed, stats.SampleBuffersQueued-e2eGameplayAudioCheckpoint.queued, stats.SampleBuffersProcessed-e2eGameplayAudioCheckpoint.processed)
		}
		if err := openal.Err(); err != nil {
			e2eError(fmt.Errorf("gameplay audio OpenAL: %w", err))
		}
	})
}
