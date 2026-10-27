#include <assert.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "defs.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "poison_state_4ee7e0.h"

static struct poison_effect { uint32_t before[8]; int32_t power; uint32_t after[8]; } effect;
static nox_object_t target,source,weapon;
static nox_player_update_data_t update,live_update;
static nox_playerInfo player,live_player;
static int facing_result,activate_result,facing_mutation,activate_mutation;
static int facing_calls,activate_calls,count;
static struct { const char* service; uint32_t direction,maximum,tx,ty,sx,sy; int index,value; } events[4];
static void event(const char* name) { assert(count<4);events[count++].service=name; }
static void mutate(int stage,int mutation) {
    if (mutation==1) target.obj_class=0;
    else if (mutation==2) target.obj_class=stage?4:2;
    else if (mutation==3) target.data_update=&live_update;
    else if (mutation==4) { effect.power=-129;target.x=77.5f;source.y=88.25f; }
}
int32_t nox_modifier_effect_getPreDamageInt(void* ptr) { assert(ptr==&effect);return effect.power; }
int nox_server_testTwoPointsAndDirection_4E6E50(float2* a,int dir,float2* b) {
    assert(a==(float2*)&target.x && b==(float2*)&source.x);
    event("facing");events[count-1].direction=(uint32_t)dir;
    memcpy(&events[count-1].tx,&target.x,4);memcpy(&events[count-1].ty,&target.y,4);
    memcpy(&events[count-1].sx,&source.x,4);memcpy(&events[count-1].sy,&source.y,4);
    facing_calls++;mutate(0,facing_mutation);return facing_result;
}
int32_t nox_xxx_activatePoison_4EE7E0(nox_object_t* obj,int32_t increment,int32_t maximum) {
    assert(obj==&target && increment==1);
    event("activate");events[count-1].maximum=(uint32_t)maximum;
    activate_calls++;mutate(1,activate_mutation);return activate_result;
}
int nox_xxx_netInformTextMsg_4DA0F0(int index,int kind,int* value) {
    assert(kind==13);event("inform");events[count-1].index=index;events[count-1].value=*value;return 0;
}

// PRODUCTION_PLAYER
// PRODUCTION_INFORM
// PRODUCTION_POISON

int main(void) {
    int cls,state,dir,power;
    while (scanf("%d %d %d %d %d %d %d %d",&cls,&state,&dir,&power,&facing_result,&activate_result,&facing_mutation,&activate_mutation)==8) {
        memset(&effect,0xa5,sizeof(effect));effect.power=power;
        memset(&target,0x96,sizeof(target));memset(&source,0x69,sizeof(source));memset(&weapon,0xc3,sizeof(weapon));
        memset(&update,0,sizeof(update));memset(&live_update,0,sizeof(live_update));memset(&player,0,sizeof(player));memset(&live_player,0,sizeof(live_player));
        player.playerInd=254;live_player.playerInd=130;update.player=&player;live_update.player=&live_player;update.state=state;
        target.obj_class=cls;target.direction1=(uint16_t)dir;target.data_update=&update;target.x=-12.75f;
        uint32_t zero=0x80000000;memcpy(&target.y,&zero,4);source.x=29.5f;source.y=-4.25f;
        unsigned char expected_effect[sizeof(effect)],expected_target[sizeof(target)],expected_source[sizeof(source)],expected_weapon[sizeof(weapon)],expected_update[sizeof(update)],expected_live_update[sizeof(live_update)],expected_player[sizeof(player)],expected_live_player[sizeof(live_player)];
        memcpy(expected_effect,&effect,sizeof(effect));memcpy(expected_target,&target,sizeof(target));memcpy(expected_source,&source,sizeof(source));memcpy(expected_weapon,&weapon,sizeof(weapon));
        memcpy(expected_update,&update,sizeof(update));memcpy(expected_live_update,&live_update,sizeof(live_update));memcpy(expected_player,&player,sizeof(player));memcpy(expected_live_player,&live_player,sizeof(live_player));
        if (sizeof(void*)>4) assert((uintptr_t)&effect>UINT32_MAX && (uintptr_t)&target>UINT32_MAX && (uintptr_t)&source>UINT32_MAX && (uintptr_t)&weapon>UINT32_MAX && (uintptr_t)&update>UINT32_MAX && (uintptr_t)&player>UINT32_MAX);
        count=facing_calls=activate_calls=0;memset(events,0,sizeof(events));
        nox_xxx_poisonEffect_4E0850(&effect,&weapon,&source,&target,NULL);
        for (int stage=0;stage<2;stage++) {
            if (!(stage?activate_calls:facing_calls)) continue;
            int mutation=stage?activate_mutation:facing_mutation;
            if (mutation==1 || mutation==2) { uint32_t c=mutation==1?0:stage?4:2;memcpy(expected_target+offsetof(nox_object_t,obj_class),&c,4); }
            else if (mutation==3) { void* p=&live_update;memcpy(expected_target+offsetof(nox_object_t,data_update),&p,sizeof(p)); }
            else if (mutation==4) { int32_t p=-129;float x=77.5f,y=88.25f;memcpy(expected_effect+offsetof(struct poison_effect,power),&p,4);memcpy(expected_target+offsetof(nox_object_t,x),&x,4);memcpy(expected_source+offsetof(nox_object_t,y),&y,4); }
        }
        assert(!memcmp(expected_effect,&effect,sizeof(effect)) && !memcmp(expected_target,&target,sizeof(target)) && !memcmp(expected_source,&source,sizeof(source)) && !memcmp(expected_weapon,&weapon,sizeof(weapon)));
        assert(!memcmp(expected_update,&update,sizeof(update)) && !memcmp(expected_live_update,&live_update,sizeof(live_update)) && !memcmp(expected_player,&player,sizeof(player)) && !memcmp(expected_live_player,&live_player,sizeof(live_player)));
        printf("{\"native_pointer_bits\":%zu,\"nonoutput_bytes_preserved\":true,\"events\":[",sizeof(void*)*8);
        for (int i=0;i<count;i++) {
            if(i)putchar(',');printf("{\"service\":\"%s\"",events[i].service);
            if(!strcmp(events[i].service,"facing"))printf(",\"direction_bits32\":%u,\"target_position_bits32\":[%u,%u],\"source_position_bits32\":[%u,%u]",events[i].direction,events[i].tx,events[i].ty,events[i].sx,events[i].sy);
            else if(!strcmp(events[i].service,"activate"))printf(",\"increment\":1,\"maximum_bits32\":%u",events[i].maximum);
            else printf(",\"player_index\":%d,\"value\":%d",events[i].index,events[i].value);
            putchar('}');
        }
        puts("]}");
    }
    return 0;
}
