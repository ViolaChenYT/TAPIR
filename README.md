# Running Unit Test
`go test -run SpecificTestName` in IR or tapir_kv subdirectory

# Running YCSB-T Benchmark 
Inside folder ycsb+t, run `make` to compile the code, if you encounter "stdlib.h not found" error on MacOS, try `export SDKROOT=$(xcrun --sdk macosx --show-sdk-path)`.

Follow the [go-ycsb](https://github.com/pingcap/go-ycsb) instruction to interact with the databse through shell or script, test example: `bin/go-ycsb run tapir -P workloads/workload_test`.
