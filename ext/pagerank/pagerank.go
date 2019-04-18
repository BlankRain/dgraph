package pagerank

import (
	"github.com/dgraph-io/dgraph/ext"
	"github.com/dgraph-io/dgraph/types"
)

func init() {
	ext.RegistProcessFunction("pagerank", PageRank, 160)
}

func PageRank(param ext.ProcessFuncParam) (map[uint64]types.Val, error) {
	ret := make(map[uint64]types.Val)
	for _, uid := range param.SrcUids.Uids {
		ret[uid] = types.Val{
			Value: int64(10086),
			Tid:   types.IntID,
		}
	}
	return ret, nil
}
