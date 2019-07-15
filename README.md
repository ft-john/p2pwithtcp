# Introduce
This is a simple application that implement P2P communication by TCP.

# Useage
After build this application, you will get a executable file "p2pwithtcp", the useage listed below:

```
Usage of p2pwithtcp.exe:
  -port int
        Source port number of a ordinary node (default 8881)
  -tracker
        Run as a tracker mode, if not contain this parameter will run as ordinary mode
```

The parameter **Tracker** determines running mode, there are two modes in this application: **Tracker** and **Ordinary**. In the **Tracker** mode, this application is running as a bootstrap, who used to help other nodes find eachother. The tracker node must have a static public IP address and open the port. In a p2p network, there is at least one and only one tracker node, the tracker node must the first node runs in the network otherwise other ordinary node will cannot startup correctly.
The default port of ordinary node is 8881, you can use **Port** to change it.
You can change tracker address and default port in the code file "node.go" line 37 - 39:
```
const (
	REMOTE_TRACKER_ADDRESS = "127.0.0.1:8880"
	LOCAL_TRACKER_ADDRESS  = "0.0.0.0:8880"
	ORDINARY_NODE_PORT     = 8881
)
```
# 