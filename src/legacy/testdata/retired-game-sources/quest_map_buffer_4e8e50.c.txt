#include "quest_map_buffer_4e8e50.h"

#include <stdint.h>

#include "memmap.h"

enum {
	quest_map_buffer_base_4e8e50 = 0x5D4594,
	quest_map_buffer_offset_4e8e50 = 1567844,
};

//----- (004E8E50) --------------------------------------------------------
char* sub_4E8E50(void) {
	return (char*)getMemAt((uintptr_t)quest_map_buffer_base_4e8e50, (uintptr_t)quest_map_buffer_offset_4e8e50);
}
