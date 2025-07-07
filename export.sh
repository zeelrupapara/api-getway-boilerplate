#!/bin/bash

# GreenLync API Gateway - Environment Variables Export Script
# Developer: zeelrupapara@gmail.com
# Last Update: Cannabis boilerplate conversion

# Basic configuration
export BASE_URL="http://127.0.0.1:8888"
export LOG_FILE="./logs/app.log"
export HTTP_HOST="0.0.0.0"
export HTTP_PORT=":8888"
export OAUTH_TOKEN_EXPIRES_IN="3600"
export OAUTH_LONG_TOKEN_EXPIRES_IN="86400"
export APP_REPORTS="./reports"

# GRPC (not used in boilerplate but required by config)
export GRPC_HOST="127.0.0.1"
export GRPC_PORT=":9999"

# Redis configuration
export REDIS_URL="127.0.0.1:6379"
export REDIS_PASSWORD=none

# MySQL configuration
export MYSQL_HOST="127.0.0.1"
export MYSQL_PORT="3306"
export MYSQL_USER="root"
export MYSQL_PASSWORD="password"
export MYSQL_DB="greenlync"
export MYSQL_USE_SSL="false"
export MYSQL_CACERT_PATH="none"

# NATS configuration
export NATS_HOST="127.0.0.1"
export NATS_PORT="4222"

# SMTP configuration (optional for testing)
export SMTP_HOST="smtp.gmail.com"
export SMTP_PORT="587"
export SMTP_FROM="noreply@greenlync.com"
export SMTP_PASSWORD="test_password"
export SMTP_LOGIN="test_user"