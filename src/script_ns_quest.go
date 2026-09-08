package opennox

import (
	"github.com/opennox/noxscript/ns/v4"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func (s noxScriptNS) GetQuestStatus(name string) int {
	return int(legacy.QuestJournalGetInt500750(name))
}

func (s noxScriptNS) GetQuestStatusFloat(name string) float32 {
	//TODO implement me
	panic("implement me")
}

func (s noxScriptNS) SetQuestStatus(status int, name string) {
	//TODO implement me
	panic("implement me")
}

func (s noxScriptNS) SetQuestStatusFloat(status float32, name string) {
	//TODO implement me
	panic("implement me")
}

func (s noxScriptNS) ResetQuestStatus(name string) {
	//TODO implement me
	panic("implement me")
}

func (s noxScriptNS) JournalEntry(obj ns.Obj, msg ns.StringID, typ ns.EntryType) {
	if obj == nil {
		for _, it := range s.s.Players.ListUnits() {
			s.s.JournalEntryAdd427500(it, string(msg), uint16(typ))
		}
		return
	}
	unit := journalObjectNS(obj)
	if unit == nil {
		return
	}
	s.s.JournalEntryAdd427500(unit, string(msg), uint16(typ))
	if (typ & 0xB) != 0 {
		s.s.Audio.EventObj(sound.SoundJournalEntryAdd, unit, 0, 0)
	}
}

func (s noxScriptNS) JournalEdit(obj ns.Obj, message ns.StringID, typ ns.EntryType) {
	if obj == nil {
		for _, it := range s.s.Players.ListUnits() {
			s.s.JournalEntryUpdate427720(it, string(message), uint16(typ))
		}
		return
	}
	s.s.JournalEntryUpdate427720(journalObjectNS(obj), string(message), uint16(typ))
}

func (s noxScriptNS) JournalDelete(obj ns.Obj, message ns.StringID) {
	if obj == nil {
		for _, it := range s.s.Players.ListUnits() {
			s.s.JournalEntryRemove427630(it, string(message))
		}
		return
	}
	s.s.JournalEntryRemove427630(journalObjectNS(obj), string(message))
}

func (s noxScriptNS) JournalEntryStr(obj ns.Obj, msg string, typ ns.EntryType) {
	s.JournalEntry(obj, ns.StringID(msg), typ)
}

func (s noxScriptNS) JournalEditStr(obj ns.Obj, message string, typ ns.EntryType) {
	s.JournalEdit(obj, ns.StringID(message), typ)
}

func (s noxScriptNS) JournalDeleteStr(obj ns.Obj, message string) {
	s.JournalDelete(obj, ns.StringID(message))
}

func journalObjectNS(obj ns.Obj) *server.Object {
	if obj == nil {
		return nil
	}
	serverObj, ok := obj.(server.Obj)
	if !ok {
		return nil
	}
	return server.ToObject(serverObj)
}
