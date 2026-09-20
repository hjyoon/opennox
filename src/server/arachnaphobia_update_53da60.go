package server

import "github.com/opennox/libs/types"

const arachnaphobiaLifetimeSeconds53DA60 = uint32(3)

// ArachnaphobiaUpdateRuntime53DA60 supplies the object creation, random, and
// deletion services reached by the native-width restoration of GAME.EXE
// 0053DA60.
type ArachnaphobiaUpdateRuntime53DA60 struct {
	NewObjectByTypeID func(string) *Object
	CreateAt          func(*Object, *Object, types.Pointf)
	RandomInt         func(int, int) int
	DelayedDelete     func(*Object)
}

type arachnaphobiaUpdateDeps53DA60 struct {
	frame             func() uint32
	fps               func() uint32
	newObjectByTypeID func(string) *Object
	createAt          func(*Object, *Object, types.Pointf)
	randomInt         func(int, int) int
	delayedDelete     func(*Object)
}

func arachnaphobiaUpdateNative53DA60(source *Object, deps arachnaphobiaUpdateDeps53DA60) {
	ageFrame := deps.frame()
	if source.Field34 < deps.frame() {
		spider := deps.newObjectByTypeID("SmallSpider")
		if spider != nil {
			deps.createAt(spider, source.ObjOwner, source.PosVec)
		}
		source.Field34 = deps.frame() + uint32(deps.randomInt(1, 5))
		ageFrame = deps.frame()
	}
	if ageFrame-source.Field32 > arachnaphobiaLifetimeSeconds53DA60*deps.fps() {
		deps.delayedDelete(source)
	}
}

// ArachnaphobiaUpdate53DA60 restores GAME.EXE 0053DA60 without reading the
// native-width Object and its owner pointer through PE32 word offsets.
func (s *Server) ArachnaphobiaUpdate53DA60(source *Object, runtime ArachnaphobiaUpdateRuntime53DA60) {
	if source == nil {
		return
	}
	arachnaphobiaUpdateNative53DA60(source, arachnaphobiaUpdateDeps53DA60{
		frame:             s.Frame,
		fps:               s.TickRate,
		newObjectByTypeID: runtime.NewObjectByTypeID,
		createAt:          runtime.CreateAt,
		randomInt:         runtime.RandomInt,
		delayedDelete:     runtime.DelayedDelete,
	})
}
