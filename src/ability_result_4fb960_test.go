package opennox

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

func TestAbilityResult4FB960ExactTableAndCallOrder(t *testing.T) {
	wantKeys := [...]string{
		"AbilityOK",
		"AbilityInUse",
		"AbilityNotReady",
		"BadSkill",
		"Illegal",
		"AbilityRestrictedByFlag",
		"AbilityRestrictedWhileJumping",
		"player.c:TooHeavy",
	}
	if abilityResultKeys4FB960 != wantKeys {
		t.Fatalf("keys = %#v, want %#v", abilityResultKeys4FB960, wantKeys)
	}

	for status, wantKey := range wantKeys {
		status := uint32(status)
		wantText := "translated:" + wantKey
		var events []string
		ok := abilityResult4FB960(status, abilityResultHooks4FB960{
			loadString: func(key, source string) string {
				events = append(events, fmt.Sprintf("load:%s:%s", key, source))
				return wantText
			},
			printCentered: func(text string) {
				events = append(events, "print:"+text)
			},
		})
		if !ok {
			t.Fatalf("status %d was rejected", status)
		}
		wantEvents := []string{
			fmt.Sprintf("load:%s:%s", wantKey, abilityResultSourcePath4FB960),
			"print:" + wantText,
		}
		if !reflect.DeepEqual(events, wantEvents) {
			t.Fatalf("status %d events = %v, want %v", status, events, wantEvents)
		}
	}
}

func TestAbilityResult4FB960RejectsUndefinedIndicesBeforeCallbacks(t *testing.T) {
	for _, status := range []uint32{uint32(len(abilityResultKeys4FB960)), math.MaxUint32} {
		calls := 0
		ok := abilityResult4FB960(status, abilityResultHooks4FB960{
			loadString: func(string, string) string {
				calls++
				return "unexpected"
			},
			printCentered: func(string) {
				calls++
			},
		})
		if ok || calls != 0 {
			t.Fatalf("status %#x = ok %t, callbacks %d; want false, 0", status, ok, calls)
		}
	}
}
