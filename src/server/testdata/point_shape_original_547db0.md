# Original point/shape and occupancy references

These fixtures execute the preserved PE32 GAME.EXE, SHA-256
`0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`.
Inputs, original outputs, typed C getters and native consumers were frozen
before changing the engine from `7009bc4a0c68a579eedf0c99316440b68607c43a`.
Preparation sources, all failures and their exact executables remain under
`ROOT/point-shape-native-entry-port`.

| Entry | Executed instruction bytes | Excluded alignment bytes |
| --- | ---: | ---: |
| 547DB0 point/shape predicate | 302 | 2 |
| 54EF00 occupancy search | 92 | 4 |
| 54EF60 occupancy candidate | 46 | 2 |

All 440 instruction bytes execute on valid inputs. The zero, diagonal scale
and search-gap constants total 12 bytes. The unchanged oracle manifest already
covers the latter two functions and all constants; the reference commit adds
exactly one previously missing 302-byte function record, retaining every
earlier record, its metadata and relative order.

`point_shape_original_547db0.json.gz` contains 987 contexts per FPU profile,
including three raw faults. Both original profiles produce byte-identical
results. Its gzip SHA-256 is
`c2a156668feeb44f178eecf9db25c44496cf45a0a1e22c79ec4068dcb3c87cb8`;
decoded SHA-256 is
`df1f713432dbdd2c97c5e635ae1d6e3b8ce94b7d9aa97037e4b34c4867963059`.
The frozen C/direct-Go consumer compares every result and raw-fault boundary
and guards the entire native allocation. The old Go helper fails nine result
leaves, including unspilled circle subtraction, unspilled box corner sums and
unordered first/fourth box gates. Its 978 pass / 11 fail events include two
failing parents; the failed executable and exact input values are preserved.

`point_occupancy_original_54ef00.json.gz` contains 1,224 contexts per FPU
profile: 1,095 standalone callbacks and 129 parent searches, including 14 raw
faults. UnitIterator517C10 is the sole declared world service. It invokes the
actual original callback and predicate; no geometry result or arithmetic is
substituted. Input mutations between callbacks are explicit. Ordered scratch
stores, rectangle bits, callback admission, point aliases, iterator completion,
arbitrary post-service DWORD results, raw faults and whole native memory are
compared. The original image is guarded outside the declared scratch/input
alias; caller and bounded callback-stack regions are guarded separately.

The original closure executes 1,217 geometry calls per profile. Startup
PC53/toward-zero (`0x0e3f`) and PC53/nearest (`0x023f`) differ in 72 rectangle
contexts; their outputs remain separate. The committed startup gzip SHA-256 is
`75fed3afdf90b7083cdf42403fd6a7545f97b97b31867363ae038e9b3fef7384`,
decoded SHA-256
`9d4ba9fc453e876795a34a8e1280ccf30109bf713ef121934b376aaf3992f093`.
The separately retained nearest gzip SHA-256 is
`399bbdc77bd20cbdf7d3940c646ca904a294efe91ca6c0a58e5f7b79e80d8a5c`.

All three old public C paths SIGSEGV after independent C/Go layout, typed
native getter, positive geometry and initial-memory controls pass above 4 GiB.
The parent case uses the actual native map index, UnitIterator517C10 and
ccall callback. Its first fixture omitted an explicit assignment after
`alloc.New`, so its query found no object and stopped before the C call. The
diagnostic records the real zero position/bounds. Adding only the missing
point initialization preserves every assertion and proves the actual fault.
The original failed fixture and binaries remain unchanged.

The reference getters use exact C11 `_Generic` function types for old and
native signatures, with explicit legacy narrowing and no incompatible
function-pointer cast. The occupancy consumer observes the declared iterator
and scratch services; actual production map composition is a separate gate.
These results do not establish the full world iterator, general monster
generator behavior, campaign completion or a completion percentage.
