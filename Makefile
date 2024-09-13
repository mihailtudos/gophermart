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

.PHONY: run/gophermart, run/db, migrate/up, migrate/down, run/accrual