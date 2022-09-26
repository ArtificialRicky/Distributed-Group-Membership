# Setup Instructions

## Clone Repo and Start Servers

Setup repos in all VM,  start servers with our python scripts. 
We have added public keys of vm 8001 to all the other VMs and public keys of all VMs to the github repo, 
so this process won't require any manual password input.
```
# Clone or pull repos on all VM
python3 setup_repo.py

# Run introducer
go run introducer/server.go

# To join the ring, ssh onto another machine and run 
# IN VM 800x
go run initiator/client.go --join

# To join the ring by batch, run the script in 8001
python3 join_ring.py --servers 2 3 4 5 6 7 8 9 10
```

## Other commands

On each machine the following command is supported when a server is running
```
# IN VM 800x
cd ~/mp1

# Leave
go run initiator/client.go --leave
# Show members
go run initiator/client.go --list_mem
# Show self ID
go run initiator/client.go --list_self_ID
```
