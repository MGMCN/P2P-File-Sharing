# P2PFileSharing
File sharing in P2P manner through LAN  
> **Warning**: This application may have bugs, please use it with caution. (ps: Although there is still a gap between this application and a real p2p application. But we have at least achieved a distributed file transfer.)
  
![image](https://img.shields.io/github/actions/workflow/status/MGMCN/P2P-File-Sharing/go_test.yml?label=test&logo=github)
[![issue](https://img.shields.io/github/issues/MGMCN/P2P-File-Sharing?logo=github)](https://github.com/MGMCN/P2P-File-Sharing/issues?logo=github)
[![license](https://img.shields.io/github/license/MGMCN/P2P-File-Sharing)](https://github.com/MGMCN/P2P-File-Sharing/blob/main/LICENSE)
![last_commit](https://img.shields.io/github/last-commit/MGMCN/P2P-File-Sharing?color=red&logo=github)
## Application architecture
Peer will start a non-blocking listener service to receive requests from other nodes and pass them to the handler. It will also start a non-blocking service to listen for input from stdin and forward these commands to the handler or CLI for execution. For example, we can request the CLI to check which nodes are online, and it will retrieve the data from the cache and provide the information. Similarly, when we want to download a file, we provide the handler with the information of the file, and it will send a download request on our behalf. Additionally, the handler will be called back when it receives a request from another node. 
It is important to note that the cache contains runtime data, including details about the currently online nodes and files shared by other nodes.  

<img src="image/arch.jpg" width = "50%" height = "50%"/>  

## Usage
Please check out the [tutorial](https://blog.mgmcn.net/posts/development/p2pnode/) on my blog.
## Contributing
Contributions must be available on a separately named branch based on the latest version of the main branch.
