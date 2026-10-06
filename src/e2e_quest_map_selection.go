package opennox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/opennox/libs/prand"
	"github.com/opennox/opennox/v1/legacy"
)

// This reference is separate from the production selector. It decodes signed
// DWORD comparisons using int64, builds an ordered candidate list and uses a
// PRIVATE RNG at the captured index. It never calls the production selection
// helpers, resets the game RNG, changes history or supplies a selected map.
// Branches are from GAME.EXE 004D0F60..004D10E4, especially the signed cooldown
// comparison after DWORD subtraction at 004D1040/004D10A9.
func e2eQuestMapReference4D0F60(state legacy.QuestMapSelectionState4D0F60) (int32, int, error) {
	if state.Count > 128 || int(state.Count) != len(state.Entries) || state.LogicIndex < 0 || state.LogicIndex >= 4096 {
		return 0, 0, fmt.Errorf("invalid Quest selection snapshot: count=%d entries=%d logic=%d", state.Count, len(state.Entries), state.LogicIndex)
	}
	random := prand.New(state.LogicIndex)
	signed := func(value uint32) int64 {
		if value&0x80000000 != 0 {
			return int64(value) - (1 << 32)
		}
		return int64(value)
	}
	switch state.Count {
	case 0:
		return -2, random.Index(), nil
	case 1:
		return 0, random.Index(), nil
	}
	maximum, equal := uint32(0), true
	for _, entry := range state.Entries {
		if signed(entry.Uses) > signed(maximum) {
			maximum = entry.Uses
		}
		if entry.Uses != state.Entries[0].Uses {
			equal = false
		}
	}
	if maximum == 0 {
		return int32(random.Int(0, len(state.Entries)-1)), random.Index(), nil
	}
	if state.LastIndex >= state.Count {
		return 0, 0, fmt.Errorf("used Quest history has invalid last index %d", state.LastIndex)
	}
	threshold := maximum
	if equal {
		threshold = uint32((uint64(maximum) + 1) & 0xffffffff)
	}
	var candidates []int32
	for index, entry := range state.Entries {
		age := uint32((uint64(state.Clock) + (1 << 32) - uint64(entry.LastUsed)) & 0xffffffff)
		if signed(entry.Uses) < signed(threshold) && uint32(index) != state.LastIndex &&
			entry.Group != state.Entries[state.LastIndex].Group && signed(age) > 4 {
			candidates = append(candidates, int32(index))
		}
	}
	if len(candidates) == 0 {
		// Original Random(0,-1) returns -1 without consuming a table entry.
		return -1, random.Index(), nil
	}
	choice := random.Int(0, len(candidates)-1)
	return candidates[choice], random.Index(), nil
}

func e2eQuestMapSelectionValidate4D0F60(selection legacy.QuestMapSelection4D0F60) (string, error) {
	want, logic, err := e2eQuestMapReference4D0F60(selection.Before)
	if err != nil {
		return "", err
	}
	after := selection.After
	after.LogicIndex = selection.Before.LogicIndex
	if selection.ResultIndex != want || selection.After.LogicIndex != logic || !reflect.DeepEqual(after, selection.Before) ||
		(selection.ResultPointer == 0) != (want == -2) {
		return "", fmt.Errorf("Quest selector disagrees with original reference: index=%d/%d logic=%d/%d untouched-history/Other=%t native=%#x",
			selection.ResultIndex, want, selection.After.LogicIndex, logic, reflect.DeepEqual(after, selection.Before), selection.ResultPointer)
	}
	if want < 0 {
		return "", fmt.Errorf("stock Quest world selected null/fallback index %d", want)
	}
	name := selection.Before.Entries[want].Name
	end := bytes.IndexByte(name[:], 0)
	if end <= 0 {
		return "", fmt.Errorf("Quest selected map has no bounded nonempty name")
	}
	return string(name[:end]), nil
}

var e2eQuestSelectedMaps4D0F60 struct {
	armed bool
	maps  []string
}

func (sc *e2eScenario) ObserveQuestMapSelection(name string) {
	sc.add(0, name, func() {
		if e2eQuestSelectedMaps4D0F60.armed {
			e2eError(fmt.Errorf("Quest selection observer already armed"))
			return
		}
		e2eQuestSelectedMaps4D0F60.armed = true
		legacy.ObserveQuestMapSelection4D0F60(func(selection legacy.QuestMapSelection4D0F60) {
			// Retain every raw input for independent offline verification.
			data, err := json.Marshal(selection)
			if err != nil {
				e2eError(err)
				return
			}
			e2eLog.Printf("QUEST SELECTION SNAPSHOT: %s", data)
			mapName, err := e2eQuestMapSelectionValidate4D0F60(selection)
			if err != nil {
				e2eError(err)
				return
			}
			e2eQuestSelectedMaps4D0F60.maps = append(e2eQuestSelectedMaps4D0F60.maps, mapName)
			e2eLog.Printf("QUEST SELECTION ORIGINAL MATCH: selection=%d map=%q native=%#x logic=%d->%d other=%d clock=%d last=%d history=untouched",
				len(e2eQuestSelectedMaps4D0F60.maps), mapName, selection.ResultPointer, selection.Before.LogicIndex,
				selection.After.LogicIndex, selection.Before.OtherIndex, selection.Before.Clock, selection.Before.LastIndex)
		})
	})
}

func (sc *e2eScenario) AssertQuestSelectedMap(stage int, name string) {
	sc.add(0, name, func() {
		if !e2eQuestSelectedMaps4D0F60.armed || stage <= 0 || len(e2eQuestSelectedMaps4D0F60.maps) != stage {
			e2eError(fmt.Errorf("Quest stage %d lacks exactly %d verified selections: got %d", stage, stage, len(e2eQuestSelectedMaps4D0F60.maps)))
			return
		}
		mapName := e2eQuestSelectedMaps4D0F60.maps[stage-1]
		state, err := e2eReadQuestPlayer()
		if err == nil {
			err = state.validate(stage, mapName)
		}
		if err != nil {
			e2eError(err)
			return
		}
		e2eLog.Printf("QUEST SELECTED MAP PLAYABLE: stage=%d selected=%q loaded=%q frame=%d wire=%d phase=%d health=%d/%d observer=false briefing=false",
			stage, mapName, state.mapName, noxServer.Frame(), state.serverCode, state.phase, state.health, state.maxHealth)
	})
}

// CaptureQuestMapAuditFrame keeps diagnostic frames separate from fixed-map
// screenshot goldens. Playability and native selection are asserted separately.
func (sc *e2eScenario) CaptureQuestMapAuditFrame(name string) {
	sc.add(0, name, func() {
		path, err := e2eWriteMagicFrame("", noxClient.r.CopyPixBuffer())
		if err != nil {
			e2eError(fmt.Errorf("capture Quest map audit frame %q: %w", name, err))
			return
		}
		e2eLog.Printf("QUEST MAP AUDIT FRAME: name=%q frame=%d path=%q", name, noxServer.Frame(), path)
	})
}
