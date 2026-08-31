package server

import (
	"unsafe"

	"github.com/opennox/libs/strman"
)

type readUseNativeDeps53F7C0 struct {
	loadFPS        func() uint32
	loadFrame      func() uint32
	mapCheck       func(*Object, *Object) int32
	primaryMessage func(*Object, *ReadableUseData, uint8)
}

func readUseNative53F7C0(
	owner, readable *Object,
	deps readUseNativeDeps53F7C0,
) int32 {
	return readUse53F7C0(readUseHooks53F7C0[*Object, *ReadableUseData]{
		loadOwnerArg: func() *Object {
			return owner
		},
		loadClassLow: func(owner *Object) uint8 {
			return uint8(owner.ObjClass)
		},
		loadReadableArg: func() *Object {
			return readable
		},
		loadFPS:   deps.loadFPS,
		loadFrame: deps.loadFrame,
		loadUseData: func(readable *Object) *ReadableUseData {
			return readable.UseData.AsReadable()
		},
		loadReadState: func(data *ReadableUseData) uint32 {
			return data.TransientReadState
		},
		mapCheck:       deps.mapCheck,
		primaryMessage: deps.primaryMessage,
		storeReadState: func(data *ReadableUseData, frame uint32) {
			data.TransientReadState = frame
		},
	})
}

func readUseServerDeps53F7C0(s *Server) readUseNativeDeps53F7C0 {
	return readUseNativeDeps53F7C0{
		loadFPS:   s.TickRate,
		loadFrame: s.Frame,
		mapCheck: func(owner, readable *Object) int32 {
			if s.MapTraceVision(owner, readable) {
				return 1
			}
			return 0
		},
		primaryMessage: func(owner *Object, data *ReadableUseData, value uint8) {
			s.NetPriMsgToPlayer(owner, strman.ID(data.TextString()), value)
		},
	}
}

// ReadUse53F7C0 binds GAME.EXE 0053F7C0 to native-width Object and UseData
// pointers. It is the registered ReadUse implementation used by signs and
// other native object-use dispatchers.
func (s *Server) ReadUse53F7C0(owner, readable *Object) int32 {
	return readUseNative53F7C0(owner, readable, readUseServerDeps53F7C0(s))
}

var (
	_ = [1]struct{}{}[unsafe.Sizeof(uintptr(0))-unsafe.Sizeof(Object{}.UseData)]
	_ = [1]struct{}{}[260-unsafe.Sizeof(ReadableUseData{})]
	_ = [1]struct{}{}[256-unsafe.Offsetof(ReadableUseData{}.TransientReadState)]
)
