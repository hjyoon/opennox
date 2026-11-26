#ifndef OPENNOX_GAME4_MAPGEN_STORAGE_FIXTURE_H
#define OPENNOX_GAME4_MAPGEN_STORAGE_FIXTURE_H

#include "game4_mapgen_state.h"

// Supply the live storage to C ABI probes and retained C references.
#define nox_file_8 (nox_game4_mapgen_storage()->archive)
#define dword_5d4594_1599480 (nox_game4_mapgen_storage()->loaded_index)
#define dword_5d4594_1599476 (nox_game4_mapgen_storage()->tiles_placed)
#define dword_5d4594_1599540 (nox_game4_mapgen_storage()->object_head)
#define dword_5d4594_1599532 (nox_game4_mapgen_storage()->wall_head)
#define dword_5d4594_1599556 (nox_game4_mapgen_storage()->tile_head)
#define dword_5d4594_1599548 (nox_game4_mapgen_storage()->waypoint_head)
#define dword_5d4594_1599588 (nox_game4_mapgen_storage()->archive_path)
#define dword_5d4594_1599592 (nox_game4_mapgen_storage()->staging_path)

#endif
