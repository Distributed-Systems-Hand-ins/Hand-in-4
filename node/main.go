package main

import (
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"math/rand"
	"net"
	pb "ricart-agrawala/proto"
	"sync"
	"time"
)

const (
	Released = "RELEASED"
	Wanted   = "WANTED"
	Held     = "HELD"
)

var clients []pb.MutexClient

var addresses = map[int]string{
	1: "localhost:5001",
	2: "localhost:5002",
	3: "localhost:5003",
}

func main() {
	id := flag.Int("id", 1, "node id")             // default id is 1
	flag.Parse()                                   // parse command line flags
	node := &Node{id: int32(*id), state: Released} // create a new node with the given id
	node.cond = sync.NewCond(&node.mu)                // initialize the condition variable
	fmt.Println("Starting node ", *id)             // print the node id
	listener, err := net.Listen("tcp", addresses[*id])
	if err != nil {
		log.Fatal(err)
	}
	server := grpc.NewServer()
	pb.RegisterMutexServer(server, node)
	go server.Serve(listener)

	for otherId, address := range addresses {
		if otherId == *id {
			continue
		}
		conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatal(err)
		}
		client := pb.NewMutexClient(conn)
		clients = append(clients, client)

	}

	for {
		time.Sleep(time.Duration(1000+rand.Intn(2000)) * time.Millisecond)
		node.mu.Lock()
		node.state = Wanted
		node.clock++
		node.myTicket = node.clock
		fmt.Println("Node", node.id, "is requesting access to the critical section with ticket", node.myTicket)
		ticket := node.myTicket
		node.mu.Unlock()

		for _, client := range clients {
			reply, err := client.RequestAccess(context.Background(), &pb.Request{NodeId: node.id, Timestamp: ticket}, grpc.WaitForReady(true))
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println("Got reply from node", reply.NodeId)

			replyTimestamp := reply.Timestamp
			node.mu.Lock()
			node.clock = max(node.clock, replyTimestamp) + 1
			node.mu.Unlock()
		}
		node.mu.Lock()
		node.state = Held
		node.mu.Unlock()
		log.Println("Node", node.id, "ENTER CS")
		time.Sleep(time.Second)
		log.Println("Node", node.id, "EXIT CS")

		node.mu.Lock()
		node.state = Released
		node.cond.Broadcast()
		node.mu.Unlock()
		
	}
}

func (n *Node) RequestAccess(ctx context.Context, req *pb.Request) (*pb.Reply, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.clock = max(n.clock, req.Timestamp) + 1
	fmt.Println("Got a request from node", req.NodeId, "with timestamp", req.Timestamp, "current clock is", n.clock)
	for n.shouldWait(req) {
		n.cond.Wait()	
	}
	log.Println("Node", n.id, "is granting access to node", req.NodeId)
	return &pb.Reply{NodeId: n.id, Timestamp: n.clock}, nil
}

func (n *Node) shouldWait(req *pb.Request) bool {
	if n.state == Held || (n.state == Wanted && (req.Timestamp > n.myTicket || (req.Timestamp == n.myTicket && req.NodeId > n.id))) {
		return true
	}
	return false
}

type Node struct {
	id int32
	pb.UnimplementedMutexServer
	state    string
	clock    int64
	myTicket int64
	mu       sync.Mutex
	cond	 *sync.Cond
}
