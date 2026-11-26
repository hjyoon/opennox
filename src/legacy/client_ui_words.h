#ifndef NOX_CLIENT_UI_WORDS_H
#define NOX_CLIENT_UI_WORDS_H

#include <stdint.h>

typedef struct {
	unsigned int wol_status;
	uint32_t mouse_primary;
	uint32_t audio_mask;
	uint32_t wol_game_data_ready;
	uint32_t create_or_join;
	uint32_t network_state;
	uint32_t player_gold;
	uint32_t shop_active;
	uint32_t shop_mode;
	uint32_t dialog_active;
	unsigned int render_bubbles;
} nox_client_ui_words_t;
_Static_assert(sizeof(nox_client_ui_words_t) == 11 * sizeof(uint32_t), "client UI word widths");
nox_client_ui_words_t* nox_client_ui_words_storage(void);

uint32_t* nox_client_chat_active_storage(void);
uint32_t* nox_client_meter_visible_storage(void);


#endif
