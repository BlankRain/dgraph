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
	"fmt"
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
func doQuery(client pb.WorkerClient, q *pb.Query) *pb.Result {
	fmt.Println()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ret, err := client.ServeTask(ctx, q)
	if err != nil {
		log.Fatalf("error in query: %v", err)
		return nil
	}
	log.Printf("%v", ret)
	for i, v := range ret.UidMatrix {
		log.Printf("uid @%d : %v", i, v)
	}
	for i, v := range ret.ValueMatrix {
		log.Printf("value @%d : %v", i, v)
	}
	for i, v := range ret.Counts {
		log.Printf("count @%d : %v", i, v)
	}
	return ret

}
func grpcdemo() {
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
	var ts uint64
	ts = 70099
	q1 := &pb.Query{
		Attr: "name",
		SrcFunc: &pb.SrcFunction{
			Name: "has",
		},

		ReadTs: ts,
	}
	// has(name)
	// Request => parser => subgraph => worker(grpc)  => executionResult => subgraph=> json

	ret := doQuery(client, q1)

	q2 := &pb.Query{
		Attr:      "friend",
		ExpandAll: true,
		UidList:   ret.UidMatrix[0],
		DoCount:   true,
		ReadTs:    ts,
	}
	fmt.Println(ret)
	doQuery(client, q2)
}
