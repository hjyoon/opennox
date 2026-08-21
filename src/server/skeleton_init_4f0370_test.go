package server

import (
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestSkeletonInit4F0370(t *testing.T) {
	// The original RET does not dereference its argument.
	SkeletonInit4F0370(nil)

	unit := &Object{
		TypeInd:     0x4321,
		ObjClass:    object.Class(0xfedcba98),
		ObjSubClass: object.SubClass(0x01234567),
		ObjFlags:    object.Flags(0x89abcdef),
		PosVec:      types.Pointf{X: 91.75, Y: -43.125},
		Direction1:  Dir16(0x01f7),
		SpeedBase:   8.25,
		Field188:    0x5aa5a55a,
	}
	before := *unit
	SkeletonInit4F0370(unit)
	if *unit != before {
		t.Fatal("single-RET SkeletonInit mutated its object argument")
	}
}
