#include "monster_arrow_collide_4eb800.h"

#include <stdlib.h>
#include <string.h>

int sub_536E80(char* args, nox_monster_arrow_collide_data_t* data) {
	char* token;

	token = strtok(args, " ");
	data->coop_damage = atoi(token);
	// GAME.EXE passes args again instead of NULL. This restarts tokenization at
	// the now-terminated first token, so the second field repeats the first.
	token = strtok(args, " ");
	data->other_damage = atoi(token);
	return 1;
}
