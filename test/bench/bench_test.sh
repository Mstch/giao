go test -run=None -benchmem  -count 5 -bench . -timeout 1h | tee benchresult && benchstat -sort delta benchresult >benchstatresult
