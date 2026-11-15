package legacy

/*
#include "GAME4_2.h"
#include "spell_accept_4fd400.h"
#include <stddef.h>
#include <stdlib.h>
#include <string.h>
typedef int (*resource_spell_callback_t)(int,void*,nox_object_t*,nox_object_t*,void*,int);
_Static_assert(_Generic(&nox_xxx_castSpellWinkORrestoreHealth_52BF20,resource_spell_callback_t:1,default:0),"native health callback ABI");
_Static_assert(_Generic(&sub_52BF50,resource_spell_callback_t:1,default:0),"native mana callback ABI");
_Static_assert(_Generic(&sub_52DD50,resource_spell_callback_t:1,default:0),"native LesserHeal callback ABI");
static resource_spell_callback_t resource_spell_entry_C(int mode) {
    switch(mode) {
    case 0:return nox_xxx_castSpellWinkORrestoreHealth_52BF20;
    case 1:return sub_52BF50;
    case 2:return sub_52DD50;
    default:abort();
    }
}
static int resource_spell_call_C(int mode,int spell,void* second,nox_object_t* caster,nox_object_t* fourth,void* arg,int level) {
    return resource_spell_entry_C(mode)(spell,second,caster,fourth,arg,level);
}
static void resource_spell_layout_C(uintptr_t* v) {
    v[0]=sizeof(nox_object_t);v[1]=offsetof(nox_object_t,obj_class);v[2]=offsetof(nox_object_t,data_update);
    v[3]=sizeof(nox_player_update_data_t);v[4]=offsetof(nox_player_update_data_t,mana_max);v[5]=offsetof(nox_player_update_data_t,player);
    v[6]=sizeof(nox_playerInfo);v[7]=offsetof(nox_playerInfo,info)+offsetof(nox_playerInfo2,playerClass);
    v[8]=sizeof(nox_spell_accept_arg_t);v[9]=offsetof(nox_spell_accept_arg_t,obj);
    v[10]=offsetof(nox_spell_accept_arg_t,x);v[11]=offsetof(nox_spell_accept_arg_t,y);v[12]=sizeof(void*);
}
// Independent x87 adaptation of the declared balance return, original binary32
// spills/multiply and integer conversion. The sealed Unicorn rows stay intact.
static int resource_spell_x87_C(uint64_t input, uint32_t factor_bits, int multiply, uint32_t* spill, int32_t* converted) {
#if (defined(__i386__) || defined(__x86_64__)) && defined(__GNUC__)
    unsigned char saved[28];
    uint16_t cw = 0xe3f;
    double balance;
    float factor, amount;
    int32_t result;
    memcpy(&balance, &input, sizeof(balance));
    memcpy(&factor, &factor_bits, sizeof(factor));
    __asm__ volatile(
        "fnstenv %[saved]\n\tfnclex\n\tfldcw %[cw]\n\t"
        "fldl %[balance]\n\tfstps %[amount]\n\t"
        "testl %[multiply], %[multiply]\n\tjz 1f\n\t"
        "flds %[factor]\n\tfmuls %[amount]\n\tfstps %[amount]\n\t"
        "1: flds %[amount]\n\tfistpl %[result]\n\tfldenv %[saved]\n\t"
        : [saved] "=m" (saved), [amount] "=m" (amount), [result] "=m" (result)
        : [balance] "m" (balance), [factor] "m" (factor), [cw] "m" (cw), [multiply] "r" (multiply)
        : "cc", "memory", "st");
    memcpy(spill, &amount, sizeof(amount));
    *converted = result;
    return 1;
#else
    (void)input; (void)factor_bits; (void)multiply; (void)spill; (void)converted;
    return 0;
#endif
}
*/
import "C"
import (
	"unsafe"

	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/server"
)

func resourceSpellLayoutC() (out [13]uintptr) {
	C.resource_spell_layout_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return out
}

func resourceSpellX87C(input uint64, factor uint32, multiply bool) (spill uint32, converted int32, available bool) {
	var m C.int
	if multiply {
		m = 1
	}
	available = C.resource_spell_x87_C(C.uint64_t(input), C.uint32_t(factor), m, (*C.uint32_t)(unsafe.Pointer(&spill)), (*C.int32_t)(unsafe.Pointer(&converted))) == 1
	return
}

func resourceSpellActualC(route, mode int, id int32, second, caster, fourth *server.Object, arg *server.SpellAcceptArg, level int32) int32 {
	if route == 0 {
		return int32(C.resource_spell_call_C(C.int(mode), C.int(id), unsafe.Pointer(second), asObjectC(caster), asObjectC(fourth), unsafe.Pointer(arg), C.int(level)))
	}
	if route == 1 {
		entry := unsafe.Pointer(C.resource_spell_entry_C(C.int(mode)))
		return int32(Nox_spells_call_intint6_go(entry, spell.ID(id), second, caster, fourth, arg, int(level)))
	}
	switch mode {
	case 0:
		return int32(Nox_xxx_castSpellWinkORrestoreHealth_52BF20(spell.ID(id), second, caster, fourth, arg, int(level)))
	case 1:
		return int32(Sub_52BF50(spell.ID(id), second, caster, fourth, arg, int(level)))
	case 2:
		return int32(Sub_52DD50(spell.ID(id), second, caster, fourth, arg, int(level)))
	default:
		panic("unknown resource spell")
	}
}
