#include "GAME2_3.h"

extern int nox_client_copyRect_current_49F6F0(int, int, int, int);
extern int nox_client_clipHorizontal_49F780(int, int);

int4* nox_client_copyRect_49F6F0(int x, int y, int width, int height) {
	return (int4*)(uintptr_t)nox_client_copyRect_current_49F6F0(x, y, width, height);
}

int4* sub_49F780(int left, int right) {
	return (int4*)(uintptr_t)nox_client_clipHorizontal_49F780(left, right);
}
