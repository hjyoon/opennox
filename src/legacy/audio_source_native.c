#include "audio_source_native.h"
#include "GAME1_3.h"

void *sub_487150_go(int index, void *format);
void *sub_4871C0_go(void *device, int index, void *format);
void *sub_487590_go(void *context, void *format);

void *sub_487150(int index, const void *format) {
    return sub_487150_go(index, (void *)format);
}
void *sub_4871C0(void *device, int index, const void *format) {
    return sub_4871C0_go(device, index, (void *)format);
}
void *sub_487590(void *context, const void *format) {
    return sub_487590_go(context, (void *)format);
}
