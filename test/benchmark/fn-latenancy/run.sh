#!/bin/bash

set -euo pipefail

ROOT=$(dirname $0)/../../..

# Create a hello world function in nodejs, test it with an http trigger
echo "Pre-test cleanup"
fission env delete --name nodejs || true

echo "Creating nodejs env"
fission env create --name nodejs --image fission/node-env --poolsize 10
trap "fission env delete --name nodejs" EXIT

sleep 10

fn=nodejs-hello-$(date +%s)

echo "Creating function"
fission fn create --name $fn --env nodejs --code $ROOT/examples/nodejs/hello.js

echo "Creating route"
fission route create --function $fn --url /$fn --method GET

echo "Waiting for router to catch up"
sleep 5

echo "Benchmarking for single cold-start time"

for i in {0..10}
do
    # -e is not support in k6 official release yet.
    # k6 run -e FN_ENDPOINT="http://$FISSION_ROUTER/$fn" sample.js
    export FN_ENDPOINT="http://$FISSION_ROUTER/$fn"

    # extract average request time from output
    k6 run --duration 60 --vus 100 sample.js | grep "http_req_duration" | awk '{print $2}'| sed 's/.*avg=*\(.*\).*/\1/' >> time.txt
    # ab -n 1 -c 1 $FN_ENDPOINT
done

fission fn delete --name $fn

# crappy cleanup, improve this later
kubectl get httptrigger -o name | tail -1 | cut -f2 -d'/' | xargs kubectl delete httptrigger

echo "All done."
#!/usr/bin/env bash
