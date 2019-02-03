# Performance Benchmark

> Toolset for measuring the performance of Orbs network

&nbsp;

## V1 Performance Optimization

### Principles

1. Reproducable - the suite should be easily runnable by anyone and automatic as possible

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

## Scenario 1

1. Setup a new virtual chain

    * No history (eg. no block persistence)
    * No impact from other virtual chains (eg. prefer not to share a dispatcher)
    * Number of nodes identical to the production scenario
    * Nodes reside in 4-6 popular AWS regions (EU, US, around 100 ms ping between them)
    * AWS machine type is pre determined
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
