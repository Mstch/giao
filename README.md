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
name        time/op
Std1C-16      11.2µs ±12%
SStd16C-16    17.1µs ± 5%
Stp1C-16      14.0µs ±50%
Stp16C-16     11.1µs ±25%

name        speed
Std1C-16     983MB/s ±11%
SStd16C-16   638MB/s ± 5%
Stp1C-16     897MB/s ±64%
Stp16C-16   1.02GB/s ±23%

name        alloc/op
Std1C-16      5.82kB ± 0%
SStd16C-16    5.86kB ± 0%
Stp1C-16      11.7kB ± 3%
Stp16C-16     17.6kB ±13%

name        allocs/op
Std1C-16        10.0 ± 0%
SStd16C-16      10.0 ± 0%
Stp1C-16        1.00 ± 0%
Stp16C-16      0.60 ±100%


```
