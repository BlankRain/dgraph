package communitydetection

import (
	"fmt"
	"strconv"

	"github.com/dgraph-io/dgraph/ext"
	"github.com/dgraph-io/dgraph/ext/communitydetection/louvain"
	"github.com/dgraph-io/dgraph/protos/pb"
	"github.com/dgraph-io/dgraph/types"
	"github.com/dgraph-io/dgraph/worker"
	"github.com/dgraph-io/dgraph/x"
)

func init() {
	ext.RegistProcessFunction("louvain", Louvain, 160)
}

func Louvain(param ext.ProcessFuncParam) (map[uint64]types.Val, error) {
	ret := make(map[uint64]types.Val)
	links, err := buildWeightGraph(param)
	if err != nil {
		return nil, err
	}

	reader := louvain.NewGraphReader()
	for _, node := range param.SrcUids.Uids {
		reader.AddNode(fmt.Sprintf("%v", node))
	}
	graph := louvain.Graph{
		Incidences: make(louvain.Edges, len(links)),
		Selfs:      make([]louvain.WeightType, reader.GetNodeSize()),
	}
	for _, link := range links {
		f := reader.GetNodeIndex(fmt.Sprintf("%v", link.from))
		t := reader.GetNodeIndex(fmt.Sprintf("%v", link.to))
		weight := louvain.WeightType(link.weight)
		graph.AddUndirectedEdge(f, t, weight)
	}
	lou := louvain.NewLouvain(graph)
	lou.Compute()
	for nodeId, commId := range lou.GetBestPertition() {
		nId := reader.GetNodeLabel(nodeId)
		id, err := strconv.ParseInt(nId, 10, 64)
		if err != nil {
			return nil, err
		}
		fmt.Printf("nodeId: %s communityId: %d \n", nId, commId)
		ret[uint64(id)] = types.Val{
			Value: int64(commId),
			Tid:   types.IntID,
		}
	}
	return ret, nil
}

type edge struct {
	from, to, weight uint64
}

func buildWeightGraph(param ext.ProcessFuncParam) (map[string]edge, error) {
	edges := make(map[string]edge)
	for _, pred := range param.ParamLabels {
		taskQuery := &pb.Query{
			Attr:    pred,
			UidList: param.SrcUids,
			ReadTs:  param.ReadTs,
		}
		result, err := worker.ProcessTaskOverNetwork(param.Context, taskQuery)
		if err != nil {
			return nil, x.Errorf("Error in Louvain { pred:%s ,msg: %v}", pred, err)
		}
		u := result.UidMatrix
		// link and update weight
		for i := 0; i < len(u); i++ {
			for _, t := range u[i].Uids {
				f := param.SrcUids.Uids[i]
				k := fmt.Sprintf("%v:%v", f, t)
				if e, ok := edges[k]; ok {
					e.weight++
				} else {
					edges[k] = edge{
						f, t, 1,
					}
				}
			}
		}
	}

	return edges, nil
}
