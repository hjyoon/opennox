#include "client__draw__slavedraw.h"
#include "GAME3_1.h"
#include "client__draw__parse__parse.h"
#include "client__draw__staticdraw.h"

//----- (004BBD30) --------------------------------------------------------
int nox_thing_slave_draw(int* a1, nox_drawable* dr) {
	nox_static_random_draw_data_t* data = dr->field_76;
	if (data && data->images && dr->field_77 < data->count) {
		nox_xxx_drawObject_4C4770_draw(a1, dr, data->images[dr->field_77]);
	}
	if (nox_thing_slave_draw == nox_thing_static_random_draw) // AntiICFoptimization
	{
		return 0;
	}
	return 1;
}

//----- (0044C120) --------------------------------------------------------
bool nox_things_slave_draw_parse(nox_thing* obj, nox_memfile* f, char* attr_value) {
	obj->draw_func = nox_thing_slave_draw;
	void* v3 = nox_xxx_spriteLoadStaticRandomData_44C000(attr_value, f);
	obj->field_5c = v3;
	obj->field_60 = 0;
	return v3 != 0;
}
