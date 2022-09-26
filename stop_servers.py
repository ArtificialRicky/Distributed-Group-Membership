import subprocess
import argparse

def RunCommandEachMachine(cmd, servers):
    netid = "siyuanc3"
    for i in servers:
        hostname = "fa22-cs425-80%02d.cs.illinois.edu" % (i,)
        print("On " + hostname + "\n")
        sshProcess = subprocess.Popen(['ssh',
                                "{}@{}".format(netid, hostname),
                                cmd]).communicate()
    

if __name__  == "__main__":
    p = argparse.ArgumentParser()
    p.add_argument('--servers', nargs="*", type=int)
    args = p.parse_args()

    stop_server = 'kill -9 `pgrep client`'
    print(args.servers)
    if args.servers != None:        
        RunCommandEachMachine(cmd = stop_server, servers=args.servers)
    else:
        RunCommandEachMachine(cmd = stop_server, servers=list(range(1, 11)))
    # RunCommandEachMachine(cmd = save_to_known_hosts)

