package legacy

import (
	"reflect"
	"testing"
)

type mapModeSetupLegacyServer417EA0 struct {
	Server
	flags  bool
	result int
	calls  []string
}

func (s *mapModeSetupLegacyServer417EA0) MapInfoSetCapflag417EA0() int {
	s.calls = append(s.calls, "ctf")
	return s.result
}

func (s *mapModeSetupLegacyServer417EA0) MapInfoSetFlags417EC0() bool {
	s.calls = append(s.calls, "flags")
	return s.flags
}

func (s *mapModeSetupLegacyServer417EA0) MapInfoSetFlagball417F30() int {
	s.calls = append(s.calls, "flagball")
	return s.result
}

func (s *mapModeSetupLegacyServer417EA0) MapInfoSetKotr4180D0() int {
	s.calls = append(s.calls, "kotr")
	return s.result
}

func TestMapModeSetupExports417EA0UseNativeRuntimeAndResultABI(t *testing.T) {
	outer := &mapModeSetupLegacyServer417EA0{}
	oldGetServer := GetServer
	GetServer = func() Server { return outer }
	t.Cleanup(func() { GetServer = oldGetServer })

	for _, result := range []int{0, 1} {
		outer.result = result
		outer.flags = result != 0
		outer.calls = nil
		if got := Nox_xxx_mapInfoSetCapflag_417EA0(); got != result {
			t.Fatalf("CTF result = %d, want %d", got, result)
		}
		if got := mapInfoSetFlagsExportCall417EC0(); got != outer.flags {
			t.Fatalf("flag scan = %t, want %t", got, outer.flags)
		}
		if got := Nox_xxx_mapInfoSetFlagball_417F30(); got != result {
			t.Fatalf("FlagBall result = %d, want %d", got, result)
		}
		if got := Nox_xxx_mapInfoSetKotr_4180D0(); got != result {
			t.Fatalf("KOTR result = %d, want %d", got, result)
		}
		if want := []string{"ctf", "flags", "flagball", "kotr"}; !reflect.DeepEqual(outer.calls, want) {
			t.Fatalf("native setup calls = %v, want %v", outer.calls, want)
		}
	}

	outer.result = 0x7f102345
	if got := Nox_xxx_mapInfoSetCapflag_417EA0(); got != outer.result {
		t.Fatalf("CTF int result = %#x, want %#x", got, outer.result)
	}
	if got := Nox_xxx_mapInfoSetKotr_4180D0(); got != outer.result {
		t.Fatalf("KOTR int result = %#x, want %#x", got, outer.result)
	}
}
