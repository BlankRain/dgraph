package pagerank

import (
	"github.com/dcadenas/pagerank"
	"github.com/dgraph-io/dgraph/ext"
	"github.com/dgraph-io/dgraph/protos/pb"
	"github.com/dgraph-io/dgraph/types"
	"github.com/dgraph-io/dgraph/worker"
	"github.com/dgraph-io/dgraph/x"
)

func init() {
	ext.RegistProcessFunction("pagerank", PageRank, 160)
}

func PageRank(param ext.ProcessFuncParam) (map[uint64]types.Val, error) {
	ret := make(map[uint64]types.Val)
	graph := pagerank.New()
	for _, edgePred := range param.ParamLabels {
		err := buildGraph(graph, edgePred, param)
		if err != nil {
			return nil, err
		}
	}
	probability_of_following_a_link := 0.85 // The bigger the number, less probability we have to teleport to some random link
	tolerance := 0.0001                     // the smaller the number, the more exact the result will be but more CPU cycles will be needed
	// init with zero
	for _, uid := range param.SrcUids.Uids {
		ret[uid] = types.Val{
			Value: 0.0,
			Tid:   types.FloatID,
		}
	}
	graph.Rank(probability_of_following_a_link, tolerance, func(identifier int, rank float64) {
		ret[uint64(identifier)] = types.Val{
			Value: rank,
			Tid:   types.FloatID,
		}
	})
	return ret, nil
}
func buildGraph(graph pagerank.Interface, pred string, param ext.ProcessFuncParam) error {
	taskQuery := &pb.Query{
		Attr:    pred,
		UidList: param.SrcUids,
		ReadTs:  param.ReadTs,
	}
	result, err := worker.ProcessTaskOverNetwork(param.Context, taskQuery)
	if err != nil {
		return x.Errorf("Error in PageRank { pred:%s ,msg: %v}", pred, err)
	}
	u := result.UidMatrix
	for i := 0; i < len(u); i++ {
		for _, t := range u[i].Uids {
			f := param.SrcUids.Uids[i]
			graph.Link(int(f), int(t))
		}
	}
	return nil
}
