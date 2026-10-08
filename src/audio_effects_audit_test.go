package opennox

import (
	"strings"
	"testing"

	"github.com/opennox/opennox/v1/common/sound"
)

func TestNativeAudioDefinitionsAudit(t *testing.T) {
	var state nativeAudioEffectsState
	if _, err := auditNativeAudioDefinitions(&state); err == nil {
		t.Fatal("unloaded bank passed the read-only audit")
	}
	entry := &nativeAudioBankEntry{name: "step", rate: 22050, flags: 8, blockSize: 256, data: make([]byte, 256)}
	state.bank = &nativeAudioBank{entries: map[string]*nativeAudioBankEntry{"step": entry}}
	state.defs[sound.SoundWalkOnStone] = nativeSoundDef{enabled: true, sampleIDs: []string{"STEP", "step"}, volume: 123}
	state.defs[sound.SoundRunOnStone] = nativeSoundDef{enabled: true}
	state.defs[sound.SoundMetalWeaponPickup] = nativeSoundDef{enabled: false, sampleIDs: []string{"disabled-missing"}}
	report, err := auditNativeAudioDefinitions(&state)
	if err != nil || report.bankSamples != 1 || report.enabledDefinitions != 2 || report.sampleReferences != 2 || report.emptyDefinitions != 1 {
		t.Fatalf("valid definitions report=%+v err=%v", report, err)
	}
	if state.defs[sound.SoundWalkOnStone].volume != 123 || state.sequence != 0 || state.next != 0 {
		t.Fatal("audio audit changed definition gain or sample selection")
	}
	state.defs[sound.SoundRunOnStone].sampleIDs = []string{"missing"}
	if _, err := auditNativeAudioDefinitions(&state); err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("unresolved enabled reference was not reported: %v", err)
	}
	state.defs[sound.SoundRunOnStone].sampleIDs = nil
	entry.flags = 0
	if _, err := auditNativeAudioDefinitions(&state); err == nil || !strings.Contains(err.Error(), "unplayable") {
		t.Fatalf("unsupported bank entry was not reported: %v", err)
	}
}
