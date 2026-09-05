package objectsql

import (
	"fmt"
	"sort"
	"strings"

	reportmodel "github.com/domainry/domainry-report-sdk/model"
	reportengine "github.com/domainry/domainry-report-sdk/query"
)

func (c *objectSQLCompiler) inferJoinCardinality(expression reportmodel.ReportObjectSQLExpression, rightAlias string) (string, []string) {
	candidates := c.joinRelationCandidates(rightAlias)
	best := ""
	walkObjectSQLExpression(expression, func(current reportmodel.ReportObjectSQLExpression) {
		if current.Kind != "comparison" || current.Operator != "=" || len(current.Arguments) != 2 {
			return
		}
		left, right := current.Arguments[0], current.Arguments[1]
		if left.Kind != "field" || right.Kind != "field" {
			return
		}
		if left.Alias == rightAlias {
			left, right = right, left
		}
		if right.Alias != rightAlias || c.aliasOrder[left.Alias] >= c.aliasOrder[rightAlias] {
			return
		}
		leftObject, rightObject := c.aliases[left.Alias], c.aliases[right.Alias]
		leftField, leftFound := objectSQLField(leftObject, left.FieldKey)
		rightField, rightFound := objectSQLField(rightObject, right.FieldKey)
		if !leftFound || !rightFound || !objectSQLJoinEqualityBacked(leftObject, leftField, rightObject, rightField) {
			return
		}
		inferred := "one_to_many"
		switch {
		case objectSQLFieldUnique(leftField) && objectSQLFieldUnique(rightField):
			inferred = "one_to_one"
		case objectSQLFieldUnique(rightField):
			inferred = "many_to_one"
		}
		if objectSQLCardinalityRank(inferred) > objectSQLCardinalityRank(best) {
			best = inferred
		}
	})
	return best, candidates
}

func objectSQLJoinEqualityBacked(leftObject reportengine.Object, left reportengine.Field, rightObject reportengine.Object, right reportengine.Field) bool {
	if objectSQLFieldUnique(left) || objectSQLFieldUnique(right) {
		return true
	}
	return objectSQLRelationTargets(left, rightObject.Key, right.Key) || objectSQLRelationTargets(right, leftObject.Key, left.Key)
}

func objectSQLRelationTargets(field reportengine.Field, objectKey, targetField string) bool {
	return strings.TrimSpace(field.Type) == "relation" && strings.TrimSpace(field.RelationTarget) == strings.TrimSpace(objectKey) && targetField == "id"
}

func objectSQLFieldUnique(field reportengine.Field) bool {
	return field.Unique || strings.TrimSpace(field.RelationCardinality) == "one_to_one"
}

func objectSQLCardinalityRank(value string) int {
	switch value {
	case "one_to_one":
		return 3
	case "many_to_one":
		return 2
	case "one_to_many":
		return 1
	default:
		return 0
	}
}

func (c *objectSQLCompiler) joinRelationCandidates(rightAlias string) []string {
	result := []string{}
	rightObject := c.aliases[rightAlias]
	for alias, object := range c.aliases {
		if alias == rightAlias || c.aliasOrder[alias] >= c.aliasOrder[rightAlias] {
			continue
		}
		for _, field := range object.Fields {
			if objectSQLRelationTargets(field, rightObject.Key, "id") {
				result = append(result, alias+"."+field.Key+" = "+rightAlias+".id")
			}
		}
		for _, field := range rightObject.Fields {
			if objectSQLRelationTargets(field, object.Key, "id") {
				result = append(result, rightAlias+"."+field.Key+" = "+alias+".id")
			}
		}
	}
	sort.Strings(result)
	return result
}

func (c *objectSQLCompiler) validateAmplification(plan reportmodel.ReportObjectSQLPlan) error {
	for projectionIndex, projection := range plan.Projections {
		var validationErr error
		walkObjectSQLExpression(projection.Expression, func(expression reportmodel.ReportObjectSQLExpression) {
			if validationErr != nil || expression.Kind != "aggregate" || expression.Distinct || expression.Name == "min" || expression.Name == "max" {
				return
			}
			sources := map[string]bool{}
			for _, argument := range expression.Arguments {
				collectObjectSQLFieldAliases(argument, sources)
			}
			for joinIndex, source := range plan.Sources {
				if source.Cardinality != "one_to_many" {
					continue
				}
				if len(sources) == 0 && expression.Name == "count" {
					validationErr = objectSQLPlanError("backend.report.join_measure_amplification", fmt.Sprintf("object_sql_v1.result_schema[%d]", projectionIndex), map[string]string{"measure": projection.Alias, "join_alias": source.Alias})
					return
				}
				for alias := range sources {
					if c.aliasOrder[alias] < joinIndex {
						validationErr = objectSQLPlanError("backend.report.join_measure_amplification", fmt.Sprintf("object_sql_v1.result_schema[%d]", projectionIndex), map[string]string{"measure": projection.Alias, "source_alias": alias, "join_alias": source.Alias})
						return
					}
				}
			}
		})
		if validationErr != nil {
			return validationErr
		}
	}
	return nil
}

func (c *objectSQLCompiler) populateSourceFields(plan *reportmodel.ReportObjectSQLPlan) {
	fields := map[string]map[string]bool{}
	for _, source := range plan.Sources {
		fields[source.Alias] = map[string]bool{}
		if source.On != nil {
			collectObjectSQLFields(*source.On, fields)
		}
	}
	for _, projection := range plan.Projections {
		collectObjectSQLFields(projection.Expression, fields)
	}
	if plan.Where != nil {
		collectObjectSQLFields(*plan.Where, fields)
	}
	if plan.Having != nil {
		collectObjectSQLFields(*plan.Having, fields)
	}
	for _, expression := range plan.GroupBy {
		collectObjectSQLFields(expression, fields)
	}
	for _, order := range plan.OrderBy {
		collectObjectSQLFields(order.Expression, fields)
	}
	for index := range plan.Sources {
		for field := range fields[plan.Sources[index].Alias] {
			plan.Sources[index].Fields = append(plan.Sources[index].Fields, field)
		}
		sort.Strings(plan.Sources[index].Fields)
	}
}

func collectObjectSQLFields(expression reportmodel.ReportObjectSQLExpression, fields map[string]map[string]bool) {
	if expression.Kind == "field" {
		fields[expression.Alias][expression.FieldKey] = true
	}
	for _, argument := range expression.Arguments {
		collectObjectSQLFields(argument, fields)
	}
	for _, when := range expression.Whens {
		collectObjectSQLFields(when.Condition, fields)
		collectObjectSQLFields(when.Value, fields)
	}
	if expression.Else != nil {
		collectObjectSQLFields(*expression.Else, fields)
	}
}

func collectObjectSQLFieldAliases(expression reportmodel.ReportObjectSQLExpression, aliases map[string]bool) {
	if expression.Kind == "field" {
		aliases[expression.Alias] = true
	}
	for _, argument := range expression.Arguments {
		collectObjectSQLFieldAliases(argument, aliases)
	}
	for _, when := range expression.Whens {
		collectObjectSQLFieldAliases(when.Condition, aliases)
		collectObjectSQLFieldAliases(when.Value, aliases)
	}
	if expression.Else != nil {
		collectObjectSQLFieldAliases(*expression.Else, aliases)
	}
}

func walkObjectSQLExpression(expression reportmodel.ReportObjectSQLExpression, visit func(reportmodel.ReportObjectSQLExpression)) {
	visit(expression)
	for _, argument := range expression.Arguments {
		walkObjectSQLExpression(argument, visit)
	}
	for _, when := range expression.Whens {
		walkObjectSQLExpression(when.Condition, visit)
		walkObjectSQLExpression(when.Value, visit)
	}
	if expression.Else != nil {
		walkObjectSQLExpression(*expression.Else, visit)
	}
}

func objectSQLResultCompatible(result reportmodel.ReportResultColumnSchema, expression reportmodel.ReportObjectSQLExpression) bool {
	if result.Type == expression.Type {
		return result.Type != "currency" && result.Type != "decimal" || result.Scale == 0 || result.Scale == expression.Scale
	}
	return result.Type == "decimal" && (expression.Type == "decimal" || expression.Type == "integer")
}
