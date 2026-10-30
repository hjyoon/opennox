package server

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/common/sound"
)

// MonsterMoveAudio534030 binds the original movement sound update to native
// records. The original AI dispatcher discards its incidental EAX/AL value.
func (s *Server) MonsterMoveAudio534030(unit *Object) {
	s.monsterMoveAudio534030(unit)
}

func (s *Server) monsterMoveAudio534030(unit *Object) {
	if unit == nil || unit.UpdateData == nil {
		return
	}
	// 00534030 reads the entry update record without a class check. The
	// class-gated sound lookup occurs only after selecting a footstep frame.
	update := (*MonsterUpdateData)(unit.UpdateData)
	def := update.MonsterDef
	if def == nil {
		return
	}
	var play bool
	if uint8(unit.ObjSubClass)&0x30 != 0 {
		frames, delay := s.PlayerAnimFrames(4)
		count, period := uint32(frames), uint32(delay)+1
		if count == 0 || period == 0 {
			return
		}
		// Both ADD and DEC wrap as DWORDs before unsigned DIV. Load frame
		// once, including when NetCode+frame or the previous tick wraps.
		tick := unit.NetCode + s.Frame()
		cur, prev := tick/period%count, (tick-1)/period%count
		play = cur != prev && (cur == def.MoveSndFrameA100 || cur == def.MoveSndFrameB104)
	} else {
		frame := uint32(update.Field120_1)
		play = (frame == def.MoveSndFrameA100 || frame == def.MoveSndFrameB104) && update.Field120_2 == 0
	}
	if !play || uint8(unit.ObjClass)&uint8(object.ClassMonster) == 0 {
		return
	}
	// Match the original 00424300 lookup through the live native record.
	sounds := (*MonsterUpdateData)(unit.UpdateData).SoundSet122
	if sounds != nil {
		id := sound.ID(*(*uint32)(unsafe.Add(sounds, 72)))
		s.Audio.EventObj(id, unit, 0, 0)
	}
}
