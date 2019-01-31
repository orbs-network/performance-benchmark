const _ = require("lodash");
const shell = require("shelljs");
const { readFileSync } = require("fs");
const { join } = require("path");
const { RTMClient } = require('@slack/client');
const { Promise } = require("bluebird");
const { deploy } = require("nebula/deploy");

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
    return exec(`API_ENDPOINT=${input.endpoint} COMMIT=${input.commit} ./extract.sh`);
}

async function uploadResults(input) {
    return exec(`aws s3 sync --region us-west-2 ${input.path} ${input.resultsBucket}/${input.path}`)
}

function getMetrics(path) {
    const metrics = JSON.parse(readFileSync(join(path, "metrics.json")).toString());
    return _.pick(metrics, "PublicApi.GetTransactionStatusProcessingTime", "PublicApi.RunQueryProcessingTime");
}

async function deployCommit(commit, vcid, {pathToConfig, regions, awsProfile}) {
    console.log(`Deploying ${commit} to ${vcid}`);
    const exitCode = await deploy({
        updateVchains: true,
        chainVersion: commit,
        pathToConfig,
        regions,
        awsProfile,
        vchains: [vcid]
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

function getEndpoint(endpoint, vcid, isGamma) {
    if (isGamma) {
        return endpoint;
    }

    return [endpoint, "vchains", vcid].join("/");
}

(async () => {
    const token = process.env.SLACK_TOKEN;
    const endpoint = process.env.API_ENDPOINT || "http://localhost:8080";
    const isGamma = process.env.GAMMA == "true";

    const tesnetConfig = process.env.TESTNET_CONFIG;
    const awsProfile = process.env.AWS_PROFILE;
    const regions = (process.env.REGIONS || "").split(",");

    const resultsBucket = "s3://orbs-performance-benchmark";
    const slack = getSlackClient(token);

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
                    awsProfile
                });
                slack.sendMessage(`successful deploy for ${commit}@${vcid}`, message.channel);

                const { path } = await extract({commit, endpoint: getEndpoint(endpoint, vcid, isGamma)});
                const metrics = getMetrics(path);

                console.log(`Extracted data to ${path}`);
                console.log(`Metrics`, metrics);

                slack.sendMessage(`extracted data for ${commit}@${vcid}`, message.channel);
                slack.sendMessage(`metrics for ${commit}@${vcid}:\n${JSON.stringify(metrics)}`, message.channel);

                await uploadResults({path, resultsBucket});

                slack.sendMessage(`<@${message.user}>, you can download the results here: \`aws s3 sync --region us-west-2 ${resultsBucket}/${path} ${path}\``, message.channel);
            } catch(e) {
                slack.sendMessage(`<@${message.user}> failed to deploy ${commit}@${vcid}: ${e}`, message.channel);
            }
        }
    });
})();
