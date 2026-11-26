#ifndef OPENNOX_CLIENT_NATIVE_STORAGE_H
#define OPENNOX_CLIENT_NATIVE_STORAGE_H

#include "defs.h"
#include "client__gui__guitrade.h"

typedef struct {
	int width;
	int height;
	uint32_t tile_width;
	uint32_t tile_stride;
	uint32_t tile_height;
	uint32_t tile_columns;
	uint32_t tile_rows;
	uint32_t edge_count;
	void * edge_images[64];
	void * health_change_allocator;
	void * health_change_head;
	void * health_change_font;
} nox_client_display_state_t;
_Static_assert(offsetof(nox_client_display_state_t, edge_images) == 8 * sizeof(uint32_t), "display image offset");
_Static_assert(sizeof(nox_client_display_state_t) == 8 * sizeof(uint32_t) + 67 * sizeof(void*), "display storage size");
nox_client_display_state_t* nox_client_display_storage(void);

typedef struct {
	void * effects_timer;
	void * music_timer;
	uint32_t music_state;
	uint32_t music_enabled;
	uint32_t music_idle;
	uint32_t music_initialized;
	uint32_t music_word_816368;
	uint32_t music_word_816372;
	uint32_t dialog_state;
	uint32_t dialog_flag;
	uint32_t dialog_initialized;
	uint32_t soundset_reference_head;
} nox_client_audio_state_t;
_Static_assert(offsetof(nox_client_audio_state_t, music_state) == 2 * sizeof(void*), "audio word offset");
nox_client_audio_state_t* nox_client_audio_storage(void);

typedef struct {
	nox_gui_trade_slot * trade_local;
	nox_gui_trade_slot * trade_remote;
	nox_gui_trade_slot * trade_source;
	nox_drawable * trade_dragged;
	uintptr_t secondary_selected;
	uint32_t secondary_previous;
	uint32_t equipped_item_code;
	uint32_t shop_scroll;
	uint32_t menu_word_815044;
	uint32_t center_message_index;
	uint32_t frame_sample_count;
} nox_client_item_state_t;
_Static_assert(offsetof(nox_client_item_state_t, secondary_previous) == 5 * sizeof(void*), "item words offset");
nox_client_item_state_t* nox_client_item_storage(void);

#endif
