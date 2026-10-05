//go:build !server

package opennox

import (
	"fmt"
	"os"

	"github.com/timshannon/go-openal/openal"
)

func e2eNPCDialogAudioStats479B00() (e2eNPCDialogPlayback479B00, error) {
	if os.Getenv("NOX_E2E_AUDIO") != "openal" || nox_enable_audio == 0 || audioDev == 0 {
		return e2eNPCDialogPlayback479B00{}, fmt.Errorf("NPC Repeat observation requires enabled native OpenAL audio")
	}
	stats := audioDev.PlaybackStats()
	if !stats.Native {
		return e2eNPCDialogPlayback479B00{}, fmt.Errorf("NPC Repeat cannot use mock playback counters")
	}
	if err := openal.Err(); err != nil {
		return e2eNPCDialogPlayback479B00{}, fmt.Errorf("NPC Repeat OpenAL: %w", err)
	}
	return e2eNPCDialogPlayback479B00{stats.DialogStreamsOpened, stats.DialogBuffersQueued, stats.DialogBuffersProcessed}, nil
}
