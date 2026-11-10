#include "game3_class_config_options.h"

// Const transport and the one mixed pointer/scalar legacy return are ABI only.
unsigned char* sub_4A61E0(nox_window* w, int group, unsigned char* rgb) {
	return (unsigned char*)game3_selcolor_match_4A61E0(w, group, rgb);
}
int sub_4A7A80(const char* s) { return game3_config_copy_4A7A80((char*)s); }
int sub_4A7AC0(const char* s) { return game3_config_copy_4A7AC0((char*)s); }
int sub_4A7B00(const char* s) { return game3_config_copy_4A7B00((char*)s); }
int sub_4A7BC0(const char* s) { return game3_config_copy_4A7BC0((char*)s); }
int sub_4A7C00(const char* s) { return game3_config_copy_4A7C00((char*)s); }
int sub_4A7D00(const char* s) { return game3_config_copy_4A7D00((char*)s); }
