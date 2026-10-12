package opennox

import (
	"fmt"
	"testing"

	"github.com/opennox/libs/script"
)

// Server.New starts with tickRate=0; the existing native audio/Lua tests also
// construct their VM before rate setup. Preserve the legacy zero bootstrap
// time while retaining exact fractional time after initialization.
func TestLuaCompatClockBootstrapSnapshot(t *testing.T) {
	s := scriptNSRoamFlagServer515C80(t)
	var game script.Game = noxScriptImpl{s: s}
	for _, frame := range []uint32{0, 1, 37, 0xffffffff} {
		t.Run(fmt.Sprintf("frame-%d", frame), func(t *testing.T) {
			s.SetTickRate(0)
			s.SetFrame(frame)
			if got := game.Time(); got != 0 {
				t.Fatalf("unconfigured server clock=%d, want original bootstrap zero", got)
			}
			if game.Frame() != int(frame) || s.TickRate() != 0 {
				t.Fatal("reading bootstrap clock changed frame or configured rate")
			}
		})
	}
}

func TestLuaCompatClockBootstrapLuaVM(t *testing.T) {
	s := scriptNSRoamFlagServer515C80(t)
	s.SetFrame(0)
	s.SetTickRate(0)
	vm := luaCompatClockVM(t, noxScriptImpl{s: s})
	if _, err := vm.Exec("calls=0; frameCalls=0; fired=-1; Nox.SecondTimer(0.25,function() calls=calls+1; fired=currentFrame; end); Nox.FrameTimer(2,function() frameCalls=frameCalls+1; end)"); err != nil {
		t.Fatal(err)
	}
	for frame := uint32(0); frame <= 2; frame++ {
		s.SetFrame(frame)
		if _, err := vm.Exec(fmt.Sprintf("currentFrame=%d", frame)); err != nil {
			t.Fatal(err)
		}
		vm.OnFrame()
	}
	if _, err := vm.Exec("assert(calls==0 and frameCalls==1, 'bootstrap timer behavior changed')"); err != nil {
		t.Fatal(err)
	}
	s.SetTickRate(30)
	for frame := uint32(2); frame <= 11; frame++ {
		s.SetFrame(frame)
		if _, err := vm.Exec(fmt.Sprintf("currentFrame=%d", frame)); err != nil {
			t.Fatal(err)
		}
		vm.OnFrame()
	}
	if _, err := vm.Exec("assert(calls==1 and fired==8 and frameCalls==1, 'initialized timer did not fire at exact deadline')"); err != nil {
		t.Fatal(err)
	}
}
