package server

import (
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestGruntInit4F0360(t *testing.T) {
	// The original RET does not dereference its argument.
	GruntInit4F0360(nil)

	unit := &Object{
		TypeInd:     0x1234,
		ObjClass:    object.Class(0x89abcdef),
		ObjSubClass: object.SubClass(0x76543210),
		ObjFlags:    object.Flags(0xfedcba98),
		PosVec:      types.Pointf{X: -12.5, Y: 37.25},
		Direction1:  Dir16(0x7f01),
		SpeedBase:   4.75,
		Field188:    0xa5a55a5a,
	}
	before := *unit
	GruntInit4F0360(unit)
	if *unit != before {
		t.Fatal("single-RET GruntInit mutated its object argument")
	}
}
