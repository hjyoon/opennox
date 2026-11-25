#include "nox_wchar.h"
#include <ctype.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

char* nox_itoa(int val, char* s, int radix);

int nox_vsnwprintf(wchar2_t* buffer, size_t count, const wchar2_t* format, va_list ap) {
	int i = 0, j, out = 0;
	wchar2_t ch;

#define EMIT(c)                                                                                                        \
	do {                                                                                                               \
		if (buffer && out < count)                                                                                     \
			buffer[out++] = c;                                                                                         \
		else                                                                                                           \
			out++;                                                                                                     \
	} while (0)

	if (count == 0)
		buffer = NULL;

	while ((ch = format[i++]) != 0) {
		char flag = 0;
		int width = -1, precision = -1;

		if (ch != '%') {
			EMIT(ch);
			continue;
		}

		ch = format[i++];
		switch (ch) {
		case '+':
		case '-':
		case '0':
			flag = ch;
			ch = format[i++];
			break;
		}

		while (ch >= '0' && ch <= '9') {
			if (width < 0)
				width = 0;
			width = width * 10 + ch - '0';
			ch = format[i++];
		}

		if (ch == '.') {
			ch = format[i++];
			precision = 0;
			while (ch >= '0' && ch <= '9') {
				precision = precision * 10 + ch - '0';
				ch = format[i++];
			}
		}

		switch (ch) {
		case 'c': {
			wchar2_t c = va_arg(ap, int);
			EMIT(c);
		} break;
		case 's': {
			static const wchar2_t null[] = {'(', 'n', 'u', 'l', 'l', ')', 0};
			const wchar2_t* pwch = va_arg(ap, const wchar2_t*);
			if (pwch == NULL)
				pwch = null;
			for (j = 0; pwch[j]; j++)
				EMIT(pwch[j]);
		} break;
		case 'S': {
			const char* pcch = va_arg(ap, const char*);
			if (pcch == NULL)
				pcch = "(null)";
			for (j = 0; pcch[j]; j++)
				EMIT(pcch[j] & 0xFF);
		} break;
		case 'd':
		case 'i': {
			char tmp[32];
			int len;

			len = strlen(nox_itoa(va_arg(ap, int), tmp, 10));

			for (j = 0; j < width - (precision > 0 ? precision : len); j++) {
				if (flag == '0')
					EMIT('0');
				else
					EMIT(' ');
			}

			for (j = 0; j < precision - len; j++)
				EMIT('0');

			for (j = 0; j < len; j++)
				EMIT(tmp[j]);
		} break;
		case 'X':
		case 'x':
		case 'o':
		case 'u': {
			char tmp[32];
			int len;

			len = strlen(nox_itoa(va_arg(ap, unsigned int), tmp, ch == 'u' ? 10 : ch == 'o' ? 8 : 16));
			for (j = 0; j < width - (precision > 0 ? precision : len); j++) {
				if (flag == '0')
					EMIT('0');
				else
					EMIT(' ');
			}

			for (j = 0; j < precision - len; j++)
				EMIT('0');

			for (j = 0; j < len; j++)
				EMIT(ch == 'X' ? toupper(tmp[j]) : tmp[j]);
		} break;
		case 'f': {
			char tmp[32];
			char tmp2[32];
			double val;
			int len;
			int final_precision = precision > 0 ? precision : 5;
			snprintf(tmp, 32, "%%.%df", final_precision);
			val = va_arg(ap, double);
			snprintf(tmp2, 32, tmp, val);
			len = strlen(tmp2);

			for (j = 0; j < len; j++) {
				EMIT(tmp2[j]);
			}

		} break;
		case '%':
			EMIT('%');
			break;
		case '!':
			EMIT('!');
			break;
		default:
			printf("Unhandled format character: '%c'", ch);
			abort();
			// EMIT(ch);
			break;
		}
	}

	EMIT(0);
	return out - 1;
#undef EMIT
}

int nox_vsnprintf(char* buffer, size_t count, const char* format, va_list ap) {
	// return vsnprintf(buffer, count, format, ap);
	int i = 0, j, out = 0;
	char ch;

#define EMIT(c)                                                                                                        \
	do {                                                                                                               \
		if (buffer && out < count)                                                                                     \
			buffer[out++] = c;                                                                                         \
		else                                                                                                           \
			out++;                                                                                                     \
	} while (0)

	if (count == 0)
		buffer = NULL;

	while ((ch = format[i++]) != 0) {
		char flag = 0;
		int width = -1, precision = -1;

		if (ch != '%') {
			EMIT(ch);
			continue;
		}

		ch = format[i++];
		switch (ch) {
		case '+':
		case '-':
		case '0':
			flag = ch;
			ch = format[i++];
			break;
		}

		while (ch >= '0' && ch <= '9') {
			if (width < 0)
				width = 0;
			width = width * 10 + ch - '0';
			ch = format[i++];
		}

		if (ch == '.') {
			ch = format[i++];
			precision = 0;
			while (ch >= '0' && ch <= '9') {
				precision = precision * 10 + ch - '0';
				ch = format[i++];
			}
		}

		switch (ch) {
		case 'c': {
			char c = (char)va_arg(ap, int);
			EMIT(c);
		} break;
		case 's': {
			const char* pcch = va_arg(ap, const char*);
			if (pcch == NULL)
				pcch = "(null)";
			for (j = 0; pcch[j]; j++)
				EMIT(pcch[j]);
		} break;
		case 'S': {
			static const wchar2_t null[] = {'(', 'n', 'u', 'l', 'l', ')', 0};
			const wchar2_t* pwch = va_arg(ap, const wchar2_t*);
			if (pwch == NULL)
				pwch = null;
			for (j = 0; pwch[j]; j++)
				EMIT(pwch[j]);
		} break;
		case 'd':
		case 'i': {
			char tmp[32];
			int len;

			len = strlen(nox_itoa(va_arg(ap, int), tmp, 10));

			for (j = 0; j < width - (precision > 0 ? precision : len); j++) {
				if (flag == '0')
					EMIT('0');
				else
					EMIT(' ');
			}

			for (j = 0; j < precision - len; j++)
				EMIT('0');

			for (j = 0; j < len; j++)
				EMIT(tmp[j]);
		} break;
		case 'X':
		case 'x':
		case 'o':
		case 'u': {
			char tmp[32];
			int len;

			len = strlen(nox_itoa(va_arg(ap, unsigned int), tmp, ch == 'u' ? 10 : ch == 'o' ? 8 : 16));
			for (j = 0; j < width - (precision > 0 ? precision : len); j++) {
				if (flag == '0')
					EMIT('0');
				else
					EMIT(' ');
			}

			for (j = 0; j < precision - len; j++)
				EMIT('0');

			for (j = 0; j < len; j++)
				EMIT(ch == 'X' ? toupper(tmp[j]) : tmp[j]);
		} break;
		case '%':
			EMIT('%');
			break;
		case '!':
			EMIT('!');
			break;
		default:
			printf("Unhandled format character: '%c'", ch);
			abort();
			break;
		}
	}

	EMIT(0);
	return out - 1;
#undef EMIT
}

int nox_vsprintf(char* str, const char* format, va_list ap) { return nox_vsnprintf(str, 0x7fffffff, format, ap); }

int nox_sprintf(char* str, const char* format, ...) {
	int ret;
	va_list ap;
	va_start(ap, format);
	ret = nox_vsprintf(str, format, ap);
	va_end(ap);
	return ret;
}

int nox_swprintf(wchar2_t* str, const wchar2_t* fmt, ...) {
	int len;
	va_list ap;
	va_start(ap, fmt);
	len = nox_vsnwprintf(str, 0x3fffffff, fmt, ap);
	va_end(ap);
	return len;
}

int nox_vswprintf(wchar2_t* str, const wchar2_t* fmt, va_list ap) { return nox_vsnwprintf(str, 0x3fffffff, fmt, ap); }
