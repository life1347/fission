#!/bin/bash

set -euo pipefail

ROOT=$(dirname $0)/../../../..

for executorType in poolmgr newdeploy
do
    for concurrency in {1..2}
    do

        concurrencyLevel=$((${concurrency}*100))
        dirName="concurrency-${concurrencyLevel}-executor-${executorType}"

        # remove old data
        rm -rf ${dirName}
        mkdir ${dirName}
        pushd ${dirName}

        # run multiple iterations to reduce impact of imbalance of pod distribution.
        for iteration in {1..3}
        do

            # Create a hello world function in nodejs, test it with an http trigger
            echo "Pre-test cleanup"
            fission env delete --name python || true

            echo "Creating python env"
            # Use short grace period time to speed up resource recycle time
            fission env create --name python --version 2 --image fission/python-env --period 30
            trap "fission env delete --name python" EXIT

            sleep 30

            fn=python-hello-$(date +%s)

            echo "Creating package"
            rm -rf pkg.zip pkg/ || true
            mkdir pkg
            cp $ROOT/examples/python/hello.py pkg/hello.py

            zip -jr pkg.zip pkg/
            pkgName=$(fission pkg create --env python --deploy pkg.zip | cut -d' ' -f 2 | cut -d"'" -f 2)

            echo "Creating function"
            fission fn create --name $fn --env python --pkg ${pkgName} --entrypoint "hello.main" --executortype ${executorType} \
                --mincpu 100 --maxcpu 100 --minmemory 128 --maxmemory 128 --minscale 3 --maxscale 3

            echo "Creating route"
            fission route create --function $fn --url /$fn --method GET

            echo "Waiting for router to catch up"
            sleep 5

            testDuration="10s"
            fnEndpoint="http://$FISSION_ROUTER/$fn"
            js="sample.js"
            usageReport="usage.txt"
            rawFile="raw-${iteration}.json"
            outFile="output-${iteration}.json"

            # cold start
            curl ${fnEndpoint}

            k6 run \
                -e FN_ENDPOINT="${fnEndpoint}" \
                --duration ${testDuration} \
                --rps ${concurrencyLevel} \
                --vus ${concurrencyLevel} \
                --no-connection-reuse \
                --out json="${rawFile}" \
                -summary-trend-stats="avg,min,med,max,p(90),p(95),p(99)" \
                ../${js} >> ${usageReport}

            jq -cr '. | select(.type=="Point" and .metric == "http_req_duration" and .data.tags.status >= "200")' ${rawFile} > ${outFile}

            echo "Clean up"
            fission fn delete --name ${fn}
            fission route list|grep ${fn}|awk '{print $1}'|xargs fission route delete --name
            fission pkg delete --name ${pkgName}
            rm -rf pkg.zip pkg

            echo "All done."
        done

        # generate report after iterations are over
        outImage="output.png"

        ../../picasso -file ${outFile} -o ${outImage}

        popd

    done
done
