#ifndef OPENNOX_DEFAULT_DAMAGE_PARENT_FULL_INITIAL_4E0B30_H
#define OPENNOX_DEFAULT_DAMAGE_PARENT_FULL_INITIAL_4E0B30_H
#include <stddef.h>
#include <stdint.h>
#include "defs.h"

// Explicit fixture views for pointer-free/external byte records. Neither is
// a claim that a corresponding production C typedef exists.
typedef struct {
 uint32_t kind; uint8_t name[80];
 uint32_t projectile,sound,spell,flags,cooldown,lastused;
 uint8_t charge,maxcharge,reserved[2]; uint32_t progress;
} ddi_wand_view_4e0b30;
typedef struct { uint8_t prefix[32]; uint32_t id; } ddi_sound_view_4e0b30;
_Static_assert(sizeof(ddi_wand_view_4e0b30)==116,"Wand storage");
_Static_assert(sizeof(uintptr_t)==sizeof(void*),"pointer storage");
_Static_assert(sizeof(((nox_object_t*)0)->func_damage)==sizeof(uintptr_t),"callback storage");

static uintptr_t ddi_c_offset_4e0b30(uint32_t id) {
 switch(id) {
 case 0:return offsetof(nox_object_t,typ_ind);
 case 1:return offsetof(nox_object_t,obj_class);
 case 2:return offsetof(nox_object_t,obj_subclass);
 case 3:return offsetof(nox_object_t,obj_flags);
 case 4:return offsetof(nox_object_t,x);
 case 5:return offsetof(nox_object_t,y);
 case 6:return offsetof(nox_object_t,prev_x);
 case 7:return offsetof(nox_object_t,prev_y);
 case 8:return offsetof(nox_object_t,owner);
 case 9:return offsetof(nox_object_t,obj_130);
 case 10:return offsetof(nox_object_t,field_131);
 case 11:return offsetof(nox_object_t,field_132);
 case 12:return offsetof(nox_object_t,field_133);
 case 13:return offsetof(nox_object_t,field_134);
 case 14:return offsetof(nox_object_t,health_data);
 case 15:return offsetof(nox_object_t,func_damage);
 case 16:return offsetof(nox_object_t,func_damage_sound);
 case 17:return offsetof(nox_object_t,use_data);
 case 18:return offsetof(nox_object_t,data_update);
 case 19:return offsetof(nox_player_update_data_t,state);
 case 20:return offsetof(nox_player_update_data_t,field_40_0);
 case 21:case 22:case 23:case 24:return UINTPTR_MAX;
 case 25:return offsetof(ddi_wand_view_4e0b30,flags);
 case 26:return offsetof(ddi_sound_view_4e0b30,id);
 default:return UINTPTR_MAX;
 }
}
static int ddi_pointer_field_4e0b30(uint32_t id) {
 return id==8 || id==9 || (id>=14 && id<=18);
}
// Actual compiled typed callback identities, never invoked by this initial
// fixture. Returning a constant is a transport stub, not a parent algorithm.
static int32_t ddi_damage_token_4e0b30(nox_object_t*a,nox_object_t*b,nox_object_t*c,int32_t d,int32_t e) {
 (void)a;(void)b;(void)c;(void)d;(void)e;return 0x12345678;
}
static void ddi_sound_token_4e0b30(nox_object_t*a,nox_object_t*b) {(void)a;(void)b;}
static uintptr_t ddi_callback_4e0b30(uint32_t kind) {
 return kind==2?(uintptr_t)&ddi_damage_token_4e0b30:kind==3?(uintptr_t)&ddi_sound_token_4e0b30:0;
}
#endif
