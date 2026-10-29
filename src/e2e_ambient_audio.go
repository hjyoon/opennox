//go:build !server

package opennox

import (
	"fmt"
	"image"
	"os"
	"strings"
	"time"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/client"
	"github.com/timshannon/go-openal/openal"
)

var e2eAmbientFixture struct {
	sources           []*client.Drawable
	seen              map[*client.Drawable]bool
	queued, processed uint64
}

// Controlled stock emitters use ordinary drawable/audio dispatch, stock AVNT
// samples/delays and the real OpenAL device. No campaign progression is claimed.
func (sc *e2eScenario) CheckAmbientAudio(mode int, name string) {
	if mode == 1 {
		sc.addWhen(0, name, 15*time.Second, func() bool {
			nativeAudioFX.mu.Lock()
			defer nativeAudioFX.mu.Unlock()
			for _, dr := range e2eAmbientFixture.sources {
				loop := nativeAudioFX.ambient[dr]
				if loop == nil || loop.event == nil {
					continue
				}
				for _, voice := range loop.event.voices {
					if voice.Status() == 4 && voice.GetSource() != nil && openal.Source(*voice.GetSource()).BuffersQueued() > 0 {
						e2eAmbientFixture.seen[dr] = true
					}
				}
			}
			stats := audioDev.PlaybackStats()
			return len(e2eAmbientFixture.sources) == 2 && len(e2eAmbientFixture.seen) == 2 && stats.SampleBuffersQueued > e2eAmbientFixture.queued && stats.SampleBuffersProcessed > e2eAmbientFixture.processed
		}, func() {
			for _, dr := range e2eAmbientFixture.sources {
				dr.PosVec = image.Pt(5800, 5800)
			}
			e2eLog.Printf("AMBIENT STOCK PLAYBACK VERIFIED: emitters=2 continuous+random-delay native=%+v; moving controlled emitters outside range", audioDev.PlaybackStats())
		})
		return
	}
	sc.add(0, name, func() {
		if os.Getenv("NOX_E2E_AUDIO") != "openal" || audioDev == 0 || noxClient == nil || noxClient.ClientPlayerUnit() == nil {
			e2eError(fmt.Errorf("ambient audit requires a loaded client and real OpenAL"))
			return
		}
		switch mode {
		case 0:
			if len(e2eAmbientFixture.sources) != 0 {
				e2eError(fmt.Errorf("ambient fixture already owns emitters"))
				return
			}
			count := 0
			var invalid []string
			nativeAudioFX.mu.Lock()
			if nativeAudioFX.bank == nil {
				nativeAudioFX.mu.Unlock()
				e2eError(fmt.Errorf("stock ambient bank is not loaded"))
				return
			}
			noxClient.Things.Each(func(typ *client.ObjectType) {
				if !strings.HasPrefix(typ.ID(), "Amb") {
					return
				}
				count++
				if typ.AudioLoop == 0 || int(typ.AudioLoop) >= len(nativeAudioFX.defs) {
					invalid = append(invalid, typ.ID())
					return
				}
				def := &nativeAudioFX.defs[typ.AudioLoop]
				if !def.enabled || !nativeAudioFX.hasPlayableEventSamplesLocked(def.sampleIDs) {
					invalid = append(invalid, typ.ID())
				}
				for _, sample := range def.sampleIDs {
					entry := nativeAudioFX.bank.entries[strings.ToLower(sample)]
					if entry == nil || len(entry.data) == 0 || entry.flags&8 == 0 || entry.blockSize == 0 || entry.rate == 0 {
						invalid = append(invalid, typ.ID()+"/"+sample)
					}
				}
			})
			nativeAudioFX.mu.Unlock()
			if count != 54 || len(invalid) != 0 {
				e2eError(fmt.Errorf("stock Amb catalog count=%d invalid=%v", count, invalid))
				return
			}
			e2eAmbientFixture.seen = make(map[*client.Drawable]bool)
			stats := audioDev.PlaybackStats()
			e2eAmbientFixture.queued, e2eAmbientFixture.processed = stats.SampleBuffersQueued, stats.SampleBuffersProcessed
			listener := noxClient.ClientPlayerUnit()
			for _, id := range []string{"AmbBrook1", "AmbBird1"} {
				dr := noxClient.Nox_xxx_spriteLoadAdd_45A360_drawable(noxClient.Things.IndByID(id), listener.PosVec)
				if dr == nil {
					e2eError(fmt.Errorf("create stock ambient %s", id))
					return
				}
				dr.ObjFlags |= object.Flags(0x1000000)
				e2eAmbientFixture.sources = append(e2eAmbientFixture.sources, dr)
				e2eLog.Printf("AMBIENT STOCK ARMED: type=%s drawable=%p sound=%d pos=%v", id, dr, dr.AudioLoop, dr.PosVec)
			}
			e2eLog.Printf("AMBIENT STOCK CATALOG VERIFIED: types=%d ordinary AVNT/bank/dispatch, original delays unchanged", count)
		case 2:
			nativeAudioFX.mu.Lock()
			for _, dr := range e2eAmbientFixture.sources {
				if nativeAudioFX.ambient[dr] != nil {
					nativeAudioFX.mu.Unlock()
					e2eError(fmt.Errorf("out-of-range ambient retains owner %p", dr))
					return
				}
			}
			nativeAudioFX.mu.Unlock()
			for _, dr := range e2eAmbientFixture.sources {
				noxClient.Nox_xxx_spriteDeleteStatic_45A4E0_drawable(dr)
			}
			e2eAmbientFixture.sources = nil
			e2eAmbientFixture.seen = nil
			e2eLog.Printf("AMBIENT STOCK CLEANUP VERIFIED: distance stop and ordinary drawable deletion released both owners")
		default:
			e2eError(fmt.Errorf("unknown ambient audit mode %d", mode))
		}
	})
}
