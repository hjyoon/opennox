#include "../drop_owned_crowns_4ed050.h"

static void (*const drop_owned_crowns_signature_4ed050)(
	nox_object_t*, nox_object_t*) = sub_4ED050;

int drop_owned_crowns_abi_test(void) {
	return drop_owned_crowns_signature_4ed050 != 0;
}
