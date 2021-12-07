#!/usr/bin/env bash

. script/utils.sh

# docker-compose files
DOCKER_COMPOSE_DB="docker/docker-compose-db.yml"
DOCKER_COMPOSE_NODE="docker/docker-compose-node.yaml"

export PATH="${PATH}:${PWD}/../bin/"
export FABRIC_CFG_PATH=${PWD}/config
export CHANNEL_NAME="atchain-channel"
export CORE_PEER_TLS_ENABLED=true

function networkDown() {
  docker-compose -f "${DOCKER_COMPOSE_DB}" -f "${DOCKER_COMPOSE_NODE}" down --volumes --remove-orphans
  docker run --rm -v "$(pwd)":/data busybox sh \
    -c 'cd /data && rm -rf system-genesis-block log.txt *.tar.gz channel-artifacts store/*.enc'
  docker run --rm -v "$(pwd)/../app":/data busybox sh -c 'cd /data && rm -rf ./user/wallets/*'

  # Remove useless containers
  CONTAINER_IDS=$(docker ps -a | awk '($2 ~ /dev-peer.*/) {print $1}')
  if [ -z "$CONTAINER_IDS" ] || [ "$CONTAINER_IDS" == " " ]; then
    infoln "No containers available for deletion"
  else
    docker rm -f "$CONTAINER_IDS"
  fi

  # Remove useless images
  DOCKER_IMAGE_IDS=$(docker images | awk '($1 ~ /dev-peer.*/) {print $3}')
  if [ -z "$DOCKER_IMAGE_IDS" ] || [ "$DOCKER_IMAGE_IDS" == " " ]; then
    infoln "No images available for deletion"
  else
    # shellcheck disable=2086
    docker rmi -f $DOCKER_IMAGE_IDS
  fi

  # delete MSP files
  docker run --rm -v "${PWD}/../organization":/data busybox sh -c "cd /data && rm -rf *"
}

function createOrgs() {
  if [ -d "../organization/peerOrganizations" ]; then
    rm -Rf organization/peerOrganizations && rm -Rf organization/ordererOrganizations
  fi

  infoln "generating cryptographic materials"

  cryptogen generate --config=./config/cryptogen/peer-crypto-config.yml --output="../organization"
  res=$?
  if [ $res -ne 0 ]; then
    fatalln "Failed to generate certificates"
  fi

  cryptogen generate --config=./config/cryptogen/orderer-crypto-config.yml --output="../organization"
  res=$?
  if [ $res -ne 0 ]; then
    fatalln "Failed to generate certificates"
  fi

  idemixgen ca-keygen --output="../organization/peerOrganizations/demo.com/idemix"

  infoln "generating CCP files"
  PEERPEM=../organization/peerOrganizations/demo.com/tlsca/tlsca.demo.com-cert.pem
  CAPEM=../organization/peerOrganizations/demo.com/ca/ca.demo.com-cert.pem
  jsonCCP $PEERPEM $CAPEM >./../organization/peerOrganizations/demo.com/connection-demo.json
  yamlCCP $PEERPEM $CAPEM >./../organization/peerOrganizations/demo.com/connection-demo.yml
}

function createConsortium() {
  mkdir -p "${PWD}/system-genesis-block/"

  infoln "generating the genesis block"
  configtxgen -profile ATChainGenesis -channelID system-channel -outputBlock ./system-genesis-block/genesis.block
  res=$?
  if [ $res -ne 0 ]; then
    fatalln "Failed to generate genesis block"
  fi

  local nodes="orderer.demo.com"
  for ((i = 0; i < 3; i++)); do
    nodes="${nodes} peer${i}.demo.com db-peer${i}"
  done
  docker-compose -f "${DOCKER_COMPOSE_DB}" -f "${DOCKER_COMPOSE_NODE}" up -d $nodes
}

function createChannel() {
  mkdir -p "channel-artifacts"

  infoln "Generating channel create transaction ${CHANNEL_NAME}.tx"
  configtxgen -profile ATChainChannel \
    -outputCreateChannelTx ./channel-artifacts/"$CHANNEL_NAME".tx -channelID "$CHANNEL_NAME"
  res=$?
  if [ $res -ne 0 ]; then
    fatalln "Failed to generate channel configuration transaction!"
  fi

  infoln "Generating anchor peer update transactions"
  configtxgen -profile ATChainChannel -outputAnchorPeersUpdate \
    ./channel-artifacts/DealerMSPanchors.tx -channelID "$CHANNEL_NAME" -asOrg DemoOrg
  res=$?
  if [ $res -ne 0 ]; then
    fatalln "Failed to generate anchor peer update transaction!"
  fi

  infoln "Creating channel ${CHANNEL_NAME}"
  setPeerEnv "0"
  rc=1
  COUNTER=1
  ORDERER_TLSCA=${PWD}/../organization/ordererOrganizations/demo.com/tlsca/tlsca.demo.com-cert.pem
  while [ $rc -ne 0 ] && [ $COUNTER -lt 5 ]; do
    sleep 1
    set -x
    peer channel create -o localhost:18860 -c "$CHANNEL_NAME" --ordererTLSHostnameOverride orderer.demo.com \
      -f ./channel-artifacts/"${CHANNEL_NAME}".tx --outputBlock ./channel-artifacts/"${CHANNEL_NAME}".block \
      --tls --cafile "$ORDERER_TLSCA" >&log.txt
    res=$?
    { set +x; } 2>/dev/null
    rc=$res
    COUNTER=$((COUNTER + 1))
  done
  cat log.txt
  if [ $res -ne 0 ]; then
    fatalln "Channel creation failed!"
    exit 1
  fi
  successln "Channel '$CHANNEL_NAME' created"
}

function joinChannel() {
  for ((i = 0; i < 3; i++)); do
    setPeerEnv ${i}
    env
    infoln "peer${i}.demo.com joining channel ${CHANNEL_NAME}"
    rc=1
    COUNTER=1
    while [ $rc -ne 0 ] && [ $COUNTER -lt 5 ]; do
      sleep 3
      set -x
      peer channel join -b ./channel-artifacts/"$CHANNEL_NAME".block >&log.txt
      res=$?
      { set +x; } 2>/dev/null
      rc=$res
      COUNTER=$((COUNTER + 1))
    done
    cat log.txt
    if [ $res -ne 0 ]; then
      fatalln "peer${i}.demo.com failed to join the channel '${CHANNEL_NAME}'!"
      exit 1
    fi
    infoln "peer${i}.demo.com successfully joined the channel '${CHANNEL_NAME}'."
  done
}

function up() {
  createOrgs
  createConsortium
  if [ ! -d "channel-artifacts" ]; then
    createChannel
  fi
  joinChannel
  docker ps
}

function deployCC() {
  CC_NAME="atchain-demo-cc"
  CC_SRC_LANGUAGE="go"
  CC_SRC_PATH="../chaincode-go"
  CC_VERSION="1.0"
  CC_SEQUENCE=""
  CC_INIT_FCN=""
  CC_COLL_CONFIG=""
  CC_END_POLICY=""

  while [[ $# -gt 1 ]]; do
    case $1 in
    -ccn)
      CC_NAME="$2"
      ;;
    -ccp)
      CC_SRC_PATH="$2"
      ;;
    -ccl)
      CC_SRC_LANGUAGE="$2"
      ;;
    -ccv)
      CC_VERSION="$2"
      ;;
    -ccs)
      CC_SEQUENCE="$2"
      ;;
    -cci)
      CC_INIT_FCN="$2"
      ;;
    -ccep)
      CC_END_POLICY="$2"
      ;;
    -cccg)
      CC_COLL_CONFIG="$2"
      ;;
    *)
      errorln "Unknown flag: $1"
      exit 1
      ;;
    esac
    shift
    shift
  done

  . ./script/deployCC.sh "$CC_NAME" "$CC_SRC_PATH" "$CC_SRC_LANGUAGE" "$CC_VERSION" "$CC_SEQUENCE" \
    "$CC_INIT_FCN" "$CC_END_POLICY" "$CC_COLL_CONFIG"

  queryInstalled "0"
  if [ -z "${CC_PACKAGE_ID}" ]; then
    packageCC
    for ((i = 0; i < 3; i++)); do
      installCC "${i}"
    done
    queryInstalled "0"
  fi

  approveForMyOrg "0"
  commitCCDef "0"
  queryCommitted "0"
}

function networkUp() {
  if [ $# -lt 1 ]; then
    up
  else
    case $1 in
    -ca)
      createOrgs
      ;;
    -nd)
      createConsortium
      ;;
    -cn)
      if [ ! -f "system-genesis-block/genesis.block" ]; then
        createChannel
      fi
      joinChannel
      ;;
    -cc)
      up
      deployCC
      ;;
    *)
      errorln "Unknown flag: $1"
      exit 1
      ;;
    esac
  fi
}

if [ $# -lt 1 ]; then
  errorln "Take at least 1 param."
  exit 0
else
  MODE=$1
fi

case $MODE in
"up")
  shift
  networkUp "$@"
  ;;
"dcc")
  shift
  deployCC "$@"
  ;;
"down")
  networkDown
  ;;
*)
  errorln "Unknown flag: $MODE"
  exit 1
  ;;
esac
