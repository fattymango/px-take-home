# Web-Based Task Manager
This is a web-based task manager that allows user to submit long running tasks that run asynchronously and track the status and logs of the task.
This project was made as a take-home assignment for Proxidize.

## Architecture
![Architecture](./docs/high_level.excalidraw.png)

## Setup

### Prerequisites
- Go
- Docker (optional, for running the server as a container)
- Make (optional, for easier commands)

### Running the server
#### Run Manually
using `Makefile`
```bash
make run
```
or manually

```bash
go run cmd/api/main.go
```

#### Run as a container
```bash
make docker-build
make docker-run
```

or manually

```bash
docker build -t px-task-manager .
docker docker run -d -p 8888:8888 --name px-task-manager px-task-manager
```


## Usage

Navigate to `http://localhost:8888` to access the task manager web client.

### Create a new task
Fill in the task name and command to run.

Then click on the "Create Task" button to create the task.

![alt](./docs/img/create_task.png)

### View all tasks

Below the create task form, you can see all the tasks created and the status and exit code of the task.

![alt](./docs/img/view_tasks.png)


### Cancel a task

Click on the `Cancel` button to cancel the task, the button only appears when the task is running.

![alt](./docs/img/cancel_task.png)

When the task is cancelled, the task will be shown in the main page with a `Cancelled` status.

![alt](./docs/img/canceled_task.png)



### View task logs

Click on `View Logs` button to view the logs of the task.

![alt](./docs/img/view_task_logs_button.png)

Then you can see the logs of the task, if the task is still running, you can see the logs in real time.

![alt](./docs/img/view_task_logs.png)



## API Specification

### Swagger
If swagger specification does not exist, you can generate it using the following command:

```bash
make swagger
```

To access the documentation:

#### Web
1. Run the server
2. Navigate to `http://localhost:8888/api/v1/docs#/`

#### Local Files
Or find the generated swagger files in the `api/swagger` directory.

## API Documentation

### API Endpoints

#### Tasks

##### Create Task
- **Method**: POST
- **Path**: `/api/v1/tasks`
- **Payload**:
  ```json
  {
    "name": "string",
    "command": "string"
  }
  ```
- **Response**:
  ```json
  {
    "id": "string",
    "name": "string",
    "command": "string",
    "status": "string",
    "created_at": "timestamp"
  }
  ```

##### Get All Tasks
- **Method**: GET
- **Path**: `/api/v1/tasks`
- **Response**:
  ```json
  {
    "tasks": [
      {
        "id": "string",
        "name": "string",
        "command": "string",
        "status": "string",
        "created_at": "timestamp"
      }
    ]
  }
  ```

##### Get Task by ID
- **Method**: GET
- **Path**: `/api/v1/tasks/:taskID`
- **Response**:
  ```json
  {
    "id": "string",
    "name": "string",
    "command": "string",
    "status": "string",
    "created_at": "timestamp"
  }
  ```

##### Get Task Logs
- **Method**: GET
- **Path**: `/api/v1/tasks/:taskID/logs`
- **Response**:
  ```json
  {
    "logs": "string"
  }
  ```

##### Cancel Task
- **Method**: DELETE
- **Path**: `/api/v1/tasks/:taskID/cancel`
- **Response**:
  ```json
  {
    "message": "Task cancelled successfully"
  }
  ```

#### Server-Sent Events (SSE)

##### Subscribe to Events
- **Method**: GET
- **Path**: `/api/v1/events`
- **Description**: Establishes an SSE connection for real-time updates
- **Response**: Event stream with task updates
