import subprocess


def RunCommandEachMachine(cmd):
    netid = "siyuanc3"
    for i in range(1, 11):
        hostname = "fa22-cs425-80%02d.cs.illinois.edu" % (i,)
        print("On " + hostname + "\n")
        sshProcess = subprocess.Popen(['ssh',
                                "{}@{}".format(netid, hostname),
                                cmd]).communicate()
    

if __name__  == "__main__":
    clone = 'if cd ~/mp2; then git pull && git submodule update --init --recursive; else git clone --recurse-submodules git@gitlab.engr.illinois.edu:siyuan-ruiqi/mp2.git ~/mp2; fi'
    save_to_known_hosts = "ssh-keyscan gitlab.engr.illinois.edu >> ~/.ssh/known_hosts"
    

    # add this to bashrc export PATH="$PATH:$(go env GOPATH)/bin"
    RunCommandEachMachine(cmd = clone)
    build_proto = "cd ~/mp2; protoc --go_out=. --go-grpc_out=. proto/introducer.proto"
    RunCommandEachMachine(cmd = build_proto)
    # RunCommandEachMachine(cmd = save_to_known_hosts)
