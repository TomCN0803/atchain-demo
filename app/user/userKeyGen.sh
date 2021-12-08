#!/usr/bin/env bash

. ./../../network/script/utils.sh

export PATH="${PATH}:${PWD}/../../bin/"

if ! idemixgen version; then
  fatalln "idemixgen binary not found"
fi

OU=$1
EID=$2
RH=$3

CAINPUT="${PWD}/../../organization/peerOrganizations/demo.com/idemix"
OUTPUT="${PWD}/wallets/${EID}-${OU}"

mkdir -p "${OUTPUT}"
mkdir -p "${OUTPUT}/conn"
cp -r "../../organization/peerOrganizations/demo.com/users/${EID}@demo.com" "${OUTPUT}/conn"

idemixgen signerconfig -u ${OU} -e ${EID} -r ${RH} --ca-input=${CAINPUT} --output=${OUTPUT}
