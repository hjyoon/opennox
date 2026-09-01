package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestWarpReadUseNative53F830Layouts(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantObjClass := uintptr(8)
	wantUseData := uintptr(736)
	wantUpdateData := uintptr(748)
	wantUpdatePlayer := uintptr(276)
	wantPlayerIndex := uintptr(2064)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantObjClass = 12
		wantUseData = 848
		wantUpdateData = 872
		wantUpdatePlayer = 336
		wantPlayerIndex = 2068
	}
	for _, check := range []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantObjClass},
		{"Object.UseData", unsafe.Offsetof(Object{}.UseData), wantUseData},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdateData},
		{"ReadableUseData size", unsafe.Sizeof(ReadableUseData{}), 260},
		{"ReadableUseData.Text", unsafe.Offsetof(ReadableUseData{}.Text), 0},
		{"ReadableUseData.TransientReadState", unsafe.Offsetof(ReadableUseData{}.TransientReadState), 256},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantUpdatePlayer},
		{"Player.PlayerInd", unsafe.Offsetof(Player{}.PlayerInd), wantPlayerIndex},
	} {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestWarpReadUseNative53F830OpenPreservesPointers(t *testing.T) {
	player := &Player{PlayerInd: 0xfe}
	update := &PlayerUpdateData{Player: player}
	owner := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	data := &ReadableUseData{}
	readable := &Object{}
	readable.UseData.SetPtr(unsafe.Pointer(data))

	frame := uint32(0x10203040)
	var (
		mapOwner       *Object
		mapReadable    *Object
		informedIndex  uint8
		informedCode   uint8
		informedValue  int32
		priorityCalled bool
	)
	deps := warpReadUseNativeDeps53F830{
		loadFPS:   func() uint32 { return 30 },
		loadFrame: func() uint32 { return frame },
		mapCheck: func(gotOwner, gotReadable *Object) int32 {
			mapOwner, mapReadable = gotOwner, gotReadable
			return 1
		},
		warpEnabled:       func() int32 { return -1 },
		currentQuestStage: func() int32 { return 7 },
		nextStageThreshold: func(stage int32) int32 {
			if stage != 7 {
				t.Fatalf("stage = %d, want 7", stage)
			}
			return math.MinInt32
		},
		informText: func(index, code uint8, value int32) {
			informedIndex, informedCode, informedValue = index, code, value
			frame = 0x89abcdef
		},
		priorityMessage: func(*Object, string, uint8) {
			priorityCalled = true
		},
	}
	if got := warpReadUseNative53F830(owner, readable, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if mapOwner != owner || mapReadable != readable {
		t.Fatalf("map args = %p/%p, want %p/%p", mapOwner, mapReadable, owner, readable)
	}
	if priorityCalled {
		t.Fatal("open path sent the closed priority message")
	}
	if informedIndex != 0xfe || informedCode != 21 || informedValue != math.MinInt32 {
		t.Fatalf("inform = %#x/%d/%d, want 0xfe/21/%d", informedIndex, informedCode, informedValue, int32(math.MinInt32))
	}
	if data.TransientReadState != 0x89abcdef {
		t.Fatalf("read state = %#x, want 0x89abcdef", data.TransientReadState)
	}

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, ptr := range map[string]uintptr{
			"owner":    uintptr(unsafe.Pointer(owner)),
			"readable": uintptr(unsafe.Pointer(readable)),
			"update":   uintptr(unsafe.Pointer(update)),
			"data":     uintptr(unsafe.Pointer(data)),
			"player":   uintptr(unsafe.Pointer(player)),
		} {
			if ptr <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, ptr)
			}
		}
	}
	runtime.KeepAlive(owner)
	runtime.KeepAlive(readable)
	runtime.KeepAlive(update)
	runtime.KeepAlive(data)
	runtime.KeepAlive(player)
}

func TestWarpReadUseNative53F830ClosedDoesNotDereferenceUpdatePlayer(t *testing.T) {
	owner := &Object{ObjClass: object.ClassPlayer}
	data := &ReadableUseData{}
	readable := &Object{}
	readable.UseData.SetPtr(unsafe.Pointer(data))

	frame := uint32(100)
	var (
		messageOwner *Object
		messageKey   string
		messageValue uint8
	)
	deps := warpReadUseNativeDeps53F830{
		loadFPS:           func() uint32 { return 20 },
		loadFrame:         func() uint32 { return frame },
		mapCheck:          func(*Object, *Object) int32 { return 1 },
		warpEnabled:       func() int32 { return 0 },
		currentQuestStage: func() int32 { t.Fatal("closed path loaded stage"); return 0 },
		nextStageThreshold: func(int32) int32 {
			t.Fatal("closed path loaded next threshold")
			return 0
		},
		informText: func(uint8, uint8, int32) {
			t.Fatal("closed path sent open inform")
		},
		priorityMessage: func(gotOwner *Object, key string, value uint8) {
			messageOwner, messageKey, messageValue = gotOwner, key, value
			frame = 0x76543210
		},
	}
	if got := warpReadUseNative53F830(owner, readable, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if messageOwner != owner || messageKey != "GeneralPrint:WarpClosed" || messageValue != 1 {
		t.Fatalf("priority = %p/%q/%d, want %p/GeneralPrint:WarpClosed/1", messageOwner, messageKey, messageValue, owner)
	}
	if data.TransientReadState != 0x76543210 {
		t.Fatalf("read state = %#x, want 0x76543210", data.TransientReadState)
	}
}

func TestWarpReadUseNative53F830NonPlayerNeedsNoRuntime(t *testing.T) {
	s := &Server{}
	if got := s.WarpReadUse53F830(
		&Object{ObjClass: object.ClassImmobile},
		nil,
		WarpReadUseRuntime53F830{},
	); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
}
