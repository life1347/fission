#!/bin/bash

set -euo pipefail

ROOT=$(dirname $0)/../../..


for s in 0 2 4 6 8 10
do

    rm -rf pkg.zip pkg/ || true
    mkdir pkg
    cp $ROOT/examples/nodejs/hello.js pkg/hello.js

    # Create empty file with give size to simulate different size of package
    gtruncate -s ${s}MiB pkg/foo

    zip -jr pkg.zip pkg/
    pkgName=$(fission pkg create --env nodejs --deploy pkg.zip | cut -d' ' -f 2 | cut -d"'" -f 2)

    for i in 0..1
    do
        # Create a hello world function in nodejs, test it with an http trigger
        echo "Pre-test cleanup"
        fission env delete --name nodejs || true

        echo "Creating nodejs env"
        # Use short grace period time to speed up resource recycle time
        fission env create --name nodejs --image fission/node-env --period 30
        trap "fission env delete --name nodejs" EXIT

        sleep 20

        fn=nodejs-hello-$(date +%s)

        echo "Creating function"
        fission fn create --name $fn --env nodejs --pkg ${pkgName}

        echo "Creating route"
        fission route create --function $fn --url /$fn --method GET

        echo "Waiting for router to catch up"
        sleep 5

        echo "Benchmarking for single cold-start time"
        # -e is not support in k6 official release yet.
        # k6 run -e FN_ENDPOINT="http://$FISSION_ROUTER/$fn" sample.js
        export FN_ENDPOINT="http://$FISSION_ROUTER/$fn"

        # extract average request time from output
        k6 run --vus 1 --duration 1s --rps 1 sample.js | grep "http_req_duration" | awk '{print $2}'| sed 's/.*avg=*\(.*\).*/\1/' >> ${s}MB-time-size.txt
        # ab -n 1 -c 1 $FN_ENDPOINT

        fission fn delete --name $fn

        # crappy cleanup, improve this later
        kubectl get httptrigger -o name | tail -1 | cut -f2 -d'/' | xargs kubectl delete httptrigger

        echo "All done."
    done
done
