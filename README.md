# Running YCSB-T Benchmark 
Run `git submodule update --init` to clone the modified [go-ycsb](https://github.com/pingcap/go-ycsb) repo.

Run `make` to compile the code, if you encounter "stdlib.h not found" error on MacOS, try `export SDKROOT=$(xcrun --sdk macosx --show-sdk-path)`.

Follow the [go-ycsb](https://github.com/pingcap/go-ycsb) instruction to interact with the databse through shell or script, test example: `bin/go-ycsb run tapir -P workloads/workload_test`.