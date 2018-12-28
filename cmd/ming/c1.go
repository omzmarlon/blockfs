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

	record_string1 := "there2"
	var record1 rfslib.Record
	copy(record1[:], record_string1)
	_, err = rfs.AppendRec("my file", &record1)
	if err != nil {
		fmt.Println("YOU ARE A DUMBASS")
	}
}
