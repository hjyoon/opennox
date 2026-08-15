package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func defaultAnkhCollideNativeDeps4EBF40() ankhCollideNativeDeps4EBF40 {
	return ankhCollideNativeDeps4EBF40{
		loadFPS:              func() uint32 { return 1 },
		loadFrame:            func() uint32 { return 0 },
		ticks:                func() uint64 { return 0 },
		loadFeedbackTicks:    func() uint64 { return 0 },
		storeFeedbackTicks:   func(uint64) {},
		loadResetName:        func() string { return "" },
		loadResetSerialFirst: func() uint8 { return 0 },
		loadBalance:          func(string) float32 { return 1 },
		floatToInt:           ankhRoundFloat32ToInt32_4EBF40,
		newObject:            func(string) *Object { return nil },
		callPickup:           func(*Object, *Object, int32, uint32) {},
		audio:                func(uint32, *Object, int32, int32) {},
		pointFX:              func(uint32, *types.Pointf) uint32 { return 0 },
		priorityMessage:      func(*Object, string, int32) {},
	}
}

func TestAnkhCollide4EBF40NativeLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantObjectClass := uintptr(8)
	wantObjectPosition := uintptr(56)
	wantObjectFrame := uintptr(136)
	wantObjectInitData := uintptr(692)
	wantObjectUpdateData := uintptr(748)
	wantUpdateSize := uintptr(556)
	wantPlayerPointer := uintptr(276)
	wantExtraLives := uintptr(320)
	wantPlayerSize := uintptr(4828)
	wantSerial := uintptr(2112)
	wantInfo := uintptr(2185)
	wantQuestState := uintptr(4792)
	wantQuestAnkhs := uintptr(4796)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantObjectClass = 12
		wantObjectPosition = 60
		wantObjectFrame = 140
		wantObjectInitData = 760
		wantObjectUpdateData = 872
		wantUpdateSize = 640
		wantPlayerPointer = 320
		wantExtraLives = 400
		wantPlayerSize = 6160
		wantSerial = 2116
		wantInfo = 2189
		wantQuestState = 6096
		wantQuestAnkhs = 6104
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantObjectClass},
		{"Object.PosVec", unsafe.Offsetof(Object{}.PosVec), wantObjectPosition},
		{"Object.Field34", unsafe.Offsetof(Object{}.Field34), wantObjectFrame},
		{"Object.InitData", unsafe.Offsetof(Object{}.InitData), wantObjectInitData},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantObjectUpdateData},
		{"PlayerUpdateData size", unsafe.Sizeof(PlayerUpdateData{}), wantUpdateSize},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantPlayerPointer},
		{"PlayerUpdateData.ExtraLives", unsafe.Offsetof(PlayerUpdateData{}.ExtraLives), wantExtraLives},
		{"Player size", unsafe.Sizeof(Player{}), wantPlayerSize},
		{"Player.SerialBuf", unsafe.Offsetof(Player{}.SerialBuf), wantSerial},
		{"Player.info", unsafe.Offsetof(Player{}.info), wantInfo},
		{"Player.Field4792", unsafe.Offsetof(Player{}.Field4792), wantQuestState},
		{"Player.QuestAnkhs", unsafe.Offsetof(Player{}.QuestAnkhs), wantQuestAnkhs},
		{"Player.QuestAnkhs size", unsafe.Sizeof(Player{}.QuestAnkhs), 5 * unsafe.Sizeof(uintptr(0))},
		{"PlayerInfo size", unsafe.Sizeof(PlayerInfo{}), 97},
		{"PlayerInfo.playerClass", unsafe.Offsetof(PlayerInfo{}.playerClass), 66},
		{"AnkhHistoryRecord size", unsafe.Sizeof(AnkhHistoryRecord{}), 80},
		{"AnkhHistoryRecord.NameBuf", unsafe.Offsetof(AnkhHistoryRecord{}.NameBuf), 0},
		{"AnkhHistoryRecord.PlayerClass", unsafe.Offsetof(AnkhHistoryRecord{}.PlayerClass), 50},
		{"AnkhHistoryRecord.SerialBuf", unsafe.Offsetof(AnkhHistoryRecord{}.SerialBuf), 51},
		{"AnkhHistoryRecord.Frame", unsafe.Offsetof(AnkhHistoryRecord{}.Frame), 76},
		{"AnkhInitData size", unsafe.Sizeof(AnkhInitData{}), 5124},
		{"AnkhInitData.Records", unsafe.Offsetof(AnkhInitData{}.Records), 0},
		{"AnkhInitData.Next", unsafe.Offsetof(AnkhInitData{}.Next), 5120},
		{"Pointf size", unsafe.Sizeof(types.Pointf{}), 8},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestAnkhCollideNative4EBF40BindsCachedAndLiveFields(t *testing.T) {
	entryData := &AnkhInitData{Next: 3}
	replacementData := &AnkhInitData{Next: 40}
	source := &Object{
		PosVec:   types.Pointf{X: 12.5, Y: -7.25},
		InitData: unsafe.Pointer(entryData),
	}
	entryPlayer := &Player{}
	entryPlayer.Info().SetName("entry")
	entryPlayer.Info().playerClass = 7
	entryPlayer.SetSerial("entry-serial")
	livePlayer := &Player{}
	livePlayer.Info().SetName("영웅")
	livePlayer.Info().playerClass = 9
	livePlayer.SetSerial("live-serial")
	entryUpdate := &PlayerUpdateData{Player: entryPlayer, ExtraLives: 1}
	replacementUpdate := &PlayerUpdateData{Player: &Player{}, ExtraLives: 99}
	target := &Object{
		ObjClass:   object.ClassPlayer | object.Class(0x40000000),
		UpdateData: unsafe.Pointer(entryUpdate),
	}
	created := &Object{}
	collision := &types.Pointf{
		X: math.Float32frombits(0x7fc12345),
		Y: math.Float32frombits(0x80000000),
	}
	wantCollisionX := math.Float32bits(collision.X)
	wantCollisionY := math.Float32bits(collision.Y)
	events := make([]string, 0, 16)
	frameReads := 0

	deps := defaultAnkhCollideNativeDeps4EBF40()
	deps.loadFPS = func() uint32 { return 1 }
	deps.loadFrame = func() uint32 {
		frameReads++
		switch frameReads {
		case ankhHistoryCount4EBF40 + 1:
			events = append(events, "source-frame")
			return 111
		case ankhHistoryCount4EBF40 + 2:
			events = append(events, "history-frame")
			return 222
		default:
			return 0
		}
	}
	deps.loadBalance = func(key string) float32 {
		events = append(events, "balance")
		if key != ankhBalanceKey4EBF40 {
			t.Fatalf("balance key = %q", key)
		}
		return 2.5
	}
	deps.newObject = func(name string) *Object {
		events = append(events, "new")
		if name != ankhTradableType4EBF40 {
			t.Fatalf("new object = %q", name)
		}
		return created
	}
	deps.callPickup = func(who, item *Object, first int32, second uint32) {
		events = append(events, "pickup")
		if who != target || item != created || first != 1 || second != 0 {
			t.Fatalf("Pickup = %p/%p/%d/%d", who, item, first, second)
		}
		source.InitData = unsafe.Pointer(replacementData)
		target.UpdateData = unsafe.Pointer(replacementUpdate)
		entryUpdate.Player = livePlayer
	}
	deps.audio = func(id uint32, obj *Object, first, second int32) {
		events = append(events, "audio")
		if id != ankhAwardAudio4EBF40 || obj != source || first != 0 || second != 0 || source.Field34 != 111 {
			t.Fatalf("audio = %d/%p/%d/%d, frame %d", id, obj, first, second, source.Field34)
		}
	}
	deps.pointFX = func(id uint32, position *types.Pointf) uint32 {
		events = append(events, "fx")
		if id != ankhAwardPointFX4EBF40 || position != &source.PosVec {
			t.Fatalf("point FX = %d/%p, want %d/%p", id, position, ankhAwardPointFX4EBF40, &source.PosVec)
		}
		position.Y = 44
		return 0xf1234567
	}
	deps.priorityMessage = func(obj *Object, message string, value int32) {
		events = append(events, "message")
		if obj != target || message != ankhAwardedExtraLifeMessage4EBF40 || value != 0 {
			t.Fatalf("message = %p/%q/%d", obj, message, value)
		}
	}
	deps.ticks = func() uint64 {
		t.Fatal("award path unexpectedly read feedback ticks")
		return 0
	}

	ankhCollideNative4EBF40(source, target, collision, deps)

	if frameReads != ankhHistoryCount4EBF40+2 {
		t.Fatalf("frame reads = %d, want %d", frameReads, ankhHistoryCount4EBF40+2)
	}
	if got, want := events, []string{
		"balance", "new", "pickup", "source-frame", "audio", "fx", "message", "history-frame",
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("events = %#v, want %#v", got, want)
	}
	if source.Field34 != 111 || source.PosVec.Y != 44 ||
		math.Float32bits(collision.X) != wantCollisionX || math.Float32bits(collision.Y) != wantCollisionY {
		t.Fatalf("source/collision = frame %d, pos %v, collision bits %#x/%#x", source.Field34, source.PosVec, math.Float32bits(collision.X), math.Float32bits(collision.Y))
	}
	if livePlayer.QuestAnkhs[0] != source {
		t.Fatalf("live Quest slot = %p, want source %p", livePlayer.QuestAnkhs[0], source)
	}
	record := &entryData.Records[3]
	if record.name4EBF40() != "영웅" || record.PlayerClass != 9 ||
		record.serial4EBF40() != "live-serial" || record.Frame != 222 || entryData.Next != 4 {
		t.Fatalf("entry history = name %q/class %d/serial %q/frame %d/next %d",
			record.name4EBF40(), record.PlayerClass, record.serial4EBF40(), record.Frame, entryData.Next)
	}
	if replacementData.Next != 40 || replacementData.Records[3].Frame != 0 {
		t.Fatal("callback replacement InitData was used instead of the cached pointer")
	}
	if replacementUpdate.Player.QuestAnkhs[0] != nil {
		t.Fatal("callback replacement UpdateData was used instead of the cached pointer")
	}
}

func TestAnkhCollide4EBF40ServerDepsBindClockResetAndPointFX(t *testing.T) {
	srv := new(Server)
	srv.SetFrame(0xfedcba98)
	srv.SetTickRate(73)
	position := types.Pointf{X: 1, Y: 2}
	stored := uint64(0)
	runtime := AnkhCollideRuntime4EBF40{
		Ticks:                func() uint64 { return 0x0123456789abcdef },
		LoadFeedbackTicks:    func() uint64 { return 0xfedcba9876543210 },
		StoreFeedbackTicks:   func(value uint64) { stored = value },
		LoadResetName:        func() string { return "reset" },
		LoadResetSerialFirst: func() uint8 { return 0x7f },
		PointFX: func(id uint32, got *types.Pointf) uint32 {
			if id != 55 || got != &position {
				t.Fatalf("point FX = %d/%p", id, got)
			}
			return 0x89abcdef
		},
	}
	deps := ankhCollideServerDeps4EBF40(srv, runtime)
	if deps.loadFrame() != 0xfedcba98 || deps.loadFPS() != 73 ||
		deps.ticks() != 0x0123456789abcdef || deps.loadFeedbackTicks() != 0xfedcba9876543210 ||
		deps.loadResetName() != "reset" || deps.loadResetSerialFirst() != 0x7f ||
		deps.floatToInt(2.5) != 2 || deps.pointFX(55, &position) != 0x89abcdef {
		t.Fatal("server dependency binding mismatch")
	}
	deps.storeFeedbackTicks(0x1122334455667788)
	if stored != 0x1122334455667788 {
		t.Fatalf("stored feedback = %#x", stored)
	}
}
