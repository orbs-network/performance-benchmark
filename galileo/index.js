const _ = require("lodash");
const shell = require("shelljs");
const { readFileSync, writeFileSync } = require("fs");
const { join } = require("path");
const { RTMClient } = require("@slack/client");
const { Promise } = require("bluebird");
const { getEndpoint, waitUntilCommit, waitUntilSync, getBlockHeight } = require("nebula/lib/metrics");
const { update, status, getNodes } = require("nebula/lib/cli");
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
    // await exec(`API_ENDPOINT=${getApiEndpoint(input.baseUrl)} BASE_URL=${input.baseUrl} VCHAIN=${input.vchain} COMMIT=${input.commit} RESULTS=${input.results} ./run.sh`);
    await exec(`API_ENDPOINT=${getApiEndpoint(input.baseUrl)} BASE_URL=${input.baseUrl} VCHAIN=${input.vchain} COMMIT=${input.commit} RESULTS=${input.results} ./extract_results.sh`);
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

async function deployCommit(commit, vcid, {configPath, contextPrefix, regions, awsProfile, vchains}) {
    console.log(`Deploying ${commit} to ${vcid}`);

    const nodes = await getNodes({ configPath });
    const endpoints = _.mapValues(nodes, (ip) => {
        return getEndpoint(ip, vcid);
    });
    const quorum = Math.round(2/3*_.size(nodes));

    const blockHeights = await Promise.props(_.mapValues(_.clone(endpoints), (endpoint) => {
        return getBlockHeight(endpoint);
    }));

    await Promise.some(_.map(regions, (region) => {
        return update({
            name: `${contextPrefix}-${region}`, // FIXME later
            region,
            configPath,
            chainVersion: vchains,
            awsProfile
        });
    }), quorum);

    await Promise.some(_.map(_.values(endpoints), (endpoint) => {
        return waitUntilCommit(endpoint, commit);
    }), quorum);

    await Promise.some(_.map(endpoints, (endpoint, region) => {
        return waitUntilSync(endpoint, blockHeights[region]);
    }), quorum);

    return (await status({ configPath, vchain: vcid })).result;
}

function formatNetworkStatus(networkStatus) {
    return _.map(networkStatus, (data, name) => {
        return `${name} ${data.status} blockHeight=${data.blockHeight} version=${data.version}@${_.truncate(data.commit, {length: 8, omission: ''})}`;
    }).join("\n");

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
        return _.mapKeys(JSON.parse(readFileSync(`${__dirname}/data/vchains.json`).toString()), (key, value) => {
            return _.parseInt(value);
        });
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

    const testnetConfig = process.env.TESTNET_CONFIG;
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

        const matchDeploy = message.text.match(/^deploy (\w+) to (\d+)/);
        if (matchDeploy) {
            const [text, commit, vcidStr] = matchDeploy;
            const vcid = _.parseInt(vcidStr);

            slack.sendMessage(`deploying <https://github.com/orbs-network/orbs-network-go/commit/${commit}|${commit}>@${vcid}, it could take some time`, message.channel);

            try {
                const networkStatus = await deployCommit(commit, vcid, {
                    configPath: testnetConfig,
                    regions,
                    awsProfile,
                    contextPrefix,
                    vchains: _.merge({}, vchainsCache, {[vcid]: commit})
                });

                // separate update after successful deployment
                vchainsCache[vcid] = commit;
                saveVchainsCache(vchainsCache);

                slack.sendMessage(formatNetworkStatus(networkStatus), message.channel);
                slack.sendMessage(`successful deploy for ${commit}@${vcid}`, message.channel);

                // FIXME run benchmark as a separate step

                // const path = join(`results`, commit, moment().format("Y-MM-DD-hhmmss"));
                // const baseUrl = getBaseUrl(endpoint, vcid, isGamma);
                //
                // const extractionOptions = {
                //     commit,
                //     baseUrl,
                //     results: path,
                //     vchain: vcid
                // };
                //
                // await extract(extractionOptions);
                // const metrics = getMetrics(path);
                //
                // console.log(`Extracted data to ${path}`);
                // console.log(`Metrics`, metrics);
                //
                // slack.sendMessage(`extracted data for ${commit}@${vcid}`, message.channel);
                // slack.sendMessage(`metrics for ${commit}@${vcid}:\n${metrics}`, message.channel);
                //
                // await uploadResults({path, resultsBucket});
                //
                // slack.sendMessage(`<@${message.user}>, you can download the results here: \`aws s3 sync --region us-west-2 ${resultsBucket}/${path} ${path}\``, message.channel);
            } catch(e) {
                console.log(e);
                slack.sendMessage(`<@${message.user}> failed to deploy ${commit}@${vcid}: ${e}`, message.channel);
            }
        }

        const matchStatus = message.text.match(/^status (\d+)/);
        if (matchStatus) {
            const [text, vcidStr] = matchStatus;
            const vcid = _.parseInt(vcidStr);

            try {
                slack.sendMessage(`determining network status for vchain ${vcid}, this may take some time`, message.channel);
                const networkStatus = (await status({ configPath: testnetConfig, vchain: vcid })).result;
                slack.sendMessage(formatNetworkStatus(networkStatus), message.channel);
            } catch(e) {
                console.log(e);
                slack.sendMessage(`<@${message.user}> failed to retrieve network status of ${vcid}: ${e}`, message.channel);
            }
        }
    });
})();
