cpu_arch ?= $(shell uname -m)

.PHONY: run-build-image
run-build-image: build-build-image
	docker run --rm std-index-build

.PHONY: build-build-image
build-build-image: run-test-image
	docker build --target build \
		-t std-index-build \
		-f Dockerfile \
		--build-arg CPU_ARCH=$(cpu_arch) .

.PHONY: run-test-image
run-test-image: build-test-image
	docker run --rm std-index-test

.PHONY: build-test-image
build-test-image:
	docker build --target tests \
		-t std-index-test \
		-f Dockerfile .

.PHONY: scan-secrets
scan-secrets:
	detect-secrets scan \
		--exclude-files '^tests/.*' \
		> .secrets.baseline
	detect-secrets audit .secrets.baseline
