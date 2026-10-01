# Go 도구체인 정책

이 포팅 브랜치의 유일한 지원 도구체인은 `go1.26.5`이다. 버전의 단일 텍스트 기준은 `go-version.txt`이며, `src/go.mod`에는 언어 기준 `go 1.26.0`과 권장 도구체인 `toolchain go1.26.5`를 함께 선언한다.

빌드와 테스트는 다음 래퍼로 실행한다.

```sh
./scripts/go.sh version
./scripts/go.sh -C src test ./internal/noxbuild
```

Linux 64비트의 native-width CGo 회귀는 `make test-linux-pie`로 root·server·legacy 전체를 시험한다. 기본 비-PIE 실행 파일에서는 C heap이 4GiB 아래에 놓일 수 있어 고주소 포인터 시험의 전제가 성립하지 않을 수 있다. PIE 시험은 그 전제를 유지하는 별도 검증 게이트이며, 일반 제품의 빌드·실행 검증을 대체하지 않는다.

Windows PowerShell에서는 다음과 같이 실행한다.

```powershell
.\scripts\go.ps1 version
.\scripts\go.ps1 -C src test ./internal/noxbuild
```

실행할 제품이 현재 clean source revision과 정확히 일치하는지는 `noxbuild -verify`로 확인한다. 이 검사는 ELF, PE, Mach-O를 호스트 종류와 무관하게 읽어 Go 1.26.5, 지원 대상 tuple, 전체 `vcs.revision`, `vcs.modified=false`를 강제한다. 특히 무시된 `build/` 아래에 남은 구 실행 파일을 다시 실행하는 일을 차단한다.

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

래퍼는 외부 `GOROOT`를 제거하고 `GOTOOLCHAIN=go1.26.5`와 빈 `GOEXPERIMENT`를 강제한 뒤, 실제 `GOVERSION`이 정확히 일치하는지 검사한다. 이로써 goenv 같은 버전 관리자가 다른 표준 라이브러리 경로를 주입하는 경우도 차단한다. `GO` 환경 변수에는 공백 없는 Go 실행 파일 경로 하나만 지정할 수 있다. `internal/noxbuild`도 자신을 컴파일한 Go와 자식 빌드에 쓰는 Go를 각각 검사한다.

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

```sh
bash ./scripts/run-headless-server-e2e.sh \
  /absolute/path/to/nox /absolute/path/to/server-e2e-output
```

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
