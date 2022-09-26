// Package main implements a server for Introducer service.
package main

import (
	"flag"
	"io"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	fd "cs425.com/mp2/failure_dection"
	intro "cs425.com/mp2/introduce_service"
	pb "cs425.com/mp2/proto_out/introducer"
)

var log_name = flag.String("log", "output.log", "Use command to set output file name")

func main() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	f, err := os.OpenFile(*log_name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	wrt := io.MultiWriter(os.Stdout, f)
	log.SetOutput(wrt)

	intro_server := &intro.Server{}
	intro_server.SelfID = pb.NodeID{
		IP: intro.GetOutboundIP().String(),
		TimeStamp: strconv.FormatInt(time.Now().Unix(), 10)}
	intro_server.Members.PushBack(&intro_server.SelfID)
	intro_server.Position = intro_server.Members.Back()
	go fd.StartUDPServer()
	intro.StartNewServer(intro.JoinPort, intro_server)

}