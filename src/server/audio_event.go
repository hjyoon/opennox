package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type audioEventObj struct {
	ID   sound.ID
	Obj  *Object
	Kind int
	Code uint32
}

type audioEventPos struct {
	ID   sound.ID
	Pos  types.Pointf
	Kind int
	Code uint32
}

type OnSoundFunc func(id sound.ID, kind int, obj *Object, pos types.Pointf)

type serverAudio struct {
	s          *Server
	alloc      alloc.ClassT[AudioEvent]
	inited     bool
	head       *AudioEvent
	bySound    [1024]audioEvent2
	bitmap     soundBitmap501E80
	onSound    []OnSoundFunc
	delayedObj []audioEventObj
	delayedPos []audioEventPos
	// inAudio indicates that the server is currently accepting audio events.
	//
	// If this flag is set, all audio events will be added directly to the list.
	// And if not, we will queue them for the next frame.
	inAudio bool
}

type AudioEvent struct {
	next0  *AudioEvent
	Sound  sound.ID
	Pos    types.Pointf
	Obj    *Object
	Kind   int
	Code   uint32
	list28 *AudioEvent
	Perc   int
}

type audioEvent2 struct {
	MaxDist int
	Flags   uint32
	Field8  uint32
	Field12 uint32
	Field16 uint32
	Field20 int
	Field24 *AudioEvent
}

func (s *serverAudio) Init(srv *Server) {
	if s.inited {
		return
	}
	s.s = srv
	s.alloc = alloc.NewClassT("AudEvent", AudioEvent{}, 128)
	for i := range s.bySound[:len(s.bySound)-1] {
		p := &s.bySound[i]
		p.MaxDist = 600
		p.Field8 = 0
		p.Field12 = 0
		p.Field16 = 0
		p.Field20 = 0
		p.Field24 = nil
	}
	s.inited = true
}

func (s *serverAudio) Free() {
	if s.inited {
		s.alloc.Free()
		s.head = nil
		s.onSound = nil
		s.inAudio = false
		s.delayedObj = nil
		s.delayedPos = nil
		s.inited = false
	}
}

func (s *serverAudio) Reset() {
	s.resetAudioEvents502100()
	s.inAudio = true
	// Replay events queued outside the audio loop.
	for _, a := range s.delayedObj {
		s.EventObj(a.ID, a.Obj, a.Kind, a.Code)
	}
	s.delayedObj = s.delayedObj[:0]
	for _, a := range s.delayedPos {
		s.EventPos(a.ID, a.Pos, a.Kind, a.Code)
	}
	s.delayedPos = s.delayedPos[:0]
}

func (s *serverAudio) OnSound(fnc OnSoundFunc) {
	s.onSound = append(s.onSound, fnc)
}

func (s *serverAudio) Flags(id sound.ID) int {
	return int(s.bySound[id].Flags)
}

func (s *serverAudio) MaxDist(id sound.ID) int {
	return s.bySound[id].MaxDist
}

func (s *serverAudio) Field12(id sound.ID) int32 {
	return int32(s.bySound[id].Field12)
}

func (s *serverAudio) Field20(id sound.ID) int {
	return s.bySound[id].Field20
}

func (s *serverAudio) newAudioEventObj(id sound.ID, iobj Obj, kind int, code uint32) {
	obj := ToObject(iobj)
	if noxflags.HasGame(noxflags.GameFlag20) {
		return
	}
	if id == 0 || obj == nil || s.Field12(id) <= 0 {
		return
	}
	p := s.alloc.NewObject()
	if p == nil {
		return
	}
	p.Sound = id
	if obj.Flags().Has(object.FlagDestroyed) {
		p.Obj = nil
	} else {
		p.Obj = obj
	}
	p.Pos = obj.Pos()
	p.Kind = kind
	p.Code = code
	p.next0 = s.head
	s.head = p
}

func (s *serverAudio) newAudioEventPos(id sound.ID, pos types.Pointf, kind int, code uint32) {
	if noxflags.HasGame(noxflags.GameFlag20) {
		return
	}
	if id == 0 || s.Field12(id) <= 0 {
		return
	}
	p := s.alloc.NewObject()
	if p == nil {
		return
	}
	p.Sound = id
	p.Pos = pos
	p.Obj = nil
	p.Kind = kind
	p.Code = code
	p.next0 = s.head
	s.head = p
}

func (s *serverAudio) EachEvent(fnc func(ev *AudioEvent)) {
	for it := s.head; it != nil; it = it.next0 {
		fnc(it)
	}
}

func (s *serverAudio) EventObj(id sound.ID, iobj Obj, kind int, code uint32) {
	obj := ToObject(iobj)
	if !s.inAudio {
		s.delayedObj = append(s.delayedObj, audioEventObj{
			ID: id, Obj: obj, Kind: kind, Code: code,
		})
		return
	}
	if noxflags.HasGame(noxflags.GameFlag20) {
		return
	}
	if id == 0 || obj == nil || s.Field12(id) <= 0 {
		return
	}
	if noxflags.HasGame(noxflags.GameModeQuest) && obj.Class().Has(object.ClassPlayer) && s.s.Players.CheckXxx(obj) {
		return
	}
	s.newAudioEventObj(id, obj, kind, code)
	for _, fnc := range s.onSound {
		fnc(id, kind, obj, obj.Pos())
	}
}

func (s *serverAudio) EventPos(id sound.ID, pos types.Pointf, kind int, code uint32) {
	if !s.inAudio {
		s.delayedPos = append(s.delayedPos, audioEventPos{
			ID: id, Pos: pos, Kind: kind, Code: code,
		})
		return
	}
	if noxflags.HasGame(noxflags.GameFlag20) {
		return
	}
	if id == 0 || s.Field12(id) <= 0 {
		return
	}
	s.newAudioEventPos(id, pos, kind, code)
	for _, fnc := range s.onSound {
		fnc(id, kind, nil, pos)
	}
}
