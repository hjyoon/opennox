#include "game3_class_config_options.h"

// Preserve the legacy pointer-shaped scalar result at the C boundary.
unsigned char* sub_4A61E0(nox_window* w, int group, unsigned char* rgb) {
	return (unsigned char*)game3_selcolor_match_4A61E0(w, group, rgb);
}
