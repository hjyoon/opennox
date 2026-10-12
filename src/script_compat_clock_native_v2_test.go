package opennox

import (
	"bytes"
	"fmt"
	"math/big"
	"testing"
	"time"
	"unsafe"

	"github.com/opennox/libs/script"
	"github.com/opennox/libs/script/lua"
	"github.com/opennox/libs/script/lua/mapv0"
)

// Exact rational nanoseconds, independent of either production clock.
func luaCompatClockExpected(frame, rate uint32) time.Duration {
	value := new(big.Int).SetUint64(uint64(frame))
	value.Mul(value, big.NewInt(int64(time.Second)))
	value.Quo(value, new(big.Int).SetUint64(uint64(rate)))
	return time.Duration(value.Int64())
}

func luaCompatClockDue(start, rate uint32, dt time.Duration) uint32 {
	deadline := luaCompatClockExpected(start, rate) + dt
	for frame := start; frame <= start+rate*2+2; frame++ {
		if luaCompatClockExpected(frame, rate) >= deadline {
			return frame
		}
	}
	panic("bounded test duration exceeds two seconds")
}

// The control is the existing native NoxScript clock, not a supplied result.
type luaCompatClockControl struct{ noxScriptImpl }

func (g luaCompatClockControl) Time() time.Duration {
	return g.s.Server.NoxScriptNS().Time()
}

func TestLuaCompatClockNativeSnapshot(t *testing.T) {
	s := scriptNSRoamFlagServer515C80(t)
	var old script.Game = noxScriptImpl{s: s}
	rates := []uint32{1, 2, 3, 7, 29, 30, 59, 60, 100, 144, 240, 1000, 65535, 0xffffffff}
	for _, rate := range rates {
		frames := []struct {
			name  string
			frame uint32
		}{
			{"zero", 0}, {"first", 1}, {"second", 2},
			{"half-second", rate / 2}, {"before-second", rate - 1},
			{"whole-second", rate}, {"after-second-or-wrap", rate + 1},
			{"signed-last", 0x7fffffff}, {"signed-first", 0x80000000},
			{"unsigned-before-last", 0xfffffffe}, {"unsigned-last", 0xffffffff},
			{"long-map", 342001},
		}
		for _, frame := range frames {
			t.Run(fmt.Sprintf("rate-%d/%s", rate, frame.name), func(t *testing.T) {
				s.SetFrame(frame.frame)
				s.SetTickRate(rate)
				before := scriptNSRoamFlagBytes515C80(unsafe.Pointer(s.Server), unsafe.Sizeof(*s.Server))
				want := luaCompatClockExpected(frame.frame, rate)
				if control := s.Server.NoxScriptNS().Time(); control != want {
					t.Fatalf("native clock control=%d, exact rational time=%d", control, want)
				}
				for n := 0; n < 2; n++ {
					if got := old.Time(); got != want {
						t.Errorf("legacy clock=%d ns, want %d ns at frame=%d rate=%d", got, want, frame.frame, rate)
					}
					if got := old.Frame(); got != int(frame.frame) {
						t.Errorf("frame=%d, want %d", got, frame.frame)
					}
				}
				after := scriptNSRoamFlagBytes515C80(unsafe.Pointer(s.Server), unsafe.Sizeof(*s.Server))
				if !bytes.Equal(before, after) {
					t.Fatal("reading clocks changed native server state")
				}
			})
		}
	}
}

func TestLuaCompatClockNativeTimers(t *testing.T) {
	s := scriptNSRoamFlagServer515C80(t)
	legacyClock := noxScriptImpl{s: s}
	controlClock := luaCompatClockControl{legacyClock}
	for _, rate := range []uint32{3, 7, 30, 60, 144, 240} {
		for _, start := range []uint32{0, 1, rate - 1, rate, rate + 5} {
			for _, delay := range []time.Duration{0, 1, 17 * time.Millisecond, 50 * time.Millisecond,
				125 * time.Millisecond, 250 * time.Millisecond, 500 * time.Millisecond, 1500 * time.Millisecond} {
				t.Run(fmt.Sprintf("rate-%d/start-%d/delay-%d", rate, start, delay), func(t *testing.T) {
					s.SetTickRate(rate)
					s.SetFrame(start)
					actual, control := script.NewTimers(legacyClock), script.NewTimers(controlClock)
					var actualCalls, controlCalls []uint32
					actual.SetTimer(script.Time(delay), func() { actualCalls = append(actualCalls, s.Frame()) })
					control.SetTimer(script.Time(delay), func() { controlCalls = append(controlCalls, s.Frame()) })
					due := luaCompatClockDue(start, rate, delay)
					for frame := start; frame <= due+3; frame++ {
						s.SetFrame(frame)
						actual.Tick()
						control.Tick()
					}
					if len(controlCalls) != 1 || controlCalls[0] != due {
						t.Fatalf("native timer control=%v, exact first due frame=%d", controlCalls, due)
					}
					if len(actualCalls) != 1 || actualCalls[0] != due {
						t.Fatalf("legacy seconds timer=%v, exact first due frame=%d; native=%v", actualCalls, due, controlCalls)
					}
				})
			}
		}
	}
}

func luaCompatClockVM(t *testing.T, game script.Game) *lua.VM {
	t.Helper()
	vm := lua.NewVM(nil, game, "")
	t.Cleanup(func() { _ = vm.Close() })
	vm.InitAPI("Nox", mapv0.New)
	return vm
}

func TestLuaCompatClockNativeLuaTimers(t *testing.T) {
	s := scriptNSRoamFlagServer515C80(t)
	legacyClock := noxScriptImpl{s: s}
	for _, rate := range []uint32{7, 30, 60, 144} {
		for _, start := range []uint32{0, 1, rate - 1} {
			for _, delay := range []time.Duration{50 * time.Millisecond, 125 * time.Millisecond,
				250 * time.Millisecond, 500 * time.Millisecond, 1500 * time.Millisecond} {
				for _, constructor := range []string{"seconds", "duration"} {
					t.Run(fmt.Sprintf("rate-%d/start-%d/delay-%d/%s", rate, start, delay, constructor), func(t *testing.T) {
						s.SetTickRate(rate)
						s.SetFrame(start)
						actual := luaCompatClockVM(t, legacyClock)
						control := luaCompatClockVM(t, luaCompatClockControl{legacyClock})
						timer := fmt.Sprintf("Nox.SecondTimer(%.9f, callback, 71, 'clock')", delay.Seconds())
						if constructor == "duration" {
							timer = fmt.Sprintf("Nox.Timer(Nox.Sec(%.9f), callback, 71, 'clock')", delay.Seconds())
						}
						program := "calls=0; fired=-1; frameCalls=0; frameFired=-1; stoppedCalls=0; argsOK=true; " +
							"callback=function(a,b) calls=calls+1; fired=currentFrame; argsOK=argsOK and a==71 and b=='clock'; end; " +
							timer + "; Nox.FrameTimer(2,function() frameCalls=frameCalls+1; frameFired=currentFrame; end); " +
							"stopped=Nox.SecondTimer(0.01,function() stoppedCalls=stoppedCalls+1; end); stopped:Stop()"
						for _, vm := range []*lua.VM{actual, control} {
							if _, err := vm.Exec(program); err != nil {
								t.Fatal(err)
							}
						}
						due := luaCompatClockDue(start, rate, delay)
						end := due + 3
						if end < start+5 {
							end = start + 5
						}
						for frame := start; frame <= end; frame++ {
							s.SetFrame(frame)
							for _, vm := range []*lua.VM{actual, control} {
								if _, err := vm.Exec(fmt.Sprintf("currentFrame=%d", frame)); err != nil {
									t.Fatal(err)
								}
								vm.OnFrame()
							}
						}
						check := fmt.Sprintf("assert(calls==1 and fired==%d and argsOK and frameCalls==1 and frameFired==%d and stoppedCalls==0, 'seconds='..calls..' fired='..fired..' frame='..frameCalls..' at='..frameFired..' stopped='..stoppedCalls)", due, start+2)
						if _, err := control.Exec(check); err != nil {
							t.Fatalf("native-clock Lua control: %v", err)
						}
						if _, err := actual.Exec(check); err != nil {
							t.Fatalf("legacy-clock Lua: %v", err)
						}
					})
				}
			}
		}
	}
}
