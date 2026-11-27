#ifndef NOX_PORT_SERVER_SYSTEM_TRADE
#define NOX_PORT_SERVER_SYSTEM_TRADE

#include "defs.h"
#include "game4_gold_consumers.h"
#include <stddef.h>

typedef struct {
	void* sessions;
	void* items;
	uint32_t head_word;
} nox_legacy_trade_pools_t;
_Static_assert(offsetof(nox_legacy_trade_pools_t, head_word) == 2 * sizeof(void*), "legacy trade head word");
nox_legacy_trade_pools_t* nox_legacy_trade_pool_storage(void);

int sub_50E7A0(void* session, nox_object_t* item);
int sub_5104F0(nox_object_t* recipient, short amount);
int sub_510540(nox_object_t* item);
int sub_5105D0(nox_object_t* item);
void* nox_xxx_createPlayerShopSession_50E8F0(nox_object_t* player, nox_object_t* merchant);
int sub_50F0F0(nox_object_t* recipient, void* session);
int sub_50F1A0(nox_object_t* recipient, void* session);
uintptr_t nox_xxx_servSendShopItems_50F280(nox_object_t* recipient, void* session);
void* nox_xxx_servShopStart_50EF10_trade(nox_object_t* starter, nox_object_t* target);
int sub_50FD60(void* head, nox_object_t* item);
int nox_xxx_tradeP2PAddOffer2_50F820_trade(void* session, nox_object_t* actor, nox_object_t* item);
void sub_510320(nox_object_t* item, void* session);
uintptr_t sub_5109C0_trade(nox_object_t* player, void* session, uint32_t code);

#endif // NOX_PORT_SERVER_SYSTEM_TRADE
