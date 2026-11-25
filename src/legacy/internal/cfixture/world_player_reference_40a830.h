#pragma once
#include <stdint.h>
#include <stddef.h>
uint32_t world_player_fixture_word(int reference,int entry,void* p,uint32_t arg);
void* world_player_fixture_pointer(int reference,int entry,void* p,void* q,int32_t a,int32_t b);
double world_player_fixture_armor(int reference,void* p);
uintptr_t world_player_fixture_layout(int field);
void world_player_fixture_tile_owner(void* head,int cap);
void world_player_fixture_modifier_next(void* p,void* next);
void* world_player_fixture_defend_callback(void);
int world_player_fixture_defend_seen(void* effect,void* item);
