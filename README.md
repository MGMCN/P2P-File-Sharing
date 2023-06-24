# P2PFileSharing
File sharing in P2P manner through LAN
## Usage
Build p2pnode from source code.
```bash
$ go get -t github.com/libp2p/go-libp2p@v0.28.0   
$ go mod tidy
$ go build -o p2pnode
```
Enter ```-help``` option to view usage.
```bash
$ ./p2pnode -help
Peer-to-peer file sharing over LAN.
  -help
    	Display Help
  -host string
    	The bootstrap node host listen address
    	 (default "0.0.0.0")
  -port int
    	node listen port
    	 (default 6666)
  -rendezvous string
    	Unique string to identify group of nodes. Share this with your friends to let them connect with you
    	 (default "default")
  -src string
    	Path to shared directory
    	 (default "./")
```
## Example
Create a directory for peer1 and store a test file.
```bash
$ ls
.
├── p2pnode
└── peer1.txt
$ cat peer1.txt
hello from peer1
```
Create a directory for peer2 and store a test file.
```bash
$ ls
.
├── p2pnode
└── peer2.txt
$ cat peer2.txt
hello from peer2
```
Execute the following command in each node's directory to start the two nodes respectively, it should be noted that the ports cannot conflict.
```bash
$ ./p2pnode -port 6667 # 6667 for peer1 6668 for peer2
2023/06/24 22:34:07 Peer listening on: 0.0.0.0 with port: 6667
```
The ```peer search``` parameter will search for resources of all online nodes in the current LAN.
```bash
2023/06/24 22:34:07 Peer listening on: 0.0.0.0 with port: 6667
peer search
2023/06/24 22:34:25 UpdateOthersSharedResources from QmUqAQ2MKJCcyzF1pKoZdNFcKsj35nk5EYdVsgrtRqpRoL
```
The ```cache list``` parameter allows you to view the details of the shared resources in the current LAN.
```bash
2023/06/24 22:34:07 Peer listening on: 0.0.0.0 with port: 6667
peer search
2023/06/24 22:34:25 UpdateOthersSharedResources from QmUqAQ2MKJCcyzF1pKoZdNFcKsj35nk5EYdVsgrtRqpRoL
cache list
2023/06/24 22:34:38 We share the following resources: | p2pnode ( 30000930 bytes ) | peer1.txt ( 17 bytes )
2023/06/24 22:34:38 Resource             | Size           | Peers
2023/06/24 22:34:38 peer2.txt            | 17 bytes       | [QmUqAQ2MKJCcyzF1pKoZdNFcKsj35nk5EYdVsgrtRqpRoL]
2023/06/24 22:34:38 p2pnode              | 30000930 bytes | [QmUqAQ2MKJCcyzF1pKoZdNFcKsj35nk5EYdVsgrtRqpRoL]
```
The ```peer download resourceName``` parameter can download the shared resources within the current LAN, and if there are multiple nodes with the same resource, the resources will be downloaded from multiple nodes in parallel.
```bash
2023/06/24 22:34:07 Peer listening on: 0.0.0.0 with port: 6667
peer search
2023/06/24 22:34:25 UpdateOthersSharedResources from QmUqAQ2MKJCcyzF1pKoZdNFcKsj35nk5EYdVsgrtRqpRoL
cache list
2023/06/24 22:34:38 We share the following resources: | p2pnode ( 30000930 bytes ) | peer1.txt ( 17 bytes )
2023/06/24 22:34:38 Resource             | Size           | Peers
2023/06/24 22:34:38 peer2.txt            | 17 bytes       | [QmUqAQ2MKJCcyzF1pKoZdNFcKsj35nk5EYdVsgrtRqpRoL]
2023/06/24 22:34:38 p2pnode              | 30000930 bytes | [QmUqAQ2MKJCcyzF1pKoZdNFcKsj35nk5EYdVsgrtRqpRoL]
peer download peer2.txt
2023/06/24 22:34:49 Caculated fileChunkSize:17 bytes
2023/06/24 22:34:49 Received file chunk from QmUqAQ2MKJCcyzF1pKoZdNFcKsj35nk5EYdVsgrtRqpRoL
2023/06/24 22:34:49 Merge chunk of peer2.txt successfully
```
Now we go back to the peer1 directory. You can see that peer2.txt has been downloaded successfully.
```bash
$ ls
.
├── p2pnode
├── peer1.txt
└── peer2.txt
$ cat peer2.txt
hello from peer2
```
## Warning
The code may have bugs, please use it with caution. However, I am continuously improving this code.