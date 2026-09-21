#ifndef NOX_PORT_CLIENT_DRAW_DRAWRAYS
#define NOX_PORT_CLIENT_DRAW_DRAWRAYS

#include "defs.h"

#define NOX_CLIENT_TRANSIENT_RAY_CAPACITY 96

nox_drawable* nox_xxx_netDrawRays_49BDD0(unsigned char* data);
void nox_client_transient_ray_set_payload(nox_drawable* dr, const unsigned char* endpoints);
bool nox_client_transient_ray_add(nox_drawable* dr);
size_t nox_client_transient_ray_count(void);
nox_drawable* nox_client_transient_ray_at(size_t index);
int nox_client_transient_ray_contains(const nox_drawable* dr);
void nox_client_transient_ray_clear(void);

#endif // NOX_PORT_CLIENT_DRAW_DRAWRAYS
