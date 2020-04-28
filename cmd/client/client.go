package main

import (
	"context"
	"flag"
	"fmt"
	"strconv"
	"time"

	pb "github.com/omzmarlon/blockfs/pkg/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
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
		status := codes.Code(response.Status.Code)
		fmt.Printf("Status: %s\n", codeToString(status))
		if status != codes.OK {
			fmt.Printf("Response from server: %s", response.Status.Message)
		}
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
		status := codes.Code(appendRes.Status.Code)
		fmt.Printf("Status: %s\n", codeToString(status))
		if status != codes.OK {
			fmt.Printf("Response from server: %s", appendRes.Status.Message)
		}
		return
	case "list":
		listResult, err := c.ListFiles(ctx, &pb.ListFilesRequest{})
		if err != nil {
			fmt.Printf("Error when calling ListFile: %s", err)
		}
		status := codes.Code(listResult.Status.Code)
		fmt.Printf("Status: %s\n", codeToString(status))
		if status == codes.OK {
			for _, filename := range listResult.FileNames {
				fmt.Printf("%s, ", filename)
			}
			fmt.Printf("\n")
		} else {
			fmt.Printf("Response from server: %s", listResult.Status.Message)
		}
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
		status := codes.Code(readRec.Status.Code)
		fmt.Printf("Status: %s\n", codeToString(status))
		if status == codes.OK {
			fmt.Printf("%s\n", string(readRec.GetRecord().GetBytes()))
		} else {
			fmt.Printf("Response from server: %s", readRec.Status.Message)
		}
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
		status := codes.Code(totalRecRes.Status.Code)
		fmt.Printf("Status: %s\n", codeToString(status))
		if status == codes.OK {
			fmt.Printf("Total number of records: %d\n", totalRecRes.GetNumRecs())
		} else {
			fmt.Printf("Response from server: %s", totalRecRes.Status.Message)
		}
		return
	default:
		fmt.Printf("Unknown command: %s\n", *cmd)
		return
	}

}

func codeToString(code codes.Code) string {
	switch code {
	case codes.OK:
		return "OK"
	case codes.NotFound:
		return "NOT_FOUND"
	case codes.InvalidArgument:
		return "INVALID_ARGUMENT"
	default:
		return fmt.Sprintf("%d", code)
	}
}
