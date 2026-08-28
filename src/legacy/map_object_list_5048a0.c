#include "map_object_list_5048a0.h"

#include <stdlib.h>

#include "defs.h"

extern uint32_t dword_5d4594_1599476;
extern uint32_t dword_5d4594_1599480;
extern void* dword_5d4594_1599540;
extern uint32_t dword_5d4594_3835396;

int nox_xxx_mapgenSaveMap_503830(int32_t map_index);
void nox_xxx_createAt_4DAA50(nox_object_t* object, nox_object_t* owner, float x, float y);
int nox_xxx_objectFreeMem_4E38A0(nox_object_t* object);

static nox_map_object_list_node_5048A0* map_object_list_head_5048A0(void) {
	return (nox_map_object_list_node_5048A0*)dword_5d4594_1599540;
}

static void map_object_list_set_head_5048A0(nox_map_object_list_node_5048A0* node) {
	dword_5d4594_1599540 = node;
}

nox_map_object_list_node_5048A0* nox_xxx_unitAddToList_5048A0(nox_object_t* object) {
	nox_map_object_list_node_5048A0* node = calloc(1, sizeof(*node));
	if (node == NULL) {
		return NULL;
	}
	node->previous = NULL;
	node->object = object;
	node->next = map_object_list_head_5048A0();
	if (node->next != NULL) {
		node->next->previous = node;
	}
	map_object_list_set_head_5048A0(node);

	object->object_prev = NULL;
	if (node->next != NULL) {
		object->object_next = node->next->object;
		node->next->object->object_prev = object;
	} else {
		object->object_next = NULL;
	}
	return node;
}

int32_t sub_504910(int32_t delta_x, int32_t delta_y) {
	const float x_offset = (float)delta_x;
	const float y_offset = (float)delta_y;
	for (nox_map_object_list_node_5048A0* node = map_object_list_head_5048A0(); node != NULL;
		 node = node->next) {
		nox_object_t* object = node->object;
		object->x = object->x + x_offset;
		object = node->object;
		object->y = object->y + y_offset;
		object = node->object;
		const float y = object->y;
		const float x = object->x;
		nox_xxx_createAt_4DAA50(object, NULL, x, y);
		object = node->object;
		object->obj_flags |= UINT32_C(0x80000000);
	}
	return 1;
}

nox_object_t* sub_504980(void) {
	if (!((dword_5d4594_1599480 == dword_5d4594_3835396 &&
		  (int32_t)dword_5d4594_1599480 != -1 && dword_5d4594_1599476 != 1) ||
		 nox_xxx_mapgenSaveMap_503830((int32_t)dword_5d4594_3835396))) {
		return NULL;
	}
	nox_map_object_list_node_5048A0* node = map_object_list_head_5048A0();
	if (node == NULL) {
		return NULL;
	}
	return node->object;
}

nox_object_t* sub_5049C0(nox_object_t* object) {
	if (object == NULL) {
		return NULL;
	}
	return object->object_next;
}

nox_map_object_list_node_5048A0* sub_5049D0(void) {
	return map_object_list_head_5048A0();
}

nox_map_object_list_node_5048A0* sub_5049E0(nox_map_object_list_node_5048A0* node) {
	if (node == NULL) {
		return NULL;
	}
	return node->next;
}

int32_t sub_504A10(nox_object_t* object) {
	if (object == NULL) {
		return 0;
	}
	nox_map_object_list_node_5048A0* node = map_object_list_head_5048A0();
	while (node != NULL && node->object != object) {
		node = node->next;
	}
	if (node == NULL) {
		return 0;
	}

	if (node->next != NULL) {
		node->next->previous = node->previous;
	}
	if (node->previous != NULL) {
		node->previous->next = node->next;
	}
	if (node == map_object_list_head_5048A0()) {
		map_object_list_set_head_5048A0(node->next);
	}
	if (object->object_next != NULL) {
		object->object_next->object_prev = object->object_prev;
	}
	if (object->object_prev != NULL) {
		object->object_prev->object_next = object->object_next;
	}
	nox_xxx_objectFreeMem_4E38A0(object);
	free(node);
	return 1;
}

nox_object_t* nox_map_object_list_node_object_5048A0(nox_map_object_list_node_5048A0* node) {
	return node->object;
}

nox_map_object_list_node_5048A0* nox_map_object_list_node_next_5048A0(nox_map_object_list_node_5048A0* node) {
	return node->next;
}

void nox_map_object_list_node_free_5048A0(nox_map_object_list_node_5048A0* node) {
	free(node);
}
