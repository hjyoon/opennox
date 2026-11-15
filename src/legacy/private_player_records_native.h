#ifndef NOX_PRIVATE_PLAYER_RECORDS_NATIVE_H
#define NOX_PRIVATE_PLAYER_RECORDS_NATIVE_H

#include "defs.h"

/* The only legacy payload consumer must share the native record layout. */
typedef struct nox_private_player_record_t {
	nox_list_item_t link;
	wchar2_t name[10];
	int32_t id;
	void* team;
	nox_list_item_t children;
} nox_private_player_record_t;

typedef struct nox_private_player_child_t {
	nox_list_item_t link;
	int32_t peer;
} nox_private_player_child_t;

#endif
