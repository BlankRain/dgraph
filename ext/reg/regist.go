package reg

import (
	_ "github.com/dgraph-io/dgraph/ext/communitydetection"
	_ "github.com/dgraph-io/dgraph/ext/indegree"
	_ "github.com/dgraph-io/dgraph/ext/outdegree"
	_ "github.com/dgraph-io/dgraph/ext/pagerank"
)

func Init() {
}
