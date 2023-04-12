package celinlineconst

import (
	"fmt"

	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"google.golang.org/protobuf/proto"
)

func InlineConst(expr *exprpb.Expr, renames map[string]*exprpb.Expr_ConstExpr) (*exprpb.Expr, error) {
	return ReplaceVariables(expr, renames, nil)
}

func ReplaceVariables(expr *exprpb.Expr, constRenames map[string]*exprpb.Expr_ConstExpr, aliases map[string]string) (*exprpb.Expr, error) {
	if len(constRenames) == 0 && len(aliases) == 0 {
		// No modifications.
		return expr, nil
	}

	if expr == nil {
		return nil, nil
	}

	switch e := expr.ExprKind.(type) {

	case *exprpb.Expr_ConstExpr:
		return expr, nil

	case *exprpb.Expr_IdentExpr:
		if constExpr, has := constRenames[e.IdentExpr.Name]; has {
			return &exprpb.Expr{
				Id:       expr.Id,
				ExprKind: constExpr,
			}, nil
		}
		if alias, has := aliases[e.IdentExpr.Name]; has {
			return &exprpb.Expr{
				Id: expr.Id,
				ExprKind: &exprpb.Expr_IdentExpr{
					IdentExpr: &exprpb.Expr_Ident{
						Name: alias,
					},
				},
			}, nil
		}
		return expr, nil

	case *exprpb.Expr_SelectExpr:
		operand, err := ReplaceVariables(e.SelectExpr.Operand, constRenames, aliases)
		if err != nil {
			return nil, err
		}
		selectExpr := proto.Clone(e.SelectExpr).(*exprpb.Expr_Select)
		selectExpr.Operand = operand
		return &exprpb.Expr{
			Id: expr.Id,
			ExprKind: &exprpb.Expr_SelectExpr{
				SelectExpr: selectExpr,
			},
		}, nil

	case *exprpb.Expr_CallExpr:
		exprCall := e.CallExpr
		target, err := ReplaceVariables(exprCall.Target, constRenames, aliases)
		if err != nil {
			return nil, err
		}
		args := make([]*exprpb.Expr, 0, len(exprCall.Args))
		for _, arg := range exprCall.Args {
			argCopy, err := ReplaceVariables(arg, constRenames, aliases)
			if err != nil {
				return nil, err
			}
			args = append(args, argCopy)
		}
		exprCallCopy := proto.Clone(exprCall).(*exprpb.Expr_Call)
		exprCallCopy.Target = target
		exprCallCopy.Args = args
		return &exprpb.Expr{
			Id: expr.Id,
			ExprKind: &exprpb.Expr_CallExpr{
				CallExpr: exprCallCopy,
			},
		}, nil

	case *exprpb.Expr_ListExpr:
		elements := make([]*exprpb.Expr, 0, len(e.ListExpr.Elements))
		for _, elem := range e.ListExpr.Elements {
			elemCopy, err := ReplaceVariables(elem, constRenames, aliases)
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
			entryCopy := proto.Clone(entry).(*exprpb.Expr_CreateStruct_Entry)
			value, err := ReplaceVariables(entry.Value, constRenames, aliases)
			if err != nil {
				return nil, err
			}
			entryCopy.Value = value
			if mapKey, ok := entry.KeyKind.(*exprpb.Expr_CreateStruct_Entry_MapKey); ok {
				mapKeyCopy, err := ReplaceVariables(mapKey.MapKey, constRenames, aliases)
				if err != nil {
					return nil, err
				}
				entryCopy.KeyKind = &exprpb.Expr_CreateStruct_Entry_MapKey{
					MapKey: mapKeyCopy,
				}
			}
			entries = append(entries, entryCopy)
		}
		createStruct := proto.Clone(e.StructExpr).(*exprpb.Expr_CreateStruct)
		createStruct.Entries = entries
		return &exprpb.Expr{
			Id: expr.Id,
			ExprKind: &exprpb.Expr_StructExpr{
				StructExpr: createStruct,
			},
		}, nil

	case *exprpb.Expr_ComprehensionExpr:
		c := proto.Clone(e.ComprehensionExpr).(*exprpb.Expr_Comprehension)
		var err error
		c.IterRange, err = ReplaceVariables(c.IterRange, constRenames, aliases)
		if err != nil {
			return nil, err
		}
		c.AccuInit, err = ReplaceVariables(c.AccuInit, constRenames, aliases)
		if err != nil {
			return nil, err
		}
		c.LoopCondition, err = ReplaceVariables(c.LoopCondition, constRenames, aliases)
		if err != nil {
			return nil, err
		}
		c.LoopStep, err = ReplaceVariables(c.LoopStep, constRenames, aliases)
		if err != nil {
			return nil, err
		}
		c.Result, err = ReplaceVariables(c.Result, constRenames, aliases)
		if err != nil {
			return nil, err
		}
		return &exprpb.Expr{
			Id: expr.Id,
			ExprKind: &exprpb.Expr_ComprehensionExpr{
				ComprehensionExpr: c,
			},
		}, nil

	default:
		return nil, fmt.Errorf("unknown expr type: %T", expr.ExprKind)
	}
}
