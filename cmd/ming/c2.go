package main

import (
	"cs416/P1-v3d0b-q4d0b/rfslib"
	"fmt"
	"log"
)

func main() {
	miner_address := "localhost:4002"
	local_ip := "localhost:5001"

	rfs, err := rfslib.Initialize(local_ip, miner_address)
	if err != nil {
		log.Fatal("Failed to initialize rfslib")
	}

	files, err := rfs.ListFiles()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(files)
}
