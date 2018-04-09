#!/bin/bash

set -euo pipefail

ROOT=$(dirname $0)/../../..

fn=nodejs-hello-$(date +%s)

 Create a hello world function in nodejs, test it with an http trigger
echo "Pre-test cleanup"
fission env delete --name nodejs || true

echo "Creating nodejs env"
fission env create --name nodejs --image fission/node-env --poolsize 10
trap "fission env delete --name nodejs" EXIT

echo "Creating function"
fission fn create --name $fn --env nodejs --code $ROOT/examples/nodejs/hello.js
trap "fission fn delete --name $fn" EXIT

echo "Creating route"
fission route create --function $fn --url /$fn --method GET

echo "Waiting for router to catch up"
sleep 10

echo "Benchmarking for single cold-start time"
# -e is not support in k6 official release yet.
# k6 run -e FN_ENDPOINT="http://$FISSION_ROUTER/$fn" sample.js
export FN_ENDPOINT="http://$FISSION_ROUTER/$fn"
k6 run sample.js

# crappy cleanup, improve this later
kubectl get httptrigger -o name | tail -1 | cut -f2 -d'/' | xargs kubectl delete httptrigger

echo "All done."
