

#################### SWAGGER ####################
.PHONY: swagger
swagger:
	@go get github.com/swaggo/swag@master
	@swag init  --parseDependency -g ./cmd/api/main.go -o ./api/swagger 



#################### Testing ####################
.PHONY: test
test:
	@go test -v ./... -cover

.PHONY: bench
bench:
	@for d in $$(go list ./...); do \
		# go test -bench=.  -benchmem -cpuprofile=$(HOME)/crm/server/pprof/cpu/cpu_$$(basename $$d).pprof -memprofile=$(HOME)/crm/server/pprof/mem/mem_$$(basename $$d).pprof $$d; \
		go test -bench=.  -benchmem $$d; \
	done
	@$(MAKE) clean-test
	

.PHONY: clean-test
clean-test:
	@rm ./*.test


#################### Linting ####################
.PHONY: format
fmt , format:
	@go fmt ./...



#################### RUN ####################
	
# Run the API server
.PHONY: run
run:
	@ $(MAKE) build
	@ ./bin/server



# Race Detector
.PHONY: race
race:
	@CGO_ENABLED=1 go run -race cmd/api/main.go


.PHONY: migrate
migrate:
	@$(MAKE) build-migrate-quick
	@./bin/migrate

#################### Profiling ####################
.PHONY: pprof-cpu
pprof-cpu:
	@go tool pprof -http=":8000" pprofbin ./profile/cpu/*.pprof 

.PHONY: pprof-mem
pprof-mem:
	@go tool pprof -http=":8000" pprofbin ./profile/mem/*.pprof 


#################### Build Executable ####################
# Build amd64	for alpine
.PHONY: build-alpine
build-alpine:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-w -s' -o ./bin/server cmd/api/*.go

# Build depending on the OS
.PHONY: build
build:
	@go build  -o ./bin/server cmd/api/*.go


#################### SQLITE ####################

.PHONY: sqlite-flush-db
sqlite-flush-db:
	@rm -f ./db/px.db

.PHONY: sqlite-create-db
sqlite-create-db:
	@touch ./db/px.db



.PHONY: fresh-start
fresh-start:
	@$(MAKE) sqlite-flush-db
	@$(MAKE) clean-logs
	@$(MAKE) run

.PHONY: install-deps
install-deps:
	@go mod tidy
	@sudo apt  install shellcheck

.PHONY: clean-logs
clean-logs:
	@rm -rf task_logs

