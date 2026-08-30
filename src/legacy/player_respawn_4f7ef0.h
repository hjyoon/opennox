#ifndef NOX_PLAYER_RESPAWN_4F7EF0_H
#define NOX_PLAYER_RESPAWN_4F7EF0_H

#include <stdint.h>

typedef struct float2 float2;
typedef struct nox_object_t nox_object_t;

int16_t nox_xxx_playerRespawn_4F7EF0(nox_object_t* unit);
int32_t sub_4F80C0(nox_object_t* gate, float2* output);
void nox_xxx_respawnPlayerImpl_53FBC0(float2* center, int32_t direction);

#endif // NOX_PLAYER_RESPAWN_4F7EF0_H
