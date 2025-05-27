##### POSTGRESQL CONFIGURATION #####
export DB_USER=crm
export DB_PASSWORD=crm.123
export DB_PORT=5432
export DB_HOST=localhost
export DB_NAME=crm
export DB_SSL_MODE=disable
export DB_MAX_OPEN_CONNS=500
export DB_MAX_IDLE_CONNS=25
export DB_MAX_CONN_LIFETIME=10

### SQLITE CONFIGURATION ###
export DB_FILE=./db/px.db


##### SERVER CONFIGURATION #####
export SERVER_PORT=8888
export SERVER_HOST=localhost
export LOG_FILE=logs/server.log
export DEBUG=true
export SWAGGER_FILE_PATH=./api/swagger/swagger.json
export CMD_VALIDATE=false
export TASK_LOGGER_DIR_PATH=./task_logs
