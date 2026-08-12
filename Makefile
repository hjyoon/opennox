GO_WRAPPER := ../scripts/go.sh
NOX_ORACLE_ROOT ?= ../nox
NOX_ORACLE_MANIFEST ?= toolchain/oracle/nox-2023-1003-01.json
NOX_CODE_MANIFEST ?= toolchain/oracle/game-exe-functions.json

.PHONY: oracle-verify oracle-code-verify oracle-test

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
