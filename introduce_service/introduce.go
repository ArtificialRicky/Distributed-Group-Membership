// Package main implements a server for Introducer service.
package introduce_service

import (
	"container/list"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	fd "cs425.com/mp2/failure_dection"
	pb "cs425.com/mp2/proto_out/introducer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

var JoinPort = "50051"
var PingPeriod = 750 // ms
// server is used to implement helloworld.IntroducerServer.
type Server struct {
	pb.UnimplementedIntroducerServer
	Members list.List
	SelfID pb.NodeID
	channel chan os.Signal
	Position *list.Element
}

func (s *Server) GetNeighbours() []*pb.NodeID {

	var out[] *pb.NodeID
	if (s.Members.Len() <= 5) {
		// for loop add everything except itself
		for m := s.Members.Front(); m != nil; m = m.Next() {
			if m != s.Position {
				out = append(out, m.Value.(*pb.NodeID))
			}
		}
	} else {
		// prev garanteed to be non empty
		var prev, prev_prev, next, next_next *list.Element
		if s.Position.Prev() != nil {	
			prev = s.Position.Prev()
		} else {
			prev = s.Members.Back()
		}
		out = append(out, prev.Value.(*pb.NodeID))

		// prev_prev garanteed to be non empty
		if (prev.Prev() != nil) {
			prev_prev = prev.Prev()
		} else {
			prev_prev = s.Members.Back()
		}
		out = append(out, prev_prev.Value.(*pb.NodeID))

		// next garanteed to be non empty
		if s.Position.Next() != nil {
			next = s.Position.Next()
		} else {
			next = s.Members.Front()
		}
		out = append(out, next.Value.(*pb.NodeID))

		// next_next garanteed to be non empty
		if next.Next() != nil {
			next_next = next.Next()
		} else {
			next_next = s.Members.Front()
		}
		out = append(out, next_next.Value.(*pb.NodeID))
	}
	return out
}

func NotifyExistedMember(existed *pb.NodeID,  in *pb.NodeID, cmd string) error {
	conn, err := grpc.Dial(existed.GetIP()+ ":" + JoinPort,
					grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewIntroducerClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()
	
	if (cmd == "accept_join") {
		r, err := c.AcceptJoin(ctx, in)
		if err != nil {
			log.Fatalf("AcceptJoin rpc dail fails: %v target=%v %v", err, existed.GetIP(), existed.GetTimeStamp())
		}
		log.Printf("%s", r.GetReplyMessage())
	} else if (cmd == "leave") {
		r, err := c.LeaveRequest(ctx, in)
		if err != nil {
			log.Printf("LeaveRequest rpc dail fails: %v target=%v %v", err, existed.GetIP(), existed.GetTimeStamp())
		}
		log.Printf("%s", r.GetReplyMessage())
	} else {
		log.Fatalf("NotifyExistedMember unknown command %v", cmd)
	}
	return nil
}

// borrowed from https://stackoverflow.com/a/37382208/10284186
func GetOutboundIP() net.IP {
    conn, err := net.Dial("udp", "8.8.8.8:80")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    localAddr := conn.LocalAddr().(*net.UDPAddr)

    return localAddr.IP
}

func (s *Server) Join(ctx context.Context, in *pb.NodeID) (*pb.JoinReply, error) {
	
	log.Printf("Join Received: %v %v", in.GetIP(), in.GetTimeStamp())
	
	for m:= s.Members.Front(); m != nil; m = m.Next() {
		member := m.Value.(*pb.NodeID)
		if member.IP != s.SelfID.IP || member.TimeStamp != s.SelfID.TimeStamp {
			NotifyExistedMember(member, in, "accept_join")
		}
	}

	s.Members.PushBack(in)

	reply := &pb.JoinReply{}
	for m:= s.Members.Front(); m != nil; m = m.Next() {
		member := m.Value.(*pb.NodeID)
		log.Println(member)
		reply.InRingServers = append(reply.InRingServers, member)
	}

	return reply, nil
}

func (s *Server) AcceptJoin(ctx context.Context, in *pb.NodeID) (*pb.MemberChangeReply, error) {
	
	log.Printf("AcceptJoin Received: %v %v", in.GetIP(), in.GetTimeStamp())
	s.Members.PushBack(in)

	for m:= s.Members.Front(); m != nil; m = m.Next() {
		member := m.Value.(*pb.NodeID)
		log.Println(member)
	}

	reply := &pb.MemberChangeReply{ReplyMessage:"AcceptJoin ACK + from " + s.SelfID.IP + " " +s.SelfID.TimeStamp}
	return reply, nil
}

func (s *Server) ListMember(ctx context.Context, in *emptypb.Empty) (*pb.JoinReply, error) {
	
	log.Printf("ListMember Received");

	reply := &pb.JoinReply{}
	for m:= s.Members.Front(); m != nil; m = m.Next() {
		member := m.Value.(*pb.NodeID)
		reply.InRingServers = append(reply.InRingServers, member)
	}
	
	return reply, nil
}

func (s *Server) ListSelfID(ctx context.Context, in *emptypb.Empty) (*pb.NodeID, error) {
	
	log.Printf("List Self's ID");

	reply := &pb.NodeID{}
	reply.IP = s.SelfID.IP
	reply.TimeStamp = s.SelfID.TimeStamp

	return reply, nil
}

func (s *Server) LeaveLocal(ctx context.Context, in *emptypb.Empty) ( *emptypb.Empty, error) {
	
	log.Printf("LeaveLocal Received");
	s.channel <- syscall.SIGTERM

	for m:= s.Members.Front(); m != nil; m = m.Next() {
		member := m.Value.(*pb.NodeID)
		if member.IP != s.SelfID.IP || member.TimeStamp != s.SelfID.TimeStamp {
			NotifyExistedMember(member, &s.SelfID , "leave")
		}
	}
	
	return &emptypb.Empty{}, nil
}


// Handle Leave request from active user
func (s *Server) LeaveRequest(ctx context.Context, in *pb.NodeID) ( *pb.MemberChangeReply, error) {
	
	log.Printf("LeaveRequest Received %v %v", in.GetIP(), in.GetTimeStamp());

	var next *list.Element
	for m:= s.Members.Front(); m != nil; m = next {
		member := m.Value.(*pb.NodeID)
		next = m.Next()
		if member.IP == in.IP && member.TimeStamp == in.TimeStamp {
			s.Members.Remove(m)
		}
	}

	log.Printf("Members after removal %v %v", in.GetIP(), in.GetTimeStamp());
	for m:= s.Members.Front(); m != nil; m = m.Next() {
		member := m.Value.(*pb.NodeID)
		log.Printf("%v %v", member.GetIP(), member.GetTimeStamp())
	}
	
	reply := &pb.MemberChangeReply{ReplyMessage:"LeaveRequest ACK + from " + s.SelfID.IP + " " +s.SelfID.TimeStamp}
	return reply, nil
}

func (s *Server) RemoveFailedNode(toRemove *pb.NodeID) error {
	var next *list.Element
	for m:= s.Members.Front(); m != nil; m = next {
		member := m.Value.(*pb.NodeID)
		next = m.Next()
		if member.IP == toRemove.IP && member.TimeStamp == toRemove.TimeStamp {
			s.Members.Remove(m)
		}
	}

	for m:= s.Members.Front(); m != nil; m = m.Next() {
		member := m.Value.(*pb.NodeID)
		if member.IP != toRemove.IP || member.TimeStamp != toRemove.TimeStamp {
			NotifyExistedMember(member, toRemove , "leave")
		}
	}
	
	return nil
}


func (s *Server) RegularPing () {
	id_to_fail_cnt := make(map[string]int)
	var ping_msg string
	for {
		out := s.GetNeighbours()
		wg := sync.WaitGroup{}
		log.Printf("Start pinging")

		for _, member := range out {
			id := member.GetIP() + ":" + member.GetTimeStamp()
			ping_msg = "ping"

			if val, ok := id_to_fail_cnt[id]; !ok {
				//do something here
				id_to_fail_cnt[id] = 0
			} else {
				if val >= fd.TimeOutCap {
					log.Printf("%s Failed for %d times", id, val)
					log.Printf("Send out LeaveRequest")
					s.RemoveFailedNode(member)
					delete(id_to_fail_cnt, id)
				}
				
			} 

			wg.Add(1)
			// log.Printf("Ping Neighbor %v %v", member.GetIP(), member.GetTimeStamp() )
			go fd.Ping(member, &wg, id_to_fail_cnt, ping_msg)
		}

		time.Sleep(time.Duration(PingPeriod) * time.Millisecond)
		wg.Wait()
	}
}


func StartNewServer(join_port string, intro_server *Server) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", join_port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpc_server := grpc.NewServer()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	wg := sync.WaitGroup{}
	wg.Add(1)
	
	go func() {
		s := <-c
		log.Printf("got signal %v, attempting graceful shutdown", s)
		if (s == syscall.SIGTERM || s == os.Interrupt) {
			grpc_server.GracefulStop()
		}
		wg.Done()
	}()

	intro_server.channel = c

	pb.RegisterIntroducerServer(grpc_server, intro_server)	
	
	go intro_server.RegularPing()
	// go fd.StartUDPServer()

	log.Printf("server listening at %v", lis.Addr())
	if err := grpc_server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}


	wg.Wait()
	log.Println("clean shutdown")
}
