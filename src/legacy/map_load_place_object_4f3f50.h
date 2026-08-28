#ifndef NOX_MAP_LOAD_PLACE_OBJECT_4F3F50_H
#define NOX_MAP_LOAD_PLACE_OBJECT_4F3F50_H

#include <stddef.h>
#include <stdint.h>

typedef struct nox_object_t nox_object_t;

typedef struct nox_map_translation_4F3F50 {
	int32_t x;
	int32_t y;
} nox_map_translation_4F3F50;

_Static_assert(offsetof(nox_map_translation_4F3F50, x) == 0,
	"wrong offset of nox_map_translation_4F3F50.x");
_Static_assert(offsetof(nox_map_translation_4F3F50, y) == 4,
	"wrong offset of nox_map_translation_4F3F50.y");
_Static_assert(sizeof(nox_map_translation_4F3F50) == 8,
	"wrong size of nox_map_translation_4F3F50");

int32_t nox_xxx_servMapLoadPlaceObj_4F3F50(
	nox_object_t* object,
	nox_object_t* owner,
	nox_map_translation_4F3F50* translation);

#endif // NOX_MAP_LOAD_PLACE_OBJECT_4F3F50_H
