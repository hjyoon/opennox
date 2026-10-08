package opennox

import (
	"fmt"
	"strings"

	"github.com/opennox/opennox/v1/common/sound"
)

type nativeAudioDefinitionsAudit struct {
	bankSamples        int
	enabledDefinitions int
	sampleReferences   int
	emptyDefinitions   int
}

// Inspect the loaded bank and every enabled AUD/AVNT reference without
// changing definitions, random streams, voices, or gameplay state. An enabled
// definition with no samples is deliberately silent and is counted separately.
func auditNativeAudioDefinitions(s *nativeAudioEffectsState) (nativeAudioDefinitionsAudit, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var report nativeAudioDefinitionsAudit
	if s.bank == nil || len(s.bank.entries) == 0 {
		return report, fmt.Errorf("native FX bank is not loaded")
	}
	var failures []string
	for name, entry := range s.bank.entries {
		report.bankSamples++
		if entry == nil || len(entry.data) == 0 || entry.flags&8 == 0 || entry.blockSize == 0 || entry.rate == 0 {
			failures = append(failures, fmt.Sprintf("unplayable bank sample %q", name))
		}
	}
	for id := 1; id < len(s.defs); id++ {
		def := &s.defs[id]
		if !def.enabled {
			continue
		}
		report.enabledDefinitions++
		if len(def.sampleIDs) == 0 {
			report.emptyDefinitions++
		}
		for _, name := range def.sampleIDs {
			report.sampleReferences++
			if s.bank.entries[strings.ToLower(name)] == nil {
				failures = append(failures, fmt.Sprintf("%s references missing sample %q", sound.ID(id), name))
			}
		}
	}
	if len(failures) != 0 {
		return report, fmt.Errorf("native audio bank/reference audit: %s", strings.Join(failures, "; "))
	}
	return report, nil
}
