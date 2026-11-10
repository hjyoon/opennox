package legacy

/*
#include "GAME3.h"
#include <string.h>

// The reference body is unchanged. These services record its actual ordering,
// coordinates, colors and final return without invoking the renderer.
static int game3_ref_colors[2] = {0x1723, 0x4567};
static int game3_ref_current, game3_ref_count, game3_ref_points;
static int game3_ref_xy[4], game3_ref_rows[15];
static int* game3_ref_mem(int blob, int off) { return &game3_ref_colors[off == 1312496]; }
static void game3_ref_color(int value) { game3_ref_current = value; }
static void game3_ref_point(int x, int y) {
	if (game3_ref_points < 2) {
		game3_ref_xy[2*game3_ref_points] = x;
		game3_ref_xy[2*game3_ref_points+1] = y;
	}
	game3_ref_points++;
}
static int game3_ref_draw(void) {
	if (game3_ref_count < 3) {
		memcpy(&game3_ref_rows[5*game3_ref_count], game3_ref_xy, sizeof(game3_ref_xy));
		game3_ref_rows[5*game3_ref_count+4] = game3_ref_current;
	}
	game3_ref_points = 0;
	return 0x100 + ++game3_ref_count;
}
#undef getMemIntPtr
#define getMemIntPtr game3_ref_mem
#define nox_client_drawSetColor_434460 game3_ref_color
#define nox_client_drawAddPoint_49F500 game3_ref_point
#define nox_client_drawLineFromPoints_49E4B0 game3_ref_draw
#define sub_4B63B0 game3_ref_original_4B63B0
static int game3_ref_original_4B63B0(int2*, int2*);
#include "testdata/game3_harpoon_line_original.h"
#undef sub_4B63B0
#undef nox_client_drawLineFromPoints_49E4B0
#undef nox_client_drawAddPoint_49F500
#undef nox_client_drawSetColor_434460
#undef getMemIntPtr

static int game3_ref_line_call(int x1, int y1, int x2, int y2, int* out, uintptr_t* address) {
	struct { uint64_t before; int2 a, b; uint64_t after; } args;
	memset(&args, 0, sizeof(args));
	args.before = args.after = UINT64_C(0x123456789abcdef0);
	args.a.field_0=x1; args.a.field_4=y1;
	args.b.field_0=x2; args.b.field_4=y2;
	*address = (uintptr_t)&args.a;
	game3_ref_count=game3_ref_points=0;
	int result = game3_ref_original_4B63B0(&args.a, &args.b);
	memcpy(out, game3_ref_rows, sizeof(game3_ref_rows));
	if (args.before != UINT64_C(0x123456789abcdef0) || args.after != args.before ||
		args.a.field_0 != x1 || args.a.field_4 != y1 || args.b.field_0 != x2 || args.b.field_4 != y2)
		return -1;
	return result;
}
*/
import "C"

func game3HarpoonOriginalControl(a, b game3LinePoint) ([15]int32, int32, uintptr) {
	var values [15]C.int
	var address C.uintptr_t
	result := C.game3_ref_line_call(C.int(a.x), C.int(a.y), C.int(b.x), C.int(b.y), &values[0], &address)
	var rows [15]int32
	for i, value := range values {
		rows[i] = int32(value)
	}
	return rows, int32(result), uintptr(address)
}
