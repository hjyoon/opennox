package opennox

import (
	"testing"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func TestFinishPlayerIncomingNative4DDF60(t *testing.T) {
	plUnit := &server.Object{PosVec: types.Pointf{X: 123.5, Y: -45.25}}
	pl := &server.Player{
		PlayerUnit: plUnit,
		Field3676:  2,
		Field4700:  1,
		Pos3632Vec: types.Pointf{X: 1, Y: 2},
	}
	finishPlayerIncomingNative4DDF60(pl, false)
	if pl.Field3676 != 3 || pl.Field4700 != 0 || pl.Pos3632Vec != plUnit.PosVec {
		t.Fatalf("incoming state = phase %d, field4700 %d, position %v", pl.Field3676, pl.Field4700, pl.Pos3632Vec)
	}
}

func TestFinishPlayerIncomingNative4DDF60CoopTeamKeepsPosition(t *testing.T) {
	pl := &server.Player{
		PlayerUnit: &server.Object{PosVec: types.Pointf{X: 123.5, Y: -45.25}},
		Field3676:  2,
		Field4700:  1,
		Pos3632Vec: types.Pointf{X: 1, Y: 2},
	}
	finishPlayerIncomingNative4DDF60(pl, true)
	if pl.Field3676 != 3 || pl.Field4700 != 0 || pl.Pos3632Vec != (types.Pointf{X: 1, Y: 2}) {
		t.Fatalf("incoming coop-team state = phase %d, field4700 %d, position %v", pl.Field3676, pl.Field4700, pl.Pos3632Vec)
	}

	finishPlayerIncomingNative4DDF60(nil, false)
}
