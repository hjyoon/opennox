#ifndef NOX_PORT_FLOOR_BRUSH_NATIVE_H
#define NOX_PORT_FLOOR_BRUSH_NATIVE_H

#include <stddef.h>
#include <stdint.h>

// Scalar brush state passed by address in GAME.EXE. This record contains no
// host pointers and retains its original32-byte layout on every architecture.
typedef struct nox_floor_brush_t {
    uint32_t tile_id;
    uint32_t variation;
    uint32_t pattern;
    int32_t anchor_x;
    int32_t anchor_y;
    uint8_t kind;
    uint8_t reserved[3];
    uint32_t edge_id;
    uint32_t edge_variation;
} nox_floor_brush_t;
_Static_assert(sizeof(nox_floor_brush_t) ==32, "wrong floor brush size");
_Static_assert(offsetof(nox_floor_brush_t, kind) ==20, "wrong floor brush kind offset");
_Static_assert(offsetof(nox_floor_brush_t, edge_id) ==24, "wrong floor brush edge offset");

// Native brush entry; GAME4_2.h retains the separate integer-token ABI.
int nox_floorBrushApplyNative_51DA70(int x, int y, nox_floor_brush_t* brush, int layer, int merge);

#endif
