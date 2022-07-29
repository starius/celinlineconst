package celinlineconst

import (
	"fmt"

	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

func InlineConst(expr *exprpb.Expr, renames map[string]*exprpb.Expr_ConstExpr) (*exprpb.Expr, error) {
	if expr == nil {
		return nil, nil
	}

	switch e := expr.ExprKind.(type) {

	case *exprpb.Expr_ConstExpr:
		return expr, nil

	case *exprpb.Expr_IdentExpr:
		if constExpr, has := renames[e.IdentExpr.Name]; has {
			return &exprpb.Expr{
				Id:       expr.Id,
				ExprKind: constExpr,
			}, nil
		}
		return expr, nil

	case *exprpb.Expr_SelectExpr:
		operand, err := InlineConst(e.SelectExpr.Operand, renames)
		if err != nil {
			return nil, err
		}
		selectExpr := *e.SelectExpr
		selectExpr.Operand = operand
		return &exprpb.Expr{
			Id: expr.Id,
			ExprKind: &exprpb.Expr_SelectExpr{
				SelectExpr: &selectExpr,
			},
		}, nil

	case *exprpb.Expr_CallExpr:
		exprCall := e.CallExpr
		target, err := InlineConst(exprCall.Target, renames)
		if err != nil {
			return nil, err
		}
		args := make([]*exprpb.Expr, 0, len(exprCall.Args))
		for _, arg := range exprCall.Args {
			argCopy, err := InlineConst(arg, renames)
			if err != nil {
				return nil, err
			}
			args = append(args, argCopy)
		}
		exprCallCopy := *exprCall
		exprCallCopy.Target = target
		exprCallCopy.Args = args
		return &exprpb.Expr{
			Id: expr.Id,
			ExprKind: &exprpb.Expr_CallExpr{
				CallExpr: &exprCallCopy,
			},
		}, nil

	case *exprpb.Expr_ListExpr:
		elements := make([]*exprpb.Expr, 0, len(e.ListExpr.Elements))
		for _, elem := range e.ListExpr.Elements {
			elemCopy, err := InlineConst(elem, renames)
			if err != nil {
				return nil, err
			}
			elements = append(elements, elemCopy)
		}
		return &exprpb.Expr{
			Id: expr.Id,
			ExprKind: &exprpb.Expr_ListExpr{
				ListExpr: &exprpb.Expr_CreateList{
					Elements: elements,
				},
			},
		}, nil

	case *exprpb.Expr_StructExpr:
		entries := make([]*exprpb.Expr_CreateStruct_Entry, 0, len(e.StructExpr.Entries))
		for _, entry := range e.StructExpr.Entries {
			entryCopy := *entry
			value, err := InlineConst(entry.Value, renames)
			if err != nil {
				return nil, err
			}
			entryCopy.Value = value
			if mapKey, ok := entry.KeyKind.(*exprpb.Expr_CreateStruct_Entry_MapKey); ok {
				mapKeyCopy, err := InlineConst(mapKey.MapKey, renames)
				if err != nil {
					return nil, err
				}
				entryCopy.KeyKind = &exprpb.Expr_CreateStruct_Entry_MapKey{
					MapKey: mapKeyCopy,
				}
			}
			entries = append(entries, &entryCopy)
		}
		createStruct := *e.StructExpr
		createStruct.Entries = entries
		return &exprpb.Expr{
			Id: expr.Id,
			ExprKind: &exprpb.Expr_StructExpr{
				StructExpr: &createStruct,
			},
		}, nil

	case *exprpb.Expr_ComprehensionExpr:
		c := *e.ComprehensionExpr
		var err error
		c.IterRange, err = InlineConst(c.IterRange, renames)
		if err != nil {
			return nil, err
		}
		c.AccuInit, err = InlineConst(c.AccuInit, renames)
		if err != nil {
			return nil, err
		}
		c.LoopCondition, err = InlineConst(c.LoopCondition, renames)
		if err != nil {
			return nil, err
		}
		c.LoopStep, err = InlineConst(c.LoopStep, renames)
		if err != nil {
			return nil, err
		}
		c.Result, err = InlineConst(c.Result, renames)
		if err != nil {
			return nil, err
		}
		return &exprpb.Expr{
			Id: expr.Id,
			ExprKind: &exprpb.Expr_ComprehensionExpr{
				ComprehensionExpr: &c,
			},
		}, nil

	default:
		return nil, fmt.Errorf("unknown expr type: %T", expr.ExprKind)
	}
}
