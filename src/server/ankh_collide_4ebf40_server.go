package server

import (
	"unsafe"

	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// AnkhHistoryRecord is the fixed-width identity record used by AnkhInit.
// Its strings remain the original 25 UTF-16 code units and 25 bytes so the
// 5124-byte object InitData layout is identical on every architecture.
type AnkhHistoryRecord struct {
	NameBuf     [25]uint16
	PlayerClass uint8
	SerialBuf   [25]byte
	Frame       uint32
}

func (r *AnkhHistoryRecord) name4EBF40() string {
	return alloc.GoString16S(r.NameBuf[:])
}

func (r *AnkhHistoryRecord) setName4EBF40(value string) {
	alloc.StrCopy16(r.NameBuf[:], value)
}

func (r *AnkhHistoryRecord) serial4EBF40() string {
	return alloc.GoStringS(r.SerialBuf[:])
}

func (r *AnkhHistoryRecord) setSerial4EBF40(value string) {
	alloc.StrCopy(r.SerialBuf[:], value)
}

// AnkhInitData is registered by AnkhInit with an exact size of 5124 bytes.
// Next is a live uint8 ring index; the three trailing bytes are alignment in
// the original record rather than native pointers.
type AnkhInitData struct {
	Records [ankhHistoryCount4EBF40]AnkhHistoryRecord
	Next    uint8
	_       [3]byte
}

// AnkhCollideRuntime4EBF40 supplies shared legacy clock/reset state and the
// point-effect service. Object, PlayerUpdateData, Player, history, and item
// pointers remain native-width inside package server.
type AnkhCollideRuntime4EBF40 struct {
	Ticks                func() uint64
	LoadFeedbackTicks    func() uint64
	StoreFeedbackTicks   func(uint64)
	LoadResetName        func() string
	LoadResetSerialFirst func() uint8
	PointFX              func(uint32, *types.Pointf) uint32
}

type ankhCollideNativeDeps4EBF40 struct {
	loadFPS              func() uint32
	loadFrame            func() uint32
	ticks                func() uint64
	loadFeedbackTicks    func() uint64
	storeFeedbackTicks   func(uint64)
	loadResetName        func() string
	loadResetSerialFirst func() uint8
	loadBalance          func(string) float32
	floatToInt           func(float32) int32
	newObject            func(string) *Object
	callPickup           func(*Object, *Object, int32, uint32)
	audio                func(uint32, *Object, int32, int32)
	pointFX              func(uint32, *types.Pointf) uint32
	priorityMessage      func(*Object, string, int32)
}

func ankhCollideNative4EBF40(
	source, target *Object,
	collision *types.Pointf,
	deps ankhCollideNativeDeps4EBF40,
) {
	ankhCollide4EBF40(
		source,
		target,
		collision,
		ankhCollideHooks4EBF40[
			*Object,
			*AnkhInitData,
			*PlayerUpdateData,
			*Player,
		]{
			loadSourceInitData: func(obj *Object) *AnkhInitData {
				return (*AnkhInitData)(obj.InitData)
			},
			loadTargetClassLow: func(obj *Object) uint8 {
				return uint8(obj.ObjClass)
			},
			loadTargetUpdate: func(obj *Object) *PlayerUpdateData {
				return (*PlayerUpdateData)(obj.UpdateData)
			},
			loadPlayer: func(update *PlayerUpdateData) *Player {
				return update.Player
			},
			loadQuestAnkh: func(player *Player, index int) *Object {
				return player.QuestAnkhs[index]
			},
			storeQuestAnkh: func(player *Player, index int, obj *Object) {
				player.QuestAnkhs[index] = obj
			},
			loadFPS:   deps.loadFPS,
			loadFrame: deps.loadFrame,
			loadRecordFrame: func(data *AnkhInitData, index int) uint32 {
				return data.Records[index].Frame
			},
			loadResetName: deps.loadResetName,
			storeRecordName: func(data *AnkhInitData, index int, value string) {
				data.Records[index].setName4EBF40(value)
			},
			loadResetSerialFirst: deps.loadResetSerialFirst,
			storeRecordSerialFirst: func(data *AnkhInitData, index int, value uint8) {
				data.Records[index].SerialBuf[0] = value
			},
			storeRecordClass: func(data *AnkhInitData, index int, value uint8) {
				data.Records[index].PlayerClass = value
			},
			storeRecordFrame: func(data *AnkhInitData, index int, value uint32) {
				data.Records[index].Frame = value
			},
			loadRecordClass: func(data *AnkhInitData, index int) uint8 {
				return data.Records[index].PlayerClass
			},
			loadPlayerClass: func(player *Player) uint8 {
				return uint8(player.Info().PlayerClass())
			},
			loadRecordName: func(data *AnkhInitData, index int) string {
				return data.Records[index].name4EBF40()
			},
			loadPlayerName: func(player *Player) string {
				return player.Info().Name()
			},
			loadRecordSerial: func(data *AnkhInitData, index int) string {
				return data.Records[index].serial4EBF40()
			},
			loadPlayerSerial: func(player *Player) string {
				return player.Serial()
			},
			storeRecordSerial: func(data *AnkhInitData, index int, value string) {
				data.Records[index].setSerial4EBF40(value)
			},
			ticks:              deps.ticks,
			loadFeedbackTicks:  deps.loadFeedbackTicks,
			storeFeedbackTicks: deps.storeFeedbackTicks,
			priorityMessage:    deps.priorityMessage,
			audio:              deps.audio,
			loadBalance:        deps.loadBalance,
			floatToInt:         deps.floatToInt,
			loadExtraLives: func(update *PlayerUpdateData) int32 {
				return int32(update.ExtraLives)
			},
			newObject:  deps.newObject,
			callPickup: deps.callPickup,
			storeSourceFrame: func(obj *Object, frame uint32) {
				obj.Field34 = frame
			},
			pointFX: func(id uint32, obj *Object) uint32 {
				return deps.pointFX(id, &obj.PosVec)
			},
			loadNextIndex: func(data *AnkhInitData) uint8 {
				return data.Next
			},
			storeNextIndex: func(data *AnkhInitData, value uint8) {
				data.Next = value
			},
		},
	)
}

func ankhCollideServerDeps4EBF40(
	s *Server,
	runtime AnkhCollideRuntime4EBF40,
) ankhCollideNativeDeps4EBF40 {
	return ankhCollideNativeDeps4EBF40{
		loadFPS:              s.TickRate,
		loadFrame:            s.Frame,
		ticks:                runtime.Ticks,
		loadFeedbackTicks:    runtime.LoadFeedbackTicks,
		storeFeedbackTicks:   runtime.StoreFeedbackTicks,
		loadResetName:        runtime.LoadResetName,
		loadResetSerialFirst: runtime.LoadResetSerialFirst,
		loadBalance: func(key string) float32 {
			return float32(s.Balance.Float(key))
		},
		floatToInt: ankhRoundFloat32ToInt32_4EBF40,
		newObject:  s.NewObjectByTypeID,
		callPickup: func(who, item *Object, first int32, second uint32) {
			_ = item.CallPickup(who, int(first), int(second))
		},
		audio: func(id uint32, obj *Object, first, second int32) {
			s.Audio.EventObj(sound.ID(id), obj, int(first), uint32(second))
		},
		pointFX: runtime.PointFX,
		priorityMessage: func(obj *Object, message string, value int32) {
			s.NetPriMsgToPlayer(obj, strman.ID(message), byte(value))
		},
	}
}

// AnkhCollide4EBF40 binds GAME.EXE 004EBF40 to native-width Object,
// PlayerUpdateData, Player, Quest-Ankh slots, and fixed-width history data.
// Collision remains in the callback signature but is deliberately unread.
func (s *Server) AnkhCollide4EBF40(
	source, target *Object,
	collision *types.Pointf,
	runtime AnkhCollideRuntime4EBF40,
) {
	ankhCollideNative4EBF40(
		source,
		target,
		collision,
		ankhCollideServerDeps4EBF40(s, runtime),
	)
}

var (
	_ = [1]struct{}{}[80-unsafe.Sizeof(AnkhHistoryRecord{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(AnkhHistoryRecord{}.NameBuf)]
	_ = [1]struct{}{}[50-unsafe.Offsetof(AnkhHistoryRecord{}.PlayerClass)]
	_ = [1]struct{}{}[51-unsafe.Offsetof(AnkhHistoryRecord{}.SerialBuf)]
	_ = [1]struct{}{}[76-unsafe.Offsetof(AnkhHistoryRecord{}.Frame)]
	_ = [1]struct{}{}[5124-unsafe.Sizeof(AnkhInitData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(AnkhInitData{}.Records)]
	_ = [1]struct{}{}[5120-unsafe.Offsetof(AnkhInitData{}.Next)]
)
