# Raw Readiness whole-consumer freeze

The original `4E0960` reference contains 68 bytes and 113 independent inputs.
Its two original FPU profiles have identical complete results. The independent
native initial memory was frozen in `d67f2a171`, before these consumers and
before changing the raw C integer ABI.

Local macOS ARM64 Go 1.27.3 normal and strict highres runs each pass 577 test
events, with no skips, failures or build warnings. Strict enables `cgocheck2`
and `checkptr=2`. All four actual test mains are preserved before execution;
both drivers and all mains exit unforced with code 0. The tested source is
`14c3e8d7004edfe32b52fc007cc922a18da6bb2b` plus the two new consumer Go files.
The complete 71-file input seal and both archives are verified before this
consumer freeze.

The additive 458 events cover 113 C pointer-argument transports, 113 frozen
typed-runtime cases, 226 complete reference-guided transports and six roots.
The C stub returns signed DWORD bits `0x87654321`; all 93 successful original
return words also cross a real scalar C echo. Every original heap, stack,
caller, image, call, return and fault certificate is checked. The original
records include 611 reads, 29 input-driven stimuli and 20 memory faults.
Native typed reads retain all 1,808 independent probes and 52 protected domains.
The three same-low32/different-full-pointer callback witnesses and the actual
unchanged-C high-pointer fault after healthy-data controls remain required.

These are transports and certificate checks. The engine hook remains nil,
the original-engine test creates no engine leaves, and the public C signature
and original body remain unchanged. The profile labels select original
records; they do not establish native CPU FPU-control, program-image or stack
equivalence. No product, campaign or integer-candidate reduction is credited.

Evidence under `ART/readiness-4e0960-native-entry-port`:

- `readiness-consumer-preexecution-input-seal-v1.json`
- `readiness-raw-consumer-normal-and-strict-qualified-v1.json`
- `local-suite-archives/readiness-raw-consumer-Go1273-normal-v1-report.zip`
- `local-suite-archives/readiness-raw-consumer-Go1273-strict-v1-report.zip`

The later engine phase must retain all these evidence bytes. Its new immutable
set contains the 70 original/reference/initial/consumer files after explicitly
excluding the dedicated ABI header. Only a separately reviewed native-phase
seal may admit that header's signature change. Existing seals stay unchanged.

## Native engine qualification

After consumer freeze `93f175870`, normal and strict highres Go 1.27.3 each
pass 1,029 events: the preceding 577 plus 452 actual original-engine leaves.
All four PREEXEC mains and both drivers exit unforced 0, with no skips,
failures or build warnings. The tested source is that consumer commit plus
the five reviewed engine/ABI files. All 70 frozen inputs remain byte exact.
The new native-phase seal admits the dedicated header change explicitly;
it does not rewrite the preceding 71-file consumer seal.

The new public C entry takes `nox_object_t*` and returns `int32_t`. Both the
direct Go and public C routes execute the same native core. It preserves the
original entry-time InitData read, class mask, slots 2/3, full pointer callback
identity, reload of the matching slot and live signed strength read. The
ordinary guarded Go alias stays unchanged. The old 449-byte C transcription
is preserved verbatim under `#if 0`.

Each mode compares all 113 original cases through Go and C against both
original-profile labels. The 452 leaves include 372 successful outcomes and
80 expected in-process fault observations. Actual C traces verify 226 entries,
186 returns and 40 faulting entries. The three native callback low32 collision
witnesses remain required. The obsolete integer-C fault child is not executed
in engine mode; its preceding successful reproduction remains archived.

Evidence: `readiness-native-preexecution-input-seal-v1.json` and
`readiness-raw-native-normal-and-strict-qualified-v1.json` in the same ART
directory. Actual pinned Go-heap/GC and product integration remain pending.
The two profile labels still do not establish native CPU FPU-control,
program-image or stack equivalence. No campaign or overall 100% is credited.
