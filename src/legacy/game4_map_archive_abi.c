#include "game4_map_archive.h"

FILE* sub_502DA0(char* name) {
	// This historical scalar success token is not the actual FILE* owner.
	// Keep address 1 out of Go's pointer roots; nox_file_8 is the real handle.
	return game4MapOpen(name) ? (FILE*)1 : NULL;
}
