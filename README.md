# BlockFS
A distributed record file fystem built on top of blockchain



## How to run miners
go run miner/main.go [path-to-miner-config-json]  
There are sample miner configs in miner/configs  
If you run miner 1 to 6 one by one, they will form a network topology that looks like:

```
--------
|Miner2|
--------
         \
          \
           --------           --------          --------          --------
           |Miner1|   -----   |Miner6|  -----   |Miner5|  -----   |Miner4|
           --------           --------          --------          --------
          /
         /
--------
|Miner3|
--------
```
Note that Miner6 connect two subnets (Subnet Miner1-2-3 and subnet Miner-4-5) if it joins last.


## How to run client
go run client/main.go -c [path-to-miner-config-json] [cmds and args]

Available commands:
* touch

   go run client/main.go -c client/configs/client.json touch fileA  

* append

   go run client/main.go -c client/configs/client.json append fileA record1  

* ls

   go run client/main.go -c client/configs/client.json ls  

* cat

   go run client/main.go -c client/configs/client.json cat fileA  

* head

   go run client/main.go -c client/configs/client.json head fileA 5  

* tail

   go run client/main.go -c client/configs/client.json tail fileA 5  
