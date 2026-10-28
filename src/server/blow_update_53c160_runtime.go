package server

import (
	"image"

	"github.com/opennox/libs/types"
)

// BlowUpdateRuntime53C160 supplies native-width world boundaries and the live
// direction-table reads. The original mass getter is the object's float field.
type BlowUpdateRuntime53C160 struct {
	IndexedDirection func(int16) image.Point
	EachInRect       func(types.Rectf, func(*Object) bool)
	CanInteract      func(*Object, *Object) bool
	DirectionFloat   func(int32) float32
}
