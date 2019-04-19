package query

import (
	"context"
	"log"

	"github.com/dgraph-io/dgraph/ext"
	"github.com/dgraph-io/dgraph/protos/pb"
	"github.com/dgraph-io/dgraph/types"
	"github.com/dgraph-io/dgraph/worker"
	"github.com/dgraph-io/dgraph/x"
)

func NodeDegree(param ext.ProcessFuncParam) (map[uint64]types.Val, error) {
	destMap := make(map[uint64]int64)
	// init
	for _, uid := range param.SrcUids.Uids {
		destMap[uid] = 0
	}
	for _, edgePred := range param.ParamLabels {
		err := doComputeNodeOutDegree(param.Context, param.SrcUids, edgePred, false, param.ReadTs, destMap)
		if err != nil {
			log.Printf("error when compute node outdegree %v", err)
			return nil, err
		}
	}
	// all schemas
	schs, err := worker.GetSchemaOverNetwork(param.Context, &pb.SchemaRequest{})
	if err != nil {
		log.Printf("error when compute node outdegree %v", err)
		return nil, err
	}
	// find reverse
	for _, edgePred := range param.ParamLabels {
		for _, sch := range schs {
			if edgePred == sch.Predicate && sch.Reverse {
				err := doComputeNodeOutDegree(param.Context, param.SrcUids, edgePred, true, param.ReadTs, destMap)
				if err != nil {
					log.Printf("error when compute node outdegree %v", err)
					return nil, err
				}
			}
		}
	}
	// trans to types.Val
	res := make(map[uint64]types.Val)
	for k, v := range destMap {
		res[k] = types.Val{
			Value: v,
			Tid:   types.IntID,
		}
	}
	return res, nil
}

func newList(data []uint64) *pb.List {
	return &pb.List{Uids: data}
}
func doComputeNodeOutDegree(ctx context.Context, uidList *pb.List,
	pred string, isReverse bool, readTS uint64, destMap map[uint64]int64) error {
	//
	taskQuery := &pb.Query{
		Attr:    pred,
		UidList: uidList,
		DoCount: true,
		Reverse: isReverse,
		ReadTs:  readTS,
	}
	result, err := worker.ProcessTaskOverNetwork(ctx, taskQuery)
	if err != nil {
		return x.Errorf("Error in NodeDegree { pred:%s ,msg: %v}", pred, err)
	}
	for i := 0; i < len(result.Counts); i++ {
		if isReverse {
			destMap[uidList.Uids[i]] -= int64(result.Counts[i])
		} else {
			destMap[uidList.Uids[i]] += int64(result.Counts[i])
		}
	}
	return nil
}

func init() {
	ext.RegistProcessFunction("nodedegree", NodeDegree, 106)
}
