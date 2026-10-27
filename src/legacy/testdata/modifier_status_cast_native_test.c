#include <assert.h>
#include <math.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "defs.h"
#include "GAME1_1.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4_2.h"
#include "spell_accept_4fd400.h"

static struct status_effect { uint32_t before[8]; int32_t power; uint32_t after[8]; } effect;
static nox_object_t target,source,weapon,other;
static nox_player_update_data_t update,live_update;
static nox_playerInfo player,live_player;
static int which,mutation,nil_source,cast_count;
static struct { const char* service; int spell,source_nil,player_index,value; uint32_t power,x,y; } events[8];
static int count;
static void event(const char* name) { assert(count<8);events[count++].service=name; }
int32_t nox_modifier_effect_getPreHitInt(void* ptr) { assert(!ptr || ptr==&effect);return ((struct status_effect*)ptr)->power; }
double nox_xxx_gamedataGetFloat_419D40(char* name) { (void)name;return 0; }
void nox_xxx_buffApplyTo_4FF380(nox_object_t* obj,int buff,short duration,char power) {
    assert(obj==&target);(void)duration;(void)power;event("inline-buff");events[count-1].value=buff;
}
void sub_4E7540(nox_object_t* owner,nox_object_t* obj) {
    assert(owner==(nil_source?NULL:&source) && obj==&target);event("inline-attribution");
}
int nox_xxx_netInformTextMsg_4DA0F0(int index,int kind,int* value) {
    assert(kind==13);event("inform");events[count-1].player_index=index;events[count-1].value=*value;return 0;
}
static int cast(int id,void* second,nox_object_t* caster,nox_object_t* fourth,void* argp,int power) {
    nox_spell_accept_arg_t* arg=argp;
    assert(id==(which?12:74) && second==(nil_source?NULL:&source) && caster==second && fourth==caster);
    assert(arg->obj==&target);
    if (sizeof(void*)>4) assert((uintptr_t)arg>UINT32_MAX);
    event("cast");events[count-1].spell=id;events[count-1].source_nil=nil_source;events[count-1].power=(uint32_t)power;
    memcpy(&events[count-1].x,&arg->x,4);memcpy(&events[count-1].y,&arg->y,4);cast_count++;
    if (mutation) {
        effect.power=1;target.x=77.5f;target.y=88.25f;arg->obj=&other;
        if (mutation==1) target.obj_class=4;
        if (mutation==2) target.obj_class=2;
        if (mutation==3) target.data_update=&live_update;
    }
    return mutation?-1:1;
}
int nox_xxx_castStun_52C2C0(int id,void* second,nox_object_t* caster,nox_object_t* fourth,void* arg,int power) { return cast(id,second,caster,fourth,arg,power); }
int nox_xxx_castConfuse_52C1E0(int id,void* second,nox_object_t* caster,nox_object_t* fourth,void* arg,int power) { return cast(id,second,caster,fourth,arg,power); }

// PRODUCTION_FLOAT2INT
// PRODUCTION_PLAYER
// PRODUCTION_INFORM
// PRODUCTION_STUN
// PRODUCTION_CONFUSE

static void invoke(void* eff,nox_object_t* obj) {
    if (which) nox_xxx_confuseEffect_4E0670(eff,&weapon,nil_source?NULL:&source,obj,NULL);
    else nox_xxx_stunEffect_4E04D0(eff,&weapon,nil_source?NULL:&source,obj,NULL);
}
int main(int argc,char** argv) {
    if (argc==3) {
        which=atoi(argv[1]);target.obj_class=2;
        invoke(atoi(argv[2])?NULL:&effect,atoi(argv[2])?&target:NULL);
        return 0;
    }
    int cls,power;
    while (scanf("%d %d %d %d %d",&which,&cls,&power,&nil_source,&mutation)==5) {
        memset(&effect,0xa5,sizeof(effect));effect.power=power;
        memset(&target,0x96,sizeof(target));memset(&source,0x69,sizeof(source));memset(&weapon,0xc3,sizeof(weapon));memset(&other,0x3c,sizeof(other));
        memset(&update,0,sizeof(update));memset(&live_update,0,sizeof(live_update));memset(&player,0,sizeof(player));memset(&live_player,0,sizeof(live_player));
        player.playerInd=254;live_player.playerInd=130;update.player=&player;live_update.player=&live_player;
        target.obj_class=cls;target.data_update=&update;target.x=-12.75f;uint32_t minus_zero=0x80000000;memcpy(&target.y,&minus_zero,4);
        unsigned char expected_effect[sizeof(effect)],expected_target[sizeof(target)],expected_source[sizeof(source)],expected_weapon[sizeof(weapon)];
        memcpy(expected_effect,&effect,sizeof(effect));memcpy(expected_target,&target,sizeof(target));memcpy(expected_source,&source,sizeof(source));memcpy(expected_weapon,&weapon,sizeof(weapon));
        if (sizeof(void*)>4) assert((uintptr_t)&effect>UINT32_MAX && (uintptr_t)&target>UINT32_MAX && (uintptr_t)&source>UINT32_MAX && (uintptr_t)&weapon>UINT32_MAX && (uintptr_t)&update>UINT32_MAX && (uintptr_t)&player>UINT32_MAX);
        count=cast_count=0;memset(events,0,sizeof(events));invoke(&effect,&target);
        if (mutation && cast_count) {
            int32_t one=1;float x=77.5f,y=88.25f;
            memcpy(expected_effect+offsetof(struct status_effect,power),&one,4);
            memcpy(expected_target+offsetof(nox_object_t,x),&x,4);memcpy(expected_target+offsetof(nox_object_t,y),&y,4);
            if (mutation==1 || mutation==2) { uint32_t c=mutation==1?4:2;memcpy(expected_target+offsetof(nox_object_t,obj_class),&c,4); }
            if (mutation==3) { void* p=&live_update;memcpy(expected_target+offsetof(nox_object_t,data_update),&p,sizeof(p)); }
        }
        assert(!memcmp(expected_effect,&effect,sizeof(effect)) && !memcmp(expected_target,&target,sizeof(target)) && !memcmp(expected_source,&source,sizeof(source)) && !memcmp(expected_weapon,&weapon,sizeof(weapon)));
        printf("{\"native_pointer_bits\":%zu,\"nonoutput_bytes_preserved\":true,\"events\":[",sizeof(void*)*8);
        for (int i=0;i<count;i++) {
            if(i)putchar(',');printf("{\"service\":\"%s\"",events[i].service);
            if(!strcmp(events[i].service,"cast"))printf(",\"spell\":%d,\"power_bits32\":%u,\"source_nil\":%s,\"position_bits32\":[%u,%u]",events[i].spell,events[i].power,events[i].source_nil?"true":"false",events[i].x,events[i].y);
            else if(!strcmp(events[i].service,"inform"))printf(",\"player_index\":%d,\"value\":%d",events[i].player_index,events[i].value);
            else printf(",\"value\":%d",events[i].value);
            putchar('}');
        }
        puts("]}");
    }
    return 0;
}
