import http from "k6/http";
import { check } from "k6";

export let options = {
 rps: `${__ENV.MAX_RPS}`,
 stages: [
    { duration: "30s", target: 10 },
    { duration: "30s", target: `${__ENV.MAX_USERS}` },
  ]
};

export default function() {
    let res = http.get(`${__ENV.FN_ENDPOINT}`)
    check(res, {
        "status is 200": (r) => r.status === 200
    });
};
