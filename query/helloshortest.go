/*
 * Copyright 2017-2018 Dgraph Labs, Inc. and Contributors
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
 */

package query

import (
	"container/heap"
	"context"
	"fmt"
	"math"

	"github.com/dgraph-io/dgraph/protos/pb"
	"github.com/dgraph-io/dgraph/x"
)

func HelloShortestPath(ctx context.Context, sg *SubGraph) ([]*SubGraph, error) {
	var err error
	if sg.Params.Alias != "helloshortest" {
		return nil, x.Errorf("Invalid hello shortest path query")
	}
	// numPaths := 1
	// return only one path
	pq := make(priorityQueue, 0)
	heap.Init(&pq)

	// Initialize and push the source node.
	srcNode := &Item{
		uid:  sg.Params.From,
		cost: 0,
		hop:  0,
	}
	heap.Push(&pq, srcNode)

	numHops := -1
	maxHops := int(sg.Params.ExploreDepth)
	if maxHops == 0 {
		maxHops = int(math.MaxInt32)
	}
	next := make(chan bool, 2)
	expandErr := make(chan error, 2)
	adjacencyMap := make(map[uint64]map[uint64]mapItem)
	go sg.expandOut(ctx, adjacencyMap, next, expandErr)

	// map to store the min cost and parent of nodes.
	dist := make(map[uint64]nodeInfo)
	dist[srcNode.uid] = nodeInfo{
		parent: 0,
		node:   srcNode,
		mapItem: mapItem{
			cost: 0,
		},
	}

	var stopExpansion bool
	var totalWeight float64
	for pq.Len() > 0 {
		item := heap.Pop(&pq).(*Item)
		if item.uid == sg.Params.To {
			totalWeight = item.cost
			break
		}
		if item.hop > numHops && numHops < maxHops {
			// Explore the next level by calling processGraph and add them
			// to the queue.
			if !stopExpansion {
				next <- true
			}
			select {
			case err = <-expandErr:
				if err != nil {
					if err == ErrStop {
						stopExpansion = true
					} else {
						return nil, err
					}
				}
			case <-ctx.Done():
				return nil, ctx.Err()
			}
			numHops++
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			if !stopExpansion {
				neighbours := adjacencyMap[item.uid]
				for toUid, info := range neighbours {
					cost := info.cost
					d, ok := dist[toUid]
					if ok && d.cost <= item.cost+cost {
						continue
					}
					if !ok {
						// This is the first time we're seeing this node. So
						// create a new node and add it to the heap and map.
						node := &Item{
							uid:  toUid,
							cost: item.cost + cost,
							hop:  item.hop + 1,
						}
						heap.Push(&pq, node)
						dist[toUid] = nodeInfo{
							parent: item.uid,
							node:   node,
							mapItem: mapItem{
								cost:  item.cost + cost,
								attr:  info.attr,
								facet: info.facet,
							},
						}
					} else {
						// We've already seen this node. So, just update the cost
						// and fix the priority in the heap and map.
						node := dist[toUid].node
						node.cost = item.cost + cost
						node.hop = item.hop + 1
						heap.Fix(&pq, node.index)
						// Update the map with new values.
						dist[toUid] = nodeInfo{
							parent: item.uid,
							node:   node,
							mapItem: mapItem{
								cost:  item.cost + cost,
								attr:  info.attr,
								facet: info.facet,
							},
						}
					}
				}
			}
		}
	}
	prettyPrint(adjacencyMap)

	next <- false
	// Go through the distance map to find the path.
	var result []uint64
	cur := sg.Params.To
	for i := 0; cur != sg.Params.From && i < len(dist); i++ {
		result = append(result, cur)
		cur = dist[cur].parent
	}
	// Put the path in DestUIDs of the root.
	if cur != sg.Params.From {
		sg.DestUIDs = &pb.List{}
		return nil, nil
	}

	result = append(result, cur)
	l := len(result)
	// Reverse the list.
	for i := 0; i < l/2; i++ {
		result[i], result[l-i-1] = result[l-i-1], result[i]
	}
	sg.DestUIDs.Uids = result
	pickedPath := FindPath(sg.Params.From, sg.Params.To, adjacencyMap)
	if pickedPath != nil {
		sg.DestUIDs.Uids = pickedPath
	}
	shortestSg := createPathSubgraph(ctx, dist, totalWeight, result)
	shortestSg.Params.Alias = "hello world"
	return []*SubGraph{shortestSg}, nil
}

func prettyPrint(adjacencyMap map[uint64]map[uint64]mapItem) {
	fmt.Println("pretty print adjacency map...")
	for row, v := range adjacencyMap {
		fmt.Print(row, "  ")
		for col, val := range v {
			fmt.Print("(", row, " -> ", col, "  ", val.attr, ") ")
		}
		fmt.Println()
	}
}

func FindPath(from uint64, to uint64, adjacencyMap map[uint64]map[uint64]mapItem) []uint64 {
	p := []uint64{from}
	pa := findPath(from, to, adjacencyMap, p)
	if pa[len(pa)-1] == to {
		return pa
	}
	return nil
}

/**
查找路径.
**/
func findPath(from uint64, to uint64, adjacencyMap map[uint64]map[uint64]mapItem, path []uint64) []uint64 {
	endNodes := expandPoint(from, adjacencyMap)
	for _, v := range endNodes {
		if v == to {
			path = append(path, v)
			return path
		}
	}
	//
	startPoints := adjacencyMap[from]
	for start, _ := range startPoints {
		p := append(path, start)
		pa := findPath(start, to, adjacencyMap, p)
		if pa[len(pa)-1] == to {
			return pa
		}
	}
	return path
}

/**
展开节点.
*/
func expandPoint(point uint64, adjacencyMap map[uint64]map[uint64]mapItem) []uint64 {
	ret := []uint64{}
	nodes := adjacencyMap[point]
	for k, _ := range nodes {
		ret = append(ret, k)
	}
	return ret
}
