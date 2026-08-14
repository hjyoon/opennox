#include "damage_collide_4e9430.h"

#include <stdlib.h>
#include <string.h>

int nox_xxx_parseDamageTypeByName_4E0A00(const char* name);

int nox_xxx_collideDamageLoad_536E10(char* args, nox_damage_collide_data_t* data) {
	char* token = strtok(args, " ");
	data->damage = (uint8_t)atoi(token);
	token = strtok(NULL, " ");
	data->damage_type = nox_xxx_parseDamageTypeByName_4E0A00(token);
	return data->damage_type != 18;
}
