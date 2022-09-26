package failure_detection

import (
	"log"
	"math/rand"
	"net"
	"sync"
	"time"

	pb "cs425.com/mp2/proto_out/introducer"
)

var DetectPort = "50061"
var UDPBufSize = 128

var UDPTimeOut = 400 // ms
var TimeOutCap = 3

var LossRate = -0.1

func Ping(target * pb.NodeID, wg * sync.WaitGroup, id_to_fail_cnt map[string]int, ping_msg string) error {

	success := false
	ping_addr := target.GetIP() + ":" + DetectPort
	s, err := net.ResolveUDPAddr("udp4", ping_addr)
	// laddr = nil -> local address will be chosen
	conn, err := net.DialUDP("udp4", nil, s)
    if err != nil {
        log.Printf("Fail to Dail %s\n", ping_addr )
		return err
    }
	

	data := []byte(ping_msg)
	rand_n := rand.Float64()
	if (rand_n > LossRate) {
		_, err = conn.Write(data)
		
		if err != nil {
			log.Printf("fail to write");
		}
	} else {
		log.Printf("Packet Loss in write during ping! %v", rand_n)
	}

	buffer := make([]byte, UDPBufSize)

	deadline := time.Now().Add(time.Duration(UDPTimeOut) * time.Millisecond)
	// conn.SetReadDeadline(deadline)
	err = conn.SetDeadline(deadline)
	
	_, addr, err := conn.ReadFromUDP(buffer)
	if err != nil {
			log.Printf("ReadFromUDP error %v from %v", err, addr);
	} else {
		// do something with packet her
		success = true
		// log.Printf("Reply: %s\n", string(buffer[0:n]))
	}

	id := target.GetIP() + ":" +target.GetTimeStamp()
	if success {
		log.Printf("ping %s succeed", target.GetIP())
		id_to_fail_cnt[id] = 0
	} else {
		log.Printf("ping %s fail", target.GetIP())
		id_to_fail_cnt[id] = id_to_fail_cnt[id] + 1
	}

	wg.Done()
	return nil
}

func StartUDPServer () error {
	s, err := net.ResolveUDPAddr("udp4", ":" + DetectPort)
	if err != nil {
		log.Fatalf("Fail to resolve Addr: %v", err)
		return err
	}
	connection, err := net.ListenUDP("udp4", s)
	if err != nil {
		log.Fatalf("Fail to ListenUDP: %v", err)
		return err
	}
	log.Printf("UDP Listening to %v", s)
	for {
		buffer := make([]byte, UDPBufSize)
		_, addr, err := connection.ReadFromUDP(buffer)
		// fmt.Printf("%v %v %v", n, addr, err)
		// if (err == nil) {
		// 	fmt.Print("--> ", string(buffer[0:n]))
		// }
		
		if ( rand.Float64() > LossRate) {
			data := []byte("ACK")
			_, err = connection.WriteToUDP(data, addr)
			if err != nil {
				log.Fatalf("Fail to write %v", err)
				return err
			}
		} else {
			log.Printf("Packet Loss in write during respond!")
		}
	}
	defer connection.Close()
	return nil
}
