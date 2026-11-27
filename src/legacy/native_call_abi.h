#ifndef OPENNOX_NATIVE_CALL_ABI_H
#define OPENNOX_NATIVE_CALL_ABI_H

#if defined(NOX_NATIVE_CALL_FALLBACK)
#define NOX_NATIVE_CALL(result, name, target, params, body) result name params { body }
#elif defined(__i386__) || defined(__x86_64__) || defined(__arm__) || defined(__aarch64__)
#define NOX_NATIVE_CALL_STRING_(value) #value
#define NOX_NATIVE_CALL_STRING(value) NOX_NATIVE_CALL_STRING_(value)
#define NOX_NATIVE_CALL(result, name, target, params, body) \
    result name params __asm__(NOX_NATIVE_CALL_STRING(__USER_LABEL_PREFIX__) #target);
#else
#define NOX_NATIVE_CALL(result, name, target, params, body) result name params;
#endif

#endif
