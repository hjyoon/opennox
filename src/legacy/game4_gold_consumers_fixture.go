package legacy

/*
#include <stdint.h>
#include "game4_gold_consumers.h"
int nox_xxx_netOnPacketRecvServ_51BAD0_net_sdecode_switch(int, unsigned char*, int, nox_playerInfo*, nox_object_t*, void*);
static unsigned game4_consumer_abi(void) {
 return _Generic(&sub_50F3A0, void(*)(void*):1, default:0)
 | (_Generic(&sub_50F7F0, uint32_t*(*)(nox_object_t*,nox_object_t*):1, default:0)<<1)
 | (_Generic(&sub_50FB90, uint32_t*(*)(void*):1, default:0)<<2)
 | (_Generic(&sub_5100C0_trade, void(*)(nox_object_t*,void*,uint16_t):1, default:0)<<3)
 | (_Generic(&sub_510640_trade, float*(*)(nox_object_t*,void*,uint16_t,uint8_t):1, default:0)<<4)
 | (_Generic(&sub_510BE0_trade, uint32_t*(*)(nox_object_t*,void*,uint16_t):1, default:0)<<5)
 | (_Generic(&sub_510AE0, uint32_t*(*)(nox_object_t*,void*,uint16_t):1, default:0)<<6)
 | (_Generic(&sub_510D10, void(*)(nox_object_t*,void*,uint16_t,uint8_t):1, default:0)<<7);
}
static uintptr_t game4_consumer_return(nox_object_t* owner,nox_object_t* gold) { return (uintptr_t)sub_50F7F0(owner,gold); }
static uintptr_t game4_consumer_reprice(void* session) { return (uintptr_t)sub_50FB90(session); }
static uintptr_t game4_consumer_sell(nox_object_t* owner,void* session,uint16_t code) { return (uintptr_t)sub_510BE0_trade(owner,session,code); }
static uintptr_t game4_consumer_repair(nox_object_t* owner,void* session,uint16_t code) { return (uintptr_t)sub_510AE0(owner,session,code); }
static uintptr_t game4_consumer_buy_many(nox_object_t* owner,void* session,uint16_t typ,uint8_t count) { return (uintptr_t)sub_510640_trade(owner,session,typ,count); }
static int game4_consumer_packet(unsigned char* data,int size,nox_object_t* owner) {
 return nox_xxx_netOnPacketRecvServ_51BAD0_net_sdecode_switch(0,data,size,NULL,owner,NULL);
}
*/
import "C"

import (
	"runtime"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func game4ConsumerFixtureABI() uint32 { return uint32(C.game4_consumer_abi()) }
func game4ConsumerFixtureReturn(owner, gold *server.Object) uintptr {
	v := uintptr(C.game4_consumer_return(asObjectC(owner), asObjectC(gold)))
	runtime.KeepAlive(owner)
	runtime.KeepAlive(gold)
	return v
}
func game4ConsumerFixtureReprice(session *server.TradeSession) uintptr {
	v := uintptr(C.game4_consumer_reprice(unsafe.Pointer(session)))
	runtime.KeepAlive(session)
	return v
}
func game4ConsumerFixtureCancel(session *server.TradeSession) { C.sub_50F3A0(unsafe.Pointer(session)) }
func game4ConsumerFixtureBuy(owner *server.Object, session *server.TradeSession, code uint16) {
	C.sub_5100C0_trade(asObjectC(owner), unsafe.Pointer(session), C.uint16_t(code))
}
func game4ConsumerFixtureBuyMany(owner *server.Object, session *server.TradeSession, typ uint16, count uint8) uintptr {
	return uintptr(C.game4_consumer_buy_many(asObjectC(owner), unsafe.Pointer(session), C.uint16_t(typ), C.uint8_t(count)))
}
func game4ConsumerFixtureSell(owner *server.Object, session *server.TradeSession, code uint16) uintptr {
	return uintptr(C.game4_consumer_sell(asObjectC(owner), unsafe.Pointer(session), C.uint16_t(code)))
}
func game4ConsumerFixtureSellMany(owner *server.Object, session *server.TradeSession, typ uint16, count uint8) {
	C.sub_510D10(asObjectC(owner), unsafe.Pointer(session), C.uint16_t(typ), C.uint8_t(count))
}
func game4ConsumerFixtureRepair(owner *server.Object, session *server.TradeSession, code uint16) uintptr {
	return uintptr(C.game4_consumer_repair(asObjectC(owner), unsafe.Pointer(session), C.uint16_t(code)))
}
func game4ConsumerFixturePacket(owner *server.Object, packet []byte) int {
	var data *C.uchar
	if len(packet) != 0 {
		data = (*C.uchar)(unsafe.Pointer(&packet[0]))
	}
	v := int(C.game4_consumer_packet(data, C.int(len(packet)), asObjectC(owner)))
	runtime.KeepAlive(owner)
	runtime.KeepAlive(packet)
	return v
}
