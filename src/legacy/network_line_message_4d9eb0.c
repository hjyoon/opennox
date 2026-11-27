//go:build !386 && !amd64 && !arm && !arm64

#include "GAME3_2.h"
#include "noxstring_format_abi.h"

intptr_t nox_net_line_message_vprintf_4D9EB0(nox_object_t* obj, nox_format_word* format, nox_format_args* args);

intptr_t nox_xxx_netSendLineMessage_4D9EB0(nox_object_t* obj, wchar2_t* format, ...) {
	nox_format_args args;
	va_start(args.list, format);
	intptr_t result = nox_net_line_message_vprintf_4D9EB0(obj, format, &args);
	va_end(args.list);
	return result;
}
