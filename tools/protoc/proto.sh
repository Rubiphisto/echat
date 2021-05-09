SCRIPT_PATH=$(dirname $(realpath "$0"))
PROTOC=$(realpath ${SCRIPT_PATH}/bin/protoc)
PROTO_PATH=$(realpath ${SCRIPT_PATH}/../../common/proto)
PB_PATH=$(realpath ${SCRIPT_PATH}/../..)

${PROTOC} --go_out=${PB_PATH} --proto_path=${PROTO_PATH} chat.proto

