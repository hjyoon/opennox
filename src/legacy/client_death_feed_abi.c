#include "GAME2_3.h"

extern int nox_death_feed_draw_495500(int* entry);

int* sub_495500(int* entry) {
	int result = nox_death_feed_draw_495500(entry);
	return entry[2] ? (int*)(intptr_t)result : entry;
}
