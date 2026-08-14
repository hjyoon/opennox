#include "mana_drain_collide_4e9490.h"

#include <stdlib.h>
#include <string.h>

int nox_xxx_collideManaDrainLoad_536E50(
	char* args,
	nox_mana_drain_collide_data_t* data) {
	char* token = strtok(args, " ");
	data->amount = (uint8_t)atoi(token);
	return 1;
}
