run/gophermart:
	go run ./cmd/gophermart/main.go

run/accrual:
	./cmd/accrual/accrual_darwin_arm64 -a=:8000

run/dbs:
	docker compose up -d

migrate/up:
	goose -dir db/migrations postgres "postgres://admin:admin@localhost:5432/db?sslmode=disable" up

migrate/down:
	goose -dir db/migrations postgres "postgres://admin:admin@localhost:5432/db?sslmode=disable" down

build/gophermart:
	cd cmd/gophermart && go build -buildvcs=false -o gophermart && cd ../..

run/autotest: build/gophermart
	gophermarttest \
    -test.v -test.run=^TestGophermart$ \
    -gophermart-binary-path=cmd/gophermart/gophermart \
    -gophermart-host=localhost \
    -gophermart-port=8080 \
    -gophermart-database-uri="postgres://admin:admin@localhost:5432/db?sslmode=disable" \
    -accrual-binary-path=cmd/accrual/accrual_darwin_arm64 \
    -accrual-host=localhost \
    -accrual-port=8000 \
    -accrual-database-uri="postgres://admin:admin@localhost:5432/db?sslmode=disable"

run/statictest:
	go vet -vettool=/usr/local/bin/statictest ./...

.PHONY: run/gophermart, run/db, migrate/up, migrate/down, run/accrual, build/gophermart, autotest/run, run/statictest


GOLANGCI_LINT_CACHE?=/tmp/praktikum-golangci-lint-cache

.PHONY: golangci-lint-run
golangci-lint-run: _golangci-lint-rm-unformatted-report

.PHONY: _golangci-lint-reports-mkdir
_golangci-lint-reports-mkdir:
	mkdir -p ./golangci-lint

.PHONY: _golangci-lint-run
_golangci-lint-run: _golangci-lint-reports-mkdir
	-docker run --rm \
    -v $(shell pwd):/app \
    -v $(GOLANGCI_LINT_CACHE):/root/.cache \
    -w /app \
    golangci/golangci-lint:v1.57.2 \
        golangci-lint run \
            -c .golangci.yml \
	> ./golangci-lint/report-unformatted.json

.PHONY: _golangci-lint-format-report
_golangci-lint-format-report: _golangci-lint-run
	cat ./golangci-lint/report-unformatted.json | jq > ./golangci-lint/report.json

.PHONY: _golangci-lint-rm-unformatted-report
_golangci-lint-rm-unformatted-report: _golangci-lint-format-report
	rm ./golangci-lint/report-unformatted.json

.PHONY: golangci-lint-clean
golangci-lint-clean:
	sudo rm -rf ./golangci-lint
