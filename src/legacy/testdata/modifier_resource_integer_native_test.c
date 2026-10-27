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
#include "unit_adjust_hp_4ee460.h"

struct modifier_resource_effect { uint32_t before[8]; float coefficient; uint32_t after[8]; };
static struct modifier_resource_effect effect;
static struct { uint32_t before; int32_t value; uint32_t after; } damage;
static nox_object_t target, owner, weapon;
static int which, mutate_service, nil_owner, nil_target;
static uint32_t available;
static struct { const char* name; uint32_t amount, x, y; int position; } events[8];
static int event_count;
static void event(const char* name, uint32_t value) {
    assert(event_count<8);
    events[event_count].name=name; events[event_count].amount=value;
    event_count++;
}
static void mutate(void) {
    if (!mutate_service) return;
    effect.coefficient=1;
    if (which==0) { target.x=77.5f; target.y=88.25f; }
}
float nox_modifier_effect_getPreHitFloat(void* ptr) { assert(ptr==&effect);return effect.coefficient; }
float nox_modifier_effect_getPreDamageFloat(void* ptr) { assert(ptr==&effect);return effect.coefficient; }
int nox_object_call_damage_native(nox_object_t* t,nox_object_t* s,nox_object_t* w,int32_t amount,int32_t typ) {
    assert(t==&target && s==(nil_owner?NULL:&owner) && w==&weapon && typ==9);
    event("damage",(uint32_t)amount);mutate();return 0;
}
int nox_xxx_netSendPointFx_522FF0(char op,float2* pos) {
    assert((uint8_t)op==129 && pos==(float2*)&target.x);
    event("point-fx",0);events[event_count-1].position=1;
    memcpy(&events[event_count-1].x,&target.x,4);memcpy(&events[event_count-1].y,&target.y,4);
    return 0;
}
void nox_xxx_aud_501960(int id,nox_object_t* obj,int kind,int code) {
    assert(id==225 && obj==&target && kind==0 && code==0);event("audio",225);
}
short nox_xxx_unitGetOldMana_4EEC80(nox_object_t* obj) {
    assert(obj==&target);event("get-mana",available&65535);mutate();return (short)available;
}
short nox_xxx_unitGetHP_4EE780(nox_object_t* obj) {
    assert(obj==&target);event("get-hp",available&65535);mutate();return (short)available;
}
uintptr_t nox_xxx_playerManaSub_4EEBF0(nox_object_t* obj,int32_t amount) {
    assert(obj==&target);event("mana-sub",(uint32_t)amount);return 0;
}
uint16_t nox_xxx_playerManaAdd_4EEB80(nox_object_t* obj,int16_t amount) {
    assert(obj==&owner);event("mana-add",(uint16_t)amount);return 0;
}
void nox_xxx_unitAdjustHP_4EE460(nox_object_t* obj,int amount) {
    assert(obj==&owner);event("hp-add",(uint32_t)amount);
}
void nox_xxx_unitDamageClear_4EE5E0(nox_object_t* obj,int amount) {
    assert(obj==&owner);event("hp-damage",(uint32_t)amount);
}

// PRODUCTION_DOUBLE2INT
// PRODUCTION_MULTIPLY
// PRODUCTION_LIGHTNING
// PRODUCTION_DRAIN_MANA
// PRODUCTION_VAMPIRISM
// PRODUCTION_SYMPATHY

int main(void) {
    uint32_t coefficient,cls,sub;
    int raw;
    while (scanf("%d %u %d %u %d %d %d %u %u",&which,&coefficient,&raw,&available,&mutate_service,&nil_owner,&nil_target,&cls,&sub)==9) {
        assert(which>=0 && which<4);
        memset(&effect,0xa5,sizeof(effect));memcpy(&effect.coefficient,&coefficient,4);
        memset(&damage,0x5a,sizeof(damage));damage.value=raw;
        memset(&target,0x96,sizeof(target));memset(&owner,0x69,sizeof(owner));memset(&weapon,0xc3,sizeof(weapon));
        target.obj_class=cls;target.obj_subclass=sub;target.func_damage=nox_object_call_damage_native;target.x=12.75f;target.y=19.25f;
        unsigned char expected_effect[sizeof(effect)],expected_damage[sizeof(damage)],expected_target[sizeof(target)],expected_owner[sizeof(owner)],expected_weapon[sizeof(weapon)];
        memcpy(expected_effect,&effect,sizeof(effect));memcpy(expected_damage,&damage,sizeof(damage));memcpy(expected_target,&target,sizeof(target));memcpy(expected_owner,&owner,sizeof(owner));memcpy(expected_weapon,&weapon,sizeof(weapon));
        uintptr_t pointers[]={(uintptr_t)&effect,(uintptr_t)&damage.value,(uintptr_t)&target,(uintptr_t)&owner,(uintptr_t)&weapon};
        if (sizeof(void*)>4) for (int i=0;i<5;i++) assert(pointers[i]>UINT32_MAX);
        event_count=0;memset(events,0,sizeof(events));
        nox_object_t* source=nil_owner?NULL:&owner;
        nox_object_t* victim=nil_target?NULL:&target;
        switch (which) {
            case 0:nox_xxx_lightngEffect_4E06F0(&effect,&weapon,source,victim,&damage.value);break;
            case 1:nox_xxx_drainMEffect_4E0740(&effect,&weapon,source,victim,&damage.value);break;
            case 2:nox_xxx_vampirismEffect_4E07C0(&effect,&weapon,source,victim,&damage.value);break;
            case 3:nox_xxx_sympathyEffect_4E08E0(&effect,&weapon,source,victim,&damage.value);break;
        }
        if (mutate_service && event_count) {
            float one=1;memcpy(expected_effect+offsetof(struct modifier_resource_effect,coefficient),&one,4);
            if (which==0) { float x=77.5f,y=88.25f;memcpy(expected_target+offsetof(nox_object_t,x),&x,4);memcpy(expected_target+offsetof(nox_object_t,y),&y,4); }
        }
        assert(!memcmp(expected_effect,&effect,sizeof(effect)) && !memcmp(expected_damage,&damage,sizeof(damage)) && !memcmp(expected_target,&target,sizeof(target)) && !memcmp(expected_owner,&owner,sizeof(owner)) && !memcmp(expected_weapon,&weapon,sizeof(weapon)));
        printf("{\"native_pointer_bits\":%zu,\"nonoutput_bytes_preserved\":true,\"events\":[",sizeof(void*)*8);
        for (int i=0;i<event_count;i++) {
            if (i) putchar(',');
            printf("{\"service\":\"%s\"",events[i].name);
            if (events[i].position) printf(",\"position_bits32\":[%u,%u]",events[i].x,events[i].y);
            else printf(",\"amount_bits32\":%u",events[i].amount);
            putchar('}');
        }
        puts("]}");
    }
    return 0;
}
