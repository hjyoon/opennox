package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestWinkGameBallReleaseNative4F7DF0Layout(t *testing.T) {
	wantTypeInd := uintptr(4)
	wantFlags := uintptr(16)
	wantPos := uintptr(56)
	wantNextOwned := uintptr(512)
	wantFirstOwned := uintptr(516)
	wantObj130 := uintptr(520)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantTypeInd = 8
		wantFlags = 20
		wantPos = 60
		wantNextOwned = 560
		wantFirstOwned = 568
		wantObj130 = 576
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.TypeInd", unsafe.Offsetof(Object{}.TypeInd), wantTypeInd},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
		{"Object.PosVec", unsafe.Offsetof(Object{}.PosVec), wantPos},
		{"Object.Field128", unsafe.Offsetof(Object{}.Field128), wantNextOwned},
		{"Object.Field129", unsafe.Offsetof(Object{}.Field129), wantFirstOwned},
		{"Object.Obj130", unsafe.Offsetof(Object{}.Obj130), wantObj130},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestWinkGameBallReleaseNative4F7DF0PreservesPointersAndFieldOrder(t *testing.T) {
	marker := &Object{}
	ball := &Object{
		TypeInd:  0x2468,
		ObjFlags: object.Flags(0xa5b6c7ff),
		Obj130:   marker,
	}
	wrong := &Object{TypeInd: 7, Field128: ball}
	player := &Object{PosVec: types.Pointf{X: -12.5, Y: 33.25}, Field129: wrong}
	ball.ObjOwner = player

	var pin runtime.Pinner
	for _, obj := range []*Object{marker, ball, wrong, player} {
		pin.Pin(obj)
	}
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]uintptr{
			"marker": uintptr(unsafe.Pointer(marker)),
			"ball":   uintptr(unsafe.Pointer(ball)),
			"wrong":  uintptr(unsafe.Pointer(wrong)),
			"player": uintptr(unsafe.Pointer(player)),
		} {
			if pointer <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, pointer)
			}
		}
	}

	cache := uint32(0)
	events := make([]string, 0, 7)
	got := winkGameBallReleaseNative4F7DF0(player, winkGameBallReleaseNativeDeps4F7DF0{
		loadTypeCache: func() uint32 {
			events = append(events, "cache")
			return cache
		},
		lookupType: func(name string) uint32 {
			events = append(events, "lookup:"+name)
			return 0x2468
		},
		storeTypeCache: func(value uint32) {
			events = append(events, "store-cache")
			cache = value
		},
		applyForce: func(gotPlayer, gotBall *Object, force float32) {
			events = append(events, "force")
			if gotPlayer != player || gotBall != ball || force != 100 || gotPlayer.PosVec != player.PosVec {
				t.Fatalf("force = player:%p ball:%p value:%g pos:%v", gotPlayer, gotBall, force, gotPlayer.PosVec)
			}
			if gotBall.ObjFlags != object.Flags(0xa5b6c7bf) || gotBall.Obj130 != marker {
				t.Fatalf("force-time ball = flags:%#x obj130:%p", gotBall.ObjFlags, gotBall.Obj130)
			}
		},
		clearOwner: func(gotBall *Object) {
			events = append(events, "clear-owner")
			if gotBall != ball || gotBall.Obj130 != nil {
				t.Fatalf("clear-owner ball = %p obj130:%p", gotBall, gotBall.Obj130)
			}
			gotBall.ObjOwner = nil
		},
		audio: func(id uint32, obj *Object, kind int32, code uint32) {
			events = append(events, "audio")
			if id != 926 || obj != player || kind != 0 || code != 0 {
				t.Fatalf("audio = %d/%p/%d/%d", id, obj, kind, code)
			}
		},
		ballStatus: func(state uint8, netCode uint16) int32 {
			events = append(events, "status")
			if state != 1 || netCode != 0 {
				t.Fatalf("status = %d/%d", state, netCode)
			}
			return -1
		},
	})
	if got != 1 || cache != 0x2468 {
		t.Fatalf("result/cache = %d/%#x, want 1/0x2468", got, cache)
	}
	wantEvents := []string{"cache", "lookup:GameBall", "store-cache", "force", "clear-owner", "audio", "status"}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	if ball.ObjFlags != object.Flags(0xa5b6c7bf) || ball.Obj130 != nil || ball.ObjOwner != nil {
		t.Fatalf("ball = flags:%#x obj130:%p owner:%p", ball.ObjFlags, ball.Obj130, ball.ObjOwner)
	}
	runtime.KeepAlive(marker)
	runtime.KeepAlive(ball)
	runtime.KeepAlive(wrong)
	runtime.KeepAlive(player)
}

func TestWinkGameBallReleaseServer4F7DF0BindsServicesAndDedicatedCache(t *testing.T) {
	s := &Server{}
	s.Types.byID = map[string]*ObjectType{
		"gameball": {ind: 0x2468},
	}
	s.Types.fast.ball = 11
	s.Types.fast.flagCollideGameBall = 22
	s.Types.fast.flagPickupGameBall = 33
	s.Types.fast.flagPickupBallStart = 44

	ball := &Object{TypeInd: 0x2468, ObjFlags: object.Flags(0xffffffff)}
	wrong := &Object{TypeInd: 9, Field128: ball}
	player := &Object{PosVec: types.Pointf{X: 5, Y: 7}, Field129: wrong}
	ball.ObjOwner = player
	forces := 0
	statuses := 0
	got := s.WinkGameBallRelease4F7DF0(player, WinkGameBallReleaseRuntime4F7DF0{
		ApplyForce: func(gotBall *Object, pos types.Pointf, force float64) {
			forces++
			if gotBall != ball || pos != player.PosVec || force != 100 {
				t.Fatalf("force = ball:%p pos:%v value:%g", gotBall, pos, force)
			}
		},
		BallStatus: func(state uint8, netCode uint16) int32 {
			statuses++
			if state != 1 || netCode != 0 {
				t.Fatalf("status = %d/%d", state, netCode)
			}
			return 0x7fffffff
		},
	})
	if got != 1 || forces != 1 || statuses != 1 {
		t.Fatalf("result/forces/statuses = %d/%d/%d", got, forces, statuses)
	}
	if s.Types.fast.winkGameBall4F7DF0 != 0x2468 {
		t.Fatalf("wink cache = %#x, want 0x2468", s.Types.fast.winkGameBall4F7DF0)
	}
	if s.Types.fast.ball != 11 || s.Types.fast.flagCollideGameBall != 22 ||
		s.Types.fast.flagPickupGameBall != 33 || s.Types.fast.flagPickupBallStart != 44 {
		t.Fatalf("unrelated caches changed: common=%d router=%d pickup=%d start=%d",
			s.Types.fast.ball, s.Types.fast.flagCollideGameBall,
			s.Types.fast.flagPickupGameBall, s.Types.fast.flagPickupBallStart)
	}
	if player.Field129 != wrong || wrong.Field128 != nil || ball.ObjOwner != nil {
		t.Fatalf("owned list = head:%p next:%p owner:%p", player.Field129, wrong.Field128, ball.ObjOwner)
	}
	if len(s.Audio.delayedObj) != 1 {
		t.Fatalf("delayed audio count = %d, want 1", len(s.Audio.delayedObj))
	}
	audio := s.Audio.delayedObj[0]
	if audio.ID != 926 || audio.Obj != player || audio.Kind != 0 || audio.Code != 0 {
		t.Fatalf("audio = %+v", audio)
	}
}
