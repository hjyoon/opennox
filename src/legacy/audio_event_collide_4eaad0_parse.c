#include "audio_event_collide_4eaad0.h"

#include <stdio.h>

int nox_xxx_utilFindSound_40AF50(char* name);

int sub_536DA0(char* args, nox_audio_event_collide_data_t* data) {
	char name[256];
	int sound;

	(void)sscanf(args, "%s", name);
	sound = nox_xxx_utilFindSound_40AF50(name);
	data->sound = (uint32_t)sound;
	return sound != 0;
}
