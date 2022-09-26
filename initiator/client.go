/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a client for Greeter service.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	fd "cs425.com/mp2/failure_dection"
	intro "cs425.com/mp2/introduce_service"
	pb "cs425.com/mp2/proto_out/introducer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

const (
	defaultName = "world"
)

var (
	introducer_addr = flag.String("introducer_addr", "fa22-cs425-8001.cs.illinois.edu", "IP address of initiator")
	is_join = flag.Bool("join", false, "Use command to Join the ring")
	is_list_mem = flag.Bool("list_mem", false, "Use command to display the membership")
	is_list_id = flag.Bool("list_self_id", false, "Use command to display self's id")
	is_leave = flag.Bool("leave", false, "Use command to leave")
	log_name = flag.String("log", "output.log", "Use command to set output file name")
)

func SelfDial() {
	rand.Seed(time.Now().UnixNano())

	conn, err := grpc.Dial("localhost"+ ":" + intro.JoinPort,
					grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewIntroducerClient(conn)
	
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	empty := emptypb.Empty{}
	if *is_list_mem {
		r, err := c.ListMember(ctx, &empty)
		if err != nil {
			log.Fatalf("rpc dail fails %v", err)
		}
		fmt.Println("Members: ")
		for _, member := range r.GetInRingServers() {
			fmt.Printf("%v %v\n", member.IP, member.TimeStamp)
		}
	}

	if *is_list_id {
		r, err := c.ListSelfID(ctx, &empty)
		if err != nil {
			log.Fatalf("rpc dail fails %v", err)
		}
		fmt.Printf("Self's ID: %v %v\n", r.IP, r.TimeStamp) 
	}

	if *is_leave {
		_, err := c.LeaveLocal(ctx, &empty)
		if err != nil {
			log.Fatalf("rpc dail fails %v", err)
		}
	}
}



func main() {
	flag.Parse()
	var err error
	var f *os.File

	if *is_list_mem || *is_list_id || *is_leave{
		f, err = os.OpenFile(*log_name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	} else {
		f, err = os.OpenFile(*log_name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	}

	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	wrt := io.MultiWriter(os.Stdout, f)
	log.SetOutput(wrt)
	
	if *is_list_mem || *is_list_id || *is_leave{
		SelfDial()
	} else if *is_join {
		
		go fd.StartUDPServer()
		// Set up a connection to the server.
		conn, err := grpc.Dial(*introducer_addr + ":" + intro.JoinPort,
						grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		// defer conn.Close()
		c := pb.NewIntroducerClient(conn)
		
		// Contact the server and print out its response.
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		
		localIP := intro.GetOutboundIP()
		unixtime := time.Now().Unix()
		request := &pb.NodeID{
			IP: localIP.String(),
			TimeStamp: strconv.FormatInt(unixtime, 10)}
		r, err := c.Join(ctx, request)

		if err != nil {
			log.Fatalf("rpc dail fails %v", err)
		}
		
		conn.Close()
		acceptjoin_server := &intro.Server{}
		acceptjoin_server.SelfID.IP = request.IP
		acceptjoin_server.SelfID.TimeStamp = request.TimeStamp

		for _, member := range r.GetInRingServers() {
			log.Printf("start with %v %v", member.IP, member.TimeStamp)
			acceptjoin_server.Members.PushBack(member)
			if (member.GetIP() == acceptjoin_server.SelfID.GetIP() && member.GetTimeStamp() == acceptjoin_server.SelfID.GetTimeStamp()) {
				acceptjoin_server.Position = acceptjoin_server.Members.Back()
			}
		}
		
		// quit := make(chan bool)
		// go intro.StartNewServer(intro.JoinPort, acceptjoin_server, &quit)
		intro.StartNewServer(intro.JoinPort, acceptjoin_server)
	}  else {
		log.Fatal("unknown Command\n")
	}
}