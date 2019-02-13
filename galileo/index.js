const _ = require("lodash");
const shell = require("shelljs");
const { readFileSync, writeFileSync } = require("fs");
const { join } = require("path");
const { RTMClient } = require("@slack/client");
const { Promise } = require("bluebird");
const { deploy } = require("nebula/deploy");
const moment = require("moment");

async function exec(command) {
    return new Promise((resolve, reject) => {
        shell.exec(command, (code, stdout, stderr) => {
            if (code != 0) {
                return reject(stderr);
            }

            resolve({path: stdout.trim()});
        });
    });
}

async function extract(input) {
    return exec(`API_ENDPOINT=${getApiEndpoint(input.baseUrl)} BASE_URL=${input.baseUrl} VCHAIN=${input.vchain} COMMIT=${input.commit} RESULTS=${input.results} ./extract.sh`);
}

async function uploadResults(input) {
    return exec(`aws s3 sync --region us-west-2 ${input.path} ${input.resultsBucket}/${input.path}`)
}

function formatHistogram(metrics, name) {
    const h = metrics[name];
    return `${name}: max ${h.Max}, p99 ${h.P99}, p95 ${h.P95}, p50 ${h.P50}, avg ${h.Avg}, min ${h.Min} (${h.Samples} samples)`;
}

function formatGauge(metrics, name, conversionFunction) {
    const g = metrics[name];
    const value = conversionFunction ? conversionFunction(g.Value) : g.Value;
    return `${name}: ${value}`;
}

function toMb(num) {
    return `${num / 1024 / 1024} Mb`;
}

// FIXME better output
function getMetrics(path) {
    const metrics = JSON.parse(readFileSync(join(path, "metrics.json")).toString());
    return [
        formatGauge(metrics, "BlockStorage.BlockHeight"),
        formatHistogram(metrics, "PublicApi.SendTransactionProcessingTime"),
        formatGauge(metrics, "Runtime.HeapAlloc", toMb),
    ].join("\n");
}

async function deployCommit(commit, vcid, {pathToConfig, contextPrefix, regions, awsProfile, vchains}) {
    console.log(`Deploying ${commit} to ${vcid}`);

    const exitCode = await deploy({
        updateVchains: true,
        pathToConfig,
        contextPrefix,
        regions,
        awsProfile,
        vchains
    });

    if (exitCode != 0) {
        throw `exit code: ${exitCode}`;
    }
}

function getSlackClient(token) {
    const client = new RTMClient(token);
    client.start();
    return client;
}

function getBaseUrl(endpoint, vcid, isGamma) {
    if (isGamma) {
        return endpoint
    }

    return `${endpoint}/vchains/${vcid}`;
}

function getApiEndpoint(baseUrl) {
    return `${baseUrl}/api/v1`
}

function loadVchainsCache() {
    try {
        return JSON.parse(readFileSync(`${__dirname}/data/vchains.json`).toString());
    } catch (e) {
        console.log(`Could not laod vchains: ${e}`);
    }
    return {};
}

function saveVchainsCache(cache) {
    writeFileSync(`${__dirname}/data/vchains.json`, JSON.stringify(cache, 2, 2));
}

(async () => {
    const token = process.env.SLACK_TOKEN;
    const endpoint = process.env.ENDPOINT || "http://localhost:8080";
    const isGamma = process.env.GAMMA == "true";

    const tesnetConfig = process.env.TESTNET_CONFIG;
    const awsProfile = process.env.AWS_PROFILE;
    const regions = (process.env.REGIONS || "").split(",");
    const contextPrefix = _.isUndefined(process.env.CONTEXT_PREFIX) ? undefined : process.env.CONTEXT_PREFIX;

    const resultsBucket = "s3://orbs-performance-benchmark";
    const slack = getSlackClient(token);

    const vchainsCache = loadVchainsCache();
    console.log(vchainsCache);

    slack.on("message", async (message) => {
        if ( (message.subtype && message.subtype === 'bot_message') || (!message.subtype && message.user === slack.activeUserId) ) {
            return;
        }

        const match = message.text.match(/^deploy (\w+) to (\d+)/);
        if (match) {
            const [text, commit, vcidStr] = match;
            const vcid = _.parseInt(vcidStr);

            slack.sendMessage(`deploying <https://github.com/orbs-network/orbs-network-go/commit/${commit}|${commit}>@${vcid}, it could take some time`, message.channel);

            try {
                await deployCommit(commit, vcid, {
                    pathToConfig: tesnetConfig,
                    regions,
                    awsProfile,
                    contextPrefix,
                    vchains: _.merge({}, vchainsCache, {[vcid]: commit})
                });

                // separate update after successful deployment
                vchainsCache[vcid] = commit;
                saveVchainsCache(vchainsCache);

                slack.sendMessage(`successful deploy for ${commit}@${vcid}`, message.channel);

                const path = join(`results`, commit, moment().format("Y-MM-DD-hhmmss"));
                const baseUrl = getBaseUrl(endpoint, vcid, isGamma);

                const extractionOptions = {
                    commit,
                    baseUrl,
                    results: path,
                    vchain: vcid
                };

                await extract(extractionOptions);
                const metrics = getMetrics(path);

                console.log(`Extracted data to ${path}`);
                console.log(`Metrics`, metrics);

                slack.sendMessage(`extracted data for ${commit}@${vcid}`, message.channel);
                slack.sendMessage(`metrics for ${commit}@${vcid}:\n${metrics}`, message.channel);

                await uploadResults({path, resultsBucket});

                slack.sendMessage(`<@${message.user}>, you can download the results here: \`aws s3 sync --region us-west-2 ${resultsBucket}/${path} ${path}\``, message.channel);
            } catch(e) {
                console.log(e);
                slack.sendMessage(`<@${message.user}> failed to deploy ${commit}@${vcid}: ${e}`, message.channel);
            }
        }
    });
})();
