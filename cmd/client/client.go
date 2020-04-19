package main

import (
	"context"
	"log"
	"time"

	pb "github.com/omzmarlon/blockfs/pkg/api"
	"google.golang.org/grpc"
)

func main() {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(":3000", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()
	c := pb.NewBlockFSClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := c.CreateFile(ctx, &pb.CreateFileRequest{FileName: "abc"})
	if err != nil {
		log.Fatalf("Error when calling CreateFile: %s", err)
	}
	log.Printf("Response from server: %s", response)
	log.Printf("Response from server: %d", response.Status.Code)
}
