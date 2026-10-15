# Go 도구체인 정책

## R233: 맵 섹션 길이의 aligned DWORD 계약

기준은 clean/pushed `d5d6bc309b4cdea0083a6ee1687f800e987c8ef7`이며 사용자 goenv Go 1.26.9를 유지한다. `maps.Reader.nextSection`의 길이 읽기 한 곳을 기존 `readAlignedU32`로 연결했다. 원본 클라이언트의 `004AC4EF`는 길이 4를 넘겨 `00426C20`을 호출한다. 암호화 블록은 8바이트지만 길이는 앞 DWORD이고 뒤 DWORD는 패딩이다. 게임 로더·체크섬·writer·기존 EOF/오류 계약은 수정하지 않는다.

새 회귀는 writer를 사용하지 않고 두 magic·다섯 unused DWORD·32개 길이의 연속 세 섹션을 직접 암호화한다. 같은 동결 입력의 수정 전 64pass/257fail·actual1, 수정 후 321pass·actual0이며 기존 R232와 함께 공식 cgocheck2/race/checkptr2 count3는 3,051pass/0skip이다. 원본 157개 맵/212,179개 객체의 수·순서·타입·opaque payload SHA는 R232와 동일하다. 모든 XFER 의미 디코딩이나 인게임 로드 수정으로 확대하지 않는다.

읽기 전용 재출력 probe는 원본 157개 중 152개에서 checksum DWORD만 다르고, TriLevel/Con09b/War02b/War09b/Wiz09b 다섯 파일에서는 마지막 8바이트도 사라짐을 확인했다. 기존 라이브러리는 수정 전후 3,098pass/158fail/0skip·actual1이며 모든 ID/action/count/package 결과가 같다. 이 158개 TestMapWrite 실패를 이번 GREEN으로 세거나 원래 CRC를 복사해 감추지 않는다. EOF marker와 원본 writer의 호출 경계는 후속 조사 대상으로 남긴다.

원본 CRC 본체 `00426BD0` 75바이트와 원래 1,024바이트 table을 변경 없이 실행한 216조건/360호출은 호출마다 마지막 반전을 하는 계약을 확인한다. private overlay의 실제 ARM64 Go 경로도 공식 strict/race/checkptr2에서 같은 360개 결과 SHA를 3회 재현했다. 독립 framing parser는 157개 모두 EOF용 zero name을 갖고 다섯 정렬 끝의 8바이트가 writer 출력에서 누락됨을 확인한다. 이 추가 경계 증거는 audit-extra-evidence-v1.log로 구분한다. 동일 바이트도 호출 분할에 따라 CRC가 달라지므로 표준 streaming CRC로 교체하지 않는다. 첫 sandbox 진단의 SIGILL과 같은 입력의 별도 실행 actual0을 함께 보존한다. 이는 전체 Windows writer·OS I/O나 원본 편집기 실행 검증이 아니다.

처음 루트 전체 검사에 NOX_DATA를 추가해 이전 기준보다 자산 검사를 넓혔다. 이 실행과 이전 reader 본체 overlay를 사용하는 동일 환경의 baseline은 둘 다 74,351pass/40skip/1,141fail·actual1이며 모든 test ID/action/count/package가 같다. 한국어 대사 1,137개의 비어 있는 기대 해시·standalone chap1.mp3 unsupported와 parent 3개이며 실제 PCM 해시/오류도 같다. 첫 private 비교기는 unrelated timestamped unsupported 로그까지 비교해 actual1이었고, 별도 v2는 실패 음성의 정확한 payload만 대조했다. 기존 테스트·기대 해시는 변경하지 않는다.

canonical 5종 게임 검사는 이전 R232와 동일하게 NOX_DATA를 지정하지 않으며 자산 확장 실패와 구분한다. 완료 여부와 전후 결과는 audit-body-v2.log를 따른다. 새 clean 네 ARM64 제품·공식 일반/HD server·기존 일반/HD Host Load 8화면 후속 판정은 audit-release-v2.log를 따른다. 자료는 `/private/tmp/opennox-map-crc-r233.aDUo52`에 보존한다. 이 README prefix·한 production 함수·새 독립 회귀 파일만 커밋하고 직후 origin port/go1.26-multiarch로 push한다. 개인 Save/설정·원본·기존 tests/YAML/PNG/oracle·R134 HOLD는 유지한다. 아직 실행하지 않은 검사를 완료로 세지 않으며 TestMapWrite/음성 자산 기대값/War01A/fixed Quest/Player SetColor·대화 API·외부 접속 원인·다른 GUI/AI/API·물리 GPU/가청 출력과 전체 포팅 목표는 미완료다.

## R232: 맵 객체의 정렬·DWORD 길이와 native int 경계

기준은 clean/pushed `862b6fcc540aef3686d267339dfcf34a03f529bd`이며 사용자 goenv Go1.26.9를 유지한다. `binenc.Reader.Align`이 이미 정렬된 위치에서도8바이트를 더 건너뛰는 계산을 교정했다. `maps.Objects.ReadObjects`는 원래 crypt block8바이트의 앞 DWORD만 길이로 읽고 나머지4바이트는 패딩으로 소비하며, native int 변환 전에 남은 길이와 비교한다. 두 production 함수 외 기존 decoder/TOC·오류/입력·게임 로더·AI·네트워크·화면 계약은 바꾸지 않는다.

동결한 새 회귀 두 파일은 정렬·반복 정렬·EOF·패딩 부족, 여러 payload 길이의 연속 객체·비어 있거나0xffffffff인 unused DWORD, 잘린 header·0x80000000/0xffffffff 길이·기존 TOC/부분 디코딩 오류를 검사한다. 수정 전18pass/678fail·actual1, 같은 입력의 수정 후696pass·actual0이고 공식 cgocheck2+race+checkptr2/count3는2,088pass/0skip·2packages다. Linux/386 테스트 바이너리 빌드는 actual0이며32비트 실행 검증으로 세지 않는다. 최소 언어 버전·전역 goenv/runtime/GC·캐시는 변경하거나 삭제하지 않는다.

별도 읽기 전용 probe는 실제 ReadObjects API의 XFER decoder만 opaque payload 관찰기로 지정해 타입·바이트 framing을 독립적인 aligned-DWORD parser와 대조한다. 같은 원본157개 맵/212,179개 객체에서 수정 전151개 맵이 unexpected EOF였고 수정 후157개 모두 객체 수·순서·타입·각 payload SHA가 일치한다. 모든 XFER 클래스의 의미 디코딩이나 인게임 load 치트 수정으로 확대하지 않는다. 현재 게임 로더는 별도의 CryptFile.ReadAlignedU32 경로이며 수정하지 않았다.

기존 라이브러리 전체 비교는 수정 전후 모두2,402pass/158fail/0skip·actual1이고 test ID/action/count/package 결과가 같다. 실패158개는157개 원본 맵의 TestMapWrite 자식과 parent다. 적어도 Estate의 원문 diff는 같은59,020bytes 외 checksum DWORD만 다르다. 이 기존 재출력 실패·기대값은 보존하며 이번 객체 읽기 수정의 GREEN으로 세지 않는다. checksum을 임의 보존하거나 기존 oracle을 수정하지 않고 별도 원본 writer 계약 조사 대상으로 남긴다.

이전 R231에서는 원본 VortexSource를 격리 로컬 게임에 배치한 세 headless 실행(등록 CGo 진단·일반 dispatch·HD dispatch)에서 draw=nil source의60개 파티클·삭제 후360tick 자연 소멸을 확인했다. 세 live 화면끼리, 세 expired 화면끼리 SHA가 같고 두 stage는 서로 다르다. 관련 기존 엄격 검사6,288pass/0skip로 R228의13개 등록 진입점 월드 경로를 모두 관찰했다. 이는 원본 캠페인 자연 발생·모든 효과·물리 출력·전체 포팅 완료 인증이 아니다. R231 자료는 `/private/tmp/opennox-vortex-world-r231.Qa7lO7`에 동결했다.

이번 자료는 `/private/tmp/opennox-map-records-r232.e1mvtT`다. 전체 게임5종 회귀·기존 결과 보존·소스/개인/원본 감사는 audit-body-v1.log, 새 clean 네 ARM64 제품·공식 일반/HD server·동일 Host Load 화면 후속 검증은 audit-release-v1.log를 따른다. 이 prefix만 추가하고 기존 README suffix·기존 tests/YAML/PNG/oracle·R134 HOLD를 유지해 커밋 직후 origin port/go1.26-multiarch로 push한다. 아직 실행하지 않은 검사는 완료로 세지 않으며 TestMapWrite·War01A/fixed Quest·Player SetColor/문자열 대화 API·외부 목록/접속 원인·다른 GUI/AI/API·physical GPU/가청 음향과 전체 포팅 목표는 미완료다.

## R230: 남은 등록 효과의 실제 월드 경로와 Imp 선공 회귀

기준은 clean/pushed `cf65497c8375bc3934b870c241fe5e2ff51f4e11`이고 사용자 goenv Go1.26.9를 유지한다. R228의 native CGo 등록 진입점13개 중 아직 월드에서 관찰하지 못한 경로를 추가 조사했다. 별도 진단 overlay는 기존 효과 계산·callback 등록·RNG·시계·HP·패킷을 바꾸지 않고5개 dispatch 호출만 실제 등록 CGo 진입점으로 우회한다.

기존 Poison/Troll 사망 독구름/Cleansing Flame YAML과 모든 기대값을 byte-identical하게 보존한9회 headless 실행(진단3·clean 일반3·clean HD3)은 모두 actual0이다. Magic의 원본 void ABI와 실제4/5개 spark 생성, SmallToxicCloud의 짝수 frame0/홀수 frame1개 생성, 네 실제 Cleansing Flame 형식의 이동과 viewport 이탈 제거를 관찰했다. 각 phase의 기존 완료 표식14개를 유지했다. 이들은 배치·대기 AI·script cast 등을 명시한 기존 fixture이며 입력만으로 하는 전투·물리 GPU/가청 출력 인증은 아니다.

원본 thing.bin/monster.bin은 ImpShot이 SparkleTrailUpdateDraw를 사용하고 stock Imp가 이를 발사함을 확인했다. 기존 선공 fixture의 지원 유형에 Imp를 추가하고 Urchin과 동일하게 사전 관찰한 소유 투사체를 인정하며 missile 관찰을 요구한다. 기존16조건·assertion·600tick deadline·HP2000·RNG·게임 AI/피해 계산은 변경하지 않는다. 새 독립 YAML은 두 대각선×clear/off-ray-box의4조건이다. 동결 새 회귀의 실제 RED를 확인한 뒤, 진단 게임은 네 조건 모두 피격 이전 자체 적 탐색/FIGHT/발사·서버 HP 감소와 client 피해 관찰을 통과했다. 실제 ImpShot CGo는 native-width viewport/drawable을 유지하며 매 호출5개 spark를 생성했다.

처음 private 수집기는 정상 공격자 본체도 허용하는 기존 계약을 잘못 좁혀 observed-projectile=true만 요구했다. 게임 actual0/네 PASS와 수집기 actual1을 함께 보존하고 별도 v2 감사는 기존 공격자 또는 사전 관찰 투사체라는 계약으로 원문을 확인했다. 기존 게임 검사나 기대값을 완화한 GREEN으로 세지 않는다. client의 음수 damage 알림은 관찰값이며 매 Imp 타격의 정확한 독립 delta로 주장하지 않는다.

기존 native 효과33IDs×3=99pass와 기존/새 AI55IDs×3=165pass는 공식 cgocheck2+race+checkptr2에서0skip/actual0이다. 전체5종 suite를 이번에 재실행했다고 주장하지 않는다. R228에서 실제 월드로 확인한8개에 이번4개를 더해12/13이며 VortexSource는 여전히 direct ABI/payload 단위검사만 확인했다. 별도 읽기 전용157개 원본 맵 조사에는 VortexSource 배치가 없었다. 처음 두 조사 도구의151개 semantic 오류는 보존하고, v3는 이미 정렬된 record에서 padding을 추가하지 않아157개를 모두 읽었다. 게임/라이브러리 맵 decoder를 수정한 것은 아니다.

자료는 `/private/tmp/opennox-effect-paths-r230.ReyUpg`다. 기존9회 감사 SHA `48910e881e8889f0a092a7b43f2d2c1a58df7b3902962b817b3ab4e975a58d9f`, 독립 Imp 감사 SHA `573bee9bdf3e95bb9d24664be86f89f9996e401229bfefb5f8ef4a8624f631bd`를 보존한다. 이번 변경은 검증 fixture3곳·새 테스트·새 YAML과 이 README prefix뿐이다. 개인 Save/설정·기존 YAML/PNG/oracle·R134 HOLD는 유지한다. 커밋 직후 push하고 새 clean 일반/HD client·server와 Host Load 후속 검증은 별도 receipt로 확인한다. R229 공개 접속2ready/3timeout/1원래75초 다운로드 deadline은 과거 결과로 유지하며 이번 재접속으로 세지 않는다. 다른 GUI/AI/API·War01A/fixed Quest·Player SetColor/R134·외부 서버 원인·물리 출력·전체 포팅 목표는 미완료다.

## R229: 공개 목록 서비스 실패와 여섯 주소의 실제 Join 분리

기준은 clean/pushed `57c472e798ffad873c3132165b852a45e42dc3fd`이며 사용자 goenv Go1.26.9를 사용한다. R228의 네 clean ARM64 제품·일반/HD 효과·실제 시계 로컬 접속/이동/종료·Host Load8화면 후속 감사는 actual0/SHA `e2367d5caa1f5456b3fdeb3c2b2bbc83c7d8937aeb1b81b7ff1084df3d2b41e4`다. 이번 게임 소스·public tests·기존 YAML/PNG/oracle 변경은0이다.

2026-10-15 20:36/20:37(KST)의 두 production discovery는 목록 API 연결 거부와 XWIS 대체 조회 시간 초과로 각각0개를 반환했다. 이를 성공한 현재 전체 서버 목록으로 세지 않는다. 앞서 사용자에게 허용받은 17:50(KST) snapshot의 여섯 주소를 새 private `game_ip.txt`로 조회했다. 현재 UDP 정보에 응답한 주소는 EU/러시아 두 곳이며 각각 exact protocol `0x0001039a`였다. 나머지 네 곳의 미관찰 protocol은 nil로 유지하고 기존 resolution fallback을 적용했다. 여섯 실제 GUI row 선택 후 C 선택 레코드까지 protocol이 보존됨을 확인했다.

| 이전 목록의 이름 / 주소 | 이번 일반 UI Join 결과 |
| --- | --- |
| Waystation / 121.200.33.119:18590 | 게임의 접속 시간 초과 오류 |
| USA Midwest #1 / 173.208.129.98:18590 | 승인 후 UDP 맵 다운로드, 원래75초 안에 입장 미완료 |
| NoXCommunity EU / 212.192.2.113:18590 | 맵 진입·HP100/mana125·관전자 idle·UI Quit/Yes·정상 disconnect |
| DELVE lobby / 217.216.92.175:18590 | 게임의 접속 시간 초과 오류 |
| Kor] Newbies / 221.153.137.109:18590 | 게임의 접속 시간 초과 오류 |
| 1:NoXWorld.ru / 45.144.64.229:18590 | 현재 UDP 이름 NoXWorld-Main, 맵 진입·HP100/mana125·관전자 idle·UI Quit/Yes·정상 disconnect |

모든 여섯 process는 actual0·signal없음이며 이는 접속 여섯 곳 성공이라는 뜻이 아니다. Join/캐릭터 생성/appearance/오류 확인·종료는 정상 UI 입력이고 실제 시계/headless·무음으로 실행했다. 공개 서버에서 spawn/보행/전투/chat/치트/비밀번호 제출·호스트 설정 변경은 없다. 관전자 eye/반투명 표시를 별도 Host Load 반투명 회귀와 혼동하지 않는다. 로컬 호스팅/이동·일반/HD Host Load 증거는 같은 게임 코드의 R228 검증이며 이번에 이를 새로 실행했다고 세지 않는다.

미국 서버는 같은 CRC `0x882e6915`의230,933byte 맵 중 연속72parts/36,864bytes(15.9631%)를20:41:09→20:42:03에 수신했다. 남은 partial 파일은 실제 수신 packet payload와 byte-identical(SHA `eb870258ec8b447df01a33bf4db2e31b77e4181aa1cbfe85fd60632f6b02a765`)이다. 원래75초 gate를 연장하거나 ready로 바꾸지 않으며 이 기록만으로 엔진 결함을 확정하지 않는다. R225의 별도87.510초 입장과 R224의3ready/2timeout/1원격 종료는 과거 결과로 보존하고 이번2ready/3접속 timeout/1다운로드 deadline과 구분한다.

기존 protocol56IDs는 공식 Go1.26.9 cgocheck2+race+checkptr2에서 count3·168pass/0skip·2packages·actual0이다. 전체5종 suite를 이번에 재실행했다고 주장하지 않는다. 검사 전후 source6,849files/protected304files/개인1,562files의 checkpoint가 동일하며 R134 HOLD·R225/R228 동결 증거·원본 자산을 유지한다. 첫 static 조회 helper가 const에 대입해 compile actual1이었던 입력/receipt는 그대로 남기고, 별도 v2는 private cwd의 정규 game_ip.txt 읽기만 사용한다. 게임 기대값/timeout 완화나 전역 goenv/runtime/GC·캐시 변경/삭제는 없다.

자료는 `/private/tmp/opennox-current-public-r229.wCUHEA`다. runtime 감사 actual0/SHA `d4f9ffac1c26a713171495b4959554fc6a59f72d266b5dc0a37d511f0e5ab1c7`, UDP 감사 actual0/SHA `9b170c02a5fea49c240f5306dd8d590168c68e158e14d12fde7d5dfb6bcadddf`를 보존한다. 이 README prefix만 추가하고 기존 suffix를 유지해 커밋 직후 origin port/go1.26-multiarch에 push한다. 새 docs-only revision의 네 clean 제품·metadata·공식 server 후속 검증은 별도 receipt로 확인하며 아직 실행하지 않은 검사를 완료로 세지 않는다. 공개 목록/timeout의 외부 원인·War01A/fixed Quest·Player SetColor/R134 HOLD·다른 GUI/AI/API·physical GPU/가청 출력·전체 포팅 목표는 미완료다.

## R202: 중앙 HUD 알림의 native 반복 경계

사용자 goenv Go1.26.9를 계속 사용한다. 실제 맵/Save 알림 초기화와 HUD 중앙 문구 출력에 남은 signed32 주소 비교 두 개만 원본의3개 알림/4개 그림자 반복 횟수로 바꿨다. 원본 두 전체 함수·unhooked UTF-16 복사 실행, 동결 compiled RED와 수정 후13,365 native 조건·실제 영어/한글 font/raster/headless402frames·관련 공식 strict/race/checkptr2 3회33pass가 통과했다. 일반 주소의 수정 전 실제 화면도 통과하므로 그것을 behavioral RED로 세지 않는다. 전체 일반/HD/CGo 각각73,731pass/47skip·server69,165pass/7skip과 모든 기존 결과가 동일하다. 일반/HD headless Quest8화면·stage1→2/이동/종료와 원본/개인 데이터·기존 기대값·R134 HOLD도 유지한다. 새 fixture 준비 실패/보존 기록·declared font/draw 서비스와 아직 미완료인 범위는 [R202 인벤토리](./PORTING-INVENTORY.md#중앙-hud-알림-초기화와-그림자의-native-반복-경계-r202)를 따른다.

## R201: 맵 생성 진행 문구의 native 포인터

사용자 goenv Go1.26.9를 계속 사용하며 `GO` 없는 공식 wrapper도 같은 Darwin/ARM64 설치를 선택한다. 실제 맵 생성 진행 문구에서 UTF-16 주소를 DWORD로 축소하던 저장/읽기를 기존 native pointer slot으로 바꾸고, 측정 폭은 포인터 대신 별도 int에 받으며3바이트 패킷의 sequence는 byte로 읽는다. 원본395바이트의681회 실행·동결 native681조건·실제 로컬/전체 C dispatcher의 영어/한국어384headless frames, 관련 공식 strict/race/checkptr2 3회30pass가 통과했다. 수정 전 두 실제 경로의 SIGSEGV와 새 fixture 준비 실패는 원문/입력으로 보존한다. 전체 일반/HD/CGo 각각73,728pass/47skip·server69,163pass/7skip 및 모든 기존 결과는 동일하다. 일반/HD headless Quest8화면·stage1→2/이동/종료도 통과했고 기존 기대값·원본/개인 데이터·R134 HOLD는 유지한다. 원본 Windows presentation/EAX와 기존 C short API의 구분, 통제 렌더링/음향 서비스 및 아직 미완료인 범위는 [R201 인벤토리](./PORTING-INVENTORY.md#맵-생성-진행-문구의-native-utf-16-포인터-r201)를 따른다.

## R200: 나침반·삼각함수 초기화의 native 경계

사용자 goenv Go1.26.9를 계속 사용한다. 활성 video initializer의 signed32 주소 비교 세 줄만 기존 카운터의4/32/320회 경계로 바꿨으며 이미지 로드/opaque handle·상태 reset·double 상수/sin/변환·store/반환은 유지한다. 원본 두 함수와 x87 helper 실행·동일 frozen fixture의 compiled RED 및 수정 후225조건·실제 CGo320 words/72파티클 좌표/1,056headless frames·관련 strict/race/checkptr2 3회612pass가 통과했다. 전체 일반/HD/CGo 각각73,725pass/47skip·server69,162pass/7skip와 기존 결과는 동일하다. 첫 전체 CGo 검사는 공유 캐시 손실로 실패했고, 원문 보존 후 전용 캐시의 동일 검사로 통과했다. 고정한 일반/HD headless Quest8화면·stage1→2/이동/종료도 통과했으며 기존 기대값·원본/개인 데이터·R134는 유지한다. 캐시/자동 checksum 복구·검증 범위·아직 미완료인 항목은 [R200 인벤토리](./PORTING-INVENTORY.md#나침반-이미지와-삼각함수-테이블의-native-초기화-경계-r200)를 따른다.

## R199: 효과 팔레트의 원본 ramp와 native 경계

사용자 goenv Go1.26.9를 계속 사용한다. 실제 맵 효과 팔레트 initializer에서 char 감소로 어두워진 첫 blue ramp를 원본 DWORD `+255`로 복원하고 네 signed32 주소 종료 조건을64/32/64/64회 순회로 바꿨다. 원본 RGB bytes/paired stores/반환과 기존 draw·시간·RNG 계약은 유지한다. 원본319바이트 실행768 RGB 호출·compiled control RED와 수정 후90조건/23,040호출·실제 CGo 맵256색/10headless frames·관련 strict/race/checkptr2 3회306pass가 통과했다. 전체 일반/HD/CGo 각각73,720pass/47skip·server69,160pass/7skip과 모든 기존 결과는 동일하다. unchanged 일반/HD headless Missile11경로/Bow3발을 확인했으며 새 private Bow 캡처는 동결 관찰이지 Windows/과거 golden이 아니다. 초기 Python/JIT/fixture 준비 실패·원본/개인 데이터/기존 기대값·R134 보존 및 아직 미완료인 범위는 [R199 인벤토리](./PORTING-INVENTORY.md#효과-팔레트의-원본-dword-색상-증가와-native-순회-경계-r199)를 따른다.

## R198: 보조무기 marker의 native 84칸 초기화

사용자 goenv Go1.26.9를 계속 사용하며 직접 실행과 GO 없는 wrapper도 같은 Darwin/ARM64 설치를 선택한다. 활성 보조무기 초기화의 signed32 주소 비교를 원래21행 카운터로 바꿔4열×21행을 처리한다. selector/encoder/sender·wire/반환·empty flag와 기존 기대값은 유지한다. compiled control RED와 수정 후87,720조건·실제 CGo1,360조건·관련 strict/race/checkptr2 3회1,164pass, 전체 일반/HD/CGo 각각73,716pass/47skip·server69,158pass/7skip 및 모든 기존 결과 동일을 확인했다. 원래 frozen inventory Screen의 일반/HD actual2는 이전 clean 제품에서도 동일한 실제 픽셀로 재현하며 그대로 보존한다. 별도 동결 기능 입력의 실제 마우스/보조무기 동기화·아이콘 확인 각2회는 일반/HD actual0이며 원래 Screen 통과를 대체하지 않는다. 준비 오류·실패 기록·원본/개인 데이터/R134 보존과 미완료 범위는 [R198 인벤토리](./PORTING-INVENTORY.md#보조무기-표시-초기화의-native-84칸-경계-r198)를 따른다.

## R197: 전사 스킬 상태의 native 다섯 칸 순회

사용자 goenv Go1.26.9와 `GO` 없는 wrapper 선택을 계속 확인한다. 실제 전사 스킬 상태/reset/활성 비트/조회 네 함수에서 signed32 주소 비교를 원래 다섯 레코드 수 경계로 바꿨다. 원본 scalar/packet/효과/시간/반환 계약과 기존 기대값은 유지한다. 최종 fixture의 compiled control actual1·수정 후11,088조건·실제 CGo7,280조건·관련 strict/race/checkptr2 3회1,101pass, 전체 일반/HD/CGo 각각73,713pass/47skip·server69,155pass/7skip 및 모든 기존 결과 동일을 확인했다. 일반 headless 다섯 스킬10cycles와 unchanged 일반/HD NPC Harpoon 각2cycles는 actual0이다. HD frozen Screen actual2는 이전 clean 제품도 같은 원본 다중 알림 문구의983픽셀 차이로 재현하며 교정/삭제하지 않았다. 초기 link 준비 실패·픽셀 실패/검증 범위·원본/개인 데이터/R134 보존과 미완료 항목은 [R197 인벤토리](./PORTING-INVENTORY.md#전사-스킬-상태-테이블의-native-다섯-레코드-경계-r197)를 따른다.

## R196: 동료 체력 테이블의 native 순회

사용자 goenv Go1.26.9로 계속 빌드·검사한다. 실제32칸 동료 HP 테이블의 조회/빈 슬롯/등록/reset에서 signed32로 축소한 주소 비교를 원래 슬롯 수 경계로 복원했다. decoder/ID·HP packing·scalar ABI·padding은 유지한다. compiled RED와 수정 후39,204조건·실제 CGo4,224조건·관련 strict/race/checkptr2 3회303pass, 전체 일반/HD/CGo 각각73,708pass/47skip·server69,150pass/7skip 및 unchanged 일반/HD headless 회복30경로/ally15기록이 통과했다. 기존 결과·기대값·원본·개인 데이터·R134 보류는 그대로이며 새 packet 준비 강화·검증 범위와 미완료 항목은 [R196 인벤토리](./PORTING-INVENTORY.md#동료-체력-테이블의-native-32칸-순회-경계-r196)를 따른다.

## R195: 퀘스트 결과 제목과 native 행 경계

사용자 goenv Go1.26.9를 계속 사용하며 `GO` 없는 wrapper 선택과 로컬 ARM64 runtime 백포트도 확인했다. 실제 Quest 결과 렌더러에서 원본에 없는 `XX1` 제목을 공백 하나로 복원하고, signed32로 축소한 포인터 종료 조건을 원래 여섯 행의 번호 경계로 바꿨다. compiled RED와 수정 후12,096조건·관련 strict/race/checkptr2 3회96pass, 전체 일반/HD/CGo 각각73,703pass/47skip·server69,145pass/7skip 및 headless stage1→2/정상 결과 제목/이동/종료를 확인했다. 기존 결과·기대값·원본·개인 데이터·R134 보류는 그대로다. 새 fixture 준비 실패/전사 교정과 private 캡처의 제한, 미완료 범위는 [R195 인벤토리](./PORTING-INVENTORY.md#퀘스트-결과-화면의-원본-제목과-native-여섯-행-경계-r195)를 따른다.

## R194: 퀘스트 보상·이름 변경 알림의 native 이름

사용자 goenv Go1.26.9를 계속 사용한다. 실제 C fallback의 스펠/도감/스킬 보상과 이름 변경 네 알림 분기에서 PE32 이름 위치를 native Player 필드로 복원했다. 이미 Go가 처리하는 hint C 중복은 건드리지 않고 원래 ID/분기/helper/settings/packet 계약을 유지한다. 기존 본체 compiled RED와 수정 후2,672조건·actual CGo156조건·strict/race/checkptr2 3회3,072pass를 확인했다. 전체 일반/HD/CGo 각각73,701pass/47skip·server69,143pass/7skip이며 기존 결과는 모두 동일하다. headless Quest stage1→2/이동/정상 종료도 확인했다. 자동 생성된 신규 캡처8개는 소스 밖에 보관하고 같은 private 입력으로 재검증했으며 기존 기대값·원본·개인 데이터·R134 보류는 유지한다. 범위와 초기 실패·미완료 항목은 [R194 인벤토리](./PORTING-INVENTORY.md#퀘스트-보상이름-변경-알림의-native-이름-접근-r194)를 따른다.

## R193: 점수·보물 알림의 native Player 필드

사용자 goenv Go1.26.9를 계속 사용한다. 실제 C dispatcher의 점수/보물 패킷 두 arm에서 PE32 위치에 쓰던 Player 데이터와 이름을 native 필드로 복원하고 해당 packed 숫자를 바이트로 읽도록 수정했다. 원본 signed Elimination·unsigned strict 보물 시각·호스트 권한·cached DWORD wrap 알림 조건과 helper 순서/packet ABI는 유지한다. 동결된 기존 본체의 compiled RED, 수정 후 3,648조건, 실제 CGo 51경로 및 strict/race/checkptr2 3회 552pass를 확인했다. 전체 일반/HD/cgocheck2 각각 73,698pass/47skip·server69,140pass/7skip이며 모든 기존 결과와 기대값은 불변이다. unchanged headless 일반/HD 보물 획득·드롭은 actual0이고 원본/개인 데이터/R134 보류도 유지한다. 전체 라운드·Windows/물리 오디오/전체 포팅 완료 인증은 아니며 [R193 인벤토리](./PORTING-INVENTORY.md#점수보물-알림-패킷의-native-필드-복원-r193)에 범위와 기록을 구분한다.

## R192: 사망 알림의 native 이름 포인터

Go1.26.9 ARM64에서 사망 알림이 Player 이름을 PE32 위치에서 읽고 주소를32비트로 축소하는 활성 C 경로를 typed native 이름 필드로 수정했다. 원본 출력 순서·버퍼·packet ABI는 유지한다. 이전 본체 control은 같은 fixture에서 실패하고 수정 후62조건·실제 CGo packet 경로 및 관련 strict/race/checkptr2 반복636pass를 확인했다. 전체 일반/HD/cgocheck2는 각각73,696pass/47skip, server는69,138pass/7skip이며 모든 기존 결과는 그대로다. unchanged 일반/HD headless 게임에서 실제 HP75→0/사망/콘솔 알림과 정상 종료를 확인했고, 초기 테스트 준비 실패·원본·개인 데이터·R134 보류를 보존했다. 상세 범위와 미완료 경로는 [R192 인벤토리](./PORTING-INVENTORY.md#사망-알림의-native-플레이어-이름-주소-복원-r192)를 따른다.

## R191: Host Game 시작 상태와 관전자 표시 검증

Goenv Go1.26.9를 계속 사용한다. 기존 Host 시작 PNG의 actual2를 보존하고 현재 native status0x10의 숨겨진 eye/방패대기 상태와 원본 low-bit 표시 조건을 대조했다. 일반 입력의 보행·공격·방패복귀·관전자 메뉴 전환은 정상/HD headless 두 실행에서 actual0이고 여섯 실제 캡처도 각각 동일하다. 게임 코드를 바꾸지 않고439조건의 compiled native 회귀를 추가하며 역사적 고정 오프셋 control actual1을 보존한다. 관련 실제 strict CGo/race/checkptr2/count3는276pass/0skip다. 기존 기대값·PNG·원본·개인 데이터·R134 HOLD는 그대로이며 상세와 제한은 [R191 인벤토리](./PORTING-INVENTORY.md#host-game-시작-상태와-관전자-아이콘의-native-계약-r191)를 따른다.

## R190: 원격 포털과 전용 서버 단일 접속자 종료 복원

사용자 goenv Go1.26.9를 계속 사용한다. 실제 시계/headless 일반 이동으로 native 포털·출발/도착 효과·위치/HUD·원본 animation 계약을 확인했다. 전용 서버가 로컬 host를 잘못 빼서 원격1명일 때 quit/ACK를 생략하던 산술만 수정했으며 compiled RED/GREEN과 실제199→200·정상 종료가 통과했다. 전체 일반/HD/strict는 각각73,693pass/47skip, server69,135pass/7skip이고 모든 기존 결과를 유지하며 새70pass만 추가했다. 초기 진단 실패·원본·개인 데이터·기존 기대값·R134 HOLD는 유지한다. 상세 범위와 캐시 정리·아직 남은 전체 목표는 [R190 인벤토리](./PORTING-INVENTORY.md#원격-포털-통과와-전용-서버의-단일-접속자-종료-ack-r190)를 따른다.

## R189: 원격 Force of Nature의 반사·직접 피격 검증

사용자 goenv Go1.26.9를 계속 사용한다. 게임 코드 수정 없이 실제 시계/headless 정상 주문책·이동·시전 입력에서 native 충전/투사체·단일 마나60 소모를 확인했다. 방패를 든 정지 전사의 반사→시전자 사망·사망 마나0, 정상 이동 중 전사의 직접 피격 HP150→0·표시150을 각각 확인했으며 일반 클릭 부활·자연 투사체/효과 제거·실제 두 접속자의 quit/ack와6 process 정상 종료도 통과했다. 관련 기존 strict CGo/race/checkptr2 반복은2,502pass/0skip다. 초기 세 진단 실패·부분 맵 decoder 실패·원본·개인 데이터·기존 기대값·R134 보류는 유지한다. 이는 모든 원격 주문/맵이나 전체 포팅 완료 인증이 아니며 [R189 인벤토리](./PORTING-INVENTORY.md#원격-force-of-nature의-방패-반사와-직접-피격-검증-r189)에 범위를 기록한다.

## R188: 원격 소환술사의 Toxic Cloud 수명 검증

Goenv Go1.26.9를 계속 사용한다. 게임 코드 수정 없이 실제 시계/headless 소환술사→전사의 정상 주문책·단축키 입력에서 시전 큐·단일 마나60 소모·독구름 피해·초록 중독 HUD·정상 이동 회피를 확인했다. 원본450틱 수명 후 독구름과 양쪽 drawable이 자연 제거됐고 중독 자동 해제·체력 관 복원·실제 접속자2명의 quit/ack 및 정상 종료도 확인했다. 관련 기존 strict CGo/race/checkptr2 회귀3회는4,506pass/0skip다. 최초 진단 helper 타입 오류를 보존하고 기존 소스·기대값·원본·개인 데이터·R134 보류는 그대로 유지한다. 모든 주문/직업/맵이나 실제 오디오 출력 인증은 아니며 [R188 인벤토리](./PORTING-INVENTORY.md#원격-소환술사의-toxic-cloud-시전과-자연-수명-검증-r188)에 범위를 기록한다.

## R187: 원격 Magic Missile과 정상 주문책 입력

사용자 goenv Go1.26.9를 계속 사용하며 GO 없는 wrapper 선택도 확인했다. 게임 코드 수정 없이 실제 시계/headless 원격 Wizard→Warrior Magic Missile을 기본 단축바와 정상 주문책 드래그 두 경로로 검증했다. 레벨3·4발의 native 소유자/추적 대상·단일 마나15 소모·서버와 양쪽 client의 피해8/16 및 정상 종료를 확인했다. v2에서는 접속자2명의 실제 quit/ack를 검사했고 초기 잘못된 종료 URL은 성공 증거에서 제외했다. 기존 소스·기대값·원본·개인 데이터·R134 보류는 그대로다. 관련 기존 엄격 CGo/race/checkptr2 회귀를 3회 반복해 855pass/0skip를 확인했다. 모든 발의 명중·모든 주문/직업/맵 인증은 아니며 자세한 범위는 [R187 인벤토리](./PORTING-INVENTORY.md#원격-magic-missile과-정상-주문책-입력-검증-r187)를 따른다.

## R186: 원격 Fireball과 전용 서버 종료 검증

사용자 goenv Go1.26.9로 계속 빌드·검사한다. 실제 시계/headless 입력의 Wizard→Warrior Fireball 두 실행에서 정상 시전 큐·단일 마나150→120 소모·서버와 client HP150→23 및 표시127을 확인했다. 실제 접속자2명인 전용 서버의 quit가 로컬 GUI 글꼴에서 nil 접근하던 크래시는 기존 GUI 네 작업만 dedicated build에서 제외해 수정했다. 같은 새 회귀 RED/GREEN과 실제 두 접속자의 quit/ack 및 정상 종료를 확인했고 전체 일반/HD/CGo 각각73,623pass/47skip·server69,065pass/7skip, 관련 엄격/race/checkptr2 반복111pass가 통과했다. 기존 결과·기대값·원본·개인 데이터·R134 보류와 초기 진단 실패를 유지한다. 모든 주문·직업 쌍·맵의 인증은 아니며 상세는 [R186 인벤토리](./PORTING-INVENTORY.md#원격-fireball-실전-검증과-전용-서버-종료-gui-분리-r186)를 따른다.

## R185: 전용 서버의 정상 멀티플레이 주문 초기화

Goenv Go1.26.9를 계속 사용하고 GO 없는 wrapper 선택도 확인했다. 전용 서버가 메뉴 시작의 EngineAdmin 설정/God Mode 해제를 빠뜨려 실제 원격 캐릭터의 주문 레벨이0으로 남는 문제를 같은 두 flag 연결로 수정했다. 동일 실제 시계/headless 입력에서 투명화 시전 큐·마나150→120 단일 소모·server/client 효과·정상 종료를 확인했다. 관련 엄격/race/checkptr2 반복2,592pass, 전체 일반/HD/CGo 각각73,622pass/47skip·server69,064pass/7skip이며 기존 결과와 기대값은 그대로다. 초기 입력 오류와 RED를 보존하고 자체 작업 종료 뒤 재생성 가능한 프로젝트 빌드 캐시만 정리했다. 모든 원격 주문·전투나 전체 포팅 완료 인증이 아니며 상세 범위는 [R185 인벤토리](./PORTING-INVENTORY.md#전용-서버의-멀티플레이-주문-초기화-복원-r185)를 따른다.

## R184: 마법사·소환술사 원격 접속 검증

같은 Go1.26.9 게임 소스에서 실제 시계/headless GUI 입력으로 마법사와 소환술사의 원격 클래스·체력/마나/속도 초기화·2초 이동·15초 높이 갱신·서버 정상 종료를 확인했다. 두 client/server 네 process는 actual0이며 게임 코드·원본·개인 데이터와 기존 테스트는 그대로다. 진단의 초기 타입 오류를 보존하고 문서만 추가한다. 이는 원격 마법·전투나 전체 포팅 완료 인증이 아니다. Player SetColor의6slot/5RGB 저장·전송 경계는 남겨두며 상세는 [R184 인벤토리](./PORTING-INVENTORY.md#원격-마법사소환술사의-실제-시계-접속-검증-r184)를 따른다.

## R183: 전용 서버 초기화와 원격 높이 보고 연결

사용자 goenv Go1.26.9로 계속 빌드·검사한다. 전용 서버가 메뉴 초기화를 건너뛰어 실제 원격 플레이어의 체력·속도가0으로 남는 문제를 기존 gamedata 서비스 연결로 수정했다. 정상 초기화 후 드러난15초 높이 갱신의 포인터 절단 크래시는 기존 native reporter로 호출을 교체했다. 실제 시계/headless 접속에서 HP150·speed2.1·2초 이동과 정상 높이 갱신, 서버 종료를 확인했다. 관련 strict/race/checkptr2 반복324pass 및 전체 일반/HD/CGo 각각73,619pass/47skip, 서버69,061pass/7skip이다. 기존 결과·기대값·개인 데이터·R134 보류와 초기 실패·공간 확보를 위한 중단 기록을 유지한다. 범위와 증거는 [R183 인벤토리](./PORTING-INVENTORY.md#전용-서버-능력치-초기화와-원격-높이-갱신-r183)를 따른다.

## R182: 원격 맵 로드 전에 로컬 클라이언트 초기화 유지

사용자 goenv의 Go 1.26.9 경로를 계속 사용한다. E2E 가상 시계 없이 실제 dedicated server에 headless 입력으로 접속한 결과, native Join이 `GameClient`까지 해제해 맵의 `ObjectTOC`를 빈 유형 테이블에 읽는 크래시를 재현했다. 호스트 역할만 해제하도록 기존 마스크 하나를 수정했다. 수정 후 유형 2,875개·맵 입장·서버 ready·플레이어 표시를 확인했고 일반/HD/엄격 CGo 전체는 각각73,612pass/47skip, 서버는69,054pass/7skip이다. 이전 테스트 결과·원본 자료·R134 보류는 유지한다. 이동 입력은 실제 server까지 도달했지만 좌표가 그대로인 별도 실패는 남아 있으며, 이를 성공이나 전체 포팅 완료로 세지 않는다. 준비 입력 오류와 공간 부족 실행을 포함한 [R182 인벤토리](./PORTING-INVENTORY.md#원격-맵-로드의-로컬-client-역할-유지-r182)를 따른다.

## R181: native 서버 참가 상태와 프레임 소유권

사용자 goenv의 Go 1.26.9를 직접 사용했고, `GO`를 생략한 wrapper도 같은 설치를 선택함을 확인했다. native 서버 목록에서 캐릭터 선택으로 넘어갈 때 사라진 참가 플래그와 접속 상태 갱신을 복원했다. 실제 UDP 응답·기존 C dispatcher의 승인/비밀번호/실패/재시도/타임아웃 검사가 통과했다. 일반·HD·엄격 CGo 전체는 각각73,611pass/47skip, 서버는69,053pass/7skip이며 기존 테스트 결과는 불변이다. 원본 자료와 R134 보류도 유지한다. 실제 맵 진입은 아직 미검증이며 가상 시계 client와 실시간 server의 혼합 검사에서 발생한 본 접속 timeout을 성공으로 보고하지 않는다. 범위와 실패 기록은 [R181 인벤토리](./PORTING-INVENTORY.md#native-서버-참가-플래그와-접속-frame-소유권-r181)에 구분한다.

## R180: Glyph 직접 callback의 native 회귀 검사

실제 C parser·고주소 local-player slot·C GlyphDraw/AnimateDraw·전체 raster/headless upload를 기존 게임 경로와15조건 비교했다. 같은 pointer의 fallback alpha와 실제 negative-height clip cache를 보존한 frozen 검사는 통과했다. 초기 새 테스트의 C allocation 초기값/연속 cache 준비 실패와 원문은 남긴다. 게임 production 수정 없이 새 test 한 파일만 추가하며 strict/race/checkptr2 반복57pass와 전체 일반/HD/CGo/server·원본 자료 대조가 통과했다. 이는 Windows stock pixel·physical SDL/GPU 인증이 아니다. 상세 범위는 [R180 인벤토리](./PORTING-INVENTORY.md#glyph-등록-c-경로의-고주소alpha애니메이션-검증-r180)를 따른다. 사용자 goenv Go1.26.9·R134 보류를 유지하고 R179 clean runtime 검증은 완료했다. 전체 포팅 목표는 계속 진행 중이다.

## R179: 실제 SDL resize event의 viewport 알림

실제 SDL dummy queue/InputTick에서 호출되지 않던 resize callback을 연결했다. 해당 window의 drawable 크기를 조회해 같은 크기·다른 창을 제외하고 첫 mouse input 이전 renderer viewport를 갱신한다. SDL 두 production files와 release에서 제외되는 opt-in test/queue fixture 두 파일만 변경했으며 renderer geometry·FPS·기존 tests/기대값은 유지한다. 동일 frozen RED/GREEN, strict/race/checkptr2 33pass 및 전체 일반/HD/CGo/server와 원본 자료 대조가 통과했다. 새 테스트의 API/checkptr 준비 실패와 viewport 비교 교정은 보존한다. physical Retina/OpenGL 검증이 아닌 headless 검사이며 상세 범위는 [R179 인벤토리](./PORTING-INVENTORY.md#sdl-native-resize-알림과-첫-입력의-viewport-연결-r179)를 따른다. 사용자 goenv Go1.26.9 정책과 R134 보류를 유지하고 R178 clean runtime 검증은 완료했다. 전체 포팅 목표는 계속 진행 중이다.

## R178: 소환 효과 직접 콜백의 native 수명 연결

실제 고주소 소환 packet/파서/렌더러 검사에서 기존 world 경로는 통과했지만 직접 등록 콜백은 PE32 오프셋 때문에 여덟 조건 모두 SIGSEGV를 일으켰다. 해당 C entry를 기존 Go 소환 효과에 연결한 네 production files와 새 test 하나만 변경했다. 동일 v5 RED/GREEN, strict/race/checkptr2 45pass와 전체 일반/HD/CGo/server 회귀 및 원본 자료 대조가 통과했다. 새 테스트 준비 실패·할당 주소 관찰 수정도 보존하며 Windows/SDL/GPU 전체 인증으로 확대하지 않는다. 자세한 범위는 [R178 인벤토리](./PORTING-INVENTORY.md#summoneffect-직접-등록-콜백의-native-child수명-연결-r178)를 따른다. R177 clean runtime 검증은 완료됐고 전체 포팅 목표는 계속 진행 중이다.

## R177: 영상 렌더러의 surface 해제와 resize 재그리기 차단

영상 종료·early close·audio-source 생성 실패 후 surface가 남고 headless resize에서 현재 GUI가 마지막 영상/빈 화면으로 덮이는 경로를 실제 stock movie로 재현했다. renderer의 graphics-thread Close와 MoviePlayer의 해당 renderer cleanup만 추가하며 active GUI·decode·FPS·viewport 계산은 유지한다. frozen RED/GREEN, 실제27resize 비교, renderer storage/lifecycle 및 기존 영상 검사와 전체 회귀가 통과했다. SDL backend는 현재 resize callback을 dispatch하지 않으므로 실제 SDL ghost redraw나 GPU 메모리 readback을 인증하지 않는다. 상세 범위는 [R177 인벤토리](./PORTING-INVENTORY.md#종료된-영상-렌더러의-surface화면-버퍼-수명-복원-r177)에 기록한다. R176 clean runtime 검증은 완료됐고 전체 포팅 목표는 계속 미완료다.

## R176: 영상 생성 실패·입력 종료의 자원 정리

영상 constructor 실패 때 열린 파일8개가 남는 누수와 실제 stock 영상의 mouse/key/window cancellation10회 중7회에서 버퍼가 남는 문제를 재현했다. 입력은 중단만 요청하고 decoder/audio worker 종료 뒤 source·buffer·file을 정리하도록 한 MoviePlayer 파일과 새 tests2개만 바꿨다. 같은 frozen tests의 strict/race/checkptr2·반복 종료·130frame 자연 재생·early close 및 전체 일반/HD/CGo/server 회귀가 통과했다. OS FD 관찰 adapter의 첫 준비 실패와 v1은 보존하며 스피커/Windows 전체 인증으로 확대하지 않는다. 상세 범위는 [R176 인벤토리](./PORTING-INVENTORY.md#영상-종료와-생성-실패의-파일오디오-자원-수명-복원-r176)에 기록한다. R175 clean runtime 검증은 완료됐고 전체 포팅 목표는 계속 미완료다.

## R175: Goenv Go 1.26.9 실제 오디오와 음악 폴더 경로

실제 headless/OpenAL-null 일반·HD에서 걷기·달리기·공격·피격·장비 획득/드롭 여섯 phase의 새 FX buffer 소비를 확인했다. 기존 strict stream267회·원본 효과음1780개 전체·macOS 기본 출력의 무음10개도 통과했다. 기존 배경음 opt-in에서 점이 있는 상위 폴더를 확장자로 오인하는 실패를 발견해 음악 filename만 검사하도록 한 조건을 수정했다. 같은 음원·테스트·경로의 재검사는 전체 약33.9초 반복을 통과했고 전체 일반/HD/CGo/server의 이전 결과와 기대값도 같다. 청음·Windows 전체 동등성 인증이 아니며 자세한 실패/성공과 범위는 [R175 인벤토리](./PORTING-INVENTORY.md#goenv-go-1269-실제-오디오-검증-및-점이-있는-음악-폴더-경로-수정-r175)에 기록한다. R174의 clean runtime 검증은 완료됐고 전체 포팅 목표는 계속 미완료다.

## R174: 실제 옵션 audit 네 경로 검증

기존 main-menu/in-game options audit를 사용자 goenv Go1.26.9의 clean ARM64 일반·HD 제품으로 실행해 각각226/228/201/203, 합계858항목이 통과했다. 실제 UI input·상태·메모리상 직렬화 검증이며 mock audio Target/flags, E2E defaults와 video/config suppression을 physical playback·disk restart로 확대하지 않는다.18private 녹화 PNG는 canonical 이미지 대조가 아니고 기존 YAML/PNG와 모든 원본/개인 파일은 그대로다. 이 단위는 문서만 추가하고 게임 동작·기대값을 바꾸지 않는다. 상세 범위와 receipt는 [R174 인벤토리](./PORTING-INVENTORY.md#실제-메뉴인게임-옵션-전체-audit-재검증-r174)를 따른다. R173의 clean3제품 및 일반/HD headless 후속 검증도 완료됐으며 전체 포팅 목표는 계속 미완료다.

## R173: 비디오 초기화의 native 콜백 저장 ABI

Go1.26.9 기준에서 렌더 타깃 초기화의 C 콜백 주소가 int/DWORD로 축소되던 setter·선언·Go 전달만 native pointer와 공용 side slot으로 복원했다. 동결된 compiled RED, 실제 C callback과 원본 descriptor 대조, strict/race/checkptr2 및 전체 일반/HD/cgocheck2/server·oracle 검사가 통과했다. 현재 renderer의 production slot-reader가 확인된 것은 아니며 화면/Windows renderer 전체 해결로 확대하지 않는다. 원본 도구의 환경/stack-guard 초기 실패와 최종 검증 범위는 [R173 인벤토리](./PORTING-INVENTORY.md#비디오-초기화-콜백-슬롯의-native-주소-보존-r173)를 따른다. R172의 clean 3제품·일반/HD headless·dedicated server 검증은 후속 완료됐고 사용자 goenv Go1.26.9 정책은 유지한다.

## R172: 사용자 설치 goenv Go 1.26.9

현재 pin·프로젝트 .go-version·빌드/제품 검사와 launcher는 Go1.26.9다. 명시한 GO를 존중하고 이미 설치된 goenv 1.26.9를 우선 사용하며 전역 기본값 1.25.12는 보존한다. [ARM64 GC 공식 백포트](https://github.com/golang/go/issues/80394)가 포함된 설치본으로 공식 재현 100회, 기존 OptionsDialog GC 진단 300회, 전체 일반/HD/cgocheck2/server 테스트 및 원본 oracle 대조가 통과했다. 구 버전의 재현 실패·도구체인 선택 RED와 기존 기대값은 유지한다. 자세한 범위·횟수와 아직 미완료인 clean runtime 검증은 [R172 인벤토리](./PORTING-INVENTORY.md#go-1269-및-설치된-goenv-도구체인-선택-r172)를 따른다. 아래 과거 Go1.26.5 기록은 역사적 증거로 보존한다.

## R114: visible LavaFountain의 native FLAME admission

- 기준은 clean/pushed R113 `e3789507db24f36ac8b284ac5461b3674f01e87b`다. 보고된 NewPlayer/type713/class2621444/flags16777732에 source==weapon LavaFountain/type788/class12582912/subclass0/flags16777748의 raw3/type1(FLAME)이 거절됐다. 이 class `0xc00000`은 IMMOBILE|VISIBLE_ENABLE로, 앞서 포팅한 LavaFountain3의 DANGEROUS|IMMOBILE(`0x410000`)과 다르다. 기존 `playerDamageWorldFlameShape4E17B0` 한 본체에 새 stationary-visible subset만 추가해 PlayerDamage/DefaultDamage가 기존 FLAME 경로를 공유하게 한다. 객체 이름/ID·피해량/유형·HP·면역은 바꾸지 않고 unit/weapon/wand/missile 및 non-FLAME 배제는 유지한다. signature/imports/본체 밖 bytes 및 기존 테스트/입력/자산/Save/config는 불변이다.
- unchanged GAME.EXE SHA=`0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`, PlayerDamage2368B SHA=`c3e71619fd8d5e8c0aff27b5d098db02ee5bed0c6827f495ce6cf54326062ed9`의 bounded228조건과 sealed6범위를 확인한다. 실제 unhooked entry는 PlayerDamage228/FISTP56/shield-selection24다. signed raw[-3,0,3,19], Quest none/0/.5, player/NPC·normal/Reflect/dead/NoUpdate/invulnerability/frame/god/shield-front/back/observer-disabled 경계를 비교한다. buff/mode/facing/audio/balance/armor-wear/DefaultDamage는 declared services며 HP 인자 경계는 Windows HP 인증이 아니다. source/guards/code 불변·memory fault0, final v4/JSON/동결 repo 재실행은 byte-identical SHA=`46ee9e5e9b145f4efcaa5263020a745474b6f275a4a8b49c94200c0cdb74ab6b`다. 초기 nil-shield synthetic 입력의 원본 +16 read fault actual1 두 건과 v1/v2/v3 원문은 보존한다; 최종 bounded shield 입력은 equipped armor record를 공급하며 원본 명령/서비스 결과를 바꾸지 않는다. 전체 original world/GUI·실제 caller CW·Windows item/death 동등성은 UNPROVEN이다.
- 최초 compiled RED306IDs/152pass/154test-fail/2package-fail·actual1 뒤 새 test2/JSON/probe4파일을 동결했다. 같은 GREEN306pass/0skip·actual0이다. RED는 class/type admission 및 기존 generic/shield fallback 순서 차이이고 컴파일/포인터 크래시가 아니다. C-owned >4GiB 실제 C dispatcher→registered PlayerDamage/DefaultDamage→UnitSetHP의11조건·3연타에서 nonimmune6조건 HP60→57→54→51, monster/NPC immune5조건 HP60 유지와 marker/attribution/frame/sync/carry/armor/status를 검증한다. damage/HP callback은 대체하지 않았다. 원본 prefix228조건은 인자/순서 비교이고 actual HP 증거는 이 C 경로다.
- Go1.26.5/GOMAXPROCS2/`-p=2` 직렬 일반·HD·cgocheck2 전체 각각51,480pass/41skip/38packages, server50,641pass/7skip/35packages·actual0이다. 모든 R113 ID/action/횟수/skip/package 결과를 유지하고 새306pass만 추가했다. strict 및 race+checkptr2+cgocheck2 count3은4,468IDs/13,404pass·0skip/0fail/5packages·actual0이다. 각 반복 lane에서 신규 actual C HP33기록(18nonimmune/15immune), 기존 NPC Sword HP2000→1997→1994→1991 세 번 및 DeathBall no-health72기록을 보존한다. original2,968code/650data verifier actual0, test audit SHA=`3ec46fed251e910a16344fdbae7ce9833b16d81a797c6f451b642310299e8d10`, one-body audit SHA=`20106c434f710b3136d8ce0cb31e847b9109a9c114a264a82cce55f9a0295c76`, functional/tested byte-identical SHA=`02fb0ff4708f8a19ecfccb089380591ba819b7eec27c17301938819dc1272389`다.
- 증거 `/private/tmp/opennox-visible-fountain-r114.EEXk6L`. protected303files/개인1562files/stock1556files 불변·삭제0이다. 커밋 직후 즉시 push 후 clean ARM64 일반/HD/server3제품 및 unchanged 환경 피해 YAML 일반/HD headless/mock를 독립 runtime ledger로 검증한다. 환경 GUI는 Flame/Spike/PeriodicSpike/SpikeBlock/SpikeBlockImmobile/RotatingSpikes/RotatingSpikesImmobile/Lava×player/monster/NPC 회귀이며 exact LavaFountain788 GUI라고 확대하지 않는다. HP/RNG/WAIT/aggression/timeout/PNG/golden retune은 없다.
- 선행 R113 DeathBall 작업은 `e3789507db24f36ac8b284ac5461b3674f01e87b` 커밋·즉시 push와 clean ARM64 9products metadata, unchanged projectile/environment/Wiz08e original Tower timer 일반/HD GUI6개 actual0까지 완료했다. runtime audit SHA=`bc06b453ecdcd272fbd9dcc897427d87d989ccaf6548feb00e68e3528428b4f0`다. 초기 private audit의 존재하지 않는 quit Log 가정 actual1은 보존하고 새 audit에서 unchanged YAML terminal quit+실제0 exit+cleanup을 확인했다; 게임 입력/코드/결과 retune이 아니다. 아래 과거 문서 suffix bytes를 보존하고 natural Skeleton·Con02A·passive-food·Quest-stage3와 전체 포팅 목표는 미완료로 유지한다.

## R113: DeathBall의 체력 없는 맵 오브젝트 피해 tail

- 기준 clean/pushed `38e40db852504c979b07901b4770e674878d7c8b`다. 보고된 PlayerStart/type714/class0/raw1/type2 및 InvisibleTeleportPentagram/type717/class4195328/raw50/type2는 DeathBall/type706/class1572865, weapon=nil 충돌이다. `DefaultDamageWorld4E0B30` 한 본체에서 non-unit의 nil-health를 누락 서비스로 오인하던 admission만 좁힌다. 원본처럼 위치·출처·buff/sound를 처리한 뒤 실제 UnitDamageClear의 nil-health/zero-Max gate에서 안전하게 끝난다. 맵 오브젝트에 체력이나 삭제 가능성을 부여하지 않으며 unit-health/필수 서비스 검사는 유지한다. signature/imports/본체 밖 bytes·기존 테스트/입력/자산/Save/config는 불변이다.
- unchanged GAME.EXE DefaultDamage1792B SHA=`6f045c2910bfb5e4a1100b5daaed3aeb5695bb401d3b447c63245c3543e0b871`와 UnitDamageClear264B SHA=`2aa2fdff219fdcd15ffdaaf41783e5b00e412ad76fa4a7fdf74baf4449f8fbdf`를 포함한 sealed11범위로160조건을 실행했다. nil/zero-Max·nil/self weapon·normal/NoUpdate/dead/invulnerable/Shield-zero·signed raw 경계를 확인했다. 원본 root return/metadata 및 nil/Max health gate는 hook이 아니고 flags/buffs/enemy/audio/Shield는 declared services다. supplied CW027f, executable/source/health/guards 불변·memory fault0이다. transcript SHA=`33de3f05a3c6afef6509c925e27d72529d574ab8af3a86cb35249eeb6c687fdd`; frozen repo probe는 private probe의 마지막 blank line 하나만 다르며 재실행 JSON은 byte-identical이다. 전체 원본 world/Windows HP·death/UI 동등성 인증은 아니다.
- 최초 compiled RED193IDs/151pass/42test-fail/2package-fail·actual1 뒤 새 test2/JSON/probe4파일을 동결했다. 같은 GREEN193pass/0skip·actual0이다. 실제 C-owned >4GiB dispatcher→registered DefaultDamage→UnitDamageClear의24조건·2연타에서 nil/zero-Max HP120→120→120, ordinary HP120→119→118 또는120→70→20 및 whole-record attribution/sync를 확인했다. 피해/HP callback은 대체하지 않았다.
- Go1.26.5/GOMAXPROCS2/`-p=2` 직렬 일반·HD·cgocheck2 각각51,174pass/41skip/38packages, server50,335pass/7skip/35packages·actual0이다. R112의 모든 기존 ID/action/횟수/skip/package terminal을 유지하고 새193pass만 추가했다. strict 및 race+checkptr2+cgocheck2 count3은4,162IDs/12,486pass·0skip/0fail/5packages·actual0이다. 기존 NPC Sword HP2000→1997→1994→1991도 보존한다. original2,968code/650data verifier actual0, functional/tested checkpoint SHA=`db8567790c454c0be4282d97d1d55a65db062a71b6aaafbbf1ab368fc88da539`, test audit SHA=`ec68b93923167da2e57c4f006e117aab063db3619baa3da2194c1519fec8bf1d`, one-body audit SHA=`6b99375f3f28d3797e67f1262f10dee46691f8bca9892b7dd417671741c401ee`다. 초기 audit의 정보성 Log 집계/EOF blank-line 판정 오류 actual1도 원문 보존한다; frozen 테스트·생산 코드 변경으로 맞추지 않았다.
- 증거 `/private/tmp/opennox-world-projectile-r113.ROjfA1`. protected303files 및 개인/stock 설치는 불변·삭제0이다. 커밋 직후 push 후 clean ARM643제품과 unchanged 투사체·환경 피해·Wiz08e 원본 Tower timer 일반/HD headless/mock를 별도 runtime ledger로 확인한다. 기존 시나리오의 배치/HP/RNG/WAIT/timeout/golden은 바꾸지 않으며 GUI 회귀를 exact no-health 충돌 또는 전체 original GUI 인증으로 확대하지 않는다. 추가 보고된 LavaFountain/type788/class12582912/raw3/type1은 다음 별도 수정 대상이다. 과거 문서 suffix bytes와 미완료 전체 포팅 게이트는 보존한다.

## R112: NPCDraw의 원본 색상 초기화 슬롯과 효과 뒤 실행 순서

- 기준 clean/pushed `cc732b2eb38ce79c2a27c7ee6f6352902e808f29`다. `src/drawable_npc.go:(*Client).DrawNPC` 한 본체에서 slots1..6의 조기 초기화를 제거하고 top effects/HP 뒤 slots0..5를 초기화한다. 원본의 slot6 유지와 missing-NPC/zero-frame early return은 보존한다. signature/imports/본체 밖, 기존 테스트/YAML/PNG/golden/자산/Save/config는 불변이다.
- unchanged GAME.EXE `004B9050`1047B를 74조건 실행했다. code SHA=`f5d3eb7b66c8b8cec76784cf0b1f1e1e3d34a075491a57511650bbe387ac545c`, 원본 transcript SHA=`079ba9fe21d1084dee49fc941d2933ca5d52f6d4a9c30f589bb329d4397b1178`. executable/guard/입력 bytes 불변·memory fault0이다. material/frame/empty equipment/cached top-effect/5551 split-pack/non-ally는 declared services이며 white7fff/blue001f는 통제 색상이다. stock blue RGB 또는 stock sprite/HP bar/preview/negative-frame/full-original-world pixels를 증명한 것이 아니다.
- 최초 compiled RED75IDs의2pass/73test-fail/1package-fail·actual1 뒤 새 test/C recorder/JSON/probe4파일을 동결했다. 같은 GREEN75pass/0skip·actual0이다. C-owned >4GiB NPC/animation/drawable/opaque RenderData와 실제 등록 C draw callbacks72회, 실제 indexed renderer의16slots×4lumas 및 headless presentation146회/9344pixels로 원본 palette·효과 순서·최종 초기화를 비교한다. 공급된 read-only top-effect callback은 stock skull/crown sprite 자체가 아니다.
- Go1.26.5/GOMAXPROCS2/`-p=2` 직렬 일반·HD·실제cgocheck2 각각50,981pass/41skip/38tested-packages, server50,142pass/7skip/35tested-packages·actual0이다. R110의 모든 기존 ID/action/multiplicity/skip/package terminal을 보존하고 !server에 새75pass와 test-only helper의 no-test-files terminal 하나만 추가했다. strict 및 cgocheck2+race+checkptr2 count3 각각3,969IDs/11,907pass·0skip/0fail/5packages·actual0이며 NPC Sword HP2000→1997→1994→1991도 유지한다.
- test audit SHA=`20e21bee2af5d0d4ebe865d243e5d34ce245340a7e4c6e170cee478a69867fe0`, one-body audit SHA=`1f2f528e4714ce0e24327efd270f794bd4613db0c3930ecbe903dcedf6bb8326`, functional/tested 동일 checkpoint SHA=`8d6717bb1409f2a02f40312636c62f1ccaa1e866804d20d0d6b4a730ad7d7f7d`. original2,968code/650data verifier actual0, protected303rows·개인/stock 설치 불변·삭제0이다. 증거 `/private/tmp/opennox-npc-palette-r112.HZ84s7`에서 커밋·즉시 push 후 clean ARM643제품과 unchanged WoundedApprentice/death-spawns/cheat-load/Troll-cloud 일반·HD headless/mock, frozen Con02A 진단을 별도 runtime ledger로 확인한다. Con02A Screen 실패를 성공으로 간주하거나 regolden하지 않는다. R107/R108 사용자 요청은 이미 완료했지만 natural Skeleton timeout·Con02A·passive-food·Quest stage3 및 전체 포팅 목표는 남는다. 아래 과거 문서 suffix bytes는 보존한다.

## R110: 현대 스크립트의 recursive 객체 그룹 Delete 연결

- 기준 clean/pushed `ebf476047b2eea1b55a018f58a885cc9edfc9e4f`에서 `src/script_ns_object.go:(nsObjGroup).Delete` 한 본체의 unconditional 미구현 panic만 제거한다. unchanged SCRIPT03 GroupDelete와 같은 recursive EachObject/객체 Delete를 사용하며, 고전 builtin 자체를 이번에 고쳤다고 주장하지 않는다. signature/imports/본체 밖·기존 테스트/입력/자산/Save/config는 불변이다.
- 최초 compiled RED81fail/0pass/81IDs·actual1 뒤 새 테스트를 동결했고 동일 GREEN81pass/0skip·actual0이다. 79leaf와 부모2개에서 C-owned >4GiB 전체 native records·중첩/중복/없는 extent·dead/destroyed/NoUpdate·비객체/빈/없는 그룹·native delayed-list 순서·unsigned ID/frame 경계 및 repeated Delete를 현대 API와 unchanged SCRIPT03 VM 양쪽에서 검사한다. 공급된 삭제 callback은 없지만 fixture의 active duration/player/team/inventory/monitored owner/Quest 분기는 새 증명이 아니라 기존 lifecycle 회귀 유지다. 새 원본 PE 실행/전체 world 동등성은 UNPROVEN이다.
- 일반·HD·실제cgocheck2 각각50,906pass/41skip/38packages, server50,142pass/7skip/35packages·actual0. R108의 기존 ID/action/multiplicity/skip/package outcomes를 모두 보존하고 새81pass만 추가했다. strict 및 cgocheck2+race+checkptr2 count3 각각3,707IDs/11,121pass·0skip/0fail/3packages·actual0이며 NPC Sword HP2000→1997→1994→1991도 유지한다. test audit SHA=`83364dc023ece28f027b03f7d89a728ada4752a5e2a524c4495455ba69c9f1df`, one-body audit SHA=`20431700f9c774a060d8bf274cf6e0f0f268d7424dcf445aa025f834f88c3f39`; functional/tested bytes 동일 SHA=`4968aeee35b4d2c779590ba02a58fa53290b5fc389ed56ea92042f9978d3802a`다.
- 증거 `/private/tmp/opennox-script-group-delete-r110.M5n0dy`. original2,968code/650data verifier actual0, protected303rows·개인/stock 설치 불변·삭제0이다. 아래 과거 문서 suffix bytes를 보존한다. 커밋 직후 push 후 clean ARM643제품과 기존 unchanged WoundedApprentice/object-death-spawns 일반·HD headless/mock를 별도 runtime ledger로 검사한다. GUI는 기존 lifecycle 회귀이며 현대 API 직접 GUI 증거가 아니다. R107 소환물 cheat-load/R108 Troll cloud 요청은 완료했으나 natural Skeleton timeout·Con02A frozen Screen actual2·passive-food·Quest stage3 및 full-port 목표는 미완료다.

## R108: 트롤 사후 독구름의 원본 lifetime 참조와 실제 중독 검증

- 기준 clean/pushed `e3e6f3dce72aac481e72b1c9c38de886460be40c`다. 기존 production은 `src/server/monster_dead_callback_549d80.go:monsterDeadTroll54A270` 한 본체만 변경한다. constructor가 반환한 cloud update-data를 CreateObjectAt/Audio/Balance 콜백 전에 보관하여 원본과 같은 cached 참조에 Duration을 저장한다. callback 이후 바뀐 참조를 다시 읽던 경계를 수정하며 nil/할당 실패·소유자/위치·수명 계산·signature/imports/본체 밖 bytes는 보존한다.
- unchanged PE32 `0054A270`154B/30entries 및 `00419A70`17B/29entries를 declared constructor/CreateAt/audio/binary32-balance 서비스와 bounded30조건으로 실행했다. supplied CW027f이며 executable bytes·guard sentinels 불변/memory fault0이다. 원본 stdout SHA=`64f78433d1b1e8b82e4f45bcc1ca8fa46d4fd87542cc75d7f961490269f3752f`. original type-cache/native constructor alias 경계, 실제 caller CW 및 원본 전체 GUI/world는 UNPROVEN이다. JIT signal4와 private API 오류 actual1도 보존한다.
- 최초 compiled RED actual1은49IDs 중42pass/7fail(6개 callback 참조 replace/nil과 부모)이며 이후 테스트/JSON/v1 입력을 고정했다. 동일 GREEN49pass·actual0. 새 C-owned >4GiB 객체의 전체 bytes·cached store·콜백 순서·nil guards·float32/nearest 수명 경계를 검사한다. 원본 raw transcript와 repo JSON은 의미상 동일하되 pretty formatting 때문에 byte-identical이라고 주장하지 않는다.
- GUI v1의 배치 공간 부족 actual2와 별개 v2의 첫 DOT 이전 partial HUD 가정 actual2를 그대로 남겼다. 별개 v3는 같은 assertion/HP/RNG/WAIT/입력/3600tick timeout을 유지하고 실제 DOT/client 반영 후에만 partial-meter pixels를 관찰한다. v3 initial compile-only actual1의 missing import만 실행 전에 추가했고 수정 전 source를 보존했다. 기존 실패 fixture/YAML·PNG/golden을 덮거나 결과를 retune하지 않았다.
- 같은 frozen headless/mock 일반·HD actual0은 정상 Troll 사망80→0/DYING, 자연 생성 SmallToxicCloud의 owner/position·450tick 수명, player150→149/Wolf40→39/NPC150→149의 첫 source-less POISON DOT와 client drawable/HP를 확인한다. contact 시 dose1만 적용되고 즉시 HP 피해는0이며 이는 Small cloud 원래 동작이다. 실제 녹색 cloud/HUD pixels1860, 독 면역 Troll HP80 유지, configured450/관찰age452의 world/client 소멸 및 보통 해독 cleanup을 확인했다. 실제 오디오 장치는 검사하지 않는다.
- 일반·HD·실제cgocheck2 전체 각각50,825pass/41skip/38packages, server50,061pass/7skip/35packages·actual0. R107의 모든 기존 ID/action/multiplicity/skip/package terminal을 보존하고 새49pass만 추가했다. strict 및 cgocheck2+race+checkptr2 count3 각각3,344IDs/10,032pass·0skip/0fail/3packages·actual0이며 NPC Sword HP2000→1997→1994→1991 회귀도 유지한다. test audit SHA=`3e6351eb7492dbe4ba53cf836547425037f6b388de19f07243fb615176a48b8c`; functional-ready/tested checkpoint bytes 동일 SHA=`f809971bc79a94f7223e96baab93c3374e371020b5ef293e08c78084fe127ea2`다.
- 증거 `/private/tmp/opennox-troll-cloud-r108.mxcNZ3`. original code2,968/data650 verifier actual0, protected 기존300rows+새 YAML3rows·개인/stock 설치/Save/config 불변·삭제0이다. 아래 과거 문서 suffix bytes를 보존한다. 커밋 직후 push 후 clean ARM64 일반/HD/서버3제품과 같은 frozen 일반/HD cloud GUI를 별도 runtime ledger로 검증한다. 소환물 cheat-load는 R107에서 완료했고 기존 natural Skeleton timeout·Con02A frozen Screen actual2·passive-food·Quest stage3·전체 포팅 목표는 남는다.

## R107: 캠페인 `load` 치트의 소환물 재배치 연결

- 기준 clean/pushed `03708475c3eda1774697812108276fc510347799`에서 기존 production은 `src/parsecmd_maps.go:noxLoadMap` 한 본체만 변경한다. 성공한 오프라인 Coop 로드에 정상 출구의 `sub_4DCBF0(1)` 재배치 플래그를 연결한다. 모드/파일 검증 실패, 온라인·Quest, team 생성, 강제 로드 및 기존 pending 상태는 보존한다. signature/imports/본체 밖 bytes·기존 테스트/자산/Save/config는 불변이다.
- 원본 PE32의 setter/getter/busy flag 및 console/exit call sites를 읽기 전용으로 확인했다. 원본 콘솔에도 재배치 flag가 없으므로 원본 콘솔 동등성 복원이 아니라 사용자 요청 캠페인 동작 개선이다. 전체 원본 GUI/world는 UNPROVEN이다.
- 최초 compiled RED actual1은 25IDs 중18pass/7fail이다. 새 테스트·서비스 fixture·YAML 및 private 입력은 이후 고정했다. 동일 입력 GREEN25pass·actual0. 잘못된 private 호출 경로의 actual127도 보존하며 통과로 세지 않는다. 실제 F1/text/Enter의 `load con02a` GUI RED actual2는 살아 있는 Wolf가 소유자를 유지한 채 거리1210.474에 남은 실패다.
- 같은 frozen GUI의 일반·HD GREEN actual0은 실제 Summon Wolf/Summon Urchin/Charm 완료 및 Pixie Swarm의5개 객체가 전환 뒤에도 같은 owner/wire/HP·deadline을 유지함을 확인한다. 30/270 destination ticks에 각각 unit 거리≤80, Pixie≤150, native >4GiB 및 실제 client drawable을 검사하고 플레이어 HP20/20을 유지했다. 새 private PNG만 생성하며 golden/override/HP/RNG/WAIT/timeout을 조정하지 않는다.
- 일반·HD·실제cgocheck2 전체 각각50,776pass/41skip/38packages, server50,012pass/7skip/35packages·actual0. strict 및 cgocheck2+race+checkptr2 count3 각각1,446IDs/4,338pass·fail0/skip0/3packages·actual0. 이전 ID/action/multiplicity/skip/package terminal을 모두 보존했다. test audit SHA=`434b9147f6acba06c94453845c2a1cb9fac0c5c6a4ffc7f685fe5e326d584370`, functional/tested checkpoint bytes 동일 SHA=`3d67444870be59c20e52f59448775053b5f437efd2b5dcf371813c2ae9b37579`다. 원본 code-verify2,968code/650data도 actual0이다.
- 증거 `/private/tmp/opennox-cheat-load-pets-r107.rCN17a`. protected 기존299rows+새 YAML1row, 개인/stock 불변·삭제0. 커밋 직후 push한 뒤 clean ARM643제품과 같은 frozen headless/mock normal/HD cheat-load를 별도 runtime ledger로 검증한다. 트롤 사후 독구름 검토와 기존 natural Skeleton timeout·Con02A frozen Screen actual2·passive-food·Quest stage3·전체 포팅 목표는 남는다. 아래 과거 문서 suffix bytes는 보존한다.

## R106: 단축 스펠을 창 밖으로 드롭할 때 원본 슬롯 삭제 복원

- 기준은 clean/pushed `ca4ffbde4b1f997058290b0e1fbd1c35a4cf31c6`다. 기존 production은 `src/legacy/GAME2.c:nox_xxx_quickBarWnd_45EF50` **한 본체만** 변경한다. signature/imports/본체 밖 bytes, 기존 테스트·YAML·PNG·원본 manifest·게임 자산·개인 Save/config는 불변이다. 새 C-owned 회귀 테스트/원본 JSON 및 별개의 새 GUI 입력 두 개만 추가한다.
- drop 반환 sentinel137의 누락된 원본 분기를 복원했다. 원래 inclusive reserved rectangle 밖이며 캡처 행이 nil 또는 현재 행일 때 spell DWORD와 self-cast BYTE만 0으로 지운다. reserved 영역·드래그 중 행 변경·다른 슬롯 이동/교환·Trap 복사 취소·capture cleanup 및 padding/나머지 항목은 보존한다. Trap의 non137 결과도 저장하되 원본처럼 main-bar repaint만 한다. 피해/HP·RNG·WAIT·golden은 조정하지 않는다.
- unchanged PE32 원본 `0045EF50`(904B/486 entries, SHA=`efc70e90c3da32286f5406549a87beff49444ca45557cf569bc7cbf68ac402e2`) 및 `004281F0`(48B/204 entries, SHA=`5d0f3a029bde7d756ce503738f8a04cebe2cf2188d0a0c16e13a1a6f3a469bfa`)의 bounded245조건을 봉인했다. window/drop/cast/repaint는 명시적 서비스이며 원본 전체 GUI/world는 UNPROVEN이다. transcript와 새 JSON은 byte-identical, SHA=`2ab9d21d12fa7c8ddf4ad3f87742b25b498b10f9e86554535c0d26c2baca1c73`다.
- 최초 private native probe actual1은 C fixture 선언 컴파일 오류로 native 실행 전이었다. 새 C fixture 선언만 교정한 뒤 최초 **compiled RED actual1: 62pass/184fail/246IDs**를 얻었고 이후 테스트/fixture를 고정했다. 동일 입력 GREEN246pass/0skip/actual0. 64-bit C-owned >4GiB record/window의 전체201B×2 payload와 metadata/padding·service order를 대조한다.
- 새 GUI v1의 hosted wizard 기본 슬롯이4인 상태에서 빈 슬롯0을 가정한 actual2(`got4 want0`)는 diagnostic 실패로 보존하며 통과했다고 표시하지 않는다. 기대값/입력을 수정하지 않고 별개 explicit occupied fixture인 v2를 추가·동결했다. 같은 v2 mouse input의 genuine RED actual2(`got27 want0`)→GREEN actual0이며 삭제·reserved 취소·이동·교환을 검사한다. live Screen에서 Fireball 및 다른 스펠 제거가854/855pixels이고 두 빈 슬롯 ROI가 동일하다. PNG/golden/override를 쓰지 않는 read-only 비교다.
- 일반·HD·실제cgocheck2 전체 각각 **50,751pass/41skip/38packages**, server49,987pass/7skip/35packages; strict 및 race/checkptr2는 각각 count3·1,305IDs/3,915pass·0skip/0fail·actual0이다. 기존 전체 ID/action/multiplicity/skip/package terminals를 유지했다. test audit SHA=`4591da155246e67e5af77abb836542c272fe6f5e77f4954cdb1199a2843edf91`. original code-verify2,968code/650data actual0 및 NPC Sword/분수 면역 회귀도 보존했다.
- functional/tested checkpoint bytes 동일 SHA=`8aaf584802fa43e4822e0230504313da406b5b4699001869af3e33b0fbc8503d`. protected 기존297rows 불변+새 YAML2rows, 개인/stock 설치 불변, 삭제0이다. 증거 `/private/tmp/opennox-shortcut-drag-r106.SpGqvx`. 커밋 직후 push한 뒤 clean ARM643제품과 headless/mock normal/HD의 새 v2 및 기존 unchanged Trap Set live-icon 회귀를 별도 runtime ledger로 검증한다. 새 소환물 cheat-load/트롤 사망 독구름 검토, 기존 natural Skeleton timeout·Con02A frozen Screen actual2·passive-food·Quest stage3 및 전체 포팅 목표는 남는다.

## R105: 플레이어 공격의 바깥 원형 탐색 정밀도 보존

- 기준은 clean/pushed `9e9ad4e38d798f5260d80359cb19255a47368bc3`다. 기존 `nox_xxx_playerAttackScanCircle_538960` **한 본체만** native Go bridge로 연결한다. signature/imports/본체 밖 및 기존 테스트·YAML·PNG·원본 자산은 변경하지 않는다. 새 bridge는 C stack center/userdata와 >4 GiB callback/candidate 포인터를 보존한다.
- 원본 PE32 `0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`의 변경 없는 `00518040`(108B, 56 entries)·`005180B0`(181B, 45 entries)을 56개 bounded 사례로 실행했다. rectangle enumeration55/box6/visit26은 명시적 services다. 공급 CW=`0e7f`이며 실제 caller CW, 원본 map/box/damage tail 및 wholeworld 동등성은 **UNPROVEN**이다. 원본 transcript/새 fixture SHA=`c940e2a9fbf6d69b4a5b58962527cf338c5ce8af16fa4225355f071cd1795710`.
- 기존 sqrtf 이른 binary32 store는 retained-positive circle 경계를 0으로 잘라 공격 대상을 누락했다. X FSTP/Y retained×storedY, FSQRT retained, strict positive·unordered 배제와 binary32 rect stores를 복원한다. live center mutation, point/circle/box, nil/negative/zero/NaN 및 callback 순서를 유지한다. HP·공격 프레임·피해량/유형·RNG/WAIT를 조정하지 않는다.
- 최초 compiled RED는 71 ID 중 69pass/2fail(해당 +0 boundary 및 부모), 실제 exit1이다. 이후 frozen 테스트/fixture를 수정하지 않았으며 같은 71 ID GREEN actual0이다. C-owned >4 GiB 실제 C attack/frame/index/ray/registered receiver 경로의 9 경계 사례를 각각 3주기 검사했다. 해당 경계는 주기당 한 번, raw36/BLADE와 RNG index 불변이다. 4-frame animation/receiver는 fixture service이며 원본 Windows HP/death 인증은 아니다.
- normal/HD/strict 각각 50,505pass/41skip/38packages, server 49,741pass/7skip/35packages, strict count3 및 race/checkptr count3 각각 1,050 IDs/3,150pass actual0이다. 기존 ID/action/multiplicity/skip/package terminals를 전부 보존했다. 원본 code-verify도 actual0(2,968 code/650 data)다. 검사 audit SHA=`ec254f1bc8333f7a9ab50c127977e8e0a1c2680dd74afd0f09213e5d1d90b507`.
- NPC Sword raw3/type7 실제 HP2000→1997→1994→1991, fountain 비면역 HP60→57→54→51 및 데몬 화염 면역 HP60 유지 회귀를 보존했다. 커밋 직후 push하고 clean ARM64 3제품·기존 unchanged headless NPC combat normal/HD 시나리오를 별도 runtime ledger로 검증한다. 과거 natural Skeleton timeout, Con02A frozen Screen actual2, passive-food·Quest stage3 및 전체 포팅 목표는 완료로 표시하지 않는다.

## R104: 용암 분수 FLAME 피해 경계와 원래 화염 면역 보존

- 기준은 clean/pushed `e907e6ef2b2a61e0de3cf938651dfabeaf97ccf5`다. `LavaFountain3`의 class4259840/`0x410000`은 **DANGEROUS|IMMOBILE**이며 FIRE|SIMPLE이 아니다. 보고된 raw3/type1은 FLAME이고 LAVA/type12로 바꾸지 않는다. 공통 `playerDamageWorldFlameShape4E17B0` 본체와 그 뒤 중복 FIRE|SIMPLE|DANGEROUS 검사를 하던 `PlayerDamageNative4E17B0` 본체를 **각각 별도의 한-본체 기능 커밋**으로 수정한다. 첫 shape-only 경계는 DefaultDamage만 복원하며 Player/NPC 경로 완료로 기록하지 않는다. signature/imports/본체 밖·피해량/유형·기존 테스트/입력/자산은 불변이다.
- `MeleeDemon`1338과 `EmberDemon`1337의 subclass332802/`0x51402`에는 FireImmune/`0x400`이 있다. 원본은 이 FLAME을 처리하되 HP를 깎지 않는다. unsupported warning을 제거하고 live 면역 조회·초기 OnFire·marker reset·nil Health 이전의 면역 반환을 보존한다. 이름/ID를 production에서 하드코딩하거나 데몬의 면역을 제거하지 않는다. 비면역 monster/NPC/player는 원래 world-fire 피해 tail을 사용한다.
- unchanged GAME.EXE의 DefaultDamage 및 원본 qualifier/owner/effect 루프를 제한 실행한144조건(두 데몬·8개 gate·raw -3/0/3·보호0/0.5/1)을 새 JSON에 봉인했다. JSON은 original-v3 stdout과 byte-identical, SHA-256 `ec628f2225943358f6614056b4cbed4eda4f848d2627199ae7f45c9cac9d9444`다. allegiance/FireProtection/audio/buff/shield/HP-argument·field-guide는 명시적 공급 서비스이고 CW027F도 공급 값이다. 원본 Windows HP/death·actual caller CW·전체 world는 미증명이다. 최초 두 private probe actual1과 그 raw 기록도 보존한다.
- 새 Go 테스트 두 파일과 JSON은 최초 **compiled red actual1·72pass/129 test fail·201 identity** 뒤 고정했다. shape-only 실행은197pass/4 test fail·actual1로, 중복 PlayerDamage guard가 막던 player/NPC3조건 및 root 실패를 그대로 보존한다. 같은 frozen 입력의 두-본체 green은201pass·fail0/skip0·2package·actual0이다. 실제 C-owned >4GiB source==weapon/type1202/class4259840/flags16777748에서 C→registered DefaultDamage/PlayerDamage→UnitSetHP를 검사한다. 11개 native 조건의 immune HP **`60→60→60→60`**, nonimmune HP **`60→57→54→51`**, attribution/인접 Health word/carry/absorption·원본 weapon bytes를 확인했다.
- Go1.26.5/GOMAXPROCS2/`-p=2` 직렬 일반·HD·실제cgocheck2 전체는 각각 **50,434pass/41skip·38package**, server49,670pass/7skip·35package·actual0이다. R103의 모든 기존 ID/outcome/횟수/skip/package terminal을 유지하고 새201pass만 추가했다. strict 표적 및 cgocheck2+race+checkptr2는 각각 count3·875 identity/**2,625pass**·fail0/skip0·2package·actual0이다. 테스트 감사 actual0 SHA-256은 `9adc7391eb1b02dc93d6bc0f9482a5b491947910d3f6793726803eb8aaa62123`다. 기존 NPC Sword raw3/type7(EXPLOSION)→Skeleton의 실제 C HP `2000→1997→1994→1991`도 두 lane에서 각각3회 유지했다.
- 원본2,968 code/650 data ranges와 protected/personal/stock 전체 rows를 유지했다. functional-both/tested checkpoint bytes는 동일하다. 증거는 `/private/tmp/opennox-lava-fountain-r104.LXHiug`다. 두 기능 커밋을 각각 즉시 push한 뒤 clean ARM64 일반/HD/서버 제품 및 unchanged 환경 피해 YAML의 headless/mock 회귀를 별도 private handoff로 검증한다. 해당 YAML은 Flame·6종 spike·Lava만 지원하므로 이를 LavaFountain GUI 인증으로 확대하지 않는다. R103의 재생성 가능 GOCACHE 정리 외에 source/게임 데이터/Save/config/증거/제품을 삭제하지 않았으며 아래 기존 문서 suffix bytes를 보존한다. 기존 Skeleton timeout·Con02A Screen actual2·외곽 circle 정밀도·passive-food·Quest stage3·전체 포팅은 미완료다.

## R103: 플레이어 근접 대상 선택의 retained 정밀도 복원

- 기준은 clean/pushed `cb8e09c1afc6d0fbd995cb39e46233fcbce74bc7`다. 기존 `nox_xxx_playerAttackSelectCandidate_538960` 본체 하나만 native-width Go 경계로 연결했다. 원본 `005386A0`의 precision53/ToZero retained Y·제곱/합/제곱근·dot·edge, X/Y와 최종 거리의 지정된 binary32 저장, box의 binary64 반환, strict ordered cone와 unordered edge 판정·첫 monster 유지/monster 우선 선택을 복원했다. allegiance를 prop에도 호출하며 콜백 뒤 live class/flags/위치를 읽는다. signature/imports/본체 밖·피해량/유형·기존 테스트/입력/자산은 불변이다.
- unchanged GAME.EXE의 unhooked candidate 413bytes를 제한 실행한58조건/65진입을 봉인했다. 원본 방향512word와 서비스 호출 순서를 비교한다. allegiance63/interaction60/box3는 공급 서비스 경계이며 CW0E7F도 공급 값이다. 실제 original caller CW·원본 전체 world/전투는 미증명이다. 독립 math/big의256방향/768점과 실제 C-owned >4GiB indexed 공격의3개 frame cycle도 검사한다.
- 최초 compiled red actual1은 package-only 방향 표 초기화 누락으로62pass·5 test fail이다. 기존 새 test/JSON3파일은 그대로 고정하고 별도 test-only 원본 방향 초기화 파일을 추가했다. 해당 파일도 다음 compiled red 뒤 고정했다. 초기화된 genuine red actual1은64pass·3 test fail로 cone miss와 retained edge 뒤 잘못된 대상 선택을 재현했다. 같은 frozen 입력 green은67pass·actual0이다. 기존 test ID/outcome/횟수/skip/package terminal을 모두 보존했다.
- Go1.26.5/GOMAXPROCS2/`-p=2` 직렬 일반·HD·실제cgocheck2 전체는 각각 **50,233pass/41skip·38package**, server49,469pass/7skip·35package·actual0이다. strict 표적 및 cgocheck2+race+checkptr2는 각각 count3·263 identity/789pass·fail0/skip0·2package·actual0이다. test 감사 v2 actual0 SHA-256은 `f13f393b67c9ac9572d6dec6ab764542e5375ae8ade1f6c8cca0c12f2937cf83`다.
- 첫 strict 전체와 별도 serial server preflight는 컴파일 archive/링커의 ENOSPC·errno28로 actual1이었다. 두 실패 로그를 삭제하거나 PASS로 바꾸지 않았다. 사용자의 캐시 정리·재개 지시에 따라 검증한 전용 GOCACHE에 Go1.26.5 `clean -cache`만 실행해54,920,852KiB를 비웠다. module/SDK/GOPATH·소스·게임 데이터·개인 Save/config·로그·기존 제품은 삭제하지 않았다. 전후 source/protected/personal checkpoint bytes는 동일하며 캐시는 빌드로 재생성된다. 캐시 감사 actual0은 `6ce0c4c1228e7a5182fc0a57cd9384a39ef72ab58b10ed32c508ac8ae9e6c180`다. 공간 확보 후 strict 재실행 actual0과 원래 actual1은 별도 기록이다.
- 보고된 NPC Sword raw3/type7(EXPLOSION)→Skeleton은 R79 `15ae300d3bfa8af59932c01e236e55c2bbbe53fc`의 복원 상태다. 두 strict lane의 실제 C→SkeletonDamage→DefaultDamage→UnitSetHP가 각각3회 **`2000→1997→1994→1991`**임을 확인했다. type7을 Blade로 재해석하거나 HP/team·입력·RNG·WAIT·golden을 바꾸지 않았다. HP2000/team2/1 통제 fixture는 자연 전투 인증과 구분한다.
- 원본2,968 code/650 data range 및 protected/personal/stock 전체 rows를 유지했다. functional/tested-final checkpoint bytes는 동일하다. 증거는 `/private/tmp/opennox-player-target-r103.zmYEUF`다. 커밋·즉시 push 뒤 clean ARM64 일반/HD/서버 제품과 headless/mock 실행은 별도 private handoff로 검증한다. 외곽 circle enumeration의 정밀도·기존 Skeleton timeout·Con02A Screen actual2·passive-food·Quest stage3·전체 포팅은 미완료다. 아래 과거 문서 suffix bytes 및 모든 기존 실패 증거/제품을 보존한다.

## R102: 근접 대상 선택의 원본 정밀도와 live 콜백 복원

- 기준은 clean/pushed `c8c8b3e80a3c836520200b02ca39acb5908160be`다. 기존 `monsterPickMeleeTarget549440` 본체 하나만 수정한다. 원본의 precision53/ToZero retained radius·X delta·제곱/합/제곱근·dot·edge와 지정된 binary32 저장을 복원한다. friendly 허용 때도 allegiance를 호출한 뒤 live 위치를 읽으며, cone의 ordered strict >와 edge의 C0-only unordered 수용을 구분한다. signature/imports/본체 밖·피해량/유형·기존 테스트/입력/자산은 불변이다.
- unchanged GAME.EXE에서 picker/candidate/default-strike 및 두 조건의 실제 rect/round walker를 제한 실행한 23개 결과를 봉인했다. 첫 private trampoline은 번역 캐시 때문에 다중 후보를 잘못 순회해 감사 actual1이며, 원본 script/log를 보존하고 별도 v2에서 해당 emulator cache만 무효화했다. 원본 code bytes를 바꾸지 않았다. 원본 world의 allegiance/trace/damage는 공급 경계이며 기록된 전투 좌표의 radius12와 CW0E7F도 공급 값이다. 전체 원본 전투 인증은 아니다.
- 새 test/JSON 두 파일을 최초 compiled red 뒤 고정했다. red actual1은 21pass·6 test fail+1 package fail로, 네 원본 조건(cone-+3/friendly live-position/unordered radius/retained search radius)과 독립 256방향 비교에서 차이를 재현했다. 같은 입력 green은 기존 control 포함27pass·actual0이다. 방향 표256쌍과 독립 math/big의768점을 검사하며 두 indexed 조건은 실제 native map 순회와 strike 전달까지 비교한다.
- 일반·HD·실제cgocheck2 전체는 각각 **50,166pass/41skip·38package**, server49,402pass/7skip·35package·actual0이다. R98의 모든 기존 ID/횟수/skip/package terminal을 보존하고 새26개만 추가했다. strict 및 cgocheck2+race+checkptr2는 각각 count3·195 identity/585pass·fail0/skip0·2package·actual0이다. test 감사 actual0 SHA-256 `4de0990c3fb35838ad0093721e71b26c143c7c44a8e997c680179d9824da66b2`다.
- 보고된 NPC1343/subclass69778·Sword1174/subclass256→Skeleton1314/subclass70210의 raw3/type7(EXPLOSION)은 R79 `15ae300d3bfa8af59932c01e236e55c2bbbe53fc` 복원 상태다. 두 strict lane에서 실제 C→SkeletonDamage→DefaultDamage→UnitSetHP를 각각3회 실행해 **`2000→1997→1994→1991`**을 확인했다. HP2000/team2/1은 기존 통제 fixture이며 type7을 Blade로 재해석하지 않았다.
- 원본2,968 code/650 data range 및 protected/personal/stock 전체 rows를 유지했다. functional/tested checkpoint bytes는 동일하다. 증거는 `/private/tmp/opennox-melee-target-r102.OSTz4Q`다. 커밋·즉시 push 뒤 clean ARM64 일반/HD/서버3제품과 headless/mock 후속 회귀를 별도로 검증한다. 이번 경계 수정만으로 기존 Skeleton timeout·Con02A Screen actual2·passive-food·Quest stage3나 전체 포팅 완료를 주장하지 않는다. 기존 증거/캐시/제품/개인 Save/config와 아래 suffix bytes를 보존한다.

## R98: NPC 무기 교체 콜백 뒤 live subclass 재조회 복원

- 기준은 clean/pushed `021f8be94ab135a88d2c19daf747a31c45726b4d`다. 기존 `npcWeaponEquipNative53A2C0` 본체 하나에서 화살통 해제 콜백 뒤와 장비 표시 콜백 뒤의 subclass를 다시 읽도록 복원했다. 원본 `0053A330`·`0053A390`의 live load와 같으며 signature/imports/본체 밖·피해 유형·기존 테스트/입력/자산은 불변이다. 새 테스트와 봉인한 원본 결과 파일만 추가했다.
- 변경하지 않은 GAME.EXE의 실제 공개 equip 진입점과 NPC equip/dequip·화살통/방패 순회를 제한 실행한60조건 중21조건이 표시 서비스 경계에서 arriving item을 변경한다. 실제 C 소유 >4GiB 객체의 public C equip 경로에서 **compiled red actual1·8조건 실패**, 같은 frozen 입력의 green61 identity/pass·actual0을 확인했다. 전체 owner/update/item/modifier bytes·콜백 순서·equipment/호환 word/효과 보존을 비교한다. PE32 포인터 word는 기존 native64 inventory/bit 계약대로0이며 포인터를32비트로 자르지 않는다. 제한 서비스 경계 비교이지 전체 원본 world나 자연 stock 콜백 변이 인증은 아니다.
- 일반·HD·실제cgocheck2 전체는 각각 **50,140pass/41skip·38package**, server49,376pass/7skip·35package·actual0이다. 이전 모든 ID/횟수/skip·패키지 결과를 유지하고 새61개만 추가했다. strict 표적 및 cgocheck2+race+checkptr2는 count3·143 identity/429pass·fail0/skip0·2package·actual0이다. 결과 집계기의 최초 actual1은 count3을 패키지 수도3배로 센 private 오류였으며 원본 집계기·로그를 보존하고 별도 v2만 교정했다.
- 보고된 NPC1343/subclass69778·Sword1174/subclass256→Skeleton1314/subclass70210의 raw3/type7(EXPLOSION)은 기존 R79 `15ae300d3bfa8af59932c01e236e55c2bbbe53fc` 복원 상태다. 두 strict lane에서 실제 C→SkeletonDamage→DefaultDamage→UnitSetHP가 각각3회 **`2000→1997→1994→1991`**임을 확인했다. type7을 Blade로 재해석하거나 HP/team·fixture를 바꾸지 않았다. 통제 C fixture와 자연 전체 전투는 구분한다.
- 원본2,968 code/650 data ranges·protected/personal/stock 전체 rows를 확인했다. 새 테스트/JSON은 최초 compiled red 뒤 고정했다. 원본 실행의 python/인수 오류 actual1 및 sandbox JIT SIGILL4는 그대로 보존하며 같은 script의 승인 실행 actual0과 구분한다. 증거는 `/private/tmp/opennox-npc-equip-r98.4TFZ9r`; test 감사 v2 actual0 SHA-256 `7e28d44075f1af27cab9e77fa6b9b57b553acb25bb1401cf88e8b8486ce73820`다. 커밋·즉시 push 뒤 clean3제품 및 headless 회귀는 별도 handoff로 검증한다.
- 앞선 R96/R97의 동일 frozen Con02A 인과 비교는 NPC equipment replay53e2eef96과 frame BYTE684946994의 변화만 확인했다. completion clear6a992ebc·thrown frame69c3ca40은 각각0픽셀 차이다. 원본과 일치하는 equipment/frame 복원을 되돌리거나 golden을 갱신하지 않았으며 이번 콜백 수정이 Con02A를 해결했다고 주장하지 않는다. 기존 Screen actual2·Skeleton timeout·passive-food·Quest stage3·전체 포팅 목표는 미완료다. 캐시/증거/개인 Save/config 및 아래 기존 suffix bytes를 보존한다.

## R95: Con02A 무기 HUD 인과 비교와 미니맵 설명 정정

- 기준은 clean/pushed R94 `6152217b92bcfc5acb305cc7c4afe2decdd4489a`다. 아래 R94의 “minimap169”는 잘못된 위치 식별이다. **169픽셀은 좌측 하단 현재 무기 HUD의 지팡이 아이콘**이고, 실제 minimap472600은 이 해상도에서 y299..469에 그려진다. 기존 문서·이미지·raw 로그는 보존하고 이 추가 절로 정정한다.
- 새 clean clone에서 `0df9319a8`와 직후 `1992e0306`를 같은 frozen YAML/golden으로 headless/mock 실행했다. 두 source의 유일한 차이는 기존 current-weapon C drawer의 packed52바이트 viewport 인수를 native104바이트 view로 바꾸는 한 줄이다. 직전 actual0/Screen PASS, 직후 actual2/Screen FAIL이며 차이는 정확히169픽셀/1성분 `[31,703)..[66,739)`뿐이다. 원본의 parent+51/+81 계약 및 기존 실제 C raster와 일치하는 아이콘 복구이므로 되돌리거나 golden을 바꾸지 않았다. 직후→R94 material의417픽셀은 player 영역만이며, 현재 R93의 추가 NPC 위치·조명 차이는 UNPROVEN이다.
- 두 실행의 소환1039·Clyde Sword prefix1077·release1257과 player HP20/20·drawables1695는 같다. dormant 대상 HP0/0이므로 실제 양수 피해 증거는 아니고 직후 quit도 미도달이다. 두 정규 `.git` 디렉터리 clone 제품은 공식 metadata actual0·Go1.26.5/ARM64·full revision·clean이며, R94 worktree metadata actual1과 원시 실패는 그대로 보존했다. SDK/verifier/provenance 변경은 없다.
- 보고된 NPC Sword raw3/type7→Skeleton은 기존 R79 `15ae300d3` 복원 상태를 다시 확인했다. unchanged viewport·C 피해 표적25 identity는 일반·HD·실제cgocheck2·strict+race/checkptr2 각각 count3/75pass·fail0/skip0·actual0이다. 정확한 class/subclass/flags를 유지한 실제 C→SkeletonDamage→DefaultDamage→UnitSetHP는 **`2000→1997→1994→1991`**이며 production·기존 테스트/입력·자산 변경은0이다. 통제 fixture와 전체 suite/자연 전투는 구분한다.
- unchanged oracle2,968 code/650 data range 및 source/protected/personal/stock 전체 rows 불변을 확인했다. 감사 `/private/tmp/opennox-con02-minimap-r95.bWuUiR/audit-runtime.log` actual0 SHA-256 `67a9f75ca1462d0e11f8c3b7827cb6dc1ff0e82c472d9b0a68961dbfe7808bb2`는 Screen actual2를 PASS로 바꾸지 않는다. 상세 입력/로그/PNG/제품 해시는 인벤토리 R95 절에 기록했다. 기존 증거·캐시·제품·개인 Save/config와 아래 suffix bytes를 보존한다. 현재 Con02A 위치/조명·Skeleton timeout·passive-food·Quest stage3·전체 목표는 미완료이며 docs-only 최신3제품은 별도 handoff다.

## R94: Con02A 과거 화면 비교와 worktree provenance 한계

- 기준은 clean/pushed R93 `47348fcbcbd0f1b0adae3e18f342d5e686d42abd`다. production/test/fixture를 바꾸지 않고 같은 frozen Con02A 입력과 golden으로 과거3개 clean source를 headless/mock 실행했다. c76594677·ed39bef3a는 Screen/cleanup actual0, 3c7abc0fa는 Screen actual2다. 세 실행의 소환1039·Clyde Sword prefix1077·release1257은 같으나 대상 HP0/0이어서 양수 피해 증거는 아니다.
- 과거 material 화면은 기준과586 pixel(player417/minimap169), 현재 R93 화면은 기준과78,693 pixel 차이다. clip snapshot은 통과했으므로 clip 수정 탓으로 단정하지 않는다. minimap 및 현재 NPC 위치/조명의 원인은 UNPROVEN이며 golden/WAIT/HP/RNG/입력으로 실패를 통과시키지 않았다.
- 과거 worktree 제품3개는 **공식 metadata actual1·vcs.revision 누락**으로 clean release 인증이 아니다. 고정 Go1.26.5의 `.git` 디렉터리 검사 때문에 worktree 파일을 인식하지 못하며 실제 `-buildvcs=true` probe도 누락을 재현했다. 일반 주 저장소 probe는 full R93 revision·clean을 기록했다. verifier/SDK/provenance를 변조하지 않고 실패를 보존했다.
- 현재 palette/clip/secret-wall/NPC Sword 표적은 일반·HD·strict·strict+race/checkptr2 각각 count3·214 identity/642 pass·fail0/skip0·actual0이다. 기존 R79의 보고된 NPC Sword raw3/type7→Skeleton 실제 C route는 `2000→1997→1994→1991`이며 damage 본체·피해 유형은 재변경하지 않았다. 통제 fixture이며 전체 suite/original world 인증으로 확대하지 않는다.
- source/protected/personal/stock 전체 rows는 wrapper 전후 불변이다. 감사 `/private/tmp/opennox-con02-visual-r94.KZ4w1r/audit-runtime.log` actual0 SHA-256 `e60cd7c094f77031638a9bee5eb380614a7884bd3bb57ef4a4a740e74fa21c97`는 위 actual1/actual2를 PASS로 바꾸지 않는다. 상세 해시는 인벤토리 R94 절에 기록하며 기존 문서 suffix·실패 로그/PNG·캐시·개인 Save/config를 보존한다. 화면 차이·Skeleton timeout·passive-food·Quest stage3·전체 목표는 미완료다. 최신 주 저장소 clean3제품은 별도 handoff다.

## R93: 실제 캠페인 피해·소환 확인과 frozen Screen 실패 보존

- 기준은 clean/pushed `2730ef8ed514fa22f60dd2f8e9fd8445e88287b6`이며 이번 production·기존 테스트/시나리오/자산 변경은0이다. Go1.26.5 ARM64 일반/HD의 기존 Wiz01A 트리거·Horvath 번개는 Urchin12 생성·6 직접 명중·12 처치, Con01A 곰 낙석은 HP200→0·client -200·player release를 확인해 각각 actual0이다. 기존 player 위치/polygon 통제 helper를 사용하므로 전체 자연 캠페인 인증은 아니다.
- Con02A도 소환1035·alive/client-visible Spider12/12·Clyde의 실제 Sword prefix1071·NPC 제거/player/dialog/cinematic release1250까지 도달했으나, 두 frozen 시나리오는 마지막 Screen에서 **actual2**다. byte-identical private YAML과 기존 golden을 그대로 유지했고 두 got/diff 화면은 일반/HD 간 byte-identical이다. NPC 위치/조명·미니맵 차이 원인은 **UNPROVEN**이며 기대값·대기·HP·RNG를 바꾸지 않았다. 대상 dormant NPC의 HP0/0이므로 이 검 명중을 양수 체력 감소로 세지 않는다. quit에는 미도달했다.
- 보고된 NPC Sword raw3/type7(EXPLOSION)→Skeleton은 기존 R79 native 복원 상태를 일반/HD/cgocheck2 및 cgocheck2+race+checkptr2로 각각 count3·19 identity/57 pass·fail0/skip0·actual0으로 재확인했다. 실제 C→SkeletonDamage→DefaultDamage→UnitSetHP는 **`2000→1997→1994→1991`**이며 type7을 변경하지 않았다. team2/1·HP2000은 기존 직접 C fixture이고 full original battle은 UNPROVEN이다.
- GUI6 및 기존 최신 release3제품의9개 metadata가 clean2730 revision·Go1.26.5/darwin/arm64/Mach-O ARM64다. headless/mock·audio-handles=false·trace=false로 실행했고 wrapper 전후 source/protected/personal/stock 전체 rows/JSON bytes가 불변이다. 통합 감사 `/private/tmp/opennox-campaign-r93.5varBt/audit-campaign.log`는 actual0 SHA-256 `9fe06527b7725234ce8d586c9e48a3ee7f77c4e7c7326a16c6a177c55078a14d`다. 두 Con02A actual2·got/diff와 기존 Skeleton actual2를 보존하며 passive-food·Quest stage3·전체 포팅도 미완료다. 상세 입력/로그/이미지 해시는 인벤토리 R93 절에 기록하고 아래 과거 suffix bytes를 유지한다.

## R92: 객체 그룹 Wander 복원과 NPC Sword 피해 유지 검증

- 기능 `5e0e66f47feb04b8979e9abf6b95e3dd036dd585`는 기존 `nsObjGroup.Wander` 본체 하나만 변경하고 새 테스트 파일만 추가했다. frozen classic GroupWander와 같은 recursive 멤버별 `obj.Wander()` dispatch를 복원하며 signature/imports/본체 밖·기존 테스트/시나리오/자산은 불변이다. 커밋 바로 다음 도구로 push actual0을 확인했다.
- 새 C-owned 고주소 matrix는7 group×4 flag×3 initial action의84 leaf와 root/malformed lookup을 합친 **86 identity**다. nested/missing/duplicate/nonobject 및 eligible/dead/destroyed/player 멤버, 전체 record bytes, recycled native waypoint slot과 roam flag 보존을 실제 public handle/API·SCRIPT03 classic VM에서 비교한다. real C group entry·중복 callback 순서/횟수·전체 원본 world는 **UNPROVEN**이다.
- 최초 missing first-child fixture fault 두 actual1과 v2 compile overflow actual1은 compiled target red로 세지 않고 보존했다. 새 fixture를 분리해 기존 malformed lookup fault도 검사했으며, compiled v3 red **actual1·1 pass·85 test fail+1 package fail** 뒤 같은 frozen 테스트로 green86 pass·actual0을 확인했다. 일반·HD·실제cgocheck2 전체50,079 pass/41skip·38package, server49,315 pass/7skip·35package·actual0이다. 기존 ID/횟수/skip를 유지하고 새86 pass만 추가했다. strict 및 cgocheck2+race+checkptr2는 count3·183 identity/549 pass·fail0/skip0·3package다.
- 보고된 NPC Sword raw3/type7→Skeleton은 기존 R79 복원을 두 strict lane에서 실제 C→SkeletonDamage→DefaultDamage→UnitSetHP로 다시 확인해 **`2000→1997→1994→1991`**이다. 피해 유형이나 production을 다시 바꾸지 않았다. clean ARM64 제품3개/help·GUI 제품4개까지7개 metadata를 검증했다. 일반/HD 단일 객체 Move/Wander4조건씩은64tick·native 슬롯 clear·HP 불변·actual0이며 새 group API의 live world 검증으로 확대하지 않는다. byte-identical YAML의 private Screen PNG8개는 fixed golden이 아닌 관찰 자료다. 기존 거미 전투도 frame764·거미 HP0/12·플레이어 HP14/20·actual0이고 전용 서버3맵 전환·4회 loop/tick 증가·quit exit0이다.
- headless/mock·audio-handles=false·trace=false로 실행하고 모든 wrapper 뒤 source/protected/personal/stock 전체 rows 불변을 확인했다. 감사 `/private/tmp/opennox-group-wander-r92.U1JyAF/audit-runtime.log`는 actual0 SHA-256 `eacb30ecf6dee0035d7e5b64f6f415e046203c8e5c1ddfc90181bede58622ec2`다. 별도 Skeleton 처치 actual2·2400tick timeout은 보존·미해결이며 이번에 재실행하지 않았다. HP/seed/RNG/WAIT/공격성/입력/timeout/기대값·캐시·기존 자료를 바꾸거나 삭제하지 않는다. passive-food·고정 Quest stage3 맵 불일치 및 전체 포팅 목표도 미완료다. 상세 해시/경계는 인벤토리 R92 절에 기록하고 과거 문서 suffix bytes를 보존한다.

## R91: 시작 능력치 제한 비교와 공격 후 live sound 조회 복원

- 기능 `0242df01c475c6f5e6bd3a8cc0429b053d5b8aa9`는 기존 `monsterActionMeleeUpdate532440` 본체 하나만 수정하고 새 테스트 파일만 추가했다. strike callback 뒤 원본424300 getter처럼 live class/UpdateData에서 sound set을 다시 읽으며, audio 뒤 완료 byte는 entry update에서 읽는 순서를 유지한다. signature/imports/다른 기존 본문·테스트·시나리오·자산은 그대로이고 커밋 직후 즉시 push actual0 후 clean ARM64 제품을 빌드했다.
- 변조하지 않은 원본 player-read-values와 x87 helper의 **24개 제한 실행**은 stock SOLO/ARENA·level1·max-stat off·세 class·명시적 max-health endpoint scale1/3 모델이다. nearest027F의12개는 HP20/20와 나머지 관찰값이 native와 같았다. toward-zero0E7F는 carry만1424/1425로 달랐으나 실제 원본 caller CW/world가 미증명이므로 전역 rounding/carry를 바꾸지 않았다. 원본 melee와 후킹하지 않은 getter의 **288개 제한 실행**에서 callback 후 live 상태 차이를 확인했다. strike/audio/pop만 격리 경계이며 전체 원본 전투는 **UNPROVEN**이다.
- 새 Go/C-owned 고주소 시험603 identity는 callback matrix와 audio/quiet/admission/fault 경계를 검사한다. compiled red **actual1**(237 pass·366 test fail+1package fail)을 보존하고 같은 frozen 테스트로 green603 pass를 확인했다. 최초 private Counter 집계 TypeError actual1도 보존했으며 private key 집계만 교정했다. 일반·HD·실제cgocheck2 전체49,993 pass/41skip·38package, server49,229 pass/7skip·35package·actual0이고 기존 ID/횟수/skip는 모두 같다. strict 및 cgocheck2+race+checkptr2는 각각 count3·641 identity/1,923 pass·3package·fail0/skip0다.
- 보고된 NPC Sword raw3/type7→Skeleton은 기존 R79 복원을 실제 C→SkeletonDamage→DefaultDamage→UnitSetHP로 재검증해 `2000→1997→1994→1991`이다. type7이나 피해 코드를 다시 바꾸지 않았다. clean Go1.26.5 ARM64 제품3개/help와 GUI 제품4개까지7개 metadata를 확인했다. 변경하지 않은 일반/HD headless 거미 전투는 frame764·거미 HP0/12·플레이어 HP14/20·actual0이고, 일반/HD 지속 공격44조건도 각각132개 distinct nonpoison 피해·반격 없이 각3회·actual0이다. 전용 서버의3맵 전환·4회 loop/tick 증가·quit exit0도 통과했다. 선언된 기존 fixture 범위이며 고정 golden/실제 오디오 장치/전체 자연 world 인증은 아니다.
- source/protected/personal/stock 전체 rows와 과거 문서 suffix를 보존했다. 통합 감사는 `/private/tmp/opennox-player-start-r91.9faufS/audit-runtime.log`, actual0 SHA-256 `143a4c90af8232efbce57c6ddc097b08ab29cd54a26a58770b8b2a9dda874c40`다. R89 Skeleton 일반/HD 처치 **actual2**·2400tick timeout은 여전히 보존·미해결이고 이번에 재실행하지 않았다. 성공한 제한 비교/GUI로 실패를 green화하거나 HP/seed/aggression/입력/timeout/기대값을 바꾸지 않는다. passive-food·고정 Quest stage3 맵 불일치 및 전체 포팅 목표도 아직 미완료다. 상세 경계·해시는 인벤토리 R91 절을 참조하며 최신 docs revision 제품은 별도 handoff로 검증한다.

## R90: 원본 공격 주기·Sword 산술의 제한 실행 비교

- 기준 문서 HEAD `22f4ae2f30c49aa474654eab31d5bc4ad5327701`, 기능 revision `e7bf3260514b8d0ef582469681a9638f9bd0faf6`의 생산 코드·테스트·자산은 이번 회차에서 변경하지 않았다. NPC Sword raw3/type7→Skeleton의 기존 R79 수정도 유지한다.
- R89 clean HD의 97-frame trace와 봉인된 원본 PE32/thing.bin을 이용했다. 원본 animation50A850, melee532440, player-attack538960, bolt-damage4EF1E0 명령의 타격/종료 게이트와 산술을 후킹하지 않고, 격리 서비스 경계만 공급하여 **86개 제한 실행**(monster60/player26)을 비교했다. Sword asset은 6 frames/duration0, Skeleton은 8 frames/duration2/Slave다.
- Skeleton 타격677/703/729, Sword 타격677/690/703/716, melee pop689/715, Sword float32 피해5/type0이 기록된 실행과 일치했다. 최초 private 분석 actual1의 count-reset 가정 오류와 원본 로그를 보존한 뒤 private 분석만 교정했다. v2 **actual0** SHA-256 `2c54c17173cc489eb159571b0978dbd060864bec31c50833ad4de26f140d6d40`이며, 기존 생산 함수·테스트·기대값을 수정하지 않았다.
- **전체 원본 전투 동등성은 UNPROVEN**이다. 실제 world 충돌/초기 능력치/AI/RNG/사망 이후를 이 실험에서 증명하지 않았다. 기존 normal/HD Skeleton 처치 E2E의 **actual2**와 2400-tick timeout은 여전히 미해결이며 새 GUI 성공으로 세지 않는다. 상세 범위·해시·증거 경로는 `PORTING-INVENTORY.md`의 R90 절을 참조한다.

## R89: 읽기 전용 전투 프레임 관찰과 NPC Sword 피해 유지 검증

기능 `e7bf3260514b8d0ef582469681a9638f9bd0faf6`는 기존 `e2eRun` 본체에 한 줄을 추가하고 새 관찰/테스트 파일만 도입했다. `NOX_E2E_COMBAT_TRACE=true`인 E2E 재생에서 실제 game frame과 detached player/monster 상태·native 주소·binary32 bits·Logic/Other index를 읽는다. holder 외 게임 상태·입력·HP·AI·inventory·clock을 쓰거나 추가 RNG draw를 하지 않는다. 240-frame cap은 최대 기록 범위이지 최소 sample 수나 변경된 timeout이 아니다. 피해 코드/기존 테스트/자산은 이번에 편집하지 않았다.

보고된 NPC1343/subclass69778·Sword1174→Skeleton1314/subclass70210의 raw3/type7은 기존 R79 복원 이후 실제 C→SkeletonDamage→DefaultDamage→UnitSetHP에서 `2000→1997→1994→1991`이다. 두 strict count3 lane에서 정확한 metadata·방어/무적 제어를 재검증했으며 type7을 다른 피해 유형으로 바꾸지 않았다. 전체 일반/HD/실제cgocheck2는 각49,390 pass/41skip·38package, server48,626 pass/7skip·35package·actual0이다. 모든 R88 ID/횟수/skip을 유지하고 새16개 관찰 identity만 추가했다. strict 표적 및 cgocheck2+race+checkptr2는 각45identity/135pass·fail0/skip0·3package·actual0이다.

보존된 같은8입력 Skeleton 진단은 일반/HD 모두 실제 C-owned 고주소 graph에서97개 실행 frame663..759를 기록했다. 플레이어 `20→12→4→0`, Skeleton `40→35→30→25→20`으로 양쪽 피해가 적용됐다. 사망 후 game frame이 멈춘 상태에서 기존2400tick 처치 대기가 timeout하여 둘 다 actual2다. 일반은 precommit dirty 관찰 제품, HD는 위 clean 기능 revision이다. target death/후속 Screen/PNG에는 미도달이다. 실제 snapshot scalar와 Logic/Other index는 두 lane이 같지만 원본 Windows 전투 타이밍·전체 RNG draw 횟수/월드 동등성은 아직 증명하지 않았다. 이 실패를 HP/seed/장비/공격성/입력/timeout 또는 기존 fixture 수정으로 통과시키지 않는다. 최초 private 분석기의240개 필수-sample 오해로 인한 actual1도 보존하고 분석기만 upper-bound 검사로 고쳤다.

clean 공식 ARM64 일반/HD/서버3제품의 Go1.26.5·revision·clean metadata와 help actual0을 확인했다. 실제 게임 실행 제품3개까지 총6개 metadata를 검증했다. 관찰을 끈 변경하지 않은 거미 전투는 일반/HD 모두 actual0·frame764·거미 HP0/12·플레이어 HP14/20·정상 cleanup이다. 전용 서버의3맵 전환·4회 loop/tick 증가·실제 quit exit0도 통과했다. 거미 회귀나 감사 actual0을 Skeleton actual2의 성공으로 대신 분류하지 않는다. GUI는 headless/mock/noaudio로 실행했으며 실제 장치 인증은 아니다.

모든 wrapper 종료 후 source/protected/personal 및 stock 전체 row 불변을 확인했다. 개인 Save/config·tests/YAML/PNG/oracle·캐시·실패 증거를 삭제/변경하지 않았다. 통합 기록은 `/private/tmp/opennox-combat-trace-r89.LkkHs3/audit-runtime.log`, SHA-256 `85bb50047d3f4c058f99e1b30c4cffb544fbb7ea9dda32004dcfa1448d459689`다. 최종 docs revision 제품은 별도 handoff로 검증하며 아래 기존 문서의 모든 바이트를 유지한다. Skeleton 처치/원본 타이밍·passive-food·고정 Quest 기대맵 불일치 및 전체 포팅 목표는 미완료다.

## R88: 사망 파편의 원본 float 저장 복원과 NPC Sword 피해 재검증

기능 `59131f3b4458fa54a3f6e1882ec276bb8a559f32`는 기존 `monsterActionDeadRuntime544D80` 본체의 `DeadFunc.RandomFloat` 연결 한 줄만 새 native helper로 교체한다. 원본 `00416030`의 binary32 scale·53-bit chop 연산과 사망 호출부 `FSTPS`의 round-toward-zero 저장을 유지하며 RNG 소비 횟수·전역 FPU 모드·기존 사망 callback/테스트/자산은 바꾸지 않는다. 실제 등록 `MECHGOLEMDEAD` 연결의 높이·수직 속도 오차를 compiled strict red로 재현했다.

변경하지 않은 GAME.EXE의 `00416030`과 실제 `00549F20` 저장 명령을 hook 없이 실행한20,480개 결과가 새 helper와 일치한다. 결과 SHA-256은 `e5368033a096fdb9a47afa2d3171d9f9ae45c91499fddb8994e7ed4b95ca2425`다. C-owned 고주소 객체를 쓰는 실제 native 사망 연결은 두 시작 index에서8개 파편씩 만들고 Logic `0→32`, `4090→26`·Other17·frame1400·HP19·파편 수명을 유지한다. 배치는 명시된 외부 관찰 서비스이며 전체 원본 사망/월드 물리 replay 인증은 아니다.

일반·HD·실제 cgocheck2 전체는 각각49,374 pass/41 skip·38 package, server는48,610 pass/7 skip·35 package·actual0이다. 기존 두 테스트 이름의 ASLR 주소만 정규화하여 모든 이전 ID/횟수/skip 불변을 확인했고 새32 identity만 추가됐다. strict 표적 및 cgocheck2+race+checkptr2는 각각 count3·152 identity/456 pass·fail0/skip0이다. 보고된 NPC1343/subclass69778·Sword1174의 raw3/type7→Skeleton1314/subclass70210은 실제 C→SkeletonDamage→DefaultDamage→UnitSetHP에서 `2000→1997→1994→1991`이다. 이 피해 수정은 기존 R79이며 이번에 피해 production을 다시 편집하지 않았다.

clean ARM64 일반/HD/서버3제품의 공식 build/metadata/help와 기존 거미 전투 headless/mock 일반·HD는 actual0이다. 두 게임 모두 frame764에 거미 HP0/12·플레이어 HP14/20을 확인하고 정상 종료했다. 별도 새 private 스켈레톤8입력 진단은 플레이어 사망 뒤 처치 대기2400tick timeout·actual2이며 사망 후 진단 화면에는 도달하지 못했다. 이 실패를 삭제하거나 HP/공격성/입력/timeout을 재조정해 통과시키지 않았고 NPC Sword C 검증 또는 사망 파편 연결 성공으로 대신 분류하지 않는다. 전용 서버의3맵 전환·4회 loop/tick 증가·실제 quit exit0도 확인했다.

모든 wrapper의 실제 종료 뒤 source/protected/personal 및 stock의 전체 row 불변을 확인했다. compiled red 이전 두 새 fixture 오류와 sandbox emulator SIGILL4도 host 승인 실행 actual0과 구분해 보존한다. 통합 기록은 `/private/tmp/opennox-death-rng-r88.zTNeq7/audit-runtime.log`, SHA-256 `fa9b38bcc109a53f29fa3ffe21b319596c7a15e3e97708318ec6683142d673d8`다. 개인 Save/config·기존 tests/YAML/PNG/oracle·캐시·실패 증거는 삭제/변경하지 않았다. 최종 docs-only revision 제품은 별도 handoff로 검증한다. 추가 스켈레톤 진단·passive-food·고정 Quest 기대맵 및 전체 포팅 목표는 미완료이며, 아래 모든 기존 문서 바이트를 유지한다.

## R87: 공개 RoamFlag 스크립트 API와 clean ARM64 제품 검증

ns/v4 단일 객체와 그룹의 `SetRoamFlag` panic을 실제 호출로 재현한 뒤 각각 기존 함수 본문 하나씩만 변경하여 native low-byte store/원본 no-stop 그룹 순회에 연결했다. 기능 커밋은 `5030576b7a6f315e91eeb7f765b6ada4c788720a`와 `f9d28d684b484785a35f995bd1bfda9023e06ed8`이며 기존 classic builtin·기존 테스트·기대값·자산은 변경하지 않았다. 새 C 소유 high-address 테스트는 중첩 그룹/ABI32 ID·nil/destroyed/nonmonster·wide/음수 flag에서 전체 객체/list/update byte 불변과 실제 SCRIPT03 결과 일치를 확인한다. 그룹 포함 strict/count3은97 identity/291 pass·fail0/skip0이다.

같은 clean 기능 revision에서 일반·HD·strict 전체 테스트 각각49,342 pass/41 skip, server48,578 pass/7 skip이며 actual0/fail0다. 조건부 기존 skip은 미실행으로 남긴다. 별도 실제 cgocheck2+race+checkptr2/count3은116 identity/348 pass·fail0/skip0이고, 보고된 NPC Sword raw3/type7 → Skeleton의 실제 C 체력 `2000→1997→1994→1991`을 유지한다. 공식 일반/HD/서버 제품3종과 실제 GUI 실행 제품을 Go1.26.5·darwin/arm64·위 revision·clean metadata로 검증했다.

headless/mock 일반·HD의 기존 선제공격16조건씩은 actual0이며 enemy가 먼저 맞지 않은 상태에서 스스로 대상을 획득하고 실제 HP/client delta를 만들었다. 기존 배치·durable HP·NPC aggression/equipment fixture를 바꾸지 않았다. 기존 Move/Wander4조건은 동일 YAML을 private copy로 실행해 actual0·ordinary tick·native slot 재사용·HP 불변이다. 이 private copy는 `Screen`의 자동 PNG 생성이 저장소에 쓰이지 않도록 격리한 것으로 기존 golden 비교를 대신하지 않는다. GUI는 기존 AI 경로의 회귀이고 새 ns/v4 API는 직접 native/SCRIPT03 비교 테스트로 검증했다. 전용 서버는 기동·3맵 전환·4회 loop/tick 증가·실제 콘솔 quit exit0을 통과했다.

모든 runtime wrapper 종료 후 source/protected/personal 및 stock checkpoint의 불변을 확인하고 문서만 덧붙였다. receipt는 `/private/tmp/opennox-script-roam-r87.IYOmJ3`, runtime audit stdout SHA-256은 `9f102ef26b50cac6abd63da37f7e0d51e662f15cb8e198ddc4b4c6a513150544`다. passive-food·고정 Quest 기대맵 불일치 및 전체 포팅 목표는 아직 미완료다. 이 연결을 음식/Quest 해결이나 전체 Windows 월드 동등성으로 확대 해석하지 않는다.

## R86: Quest 20단계 런타임 관찰과 NPC Sword 피해 재검증

기존 clean revision `7ad4f63155e054142b4852a2023bad776bf53717`에서 변경하지 않은 보조 `host-quest-twenty-stages-map-oracle.yaml`을 일반·HD ARM64/headless/mock으로 각각 실행해 actual0을 확인했다. 각 실행은 실제 선택 맵과 로드된 맵의 일치, 20단계 플레이 가능 상태·generator 설정, 19회 일반 입력 이동, 19회 exit callback 전환, 5단계 Hecubah/Necromancer·20단계 Hecubah 생존을 관찰했다. exit 진입은 기존 시나리오의 명시적 위치 배치 fixture이며 전체 자연 Quest 전투 완료 주장이 아니다.

별도 읽기 전용 audit는 총40개 detached 선택 전후 snapshot에서 signed DWORD 원본 선택 규칙·RNG table 결과·Logic 소비 및 history/frame/Other 불변을 확인했다. 일반/HD 입력·결과는 native 주소 외 동일하다. 진단 PNG10개는 5·6·20단계 플레이 화면과 직전 스테이지 score를 검사했고 기존 golden을 수정하지 않았다. 같은 입력에서 selector가 원본과 맞는다는 증거이지 전체 월드 RNG 소비의 Windows 동등성 인증은 아니다. 기존 고정 맵 검사는 재실행 actual2이며 3단계 `g_lotdd` 기대와 `g_crypts` 실제의 불일치를 그대로 보존한다. 보조 검증으로 이 실패를 대체하지 않는다.

보고된 NPC1343/subclass69778 → Skeleton1314/subclass70210, Sword1174의 raw3/type7은 기존 기능 `15ae300d3bfa8af59932c01e236e55c2bbbe53fc` 이후 실제 C 경로에서 HP `2000→1997→1994→1991`이다. Go1.26.5·실제 cgocheck2/count3 재검증은19 identity/57 pass·fail0/skip0·actual0이다. 직접 C damage dispatcher의 제어 fixture이며 자연 NPC 전투 전체 검증과 구분한다. 이번에는 production·기존 테스트·YAML·기대값·RNG·HP·자산을 편집하지 않았고 모든 wrapper 종료 뒤 문서만 덧붙인다. 음식 실패와 전체 포팅 목표는 미완료다.

독립 기록은 `/private/tmp/opennox-quest-stage-r86.c65oUK/audit-runtime-v1.log`, SHA-256 `032c6dc8baa759392829445b314853e2ffa0d018ea8f78a64079a008f49c512c`이다. source·보호 대상·개인 설치는 게임 전후 byte-identical이며 stock oracle tree도 기존 봉인값과 같다. 첫 strict 실행의 상대 경로 launch 오류는 child 미실행으로 기록하고 루트에서 실행한 v2 통과와 구분했다. 캐시·제품·실패 증거를 삭제하지 않았다. 아래 모든 기존 기록의 바이트를 그대로 유지한다.

## R85: Turn Undead 충돌의 공통 피해 경로 복원

기능 `babe3832dd50af3dc434946cf8af26537a6b9c7e`은 기존 `DefaultDamageWorld4E0B30` production 본체 하나만 변경하고 새 테스트 두 파일을 추가했다. MISSILE이 아닌 stock LIGHT|CLIENT_PREDICT collider의 type6/DISPEL_UNDEAD와 무소유 terminal-owner identity를 공통 damage tail에 연결한다. 수정 전 실제 C 충돌은 예산을 소모하고도 HP를 줄이지 않았다. signed raw type6·원래 적대/무적/NoUpdate/Shield·피격/효과음/attribution 순서는 유지하며 PLAYER 전용 type6 tail 또는 전체 사망/삭제 lifecycle을 새로 이식했다고 주장하지 않는다.

unchanged PE32 원본576행/49,252바이트의 결과·필드·순서와 native가 일치했다. SHA-256 `98d9f58f6d869d60a99d9c4bd9eb9c8a22e51371ef4b15543389911d63025587`이며 명시된 game/enchant/enemy/sound/guide/Shield/HP 외부 서비스를 쓰는 bounded 비교다. 별도 실제 C-owned 고주소 C collision→등록 SkeletonDamage→DefaultDamage→UnitSetHP에서 player/NPC/unowned source의 입력 예산7마다 HP `2000→1993→1986→1979`를 확인했다. 피해/HP 결과는 대체하지 않으며 DelayedDelete만 월드 서비스 관찰 경계다. 무적·NoUpdate·non-undead 음성 조건과 source metadata도 검증했다. 기존 NPC→Skeleton Sword raw3/type7의 실제 C 경로는 HP `2000→1997→1994→1991`을 유지했다.

새36 test identity는 일반/HD/실제 cgocheck2/race+checkptr=2 각각 count3 통과했다. 각 표적 lane actual0·570 정규화 ID/1,716 pass·fail0/skip0이다. 주소만 정규화한 기존 한 fault ID는9회, 나머지는3회이며 R84 표적477 ID/횟수를 보존했다. full 일반/HD/strict는 각각49,249 pass/41 skip, server48,485 pass/7 skip·actual0이고 기존 모든 ID/횟수·skip은 불변이다. compiled red와 새 private enchant hook/InitData fixture 실패도 별도 로그에 보존했다.

clean 기능 revision의 공식 ARM64 일반/HD/서버3제품 build/metadata/help 및 기존 Turn Undead 일반·HD headless/mock 시나리오는 actual0이다. GUI는43개 투사체의 생성/FX/갱신/취소를 검증하며 자연 target 피해나 실제 화면/스피커 인증은 아니다. 실제 피해 검사는 위 C 충돌 테스트와 구분한다. 개인 Save/config·원본 자산·기존 테스트/YAML/golden은 불변이고 캐시는 삭제하지 않았다. 최종 docs revision 제품은 별도 handoff다.

자세한 범위와 한계는 인벤토리 R85 절 및 `/private/tmp/opennox-turn-undead-damage-r85.xnmwfD/audit-tests.log`(SHA-256 `6b6f427f2c86c4bcaa4e5bb7049f048456bb7523c572254f1ed2ca1dba7068f1`), `audit-games.log`(`550f4d4b31bad269046e99138463685e30f2b77e18897db90b4379aaae42308c`)를 따른다. R82 음식/고정 Quest actual2는 재실행하거나 해결했다고 주장하지 않으며 전체 포팅 목표도 미완료다. 아래 모든 기존 기록의 바이트는 그대로 유지한다.

## R84: armed EXPLOSION GreatSword 방어 후속 완료

기능 `903de8ca0a58f807d4c524d168b56415332d13bc`은 기존 armed EXPLOSION 한 production 본체의 rejection 한 줄만 finisher 호출 세 줄로 교체하고 새 helper 1개/테스트 3파일을 추가했다. admission 뒤 callback-created MISSILE의 GreatSword reflection·owner·sound·state/action·live inventory·실제 내구도와 broken-item 종료를 원본 순서대로 처리하며 prefix를 반복하거나 type7을 다른 피해 유형으로 바꾸지 않는다. ordinary stock Sword가 자연스럽게 MISSILE로 바뀐다는 주장은 아니다.

unchanged 원본의 5,904개 제어 transcript와 nil-item 실제 fault 2개, native service-use와 invalid-layout 음성 경계, C-owned 고주소/등록 생산 HP·wear 검사가 통과했다. 새 45 test identity를 일반/HD/cgocheck2/race+checkptr=2 각각 count3 검증했으며 각 lane actual0·1,437 pass·fail0/skip0이다. 주소만 정규화한 477 ID 중 기존 한 fault ID는 9회, 나머지는 3회다. full 일반/HD/strict는 각49,213 pass/41 skip, server48,449 pass/7 skip·actual0이며 기존 ID/횟수는 불변이다. 보고된 NPC→Skeleton Sword raw3/type7의 실제 C 피해 경로도 각 lane에서 HP `2000→1997→1994→1991`을 유지했다. 새 CallDamage registry fixture와 이 실제 C-entry 검사는 구분한다.

clean 기능 revision의 공식 macOS ARM64 3제품 build/metadata/help와 일반·HD headless/mock 기존 GreatSword 공개 시나리오 각21결과가 actual0이다. 정면11조건의 HP 보호·대검 wear와 후면10조건의 HP 피해·wear 없음, 자연 NPC Fireball을 관찰했다. 이 GUI는 공통 방어 회귀이며 callback-created armed-Sword의 자연 발생·실제 화면/스피커 인증은 아니다. 개인 Save/config·원본 assets·기존 테스트/YAML/golden은 불변이고 캐시는 삭제하지 않았다. 최종 docs revision 제품은 별도 handoff다.

상세 scope·원본 경계·compiled red와 새 fixture의 GameBall registration 실패·private auditor 표시 공백 실패 보존은 인벤토리 R84 절을 따른다. 증거는 `/private/tmp/opennox-armed-greatsword-r84.v3Aob4/audit-tests.log`(SHA-256 `22c40feb6ebe7f6a2033988842e4e84dd7c30c7a688ca5cbc7666b1e8fc4be44`) 및 `audit-games-v2.log`(`857667c01b4b6345f1ae7149739aa70e1f3891a19117e9e7fb18f9ffd3805f89`)이다. 기존 R82 음식/고정 Quest actual2와 전체 포팅 목표 미완료는 유지한다. 아래 모든 기존 기록은 원래 바이트 그대로 보존한다.

## R83: 콜백·충돌 감사 및 NPC Sword 피해 재검증

기존 코드·테스트·e2e/oracle·자산을 바꾸지 않은 읽기 전용 조사에서 원본 root 갱신 순서, 음식 producer의 native-width 인자, MOVE_TO/PICKUP dispatch와 72개 action metadata를 대조했다. 원본 음식 충돌은 몬스터 자동 줍기를 처리하지 않는다. 이 근거만으로 강제 pop/소비를 추가하지 않았다. 원본 전체 root·월드 물리를 새로 replay한 결과나 공개 음식 게이트 해결로 분류하지 않는다.

일반/HD/cgocheck2/race+checkptr=2의 관련 회귀는 각 count3/636 identity/1,908 pass·fail/skip0·actual0이며, native pickup collision strict도 5 identity/15 pass·actual0이다. 보고된 NPC→Skeleton Sword raw3/type7은 실제 C 피해 경로에서 HP `2000→1997→1994→1991`을 네 lane 모두 유지한다. R79/R80/R81 기능 재검증이며 이번 production 변경은 없다. 원본 phase 26개와 반복 native 312개 phase 비교·불변 checkpoint·actual terminal은 인벤토리 R83 절과 `/private/tmp/opennox-passive-move-r83.Qji3qe/audit-diagnosis.log`(SHA-256 `9201b5c27cbfe84803adf66656e3e181a485cdbf7cbf0dc9814be56b6e2f6942`)를 따른다. R82 음식 소비 timeout·고정 Quest stage3 actual2와 전체 목표 미완료는 유지한다.

## R82: 수동적 음식 접근 진단과 남은 실행 게이트

`72dd429054b980de144a57cd9b687e4f5dfdc74c`은 새 AI 회귀 테스트 한 파일만 추가한 커밋이다. 기존 생산 코드·테스트·e2e 기대값·oracle·게임 데이터는 수정하지 않았다. Troll의 두 stall 구간과 NPC/synthetic due-tick 조합에서 원본 Main→dependency→close MOVE_TO의 26개 상태와 native가 일치했다. 원본도 해당 제한 입력에서 WAIT 후 IDLE로 전환하며 음식을 소비하지 않았다. 이는 실제 월드 이동·물리·regen·전체 RNG 스트림의 동등성 또는 음식 소비 버그 해결을 뜻하지 않는다.

고유 leaf 6개는 일반/HD/cgocheck2/race/checkptr=2 각 `-count=3`에서 통과했다. 네 full lane까지 합쳐 494개 phase row를 비교했다. 전체 일반·HD·strict CGo는 각각 actual exit 0, pass event 49,168/skip 41, 서버 전용은 pass 48,404/skip 7이다. 기존 pass/skip과 횟수는 유지됐고 새 leaf 및 parent의 pass event 7개만 추가됐다. 원본 비교의 서비스 입력과 검증 한계, raw evidence digest는 인벤토리 R82 절을 따른다.

공식 client/client-hd/server 세 파일은 Go 1.26.5, darwin/arm64, clean 테스트 커밋의 metadata와 `-help` actual exit 0을 확인했다. headless/mock GUI 실행은 개인 Save/config를 격리했고 보호 대상 및 개인 설치 파일의 tree hash는 실행 전후 동일했다. 이번 반복에서 캐시는 제거하지 않았다.

공개 게임 내 게이트는 아직 두 개가 실패한다. 음식 게이트는 production 동일 parent에서 첫 Troll ascending apple의 비강제 소비 600tick timeout으로 actual exit 2이며 나머지 조합의 완료를 주장하지 않는다. 고정 20단계 Quest는 이번 커밋으로 새로 실행하여 stage 3 기대 `g_lotdd`/실제 `g_crypts`, actual exit 2를 재현했다. 기대값·RNG·맵·timeout·golden을 바꾸지 않았으며 전체 포팅 목표도 완료로 표시하지 않는다. 다음 단계는 실제 음식 접근의 물리·충돌/완료 처리와 전체 runtime RNG 소비를 조사하는 것이다.

통합 증거: `/private/tmp/opennox-passive-food-diagnosis-r82.jezZz5/audit-evidence.log`, SHA-256 `e4c57b48b3e5e73fe41f4e7e3437d2ba859734472ea9edf050b95f0d0d36d63f`.

2026-10-11 R81 `fc8b903e2`는 R79/R80의 armed EXPLOSION(7) 공통 피해 복원에 이어 PLAYER 및 MONSTER NPC의 PlayerDamage 방어 진입을 복원했다. 기존 `PlayerDamageNative4E17B0` 본체에4줄만 추가하며 Sword/qualifying WAND/Hammer의 실제 type7·attribution을 그대로 유지한다. 원본7194조건/638529byte 결과·순서와 실제4GiB 초과 C-owned NPC/player의 방어구 흡수·fractional carry·내구도·반복 HP 감소를 확인했다. 새113 test identity의 strict count3/339pass, 이전 모든 identity/action/multiplicity/skip 보존, 일반·HD·strict 전체 각49161pass/41skip·server48397pass/7skip 및13개 직렬 gate가 actual0이다. clean/pushed ARM643제품과 일반·HD headless/mock의 기존 Wiz01a 사망·실제 Done 이후5tick 시체 이동도 유지됐다. [R81 근거와 한계](PORTING-INVENTORY.md#armed-explosion-공격의-playerdamage-방어-진입-복원)에 통제된 원본 replay·실제 C 피해 검증·기존 GUI 회귀를 구분한다. 기존 문서 모든 바이트·원본 자산·개인 Save/config는 보존하며 docs-only 제품은 별도 handoff다. 원본 nil-shield fault는 양성 pass가 아니고 optional gameex 방어는 stock oracle 인증이 아니다. 자연 전투 전체·retained R77 Quest/passive-food·원격 transport·전체 port goal은 여전히 미완료다.

2026-10-11 R80 `78d392c6d`는 R79의 NPC Sword EXPLOSION(7) 복원에 이어 이미 계산된 피해를 받는 PLAYER의 공통 DefaultDamage 경로를 복원했다. 기존 `DefaultDamageWorld4E0B30` 본체 하나만 변경하며 PlayerDamage의 방어구 흡수·carry·hit marker·GodMode/Quest prefix를 다시 실행하지 않는다. unchanged GAME.EXE의774조건/78268byte 결과·순서가 일치하고 실제4GiB 초과 C-owned NPC/player의 Sword/Wand/Hammer raw3 반복 HP2000→1997→1994→1991 및 무적을 확인했다. 새30 test identity의 strict count3/90pass, 기존 모든 identity/action/multiplicity/skip 보존, 일반·HD·strict 전체 각49048pass/41skip·server48284pass/7skip와13개 직렬 gate는 actual0다. clean/pushed ARM643제품과 일반·HD headless/mock의 기존 Wiz01a 사망·실제 Done 이후5tick 시체 이동도 유지됐다. [R80 근거와 한계](PORTING-INVENTORY.md#플레이어의-armed-explosion-공통-피해-경로-복원)에 통제된 C DefaultDamage 검증과 자연 GUI 전투를 구분한다. 기존 문서 모든 바이트·원본 자산·개인 Save/config는 보존하며 docs-only 제품은 별도 handoff다. 같은 type7의 PlayerDamage 전용 PLAYER/NPC 방어 진입은 아직 미이식으로 다음 수정 단위이며 retained R77 Quest/passive-food·원격 transport·전체 port goal도 미완료다.

2026-10-11 R79 `15ae300d3`는 NPC의 Sword가 Skeleton에 전달하는 EXPLOSION(7) 피해를 복원했다. 기존 `DefaultDamageWorld4E0B30` 본체 하나만 변경하며, 원본의 적대 관계·무적·화염 저항·보호막·피격/효과음 순서를 유지한다. 보고된 type/class/subclass/flags 조합의 실제 C 피해 처리에서 raw3 반복 HP `2000→1997→1994→1991`을 확인했고, 새 39 test identity가 실제 strict count3에서 117pass였다. 변조하지 않은 GAME.EXE의 918개 조건과 결과/순서가 일치하며, 일반·HD·strict 전체 각49018pass/41skip, server48254pass/7skip와 13개 직렬 gate가 actual0이다. 기존 PCM/OpenAL null 테스트·모든 이전 test identity/action/multiplicity/skip·원본 자산·개인 Save/config를 보존했다. Clean/pushed ARM64 일반·HD·서버 제품과 일반·HD headless/mock의 기존 Wiz01a 사망·실제 Done 이후5tick 시체 이동도 통과했다. [R79 근거와 한계](PORTING-INVENTORY.md#npc의-sword-explosion-피해-복원)에 통제된 C 피해 검증과 자연 GUI 검 전투를 구분한다. 기존 문서 모든 바이트를 보존하며 docs-only 제품은 별도 handoff다. 자연 검 전투·retained R77 Quest/passive-food·원격 transport·전체 port goal은 별도 미완료다.

2026-10-11 R78 `0731ee877`는 TODO였던 16-bit PCM WAV의 정확한 프레임 재개를 복원했다. 기존 `pcmDecoder.Seek` 본체 하나만 변경하고 독립 PCM suffix·read-ahead·여러 data chunk·EOF/rewind·잘린 파일·음수/MaxInt 위치의 새 129 pure identity와 OpenAL null의 17 identity를 추가했다. 실제 strict/null count3은 새 146 identity 모두 통과했고, 일반·HD·race의 전체 오디오도 각 996 pass/24 optional skip이다. 기존 모든 test identity/action/multiplicity/skip를 유지하며 일반·HD·strict 전체 각 48979 pass/41 skip, server 48215 pass/7 skip, 원본 oracle·strict NXZ와 clean/pushed ARM64 세 제품 build/revision/help가 actual0이다. 같은 clean revision의 일반·HD headless/mock에서도 원본 Wiz01a의 nil/nil raw500 BLADE 사망·실제 Done 이후 5tick 이동·client animation10이 유지됐다. [R78 근거와 한계](PORTING-INVENTORY.md#pcm-wav의-정확한-프레임-재개-복원)에 PCM 위치와 실제 청취 위치, 기존 R77 두 retained 실패를 구분한다. 기존 문서 모든 바이트·원본 자산·개인 Save/config는 보존하며 docs-only 제품은 별도 handoff다. 물리 스피커·모든 WAV 형식·retained Quest/passive-food·원격 transport·전체 port goal은 여전히 미완료다.

2026-10-11 R76 `593227043`는 공격자·무기가 없는 NPC BLADE(0) 피해를 기존 방어·내구도·체력·사망 경로에 연결했다. 기존 PlayerDamage monster 본체에3줄만 추가하고 원본의 무적/Dead/NoUpdate·cached absorption/marker·live carry·armor wear·late Quest 순서를 유지한다. `25b17fa88`는 보고된 NPC type1343/class2097154/subclass69650/flags150995460의 raw300 C→Go→native HP 감소(1000→700→400→100)와 원본 Wiz01a 대화 관찰을 추가했다. 일반·HD headless/mock에서 ApprenticeStart의 nil/nil raw500 BLADE가 미리 배치된 Apprentice2를 죽이고, 실제 Done 이후 원본5프레임 타이머가 시체를 이동하며 클라이언트 dead animation10을 표시한다. 새 NPC 생성이나 원본 map/script·체력·사망 상태의 강제 변경은 없다. 일반·HD·실제 strict 전체 각48850pass/40skip, server48215pass/7skip와 count3 표적/race·oracle·NXZ·clean/pushed ARM643제품 build/revision/help가 actual0이고 모든 기존 test 결과·자산·개인 Save/config는 불변이다. [R76 근거와 범위](PORTING-INVENTORY.md#공격자-없는-blade-npc-피해와-wiz01a-대화-후-사망-복원)에 fixture·실패 관찰·native HP와 실제 사망 검증을 구분한다. 이전 문서 모든 바이트를 보존하며 docs-only 최종 제품은 별도 handoff다. 전체 챕터 자연 진행·물리 화면/오디오·retained Quest/passive-food·원격 transport·전체 port goal은 별도 미완료다.

2026-10-11 R75 `31518bf50`는 WAV MP3의 채널별 frame 수를 interleaved sample 수로 정확히 변환하고 작은 버퍼의 미소비 PCM·정확한 seek/Position을 복원했다. 기존 wavReader.open 본체에3줄만 추가하며 raw minimp3·PCM/ADPCM·standalone MP3·tests/golden/자산은 불변이다. compiled red actual1과 strict green110pass, 새67 pure identity, 원래 한국어 대사9위치×count3/30pass, 실제 OpenAL null153pass가 확인됐다. R73/R74의 약33.9초 전체 chap1 반복·ADPCM 정확한 seek도 유지한다. 일반·HD·실제 strict 전체 각48495pass/40skip, server47860pass/7skip, 엄격 표적/race/checkptr2·unchanged oracle·strict NXZ 및 clean/pushed ARM643제품 build/revision/help·일반/HD headless/mock가 actual0다. 개인 전체 오디오 actual1은 누락된 한국어 기준1137개와 추가 raw MP3 때문에 compiled R74와 action/PCM hash/길이가 동일하다. [R75 근거와 한계](PORTING-INVENTORY.md#wav-mp3-스트림의-채널-샘플-수와-정확한-재개-복원)에 frame/실제 청취 위치·encoder padding·스피커 미검증을 구분한다. 기존 문서 모든 바이트·자산·Save/config를 보존하며 docs-only 제품은 별도 handoff다. Wiz01a/WoundedApprentice·retained Quest/passive-food·원격 transport·전체 port goal은 다음 작업으로 남는다.

2026-10-11 R74 `78c42466e`는 R73의 배경음 반복 상태 복원에 이어 ADPCM stream의 정확한 프레임 재개를 복원했다. 기존 wavReader.Seek 본체 하나만 변경하며 블록의 첫 predictor frame, 블록 중간의 미소비 PCM, 여러 block-aligned data chunk와 EOF/rewind를 새 adapter에서 처리한다. 기존 일반 Decode·PCM/MP3 seek·원본 WAV와 tests/golden은 불변이다. strict 최초 entry77pass, stock chap1의10위치×count3/33pass와 실제 OpenAL null lifecycle54pass가 통과하고 약33.9초 전체 음악 반복도 다시 확인했다. 일반·HD·실제 strict 전체 각48428pass/38skip, server47860pass/7skip, 엄격 표적/race/checkptr2·unchanged oracle·strict NXZ 및 clean/pushed 기능 revision의 ARM643제품 build/revision/help·일반/HD headless/mock가 actual0다. 전체 개인 오디오 검사는 기준 해시가 없는 한국어 대사1137개와 추가 chap1.mp3로 actual1이며, 변경 전 코드의 결과·PCM hash/길이와 완전히 동일하므로 통과로 바꾸지 않았다. 음악 WAV29개는 기존 기준을 통과했다. [R74 근거와 한계](PORTING-INVENTORY.md#adpcm-배경음-재개의-정확한-프레임과-블록-꼬리-복원)에 실패/재개 로그·optional skip·decoder frame과 실제 청취 위치의 구분을 남긴다. 이전 문서 전체 바이트·자산·개인 Save/config는 보존하며 docs-only 최종 제품은 별도 handoff다. 물리 스피커·모든 오디오 형식·Windows/Linux 전체 실행과 retained Quest/passive-food·원격 transport·전체 port goal은 여전히 미인증/미완료다.

2026-10-11 R73 `24d0c316a`는 chap1 배경음이 저장된 후반 재개 위치에서 짧게 반복하던 오류를 수정했다. 기존 Music.Module.Update 본체 한 줄만 바꿔 선택 슬롯을 복사본 대신 포인터로 갱신하며 Position 초기화와 Loaded 상태를 보존한다. 최초 재개 후 다음 반복은0에서 시작한다. compiled mock/native red와11mock 사례·유효 선택별4반복, 실제 OpenAL null에서 stock747519frame/22050Hz/약33.9초 전체 반복을 두 번 확인했다. 원본 WAV·디코더와 기존 PCM hash는 불변이다. 일반·HD·실제 strict 전체 각48370pass/36skip, server47860pass/7skip, 엄격 표적/race/checkptr2·기존 OpenAL lifecycle30pass·standalone ABI32 sanitizer·unchanged oracle·strict NXZ가 actual0다. clean/pushed 기능 revision의 공식 ARM643제품 build/revision/help 및 일반·HD headless/mock 기존 사망 시나리오도 actual0다. [R73 근거와 한계](PORTING-INVENTORY.md#chap1-배경음의-짧은-구간-반복과-재개-위치-상태-복원)에 재개 위치 복사 오류·실제 null 재생·optional skip와 물리 스피커 미검증을 구분했다. 이전 tests/문서 바이트·자산·개인 Save/config를 보존하며 docs-only 최종 제품은 별도 handoff다. 앞선 R72 handoff는 완료했지만 retained Quest/passive-food·disabled statistics gate·원격 transport·전체 port goal은 여전히 미완료다.

2026-10-11 R72 `39d8f729e`는 OblivionHeart의 ELECTRIC/AIRBORNE_ELECTRIC 피해를 기존 native DefaultDamage의 armed-unit 경로에 연결했다. 기존 production 본체 하나만 수정하며 signature·바깥 바이트·PlayerDamage admission은 유지한다. unchanged PE32 2598조건의 결과와 실제4GiB 초과 C-owned Troll/Urchin/SmallAlbinoSpider의18변형·각3hit를 확인했다. 원래 Shock/Enemy qualifier 순서·live 면역·방패·전기 보호 계산을 보존하고 TypeID/name 특례는 없다. 일반·HD·실제 strict 전체 각48369pass/34skip, server47860pass/7skip, 엄격 표적/race/checkptr2 각39354pass·3package와 standalone ABI32 sanitizer·unchanged oracle·strict NXZ는 actual0다. clean/pushed 기능 revision의 공식 ARM643제품 build/revision/help 및 일반·HD headless/mock 기존 사망 시나리오도 actual0다. [R72 근거와 한계](PORTING-INVENTORY.md#oblivionheart의-armed-electric-defaultdamage-복원)에 원본 제어 서비스·첫 fixture 실패·자연 발생 전투의 미인증 범위를 구분했다. 기존 tests/문서/자산/개인 Save/config는 보존하며 docs-only 최종 제품은 별도 handoff다. retained Quest/passive-food·disabled statistics gate·원격 transport·전체 port goal과 다음 배경음 반복 요청은 이 단위로 완료되지 않는다.

2026-10-11 R70 `8367816d4`는 self-sourced GolemArrow의 PIERCE 피해를, R71 `6c2e8eac8`는 PressurePlate/LargeFist의 CRUSH와 class-zero ImaginaryCaster/Fireball의 FLAME 피해를 기존 native PlayerDamage 경로에 연결했다. 각 기능 커밋은 기존 class admission 본체 하나만 변경하며 TypeID/name 특례나 방어구·저항·면역 계산 변경은 없다. unchanged PE32 prefix 각1164조건, 총3492조건의 sealed 결과와 실제4GiB 초과 C-owned damage→DefaultDamage→UnitSetHP 반복 호출이 통과했다. 일반·HD·실제 strict 전체 각48330pass/34skip, 지원 server47821pass/7skip, 엄격 표적/race/checkptr2 각39237pass·3package와 standalone ABI32 sanitizer·unchanged oracle·strict NXZ는 actual0다. clean/pushed R71의 공식 일반·HD·server ARM643제품 build/revision/help 및 일반·HD headless/mock의 기존 Spider 피해·사망 시나리오도 actual0다. 이는 새 공격 형태의 isolated C 검증과 common death tail GUI 회귀이며 세 공격의 자연 발생 맵 시나리오 전체 인증은 아니다. [R70/R71 근거와 한계](PORTING-INVENTORY.md#비유닛-공격-playerdamage-admission-복원)에 실패 로그·원본 외부 서비스·검증 범위를 구분했다. 이전 tests·문서 바이트·자산·개인 Save/config를 보존하며 docs-only 최종 제품은 별도 handoff다. retained Quest/passive-food·disabled statistics gate·원격 transport와 전체 port goal은 미완료다.

2026-10-11 R68 `4166d33e8`는 경쟁 모드 PlayerDie의 shared native core를 복원하고, R69 `64d4569d0`는 기존 사망 callback 본체 하나를 완성된 Arena/KotR/Elimination 서비스에 연결했다. 각 기능 커밋은 기존 production 본체 하나만 변경하며 signature·바깥 바이트와 기존 limited entry의 admission을 보존한다. 원본 PE32 사망/terminal-parent root11440조건·237fault의 전체 snapshot/순서와 실제4GiB 초과 C-owned CGo 사망17사례·lesson/death41패킷이 통과했다. signed death DWORD와 unsigned limit WORD의 두 번 live read, full-DWORD tracking, duplicate assist, cached update/common Quest·Coop·online tail을 유지한다. 일반·HD·실제 strict 전체48229pass/34skip, 지원 server47720pass/7skip, strict 표적/race/checkptr2 각131identity×count3, standalone ABI32 sanitizer·unchanged oracle·strict NXZ·clean/pushed ARM643제품 build/revision/help가 actual0다. 같은 clean 기능 revision의 일반·HD headless/mock 기존 사망 시나리오는 각각 실제 Spider 공격으로 HP75→0·state4·고주소 server/client binding을 관찰하고 정상 종료했다. 이는 common death tail 검증이며 실 경쟁 상대·경쟁 match 전체 인증은 아니다. [R68/R69 근거와 한계](PORTING-INVENTORY.md#경쟁-모드-playerdie의-native-core와-production-callback-연결)에 초기 원본 parent-service 가정·compile/path·fixture 틱/할당 오류와 보존된 red를 구분했다. 기존 tests/자산/개인 Save/config/ledger는 보존하며 docs 제품은 별도 handoff다. disabled statistics gate·원격 transport·retained Quest/passive-food·전체 port goal 및 다음 요청의 self-sourced GolemArrow 피해 admission은 이 단위로 완료되지 않는다.

2026-10-10 R67의 `016c6bacc`는 KotR scoring의 기존54DC40 production 본체 하나를 native pointer entry에 연결했다. 원본 crown/balance 선택·반복 live read·binary32 spill/FISTP 반올림·score/death/team wrap·lesson/observer/drop 순서를 보존한다. unchanged PE32 root/add/sub/FISTP의9944조건/224fault 전체 snapshot과 실제4GiB 초과 C-owned chain의 CGo10호출·lesson13/team3 packet이 통과했다. configured x87 replay의 synthetic NaN bit 일치는 실제 물리 x87 인증과 구분한다. 공간 부족으로 멈췄던 검증은 사용자가 허가한 전용 Go build cache만 pinned `go clean -cache`로 정리해 약14.9GiB를 확보한 뒤 재개했다. module cache·기존 제품·소스·게임/Save·검증 로그는 전후 hash가 같고, 제거한 cache는 다음 빌드에서 재생성된다. 일반·HD·실제 strict 전체48226pass/34skip, 지원 server47717pass/7skip 및 엄격 CGo/race/checkptr2 표적 각128identity×count3·standalone ABI32 C ASan/UBSan·unchanged oracle/strict NXZ·clean/pushed 기능 ARM643제품 build/revision/help는 actual0다. [R67 근거와 범위](PORTING-INVENTORY.md#kotr-scoring의-native-포인터-entry와-원본-crownbalance-순서-보존)에 실제 truncation red·NaN/인자 실패·retained no-space/partial audit·실행하지 않은 compression 거부·cache-only 재개와 외부 서비스 한계를 구분한다. 기존 tests/자산/개인 데이터/ledger와 Trap Set 회귀를 보존하고 이번 GUI 실행은 없다. raw competitive death admission·disabled statistics gameplay gate·다른 dependency·원격 transport·retained Quest/passive-food와 전체 port goal은 미완료다. docs 제품은 별도 handoff다.

2026-10-10 R66의 `9084ee11f`는 Elimination scoring의 기존54D7A0 production 본체 하나만 native 포인터 entry에 연결했다. 원래 score/death/team DWORD wrap, cached update/team·live Player load, suicide/environment/kill의 report/observer 순서를 보존한다. unchanged 원본2496조건/84fault의 전체 snapshot·순서가 native와 같고 실제4GiB 초과 C-owned chain의 CGo7호출·lesson10/team3 packet도 통과했다. 일반·HD·실제 strict 전체48222pass/34skip, 지원 server47713pass/7skip, 엄격 CGo/race/checkptr2 표적 각124identity×count3·standalone ABI32 C ASan/UBSan·unchanged oracle/strict NXZ·clean/pushed 기능 ARM643제품 build/revision/help는 actual0다. [R66 근거와 범위](PORTING-INVENTORY.md#elimination-scoring의-native-포인터-entry와-원본-deathteam-순서-보존)에 실제 raw truncation red·test 전 compile 실패·성공한 최초 제품 로그 수집의 overlap 및 보존/재검증, 외부 서비스와 raw caller 한계를 구분했다. 기존 tests/자산/Save/config/ledger와 Trap Set 회귀를 보존하고 이번 GUI는 실행하지 않았다. competitive death admission·비활성 통계 gameplay gate는 닫힌 상태이며 다른 미이식 dependency·원격 transport·retained Quest/passive-food·전체 port goal은 미완료다. docs 제품은 별도 handoff다.

2026-10-10 R65의 `c628a5228`는 Arena scoring의 기존 `54D980` production 본체 하나만 native 포인터 entry에 연결했다. 원래 DWORD score wrap·tracking nonzero·cached update/team과 live Player load·lesson/death/observer 호출 순서 및 실행된 PE32 fault 경계를 보존한다. raw int entry는 이미 잘린 고주소를 복원하지 않으며 competitive death admission은 계속 닫혀 있다. unchanged 원본3516조건/206fault와 scalar38조건/10fault의 전체 snapshot·순서가 native 결과와 일치하고, 실제4GiB 초과 C-owned chain의 CGo3호출·lesson packet5개도 통과했다. 일반·HD·실제 strict 전체48219pass/34skip, 지원 server47710pass/7skip 및 strict 표적/race/checkptr2 각121identity×count3·standalone ABI32 C ASan/UBSan·unchanged oracle/strict NXZ·clean/pushed 기능 ARM643제품 build/revision/help는 actual0다. [R65 근거와 범위](PORTING-INVENTORY.md#arena-scoring의-native-포인터-entry와-원본-scoreobserver-순서-보존)에 실제 truncation red와 disk/fixture/orchestration 실패, 명시적인 외부 서비스·raw caller 한계를 구분한다. 기존 증거·제품·전용 cache는 삭제하지 않고 byte-identical 압축으로 공간을 확보했으며 tests/자산/Save/config/ledger와 Trap Set 회귀를 유지한다. 이번 GUI 실행은 없고 docs 제품은 별도 handoff다. 다른 raw competitive death dependency·비활성 통계 gameplay gate·원격 transport·retained Quest/passive-food·전체 port goal은 여전히 미완료다.

2026-10-10 R64의 `1b41d026b`는 unsigned IP index/OwnIP/Streams 대응을, `f0d5f1da0`는 통계 세션 종료의 native Quest column/packet 연결을 각각 기존 production 본체 하나씩 복원했다. 원래 signed WORD count·live read/할당/저장/해제 순서·BYTE tag wrap·full DWORD sequence·packed header slots를 보존하며 gameplay 통계와 competitive death admission은 열지 않았다. unchanged PE32 root72/builder16/packet49조건 전체 snapshot·IP524조건의 실제 export 결과·실제 고주소 empty4종/nonempty3→0→1→3 CGo 호출이 통과했다. 일반·HD·실제 strict 전체48214pass/34skip, 지원 server47705pass/7skip과 엄격 CGo/race/checkptr2·standalone C ASan/UBSan·unchanged oracle/strict NXZ·clean/pushed ARM643제품 build/revision/help는 actual0다. disk no-space의 최초 lane 실패·serial 재실행, IP slot128 vacant 한계·외부 제어 서비스·원래 fault와 unsupported transport는 [R64 근거와 범위](PORTING-INVENTORY.md#통계-세션-종료의-native-quest-column과-unsigned-ip-인덱스-보존)에 구분한다. 기존 Trap Set 수정·tests/자산/Save/config/ledger를 보존하며 이번 GUI 실행은 없다. 다른 raw helper 직접 caller·비활성 gameplay gate·competitive death admission·원격 transport·retained Quest/passive-food와 전체 port goal은 여전히 미완료다.

2026-10-10 R63의 `dbbd4ab8c`는 통계 세션 시작의 기존426060 production 본체 하나를 native header·Quest cleanup·online report helper에 연결했다. 팀 목록의 pointer→int 축소를 제거하고 원본 packed header mode 기록·BYTE team wrap·signed WORD port·다섯 column 초기화 순서를 보존한다. unchanged PE32의 header320/root72/cleanup41조건 전체 snapshot과 실제 nonempty player/team CGo12회가 통과했다. 일반·HD·실제 strict 전체48203pass/34skip, 지원 server47694pass/7skip, 엄격 CGo/race/checkptr2·standalone C ASan/UBSan·unchanged oracle/strict NXZ와 clean/pushed ARM643제품 build/revision/help는 actual0다. [R63 근거와 범위](PORTING-INVENTORY.md#통계-세션-시작의-native-팀-목록과-packed-header-보존)에 외부 제어 서비스·원래 fault·unsupported transport·retained red를 구별했다. 기존 Trap Set 수정·tests/자산/Save/config/ledger는 불변이며 이번 GUI 실행은 없다. raw end·Quest builder/end와 비활성 gameplay gate·competitive death admission·retained Quest/passive-food 및 전체 port goal은 여전히 미완료다.

2026-10-10 R62의 `3c53e9c99`는 주기적 통계 bulk flush의 기존4263C0 한 production 본체를 native pointer/column/packet chain에 연결했다. 원래 DWORD count·sequence, WORD/BYTE 길이·tag wrap, record 갱신·host 우선 재등록과 buffer 초기화를 유지한다. unchanged 원본 packet34/flush32/lifecycle5조건의 전체 결과·순서가 같고 실제 빈 player list의 CGo/압축/포장 반복 호출도 통과했다. 일반·HD·실제 strict 전체48191pass/34skip, 지원 server47682pass/7skip, 엄격 CGo/race/checkptr2·ASan/UBSan·unchanged oracle/strict NXZ와 clean/pushed ARM643제품 build/revision/help는 actual0다. [R62 근거와 범위](PORTING-INVENTORY.md#주기적-통계-bulk-flush의-native-포인터와-원본-패킷-보존)에 제어 서비스·원래 fault·기존 unsupported transport abort를 구별했다. 기존 Trap Set 수정·tests/자산/Save/config/ledger는 불변이며 이번 GUI 실행은 없다. 다른 raw startup/end caller·competitive death admission·retained Quest/passive-food와 전체 port goal은 아직 미완료다.

2026-10-10 R61의 `dd8ecbffb`는 통계 포장 node/cursor의 native 포인터, `b14f3f51f`는 압축의 BYTE 비교를 각각 기존 production 본체 하나씩 복원했다. unchanged 원본 envelope155조건·compressor2879조건의 전체 output/길이/순서가 native와 일치하고 원래 WORD/DWORD/run wrap·fault 경계도 유지한다. 일반·HD·실제 strict 전체48179pass/34skip, 지원 server47670pass/7skip과 sanitizer·race/checkptr2·unchanged oracle 및 clean ARM643제품 build/revision/help가 통과했다. allocator/time/RNG 등 제어 서비스와 이전154/155 불일치·red·auditor counter 정정은 [R61 근거와 범위](PORTING-INVENTORY.md#통계-포장-노드의-native-폭과-압축-byte-비교-보존)에 구별한다. 이번 GUI 실행은 없고 기존 Trap Set 아이콘 수정·tests/자산/Save/config/ledger를 보존한다. online bulk caller·retained Quest/passive-food와 전체 port goal은 여전히 미완료다.

2026-10-10 R60의 `2821ad417`에서 통계 패킷 난독화 한 C 본체의 native buffer와 packed DWORD 길이를 복원했다. 원본 packet/mask/CRT 변환 코드의750개 제어 fixture와 전체 output·length·호출 순서가 일치한다. allocation/time/RNG 등은 명시적인 외부 서비스이며 원본 전체 engine/online 통계 실행의 인증은 아니다. 일반·HD·실제 strict 전체48170pass/34skip, 지원 server47661pass/7skip과 sanitizer·race/checkptr2·unchanged oracle 및 clean ARM643제품 build/revision/help가 통과했다. 기존 tests·자산·Save/config·ledger를 보존하고, 아직 ABI32인 bulk caller 및 retained Quest/passive-food는 [별도 미완료 항목](PORTING-INVENTORY.md#통계-패킷-난독화의-native-버퍼와-dword-길이-보존)으로 남긴다. 전체 port goal은 active다.

2026-10-09 R48에서 원본 sound별 동시 event 제한을 복원해 과적 HeavyExertion의 반복 재시작/중첩을 막고, Monster/NPC의 world IMPALE·LAVA 피해 입구를 복원했다. 각 기능은 기존 production 본체 하나씩 수정한 뒤 즉시 push했다. stock 보호/immune/방패·armor carry·Quest 판정은 유지한다. real OpenAL null strict 회귀는 lifecycle만 확인하며 물리 청취로 확대하지 않는다.

기능 revision `ce88d2456`의 공식 ARM64 일반·HD·server3제품 build/clean revision/help와 unchanged stock/code2968/data650/strict NXZ oracle를 통과했다. 최종 전체 일반·HD·실제 cgocheck2는 각각47322pass/34skip, server는46813pass/7skip이다. 일반·HD headless/mock에서8종 환경 피해×Player/Monster/NPC의24개 실제 충돌과 client HP를 각각 확인했다. 초기 fixture의 actual2, 전사의 정상 shield block, spectator 위치 및 원본 정수 marker 기대값 보정은 [R48 근거와 한계](PORTING-INVENTORY.md#과적-효과음-동시-재생-제한과-세-unit-종류의-환경-피해)에 구분했다. 개인 Save/config·원본 assets·기존 ledger/golden을 보존하고 retained Quest/passive-food 및 전체 goal은 완료로 바꾸지 않는다.

2026-10-09 R33에서 Shopkeeper 판매 수량 창의 아이템 아이콘과 주문 습득 시 원본 자기 시전 기본값을 복원했다. 기존 GUI C 본체3개를 각각 별도 커밋으로 수정하고 즉시 push했다. PE32의52byte viewport를 native104byte 그리기 인자로 변환하고, 주문 보상 종류 DWORD를 옆의 입자 수와 함께 pointer 폭으로 읽던 오류를 제거했다. 원본 `0x600` 기본값·공격 주문의 다른 대상 기본값·수동 대상 설정은 유지한다.

clean/pushed 기능 revision `b6c1eecbe`의 공식 ARM64 일반·HD·server3제품 build/revision/help와 모든 Go 패키지 일반/실제 cgocheck2, 지원 server, 표적 highres/race+checkptr2 검사를 통과했다. 일반·HD headless/mock 각3개/총6개 actual exits는0이며, 실제 구매·판매·수리 아이콘6건, 자연 주문 보상 기본값18건, 실제 mouse 대상 변경8건과 재습득 후 설정 보존8건을 확인했다. LIVE inputs6019/제품3/helper2는 불변이다. 초기 fixture·검증 코드 오류와 optional/child skip, 원본 대조의 범위는 [R33 상세 근거](PORTING-INVENTORY.md#상점-판매-아이콘과-주문-자기-시전-기본값-복원)에 구분해 기록했다. 기존 Quest/passive-food 미해결 gate와 전체 port goal은 별도로 active를 유지한다.

2026-10-09 R32의 clean 기능 revision `ef8a3181c`에서 기존 waypoint 도착 판정 한 본체의 연산을 원본 x87 precision53/ToZero 순서로 복원했다. 새80조건의 수정 전32실패/48통과, 수정 후960leaf 통과와 unchanged original206byte의80개 제어 fixture를 구분해 보존했다. 일반/실제 cgocheck2/지원 server의 `. ./server ./legacy`와 highres/race+checkptr2 표적은 actual0이며 unrelated optional/child skip도 기록했다. 원본 map/physics 전체 실행으로 확대하지 않는다.

공식 ARM643제품 clean build/revision/help 및 일반·HD headless/mock 각8개/총16개 실제 종료는0이다. 실제 몬스터264타격·활/석궁3발·Charm PNG/owner 전환·캠페인 대화 잠금/종료 후 이동을 재확인했고 LIVE inputs6007/제품3/helper4는 불변이다. 별도 retained Quest 고정 맵/패시브 음식 검증은 actual2/2, batch1로 여전히 실패한다. 최종 unchanged stock/code2968/data650/strict NXZ 검사를 통과했으며 [R32 상세 근거와 한계](PORTING-INVENTORY.md#ai-웨이포인트-도착-판정의-원본-반올림-복원)를 남겼다. 이전 ledger는 수정하지 않고 전체 port goal을 active로 유지한다.

2026-10-09 후속 clean 기능 revision `29deffcb4`에서 AI push가 재사용 슬롯의 인자를 지우던 문제를 원본 `0050A30F/0050A311`의 action/start-flag 부분 기록으로 복원했다. production 본체는 하나만 변경했다. 새27개 leaf의 red와 수정 후 일반/실제 cgocheck2/지원 server 및 highres/race+checkptr2 검사, 원본 push28개/호출자6개 fixture를 구분해 보존했다. 새 실제 MainAI capture Quest12개·food600개의 original entry replay는 raw24-slot/관찰 scalar 차이0이다. world 결과는 기록값이며 원본 Windows 전체 engine 실행의 증명은 아니다.

공식 ARM643제품 rebuild/revision 검사와 다섯 이슈의 일반·HD headless/mock 각8개/총16개 시나리오가 모두 actual0이다. 몬스터의 실제264타격, Bow/CrossBow3발·탄약20→17, Charm 녹색 효과/owner 전환, 캠페인 상점/대화 잠금과 종료 후 이동을 재확인했다. LIVE 동안6,006 tracked inputs/3 products는 불변이다. 별도 retained Quest 고정 맵/패시브 먹이 소비 scenario는 각각 actual2를 유지하며 성공으로 세지 않는다. conditional optional/child test skip과 원본 대조의 한계도 [상세 기록](PORTING-INVENTORY.md#ai-행동-슬롯-인자-보존과-다섯-이슈의-후속-clean-검증)에 남겼다. 전체 port goal은 active다.

2026-10-09 참 크리처 효과·캠페인 Shopkeeper 이동 잠금·Bow/CrossBow 발사·몬스터의 실제 플레이어 피해와 반복 공격을 원래1.2b 계약에 맞춰 복원했다. clean/pushed 기능 revision `9de4193e5`의 공식 ARM64 일반·HD·server3제품과 전체 일반/실제 cgocheck2, 표적 highres/race+checkptr2 및 지원 server 패키지 검사가 통과했다. 두 client 제품의 headless/mock 순차8개씩 총16실행이 actual0이며 입력6,005개·제품3개의 해시는 전후 불변이다. [상세 원인·실패 이력·검증 범위](PORTING-INVENTORY.md#참-크리처상점-이동활몬스터-피해-및-반복-공격-복원)를 구분해 기록했다.

새 공개 시나리오는 `host-conjurer-charm-effect.yaml`, `host-conjurer-bow.yaml`, `host-conjurer-crossbow.yaml`, `host-game-ai-sustained-attack.yaml`, `solo-conjurer-map-shop-movement.yaml`, `solo-warrior-dialog-movement.yaml`, `host-game-shop-movement.yaml`이다. Charm의 녹색 입자·자연 owner 전환/제거, 실제 inventory/mouse 발사3회와 탄약20→17, 몬스터11종×방향2×lane2의44조건에서 각각3회 선제 타격을 검사한다. poison DOT를 반복 타격으로 계산하지 않으며 적에게 먼저 공격하지 않는다. 캠페인 상점과 대화는30프레임 실제 이동 입력에도 잠기고 Exit/Done 뒤 이동이 재개된다. 일반 멀티플레이 상점의 원래 이동 허용은 보존한다. 기존 콘솔44명령/8text-mode도 두 제품에서 통과했다.

GUI 검사는 아래 기존 headless runner에 해당 YAML을 전달하고 일반/HD를 순차 실행한다. stock 지급·위치/HP·통상 aggression 등의 명시적인 fixture와 실제 게임 결과 관찰은 구분하며, personal Save/config·원본 자산·기존 PNG/golden은 바꾸지 않는다. 물리 SDL/OpenAL·원본 Windows runtime 전체를 검증했다는 주장은 아니다. 이전 fixed-map Quest/passive-retreat-food 미해결 gate와 전체 포팅 목표는 별도로 남아 있다.

비밀벽 observer `7b68b3bb0`의 최종 clean 검증을 완료했다. source/remote revision이 일치하는 ARM64 일반·highres·server 3제품 build/revision/help와 전후 oracle이 통과했다. 일반/HD headless/mock 실제 비밀벽 각 5관찰/합계 10개, Quest key 각 16관찰/합계 32개 및 대검 21·Fireball 11·Magic Missile 11·상태효과 10의 양방향 마법 각 53결과/합계 106개, 총 12실행이 정상 exit 0이다. 종료 후 clean source/remote·code 2,935/data 638개·NXZ 50쌍·stock tree·개인 파일 불변을 확인했다. 아래의 비밀벽 observer clean 후속 대기 기록은 이 결과로 마무리하며, 새 단위는 문서 기록만 갱신하고 runtime/manifest는 변경하지 않는다.

2026-10-04 공개 `host-warrior-secret-wall-touch.yaml`과 읽기 전용 observer로 stock G_Crypts의 닫힌 접촉식 비밀벽을 확인했다. 실제 Warrior host 메뉴·정상 맵 로드·queued 이동으로 접촉하며, player 위치 준비 외에 wall state/flags/delay·CollisionWall·force·오디오·packet·client 위치를 공급하지 않는다. 일반/HD headless/mock 진단 모두 closed 1/0→실제 CollisionWall 기록→opening 4의 delay 1..23→open 3/23, delay>11의 정상 trace 통과, 서버/client drawable의 왕복 통과 및 정확히 한 tile open sound event를 확인했다. 각 5개 관찰/합계 10개와 정상 exit 0이다. native wall/secret/unit/update 포인터는 4GiB를 넘고 원래 연결·좌표·방향·tile·flags·wait를 유지한다.

첫 War01a 진단은 해당 맵의 모든 비밀벽이 flags=0인 스크립트식이라 준비 조건에서 실패했다. stock 맵을 read-only로 확인해 원래 touch flag가 있는 G_Crypts를 선택했으며 맵/플래그나 게임 동작을 바꿔 테스트를 통과시키지 않았다. 결과 오염 감도·geometry·bounded schedule·공개 action 연결·AST 결과 주입 금지 회귀와 일반/실제 cgocheck2/race/checkptr/highres 각 3회, 전체 일반/strict·server-tag·oracle 각 1회도 통과했다. 이 observer 단위의 clean 3제품 및 이전 Quest key/양방향 마법 재검증은 후속 단계이며 dirty 진단과 구별한다. 전체 wall-animation 픽셀·remote client·물리 화면/스피커·원본 Windows runtime을 검증했다는 주장은 아니다. 원본 자산·개인 Save/config·기존 PNG/golden은 불변이다.

2026-10-04 비밀벽 접촉 `00548100`의 PE32 필드 접근을 native `nox_secret_wall_t`로 복원했다. LP64에서 기존 +20/+21/+22는 flags/state/delay가 아니라 wall 포인터 내부였으며, 실제 필드는 +28/+29/+30이다. 한 C 본체만 변경해 원래 플레이어/몬스터 접촉 gate, 닫힘→열리는 중 전환, delay=0, signed DWORD 좌표 wrap과 tile별 open sound 순서를 유지한다. 원본 177바이트·뒤 15-NOP 해시를 구현 전에 확인해 봉인했다. 실제 production 본체·layout/header를 독립 C 실행 파일로 컴파일하는 회귀는 수정 전 닫힘 유지 red와 수정 후 256×256×8 class/flag/state 조합, 반복 무중복, 음수/overflow 좌표 및 고주소 레코드의 인접 필드/포인터 불변을 확인한다. 관련 일반/실제 cgocheck2/race/checkptr/실제 highres 각 3회, 전체 일반/strict·server-tag 각 1회와 oracle이 통과했다. 이 단계는 외부 서비스를 가로채는 본체 계약 검사이며 실제 맵 접촉/GUI는 후속 단계다. code 2,935/data 638개·stock tree·NXZ 50쌍 및 개인 파일은 유지한다.

이전 Quest key observer `d36a7dabd`의 후속 clean 검증도 완료했다. ARM64 3제품 build/revision/help, 일반/HD 실제 Quest key 각 16회/합계 32회와 양방향 마법 8실행/106결과 및 전후 oracle이 통과했고 source/remote revision과 stock 자산은 불변이다. 아래 observer의 clean 검증 대기 기록은 이 결과로 마무리한다.

이 포팅 브랜치의 유일한 지원 도구체인은 `go1.26.9`이다. 버전의 단일 텍스트 기준은 `go-version.txt`이며, `src/go.mod`에는 언어 기준 `go 1.26.0`과 권장 도구체인 `toolchain go1.26.9`를 함께 선언한다. 저장소의 `.go-version`은 goenv의 프로젝트 선택을 1.26.9로 지정하며 전역 선택은 변경하지 않는다. 아래 과거 검증 기록의 1.26.5는 당시 실제 사용 버전으로 보존한다.

빌드와 테스트는 다음 래퍼로 실행한다.

```sh
./scripts/go.sh version
./scripts/go.sh -C src test ./internal/noxbuild
```

Linux 64비트의 native-width CGo 회귀는 `make test-linux-pie`로 root·server·legacy 전체를 시험한다. 기본 비-PIE 실행 파일에서는 C heap이 4GiB 아래에 놓일 수 있어 고주소 포인터 시험의 전제가 성립하지 않을 수 있다. PIE 시험은 그 전제를 유지하는 별도 검증 게이트이며, 일반 제품의 빌드·실행 검증을 대체하지 않는다.

2026-10-04 player self-report의 장비 인챈트 바이트 보고 누락을 복원했다. 원본 `004D99A7..004D99E1`은 진입 시의 update 포인터에서 Player를 읽어 마지막 보고 바이트와 object `+440`의 하위 바이트를 비교하고, 변경 시 `004D8840`의 reliable `5B <item-enchantment-byte>` 패킷을 보낸다. 전송 실패도 결과를 무시하며, 전송 뒤 같은 update 포인터에서 Player와 해당 바이트를 다시 읽어 캐시를 갱신한다. `+440`은 native `Field110`이며, 중독 강도 `Poison540` (`+540`)과는 무관하다. 최초 연결의 필드 오해를 headless 실행에서 발견해 바로잡았다. 새 high-C-owned root 회귀는 65,536개 바이트 비교 조합, 256개 unsigned 수신자, 골드 콜백 이후 재로딩, 전송 중 update/Player/마스크가 바뀌는 경우 및 실패 후 캐시 갱신을 검사한다. 상위 24비트만 바뀌거나 중독이 적용되었을 때 잘못 보고하지 않고, 원본 마스크와 중독 상태·타이머도 보존하는 독립 회귀를 추가했다. 원본 코드 manifest에는 보고 slice·sender·3바이트 padding의 해시 범위만 추가했다. 이 단계에서 남아 있던 Quest key 누락은 아래의 별도 단위로 복원했다.

Quest 은·금 열쇠 보고 `004D9A3F..004D9B48`도 native self-report에 연결했다. 생명 loop 앞의 단일 Quest admission과 entry update를 유지하며, 각 32개 슬롯의 zero type-cache lookup/publish, unitless recipient 포함, 후보마다 DWORD cache→WORD type 비교, 전송 실패 뒤 precomputed presence 저장을 복원한다. reliable `F0/22`·`F0/23` 패킷의 raw NetCode 하위 WORD와 signed recipient/result는 유지한다. 고주소 C-owned 회귀와 일반/실제 cgocheck2/race/checkptr/HD 각 3회, 전체 일반/strict·server-tag 검사 및 code 2,933/data 638개·NXZ 50쌍·stock tree 대조가 통과했다. 복원 커밋 `4566ad0b3`의 clean ARM64 3제품 빌드·revision 검증·도움말과 일반/HD 양방향 마법 8실행/106결과도 통과했다. 기존 raw C 본체·원본 자산·개인 Save/config·PNG 기대값은 변경하지 않는다.

후속 `host-quest-key-report.yaml`은 실제 Quest 메뉴, 정상 stock key 지급, queued stack 수량 확인·금지된 수동 드롭 시도와 정상 stage 전환을 검사한다. 읽기 전용 observer는 실제 server/client inventory 수량·self-report cache·기존 C decoder의 native Player tail 수신 바이트를 대조한다. 중복 2개도 presence=1이며, 진행 중 수량 대화창/drag는 성공으로 세지 않는다. 첫 두 진단의 수량 확인 누락과 “전환 시 모든 열쇠 삭제”라는 잘못된 기대값을 바로잡았고 game 동작은 바꾸지 않았다. 최종 일반/HD 진단은 각 16회, 합계 32회 관찰이 통과했으며 관련 강화 검사 9종도 통과했다. 이 observer 커밋의 clean 제품 재검증은 별도 단계다. 실제 소비 후 소실, remote client·key rank 픽셀·물리 화면·Windows runtime 전체 검증으로 확대하지 않는다.

같은 보고 경로가 사용하는 장비 아이콘 `00413420`의 이름·이미지 캐시 조회를 native pointer side slot에 연결했다. 공용 `memmap.PtrPtr` 저장소는 이미 안전했지만 기존 C 본체는 packed PE32 슬롯을 직접 8바이트로 읽고 이미지 반환을 DWORD로 축소했다. 반환 ABI와 호출자 변수 타입은 별도 커밋으로 먼저 복원했다. 이번 한 본체 변경은 여섯 이미지의 순서대로 lazy load, 마지막 loaded=1 게시, live flag byte의 exact match 및 nil 반환을 보존한다. 모든 256개 입력·nonzero loaded 값·nil loader/reset·로드 후 flag 변경·4GiB 초과 C-owned 이름/이미지와 packed 120바이트 불변을 실제 C 진입점까지 검사했다. 원본 body 91바이트·뒤 5-NOP·6행 테이블과 여섯 이름의 해시만 oracle에 더했으며 원본 자산은 공개하지 않는다. Tooltip 조회와 실제 inventory hover는 아직 별도 검증 단계다.

장비 효과 tooltip `00413480`도 같은 side slot의 native 키 포인터로 복원했다. 한 C 본체 변경이며 six-row live byte 비교·첫 exact match의 late key read·원본 `Modifier.c`/line 2087·null variant output·매 호출 string lookup 및 nil 반환을 유지한다. 원본 본체 61바이트·뒤 3-NOP·여섯 문자열 키와 source path의 해시를 따로 봉인했다. 모든 256개 byte, 조회 순서/콜백 사이 flag·key 변경/중복 first-wins, C 진입점의 4GiB 초과 C-owned UTF-16 반환 및 packed table/아이콘 loaded 상태 불변을 회귀로 검사한다. GUI hover 관찰과 Windows/물리 화면 전체의 검증은 별도다.

공개 `host-game-player-item-enchantment-report.yaml`은 stock `Sword`/`FireProtect1`의 정상 inventory grant 후 실제 mouse/`MSG_TRY_EQUIP`·`MSG_TRY_DEQUIP`을 두 번 반복한다. observer는 server `Field110`·report cache·opcode 91의 실제 client 수신값, 정상 inventory draw가 채운 4GiB 초과 image handle, 실제 hover의 한국어 `화염 저항` 표시와 해제 후 빈 tooltip을 읽기만 한다. 장비·효과 mask·packet·tooltip·draw 결과를 주입하지 않으며 결과 변조 감도·bounded schedule·공개 action 연결 및 observer의 금지된 writes/calls를 별도 단위 테스트로 검사한다. headless/mock audio 검증이며 원본 PNG golden이나 개인 save/config는 변경하지 않는다.

Windows PowerShell에서는 다음과 같이 실행한다.

```powershell
.\scripts\go.ps1 version
.\scripts\go.ps1 -C src test ./internal/noxbuild
```

실행할 제품이 현재 clean source revision과 정확히 일치하는지는 `noxbuild -verify`로 확인한다. 이 검사는 ELF, PE, Mach-O를 호스트 종류와 무관하게 읽어 Go 1.26.9, 지원 대상 tuple, 전체 `vcs.revision`, `vcs.modified=false`를 강제한다. 특히 무시된 `build/` 아래에 남은 구 실행 파일을 다시 실행하는 일을 차단한다.

```sh
./scripts/go.sh -C src run ./internal/noxbuild \
  -go=../scripts/go.sh -verify \
  ../build/linux-amd64/opennox ../build/linux-amd64/opennox-server
```

```powershell
.\scripts\go.ps1 -C src run ./internal/noxbuild `
  -go=go -verify `
  ..\build\windows-amd64\opennox.exe ..\build\windows-amd64\opennox-server.exe
```

래퍼는 외부 `GOROOT`를 제거하고 `GOTOOLCHAIN=go1.26.9`와 빈 `GOEXPERIMENT`를 강제한 뒤, 실제 `GOVERSION`이 정확히 일치하는지 검사한다. `GO`를 명시하면 그 실행 파일을 우선 사용한다. 그렇지 않으면 `goenv prefix 1.26.9`로 이미 설치된 실행 파일을 찾고, 사용할 수 없을 때만 PATH의 `go`로 돌아간다. `goenv global/local/install`을 실행하거나 전역 선택을 변경하지 않는다. `GO`에는 인자 없이 Go 실행 파일 경로 하나만 지정하며 공백이 있는 경로도 지원한다. `internal/noxbuild`도 자신을 컴파일한 Go와 자식 빌드에 쓰는 Go를 각각 검사한다.

따라서 `GOEXPERIMENT=cgocheck2 ./scripts/go.sh ...`는 strict CGo 검증이 아니다. POSIX 환경의 테스트 전용 실행 경로는 고정 도구체인을 먼저 찾은 뒤 `cgocheck2`와 `CGO_ENABLED=1`의 실제 적용을 확인한다. 릴리스 빌드 래퍼의 정책은 바꾸지 않으며, 테스트 인수를 생략하면 전체 패키지를 실행한다.

```sh
sh ./scripts/test-cgocheck2.sh
sh ./scripts/test-cgocheck2.sh -run TestNative -count=3 . ./legacy ./server
```

macOS/ARM64의 GUI 회귀는 `NOX_E2E_SEAT=headless`로 실행한다. 다음 통합 시나리오는 Con01a 곰 전투, Con02a 네크로맨서 소환·공격·퇴장, 늑대 Charm 60회와 Henrick 추종, 이후 ESC 메뉴 입력을 같은 세션에서 검증한다. 맵 trigger로 플레이어 위치만 준비하며 대화 버튼은 실제 mouse 입력으로 누른다. 일반·HD 제품 모두 통과했으며, 네크로맨서 소환 시 frozen 상태와 이벤트 종료 후 조작/cinematic 복귀를 구분한다. 이 검증은 SDL 화면 출력이나 실제 OpenAL 재생 검증을 대신하지 않는다.

```sh
NOX_E2E_SEAT=headless bash ./scripts/run-headless-gui-e2e.sh \
  /absolute/path/to/nox ./scripts/e2e/solo-conjurer-headless.yaml
NOX_E2E_SEAT=headless NOX_E2E_CLIENT_TARGET=client-hd \
  bash ./scripts/run-headless-gui-e2e.sh /absolute/path/to/nox \
  ./scripts/e2e/solo-conjurer-headless.yaml
```

`host-game-flame-monster-food.yaml`도 일반·HD headless 실행에서 player HP `150→148`, Spider의 RedApple `40→45`, Meat `40→50`, Mushroom 독 `4→0`과 소비된 필드 오브젝트 제거를 확인했다. 휴면 tutorial NPC/Wolf의 원본 `Cur=Max=0`은 사망 상태가 아니며, 일반 몬스터의 food/retreat AI와 구분한다. 통합 이벤트의 종료가 통과하더라도 미이식 피해 분기는 별도로 검사해야 한다.

`host-game-fist-unit-damage.yaml`은 실제 Wizard host 메뉴 시작 뒤 정상 NoxScript object-to-position API로 Fist 1..5레벨을 플레이어→stock NPC 및 NPC→플레이어 방향으로 시전한다. 두 유닛의 원래 `PlayerDamage` callback, 자연 충돌, 서버 HP/피격 marker/귀속, fractional carry, 실제 client 피해 표시와 발사체 제거를 검사하며 일반·HD headless 각각 10회, 합계 20회가 exit 0이다. 원본 피해 `50/100/200/300/400`, Wizard의 시작 armor `0.23000002` 및 연속 시전의 소수 누적을 유지하고 NPC→플레이어 실효 피해 `44/89/177/265/354`를 확인했다. 모든 시전에서 중복 거부, 18틱의 명중, 36틱의 world/owned/client drawable 제거를 관찰하며 native object/update 포인터는 4GiB를 넘는다. 위치·검증용 HP 2,000·NPC 대기 AI만 fixture이고 피해 결과·패킷·물리 상태를 주입하지 않는다. 자율 NPC 주문 선택, incantation/mana, 유닛 사망·campaign trigger·다른 모든 피해 종류나 Linux/Windows/SDL/OpenAL 검증을 대신하지 않는다.

기존 배럴 Fist 제어 시나리오의 첫 historical PNG는 앞선 색상 복원 이후 불일치한다. 이번 피해 수정 전 `e882c931b`의 별도 clean worktree에서도 같은 불일치를 확인했고 현재 `_got` PNG와 byte-identical이었다. 기존 공개 기준 PNG를 바꾸거나 override하지 않았다. 피해 수정 전 worktree에서 얻은 private 비교 프레임 10개를 기준으로 일반·HD의 배럴 5레벨 피해/제거 및 RGBA 비교를 모두 통과했으며 이를 historical PNG 통과로 세지 않는다. 원본 자산·개인 Save/config와 private 참조 이미지는 저장소에 추가하지 않는다. 다른 GUI E2E와 동시에 실행하지 않는다.

```sh
NOX_E2E_SEAT=headless bash ./scripts/run-headless-gui-e2e.sh \
  /absolute/path/to/nox ./scripts/e2e/host-game-fist-unit-damage.yaml
NOX_E2E_SEAT=headless NOX_E2E_CLIENT_TARGET=client-hd \
  bash ./scripts/run-headless-gui-e2e.sh /absolute/path/to/nox \
  ./scripts/e2e/host-game-fist-unit-damage.yaml
```

`host-game-push-spell.yaml`은 실제 Wizard host 시작 뒤 플레이어의 정상 script cast 1..5레벨과 stock NPC의 `AIActionCastOnObj`→자연 cast-animation frame→`CastSpellByUser4FDD20`→Push selector 경로를 검사한다. 일반·HD headless 각각 6회가 exit 0이며, NPC는 원본 mode-selected power 3으로 11틱 후 animation frame 4에서 시전하고 12틱 후 행동 종료와 서버·client 이동을 확인했다. 원본 `PushPowerCoeff=15`, 반경 600/10, 거리 감쇠와 대상 mass를 유지하며 정확한 서버 force, 실제 packet을 받은 drawable 이동, 1회 cast audio event와 HP 불변을 확인한다. native 시전자·대상·NPC update 포인터는 4GiB를 넘는다. 기존 6-int C 진입점의 시전자 절단 및 PE32 `+56` 위치 참조를 실제 public wrapper의 실패 회귀로 재현했고, native `PosVec`(wide host `+60`)를 쓰는 Go cast로 복원했다. 원본 executable/자산은 검사 전후 동일하다. 위치·대기 AI·명시적인 주문/대상 선택만 fixture이며 force·cast frame·피해·이동 packet·client 위치를 주입하지 않는다. 자율 NPC 주문 선택, 플레이어 incantation/mana, 특정 campaign trigger 또는 Linux/Windows/SDL/OpenAL 검증을 대신하지 않는다. 다른 GUI E2E와 동시에 실행하지 않는다.

```sh
NOX_E2E_SEAT=headless bash ./scripts/run-headless-gui-e2e.sh \
  /absolute/path/to/nox ./scripts/e2e/host-game-push-spell.yaml
NOX_E2E_SEAT=headless NOX_E2E_CLIENT_TARGET=client-hd \
  bash ./scripts/run-headless-gui-e2e.sh /absolute/path/to/nox \
  ./scripts/e2e/host-game-push-spell.yaml
```

`host-game-pull-spell.yaml`은 같은 실제 Wizard host 시작과 정상 script cast 1..5레벨, stock NPC의 자연 `AIActionCastOnObj` 시전 경로에서 Pull을 검사한다. 일반·HD headless 각각 6회, 합계 12회가 exit 0이다. 원본 `PullPowerCoeff=15`, 반경 600/10, 거리 감쇠와 대상 mass를 유지하며, 서버의 정확한 음수 force와 실제 이동 packet을 받은 client drawable의 **안쪽 이동**을 확인했다. NPC는 원본 mode-selected power 3으로 11틱 후 animation frame 4에서 시전하고 12틱 후 행동 종료와 양쪽 이동을 확인했다. 모든 시전에서 cast audio event는 1회이며 NPC HP 150과 Wizard HP 75는 변하지 않는다. 4GiB 초과 native 포인터를 실제 공개 Pull wrapper에 넘기면 기존 6-int C callee의 시전자 절단 및 PE32 `+56` 접근으로 크래시가 발생하는 실패 회귀를 먼저 재현했고, wide host의 실제 `PosVec`(offset `+60`)를 사용하는 Go cast로 복원했다. 원본의 signed DWORD level 곱셈→FCHS 부호 반전→binary32 spill 순서, 음수 0과 세 번째 오브젝트의 cast audio 귀속은 별도 함수 회귀에서 검사한다. 기존 Push 시나리오도 일반·HD 각각 6회 재실행해 바깥쪽 이동이 유지됨을 확인했으며 원본 executable/자산은 검사 전후 동일하다. 위치·대기 AI·명시적인 주문/대상 선택만 fixture이고 force 결과·cast frame·HP·이동 packet·client 위치는 주입하지 않는다. 자율 NPC 주문 선택, 플레이어 incantation/mana, campaign trigger 또는 Linux/Windows/SDL/OpenAL 검증을 대신하지 않는다. 다른 GUI E2E와 동시에 실행하지 않는다.

```sh
NOX_E2E_SEAT=headless bash ./scripts/run-headless-gui-e2e.sh \
  /absolute/path/to/nox ./scripts/e2e/host-game-pull-spell.yaml
NOX_E2E_SEAT=headless NOX_E2E_CLIENT_TARGET=client-hd \
  bash ./scripts/run-headless-gui-e2e.sh /absolute/path/to/nox \
  ./scripts/e2e/host-game-pull-spell.yaml
```

`host-game-fumble-spell.yaml`은 실제 Warrior host 메뉴 시작 뒤 플레이어→stock NPC의 정상 script object cast 요청 1..5, 플레이어→일반 Troll의 drop-all 분기, stock NPC→플레이어의 자연 cast-animation 경로를 검사한다. 일반·HD headless 각각 7회, 합계 14회가 exit 0이다. 원본 targeted Magic projectile은 요청 level과 별개로 mode-selected power 3을 유지하며 Fumble 효과는 level을 사용하지 않는다. 장착 무기·wand·방패만 떨어뜨리고 몸통 방어구와 미장착 예비 장비를 유지하는 판정은 함수 회귀로 고정했고, 실제 NPC/플레이어의 무기·방패 드롭 및 client 장비 mask 동기화도 확인했다. 일반 몬스터는 인벤토리 두 개가 모두 떨어지고, 최초 시전자가 아닌 **명중한 Magic의 위치**를 기준으로 원본 force 50과 mass에 따른 정확한 힘 및 서버·client의 바깥쪽 이동을 확인한다. NPC HP 150, Troll HP 80, Warrior HP 150은 변하지 않는다. NPC는 11틱 후 animation frame 4에서 자연 시전하고 20틱 후 명중, 22틱 후 행동 종료·투사체 제거·client replay가 완료됐다. native unit/item/projectile/update 포인터는 4GiB를 넘는다. 모든 드롭의 실제 서버 상태와 시야 안 드롭의 client drawable을 검사하며, 원본의 무작위 reachable 위치가 벽 뒤인 드롭은 표시 결과 검증에서 구분한다. 위치·대기 AI·정상 아이템 지급/장착·명시적인 주문/대상 선택만 fixture이고 드롭·장비 flag·힘·cast frame·HP·투사체·client 결과를 주입하지 않는다. 원본 executable/자산은 검사 전후 동일하다. GameBall/Shopkeeper 분기는 별도 함수 회귀 범위이며, 자율 NPC 주문 선택·플레이어 incantation/mana·campaign trigger·Linux/Windows/SDL/OpenAL 검증을 대신하지 않는다. 다른 GUI E2E와 동시에 실행하지 않는다.

```sh
NOX_E2E_SEAT=headless bash ./scripts/run-headless-gui-e2e.sh \
  /absolute/path/to/nox ./scripts/e2e/host-game-fumble-spell.yaml
NOX_E2E_SEAT=headless NOX_E2E_CLIENT_TARGET=client-hd \
  bash ./scripts/run-headless-gui-e2e.sh /absolute/path/to/nox \
  ./scripts/e2e/host-game-fumble-spell.yaml
```

`host-game-player-status-animation.yaml`은 실제 메뉴로 시작한 regular host에서 정상 enchant API로 제자리 플레이어에게 Held(스턴)·Confused·AntiMagic·Charming·Shield·Slowed를 각각 90틱 적용한다. 초기 무적은 강제로 제거하지 않고 자연 만료를 기다린다. 일반·HD headless와 일반 독립 재실행에서 실제 클라이언트 패킷의 buff/HUD 동기화, 두 시점의 원본 효과 sprite 픽셀 및 서로 다른 애니메이션 프레임, 자연 만료 후 제거를 확인했다. Slowed는 실제 화면 안의 YellowBubbleParticle 생성과 만료 후 0개를 검사한다. 효과별 시작·진행·해제의 18개 PNG도 모든 실행에서 일치했다. 원래 로더의 packed DWORD 저장으로 native 애니메이션 cache가 nil이 되는 실패를 먼저 재현했으며, cache 초기화와 세션 정리를 별도 함수 단위로 복원했다. 상태 준비만 fixture이고 client buff·화면 출력·만료를 주입하지 않는다. Stun 주문의 조건별 Held/Slowed 선택, 적의 주문 명중·이동 제어, 다른 상태 종류·캐릭터 자세나 SDL/OpenAL 검증을 대신하지 않는다. 다른 GUI E2E와 동시에 실행하지 않는다.

```sh
NOX_E2E_SEAT=headless bash ./scripts/run-headless-gui-e2e.sh \
  /absolute/path/to/nox ./scripts/e2e/host-game-player-status-animation.yaml
NOX_E2E_SEAT=headless NOX_E2E_CLIENT_TARGET=client-hd \
  bash ./scripts/run-headless-gui-e2e.sh /absolute/path/to/nox \
  ./scripts/e2e/host-game-player-status-animation.yaml
```

`host-game-player-status-spell.yaml`은 실제 Wizard 메뉴 선택 뒤 level 1 Confuse와 Stun을 self-target으로 시전한다. `host-game-player-status-spell-warrior.yaml`은 실제 Warrior 메뉴로 Stun의 Slowed 분기를 확인한다. 정상 `SpellAccept4FD400` selector→6-argument C cast 진입→buff 적용→실제 client packet/HUD→그리기→자연 만료 경로이며, enchant를 직접 적용하는 앞 시나리오와 구분한다. 원본 Confuse 90틱·Stun 60틱과 class 선택을 바꾸지 않는다. Wizard의 Held 머리 위 효과는 두 서로 다른 sprite frame에서 픽셀이 일치했고, Warrior는 원본대로 Held가 아닌 Slowed의 노란 입자가 두 시점에서 7개·6개 표시된 뒤 사라졌다. 4GiB 초과 native player/argument 포인터를 유지하며, C-heap argument를 명시적으로 초기화하는 회귀도 검사한다.

일반·HD headless 네 실행에서 시작·진행·해제의 새 PNG 9개가 decoded RGBA 기준으로 정확히 일치했다. 공유 시각 검사 helper 변경 후 기존 6종 상태의 PNG 18개도 일반 회귀 실행에서 그대로 일치했다. self-target server cast 준비만 fixture이며 적의 주문 명중·플레이어 incantation 입력·mana 소비, 모든 캐릭터 자세나 SDL/OpenAL 출력을 대신하지 않는다. Conjurer/기타 class byte와 Monster Mass 분기는 별도 native 함수 회귀 범위다. 다른 GUI E2E와 동시에 실행하지 않는다.

```sh
NOX_E2E_SEAT=headless bash ./scripts/run-headless-gui-e2e.sh \
  /absolute/path/to/nox ./scripts/e2e/host-game-player-status-spell.yaml
NOX_E2E_SEAT=headless NOX_E2E_CLIENT_TARGET=client-hd \
  bash ./scripts/run-headless-gui-e2e.sh /absolute/path/to/nox \
  ./scripts/e2e/host-game-player-status-spell.yaml
NOX_E2E_SEAT=headless bash ./scripts/run-headless-gui-e2e.sh \
  /absolute/path/to/nox ./scripts/e2e/host-game-player-status-spell-warrior.yaml
NOX_E2E_SEAT=headless NOX_E2E_CLIENT_TARGET=client-hd \
  bash ./scripts/run-headless-gui-e2e.sh /absolute/path/to/nox \
  ./scripts/e2e/host-game-player-status-spell-warrior.yaml
```

`host-game-warrior-abilities.yaml`은 regular host의 정상 초기화로 습득한 Warrior 스킬 5개를 실제 A/S/D/F/G 키로 각각 두 번 사용한다. 돌진 피해·범위 내 함성의 anti-magic/stun·작살의 피해와 끌어오기·가볍게 걷기의 은신 유지 및 이동 중 공격으로 해제·늑대의 눈의 투명 적 감지를 확인한다. 각 사용의 서버 상태·실제 패킷의 HUD·종료·쿨다운 중 재입력 거부·ready 보고·재사용을 검사하며 일반·HD headless 실행 모두 통과했다. Troll HP 2,000, 정지 AI 대상·위치·투명 Spider와 quickbar 배치는 명시적인 fixture다. 스킬 실행·종료·쿨다운을 직접 주입하지 않는다. 원본 HarpoonDuration=0과 TreadLightlyDuration=99,999는 변경하지 않으며, 벽 돌진·PvP·campaign의 레벨별 검증이나 SDL/OpenAL 검증을 대신하지 않는다. 다른 GUI E2E와 동시에 실행하지 않는다.

```sh
NOX_E2E_SEAT=headless bash ./scripts/run-headless-gui-e2e.sh \
  /absolute/path/to/nox ./scripts/e2e/host-game-warrior-abilities.yaml
NOX_E2E_SEAT=headless NOX_E2E_CLIENT_TARGET=client-hd \
  bash ./scripts/run-headless-gui-e2e.sh /absolute/path/to/nox \
  ./scripts/e2e/host-game-warrior-abilities.yaml
```

`host-game-warrior-charge-collisions.yaml`은 실제 A 키 입력으로 서버 플레이어 fixture와 기존 stock 벽에 각각 두 번 돌진한다. 일반·HD headless에서 플레이어 HP `2,000→1,850`, 벽 충돌의 자기 피해·Held·종료·쿨다운 재입력 거부·HUD ready·재사용을 확인했다. 정상 join/관전자 종료로 초기화한 두 번째 서버 플레이어의 HP·배치와 quickbar만 준비하며 능력·collision callback·피해·CollisionWall·buff 종료·쿨다운을 주입하지 않는다. 원격 클라이언트의 네트워크 입력 대결이나 FlagBall GameBall drop을 대신하지 않으며, 다른 GUI E2E와 동시에 실행하지 않는다.

```sh
NOX_E2E_SEAT=headless bash ./scripts/run-headless-gui-e2e.sh \
  /absolute/path/to/nox ./scripts/e2e/host-game-warrior-charge-collisions.yaml
NOX_E2E_SEAT=headless NOX_E2E_CLIENT_TARGET=client-hd \
  bash ./scripts/run-headless-gui-e2e.sh /absolute/path/to/nox \
  ./scripts/e2e/host-game-warrior-charge-collisions.yaml
```

`solo-conjurer-chapter2-field-guide-shop.yaml`은 일반·HD headless 클라이언트에서 Con02a의 실제 Mystic 상점 열기, Wasp 책 구매·인벤토리 사용·습득 알림, 상점 재입장과 세션 해제를 검증한다. 구매 자금 10,000과 Urchin 보상은 명시적인 fixture이며, Wasp는 실제 맵 상점 정의에서 생성된다. 실제 마우스 입력과 서버·클라이언트 패킷 처리로 Wasp 가격 100, 잔액 `10,000→9,900`, 재고 `8→7`, 책 소비 `1→0`, 습득 레벨 `0→1`을 확인하고 재입장 후에도 품절을 검사한다. 이 시나리오도 다른 게임 E2E와 동시에 실행하지 않는다.

`solo-conjurer-pet-autosave-load.yaml`은 seed 프로세스에서 실제 주문으로 만든 Wolf 1마리·Urchin 2마리와 Con02a AUTOSAVE를 준비하고 정상 종료한 뒤, 새 프로세스의 실제 메뉴에서 로드한다. 두 일반·HD headless 실행에서 script ID·HP·owner·HUD·minimap·client drawable, 이어지는 Con03a 출구 전환과 player 이동을 확인했다. 기대값 JSON은 격리된 save 루트의 저장 슬롯 밖에 있으며 소환수를 생성·복원하지 않는다. Pixie 저장·Quest 단계·개인 기존 save는 이 시나리오의 범위가 아니다. seed와 load YAML을 같은 디렉터리에 두고 다른 GUI E2E와 동시에 실행하지 않는다.

```sh
NOX_E2E_SEAT=headless bash ./scripts/run-headless-gui-e2e.sh \
  /absolute/path/to/nox ./scripts/e2e/solo-conjurer-pet-autosave-load.yaml
```

`host-quest-twenty-stages.yaml`은 실제 Warrior Quest 메뉴부터 stock 출구로 1→20단계를 진행한다. 각 단계의 지도·playable player·native 생성기 최대 수, 전환 후 실제 이동 19회, 5/20단계 미니언의 HP와 inventory 연결을 검사한다. 일반·HD headless 모두 제품당 생성기 2,596개 기록과 20단계 강화 경계의 High `8→16`·Normal `5→10`·Low `2→4`·Singular `1→1`을 확인하고 정상 종료했다. player의 출구 배치만 fixture이며 stage·spawn·HP·생성기 cap·UI 해제·이동은 주입하지 않는다. 생성기의 실제 최대 spawn 개체 수·boss 전투/drop·온라인 Quest 소환수 보존·20단계 이후나 SDL/OpenAL 검증은 별도다. 다른 GUI E2E와 동시에 실행하지 않는다.

```sh
NOX_E2E_SEAT=headless bash ./scripts/run-headless-gui-e2e.sh \
  /absolute/path/to/nox ./scripts/e2e/host-quest-twenty-stages.yaml
NOX_E2E_SEAT=headless NOX_E2E_CLIENT_TARGET=client-hd \
  bash ./scripts/run-headless-gui-e2e.sh /absolute/path/to/nox \
  ./scripts/e2e/host-quest-twenty-stages.yaml
```

`host-quest-minion-combat.yaml`은 실제 Quest 메뉴와 네 번의 stock 출구 전환으로 5단계 네크로맨서를 생성한 뒤, 플레이어만 근처에 배치한다. 일반·HD headless에서 원래 AI가 2틱 뒤 적을 감지하고 실제 전기 피해로 player HP `450→397`과 client 피해 표시를 만들었다. 실제 mouse 추적·이동·공격으로 Necromancer HP `100→0`, 사망과 원래 inventory의 CommonSpellBook 필드 방출을 확인했다. 몬스터 생성·AI·aggression·HP·damage·death·보상을 주입하지 않으며 일반 독립 재실행에서도 두 전투 PNG와 기존 단계 결과 PNG가 일치했다. 이는 해당 네크로맨서의 전투 검증이며 Hecubah나 모든 피해 종류, 주변 FlyingGolem의 별도 PIERCE shape guard, 온라인 대전 검증을 포함하지 않는다.

```sh
NOX_E2E_SEAT=headless bash ./scripts/run-headless-gui-e2e.sh \
  /absolute/path/to/nox ./scripts/e2e/host-quest-minion-combat.yaml
NOX_E2E_SEAT=headless NOX_E2E_CLIENT_TARGET=client-hd \
  bash ./scripts/run-headless-gui-e2e.sh /absolute/path/to/nox \
  ./scripts/e2e/host-quest-minion-combat.yaml
```

`solo-warrior-purchased-wolf-transition.yaml`은 War01a 원본 세 NPC 이벤트를 끝내고 stock War03b를 로드한 다음, Henrick의 실제 Yes/Done 입력 두 번으로 원래 Wolf1/Wolf2를 구매한다. 일반·HD headless에서 stock 가격 200, 골드 `400→200→0`, 서로 독립적인 Migrate/Monitor bit를 확인하고 War03b→새 War03c→저장된 War03b를 stock 출구 collision으로 왕복했다. 전환 직후와 추가 240틱 뒤의 네 assertion에서 같은 두 객체·script ID·wire ID·HP `40/40`·host owner·client drawable과 player 근처 복원을 확인하며 중복 생성도 검사한다. 네 PNG는 구매 대화와 두 목적지 화면을 기록한다. 초기 War03b 로드, 구매 자금 400, stock StartDialog 서비스 호출과 player의 출구 배치만 fixture다. 늑대 생성·재소유·HP/AI/migration flag·출구 callback을 직접 주입하지 않는다. 전체 campaign 도보 진행·다른 챕터·온라인 Quest의 소환수 보존이나 SDL/OpenAL 검증은 별도이며 다른 GUI E2E와 동시에 실행하지 않는다.

```sh
NOX_E2E_SEAT=headless bash ./scripts/run-headless-gui-e2e.sh \
  /absolute/path/to/nox ./scripts/e2e/solo-warrior-purchased-wolf-transition.yaml
NOX_E2E_SEAT=headless NOX_E2E_CLIENT_TARGET=client-hd \
  bash ./scripts/run-headless-gui-e2e.sh /absolute/path/to/nox \
  ./scripts/e2e/solo-warrior-purchased-wolf-transition.yaml
```

독립 서버의 실제 기동·맵 전환·정상 종료는 다음 headless 회귀로 확인한다. `curl`·`jq`가 필요하며, 고정 도구체인으로 `server` 제품을 빌드한 뒤 개인 save/config를 제외한 임시 데이터 뷰에서 실행한다. E2E 모드로 NAT 포워딩과 공개 서버 등록을 차단하고, 인증된 로컬 API로 `so_beach→estate→trilevel→estate`를 로드한다. 각 맵의 mode·player 수, 실제 game tick 증가와 map load 횟수를 확인하고 console `quit` 이후 10초 안에 프로세스가 exit 0으로 끝나야 통과한다. 원본 데이터는 변경하지 않고 출력 디렉터리에 로그·임시 런타임 데이터를 남긴다. 기본 게임/API 포트는 18610, metrics 포트는 6062이며 다른 서버 E2E와 동시에 쓰지 않는다. `NOX_E2E_SERVER_PORT`·`NOX_E2E_SERVER_METRICS_PORT`로 바꿀 수 있다. 이미 빌드한 제품을 검사할 때에는 절대 경로의 `NOX_E2E_SERVER_BINARY`를 지정한다. 이 검증은 접속한 원격 플레이어의 게임플레이·공개 서버 발견·Quest 소환수 보존을 대신하지 않는다.

이 검사 실행기는 초기 so_beach와 이후 API 맵 전환을 분리하기 위해 `-autoexec ""`를 명시한다. 일반 전용 서버에서 옵션을 생략하면 기존 CLI 기본 `load estate`가 실행된다. 검사 편의를 위해 게임의 기본 명령이나 기존 판정값을 변경하지 않는다.

```sh
bash ./scripts/run-headless-server-e2e.sh \
  /absolute/path/to/nox /absolute/path/to/server-e2e-output
```

`host-quest-player-death.yaml`은 실제 Quest 메뉴와 네 번의 stock 출구 전환 뒤 자연 전투로 Warrior를 세 번 사망시킨다. 일반·HD headless와 일반 새 프로세스 재실행에서 서버 목숨 `2→1→0→2`, death counter `0→1→2→0`, HP `450→0→450`, 같은 native player identity, 마지막 사망의 결과창과 실제 Continue 입력 부활을 확인했다. 20 ordinary dead frame 뒤 native host PlayerInd 31과 실제 timer child 10712의 empty StaticText data를 관찰하고 Continue 직전 화면도 캡처한다. player placement만 fixture이며 timer/text event·GUI 값·HP·AI·목숨·통계·사망·페널티·부활은 주입하지 않는다. PNG 8개가 일반·HD에서 바이트 단위로 같고 독립 재실행 baseline도 통과했다.

최종 발사체는 제거되었을 수 있어 독점적 lethal attribution은 단정하지 않는다. 실제 전투 골드와 generator/monster/secret counter가 0인 한계, ankh HUD `X 0`의 별도 미완료 경계, remote non-host 결과창·SDL/OpenAL·부활 후 새 이동·다른 캐릭터/피격·온라인 Quest 소환수 보존 범위를 구분한다. host countdown은 `0049B6E0` 복원 뒤 empty text와 실제 화면으로 후속 검증했다. 원본 자산과 개인 Save/config는 변경하지 않는다. PNG가 저장소에 생성되지 않도록 YAML을 임시 시나리오 디렉터리에 복사해 실행하고 다른 GUI E2E와 동시에 실행하지 않는다.

```sh
NOX_E2E_SEAT=headless bash ./scripts/run-headless-gui-e2e.sh \
  /absolute/path/to/nox /absolute/path/to/temp-e2e/host-quest-player-death.yaml
NOX_E2E_SEAT=headless NOX_E2E_CLIENT_TARGET=client-hd \
  bash ./scripts/run-headless-gui-e2e.sh /absolute/path/to/nox \
  /absolute/path/to/temp-e2e/host-quest-player-death.yaml
```

옵션 audit 시나리오 `main-menu-options-audit.yaml`과 `host-game-client-options-audit.yaml`은 일반·HD headless의 실제 입력, native 설정값과 직렬화 결과를 점검한다. 초기 실패와 함수별 보정 이력은 [포팅 인벤토리](PORTING-INVENTORY.md)에 보존한다. 최신 private 회귀의 일반/HD main은 226/228개, 게임 내는 201/203개 assertion 통과·실패 0·정상 exit 0이며 기존 PNG 18개를 override 없이 비교했다. 이는 반복 assertion 수이지 고유 기능 수가 아니다. checkbox·음량/mute·해상도 pending 값·mouse/키 재지정/스크롤·Reset/Defaults·Back/Apply/ESC/Close를 검사하며 E2E의 해상도/디스크 변경 억제는 유지한다.

비-E2E headless framebuffer/API 회귀는 별도로 실제 C apply와 game-entry/menu reset을 호출한다. 일반 8개/HD 10개 해상도·세 signed window mode에서 98/122개 frame 검사와 기존 GUI 네 폰트의 16개 실제 pixel/binding 검사를 각각 3회 통과했다. 버퍼 재생성 뒤 폰트 handle 연결이 끊기던 문제를 고쳤다. 생성한 font/seat fixture를 stock GUI/game-loop 전체의 해상도 적용이나 실제 macOS 창·Retina/물리 오디오 검증으로 확대하지 않는다. 원본 renderer의 기존 encoded PNG MD5 불일치는 변경 전 코드에서도 같으며 별도 미해결 항목이다. 설정 파일의 새 프로세스 복원·원본 default.cfg Reset·실제 OpenAL Soft null 검증도 각각 독립된 근거와 한계를 인벤토리에 기록한다.

플레이어/NPC 상호 마법 회귀는 `host-game-fireball-unit-damage.yaml`과 `host-game-magic-missile-unit-damage.yaml`에서 각각 1..5레벨 양방향 10회와 NPC 자연 시전 1회를 검사한다. 위치·대기 AI·내구성 있는 대상 최대 HP·정상 주문/대상 선택만 fixture다. 실제 시전·투사체·명중·피해·fractional carry·방어구·hit marker·audio·HP packet/HUD·투사체 제거는 주입하지 않는다. Fireball의 FLAME과 Magic Missile의 direct/splash EXPLOSION을 NPC 및 플레이어의 PlayerDamage→DefaultDamage 경로에 복원하고, 일반 방패로 막은 직접 명중의 source/type marker를 유지한다. 방어구 carry와 splash 반경 밖의 직접 명중도 구분한다.

`host-warrior-greatsword-missile-defense.yaml`은 실제 Warrior inventory 입력으로 GreatSword를 장착한 뒤 Fireball 정면/후면·1..5레벨·양방향 20회와 NPC 자연 시전 1회를 검사한다. 일반·HD headless/mock에서 각각 21회가 통과했다. 실제 반사 ownership/velocity·방어 audio·현재 block 자세/action·HP 보호·원본 binary32 내구도 carry와 후면의 실제 피해를 관찰하며 결과를 주입하지 않는다. GreatSword 장착의 정상 shield 자동 해제 뒤 잘못 남던 자세도 원본 subclass 필드로 복원했다. Fireball은 원본대로 대검 방어 뒤에도 폭발·제거되므로 돌아간 투사체가 caster에게 다시 명중했다는 주장이 아니다.

`host-game-mutual-status-spells.yaml`은 Confuse·Stun·Slow·Freeze·Blind의 양방향 10회에서 실제 명중 대상, duration/power, client/HUD 동기화, 자연 만료를 검사한다. Confuse/Stun은 원본 sprite와 두 시점의 실제 효과 픽셀·서로 다른 프레임을 비교하고, Slow는 대상 근처의 실제 YellowBubbleParticle을 검사한다. `host-game-shock-spell.yaml`은 player/NPC self-cast 1..5레벨, NPC 자연 시전 및 양방향 Shock Glyph 피해를 별도로 검사한다. 이들 테스트는 실제 입력의 incantation/mana, 모든 주문·장비의 방어·possessed unit·모든 class/자세·원격 대전을 검증했다는 의미가 아니다.

앞서 기록한 status 전체 화면 PNG 일치는 당시 실행 결과다. 후속 실행에서 횃불·크리스털 등 주변 배경 애니메이션의 프레임 차이로 전체 화면 비교가 실패했으나 실제 Confuse sprite는 55/55 픽셀 일치했다. 현재 status와 Shock의 판정은 기존의 실제 효과 픽셀/입자, 프레임 진행, buff/HUD, 만료 assertion을 유지한다. 전체 화면은 `CaptureMagicFrame`으로 매번 새 private 임시 PNG에 진단용으로 저장하며 기존 baseline이나 `NOX_E2E_OVERRIDE`를 변경하지 않는다. 이는 고정 배경의 전체 화면 golden을 새로 통과시켰다는 주장이 아니다.

패치 버전을 올릴 때에는 다음 항목을 한 변경으로 갱신한다.

1. `toolchain/go-version.txt`
2. `src/go.mod`의 `toolchain` 지시자
3. `src/internal/noxbuild/main.go`의 `requiredGoVersion`
4. Docker 이미지 태그와 CI 매트릭스
5. `BASELINE-linux-386.md` 및 전체 대상 빌드 결과

현재 포팅 단계, 정적 검색 재현법과 확인된 64비트 ABI 단절은 `PORTING-INVENTORY.md`에 기록한다.

첫 구조체 분리의 근거, C/Go 오프셋과 실제 Linux/386 산출물 검증값은 [`ABI-OBJECT.md`](ABI-OBJECT.md)에 기록한다.

사용자가 보유한 원본 데이터에 의존하는 시험은 [`oracle/README.md`](oracle/README.md)의 O0 무결성 게이트를 먼저 통과해야 한다. `make oracle-test`는 봉인된 `nox/` 이외의 데이터로 기대 결과가 바뀌는 것을 차단한다.

임의의 `GOEXPERIMENT`, 호스트 기본 아키텍처 최적화, `latest` 도구체인은 릴리스 산출물에 사용하지 않는다.

Linux Docker 빌더의 Go 베이스는 다음 멀티아키텍처 인덱스로 고정한다.

```text
golang:1.26.5-trixie@sha256:98988b42f3293b627bf07c884ff17181a59501769cd8c06c7ba901e0ce2c9853
```

이 인덱스에서 이 프로젝트가 사용하는 플랫폼은 `linux/386`, `linux/amd64`, `linux/arm/v7`, `linux/arm64/v8`이다.
