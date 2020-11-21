# Fastfifo
![goreportcard](https://goreportcard.com/badge/github.com/avbasyrov/fastfifo)

## Overview
FIFO queue with fixed size ring buffer and no memory allocations during Push/Pop.

Thread safe.

## Benchmarks
```
BenchmarkFastFifo_MemoryAllocations
BenchmarkFastFifo_MemoryAllocations-8  	59525016	 21.1 ns/op   0 B/op    0 allocs/op
```
