import subprocess
from pathlib import Path
import argparse
import os
import time

def RunEachVM(cmd, servers):
    netid = "siyuanc3"
    Path("server_logs").mkdir(parents=True, exist_ok=True)
    for i in servers:
        hostname = "fa22-cs425-80%02d.cs.illinois.edu" % (i,)
        print("Start server on " + hostname + "\n")

        stdout_file_name = "server_logs/80%02dstdout.log" % (i,)
        stderr_file_name = "server_logs/80%02dstderr.log" % (i,)
        with open(stdout_file_name,"wb") as out, open(stderr_file_name,"wb") as err:
            sshProcess = subprocess.Popen(['ssh',
                                    "{}@{}".format(netid, hostname),
                                    cmd], stdout=out, stderr=err)
        # wait until server is listening
        while os.path.getsize(stdout_file_name) <= 0:
            time.sleep(0.05)
    

if __name__  == "__main__":
    p = argparse.ArgumentParser()
    p.add_argument('--servers', nargs="*", type=int)
    args = p.parse_args()  

    "protoc --go_out=. --go-grpc_out=. proto/introducer.proto"
    
    # clone = 'if cd ~/mp1; then git pull; else git clone git@gitlab.engr.illinois.edu:siyuan-ruiqi/mp1.git ~/mp1; fi'
    # save_to_known_hosts = "ssh-keyscan gitlab.engr.illinois.edu >> ~/.ssh/known_hosts"
    run_server = "cd ~/mp1; go run server/server.go "
    if args.servers != None:        
        RunEachVM(cmd = run_server, servers=args.servers)
    else:
        RunEachVM(cmd = run_server, servers=list(range(1, 11)))
    # RunCommandEachMachine(cmd = save_to_known_hosts)

