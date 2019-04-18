package ext

import (
	"context"

	"github.com/dgraph-io/dgraph/protos/pb"
	"github.com/dgraph-io/dgraph/types"
)

type ParamTree struct {
	Var string
}
type ProcessFuncParam struct {
	Context     context.Context
	ReadTS      uint64
	FuncName    string
	SrcUids     *pb.List
	ParamLabels []string
	ParamTree   []ParamTree
}

type ProcessFunc func(param ProcessFuncParam) (map[uint64]types.Val, error)

// a map to store functions for process graph
var processMathFunctionPool = make(map[string]ProcessFunc)

var mathOpPrecedence = make(map[string]int)

func RegistProcessFunction(name string, fn ProcessFunc, precedence int) {
	processMathFunctionPool[name] = fn
	RegistFunctionPrecedence(name, precedence)
}
func HasProcessFunc(funcName string) bool {
	_, flag := processMathFunctionPool[funcName]
	return flag
}
func RegistFunctionPrecedence(name string, precedence int) {
	mathOpPrecedence[name] = precedence
}

func PickProcessFunc(name string) ProcessFunc {
	return processMathFunctionPool[name]
}

func PickExtMathOpPrecedence(op string) int {
	return mathOpPrecedence[op]
}
