- [x] basic rpc logic
- [ ] conn pool
- [ ] graceful shutdown
- [ ] humanize api
- [ ] goroutine safe off heap memory pool

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
Stp1C-8         36.1µs ±63%
Stp16C-8        16.5µs ±35%
SyncStd1C-8     43.7µs ±16%
SyncStd16C-8    20.8µs ± 8%
Grpc1C-8         180µs ±11%
Grpc16C-8       43.0µs ± 6%

name          speed
Stp1C-8       377MB/s ±117%
Stp16C-8       701MB/s ±30%
SyncStd1C-8    253MB/s ±15%
SyncStd16C-8   527MB/s ± 8%
Grpc1C-8      61.2MB/s ±11%
Grpc16C-8      254MB/s ± 5%
name          alloc/op
Stp1C-8          4.40B ±55%
Stp16C-8         30.0B ± 3%
SyncStd1C-8       417B ±51%
SyncStd16C-8      282B ± 1%
Grpc1C-8        23.3kB ± 0%
Grpc16C-8       23.7kB ± 0%

name          allocs/op
Stp1C-8           0.00     
Stp16C-8          0.00     
SyncStd1C-8       5.40 ±11%
SyncStd16C-8      5.00 ± 0%
Grpc1C-8          96.0 ± 0%
Grpc16C-8         98.0 ± 0%


```
