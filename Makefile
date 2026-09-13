GO_WRAPPER := ../scripts/go.sh
NOX_ORACLE_ROOT ?= ../nox
NOX_ORACLE_MANIFEST ?= toolchain/oracle/nox-2023-1003-01.json
NOX_CODE_MANIFEST ?= toolchain/oracle/game-exe-functions.json

.PHONY: oracle-verify oracle-code-verify oracle-test test-linux-pie test-linux-386 test-linux-armv7 test-darwin-amd64-server

# Linux non-PIE executables may place the C heap below 4 GiB. The native-width
# CGo tests intentionally require high addresses, so run their full gate as PIE.
test-linux-pie:
	./scripts/go.sh -C src test -buildmode=pie . ./server ./legacy -count=1

# Requires a Linux builder with i386 multilib and 32-bit OpenAL/ALSA/SDL2.
test-linux-386:
	CGO_CFLAGS_ALLOW='-f.*' GOARCH=386 CGO_ENABLED=1 CC='gcc -m32' \
		PKG_CONFIG_LIBDIR='/usr/lib/i386-linux-gnu/pkgconfig:/usr/share/pkgconfig' \
		./scripts/go.sh -C src test . ./server ./legacy -count=1

# Requires a native/emulated Linux arm/v7 builder with OpenAL/ALSA/SDL2.
test-linux-armv7:
	CGO_CFLAGS_ALLOW='-f.*' GOARCH=arm GOARM=7 CGO_ENABLED=1 \
		./scripts/go.sh -C src test . ./server ./legacy -count=1

# Requires macOS with an x86_64 Clang target; Apple Silicon also needs Rosetta.
# The server package gate does not require client-side OpenAL dependencies.
test-darwin-amd64-server:
	CGO_CFLAGS_ALLOW='-f.*' GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 \
		CC='clang -arch x86_64' ./scripts/go.sh -C src test ./server -count=1

oracle-verify:
	./scripts/go.sh -C src run ./internal/noxoracle verify \
		-root "$(abspath $(NOX_ORACLE_ROOT))" \
		-manifest "$(abspath $(NOX_ORACLE_MANIFEST))"

oracle-code-verify: oracle-verify
	./scripts/go.sh -C src run ./internal/noxoracle code-verify \
		-root "$(abspath $(NOX_ORACLE_ROOT))" \
		-manifest "$(abspath $(NOX_CODE_MANIFEST))" \
		-oracle-manifest "$(abspath $(NOX_ORACLE_MANIFEST))"

oracle-test: oracle-code-verify
	NOX_DATA="$(abspath $(NOX_ORACLE_ROOT))" NOX_ORACLE_STRICT=1 \
		./scripts/go.sh -C src test ./legacy/cnxz -run 'Test(Decompress|Compress)$$' -count=1
	./scripts/go.sh -C src run ./internal/noxoracle verify \
		-root "$(abspath $(NOX_ORACLE_ROOT))" \
		-manifest "$(abspath $(NOX_ORACLE_MANIFEST))"
	./scripts/go.sh -C src run ./internal/noxoracle code-verify \
		-root "$(abspath $(NOX_ORACLE_ROOT))" \
		-manifest "$(abspath $(NOX_CODE_MANIFEST))" \
		-oracle-manifest "$(abspath $(NOX_ORACLE_MANIFEST))"

format:
	clang-format --verbose --style=file -i ./src/*.c ./src/*.h ./src/*/*.c ./src/*/*.h ./src/*/*/*.c ./src/*/*/*.h  ./src/*/*/*/*.c ./src/*/*/*/*.h

build-server:
	cd ./src; \
	$(GO_WRAPPER) run ./internal/noxbuild -go=$(GO_WRAPPER) -arch=386 server

build-client:
	cd ./src; \
	$(GO_WRAPPER) run ./internal/noxbuild -go=$(GO_WRAPPER) -arch=386 client client-hd

build-client-win:
	cd ./src; \
	$(GO_WRAPPER) run ./internal/noxbuild -go=$(GO_WRAPPER) -os=windows -arch=386 client client-hd

build-server-docker:
	GIT_SHA=$$(git rev-parse --short HEAD); \
	GIT_TAG=$$(git name-rev --tags --name-only $$GIT_SHA); \
	docker build -t ghcr.io/opennox/opennox:dev -f ./docker/Dockerfile_server --target=server --build-arg GIT_SHA=$$GIT_SHA  --build-arg GIT_TAG=$$GIT_TAG ./src

build-server-demo-docker:
	GIT_SHA=$$(git rev-parse --short HEAD); \
	GIT_TAG=$$(git name-rev --tags --name-only $$GIT_SHA); \
	docker build -t ghcr.io/opennox/opennox:dev-demo -f ./docker/Dockerfile_server --target=demo --build-arg GIT_SHA=$$GIT_SHA  --build-arg GIT_TAG=$$GIT_TAG ./src

build-client-docker:
	GIT_SHA=$$(git rev-parse --short HEAD); \
	GIT_TAG=$$(git name-rev --tags --name-only $$GIT_SHA); \
	docker build -t ghcr.io/opennox/opennox-client:dev -f ./docker/Dockerfile_client --build-arg GIT_SHA=$$GIT_SHA  --build-arg GIT_TAG=$$GIT_TAG ./src
	mkdir -p ./build
	ID=$$(docker create ghcr.io/opennox/opennox-client:dev) && \
	docker cp $$ID:/home/runner/opennox/opennox ./build/ && \
	docker cp $$ID:/home/runner/opennox/opennox-hd ./build/ && \
	docker rm -f $$ID
