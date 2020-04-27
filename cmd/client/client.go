package main

import (
	"context"
	"flag"
	"fmt"
	"strconv"
	"time"

	pb "github.com/omzmarlon/blockfs/pkg/api"
	"google.golang.org/grpc"
)

func main() {
	miner := flag.String("miner", "", "Address of miner to connect to")
	cmd := flag.String("cmd", "", "The command to issue to the miner")

	flag.Parse()
	tailArgs := flag.Args()

	var conn *grpc.ClientConn
	conn, err := grpc.Dial(*miner, grpc.WithInsecure())
	if err != nil {
		fmt.Printf("Could not connect to miner due to err: %s", err)
	}
	defer conn.Close()
	c := pb.NewBlockFSClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	switch *cmd {
	case "create":
		if len(tailArgs) != 1 {
			fmt.Println("create command expects a file name as argument")
			return
		}
		fileName := tailArgs[0]
		fmt.Printf("Creating a new file: %s\n", fileName)
		response, err := c.CreateFile(ctx, &pb.CreateFileRequest{FileName: fileName})
		if err != nil {
			fmt.Printf("Error when calling CreateFile: %s\n", err)
		}
		fmt.Printf("Response from server: %s", response)
		fmt.Printf("Response from server: %d", response.Status.Code)
		return
	case "append":
		if len(tailArgs) != 2 {
			fmt.Println("append command expects a file name and a string as arguments")
			return
		}
		fileName := tailArgs[0]
		data := tailArgs[1]
		appendRes, err := c.AppendRec(ctx, &pb.AppendRecRequest{FileName: fileName, Record: &pb.Record{Bytes: []byte(data)}})
		if err != nil {
			fmt.Printf("Error when calling AppendRec: %s\n", err)
		}
		fmt.Printf("AppendRec Response from server: %s\n", appendRes)
		fmt.Printf("Response from server: %d\n", appendRes.Status.Code)
		return
	case "list":
		listResult, err := c.ListFiles(ctx, &pb.ListFilesRequest{})
		if err != nil {
			fmt.Printf("Error when calling ListFile: %s", err)
		}
		fmt.Printf("ListFiles Response from server: %s\n", listResult)
		fmt.Printf("Response from server: %d\n", listResult.Status.Code)
		return
	case "read":
		if len(tailArgs) != 2 {
			fmt.Println("append command expects a file name and an index as arguments")
			return
		}
		fileName := tailArgs[0]
		index, err := strconv.Atoi(tailArgs[1])
		readRec, err := c.ReadRec(ctx, &pb.ReadRecRequest{FileName: fileName, RecordNum: uint32(index)})
		if err != nil {
			fmt.Printf("Error when calling ReadRec: %s", err)
		}
		fmt.Printf("TotalRecs Response from server: %s", readRec)
		fmt.Printf("Response from server: %d", readRec.Status.Code)
	// case "read_all":
	// TODO
	case "total_recs":
		if len(tailArgs) != 1 {
			fmt.Println("total_recs command expects a file name as argument")
			return
		}
		fileName := tailArgs[0]
		fmt.Printf("Getting total number of recs in file %s\n", fileName)
		totalRecRes, err := c.TotalRecs(ctx, &pb.TotalRecsRequest{FileName: fileName})
		if err != nil {
			fmt.Printf("Error when calling TotalRecs: %s", err)
		}
		fmt.Printf("TotalRecs Response from server: %s", totalRecRes)
		fmt.Printf("Response from server: %d", totalRecRes.Status.Code)
		return
	default:
		fmt.Printf("Unknown command: %s\n", *cmd)
		return
	}

}
