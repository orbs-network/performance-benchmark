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
