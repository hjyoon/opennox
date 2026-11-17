#ifndef NOX_MONSTER_SPAWN_REGISTRY_50D780_H
#define NOX_MONSTER_SPAWN_REGISTRY_50D780_H
#include "defs.h"

typedef struct nox_monster_spawn_link {
	nox_object_t* object;
	struct nox_monster_spawn_link* older;
	struct nox_monster_spawn_link* newer;
} nox_monster_spawn_link;

typedef struct nox_monster_spawn_screen {
	nox_object_t* object;
	float average;
	uint32_t mask;
	float distances[32];
	struct nox_monster_spawn_screen* older;
	struct nox_monster_spawn_screen* newer;
} nox_monster_spawn_screen;

int nox_xxx_allocMonsterRelatedArrays_50D780(void);
void sub_50D7E0(void);
void sub_50D820(void);
unsigned int sub_50D890(void);
void sub_50D8D0(void);
int sub_50D960(void);
int sub_50DE10(void);
nox_monster_spawn_screen* sub_50DE60(nox_object_t* object);
int sub_50DE80(nox_object_t* generator, float2* point);
void sub_50DFB0(nox_object_t* object, nox_object_t* player);
int sub_50E030(nox_object_t* generator, nox_object_t* spawned);
nox_monster_spawn_link* sub_50E110(nox_monster_spawn_link* link);
void sub_50E140(nox_object_t* object);
nox_monster_spawn_link* sub_50E1B0(nox_monster_spawn_link* link);
void sub_50E1E0(nox_object_t* object);
void sub_50E210(nox_object_t* object);
#endif
