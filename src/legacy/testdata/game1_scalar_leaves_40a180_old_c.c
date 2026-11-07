/* Exact old function definitions replace the seven markers before compilation.
 * This standalone program links no new Go implementation and computes no
 * substitute model for a function result. Input/output are little-endian. */
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

_Static_assert(sizeof(int) == 4 && sizeof(short) == 2 && sizeof(unsigned char) == 1, "old public scalar ABI");
_Alignas(4) static unsigned char blob587000[8192];
_Alignas(4) static unsigned char blob5d4594[8192];
static void* getMemAt(uint32_t base, size_t offset) {
    if (offset >= 8192) exit(81);
    if (base == 0x587000) return blob587000 + offset;
    if (base == 0x5d4594) return blob5d4594 + offset;
    exit(82);
}
static uint32_t* getMemU32Ptr(uint32_t base, size_t offset) { return getMemAt(base, offset); }
static unsigned char getMemByte(uint32_t base, size_t offset) { return *(unsigned char*)getMemAt(base, offset); }

// ORIGINAL_sub_409A70
// ORIGINAL_sub_40A180
// ORIGINAL_sub_40A220
// ORIGINAL_sub_40A3C0
// ORIGINAL_sub_40A6A0
// ORIGINAL_sub_40A6B0
// ORIGINAL_sub_40A710

static uint32_t read32(void) {
    unsigned char b[4]; if (fread(b, 1, 4, stdin) != 4) exit(83);
    return (uint32_t)b[0] | (uint32_t)b[1]<<8 | (uint32_t)b[2]<<16 | (uint32_t)b[3]<<24;
}
static void write32(uint32_t value) {
    unsigned char b[4] = {(unsigned char)value, (unsigned char)(value>>8), (unsigned char)(value>>16), (unsigned char)(value>>24)};
    if (fwrite(b, 1, 4, stdout) != 4) exit(84);
}
static int signed32(uint32_t value) { int word; memcpy(&word, &value, 4); return word; }
static short signed16(uint16_t value) { short word; memcpy(&word, &value, 2); return word; }
int main(void) {
    uint32_t endian = 1; if (*(unsigned char*)&endian != 1) return 85;
    if (read32() != 0x36313041) return 86;
    uint32_t configurations = read32(); if (configurations != 9) return 87;
    for (uint32_t config = 0; config < configurations; config++) {
        for (unsigned i = 0; i < 6; i++) *getMemU32Ptr(0x587000, 4704+4*i) = read32();
        if (fread(getMemAt(0x5d4594, 3500), 1, 6, stdin) != 6) return 88;
        for (uint32_t input = 0; input <= 65535; input++) {
            unsigned char result = sub_40A180(signed16((uint16_t)input));
            if (fwrite(&result, 1, 1, stdout) != 1) return 89;
        }
    }
    uint32_t cases = read32(); if (cases > 1000) return 90;
    for (uint32_t index = 0; index < cases; index++) {
        uint32_t kind = read32(), argument = read32(), result = 0;
        if (fread(getMemAt(0x587000, 4652), 1, 76, stdin) != 76) return 91;
        if (fread(getMemAt(0x5d4594, 3500), 1, 92, stdin) != 92) return 92;
        switch (kind) {
        case 0: result = (uint32_t)sub_40A220(); break;
        case 1: result = sub_40A3C0(argument); break;
        case 2: result = (uint32_t)sub_40A6A0(signed32(argument)); break;
        case 3: result = (uint32_t)sub_40A6B0(); break;
        case 4: result = (uint32_t)sub_40A710(signed32(argument)); break;
        default: return 93;
        }
        write32(result);
        if (fwrite(getMemAt(0x587000, 4652), 1, 76, stdout) != 76) return 94;
        if (fwrite(getMemAt(0x5d4594, 3500), 1, 92, stdout) != 92) return 95;
    }
    if (getchar() != EOF || ferror(stdin) || fflush(stdout)) return 96;
    return 0;
}
