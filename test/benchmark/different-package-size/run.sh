#!/bin/bash

set -euo pipefail

ROOT=$(dirname $0)/../../..


for i in 0..1
do
    for s in 0
    do

        # Create a hello world function in nodejs, test it with an http trigger
        echo "Pre-test cleanup"
        fission env delete --name python || true

        echo "Creating python env"
        # Use short grace period time to speed up resource recycle time
        fission env create --name python --version 2 --image fission/python-env --period 30
        trap "fission env delete --name python" EXIT

        sleep 20

        fn=python-hello-$(date +%s)


        echo "Creating package"
        rm -rf pkg.zip pkg/ || true
        mkdir pkg
        cp $ROOT/examples/python/hello.py pkg/hello.py

        # Create empty file with give size to simulate different size of package
        gtruncate -s ${s}MiB pkg/foo

        zip -jr pkg.zip pkg/
        pkgName=$(fission pkg create --env python --deploy pkg.zip | cut -d' ' -f 2 | cut -d"'" -f 2)


        echo "Creating function"
        fission fn create --name $fn --env python --pkg ${pkgName} --entrypoint "hello.main"

        echo "Creating route"
        fission route create --function $fn --url /$fn --method GET

        echo "Waiting for router to catch up"
        sleep 5

        echo "Benchmarking for single cold-start time"
        # -e is not support in k6 official release yet.
        # k6 run -e FN_ENDPOINT="http://$FISSION_ROUTER/$fn" sample.js
        export FN_ENDPOINT="http://$FISSION_ROUTER/$fn"
        export MAX_USERS=${i}

        # extract average request time from output
        k6 run --out json=l1.json sample.js


        fission fn delete --name $fn

        # crappy cleanup, improve this later
        echo kubectl get httptrigger -o name | tail -1 | cut -f2 -d'/' | xargs kubectl delete httptrigger

        echo "All done."
    done
done
