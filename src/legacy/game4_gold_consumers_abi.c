#include <stdint.h>
#include "GAME4_1.h"
#include "server__system__trade.h"

uintptr_t nox_game4_goldReturn_50F7F0(nox_object_t*, nox_object_t*);
uintptr_t nox_game4_tradeReprice_50FB90(void*);
uintptr_t nox_game4_shopBuyByType_510640(nox_object_t*, void*, uint16_t, uint8_t);
uintptr_t nox_game4_shopSell_510BE0(nox_object_t*, void*, uint16_t);
uintptr_t nox_game4_shopRepair_510AE0(nox_object_t*, void*, uint16_t);

uint32_t* sub_50F7F0(nox_object_t* owner, nox_object_t* gold) {
	return (uint32_t*)nox_game4_goldReturn_50F7F0(owner, gold);
}
uint32_t* sub_50FB90(void* session) {
	return (uint32_t*)nox_game4_tradeReprice_50FB90(session);
}
float* sub_510640_trade(nox_object_t* owner, void* session, uint16_t type, uint8_t count) {
	return (float*)nox_game4_shopBuyByType_510640(owner, session, type, count);
}
uint32_t* sub_510BE0_trade(nox_object_t* owner, void* session, uint16_t code) {
	return (uint32_t*)nox_game4_shopSell_510BE0(owner, session, code);
}
uint32_t* sub_510AE0(nox_object_t* owner, void* session, uint16_t code) {
	return (uint32_t*)nox_game4_shopRepair_510AE0(owner, session, code);
}
