package dsl

import (
	"errors"
	"fmt"
	"log"
)

// ParseExpression parses a single expression string.
func ParseExpression(str string) (Expression, error) {
	var (
		expr Expression
		not  bool
	)
	_, c := lex(str)
	for item := range c {
		switch item.itemType {
		case itemNot:
			not = !not

		case itemAll:
			expr.selector = selectAll

		case itemConnected:
			expr.selector = selectConnected

		case itemNonlocal:
			expr.selector = selectNonlocal

		case itemLike:
			item = <-c
			switch item.itemType {
			case itemRegex:
				expr.selector = selectLike(item.literal)
			default:
				return Expression{}, fmt.Errorf("bad LIKE: want %s, got %s", itemRegex, item.itemType)
			}

		case itemWith:
			item = <-c
			switch item.itemType {
			case itemKeyValue:
				expr.selector = selectWith(item.literal)
			default:
				return Expression{}, fmt.Errorf("bad WITH: want %s, got %s", itemKeyValue, item.itemType)
			}

		case itemHighlight:
			expr.transformer = transformHighlight

		case itemRemove:
			expr.transformer = transformRemove

		case itemShowOnly:
			expr.transformer = transformShowOnly

		case itemMerge:
			expr.transformer = transformMerge

		case itemGroupBy:
			item = <-c
			switch item.itemType {
			case itemKeyList:
				expr.transformer = transformGroupBy(item.literal)
			default:
				return Expression{}, fmt.Errorf("bad GROUPBY: want %s, got %s", itemKeyList, item.itemType)
			}

		case itemJoin:
			item = <-c
			switch item.itemType {
			case itemKey:
				expr.transformer = transformJoin(item.literal)
			default:
				return Expression{}, fmt.Errorf("bad JOIN: want %s, got %s", itemKey, item.itemType)
			}

		default:
			return Expression{}, errors.New(item.literal)
		}
	}
	if not {
		expr.selector = selectNot(expr.selector)
	}
	if expr.transformer == nil {
		expr.transformer = transformHighlight
	}
	return expr, nil
}

// ParseExpressions parses multiple expression strings.
func ParseExpressions(strs ...string) Expressions {
	var exprs Expressions
	for _, str := range strs {
		expr, err := ParseExpression(str)
		if err != nil {
			log.Printf("%s: %v", str, err)
			continue
		}
		log.Printf("%s: OK", str)
		exprs = append(exprs, expr)
	}
	return exprs
}
