package cfixture

/*
#include "GAME2_2.h"

static uintptr_t floor_edge_positions[2];
static uintptr_t floor_edge_nodes[2];
static int floor_edge_calls;

static void floor_capture_edge(int2* pos, nox_tile_list_node_t* edge) {
    int index = floor_edge_calls++;
    if (index < 2) {
        floor_edge_positions[index] = (uintptr_t)pos;
        floor_edge_nodes[index] = (uintptr_t)edge;
    }
}

static void* floor_edge_callback(void) {
    return (void*)floor_capture_edge;
}

static int floor_test_call_edges(void) {
    int2 pos = {-19, 31};
    nox_tile_list_node_t edges[2] = {{0}};
    edges[0].next = &edges[1];
    floor_edge_calls = 0;
    nox_xxx_tileCallDrawEdges_481BF0(&pos, edges);
    if (floor_edge_calls == 0) {
        return 0;
    }
    if (floor_edge_calls != 2 ||
        floor_edge_positions[0] != (uintptr_t)&pos || floor_edge_positions[1] != (uintptr_t)&pos ||
        floor_edge_nodes[0] != (uintptr_t)&edges[0] || floor_edge_nodes[1] != (uintptr_t)&edges[1]) {
        return -1;
    }
    return floor_edge_calls;
}
*/
import "C"

import "unsafe"

func FloorEdgeCallback() unsafe.Pointer {
	return C.floor_edge_callback()
}

func FloorEdgeCallbackCalls() int {
	return int(C.floor_test_call_edges())
}
