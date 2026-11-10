#ifndef OPENNOX_GAME4_GOLD_CONSUMERS_H
#define OPENNOX_GAME4_GOLD_CONSUMERS_H
#include <stdint.h>
#include "defs.h"

/* TradeSession is opaque here; its native pointer fields are Go-owned. */
void sub_50F3A0(void* session);
uint32_t* sub_50F7F0(nox_object_t* owner, nox_object_t* gold);
uint32_t* sub_50FB90(void* session);
void sub_5100C0_trade(nox_object_t* owner, void* session, uint16_t net_code);
float* sub_510640_trade(nox_object_t* owner, void* session, uint16_t type, uint8_t count);
uint32_t* sub_510BE0_trade(nox_object_t* owner, void* session, uint16_t net_code);
uint32_t* sub_510AE0(nox_object_t* owner, void* session, uint16_t net_code);
void sub_510D10(nox_object_t* owner, void* session, uint16_t type, uint8_t count);
int nox_server_netTradeNative_GAME4_51BAD0(unsigned char* data, int size, nox_object_t* owner);
#endif
