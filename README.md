- [x] basic rpc logic
- [ ] conn pool
- [ ] graceful shutdown
- [ ] humanize api
- [ ] goroutine safe off heap memory pool

bench result  
"Stp" means stupid,a milestone of the framework  
the test cases are 2,4,8~4k that 343 bytes on average protobuf message  
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
Stp1C-8         10.0µs ±15%
Stp16C-8        7.60µs ±34%
SyncStd1C-8     23.7µs ±13%
SyncStd16C-8    14.8µs ± 2%
Grpc1C-8         113µs ± 7%
Grpc16C-8       25.0µs ± 4%

name          speed
Stp1C-8       69.0MB/s ±14%
Stp16C-8      92.6MB/s ±27%
SyncStd1C-8   29.1MB/s ±13%
SyncStd16C-8  46.2MB/s ± 2%
Grpc1C-8      6.10MB/s ± 7%
Grpc16C-8     27.5MB/s ± 4%

name          alloc/op
Stp1C-8          0.00B     
Stp16C-8         24.0B ± 8%
SyncStd1C-8       342B ± 0%
SyncStd16C-8      278B ± 0%
Grpc1C-8        5.79kB ± 0%
Grpc16C-8       5.88kB ± 0%

name          allocs/op
Stp1C-8           0.00     
Stp16C-8          0.00     
SyncStd1C-8       5.00 ± 0%
SyncStd16C-8      5.00 ± 0%
Grpc1C-8          95.0 ± 0%
Grpc16C-8         96.0 ± 0%


```
