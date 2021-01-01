- [x] basic rpc logic
- [ ] graceful shutdown
- [ ] humanize api
- [ ] goroutine safe off heap memory pool
- [x] client conn pool for a stateless server rpc  
- [ ] add auto batch write(may depends on the memory pool)         

bench result  
"Stp" means stupid,a milestone of the framework  
the test cases are 64~64k bytes that 5462 bytes on average protobuf message  
SyncStd means use net/rpc with default codec and just for Compared   
the detail of this bench please move to [test/bench](test/bench)  
"1C,16C" means open how many client to test  
benchmark platform :
- CPU: 2.2 GHz Intel Core i7
- Memory: 16 GB 1600 MHz DDR3
- Platform: MacBook Pro 2015
- OS: macOS Mojave 10.14.6
- Go: 1.15


benchmark data :
```
name          time/op
Stp16C-16       9.41µs ±21%
Stp1C-16        18.0µs ±55%
Std1C-16        18.3µs ±14%
Std16C-16       22.5µs ± 7%
Rpcx16C-16      23.6µs ±16%
Hprose16C-16    37.8µs ±14%
Hprose1C-16     68.8µs ±10%
Rpcx1C-16        113µs ±11%

name          speed
Stp16C-16     1.13GB/s ±32%
Stp1C-16       651MB/s ±71%
Std1C-16       603MB/s ±15%
Std16C-16      486MB/s ± 7%
Rpcx16C-16     467MB/s ±18%
Hprose16C-16   290MB/s ±13%
Hprose1C-16    159MB/s ±11%
Rpcx1C-16     97.2MB/s ±10%

name          alloc/op
Stp1C-16          535B ±62%
Stp16C-16       1.00kB ±60%
Std1C-16        5.82kB ± 0%
Std16C-16       5.89kB ± 0%
Rpcx1C-16       47.3kB ± 1%
Rpcx16C-16      48.8kB ± 0%
Hprose1C-16     60.8kB ± 0%
Hprose16C-16    61.2kB ± 0%

name          allocs/op
Stp1C-16          0.00
Stp16C-16         0.00
Std1C-16          10.0 ± 0%
Std16C-16         10.0 ± 0%
Rpcx16C-16        27.4 ± 2%
Rpcx1C-16         28.0 ± 0%
Hprose1C-16       34.0 ± 0%
Hprose16C-16      35.0 ± 0%


```
