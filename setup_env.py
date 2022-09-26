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
    # clone = 'if cd ~/mp2; then git pull && git submodule update --init --recursive; else git clone --recurse-submodules git@gitlab.engr.illinois.edu:siyuan-ruiqi/mp2.git ~/mp2; fi'
    # save_to_known_hosts = "ssh-keyscan gitlab.engr.illinois.edu >> ~/.ssh/known_hosts"
    cmd = """
    PB_REL=\"https://github.com/protocolbuffers/protobuf/releases\" && 
    curl -LO $PB_REL/download/v3.15.8/protoc-3.15.8-linux-x86_64.zip &&
    unzip protoc-3.15.8-linux-x86_64.zip -d $HOME/.local &&
    go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28 &&
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2 &&
    """ 

    add_local_bin = """
    echo 'export PATH="$PATH:$HOME/.local/bin"' >> ~/.bashrc
    """
    add_go_path_bin = """   
        echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.bashrc 
    """

    remove_last_four_lines = "head -n -4 ~/.bashrc > tmp.txt && mv tmp.txt ~/.bashrc"
    # add this to bashrc export PATH="$PATH:$(go env GOPATH)/bin"
    RunCommandEachMachine(cmd = cmd)
    RunCommandEachMachine(cmd = add_local_bin)
    RunCommandEachMachine(cmd = add_go_path_bin)
    # RunCommandEachMachine(cmd = save_to_known_hosts)

