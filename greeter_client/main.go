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
	"log"
	"time"

	"github.com/dgraph-io/dgraph/protos/pb"
	"google.golang.org/grpc"
)

const (
	address = "localhost:7080"
)

func newList(data []uint64) *pb.List {
	return &pb.List{Uids: data}
}
func doQuery(client pb.WorkerClient, q *pb.Query) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ret, err := client.ServeTask(ctx, q)
	if err != nil {
		log.Fatalf("error in query: %v", err)
		return
	}
	// log.Printf("%v", ret)
	for i, v := range ret.UidMatrix {
		log.Printf("%d %v", i, v)
	}
	for i, v := range ret.ValueMatrix {
		log.Printf("value :%d %v", i, v)
	}

}
func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	log.Printf("%s", "connected.")
	client := pb.NewWorkerClient(conn)
	input := []*pb.List{
		newList([]uint64{1, 3, 6, 8, 10}),
		newList([]uint64{2, 4, 5, 7, 15}),
	}
	log.Printf("%s", input)

	q1 := &pb.Query{
		Attr:      "name",
		ExpandAll: true,
		UidList:   newList([]uint64{0x1, 2, 3, 4}),
		ReadTs:    70086,
	}
	q2 := &pb.Query{
		Attr: "name",
		SrcFunc: &pb.SrcFunction{
			Name: "has",
		},
		ReadTs: 70086,
	}
	doQuery(client, q1)
	doQuery(client, q2)

	// c := pb.NewGreeterClient(conn)

	// // Contact the server and print out its response.
	// name := defaultName
	// if len(os.Args) > 1 {
	// 	name = os.Args[1]
	// }
	// ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// defer cancel()
	// r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
	// if err != nil {
	// 	log.Fatalf("could not greet: %v", err)
	// }
	// log.Printf("Greeting: %s", r.Message)
}
