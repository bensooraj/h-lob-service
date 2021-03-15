run:
	go run *go

test:
	go clean -testcache
	# go test -timeout 30s -run ^TestLoB_UpdateOrAdd_Small$ gitlab.com/hooklabs-backend/order-management-system-engine/h-lob-service/limitorderbook
	# go test -timeout 30s -run ^TestLoB_UpdateOrAdd_Large$ gitlab.com/hooklabs-backend/order-management-system-engine/h-lob-service/limitorderbook