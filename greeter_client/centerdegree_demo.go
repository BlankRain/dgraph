package main

import (
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/dgraph/gql"
)

func main() {

	q := `
	query q{
		x(func: has(name)){
			name
			cd :math(nodedegree(
				<friend>,<name>,<hello>
			)
		)
		}
	}
	`
	fmt.Println(q)
	ret, err := gql.Parse(gql.Request{Str: q})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	pprint(ret)

	x := []int{5, 3, 2, 1}
	for i, e := range x {
		fmt.Println(i, e)
	}
}

func pprint(r gql.Result) {
	b, err := json.MarshalIndent(r, "", "  ")

	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(b))
}
