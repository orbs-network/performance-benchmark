# Performance Benchmark

> Toolset for measuring the performance of Orbs network

&nbsp;

## V1 Performance Optimization

### Principles

1. Reproducible - the suite should be easily runnable by anyone and automatic as possible

2. Production oriented - use case under test should resemble a real production environment

3. Start simple - behavior of distributed systems is complex and nuanced, aim for the basics first

4. Pareto principle - focus efforts on the easiest 20% that bring 80% of results, ignore the long tail

5. Evidence based - focus on empirical data and numbers only, ignore gut feelings 

### KPIs

1. TPS - maximum number of simple token transfer transactions within a pool of 100K addresses

2. Confirmation time - time from reception of such transaction to reply of the receipt

3. Cost - infrastructure cost for node operation (mostly AWS machine price)

How to combine the multiple KPIs? Limit (2) and (3) to reasonable values (eg. 95% of transactions are committed under 5 seconds on a medium sized AWS machine) and measure the maximum value of (1)

### General workflow

1. Measure a baseline (KPI value on a set git commit and set configuration profile)

2. Extract profiling information during test (from node metrics + golang pprof in production)

3. Analyze profile samples and identify top bottlenecks

4. Propose an optimization (in code or configuration) to improve one of the top bottlenecks

5. Measure the proposal (KPI value on same exact scenario but incorporating proposed change)

6. Accept proposed change if KPI improved

7. Rinse and repeat

&nbsp;

## Scenarios

### Basic

1. Setup a new virtual chain

    * No history (eg. no block persistence)
    * No impact from other virtual chains (eg. prefer not to share a dispatcher)
    * Number of nodes identical to the production scenario
    * Nodes reside in 4-6 popular AWS regions (EU, US, around 100 ms ping between them)
    * AWS machine type is predetermined
    * Code base for production (eg. without Info logs)
    
2. Simulate client traffic

    * Using one of the official client SDKs
    * Generate a significant number of `BenchmarkToken.transfer` transactions
    * All transactions are sent by the contract deployer (owner of all supply) where 1 token is transferred to a random address (from a pool of 100K addresses)
    * Transactions are sent evenly to all gateways (all nodes)
    * Transactions should be sent from multiple machines if the processing rate is faster than the send rate
    
3. Measure main KPIs periodically during the scenario

    * TPS
        * Rely on the metrics system TPS measure
        * Double check with the metrics system measures of total committed transactions (over time)
    * Confirmation time
        * Rely on the metrics system confirmation time measure
        * Double check with client view of this time
    * Cost
        * Predetermined
        
4. Extract profiling information periodically during the scenario

    * Separately from each node (it's enough to sample the nodes)
    * All basic profiling types of Golang (cpu, heap, goroutines, locks, etc.)
    * Core bottleneck metrics from the machines (cpu usage, network usage, etc.)
    
5. Stop the scenario once we have enough stable measurements


## StabilityNet user guide

The Stability network lets us test a long running network. 
Each node exposes JSON metrics with the `/metrics` endpoint, and prints out metrics and errors to logs.
It can optionally print out all logs for the purpose of debugging.  

### Metrics
* Accessible on `http://<ip>/vchains/<vchain_id>/metrics` endpoint

#### Metrics collector
We use a custom metrics collector, see repo `https://github.com/orbs-network/metrics-processor`
The metrics collector asks metrics from every node every 20 seconds, aggregates the data and sends to Geckoboard.
The IP addresses of the nodes are available from an environment variable.
It is written in Node.js.

#### Dashboard
We presently use Geckoboard, this is subject to change.

### Logging
* Enable logs: curl -XPOST http://<ip>/vchains/<vchain_id>/debug/logs/filter-off
* Disable logs: curl -XPOST http://<ip>/vchains/<vchain_id>/debug/logs/filter-off
Logs should not be enabled for long, as they are very verbose.

Logs are sent to logz.io
#### Logz.io configuration
TBD how to send there logs 


### IP Addresses of the nodes (topology)
The IP addresses for the nodes is stored in the file `/opt/galileo/testnet-configuration/benchmark/ips.json` on the client machine.
These IPs must match the actual AWS IPs where the nodes are installed.

### Tampering with IP addresses
For testing purposes, such as preventing a specific node from communicating with other nodes, you can do the following:
* `ssh ec2-user@34.216.213.19`
* `sudo su`
* `cd /opt/galileo/testnet-configuration/benchmark`
* Make a backup of the file `ips.json`
* Modify the file `ips.json`
* Redeploy the app from Slackbot: `deploy <commit> <vchain>` - this will recreate each node's config file based on `ips.json`, then redeploy and restart the nodes, applying the new IP configuration
* When the test is done and you wish to re-enable all nodes, restore the `ips.json` file and redeploy. 

### Updating with new build

* Slackbot: ... not yet
 
go to performance_benchmark project
cd galileo
export API_ENDPOINT=http://18.219.170.177/vchains/2000/api/v1/
export BASE_URL=http://18.219.170.177/vchains/2000
export STRESS_TEST_NUMBER_OF_TRANSACTIONS=100 
export VCHAIN=2000
./extract.sh


### Extracting measurements from live network

* Goroutine stack traces
* Performance metrics

