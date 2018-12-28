package main

import (
	"cs416/P1-v3d0b-q4d0b/rfslib"
	"fmt"
	"log"
)

func main() {
	miner_address := "localhost:4001"
	local_ip := "localhost:5001"

	rfs, err := rfslib.Initialize(local_ip, miner_address)
	if err != nil {
		log.Fatal("Failed to initialize rfslib")
	}

	err = rfs.CreateFile("my file")
	if err != nil {
		fmt.Println("CREATE ERROR", err)
	}

	err = rfs.CreateFile("my file")
	if err == nil {
		fmt.Println("SHOULD NOT SEE THIS")
	}

	files, err := rfs.ListFiles()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(files)

	record_string := "hello"
	var record rfslib.Record
	copy(record[:], record_string)
	recordNum, err := rfs.AppendRec("my file", &record)
	fmt.Println(recordNum)

	record_string1 := "there2"
	var record1 rfslib.Record
	copy(record1[:], record_string1)
	recordNum1, err := rfs.AppendRec("my file", &record1)
	fmt.Println(recordNum1)

	recNum, err := rfs.TotalRecs("my file")
	fmt.Println(recNum)

	var record4 rfslib.Record
	rfs.ReadRec("my file", 0, &record4)
	fmt.Println(record4)

	var record3 rfslib.Record
	rfs.ReadRec("my file", 1, &record3)
	fmt.Println(record3)

	fmt.Println("DONE")
}
