//go:build !server

package opennox

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/server"
)

func TestShouldDrawInvisibleCreatureEffects(t *testing.T) {
	const (
		localCode  = 10
		targetCode = 20
	)
	local := &client.Drawable{NetCode32: localCode}
	target := &client.Drawable{NetCode32: targetCode}
	teamOne := &server.ObjectTeam{ID: 1}
	teamTwo := &server.ObjectTeam{ID: 2}

	tests := []struct {
		name  string
		dr    *client.Drawable
		local *client.Drawable
		teams map[int]*server.ObjectTeam
		buffs uint32
		want  bool
	}{
		{
			name:  "local drawable",
			dr:    local,
			local: local,
			want:  true,
		},
		{
			name:  "infravision",
			dr:    target,
			local: local,
			buffs: 1 << server.ENCHANT_INFRAVISION,
			want:  true,
		},
		{
			name:  "same team",
			dr:    target,
			local: local,
			teams: map[int]*server.ObjectTeam{localCode: teamOne, targetCode: teamOne},
			want:  true,
		},
		{
			name:  "different teams",
			dr:    target,
			local: local,
			teams: map[int]*server.ObjectTeam{localCode: teamOne, targetCode: teamTwo},
			want:  false,
		},
		{
			name:  "missing target team",
			dr:    target,
			local: local,
			teams: map[int]*server.ObjectTeam{localCode: teamOne},
			want:  false,
		},
		{
			name: "no local drawable",
			dr:   target,
			want: false,
		},
		{
			name: "nil target",
			want: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.local != nil {
				test.local.Buffs = test.buffs
				defer func() { test.local.Buffs = 0 }()
			}
			lookup := func(code int) *server.ObjectTeam {
				return test.teams[code]
			}
			if got := shouldDrawInvisibleCreatureEffects(test.dr, test.local, localCode, lookup); got != test.want {
				t.Fatalf("should draw = %v, want %v", got, test.want)
			}
		})
	}
}

func TestShouldDrawInvisibleCreatureEffectsPreservesNativePointers(t *testing.T) {
	local := &client.Drawable{NetCode32: 10}
	target := &client.Drawable{NetCode32: 20}
	var pin runtime.Pinner
	defer pin.Unpin()
	for _, ptr := range []unsafe.Pointer{unsafe.Pointer(local), unsafe.Pointer(target)} {
		pin.Pin(ptr)
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= math.MaxUint32 {
			t.Fatalf("drawable pointer = %p, want native address above 4 GiB", ptr)
		}
	}

	teamOne := &server.ObjectTeam{ID: 1}
	teamTwo := &server.ObjectTeam{ID: 2}
	lookup := func(code int) *server.ObjectTeam {
		if code == int(local.NetCode32) {
			return teamOne
		}
		return teamTwo
	}
	if shouldDrawInvisibleCreatureEffects(target, local, int(local.NetCode32), lookup) {
		t.Fatal("enemy invisible creature effects should remain hidden")
	}

	local.Buffs = 1 << server.ENCHANT_INFRAVISION
	if !shouldDrawInvisibleCreatureEffects(target, local, int(local.NetCode32), lookup) {
		t.Fatal("infravision should reveal invisible creature effects")
	}
	runtime.KeepAlive(local)
	runtime.KeepAlive(target)
}
