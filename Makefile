

#################### SWAGGER ####################
.PHONY: swagger
swagger:
	@go get github.com/swaggo/swag@master
	@swag init  --parseDependency -g ./cmd/api/main.go -o ./api/swagger 


#################### GRPC ####################
.PHONY: init
init:
	@sudo apt install protobuf-compiler
	@go get -u google.golang.org/protobuf@v1.26.0 
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest	
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
	


# Generate the proto files
.PHONY: proto
proto:
	@protoc --go_out=./proto/ --go-grpc_out=./proto/ proto/*.proto



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
	@ $(MAKE) build-quick
	@ $(shell source exports.sh)
	@ ./bin/server

.PHONY: seed
seed:
	@ go run cmd/seed/main.go


# Race Detector
.PHONY: race
race:
	@CGO_ENABLED=1 go run -race cmd/main.go


.PHONY: migrate
migrate:
	@$(MAKE) build-migrate-quick
	@./bin/migrate

#################### Profiling ####################
.PHONY: pprof-cpu
pprof-cpu:
	@go tool pprof -http=":8000" pprofbin ./pprof/cpu/*.pprof 

.PHONY: pprof-mem
pprof-mem:
	@go tool pprof -http=":8000" pprofbin ./pprof/mem/*.pprof 


#################### Build Executable ####################
# Build amd64	for alpine
.PHONY: build
build:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-w -s' -o ./bin/server cmd/api/*.go

# Build depending on the OS
.PHONY: build-quick
build-quick:
	@go build  -o ./bin/server cmd/api/*.go


.PHONY: build-migrate
build-migrate:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-w -s' -o ./bin/migrate cmd/migrate/*.go

.PHONY: build-migrate-quick
	@go build  -o ./bin/migrate cmd/migrate/*.go




#################### Docker Compose ####################

# Stack
.PHONY: up
up:
	@docker compose up -d --build --force-recreate --remove-orphans

.PHONY: down
down:
	@docker compose down 

.PHONY: top
top:
	@docker stats
	
.PHONY: down-volumes
down-volumes:
	@docker compose down -v



.PHONY: push-dev
push-dev:
	@docker login registry.gitlab.com
	@ $(MAKE) swagger
	@ $(MAKE) build
	@docker build -t registry.gitlab.com/:dev .
	@docker push registry.gitlab.com/:dev

#################### Logs ####################

.PHONY: logs-server
logs-server:
	@docker logs server -f


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
	@$(MAKE) sqlite-create-db
	@$(MAKE) clean
	@$(MAKE) run

.PHONY: install-deps
install-deps:
	@go mod tidy
	@sudo apt  install shellcheck

.PHONY: clean
clean:
	@rm -r task_logs
