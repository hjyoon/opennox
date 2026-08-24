// Command nativelayout reports native Go structure sizes that cross legacy C
// boundaries. It is intentionally C-free so it still runs while a C layout is
// being ported.
package main

import (
	"fmt"
	"unsafe"

	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/server"
)

func main() {
	for _, typ := range []struct {
		name string
		size uintptr
	}{
		{"gui.WindowData", unsafe.Sizeof(gui.WindowData{})},
		{"gui.Window", unsafe.Sizeof(gui.Window{})},
		{"gui.Anim", unsafe.Sizeof(gui.Anim{})},
		{"gui.ScrollListBoxItem", unsafe.Sizeof(gui.ScrollListBoxItem{})},
		{"gui.ScrollListBoxData", unsafe.Sizeof(gui.ScrollListBoxData{})},
		{"gui.EntryFieldData", unsafe.Sizeof(gui.EntryFieldData{})},
		{"binfile.MemFile", unsafe.Sizeof(binfile.MemFile{})},
		{"server.TileDef", unsafe.Sizeof(server.TileDef{})},
		{"server.MissileUpdateData", unsafe.Sizeof(server.MissileUpdateData{})},
		{"server.ElevatorUpdateData", unsafe.Sizeof(server.ElevatorUpdateData{})},
		{"server.MoverUpdateData", unsafe.Sizeof(server.MoverUpdateData{})},
		{"server.MonsterAnim", unsafe.Sizeof(server.MonsterAnim{})},
		{"server.Modifier", unsafe.Sizeof(server.Modifier{})},
		{"server.ModifierEff", unsafe.Sizeof(server.ModifierEff{})},
		{"server.Waypoint", unsafe.Sizeof(server.Waypoint{})},
		{"server.WallDef", unsafe.Sizeof(server.WallDef{})},
		{"server.MonsterUpdateData", unsafe.Sizeof(server.MonsterUpdateData{})},
		{"server.PlayerUpdateData", unsafe.Sizeof(server.PlayerUpdateData{})},
		{"server.ClassStats", unsafe.Sizeof(server.ClassStats{})},
	} {
		fmt.Printf("%-28s %d\n", typ.name, typ.size)
	}
}
