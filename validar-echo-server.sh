#!/bin/bash

NETWORK="tp0_testing_net"
SERVER="server"
PORT="12345"
MESSAGE="ping"

RESPONSE=$(echo "$MESSAGE" | docker run --rm -i \
  --network "$NETWORK" \
  alpine \
  sh -c "nc -w 2 $SERVER $PORT")

if [ "$RESPONSE" = "$MESSAGE" ]; then
  echo "action: test_echo_server | result: success"
else
  echo "action: test_echo_server | result: fail"
fi