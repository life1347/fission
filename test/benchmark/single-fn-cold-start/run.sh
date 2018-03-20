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

for i in {0..10}
do
    fn=nodejs-hello-$(date +%s)

    echo "Creating function"
    fission fn create --name $fn --env nodejs --code $ROOT/examples/nodejs/hello.js

    echo "Creating route"
    fission route create --function $fn --url /$fn --method GET

    echo "Waiting for router to catch up"
    sleep 5

    echo "Benchmarking for single cold-start time"
    # -e is not support in k6 official release yet.
    # k6 run -e FN_ENDPOINT="http://$FISSION_ROUTER/$fn" sample.js
    export FN_ENDPOINT="http://$FISSION_ROUTER/$fn"
    k6 run sample.js
    # ab -n 1 -c 1 $FN_ENDPOINT

    fission fn delete --name $fn
done

# crappy cleanup, improve this later
kubectl get httptrigger -o name | tail -1 | cut -f2 -d'/' | xargs kubectl delete httptrigger

echo "All done."
