#include "client__draw__drawrays.h"

bool noxClientTransientRayAdd(uintptr_t addr);
uintptr_t noxClientTransientRayAddress(size_t index);
int noxClientTransientRayContains(uintptr_t addr);
void noxClientTransientRaySetPayload(nox_drawable* dr, unsigned char* endpoints);

bool nox_client_transient_ray_add(nox_drawable* dr) {
	return noxClientTransientRayAdd((uintptr_t)dr);
}

nox_drawable* nox_client_transient_ray_at(size_t index) {
	return (nox_drawable*)noxClientTransientRayAddress(index);
}

int nox_client_transient_ray_contains(const nox_drawable* dr) {
	return noxClientTransientRayContains((uintptr_t)dr);
}

void nox_client_transient_ray_set_payload(nox_drawable* dr, const unsigned char* endpoints) {
	noxClientTransientRaySetPayload(dr, (unsigned char*)endpoints);
}
