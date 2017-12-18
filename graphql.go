package graphql

import (
	"context"

	"github.com/ChargePoint/graphql/gqlerrors"
	"github.com/ChargePoint/graphql/language/parser"
	"github.com/ChargePoint/graphql/language/source"
)

type Params struct {
	// The GraphQL type system to use when validating and executing a query.
	Schema Schema

	// A GraphQL language formatted string representing the requested operation.
	RequestString string

	// The value provided as the first argument to resolver functions on the top
	// level type (e.g. the query object type).
	RootObject map[string]interface{}

	// A mapping of variable name to runtime value to use for all variables
	// defined in the requestString.
	VariableValues map[string]interface{}

	// The name of the operation to use if requestString contains multiple
	// possible operations. Can be omitted if requestString contains only
	// one operation.
	OperationName string

	// Context may be provided to pass application-specific per-request
	// information to resolve functions.
	Context context.Context
}

func do(p Params, synchronous bool) *Result {
	source := source.NewSource(&source.Source{
		Body: []byte(p.RequestString),
		Name: "GraphQL request",
	})
	AST, err := parser.Parse(parser.ParseParams{Source: source})
	if err != nil {
		return &Result{
			Errors: gqlerrors.FormatErrors(err),
		}
	}
	validationResult := ValidateDocument(&p.Schema, AST, nil)

	if !validationResult.IsValid {
		return &Result{
			Errors: validationResult.Errors,
		}
	}

	return execute(ExecuteParams{
		Schema:        p.Schema,
		Root:          p.RootObject,
		AST:           AST,
		OperationName: p.OperationName,
		Args:          p.VariableValues,
		Context:       p.Context,
	}, synchronous)
}

func Do(p Params) *Result {
	return do(p, false)
}

// DoSynchronously will execute the target handler within the current goroutine.
// This is intended to be used by unit tests to run the handler under test on the
// test goroutine so any calls to testing.Fatalf through interactions between
// the handler and injected mocks will be properly handled and reported by Go's
// mock controller.
func DoSynchronously(p Params) *Result {
	return do(p, true)
}
