#!/bin/bash

set -euo pipefail

ROOT=$(dirname $0)/../../../..

for executorType in poolmgr newdeploy
do
    for packagesize in 0 1 5 10 15 20
    do

        testDuration="5"
        dirName="package-size-${packagesize}-executor-${executorType}"

        # remove old data
        rm -rf ${dirName}
        mkdir ${dirName}
        pushd ${dirName}

        # run multiple iterations to reduce impact of imbalance of pod distribution.
        for iteration in {1..10}
        do

            # Create a hello world function in nodejs, test it with an http trigger
            echo "Pre-test cleanup"
            fission env delete --name python || true

            echo "Creating python env"
            # Use short grace period time to speed up resource recycle time
            # Use high min/max CPU so that K8S will distribute pod in different nodes

            version=2

            if [[ "${packagesize}" == "0" ]]
            then
                version=1
            fi

            fission env create --name python --version ${version} --image fission/python-env --period 5 --mincpu 300 --maxcpu 300 --minmemory 256 --maxmemory 256

            trap "fission env delete --name python" EXIT

            sleep 30

            fn=python-hello-$(date +%s)

            pkgName=""

            if [[ "${packagesize}" == "0" ]]
            then
                echo "Creating function"
                fission fn create --name $fn --env python --code $ROOT/examples/python/hello.py --executortype ${executorType} --minscale 3 --maxscale 3
            else
                echo "Creating package"
                rm -rf pkg.zip pkg/ || true
                mkdir pkg
                cp $ROOT/examples/python/hello.py pkg/hello.py

                # Create empty file with give size to simulate different size of package
                truncate -s ${packagesize}MiB pkg/foo

                zip -jr pkg.zip pkg/
                pkgName=$(fission pkg create --env python --deploy pkg.zip | cut -d' ' -f 2 | cut -d"'" -f 2)

                echo "Creating function"
                fission fn create --name $fn --env python --pkg ${pkgName} --entrypoint "hello.main" --executortype ${executorType} --minscale 3 --maxscale 3
            fi

            echo "Creating route"
            fission route create --function $fn --url /$fn --method GET

            echo "Waiting for router to catch up"
            sleep 5

            fnEndpoint="http://$FISSION_ROUTER/$fn"
            js="sample.js"
            rawFile="raw-${iteration}.json"
            rawUsageReport="raw-usage.txt"

            k6 run \
                -e FN_ENDPOINT="${fnEndpoint}" \
                --duration "${testDuration}s" \
                --rps 1 \
                --vus 1 \
                --no-connection-reuse \
                --out json="${rawFile}" \
                --summary-trend-stats="avg,min,med,max,p(5),p(10),p(15),p(20),p(25),p(30),p(35),p(40),p(45),p(50),p(55),p(60),p(65),p(70),p(75),p(80),p(85),p(90),p(95),p(100)" \
                ../${js} >> ${rawUsageReport}

            echo "Clean up"
            fission env delete --name python
            fission fn delete --name ${fn}
            fission route list| grep ${fn}| awk '{print $1}'| xargs fission route delete --name

            if [[ ! -z "${pkgName}" ]]
            then
                fission pkg delete --name ${pkgName} || true
                rm -rf pkg.zip pkg
            fi

            kubectl -n fission-function get pod -o name|xargs -I@ bash -c "kubectl -n fission-function delete @" || true

            echo "All done."
        done

        usageReport="usage.txt"
        cat ${rawUsageReport}| grep "http_req_duration"| cut -f2 -d':' > ${usageReport}

        popd

        # generate report after iterations are over
        outImage="${dirName}.png"
        ../picasso -file ${dirName} -format png -o ${outImage}

    done
done
