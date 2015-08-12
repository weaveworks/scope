package main

import (
	"fmt"
	"log"
)

func parseExpressions(strs []string) []expression {
	var exprs []expression
	for _, str := range strs {
		expr, err := parseExpression(str)
		if err != nil {
			log.Printf("%s: %v", str, err)
			continue
		}
		log.Printf("%s: OK", str)
		exprs = append(exprs, expr)
	}
	return exprs
}

func parseExpression(str string) (expression, error) {
	var (
		expr expression
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
				return expression{}, fmt.Errorf("bad LIKE: want %s, got %s", itemRegex, item.itemType)
			}

		case itemWith:
			item = <-c
			switch item.itemType {
			case itemKeyValue:
				expr.selector = selectWith(item.literal)
			default:
				return expression{}, fmt.Errorf("bad WITH: want %s, got %s", itemKeyValue, item.itemType)
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
				return expression{}, fmt.Errorf("bad GROUPBY: want %s, got %s", itemKeyList, item.itemType)
			}

		case itemJoin:
			item = <-c
			switch item.itemType {
			case itemKey:
				expr.transformer = transformJoin(item.literal)
			default:
				return expression{}, fmt.Errorf("bad JOIN: want %s, got %s", itemKey, item.itemType)
			}

		default:
			return expression{}, fmt.Errorf("%s: %s", str, item.literal)
		}
	}
	if not {
		expr.selector = selectNot(expr.selector)
	}
	return expr, nil
}
