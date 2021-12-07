#!/usr/bin/env bash

CC_NAME=$1     # chaincode name
CC_SRC_PATH=$2 # chaincode path
CC_SRC_LANGUAGE=${3:-"go"}
CC_VERSION=${4:-"1.0"}
CC_SEQUENCE=${5:-"1"}
CC_INIT_FCN=${6:-"NA"}    # init function
CC_END_POLICY=${7:-"NA"}  # chaincode endorsement policy
CC_COLL_CONFIG=${8:-"NA"} # private collection configuration

INIT_REQUIRED=""
if [ "$CC_INIT_FCN" != "NA" ]; then
  INIT_REQUIRED="--init-required"
fi

if [ "$CC_END_POLICY" = "NA" ]; then
  CC_END_POLICY=""
else
  CC_END_POLICY="--signature-policy $CC_END_POLICY"
fi

# Check private data collection
if [ "$CC_COLL_CONFIG" = "NA" ]; then
  CC_COLL_CONFIG=""
else
  CC_COLL_CONFIG="--collections-config $CC_COLL_CONFIG"
fi

function packageCC() {
  if [ ! -d "$CC_SRC_PATH" ]; then
    fatalln "Path to chaincode does not exist. Please provide different path!"
    exit 1
  fi

  CC_RUNTIME_LANGUAGE=""
  case $CC_SRC_LANGUAGE in
  "go")
    CC_RUNTIME_LANGUAGE="golang"
    infoln "Vendoring Go dependencies at $CC_SRC_PATH"
    pushd "$CC_SRC_PATH" || exit 1
    GO111MODULE=on go mod vendor
    popd || exit 1
    successln "Finished vendoring Go dependencies"
    ;;
  "java")
    CC_RUNTIME_LANGUAGE=java
    infoln "Compiling Java code..."
    pushd "$CC_SRC_PATH" || exit 1
    ./gradlew installDist
    popd || exit 1
    successln "Finished compiling Java code"
    CC_SRC_PATH=$CC_SRC_PATH/build/install/$CC_NAME
    ;;
  "javascript")
    CC_RUNTIME_LANGUAGE=node
    ;;
  *)
    fatalln "The chaincode language ${CC_SRC_LANGUAGE} is not supported by this script. Supported chaincode languages are: go, java, and javascript!"
    exit 1
    ;;
  esac

  infoln "Packaging the chaincode"
  set -x
  peer lifecycle chaincode package "${CC_NAME}".tar.gz --path "${CC_SRC_PATH}" --lang "${CC_RUNTIME_LANGUAGE}" \
    --label "${CC_NAME}"_"${CC_VERSION}" >&log.txt
  res=$?
  { set +x; } 2>/dev/null
  cat log.txt
  if [ $res -ne 0 ]; then
    fatalln "Chaincode packaging has failed!"
    exit 1
  fi
  successln "Chaincode is packaged"
}

function installCC() {
  local PEER=$1
  setPeerEnv $PEER
  infoln "Installing chaincode on ${PEER}.demo.com..."
  set -x
  peer lifecycle chaincode install "${CC_NAME}".tar.gz >&log.txt
  res=$?
  { set +x; } 2>/dev/null
  cat log.txt
  if [ $res -ne 0 ]; then
    fatalln "Chaincode installation on peer${PEER}.demo.com has failed!"
    exit 1
  fi
  successln "Chaincode is successfully installed on peer${PEER}.demo.com"
}

function queryInstalled() {
  local PEER=$1
  setPeerEnv $PEER
  set -x
  peer lifecycle chaincode queryinstalled >&log.txt
  res=$?
  { set +x; } 2>/dev/null
  cat log.txt
  CC_PACKAGE_ID=$(sed -n "/${CC_NAME}_${CC_VERSION}/{s/^Package ID: //; s/, Label:.*$//; p;}" log.txt)
  if [ $res -ne 0 ]; then
    fatalln "Query installed on peer${PEER}.demo.com has failed!"
    exit 1
  fi
  successln "Query installed successful on peer${PEER}.demo.com on channel"
}

function approveForMyOrg() {
  local PEER=$1
  setPeerEnv $PEER
  set -x
  # shellcheck disable=2086
  peer lifecycle chaincode approveformyorg -o localhost:18860 --ordererTLSHostnameOverride orderer.demo.com \
    --tls --cafile $ORDERER_TLSCA --channelID $CHANNEL_NAME --name ${CC_NAME} --version ${CC_VERSION} \
    --package-id ${CC_PACKAGE_ID} --sequence ${CC_SEQUENCE} ${INIT_REQUIRED} ${CC_END_POLICY} ${CC_COLL_CONFIG} \
    >&log.txt
  res=$?
  { set +x; } 2>/dev/null
  cat log.txt
  if [ $res -ne 0 ]; then
    fatalln "Chaincode definition approved on peer${PEER}.demo.com on channel '$CHANNEL_NAME' failed"
    exit 1
  fi
  successln "Chaincode definition approved on peer${PEER}.demo.com on channel '$CHANNEL_NAME'"
}

function commitCCDef() {
  local PEER=$1
  setPeerEnv $PEER
  PEER_CONN_PARAMS="--peerAddresses localhost:1885${PEER} --tlsRootCertFiles ${PWD}/../organization/peerOrganizations/demo.com/peers/peer${PEER}.demo.com/tls/ca.crt"

  set -x
  # shellcheck disable=2086
  peer lifecycle chaincode commit -o localhost:18860 --ordererTLSHostnameOverride orderer.demo.com --tls \
    --cafile $ORDERER_TLSCA --channelID $CHANNEL_NAME --name ${CC_NAME} $PEER_CONN_PARAMS --version ${CC_VERSION} \
    --sequence ${CC_SEQUENCE} ${INIT_REQUIRED} ${CC_END_POLICY} ${CC_COLL_CONFIG} >&log.txt
  res=$?
  { set +x; } 2>/dev/null
  cat log.txt
  if [ $res -ne 0 ]; then
    fatalln "Chaincode definition commit failed on channel '$CHANNEL_NAME' failed"
    exit 1
  fi
  successln "Chaincode definition committed on channel '$CHANNEL_NAME'"
}

function queryCommitted() {
  local PEER=$1
  setPeerEnv $PEER
  EXPECTED_RESULT="Version: ${CC_VERSION}, Sequence: ${CC_SEQUENCE}, Endorsement Plugin: escc, Validation Plugin: vscc"
  infoln "Querying chaincode definition on ${PEER}.demo.com on channel '$CHANNEL_NAME'..."
  local rc=1
  local COUNTER=1
  while [ $rc -ne 0 ] && [ $COUNTER -lt 5 ]; do
    sleep 2
    infoln "Attempting to Query committed status on ${PEER}.demo.com, Retry after 2 seconds."
    set -x
    peer lifecycle chaincode querycommitted --channelID "$CHANNEL_NAME" --name "${CC_NAME}" >&log.txt
    res=$?
    { set +x; } 2>/dev/null
    # shellcheck disable=2002
    test $res -eq 0 && VALUE=$(cat log.txt | grep -o '^Version: '"$CC_VERSION"', Sequence: [0-9]*, Endorsement Plugin: escc, Validation Plugin: vscc')
    test "$VALUE" = "$EXPECTED_RESULT" && ((rc = 0))
    ((COUNTER++))
  done
  cat log.txt
  if [ "$rc" -eq 0 ]; then
    successln "Query chaincode definition successful on peer${PEER}.demo.comon channel '$CHANNEL_NAME'"
  else
    fatalln "After $MAX_RETRY attempts, Query chaincode definition result on peer${PEER}.demo.com is INVALID!"
  fi
}
