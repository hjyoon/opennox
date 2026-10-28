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
	ArachnaphobiaUpdateRaw53DA60(source, ArachnaphobiaRawRuntime53DA60{
		Frame: deps.frame, FPS: deps.fps,
		LoadTypeIndex:      func() uint32 { return 1 },
		NewObjectByTypeInd: func(int32) *Object { return deps.newObjectByTypeID("SmallSpider") },
		CreateAt:           deps.createAt, RandomInt: deps.randomInt, DelayedDelete: deps.delayedDelete,
	})
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
