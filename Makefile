# things to do: 
# download dependencies such as protoc
# dependencies needed seem to include:
# go get google.golang.org/grpc # go source code for grpc lib impl
# go get -u github.com/golang/protobuf/proto # go source code for protbuf lib impl
# go get -u github.com/golang/protobuf/protoc-gen-go # code golang code gen


# protobuf code gen
# generate miner & client binaries

PROTOBUF_DEP := -I ./third_party/protobuf/googleapis -I ./api/protobuf

build: build-miner build-client

build-client: proto-gen
	go mod vendor
	go build -o build/main cmd/client/client.go

build-miner: proto-gen
	go mod vendor
	go build -o build/miner cmd/miner/miner.go

clean:
	find ./pkg/api -name "*.pb.go" -type f -delete
	rm -rf vendor
	rm -rf third_party/protobuf
	rm -rf build/*

proto-gen:       
	if [ ! -d "third_party/protobuf/googleapis" ]; then git clone git@github.com:googleapis/googleapis.git third_party/protobuf/googleapis; fi
	find ./pkg/api -name "*.pb.go" -type f -delete
	protoc $(PROTOBUF_DEP) --go_out=plugins=grpc:. api/protobuf/*.proto

run:                                   
	./build/main