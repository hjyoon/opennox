#ifndef NOX_MAP_OBJECT_LIST_5048A0_H
#define NOX_MAP_OBJECT_LIST_5048A0_H

#include <stddef.h>
#include <stdint.h>

typedef struct nox_object_t nox_object_t;

typedef struct nox_map_object_list_node_5048A0 {
	nox_object_t* object;
	struct nox_map_object_list_node_5048A0* next;
	struct nox_map_object_list_node_5048A0* previous;
} nox_map_object_list_node_5048A0;

_Static_assert(offsetof(nox_map_object_list_node_5048A0, object) == 0,
	"wrong offset of nox_map_object_list_node_5048A0.object");
_Static_assert(offsetof(nox_map_object_list_node_5048A0, next) == sizeof(void*),
	"wrong offset of nox_map_object_list_node_5048A0.next");
_Static_assert(offsetof(nox_map_object_list_node_5048A0, previous) == 2 * sizeof(void*),
	"wrong offset of nox_map_object_list_node_5048A0.previous");
_Static_assert(sizeof(nox_map_object_list_node_5048A0) == 3 * sizeof(void*),
	"wrong size of nox_map_object_list_node_5048A0");

nox_map_object_list_node_5048A0* nox_xxx_unitAddToList_5048A0(nox_object_t* object);
int32_t sub_504910(int32_t delta_x, int32_t delta_y);
nox_object_t* sub_504980(void);
nox_object_t* sub_5049C0(nox_object_t* object);
nox_map_object_list_node_5048A0* sub_5049D0(void);
nox_map_object_list_node_5048A0* sub_5049E0(nox_map_object_list_node_5048A0* node);
int32_t sub_504A10(nox_object_t* object);

nox_object_t* nox_map_object_list_node_object_5048A0(nox_map_object_list_node_5048A0* node);
nox_map_object_list_node_5048A0* nox_map_object_list_node_next_5048A0(nox_map_object_list_node_5048A0* node);
void nox_map_object_list_node_free_5048A0(nox_map_object_list_node_5048A0* node);

#endif // NOX_MAP_OBJECT_LIST_5048A0_H
