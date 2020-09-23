- [x] basic rpc logic
- [ ] conn pool
- [ ] graceful shutdown
- [ ] humanize api
- [ ] goroutine safe off heap memory pool

bench result  
"Stp" means stupid,a milestone of the framework  
the test cases are 2,4,8~4k protobuf message  
SyncStd means use net/rpc with default codec and just for Compared   
the detail of this bench please move to [test/bench](test/bench)

benchmark platform :
- CPU: 2.2 GHz Intel Core i7
- Memory: 16 GB 1600 MHz DDR3
- Platform: MacBook Pro 2015
- OS: macOS Mojave 10.14.6
- Go: 1.15


benchmark data :
```
name          time/op
Stp1C-8         11.2µs ±13%
Stp16C-8        8.52µs ±58%
SyncStd1C-8     22.9µs ± 7%
SyncStd16C-8    15.3µs ± 2%

name          speed
Stp1C-8       61.9MB/s ±13%
Stp16C-8      85.8MB/s ±40%
SyncStd1C-8   30.0MB/s ± 7%
SyncStd16C-8  45.0MB/s ± 2%

name          alloc/op
Stp1C-8          0.00B     
Stp16C-8         26.6B ± 5%
SyncStd1C-8       330B ± 1%
SyncStd16C-8      262B ± 0%

name          allocs/op
Stp1C-8           0.00     
Stp16C-8          0.00     
SyncStd1C-8       5.00 ± 0%
SyncStd16C-8      5.00 ± 0%

```