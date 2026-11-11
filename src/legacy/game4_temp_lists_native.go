package legacy

/*
#include "game4_list_vote_native.h"
*/
import "C"

import "unsafe"

func game4WallHead() *C.nox_map_wall_list_node_504290 {
	return (*C.nox_map_wall_list_node_504290)(C.dword_5d4594_1599532)
}
func game4SetWallHead(node *C.nox_map_wall_list_node_504290) {
	C.dword_5d4594_1599532 = unsafe.Pointer(node)
}
func game4WaypointHead() *C.nox_map_waypoint_list_node_5044B0 {
	return (*C.nox_map_waypoint_list_node_5044B0)(C.dword_5d4594_1599548)
}
func game4SetWaypointHead(node *C.nox_map_waypoint_list_node_5044B0) {
	C.dword_5d4594_1599548 = unsafe.Pointer(node)
}

//export sub_504290
func sub_504290(x, y C.int8_t) *C.nox_map_wall_list_node_504290 {
	node := (*C.nox_map_wall_list_node_504290)(game4SupportAllocate(1, C.sizeof_nox_map_wall_list_node_504290))
	if node == nil {
		return nil
	}
	node.wall = (*C.nox_map_wall_504290)(game4SupportAllocate(1, C.sizeof_nox_map_wall_504290))
	if node.wall == nil {
		game4SupportFree(unsafe.Pointer(node))
		return nil
	}
	node.wall.x, node.wall.y = C.uint8_t(x), C.uint8_t(y)
	node.next = game4WallHead()
	if node.next != nil {
		node.next.previous = node
	}
	game4SetWallHead(node)
	return node
}

//export nox_xxx_cliWallGet_5042F0
func nox_xxx_cliWallGet_5042F0(x, y C.int32_t) *C.nox_map_wall_list_node_504290 {
	for node := game4WallHead(); node != nil; node = node.next {
		if C.int32_t(node.wall.x) == x && C.int32_t(node.wall.y) == y {
			return node
		}
	}
	return nil
}

//export sub_504330
func sub_504330(deltaX, deltaY C.int32_t) C.int32_t {
	for node := game4WallHead(); node != nil; node = node.next {
		source := node.wall
		// Both operands remain signed DWORDs before truncating division.
		x := (int32(deltaX) + 23*int32(source.x)) / 23
		y := (int32(deltaY) + 23*int32(source.y)) / 23
		wall := (*C.nox_map_wall_504290)(C.nox_server_getWallAtGrid_410580(C.int(x), C.int(y)))
		if wall != nil {
			if C.dword_5d4594_3835368 != 0 {
				wall.dir = C.nox_xxx_wall_42A6C0(wall.dir, source.dir)
			} else {
				wall.dir = source.dir
			}
		} else {
			wall = (*C.nox_map_wall_504290)(C.nox_xxx_wallCreateAt_410250(C.int(x), C.int(y)))
			if wall == nil {
				return 0
			}
			wall.dir = source.dir
		}
		wall.tile, wall.variation, wall.health = source.tile, source.variation, source.health
		if source.flags&0x80 != 0 {
			wall.flags |= 0x80
		}
		if wall.variation >= C.nox_xxx_mapWallMaxVariation_410DD0(C.int(wall.tile), C.int(wall.dir), 0) {
			wall.variation = 0
		}
		if wall.flags&0x04 != 0 {
			C.sub_4107A0(wall.data)
			wall.flags &^= 0x04
			wall.data = nil
		}
		if source.flags&0x04 != 0 {
			C.nox_server_wallAttachSecret(unsafe.Pointer(wall), source.data, source.field_10)
			C.nox_xxx_wallSecretBlock_410760((*C.nox_secret_wall_t)(source.data))
			source.data = nil
		}
		if source.flags&0x08 != 0 && wall.flags&0x08 == 0 {
			wall.flags |= 0x08
			C.nox_xxx_wallBreackableListAdd_410840(unsafe.Pointer(wall))
		}
		if source.flags&0x40 != 0 {
			wall.flags |= 0x40
		}
	}
	return 1
}

//export sub_5044B0
func sub_5044B0(index C.int32_t, x, y C.float) *C.nox_map_waypoint_list_node_5044B0 {
	node := (*C.nox_map_waypoint_list_node_5044B0)(game4SupportAllocate(1, C.sizeof_nox_map_waypoint_list_node_5044B0))
	if node == nil {
		return nil
	}
	node.waypoint = C.game4_waypoint_alloc_native()
	if node.waypoint == nil {
		game4SupportFree(unsafe.Pointer(node))
		return nil
	}
	node.next = game4WaypointHead()
	if node.next != nil {
		node.next.previous = node
	}
	game4SetWaypointHead(node)
	node.waypoint.ind = C.uint32_t(index)
	node.waypoint.pos.field_0, node.waypoint.pos.field_4 = x, y
	node.waypoint.prev = nil
	if node.next != nil {
		node.waypoint.next = node.next.waypoint
		node.next.waypoint.prev = node.waypoint
	} else {
		node.waypoint.next = nil
	}
	return node
}

//export sub_504560
func sub_504560(deltaX, deltaY C.int32_t) C.int32_t {
	x, y := C.float(deltaX), C.float(deltaY)
	for node := game4WaypointHead(); node != nil; node = node.next {
		node.waypoint.pos.field_0 += x
		node.waypoint.pos.field_4 += y
		C.sub_579E90(node.waypoint)
	}
	return 1
}

//export nox_map_wall_list_value_504290
func nox_map_wall_list_value_504290(node *C.nox_map_wall_list_node_504290) *C.nox_map_wall_504290 {
	if node == nil {
		return nil
	}
	return node.wall
}

//export nox_map_waypoint_list_value_5044B0
func nox_map_waypoint_list_value_5044B0(node *C.nox_map_waypoint_list_node_5044B0) *C.nox_waypoint_t {
	if node == nil {
		return nil
	}
	return node.waypoint
}

//export nox_mapgen_free_waypoint_list_503F40
func nox_mapgen_free_waypoint_list_503F40(freePayloads C.int32_t) {
	for node := game4WaypointHead(); node != nil; {
		next := node.next
		if freePayloads != 0 && node.waypoint != nil {
			game4SupportFree(unsafe.Pointer(node.waypoint))
		}
		game4SupportFree(unsafe.Pointer(node))
		node = next
	}
	game4SetWaypointHead(nil)
}

//export nox_mapgen_free_tile_list_503F40
func nox_mapgen_free_tile_list_503F40() {
	for entry := C.dword_5d4594_1599556; entry != nil; {
		next := entry.prev
		if entry.layer != nil {
			C.nox_xxx_tileFreeTile_422200(&entry.layer.subtiles)
			game4SupportFree(unsafe.Pointer(entry.layer))
		}
		game4SupportFree(unsafe.Pointer(entry))
		entry = next
	}
	C.dword_5d4594_1599556 = nil
}

//export nox_mapgen_free_wall_list_503F40
func nox_mapgen_free_wall_list_503F40(freePayloads C.int32_t) {
	for node := game4WallHead(); node != nil; {
		next := node.next
		if node.wall != nil {
			if freePayloads != 0 && node.wall.flags&0x04 != 0 {
				game4SupportFree(node.wall.data)
			}
			game4SupportFree(unsafe.Pointer(node.wall))
		}
		game4SupportFree(unsafe.Pointer(node))
		node = next
	}
	game4SetWallHead(nil)
}

//export nox_mapgen_adjust_waypoint_506260
func nox_mapgen_adjust_waypoint_506260(bounds unsafe.Pointer, position *C.float2) {
	if bounds == nil || position == nil {
		return
	}
	size := (*[2]C.int32_t)(unsafe.Pointer(C.game4_wall_size_native()))
	var origin C.int4
	C.sub_428170(bounds, &origin)
	x := float32(position.field_0) - float32(23*int32(size[0]))
	x = x + float32(origin.field_0)
	x = x - 11
	position.field_0 = C.float(x)
	y := float32(position.field_4) - float32(23*int32(size[1]))
	y = y + float32(origin.field_4)
	y = y - 11
	position.field_4 = C.float(y)
}
