go test -benchmem  -count 5 -bench . -timeout 1h | tee results && benchstat results

#name          time/op
#Stp1C-8       1.30µs ±130%
#Stp16C-8       7.34µs ± 9%
#SyncStd1C-8    20.8µs ±11%
#SyncStd16C-8   14.5µs ± 7%
#
#name          alloc/op
#Stp1C-8         1.00B ± 0%
#Stp16C-8         742B ±20%
#SyncStd1C-8      842B ± 9%
#SyncStd16C-8     790B ± 0%
#
#name          allocs/op
#Stp1C-8          0.00
#Stp16C-8         2.00 ± 0%
#SyncStd1C-8      12.0 ± 0%
#SyncStd16C-8     12.0 ± 0%
