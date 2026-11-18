#ifndef NOX_PORT_SERVER_SYSTEM_TRADE
#define NOX_PORT_SERVER_SYSTEM_TRADE

#include "defs.h"
#include "game4_gold_consumers.h"

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
uint32_t* sub_5109C0_trade(int* a1, int a2, uint32_t* a3);

#endif // NOX_PORT_SERVER_SYSTEM_TRADE
