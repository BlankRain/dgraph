package ext

import (
	"context"

	"github.com/dgraph-io/dgraph/protos/pb"
	"github.com/dgraph-io/dgraph/types"
)

type ParamTree struct {
	Var   string
	Child []ParamTree
}
type ProcessFuncParam struct {
	Context     context.Context
	ReadTs      uint64
	FuncName    string
	SrcUids     *pb.List
	ParamLabels []string
	ParamTree   []ParamTree
}

type ProcessFunc func(param ProcessFuncParam) (map[uint64]types.Val, error)

// a map to store functions for process graph
var processMathFunctionPool = make(map[string]ProcessFunc)

// a map to extend the `mathOpPrecedence` in gql package
var extMathOpPrecedence = make(map[string]int)

func RegistProcessFunction(name string, fn ProcessFunc, precedence int) {
	processMathFunctionPool[name] = fn
	RegistFunctionPrecedence(name, precedence)
}

//
func HasProcessFunc(funcName string) bool {
	_, flag := processMathFunctionPool[funcName]
	return flag
}

// regist a op  with precedence
func RegistFunctionPrecedence(op string, precedence int) {
	extMathOpPrecedence[op] = precedence
}

// return a processfunc for the name
func PickProcessFunc(name string) ProcessFunc {
	return processMathFunctionPool[name]
}

// read a precedence value of op
func PickExtMathOpPrecedence(op string) int {
	return extMathOpPrecedence[op]
}
