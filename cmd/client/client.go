package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/omzmarlon/blockfs/pkg/rfslib"
)

var (
	Trace = log.New(os.Stdout, "[TRACE] ", 0)
	Error = log.New(os.Stderr, "[ERROR] ", 0)
)

func main() {
	configPath := flag.String("c", "./client/configs/client.json", "client config")
	flag.Parse()

	Trace.Printf("Using client config at: %s", *configPath)
	config, cerr := getClientConfig(*configPath)
	if cerr != nil {
		Error.Printf("Could not resolve client config: %s", cerr)
		os.Exit(1)
	}
	rfs, rfserr := rfslib.Initialize(config.ClientIPPort, config.MinerIPPort)
	if rfserr != nil {
		Error.Printf("RFS client initialization failed: %s", rfserr)
		os.Exit(1)
	}

	args := flag.Args()
	cmd := args[0]
	args = args[1:]
	switch strings.ToLower(cmd) {
	case "append":
		fname := args[0]
		recordStr := args[1]
		if len(args) != 2 {
			Error.Println("Incorrect number of arguments")
			os.Exit(1)
		}
		append(rfs, fname, recordStr)
	case "cat":
		if len(args) != 1 {
			Error.Println("Incorrect number of arguments")
			os.Exit(1)
		}
		fname := args[0]
		cat(rfs, fname)
	case "head":
		if len(args) != 2 {
			Error.Println("Incorrect number of arguments")
			os.Exit(1)
		}
		fname := args[0]
		k, aerr := strconv.Atoi(args[1])
		if aerr != nil {
			Error.Println("K must be an integer")
			os.Exit(1)
		}
		head(rfs, fname, k)
	case "ls":
		ls(rfs)
	case "tail":
		if len(args) != 2 {
			Error.Println("Incorrect number of arguments")
			os.Exit(1)
		}
		fname := args[0]
		k, aerr := strconv.Atoi(args[1])
		if aerr != nil {
			Error.Println("K must be an integer")
			os.Exit(1)
		}
		tail(rfs, fname, k)
	case "touch":
		fname := args[0]
		if len(args) != 1 {
			Error.Println("Incorrect number of arguments")
			os.Exit(1)
		}
		touch(rfs, fname)
	default:
		Error.Printf("Unsupported command: %s", cmd)
	}
}

func getClientConfig(configPath string) (*ClientConfig, error) {
	rawConf, rerr := ioutil.ReadFile(configPath)
	if rerr != nil {
		return nil, rerr
	}
	config := &ClientConfig{}
	jerr := json.Unmarshal(rawConf, config)
	if jerr != nil {
		return nil, jerr
	}
	return config, nil
}

func append(rfs rfslib.RFS, fname string, recordStr string) {
	var record rfslib.Record

	copy(record[:], recordStr)
	recordNum, err := rfs.AppendRec(fname, &record)
	if err != nil {
		Error.Printf("Failed to append %s to file %s, for error: %s\n", recordStr, fname, err)
		os.Exit(1)
	}
	Trace.Printf("Succesfully appended a record to file %s at index %d", fname, recordNum)
}

func touch(rfs rfslib.RFS, fname string) {
	crerr := rfs.CreateFile(fname)
	if crerr != nil {
		Error.Println("Failed to create file: ", fname, "  ", crerr)
		os.Exit(1)
	}
	Trace.Println("Successfully created file:", fname)
}

func cat(rfs rfslib.RFS, fname string) {
	numRecs, recerr := rfs.TotalRecs(fname)
	if recerr != nil {
		Error.Print("Failed to obtain total number of records for file: ", fname, "  ", recerr)
		os.Exit(1)
	}

	Trace.Printf("Contents of %s:", fname)
	var i uint16
	for i = 0; i < numRecs; i++ {
		var record rfslib.Record
		recerr = rfs.ReadRec(fname, i, &record)
		if recerr != nil {
			Error.Print("Failed to obtain record %d for %s\n", i, fname, "  ", recerr)
			continue
		}
		Trace.Println(string(record[:]))
	}
}

func ls(rfs rfslib.RFS) {
	flist, lserr := rfs.ListFiles()
	if lserr != nil {
		Error.Println("Failed to obtain list of files", "  ", lserr)
		os.Exit(1)
	}

	Trace.Printf("Retrieve %d files:", len(flist))
	for _, fname := range flist {
		numRecs, recerr := rfs.TotalRecs(fname)
		if recerr != nil {
			Error.Println("Failed to obtain total number of records for: ", fname, "  ", recerr)
			continue
		}
		Trace.Printf("File name: %s, number of records: %d", fname, numRecs)
	}
}

func head(rfs rfslib.RFS, fname string, k int) {
	numRecs, recerr := rfs.TotalRecs(fname)
	if recerr != nil {
		Error.Println("Failed to obtain total number of records for: ", fname, "  ", recerr)
		os.Exit(1)
	}

	var i uint16

	for i = 0; i < numRecs; i++ {
		if i < uint16(k) {
			var record rfslib.Record
			readerr := rfs.ReadRec(fname, i, &record)
			if readerr != nil {
				Error.Print("Failed to obtain record %d for %s\n", i, fname, "  ", readerr)
				continue
			}
			Trace.Println(string(record[:]))
		}
	}
}

func tail(rfs rfslib.RFS, fname string, k int) {
	numRecs, recerr := rfs.TotalRecs(fname)
	if recerr != nil {
		Error.Println("Failed to obtain total number of records for file: ", fname, "  ", recerr)
		os.Exit(1)
	}

	for i := max(0, int(numRecs)-k); i < int(numRecs); i++ {
		var record rfslib.Record
		err := rfs.ReadRec(fname, uint16(i), &record)
		if err != nil {
			Error.Print("Failed to obtain record %d for %s\n", i, fname, "  ", err)
		}
		Trace.Println(string(record[:]))
	}
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

type ClientConfig struct {
	MinerIPPort  string
	ClientIPPort string
}
