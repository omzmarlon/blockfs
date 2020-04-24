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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	fileName := "abc"
	response, err := c.CreateFile(ctx, &pb.CreateFileRequest{FileName: fileName})
	if err != nil {
		log.Fatalf("Error when calling CreateFile: %s", err)
	}
	log.Printf("Response from server: %s", response)
	log.Printf("Response from server: %d", response.Status.Code)

	// ----- create same file on purpose ---
	response, err = c.CreateFile(ctx, &pb.CreateFileRequest{FileName: fileName})
	if err != nil {
		log.Fatalf("Error when calling CreateFile: %s", err)
	}
	log.Printf("Response from server: %s", response)
	log.Printf("Response from server: %d", response.Status.Code)
	// ------------------------------------

	listResult, err := c.ListFiles(ctx, &pb.ListFilesRequest{})
	if err != nil {
		log.Fatalf("Error when calling ListFile: %s", err)
	}
	log.Printf("ListFiles Response from server: %s", listResult)
	log.Printf("Response from server: %d", listResult.Status.Code)

	appendRes, err := c.AppendRec(ctx, &pb.AppendRecRequest{FileName: fileName, Record: &pb.Record{Bytes: []byte("hello world")}})
	if err != nil {
		log.Fatalf("Error when calling AppendRec: %s", err)
	}
	log.Printf("AppendRec Response from server: %s", appendRes)
	log.Printf("Response from server: %d", appendRes.Status.Code)

	totalRecRes, err := c.TotalRecs(ctx, &pb.TotalRecsRequest{FileName: fileName})
	if err != nil {
		log.Fatalf("Error when calling TotalRecs: %s", err)
	}
	log.Printf("TotalRecs Response from server: %s", totalRecRes)
	log.Printf("Response from server: %d", totalRecRes.Status.Code)

	readRec, err := c.ReadRec(ctx, &pb.ReadRecRequest{FileName: fileName, RecordNum: 0})
	if err != nil {
		log.Fatalf("Error when calling ReadRec: %s", err)
	}
	log.Printf("TotalRecs Response from server: %s", readRec)
	log.Printf("Response from server: %d", readRec.Status.Code)
}
