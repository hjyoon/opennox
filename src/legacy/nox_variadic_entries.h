// The architecture's assembler macro passes the named arguments followed
// by a pointer to the native nox_format_args cursor.
#define NOX_VARIADIC_SYMBOL_(prefix, name) prefix##name
#define NOX_VARIADIC_SYMBOL(prefix, name) NOX_VARIADIC_SYMBOL_(prefix, name)
#define NOX_VARIADIC_ENTRY(entry, handler, named) \
	nox_variadic_entry NOX_VARIADIC_SYMBOL(__USER_LABEL_PREFIX__, entry), \
		NOX_VARIADIC_SYMBOL(__USER_LABEL_PREFIX__, handler), named

NOX_VARIADIC_ENTRY(nox_fs_fprintf, nox_fs_vfprintf_go, 2)
NOX_VARIADIC_ENTRY(nox_xxx_networkLog_printf_413D30, nox_network_log_vprintf_413D30, 1)
NOX_VARIADIC_ENTRY(nox_xxx_printToAll_4D9FD0, nox_print_to_all_go_4D9FD0, 2)
NOX_VARIADIC_ENTRY(nox_ai_debug_printf_5341A0, nox_ai_debug_vprintf_5341A0, 1)
NOX_VARIADIC_ENTRY(nox_xxx_netSendLineMessage_4D9EB0, nox_net_line_message_vprintf_4D9EB0, 2)
NOX_VARIADIC_ENTRY(nox_sprintf, nox_sprintf_args_go, 2)
NOX_VARIADIC_ENTRY(nox_swprintf, nox_swprintf_args_go, 2)

#undef NOX_VARIADIC_ENTRY
#undef NOX_VARIADIC_SYMBOL
#undef NOX_VARIADIC_SYMBOL_
