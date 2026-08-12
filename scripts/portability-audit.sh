#!/bin/sh
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
repo_dir=$(CDPATH= cd -- "$script_dir/.." && pwd)
cd "$repo_dir"
export LC_ALL=C

audit() {
	label=$1
	pattern=$2
	glob=$3
	rg -n "$pattern" src -g "$glob" |
		awk -F: -v label="$label" '
			{ files[$1] = 1; matches++ }
			END {
				for (file in files) file_count++
				printf "%s\t%d\t%d\n", label, matches, file_count
			}'
}

printf 'category\tmatches\tfiles\n'
audit go_layout 'unsafe\.(Sizeof|Offsetof)' '*.go'
audit go_pointer_conversion '(uint32|int32|C\.int)\s*\(\s*uintptr|uintptr\s*\([^\n]*\)\s*\)' '*.go'
audit go_unsafe 'unsafe\.(Pointer|Add|Slice|String|Sizeof|Offsetof)' '*.go'
audit c_static_assert '_Static_assert|static_assert\s*\(' '*.{c,h}'
audit x86_isa '(?i)\b(x86|i[3-6]86|amd64|sse2?|mmx|x87|rdtsc|__asm|asm\s*\()\b' '*.{go,c,h,s,S}'
audit c_pointer_integer_cast '\((u?int(32|64)?_t|unsigned (int|long)|int|long)\)\s*[^;\n]*\*|\([^)]*\*\)\s*\(?\s*(u?int(32|64)?_t|unsigned (int|long)|int|long)' '*.{c,h}'
audit unsafe_literal_offset 'unsafe\.Add\([^,]+,\s*(0x[0-9A-Fa-f]+|[0-9]+)\s*\)' '*.go'
audit cgo_import '^import "C"|^\s*import "C"' '*.go'
