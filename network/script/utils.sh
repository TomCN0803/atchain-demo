#!/bin/env bash

C_RESET='\033[0m'
C_RED='\033[0;31m'
C_GREEN='\033[0;32m'
C_BLUE='\033[0;34m'
C_YELLOW='\033[1;33m'

# println echos string
function println() {
  echo -e "$1"
}

# errorln echos i red color
function errorln() {
  println "${C_RED}${1}${C_RESET}"
}

# successln echos in green color
function successln() {
  println "${C_GREEN}${1}${C_RESET}"
}

# infoln echos in blue color
function infoln() {
  println "${C_BLUE}${1}${C_RESET}"
}

# warnln echos in yellow color
function warnln() {
  println "${C_YELLOW}${1}${C_RESET}"
}

# fatalln echos in red color and exits with fail status
function fatalln() {
  errorln "$1"
  exit 1
}

# shut the shell up
function die() {
  exit 2
}

function oneLinePem() {
  # shellcheck disable=2005
  echo "$(awk 'NF {sub(/\\n/, ""); printf "%s\\\\n",$0;}' "$1")"
}

function jsonCCP() {
  local PP=$(oneLinePem $1)
  local CP=$(oneLinePem $2)
  sed -e "s#\${PEERPEM}#$PP#" \
    -e "s#\${CAPEM}#$CP#" \
    ./config/ccp-template.json
}

function yamlCCP() {
  local PP=$(oneLinePem $1)
  local CP=$(oneLinePem $2)
  sed -e "s#\${PEERPEM}#$PP#" \
    -e "s#\${CAPEM}#$CP#" \
    ./config/ccp-template.yml | sed -e $'s/\\\\n/\\\n          /g'
}

function setPeerEnv() {
  local PEER=$1
  export CORE_PEER_LOCALMSPID=DemoMSP
  export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/../organization/peerOrganizations/demo.com/peers/peer${PEER}.demo.com/tls/ca.crt
  export CORE_PEER_MSPCONFIGPATH=${PWD}/../organization/peerOrganizations/demo.com/users/Admin@demo.com/msp
  export CORE_PEER_ADDRESS=localhost:1885${PEER}
}

export -f println
export -f errorln
export -f successln
export -f infoln
export -f warnln
export -f fatalln
export -f die
export -f oneLinePem
export -f jsonCCP
export -f yamlCCP
export -f setPeerEnv
