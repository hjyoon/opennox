package server

import (
	"testing"
	"unsafe"
)

func TestWallDrawScalarPrefixLayout(t *testing.T) {
	fields := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Dir0", unsafe.Offsetof(Wall{}.Dir0), 0},
		{"Tile1", unsafe.Offsetof(Wall{}.Tile1), 1},
		{"Field2", unsafe.Offsetof(Wall{}.Field2), 2},
		{"Field3", unsafe.Offsetof(Wall{}.Field3), 3},
		{"Flags4", unsafe.Offsetof(Wall{}.Flags4), 4},
		{"X5", unsafe.Offsetof(Wall{}.X5), 5},
		{"Y6", unsafe.Offsetof(Wall{}.Y6), 6},
		{"Health7", unsafe.Offsetof(Wall{}.Health7), 7},
		{"Field8", unsafe.Offsetof(Wall{}.Field8), 8},
		{"Field10", unsafe.Offsetof(Wall{}.Field10), 10},
		{"Field12", unsafe.Offsetof(Wall{}.Field12), 12},
		{"NextByPos16", unsafe.Offsetof(Wall{}.NextByPos16), 16},
	}
	for _, field := range fields {
		if field.got != field.want {
			t.Errorf("Wall.%s offset = %d, want %d", field.name, field.got, field.want)
		}
	}
}
