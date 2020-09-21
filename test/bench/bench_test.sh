go test -bench=Stp1C -benchmem | grep Benchmark
go test -bench=Stp16C -benchmem | grep Benchmark
go test -bench=SyncStd1C -benchmem | grep Benchmark
go test -bench=AStd1C -benchmem | grep Benchmark
go test -bench=ASyncStd16C -benchmem | grep Benchmark
