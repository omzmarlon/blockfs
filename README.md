# BlockFS
A distributed record file system built on top of blockchain. 

Clients connect to miners to submit file system operations, miners work together 
to build a blockchain to reach distributed consensus on the content of the file system.

## How to run miners
go run cmd/miner/miner.go -c [path-to-miner-config-json]  
There are sample miner configs in `/configs`
If you run miner 1 to 3 one by one, they will form a very simple topology:

```
--------
|Miner2|
--------
|        \
|         \
|          \
|           --------
|           |Miner1|
|           --------
|          /
|         /
|        /
--------
|Miner3|
--------
```

## How to run client
go run cmd/client/client.go -miner=[miner IP:Port address] -cmd=[create|append|list|read|total_recs] [arg0] [arg1]

Examples:
```
go run cmd/client/client.go -miner=127.0.0.1:3000 -cmd=create file0
go run cmd/client/client.go -miner=127.0.0.1:3000 -cmd=append file0 hello_world
go run cmd/client/client.go -miner=127.0.0.1:3000 -cmd=list
go run cmd/client/client.go -miner=127.0.0.1:3000 -cmd=total_recs file0
go run cmd/client/client.go -miner=127.0.0.1:3000 -cmd=read file0 0
```
