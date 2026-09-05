package opennox

import (
	"math"
	"testing"

	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/server"
)

func TestSpellAcceptRoot4FD400DoesNotRequireDefinitionForDefaultHole(t *testing.T) {
	s := &Server{Server: new(server.Server)}
	second := new(server.Object)
	third := new(server.Object)
	arg := new(server.SpellAcceptArg)

	if got := s.SpellAccept4FD400(spell.ID(7), second, third, nil, arg, math.MinInt32); got != 1 {
		t.Fatalf("undefined selector hole result = %d, want 1", got)
	}
}

func TestSpellAcceptRoot4FD400RestoresTrivialSuccessCallbacks(t *testing.T) {
	s := &Server{Server: new(server.Server)}
	second := new(server.Object)
	third := new(server.Object)
	arg := new(server.SpellAcceptArg)

	for _, id := range []spell.ID{6, 18, 57, 63} {
		if got := s.SpellAccept4FD400(id, second, third, nil, arg, math.MaxInt32); got != 1 {
			t.Errorf("spell %d result = %d, want 1", id, got)
		}
	}
}
