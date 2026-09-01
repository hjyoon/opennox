package server

import (
	"errors"
	"reflect"
	"testing"

	"github.com/opennox/noxscript/ns/asm"
)

func mapInitializeScriptFunc4FC590(name string) ScriptFunc {
	return ScriptFunc{FuncDef: asm.FuncDef{Name: name}}
}

func TestMapInitializeServer4FC590PreservesGatesAndHookOrder(t *testing.T) {
	for _, tc := range []struct {
		name      string
		state     int32
		withUnit  bool
		wantState int32
		wantHooks []string
	}{
		{name: "zero state", withUnit: true},
		{name: "no player unit", state: -1985229329, wantState: -1985229329},
		{name: "dispatch", state: -1985229329, withUnit: true, wantHooks: []string{"before", "after"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := new(Server)
			s.NoxScriptVM.Init(s)
			s.SetMapInitState4FC570(tc.state)
			if tc.withUnit {
				s.Players.list = []Player{{PlayerUnit: new(Object)}}
			}
			var hooks []string
			got := s.MapInitialize4FC590(MapInitializeRuntime4FC590{
				BeforeLegacy: func() {
					hooks = append(hooks, "before")
					if value := s.MapInitState4FC570(); value != tc.state {
						t.Fatalf("state before legacy = %#08x, want %#08x", uint32(value), uint32(tc.state))
					}
				},
				AfterLegacy: func() {
					hooks = append(hooks, "after")
					if value := s.MapInitState4FC570(); value != tc.state {
						t.Fatalf("state before clear = %#08x, want %#08x", uint32(value), uint32(tc.state))
					}
				},
			})
			if got != 0 {
				t.Fatalf("result = %#08x, want zero", uint32(got))
			}
			if value := s.MapInitState4FC570(); value != tc.wantState {
				t.Fatalf("state = %#08x, want %#08x", uint32(value), uint32(tc.wantState))
			}
			if !reflect.DeepEqual(hooks, tc.wantHooks) {
				t.Fatalf("hooks = %v, want %v", hooks, tc.wantHooks)
			}
		})
	}
}

func TestMapInitializeServer4FC590ReloadsVMTableAndCount(t *testing.T) {
	s := new(Server)
	s.NoxScriptVM.Init(s)
	s.SetMapInitState4FC570(1)
	s.Players.list = []Player{{PlayerUnit: new(Object)}}
	s.NoxScriptVM.vm.funcs = []ScriptFunc{
		mapInitializeScriptFunc4FC590("MapInitializeFirst"),
	}

	var events []string
	s.NoxScriptVM.virtual.funcs[0] = nsCallback{Name: "first", Fnc: func() error {
		events = append(events, "legacy:0")
		s.NoxScriptVM.vm.funcs = []ScriptFunc{
			mapInitializeScriptFunc4FC590("ignored"),
			mapInitializeScriptFunc4FC590("Other"),
			mapInitializeScriptFunc4FC590("MapInitializeAdded"),
		}
		return nil
	}}
	s.NoxScriptVM.virtual.funcs[2] = nsCallback{Name: "added", Fnc: func() error {
		events = append(events, "legacy:2")
		return nil
	}}

	got := s.MapInitialize4FC590(MapInitializeRuntime4FC590{
		BeforeLegacy: func() { events = append(events, "before") },
		AfterLegacy:  func() { events = append(events, "after") },
	})
	if got != 0 || s.MapInitState4FC570() != 0 {
		t.Fatalf("result/state = %d/%d, want 0/0", got, s.MapInitState4FC570())
	}
	if want := []string{"before", "legacy:0", "legacy:2", "after"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestMapInitializeServer4FC590ScriptErrorDoesNotSkipTailOrClear(t *testing.T) {
	s := new(Server)
	s.NoxScriptVM.Init(s)
	s.SetMapInitState4FC570(1)
	s.Players.list = []Player{{PlayerUnit: new(Object)}}
	s.NoxScriptVM.vm.funcs = []ScriptFunc{
		mapInitializeScriptFunc4FC590("MapInitialize"),
	}
	s.NoxScriptVM.virtual.funcs[0] = nsCallback{Name: "failing", Fnc: func() error {
		return errors.New("expected script failure")
	}}

	var after bool
	if got := s.MapInitialize4FC590(MapInitializeRuntime4FC590{
		AfterLegacy: func() { after = true },
	}); got != 0 {
		t.Fatalf("result = %d, want zero", got)
	}
	if !after || s.MapInitState4FC570() != 0 {
		t.Fatalf("after/state = %v/%d, want true/0", after, s.MapInitState4FC570())
	}
}

func TestMapInitializeServer4FC590TailPanicPreservesState(t *testing.T) {
	s := new(Server)
	s.NoxScriptVM.Init(s)
	s.SetMapInitState4FC570(-1985229329)
	s.Players.list = []Player{{PlayerUnit: new(Object)}}

	defer func() {
		if recover() == nil {
			t.Fatal("tail panic was not propagated")
		}
		if got := s.MapInitState4FC570(); uint32(got) != 0x89abcdef {
			t.Fatalf("state after tail panic = %#08x, want unchanged", uint32(got))
		}
	}()
	s.MapInitialize4FC590(MapInitializeRuntime4FC590{
		AfterLegacy: func() { panic("tail") },
	})
}
