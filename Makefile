run:
	go run src/main.go

protobuf:
	protoc -I calculator/calculator --go_out=calculator/calculator/ calculator/calculator/calculator.proto 
