run:
	go run *go

test:
	go clean -testcache
	# go test -timeout 30s -run ^TestLoB_UpdateOrAdd_Small$ github.com/bensooraj/h-lob-service/limitorderbook
	# go test -timeout 30s -run ^TestLoB_UpdateOrAdd_Large$ github.com/bensooraj/h-lob-service/limitorderbook