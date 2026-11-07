/* The six definitions are substituted byte for byte from preserved GAME1.c.
 * This program executes those bodies, with independently selected live cells
 * and stable clock inputs. It contains no substitute return-value model. */
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
_Static_assert(sizeof(int)==4 && sizeof(unsigned int)==4, "exact DWORD ABI");
_Alignas(4) static unsigned char test_blob[12000];
static uint32_t input_frame,input_fps;
uint32_t dword_5d4594_10984;
static uint32_t gameFrame(void) { return input_frame; }
static uint32_t gameFPS(void) { return input_fps; }
static uint32_t* getMemU32Ptr(uint32_t base,size_t offset) {
    if(base!=0x5d4594||offset>sizeof(test_blob)-4) exit(71);
    return (uint32_t*)(void*)(test_blob+offset);
}
// ORIGINAL_sub_40AA00
// ORIGINAL_sub_40AA20
// ORIGINAL_sub_40AA30
// ORIGINAL_sub_40AA40
// ORIGINAL_sub_40AA60
// ORIGINAL_sub_40E090
static uint32_t read32(void) {
    unsigned char b[4];if(fread(b,1,4,stdin)!=4) exit(72);
    return (uint32_t)b[0]|(uint32_t)b[1]<<8|(uint32_t)b[2]<<16|(uint32_t)b[3]<<24;
}
static void write32(uint32_t v) {
    unsigned char b[4]={(unsigned char)v,(unsigned char)(v>>8),(unsigned char)(v>>16),(unsigned char)(v>>24)};
    if(fwrite(b,1,4,stdout)!=4) exit(73);
}
static int signed32(uint32_t v) { int result;memcpy(&result,&v,4);return result; }
int main(void) {
    uint32_t endian=1;if(*(unsigned char*)&endian!=1) return 74;
    if(read32()!=0x36414153) return 75;
    uint32_t count=read32();if(count!=112) return 76;
    for(uint32_t i=0;i<count;i++) {
        uint32_t entry=read32(),arg=read32(),result=0;
        input_fps=read32();input_frame=read32();dword_5d4594_10984=read32();
        if(fread(test_blob+3468,1,80,stdin)!=80||fread(test_blob+10976,1,20,stdin)!=20) return 77;
        switch(entry) {
        case 0x40aa00:result=(uint32_t)sub_40AA00();break;
        case 0x40aa20:result=(uint32_t)sub_40AA20();break;
        case 0x40aa30:result=(uint32_t)sub_40AA30(signed32(arg));break;
        case 0x40aa40:result=(uint32_t)sub_40AA40();break;
        case 0x40aa60:result=(uint32_t)sub_40AA60(signed32(arg));break;
        case 0x40e090:sub_40E090();break;
        default:return 78;
        }
        write32(result);
        if(fwrite(test_blob+3468,1,80,stdout)!=80) return 79;
        write32(dword_5d4594_10984);
        if(fwrite(test_blob+10976,1,20,stdout)!=20) return 80;
    }
    if(getchar()!=EOF||ferror(stdin)||fflush(stdout)) return 81;
    return 0;
}
