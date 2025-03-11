# ZK-PIG Metrics Documentation

This document describes all Prometheus metrics exposed by the ZK-PIG application.

## Generator Metrics

All generator metrics use the subsystem label `generator`.

### Blocks
- **Name**: `generator_blocks`
- **Type**: Gauge Vector
- **Labels**: 
  - `blocknumber`: The block number being processed
- **Description**: Tracks blocks for which the generation of prover input is currently running.

### Generation Time
- **Name**: `generator_generation_time`
- **Type**: Histogram Vector
- **Labels**:
  - `final_step`: The final step reached in generation process
- **Description**: Time spent to generate prover input (in seconds)
- **Buckets**: [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 25, 50, 100, 250, 500]

### Count of Blocks Per Step
- **Name**: `generator_count_of_blocks_per_step`
- **Type**: Gauge Vector
- **Labels**:
  - `step`: The current processing step
- **Description**: Count of blocks for which the generation of prover input is running at each step

### Generation Time Per Step
- **Name**: `generator_time_per_step`
- **Type**: Histogram Vector
- **Labels**:
  - `step`: The processing step
- **Description**: Time spent per step to generate prover input (in seconds)
- **Buckets**: [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 25, 50, 100, 250, 500]

### Generate Error Count
- **Name**: `generator_generate_error_count`
- **Type**: Gauge Vector
- **Labels**:
  - `step`: The step where the error occurred
- **Description**: Count of errors during the generation of prover input

## Daemon Metrics

### Latest Block Number
- **Name**: `generator_latest_block_number`
- **Type**: Gauge
- **Description**: The latest block number seen by the daemon

## Steps

The following steps are tracked in the metrics:
1. `preflight`: Initial validation and data gathering
2. `storePreflightData`: Storing preflight check data
3. `loadPreflightData`: Loading stored preflight data
4. `prepare`: Preparation of prover input
5. `storeProverInput`: Storing the prover input
6. `loadProverInput`: Loading stored prover input
7. `execute`: Execution of the prover
8. `final`: Completion of all steps
9. `error`: Error state

-----

# EthRPC [WiP]

- ChainID Gauge
- HeaderBlock Gauge

# JSON-RPC [WiP]

- Count of JSON-RPC requests GaugeVec (per JSON-RPC method)
- Time per JSON-RPC call HistogramVec (per JSON-RPC method)
- JSON-RPC call error count GaugeVec (per JSON-RPC method)

# Stores [WiP]

- Time per request HistogramVec (per method store/load)
- Request error count GaugeVec (per method store/load)