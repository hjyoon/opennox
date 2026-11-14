package legacy

/*
#include "audio_source_native.h"
void sub_431270(void);
void sub_431290(void);
uintptr_t sub_43F130(void);
extern int goAudioNativeFixtureCallback(int kind, void *owner);
#define ASN_CALLBACK(n) \
static int asn_i_##n(void *p) { return goAudioNativeFixtureCallback(n, p); } \
static void asn_v_##n(void *p) { goAudioNativeFixtureCallback(n, p); }
ASN_CALLBACK(0) ASN_CALLBACK(1) ASN_CALLBACK(2) ASN_CALLBACK(3)
ASN_CALLBACK(4) ASN_CALLBACK(5) ASN_CALLBACK(6) ASN_CALLBACK(7)
ASN_CALLBACK(8) ASN_CALLBACK(9) ASN_CALLBACK(10) ASN_CALLBACK(11)
ASN_CALLBACK(12) ASN_CALLBACK(13) ASN_CALLBACK(14) ASN_CALLBACK(15)
static void *asn_callback(int n, int notification) {
    static void *is[] = {asn_i_0,asn_i_1,asn_i_2,asn_i_3,asn_i_4,asn_i_5,asn_i_6,asn_i_7,asn_i_8,asn_i_9,asn_i_10,asn_i_11,asn_i_12,asn_i_13,asn_i_14,asn_i_15};
    static void *vs[] = {asn_v_0,asn_v_1,asn_v_2,asn_v_3,asn_v_4,asn_v_5,asn_v_6,asn_v_7,asn_v_8,asn_v_9,asn_v_10,asn_v_11,asn_v_12,asn_v_13,asn_v_14,asn_v_15};
    return notification ? vs[n] : is[n];
}
#define ASN_OFFSET(t,f) offsetof(audio_native_##t,f)
static size_t asn_layout(int type, int field) {
    static size_t voice[] = {sizeof(audio_native_voice),ASN_OFFSET(voice,link),ASN_OFFSET(voice,kind),ASN_OFFSET(voice,group),ASN_OFFSET(voice,owner_group),ASN_OFFSET(voice,master),ASN_OFFSET(voice,priority),ASN_OFFSET(voice,flags),ASN_OFFSET(voice,repeat),ASN_OFFSET(voice,context),ASN_OFFSET(voice,on_refill),ASN_OFFSET(voice,on_restart),ASN_OFFSET(voice,on_end),ASN_OFFSET(voice,on_stop),ASN_OFFSET(voice,user_data),ASN_OFFSET(voice,reserved),ASN_OFFSET(voice,methods),ASN_OFFSET(voice,mixed),ASN_OFFSET(voice,sample),ASN_OFFSET(voice,refill),ASN_OFFSET(voice,restart),ASN_OFFSET(voice,finish),ASN_OFFSET(voice,buffer),ASN_OFFSET(voice,chunk),ASN_OFFSET(voice,cursor),ASN_OFFSET(voice,available),ASN_OFFSET(voice,remaining),ASN_OFFSET(voice,field308)};
    static size_t context[] = {sizeof(audio_native_context),ASN_OFFSET(context,link),ASN_OFFSET(context,kind),ASN_OFFSET(context,users),ASN_OFFSET(context,device),ASN_OFFSET(context,index),ASN_OFFSET(context,field28),ASN_OFFSET(context,format32),ASN_OFFSET(context,format60),ASN_OFFSET(context,group),ASN_OFFSET(context,master),ASN_OFFSET(context,priority),ASN_OFFSET(context,count),ASN_OFFSET(context,limit),ASN_OFFSET(context,voices),ASN_OFFSET(context,busy),ASN_OFFSET(context,update),ASN_OFFSET(context,field220),ASN_OFFSET(context,interval),ASN_OFFSET(context,delta),ASN_OFFSET(context,peak),ASN_OFFSET(context,last),ASN_OFFSET(context,methods),ASN_OFFSET(context,field260)};
    static size_t buffer[] = {sizeof(audio_native_buffer),ASN_OFFSET(buffer,data),ASN_OFFSET(buffer,size),ASN_OFFSET(buffer,chunks),ASN_OFFSET(buffer,format),ASN_OFFSET(buffer,field24)};
    static size_t chunk[] = {sizeof(audio_native_chunk),ASN_OFFSET(chunk,link),ASN_OFFSET(chunk,data),ASN_OFFSET(chunk,size),ASN_OFFSET(chunk,buffer)};
    static size_t device[] = {sizeof(audio_native_device),ASN_OFFSET(device,link),ASN_OFFSET(device,class_),ASN_OFFSET(device,count),ASN_OFFSET(device,limit),ASN_OFFSET(device,contexts)};
    static size_t class_[] = {sizeof(audio_native_class),ASN_OFFSET(class,name),ASN_OFFSET(class,field4),ASN_OFFSET(class,flags),ASN_OFFSET(class,state),ASN_OFFSET(class,reserved),ASN_OFFSET(class,create),ASN_OFFSET(class,destroy),ASN_OFFSET(class,context_create),ASN_OFFSET(class,context_destroy),ASN_OFFSET(class,methods)};
    static size_t methods[] = {sizeof(audio_native_methods),ASN_OFFSET(methods,reserved),ASN_OFFSET(methods,create),ASN_OFFSET(methods,destroy),ASN_OFFSET(methods,play),ASN_OFFSET(methods,stop),ASN_OFFSET(methods,pause),ASN_OFFSET(methods,resume),ASN_OFFSET(methods,status),ASN_OFFSET(methods,refresh),ASN_OFFSET(methods,restart)};
    static size_t global[] = {sizeof(audio_native_global),ASN_OFFSET(global,devices),ASN_OFFSET(global,contexts),ASN_OFFSET(global,busy),ASN_OFFSET(global,group),ASN_OFFSET(global,group_initialized)};
    static size_t sample[] = {sizeof(audio_native_sample),ASN_OFFSET(sample,driver),ASN_OFFSET(sample,source),ASN_OFFSET(sample,handle),ASN_OFFSET(sample,eof),ASN_OFFSET(sample,format),ASN_OFFSET(sample,scratch0),ASN_OFFSET(sample,scratch1),ASN_OFFSET(sample,finished)};
    static size_t cache[] = {sizeof(audio_native_cache),ASN_OFFSET(cache,bank),ASN_OFFSET(cache,chunks),ASN_OFFSET(cache,samples),ASN_OFFSET(cache,lru),ASN_OFFSET(cache,chunk_size)};
    static size_t cached[] = {sizeof(audio_native_cached_sample),ASN_OFFSET(cached_sample,link),ASN_OFFSET(cached_sample,references),ASN_OFFSET(cached_sample,id),ASN_OFFSET(cached_sample,loaded),ASN_OFFSET(cached_sample,buffer),ASN_OFFSET(cached_sample,cache),ASN_OFFSET(cached_sample,format)};
    static size_t *all[] = {voice,context,buffer,chunk,device,class_,methods,global,sample,cache,cached};
    return all[type][field];
}
static int asn_scalar(int op, void *p, void *q, int a) {
    switch (op) {
    case 0:return sub_43EE00(p);
    case 1:sub_43EDB0((HSAMPLE)p);return 0;
    case 2:return sub_4BD8C0(p);
    case 3:return sub_4BD940(p);
    case 4:return sub_4BD9B0(p);
    case 5:return sub_4BDA80(p);
    case 6:return sub_4BDB40(p);
    case 7:return sub_487790(p,a);
    case 8:return sub_4873C0(p);
    case 9:return sub_487C50(p,q);
    case 10:return sub_4BD660(p);
    case 11:return sub_4BD680(p);
    case 12:return sub_4BD600(p);
    case 13:return sub_4875F0();
    case 14:return sub_487910(p,a);
    case 15:return sub_487D00(p);
    }
    return 0;
}
static void *asn_pointer(int op, void *p, void *q, int a, int b, int c) {
    switch (op) {
    case 0:return sub_4BDB20(p);
    case 1:return sub_4BDB30(p);
    case 2:return sub_4BDC00(p);
    case 3:sub_4BDB90(p,q);return 0;
    case 4:return sub_4BD7C0(p);
    case 5:return sub_4BD720(p);
    case 6:return sub_486FA0(p);
    case 7:return sub_487150(a,q);
    case 8:return sub_487810(p,a);
    case 9:return sub_4BD280(a,b);
    case 10:return sub_4BD2E0(p);
    case 11:return sub_4BD300(p,q);
    case 12:sub_4BD2D0(p);return 0;
    case 13:return sub_4BD340(p,a,b,c);
    case 14:return sub_4BD470(p,a);
    case 15:return sub_4BD420(p,a);
    case 16:return sub_4BD650(p);
    case 17:return sub_4BD710(p);
    case 18:sub_4BD3C0(p);return 0;
    case 19:sub_487C30(p);return 0;
    case 20:return sub_487D30(p,q,a);
    case 21:sub_487680(p);return 0;
    case 22:sub_4870A0();return 0;
    case 23:sub_4BDA60(p);return 0;
    case 24:sub_431270();return 0;
    case 25:sub_431290();return 0;
    }
    return 0;
}
static int asn_bank_scalar(int op, void *bank, void *data, int a) {
    switch(op) {
    case 1:return (int)sub_486A10(bank,data);
    case 2:return sub_486AA0(bank,a,data);
    case 3:return sub_486B60(bank,a);
    case 4:return sub_486DB0(bank,data,a);
    }
    return 0;
}
static void *asn_bank_pointer(int op,void *bank,int a) {
    return op == 0 ? sub_4866D0(bank,a) : sub_486E00(bank);
}
static uintptr_t asn_driver(void) { return sub_43F130(); }
*/
import "C"

import "unsafe"

var audioNativeFixtureCallbackService func(int, unsafe.Pointer) int32

//export goAudioNativeFixtureCallback
func goAudioNativeFixtureCallback(kind C.int, owner unsafe.Pointer) C.int {
	return C.int(audioNativeFixtureCallbackService(int(kind), owner))
}

func asnCallback(kind int, notification bool) unsafe.Pointer {
	return C.asn_callback(C.int(kind), C.int(bool2int(notification)))
}
func asnLayout(kind, field int) uintptr { return uintptr(C.asn_layout(C.int(kind), C.int(field))) }
func asnScalar(op int, p, q unsafe.Pointer, arg int32) int32 {
	return int32(C.asn_scalar(C.int(op), p, q, C.int(arg)))
}
func asnPointer(op int, p, q unsafe.Pointer, a, b, c int32) unsafe.Pointer {
	return C.asn_pointer(C.int(op), p, q, C.int(a), C.int(b), C.int(c))
}

func AudioNativeBankScalar(op int, bank, data unsafe.Pointer, arg int32) int32 {
	return int32(C.asn_bank_scalar(C.int(op), bank, data, C.int(arg)))
}
func AudioNativeBankPointer(op int, bank unsafe.Pointer, arg int32) unsafe.Pointer {
	return C.asn_bank_pointer(C.int(op), bank, C.int(arg))
}
func asnDriver() uintptr { return uintptr(C.asn_driver()) }
