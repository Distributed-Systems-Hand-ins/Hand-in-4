package main

import (
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
	pb "ricart-agrawala/proto"
)

var addresses = map[int]string{
	1: "localhost:5001",
	2: "localhost:5002",
	3: "localhost:5003",
}

func main() {
	id := flag.Int("id", 1, "node id") // default id is 1
	flag.Parse()                       // parse command line flags
	node := &Node{id: int32(*id)}      // create a new node with the given id
	fmt.Println("Starting node ", *id) // print the node id
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
		reply, err := client.RequestAccess(context.Background(), &pb.Request{NodeId: node.id}, grpc.WaitForReady(true))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Got reply from node", reply.NodeId)
	}

	select {}

}

func (n *Node) RequestAccess(ctx context.Context, req *pb.Request) (*pb.Reply, error) {
	fmt.Println("Got a request from node", req.NodeId)
	return &pb.Reply{NodeId: n.id}, nil
}

type Node struct {
	id int32
	pb.UnimplementedMutexServer
}
