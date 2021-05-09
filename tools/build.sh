#!/bin/bash

SCRIPT_PATH=$(dirname $(realpath "$0"))
PROJECT_PATH=$(realpath ${SCRIPT_PATH}/..)

cd ${PROJECT_PATH}

go build -o ${PROJECT_PATH}/bin/server echat/server
go build -o ${PROJECT_PATH}/bin/client echat/client

