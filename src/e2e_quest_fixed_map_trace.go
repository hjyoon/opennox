package opennox

import (
	"encoding/json"

	"github.com/opennox/opennox/v1/legacy"
)

// Fixed-map assertions retain their original expectations. This separate trace
// records choices after the first successful checkpoint; it does not supply a
// map, advance either RNG, change history, or turn a failed assertion into a
// passing one. Like the explicit map-oracle observer, it is E2E-process scoped.
type e2eQuestFixedMapTrace4D0F60 struct {
	armed bool
	stop  func()
}

var e2eQuestFixedMaps4D0F60 e2eQuestFixedMapTrace4D0F60

func (trace *e2eQuestFixedMapTrace4D0F60) arm(
	stage int,
	mapName string,
	install func(func(legacy.QuestMapSelection4D0F60)) func(),
	emit func(string, ...any),
) {
	if stage != 1 || mapName == "" || trace.armed {
		return
	}
	trace.armed = true
	trace.stop = install(func(selection legacy.QuestMapSelection4D0F60) {
		data, err := json.Marshal(selection)
		if err != nil {
			emit("QUEST FIXED MAP TRACE ERROR: snapshot=%v", err)
			return
		}
		emit("QUEST FIXED MAP SELECTION SNAPSHOT: %s", data)
		selected, err := e2eQuestMapSelectionValidate4D0F60(selection)
		if err != nil {
			emit("QUEST FIXED MAP REFERENCE DIFFERENCE: %v", err)
			return
		}
		emit("QUEST FIXED MAP ORIGINAL MATCH: map=%q native=%#x logic=%d->%d other=%d clock=%d last=%d history=untouched",
			selected, selection.ResultPointer, selection.Before.LogicIndex, selection.After.LogicIndex,
			selection.Before.OtherIndex, selection.Before.Clock, selection.Before.LastIndex)
	})
}

func e2eQuestTraceFixedMaps4D0F60(stage int, mapName string) {
	e2eQuestFixedMaps4D0F60.arm(stage, mapName, legacy.ObserveQuestMapSelection4D0F60, e2eLog.Printf)
}
