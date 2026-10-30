//go:build !server

package opennox

import (
	"fmt"
	"os"
	"time"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/timshannon/go-openal/openal"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/client/audio/ail"
	"github.com/opennox/opennox/v1/server"
)

type e2eDoorAudioRecord struct {
	voice             ail.Sample
	source            openal.Source
	gain              float32
	queued, processed int
	completed         bool
}

var e2eDoorAudioProbe *struct {
	expected sound.ID
	received bool
	records  []*e2eDoorAudioRecord
	failure  string
}

var e2eDoorAudioPairs = [5][2]sound.ID{
	{sound.SoundOpenWoodenDoor, sound.SoundCloseWoodenDoor},
	{sound.SoundOpenGate, sound.SoundCloseGate},
	{sound.SoundOpenWoodenGate, sound.SoundCloseWoodenGate},
	{sound.SoundOpenHeavyWoodenDoor, sound.SoundCloseHeavyWoodenDoor},
	{sound.SoundStoneDoorOpen, sound.SoundStoneDoorClose},
}

func e2eObserveDoorAudioPacket(op netmsg.Op, id sound.ID, volume int, submitted bool) {
	probe := e2eDoorAudioProbe
	if probe == nil {
		return
	}
	known := false
	for _, pair := range e2eDoorAudioPairs {
		known = known || id == pair[0] || id == pair[1]
	}
	if !known {
		return
	}
	if probe.received || id != probe.expected || op != netmsg.MSG_AUDIO_EVENT || !submitted || volume <= 0 {
		probe.failure = fmt.Sprintf("door sound=%s expected=%s received=%t op=%s submitted=%t volume=%d", id, probe.expected, probe.received, op, submitted, volume)
		return
	}
	nativeAudioFX.mu.Lock()
	defer nativeAudioFX.mu.Unlock()
	events := 0
	for _, event := range nativeAudioFX.events {
		if event.id != id {
			continue
		}
		events++
		for _, voice := range event.voices {
			ptr := voice.GetSource()
			if ptr == nil {
				probe.failure = "stock door event has no native OpenAL source"
				return
			}
			source := openal.Source(*ptr)
			gain := source.Getf(openal.AlGain)
			if !(gain > 0) {
				probe.failure = fmt.Sprintf("stock door source gain=%g", gain)
				return
			}
			probe.records = append(probe.records, &e2eDoorAudioRecord{voice: voice, source: source, gain: gain})
			e2eLog.Printf("DOOR AUDIO SUBMITTED: sound=%s opcode=%s source=%d gain=%g wire-volume=%d", id, op, source, gain, volume)
		}
	}
	if events != 1 || len(probe.records) == 0 {
		probe.failure = fmt.Sprintf("stock door event ownership: events=%d voices=%d", events, len(probe.records))
	}
	probe.received = true
}

func e2eObserveDoorAudioRecord(record *e2eDoorAudioRecord) {
	if record.completed {
		return
	}
	record.queued = int(record.source.BuffersQueued())
	record.processed = int(record.source.BuffersProcessed())
	// Status stays playing until the encoded input reaches EOS and the source
	// finishes. Earlier recycled buffers have already been hardware-processed;
	// require every remaining submitted buffer too, before voice reinitialization.
	if record.queued > 0 && record.queued == record.processed && record.voice.Status() != 4 && record.source.State() == openal.Stopped {
		record.completed = true
		e2eLog.Printf("DOOR AUDIO DRAINED: source=%d final-queued=%d final-processed=%d gain=%g input-status=%d", record.source, record.queued, record.processed, record.gain, record.voice.Status())
	}
}

func e2eObserveDoorAudioVoiceReuseLocked(state *nativeAudioEffectsState, voice ail.Sample) {
	probe := e2eDoorAudioProbe
	if probe == nil || state != &nativeAudioFX {
		return
	}
	for _, record := range probe.records {
		if record.voice != voice || record.completed {
			continue
		}
		e2eObserveDoorAudioRecord(record)
		if !record.completed {
			probe.failure = "stock door voice reused before input EOS and full hardware drain"
		}
	}
}

func (sc *e2eScenario) CheckDoorAudio(name string) {
	var doors [5]*server.Object
	sc.add(0, name+" stock fixtures", func() {
		player := noxServer.Players.HostUnit()
		if os.Getenv("NOX_E2E_AUDIO") != "openal" || audioDev == 0 || player == nil || e2eDoorAudioProbe != nil {
			e2eError(fmt.Errorf("door playback audit requires one live host and real OpenAL"))
			return
		}
		for _, typ := range noxServer.Types.List() {
			if !typ.Class().Has(object.ClassDoor) {
				continue
			}
			subclass, material, category := uint32(typ.SubClass()), uint32(typ.Material()), 0
			switch {
			case subclass&4 != 0:
				category = 1
				if material&8 != 0 {
					category = 2
				}
			case subclass&1 != 0:
				category = 3
			case subclass&0x1000 != 0:
				category = 4
			}
			if doors[category] != nil {
				continue
			}
			door := noxServer.NewObjectByTypeID(typ.ID())
			if door == nil || door.UpdateData == nil || door.Update == nil {
				e2eError(fmt.Errorf("stock door %s lacks its native update", typ.ID()))
				return
			}
			pos := player.Pos()
			pos.X += 40
			pos.Y += float32(category * 5)
			noxServer.CreateObjectAt(door, nil, pos)
			noxServer.ObjectsAddPending()
			noxServer.Objs.RemoveFromUpdatable(door)
			if uint32(door.ObjSubClass) != subclass || uint32(door.Material) != material {
				e2eError(fmt.Errorf("stock door %s subtype/material changed during initialization", typ.ID()))
				return
			}
			doors[category] = door
			e2eLog.Printf("DOOR AUDIO STOCK: category=%d type=%s subclass=%#x material=%#x update=%p", category, typ.ID(), subclass, material, door.Update)
		}
		for i, door := range doors {
			if door == nil {
				e2eError(fmt.Errorf("stock door category %d is unavailable", i))
				return
			}
		}
		e2eDoorAudioProbe = &struct {
			expected sound.ID
			received bool
			records  []*e2eDoorAudioRecord
			failure  string
		}{}
	})
	for category, pair := range e2eDoorAudioPairs {
		for phase, id := range pair {
			sc.add(0, name+" "+id.String(), func() {
				probe := e2eDoorAudioProbe
				if probe == nil {
					return
				}
				probe.expected, probe.received, probe.records = id, false, nil
				door := doors[category]
				update := door.UpdateDataDoor()
				update.TargetDirection, update.SyncedDirection, update.CurrentDirection = 3, 3, 4
				if phase == 1 {
					update.SyncedDirection, update.CurrentDirection = 4, 3
				}
				// Controlled selector phases, not a door-motion or campaign claim.
				// Keep stock subtype/material, normal update dispatch, server audio,
				// client packet decoding, mixer and sample data intact.
				asObjectS(door).CallUpdate()
				noxServer.Objs.RemoveFromUpdatable(door)
			})
			sc.addWhen(0, name+" drain "+id.String(), 10*time.Second, func() bool {
				probe := e2eDoorAudioProbe
				if probe == nil {
					return false
				}
				if probe.failure != "" {
					e2eError(fmt.Errorf("%s", probe.failure))
					return false
				}
				if !probe.received {
					return false
				}
				nativeAudioFX.mu.Lock()
				for _, record := range probe.records {
					e2eObserveDoorAudioRecord(record)
				}
				nativeAudioFX.mu.Unlock()
				if probe.failure != "" {
					e2eError(fmt.Errorf("%s", probe.failure))
					return false
				}
				for _, record := range probe.records {
					if !record.completed {
						return false
					}
				}
				return true
			}, func() { e2eLog.Printf("DOOR AUDIO VERIFIED: sound=%s all-owned-voices-drained=true", id) })
		}
	}
	sc.add(0, name+" complete", func() {
		e2eLog.Printf("DOOR AUDIO COMPLETE: stock-types=5 selector-phases=10 real-source-gain-and-drain=true")
	})
}
