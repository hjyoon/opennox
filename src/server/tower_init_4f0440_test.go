package server

import (
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestTowerInit4F0440(t *testing.T) {
	// The original RET does not dereference its argument.
	TowerInit4F0440(nil)

	unit := &Object{
		TypeInd:     0x6b6b,
		ObjClass:    object.Class(0x13579bdf),
		ObjSubClass: object.SubClass(0x2468ace0),
		ObjFlags:    object.Flags(0x89abcdef),
		PosVec:      types.Pointf{X: -91.125, Y: 27.75},
		Direction1:  Dir16(0x4e20),
		SpeedBase:   23.5,
		Field188:    0x96963c3c,
	}
	before := *unit
	TowerInit4F0440(unit)
	if *unit != before {
		t.Fatal("single-RET TowerInit mutated its object argument")
	}
}
