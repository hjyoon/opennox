package server

import (
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestProjectileInit4F0380(t *testing.T) {
	// The original RET does not dereference its argument.
	ProjectileInit4F0380(nil)

	unit := &Object{
		TypeInd:     0x5a5a,
		ObjClass:    object.Class(0x76543210),
		ObjSubClass: object.SubClass(0xfedcba98),
		ObjFlags:    object.Flags(0x01234567),
		PosVec:      types.Pointf{X: -77.5, Y: 14.625},
		Direction1:  Dir16(0x5f10),
		SpeedBase:   19.75,
		Field188:    0xc33c9669,
	}
	before := *unit
	ProjectileInit4F0380(unit)
	if *unit != before {
		t.Fatal("single-RET ProjectileInit mutated its object argument")
	}
}
