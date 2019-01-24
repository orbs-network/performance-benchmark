const _ = require("lodash");
const shell = require("shelljs");
const { readFileSync } = require("fs");
const { join } = require("path");

function extract(input) {
    const result = shell.exec(`API_ENDPOINT=${input.endpoint} COMMIT=${input.commit} ./extract.sh`, {        
        async: false,
    });
    
    return {
        path: result.stdout.trim(),
    };
}

function getMetrics(path) {
    const metrics = JSON.parse(readFileSync(join(path, "metrics.json")).toString());
    return _.pick(metrics, "PublicApi.GetTransactionStatusProcessingTime", "PublicApi.RunQueryProcessingTime");
}

(async () => {
    const { path } = extract({
        commit: "master",
        endpoint: "http://localhost:8080"        
    });

    console.log(`Extracted data to ${path}`);
    console.log(`Metrics`, getMetrics(path));
})();
