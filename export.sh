##### DATABASE CONFIGURATION #####
export DB_USER=crm
export DB_PASSWORD=crm.123
export DB_PORT=5432
export DB_HOST=localhost
export DB_NAME=crm
export DB_SSL_MODE=disable
export DB_MAX_OPEN_CONNS=500
export DB_MAX_IDLE_CONNS=25
export DB_MAX_CONN_LIFETIME=10
export DB_FILE=./db/px.db

##### REDIS CONFIGURATION #####
export REDIS_HOST=localhost
export REDIS_PORT=6379

##### NATS CONFIGURATION #####
export NATS_PORT=4222
export NATS_HOST=localhost

#### SMTP ENV ####
export SMTP_HOST=smtp.gmail.com
export SMTP_PORT=587
export SMTP_USERNAME=username
export SMTP_PASSWORD=password
export SMTP_FROM=from
export IMAP_HOST=imap.gmail.com
export IMAP_PORT=993

##### SERVER CONFIGURATION #####
export SERVER_PORT=8888
export SERVER_HOST=localhost
export LOG_FILE=logs/server.log
export DEBUG=true
export SWAGGER_FILE_PATH=./api/swagger/swagger.json

