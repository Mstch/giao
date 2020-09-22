- [x] basic rpc logic
- [ ] goroutine safe off heap memory pool
- [ ] graceful shutdown
- [ ] humanize api



bench result  
"Stp" means stupid,a milestone of the framework  
the test cases are 2,4,8~4k protobuf message  
SyncStd means use net/rpc with default codec  
the detail of this bench please move to [test/bench](test/bench)
```
name          time/op
Stp1C-8        925ns ±39%
Stp16C-8      7.61µs ±30%
SyncStd1C-8   22.6µs ± 9%
SyncStd16C-8  14.1µs ± 2%

name          alloc/op
Stp1C-8        0.00B     
Stp16C-8       24.4B ± 6%
SyncStd1C-8     402B ±64%
SyncStd16C-8    262B ± 0%

name          allocs/op
Stp1C-8         0.00     
Stp16C-8        0.00     
SyncStd1C-8     5.40 ±11%
SyncStd16C-8    5.00 ± 0%


```