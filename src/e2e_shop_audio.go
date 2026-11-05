//go:build !server

package opennox

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/timshannon/go-openal/openal"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/client/audio/ail"
)

type e2eShopAudioRecord struct {
	voice             ail.Sample
	source            openal.Source
	blocks            int
	gain              float32
	queued, processed int
	completed         bool
}

type e2eShopAudioProbeState struct {
	records        []*e2eShopAudioRecord
	failure        string
	id             sound.ID
	submissionOnly bool
}

var e2eShopAudioProbe *e2eShopAudioProbeState

// This probe reads the source and its hardware counters without replacing
// audio, callbacks, clocks, sample data or RNG. It works on a remote client
// without a local HostUnit or server trade session.
func e2eObserveShopAudioPacket(op netmsg.Op, id sound.ID, volume, pan int, submitted bool) {
	probe := e2eShopAudioProbe
	if probe == nil || id != probe.id {
		return
	}
	if op != netmsg.MSG_AUDIO_PLAYER_EVENT || !submitted || volume <= 0 {
		probe.failure = fmt.Sprintf("shop audio packet=%s submitted=%t volume=%d", op, submitted, volume)
		return
	}
	nativeAudioFX.mu.Lock()
	defer nativeAudioFX.mu.Unlock()
	def := &nativeAudioFX.defs[id]
	if probe.submissionOnly {
		if nativeAudioFX.bank == nil || len(nativeAudioFX.voices) == 0 || len(def.sampleIDs) == 0 {
			probe.failure = "sale audio requires the stock effect bank and native voices"
			return
		}
		voice := nativeAudioFX.voices[(nativeAudioFX.next+len(nativeAudioFX.voices)-1)%len(nativeAudioFX.voices)]
		owned := false
		for _, event := range nativeAudioFX.events {
			if event.id == id {
				for _, candidate := range event.voices {
					owned = owned || candidate == voice
				}
			}
		}
		ptr := voice.GetSource()
		if !owned || ptr == nil {
			probe.failure = "sale effect has no owned native OpenAL source"
			return
		}
		source := openal.Source(*ptr)
		gain, queued := source.Getf(openal.AlGain), int(source.BuffersQueued())
		if !(gain > 0) || queued == 0 {
			probe.failure = fmt.Sprintf("sale source gain=%g queued=%d", gain, queued)
			return
		}
		probe.records = append(probe.records, &e2eShopAudioRecord{voice: voice, source: source, gain: gain, queued: queued, completed: true})
		e2eLog.Printf("SHOP SALE AUDIO SUBMITTED: event=%d original-id=%d opcode=%s samples=%q voice=%#x source=%d gain=%g queued=%d wire-volume=%d pan=%d", len(probe.records), id, op, def.sampleIDs, voice, source, gain, queued, volume, pan)
		return
	}
	if nativeAudioFX.bank == nil || len(nativeAudioFX.voices) == 0 || len(def.sampleIDs) != 1 {
		probe.failure = "shop audio requires one stock InventoryPickup sample and a real voice pool"
		return
	}
	entry := nativeAudioFX.bank.entries[strings.ToLower(def.sampleIDs[0])]
	if entry == nil || entry.blockSize == 0 || len(entry.data) == 0 || len(entry.data)%int(entry.blockSize) != 0 {
		probe.failure = "invalid stock InventoryPickup ADPCM metadata"
		return
	}
	blocks := len(entry.data) / int(entry.blockSize)
	// The stock click fits in the four hardware buffers. No buffer has to
	// be recycled before its complete, independently readable drain.
	if blocks != 4 || entry.name != "otrapclk" {
		probe.failure = fmt.Sprintf("unexpected stock InventoryPickup sample=%q blocks=%d", entry.name, blocks)
		return
	}
	voice := nativeAudioFX.voices[(nativeAudioFX.next+len(nativeAudioFX.voices)-1)%len(nativeAudioFX.voices)]
	owned := false
	for _, event := range nativeAudioFX.events {
		if event.id == id {
			for _, candidate := range event.voices {
				owned = owned || candidate == voice
			}
		}
	}
	ptr := voice.GetSource()
	if !owned || ptr == nil {
		probe.failure = "submitted shop click has no owned OpenAL source"
		return
	}
	source := openal.Source(*ptr)
	gain := source.Getf(openal.AlGain)
	if !(gain > 0) {
		probe.failure = fmt.Sprintf("shop click source gain=%g", gain)
		return
	}
	record := &e2eShopAudioRecord{voice: voice, source: source, gain: gain, blocks: blocks}
	probe.records = append(probe.records, record)
	e2eLog.Printf("SHOP AUDIO SUBMITTED: event=%d opcode=%s sample=%s bytes=%d block=%d expected=%d voice=%#x source=%d gain=%g wire-volume=%d pan=%d", len(probe.records), op, entry.name, len(entry.data), entry.blockSize, blocks, voice, source, gain, volume, pan)
}

func e2eObserveShopAudioRecord(record *e2eShopAudioRecord) {
	if record.completed {
		return
	}
	record.queued = int(record.source.BuffersQueued())
	record.processed = int(record.source.BuffersProcessed())
	if record.queued == record.blocks && record.processed == record.blocks && record.source.State() == openal.Stopped {
		record.completed = true
		e2eLog.Printf("SHOP AUDIO DRAINED: voice=%#x source=%d queued=%d processed=%d gain=%g", record.voice, record.source, record.queued, record.processed, record.gain)
	}
}

// Called with the FX mutex held, immediately before the chosen voice can be
// reinitialized. Read its old counters before Init clears its hardware queue.
func e2eObserveShopAudioVoiceReuseLocked(state *nativeAudioEffectsState, voice ail.Sample) {
	probe := e2eShopAudioProbe
	if probe == nil || state != &nativeAudioFX {
		return
	}
	for _, record := range probe.records {
		if record.voice != voice || record.completed {
			continue
		}
		e2eObserveShopAudioRecord(record)
		if !record.completed {
			probe.failure = fmt.Sprintf("shop click reused before complete drain: queued=%d processed=%d expected=%d", record.queued, record.processed, record.blocks)
		}
	}
}

func (sc *e2eScenario) CheckShopAudio(mode, count int, name string) {
	if mode == 0 || mode == 2 {
		sc.add(0, name, func() {
			if mode == 2 && e2eShopAudioProbe != nil {
				if e2eShopAudioProbe.failure != "" {
					e2eError(fmt.Errorf("previous shop audio probe failed: %s", e2eShopAudioProbe.failure))
					return
				}
				for _, record := range e2eShopAudioProbe.records {
					if !record.completed {
						e2eError(fmt.Errorf("previous shop audio has not finished"))
						return
					}
				}
				e2eShopAudioProbe = nil
			}
			if os.Getenv("NOX_E2E_AUDIO") != "openal" || audioDev == 0 || !nox_client_isConnected() || e2eShopAudioProbe != nil {
				e2eError(fmt.Errorf("shop audio audit requires one connected client and real OpenAL"))
				return
			}
			e2eShopAudioProbe = &e2eShopAudioProbeState{id: sound.SoundInventoryPickup}
			if mode == 2 {
				e2eShopAudioProbe.id = sound.ID(307)
				e2eShopAudioProbe.submissionOnly = true
				e2eLog.Printf("SHOP SALE AUDIO ARMED: original-id=307 real stock effect and native source submission")
				if noxServer != nil {
					e2eLog.Printf("SHOP SALE AUDIO DESCRIPTORS: original307=%d previous791=%d", noxServer.Audio.Field12(307), noxServer.Audio.Field12(791))
				}
			} else {
				e2eLog.Printf("SHOP AUDIO ARMED: real stock InventoryPickup source gain and complete hardware drain")
			}
		})
		return
	}
	if mode != 1 && mode != 3 || count < 1 {
		sc.add(0, name, func() { e2eError(fmt.Errorf("invalid shop audio mode/count: %d/%d", mode, count)) })
		return
	}
	sc.addWhen(0, name, 10*time.Second, func() bool {
		probe := e2eShopAudioProbe
		if probe == nil {
			e2eError(fmt.Errorf("shop audio probe was not armed"))
			return false
		}
		if probe.submissionOnly != (mode == 3) {
			e2eError(fmt.Errorf("shop audio verification mode does not match its armed probe"))
			return false
		}
		nativeAudioFX.mu.Lock()
		for _, record := range probe.records {
			e2eObserveShopAudioRecord(record)
		}
		nativeAudioFX.mu.Unlock()
		if probe.failure != "" {
			e2eError(fmt.Errorf("%s", probe.failure))
			return false
		}
		if len(probe.records) > count {
			e2eError(fmt.Errorf("shop audio event count=%d, expected=%d", len(probe.records), count))
			return false
		}
		if len(probe.records) != count {
			return false
		}
		for _, record := range probe.records {
			if !record.completed {
				return false
			}
		}
		return true
	}, func() {
		if mode == 3 {
			e2eLog.Printf("SHOP SALE AUDIO VERIFIED: events=%d original-id=307 all-client-submitted=true all-positive-gain=true all-native-buffers-queued=true", count)
		} else {
			e2eLog.Printf("SHOP AUDIO VERIFIED: events=%d all-positive-gain=true all-four-buffers-hardware-consumed=true", count)
		}
	})
}
