#include "GAME2_3.h"

extern nox_render_data_t* nox_draw_curDrawData_3799572;
extern int nox_client_copyRect_native_49F6F0(nox_render_data_t*, int, int, int, int);
extern int nox_client_clipHorizontal_49F780(int, int);

int4* nox_client_copyRect_49F6F0(int x, int y, int width, int height) {
	return (int4*)(uintptr_t)nox_client_copyRect_native_49F6F0(nox_draw_curDrawData_3799572, x, y, width, height);
}

int4* sub_49F780(int left, int right) {
	return (int4*)(uintptr_t)nox_client_clipHorizontal_49F780(left, right);
}
