go test -run=StpServer &
echo "$!"
go test -run=None -benchmem -count 5 -bench Stp1C -timeout 1h >benchresult
go test -run=None -benchmem -count 5 -bench Stp16C -timeout 1h >>benchresult
echo "$!"
kill "$!"
sleep 10s

go test -run=StdServer &
echo "$!"
go test -run=None -benchmem -count 5 -bench SyncStd1C -timeout 1h >>benchresult
go test -run=None -benchmem -count 5 -bench SyncStd16C -timeout 1h >>benchresult
echo "$!"
kill "$!"
sleep 10s

go test -run=GRpcServer &
echo "$!"
go test -run=None -benchmem -count 5 -bench Grpc1C -timeout 1h >>benchresult
go test -run=None -benchmem -count 5 -bench Grpc16C -timeout 1h >>benchresult
echo "$!"
kill "$!"

benchstat -sort delta benchresult >benchstatresult
