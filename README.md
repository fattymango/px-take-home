

## Documentation
Refer to [docs](./docs/) and wiki for the API documentation.

## Run
### Dependencies
To run the server, you need to have [Docker](https://docs.docker.com/get-docker/) installed.
then start the docker compose file in the root directory.

using `Makefile`
```bash
make up
```

or manually using `docker compose`
```bash
docker compose up -d
```


### Environment Variables

The server uses environment variables to configure the database connection and other settings.

Load the environment variables from the `export.sh` file:

```bash
source export.sh
```

### gRPC Setup

To generate the proto files, you can use the `Makefile` to generate the proto files.

First install the dependencies if not installed already
```bash
make init
```

Then generate the proto files
```bash
make proto
```

### Run the server

To run the server, you can use `Makefile` to start the server and other services.

```bash
make run
```

or manually start the server

```bash
go run cmd/api/main.go
```



## Seed 

To seed the database, you can use the `Makefile` to seed the database.

*Note: To run the seed, you need to have the dependencies running and the environment variables loaded.*

```bash
make seed
```

## API Documentation

API documentation is generated using Swagger. To access the documentation:

1. Run the server
2. Navigate to `http://localhost:8888/api/v1/docs#/`

To regenerate the Swagger documentation after making changes:

```bash
make swagger
```

or manually

```bash
go get github.com/swaggo/swag@master
swag init  --parseDependency -g ./cmd/api/main.go -o ./api/swagger 
```


## Docker Deployment

**Build and push Docker image**:
```bash
make push-dev
```


## Project Structure

```
server/
├── api/              # API definitions and Swagger documentation
├── app/              # Application initialization and configuration
├── bin/              # Compiled binaries
├── cmd/              # Entry points for applications
│   ├── api/          # Main API server
│   ├── migrate/      # Database migration tool
│   └── seed/         # Database seeding tool
├── docs/             # Documentation
├── dto/              # Data Transfer Objects (request/response models)
├── handler/          # HTTP request handlers
├── internal/         # Internal packages (core business logic)
│   ├── appointment/  # Appointment module
│   ├── authentication/ # Authentication module
│   ├── user/         # User module
│   ├── business/     # Business module
│   ├── ... other modules          # Car module
│   ├── casbin/       # Casbin authorization
│   ├── middleware/   # HTTP middleware
│   ├── shared/       # Shared data models and utilities
│   ├── utils/        # Utility functions
└── pkg/              # Shared libraries and wrappers
```

### Key Components
- **cmd**: Contains the main entry points for the application
- **app**: Application initialization and configuration
- **dto**: Contains data transfer objects for API requests and responses
- **handler**: HTTP request handlers for each module
- **internal**: Core business logic organized by domain
- **pkg**: Shared libraries and wrappers
- **api**: API definitions and Swagger documentation
- **docs**: Documentation



### Branch and Commit Conventions

- **Branch Name**: `{ticket_id}-{ticket_description}`
- **Commit Message**: `{ticket_id} - {brief_description_of_the_changes}`
- **MR Title**: Use gitlab merge request tool to create a new merge request.


### Code Conventions

- Do not use fmt.Println to print the logs, use the logger instead. If you want to debug something, use the logger.Debug() so we can disable it when we want to using the DEBUG=false.
  
- Handlers should be responsible for validating the request and response.
- Handlers should only orchestrate the data flow between the services/modules and should not contain any business logic.
- Use the predefined dto BaseResponse function to return data to the client.
- Write swagger documentation for all the handlers.