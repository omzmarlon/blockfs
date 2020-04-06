package flooding

// when received an op, append to queue
// when received a block, append to queue
// when generated a block, flood to others

// need to start a grpc service
// and then maintain a peer list of clients

type Flooder struct {
}
