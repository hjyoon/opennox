#include "../object_extended_data_4f40a0.h"

typedef int8_t (*object_extended_data_fn_4F40A0)(nox_object_t*);

_Static_assert(_Generic(&sub_4F40A0, object_extended_data_fn_4F40A0: 1, default: 0),
	"sub_4F40A0 must preserve its native object pointer and signed byte result");

static object_extended_data_fn_4F40A0 const object_extended_data_signature_4F40A0 =
	sub_4F40A0;

int object_extended_data_header_test_4F40A0(void) {
	return object_extended_data_signature_4F40A0 != 0;
}
