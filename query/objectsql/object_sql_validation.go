package objectsql

import (
	"fmt"
	"strconv"
	"strings"

	reportmodel "github.com/domainry/domainry-report-sdk/model"
	reportengine "github.com/domainry/domainry-report-sdk/query"
	"vitess.io/vitess/go/vt/sqlparser"
)

const (
	reportObjectSQLMaximumLength = 32 << 10
	// A payroll statement can legitimately exceed 1,000 parser nodes while
	// staying within the fixed 64-column, eight-join and 32 KiB SQL bounds.
	// Keep one bounded backend calculation instead of forcing callers to split
	// financial formulas across independently observed reports.
	reportObjectSQLMaximumNodes   = 2000
	reportObjectSQLMaximumJoins   = 8
	reportObjectSQLMaximumColumns = 64
	reportObjectSQLMaximumLimit   = 10000
	reportObjectSQLDefaultLimit   = 1000
)

type objectSQLCompiler struct {
	schema          reportmodel.ReportObjectSQLSchema
	objects         map[string]reportengine.Object
	aliases         map[string]reportengine.Object
	aliasOrder      map[string]int
	parameters      map[string]reportmodel.ReportObjectSQLParameter
	cardinality     map[string]string
	cardinalityPath map[string]string
	results         map[string]reportmodel.ReportResultColumnSchema
	resultIndex     map[string]int
	projections     map[string]reportmodel.ReportObjectSQLExpression
	declaredObjects map[string]bool
}

// CompileReportObjectSQL parses the canonical object_sql_v1 subset with a
// maintained third-party parser, binds every table/field to published
// metadata, and lowers only allowlisted AST nodes into an internal plan.
func CompileReportObjectSQL(schema reportmodel.ReportObjectSQLSchema, objects map[string]reportengine.Object) (reportmodel.ReportObjectSQLPlan, error) {
	if len(schema.SQL) == 0 || len(schema.SQL) > reportObjectSQLMaximumLength {
		return reportmodel.ReportObjectSQLPlan{}, objectSQLPlanError("backend.report.object_sql_length_invalid", "object_sql_v1.sql", nil)
	}
	if len(schema.ResultSchema) > reportObjectSQLMaximumColumns {
		return reportmodel.ReportObjectSQLPlan{}, objectSQLPlanError("backend.report.object_sql_result_schema_invalid", "object_sql_v1.result_schema", nil)
	}
	compiler := objectSQLCompiler{
		schema: schema, objects: objects, aliases: map[string]reportengine.Object{}, aliasOrder: map[string]int{},
		parameters: map[string]reportmodel.ReportObjectSQLParameter{}, cardinality: map[string]string{}, cardinalityPath: map[string]string{},
		results: map[string]reportmodel.ReportResultColumnSchema{}, resultIndex: map[string]int{}, projections: map[string]reportmodel.ReportObjectSQLExpression{}, declaredObjects: map[string]bool{},
	}
	if err := compiler.compileMetadata(); err != nil {
		return reportmodel.ReportObjectSQLPlan{}, err
	}
	statement, err := sqlparser.NewTestParser().Parse(schema.SQL)
	if err != nil {
		return reportmodel.ReportObjectSQLPlan{}, objectSQLPlanError("backend.report.object_sql_parse_failed", "object_sql_v1.sql", map[string]string{"reason": err.Error()})
	}
	selectStatement, ok := statement.(*sqlparser.Select)
	if !ok {
		return reportmodel.ReportObjectSQLPlan{}, objectSQLPlanError("backend.report.object_sql_statement_forbidden", "object_sql_v1.sql", map[string]string{"statement": fmt.Sprintf("%T", statement)})
	}
	if selectStatement.With != nil || len(selectStatement.Windows) > 0 || selectStatement.Lock != sqlparser.NoLock || selectStatement.Into != nil || selectStatement.Distinct || selectStatement.GroupBy != nil && selectStatement.GroupBy.WithRollup {
		return reportmodel.ReportObjectSQLPlan{}, objectSQLPlanError("backend.report.object_sql_syntax_forbidden", "object_sql_v1.sql", nil)
	}
	nodeCount := 0
	hasWindow := false
	_ = sqlparser.Walk(func(node sqlparser.SQLNode) (bool, error) {
		nodeCount++
		if _, ok := node.(*sqlparser.OverClause); ok {
			hasWindow = true
		}
		return nodeCount <= reportObjectSQLMaximumNodes, nil
	}, statement)
	if nodeCount > reportObjectSQLMaximumNodes {
		return reportmodel.ReportObjectSQLPlan{}, objectSQLPlanError("backend.report.object_sql_complexity_exceeded", "object_sql_v1.sql", map[string]string{"limit": strconv.Itoa(reportObjectSQLMaximumNodes)})
	}
	if hasWindow {
		return reportmodel.ReportObjectSQLPlan{}, objectSQLPlanError("backend.report.object_sql_window_forbidden", "object_sql_v1.sql", nil)
	}
	plan := reportmodel.ReportObjectSQLPlan{Parameters: compiler.parameters, NodeCount: nodeCount}
	if err := compiler.compileSources(selectStatement.From, &plan); err != nil {
		return reportmodel.ReportObjectSQLPlan{}, err
	}
	if err := compiler.compileSelect(selectStatement, &plan); err != nil {
		return reportmodel.ReportObjectSQLPlan{}, err
	}
	if err := compiler.validateAmplification(plan); err != nil {
		return reportmodel.ReportObjectSQLPlan{}, err
	}
	return plan, nil
}

func (c *objectSQLCompiler) compileMetadata() error {
	for index, raw := range c.schema.SourceObjects {
		objectKey := strings.TrimSpace(raw)
		if objectKey == "" || c.declaredObjects[objectKey] || c.objects[objectKey].Key == "" {
			return objectSQLPlanError("backend.report.source_object_not_found", fmt.Sprintf("object_sql_v1.source_objects[%d]", index), map[string]string{"object": objectKey})
		}
		c.declaredObjects[objectKey] = true
	}
	for index, parameter := range c.schema.Parameters {
		key, parameterType := strings.TrimSpace(parameter.Key), strings.TrimSpace(parameter.Type)
		if key == "" || c.parameters[key].Key != "" || !objectSQLValueType(parameterType) || strings.HasPrefix(key, "__") {
			return objectSQLPlanError("backend.report.object_sql_parameter_invalid", fmt.Sprintf("object_sql_v1.parameters[%d]", index), map[string]string{"parameter": key, "type": parameterType})
		}
		parameter.Key, parameter.Type = key, parameterType
		c.parameters[key] = parameter
	}
	for index, result := range c.schema.ResultSchema {
		key := strings.TrimSpace(result.Key)
		if key == "" || c.results[key].Key != "" {
			return objectSQLPlanError("backend.report.object_sql_result_schema_invalid", fmt.Sprintf("object_sql_v1.result_schema[%d].key", index), map[string]string{
				"field": "key", "invalid_field": "key", "result_key": key, "actual": key, "reason": "empty_or_duplicate_result_alias",
			})
		}
		result.Key = key
		c.results[key] = result
		c.resultIndex[key] = index
	}
	for index, contract := range c.schema.JoinCardinalities {
		alias, cardinality := strings.TrimSpace(contract.Alias), strings.TrimSpace(contract.Cardinality)
		if alias == "" || c.cardinality[alias] != "" || cardinality != "one_to_one" && cardinality != "many_to_one" && cardinality != "one_to_many" {
			return objectSQLPlanError("backend.report.join_cardinality_invalid", fmt.Sprintf("object_sql_v1.join_cardinalities[%d]", index), map[string]string{
				"alias": alias, "actual": cardinality, "allowed_values": "one_to_one,many_to_one,one_to_many",
			})
		}
		c.cardinality[alias] = cardinality
		c.cardinalityPath[alias] = fmt.Sprintf("object_sql_v1.join_cardinalities[%d]", index)
	}
	if c.schema.TimeoutMilliseconds < 0 || c.schema.TimeoutMilliseconds > 30000 {
		return objectSQLPlanError("backend.report.object_sql_timeout_invalid", "object_sql_v1.timeout_milliseconds", nil)
	}
	return nil
}

// CanonicalReportObjectSQL compiles one authoring definition and returns the
// publication form. Structural facts come from the bound SQL plan; legacy
// source/result/cardinality declarations are assertions only.
func CanonicalReportObjectSQL(schema reportmodel.ReportObjectSQLSchema, objects map[string]reportengine.Object) (reportmodel.ReportObjectSQLSchema, reportmodel.ReportObjectSQLPlan, error) {
	plan, err := CompileReportObjectSQL(schema, objects)
	if err != nil {
		return reportmodel.ReportObjectSQLSchema{}, reportmodel.ReportObjectSQLPlan{}, err
	}
	canonical := schema
	canonical.SourceObjects = nil
	seen := map[string]bool{}
	for _, source := range plan.Sources {
		if !seen[source.ObjectKey] {
			seen[source.ObjectKey] = true
			canonical.SourceObjects = append(canonical.SourceObjects, source.ObjectKey)
		}
	}
	canonical.ResultSchema = append([]reportmodel.ReportResultColumnSchema(nil), plan.ResultSchema...)
	canonical.JoinCardinalities = nil
	return canonical, plan, nil
}

func (c *objectSQLCompiler) compileSources(from []sqlparser.TableExpr, plan *reportmodel.ReportObjectSQLPlan) error {
	if len(from) != 1 {
		return objectSQLPlanError("backend.report.object_sql_source_invalid", "object_sql_v1.sql.from", map[string]string{"reason": "exactly one joined source tree is required"})
	}
	if err := c.compileSourceNode(from[0], plan, true); err != nil {
		return err
	}
	if len(plan.Sources)-1 > reportObjectSQLMaximumJoins {
		return objectSQLPlanError("backend.report.object_sql_complexity_exceeded", "object_sql_v1.sql.from", map[string]string{"join_limit": strconv.Itoa(reportObjectSQLMaximumJoins)})
	}
	joinAliases := map[string]bool{}
	for _, source := range plan.Sources[1:] {
		joinAliases[source.Alias] = true
	}
	for alias, actual := range c.cardinality {
		if !joinAliases[alias] {
			return objectSQLPlanError("backend.report.join_cardinality_invalid", c.cardinalityPath[alias], map[string]string{
				"alias": alias, "actual": actual, "reason": "legacy cardinality assertion names no parsed JOIN alias",
			})
		}
	}
	usedObjects := map[string]bool{}
	for _, source := range plan.Sources {
		usedObjects[source.ObjectKey] = true
	}
	if len(c.declaredObjects) > 0 && len(usedObjects) != len(c.declaredObjects) {
		return objectSQLPlanError("backend.report.source_object_invalid", "object_sql_v1.source_objects", map[string]string{"reason": "declared objects must exactly match parsed base objects"})
	}
	for objectKey := range c.declaredObjects {
		if !usedObjects[objectKey] {
			return objectSQLPlanError("backend.report.source_object_invalid", "object_sql_v1.source_objects", map[string]string{"object": objectKey})
		}
	}
	return nil
}

func (c *objectSQLCompiler) compileSourceNode(node sqlparser.TableExpr, plan *reportmodel.ReportObjectSQLPlan, root bool) error {
	switch value := node.(type) {
	case *sqlparser.AliasedTableExpr:
		return c.compileBaseSource(value, plan, root, "", nil)
	case *sqlparser.JoinTableExpr:
		if err := c.compileSourceNode(value.LeftExpr, plan, root); err != nil {
			return err
		}
		joinType := strings.ToLower(strings.TrimSpace(value.Join.ToString()))
		if joinType == "join" || joinType == "inner join" {
			joinType = "inner"
		} else if joinType == "left join" {
			joinType = "left"
		} else {
			return objectSQLPlanError("backend.report.object_sql_join_forbidden", "object_sql_v1.sql.from", map[string]string{"join": joinType})
		}
		right, ok := value.RightExpr.(*sqlparser.AliasedTableExpr)
		if !ok || value.Condition == nil || value.Condition.On == nil || len(value.Condition.Using) > 0 {
			return objectSQLPlanError("backend.report.object_sql_join_invalid", "object_sql_v1.sql.from", nil)
		}
		if err := c.compileBaseSource(right, plan, false, joinType, value.Condition.On); err != nil {
			return err
		}
		return nil
	default:
		return objectSQLPlanError("backend.report.object_sql_source_invalid", "object_sql_v1.sql.from", map[string]string{"node": fmt.Sprintf("%T", node)})
	}
}

func (c *objectSQLCompiler) compileBaseSource(table *sqlparser.AliasedTableExpr, plan *reportmodel.ReportObjectSQLPlan, root bool, joinType string, on sqlparser.Expr) error {
	name, ok := table.Expr.(sqlparser.TableName)
	alias := strings.TrimSpace(table.As.String())
	if !ok || !name.Qualifier.IsEmpty() || alias == "" || len(table.Partitions) > 0 || table.Hints != nil || len(table.Columns) > 0 {
		return objectSQLPlanError("backend.report.object_sql_source_invalid", "object_sql_v1.sql.from", nil)
	}
	objectKey := strings.TrimSpace(name.Name.String())
	object, exists := c.objects[objectKey]
	if !exists || len(c.declaredObjects) > 0 && !c.declaredObjects[objectKey] || c.aliases[alias].Key != "" {
		return objectSQLPlanError("backend.report.source_object_not_found", "object_sql_v1.sql.from", map[string]string{"object": objectKey, "alias": alias})
	}
	c.aliases[alias], c.aliasOrder[alias] = object, len(plan.Sources)
	source := reportmodel.ReportObjectSQLSource{ObjectKey: objectKey, Alias: alias, JoinType: joinType}
	if !root {
		expression, err := c.compileExpression(on, false)
		if err != nil {
			return err
		}
		inferred, candidates := c.inferJoinCardinality(expression, alias)
		if inferred == "" {
			return objectSQLPlanError("backend.report.join_cardinality_unprovable", "object_sql_v1.sql.from", map[string]string{
				"alias": alias, "join": objectKey + " " + alias, "reason": "JOIN must contain an equality backed by a declared relation or unique field", "candidate_relation_fields": strings.Join(candidates, ","),
			})
		}
		if asserted := c.cardinality[alias]; asserted != "" && asserted != inferred {
			return objectSQLPlanError("backend.report.join_cardinality_mismatch", c.cardinalityPath[alias], map[string]string{
				"alias": alias, "actual": asserted, "inferred": inferred, "allowed_values": inferred,
			})
		}
		source.Cardinality = inferred
		source.On = &expression
	}
	plan.Sources = append(plan.Sources, source)
	return nil
}

func (c *objectSQLCompiler) compileSelect(statement *sqlparser.Select, plan *reportmodel.ReportObjectSQLPlan) error {
	if statement.SelectExprs == nil || len(statement.SelectExprs.Exprs) == 0 || len(statement.SelectExprs.Exprs) > reportObjectSQLMaximumColumns {
		return objectSQLPlanError("backend.report.object_sql_result_schema_invalid", "object_sql_v1.sql.select", nil)
	}
	usedResults := map[string]bool{}
	for index, raw := range statement.SelectExprs.Exprs {
		if _, star := raw.(*sqlparser.StarExpr); star {
			return objectSQLPlanError("backend.report.object_sql_star_forbidden", "object_sql_v1.sql.select", nil)
		}
		aliased, ok := raw.(*sqlparser.AliasedExpr)
		alias := ""
		if ok {
			alias = strings.TrimSpace(aliased.As.String())
		}
		if !ok || alias == "" {
			return objectSQLPlanError("backend.report.object_sql_result_schema_invalid", "object_sql_v1.sql.select", map[string]string{
				"field": "key", "invalid_field": "key", "actual": alias, "reason": "every SELECT expression requires a stable result alias",
			})
		}
		if c.projections[alias].Kind != "" {
			return objectSQLPlanError("backend.report.object_sql_result_schema_invalid", "object_sql_v1.sql.select", map[string]string{
				"field": "key", "invalid_field": "key", "result_key": alias, "actual": alias, "reason": "duplicate_result_alias",
			})
		}
		expression, err := c.compileExpression(aliased.Expr, false)
		if err != nil {
			return err
		}
		result, declared := c.results[alias]
		if !declared {
			result = reportmodel.ReportResultColumnSchema{Key: alias}
		}
		resultKey := alias
		resultPathIndex := index
		if declared {
			resultPathIndex = c.resultIndex[alias]
		}
		if result.Type == "" {
			result.Type, result.Precision, result.Scale = expression.Type, expression.Precision, expression.Scale
		} else if !objectSQLResultType(result.Type) {
			return objectSQLPlanError("backend.report.object_sql_result_schema_invalid", fmt.Sprintf("object_sql_v1.result_schema[%d].type", resultPathIndex), map[string]string{
				"field": "type", "invalid_field": "type", "result_key": resultKey, "actual": result.Type, "allowed_values": strings.Join(objectSQLResultTypes(), ","),
			})
		}
		if result.Kind == "" {
			result.Kind = "dimension"
			if objectSQLExpressionAggregate(expression) {
				result.Kind = "measure"
			}
		} else if result.Kind != "dimension" && result.Kind != "measure" {
			params := map[string]string{
				"field": "kind", "invalid_field": "kind", "result_key": resultKey, "actual": result.Kind, "allowed_values": "dimension,measure",
			}
			if result.Kind == "metric" {
				params["replacement_value"] = "measure"
			}
			return objectSQLPlanError("backend.report.object_sql_result_schema_invalid", fmt.Sprintf("object_sql_v1.result_schema[%d].kind", resultPathIndex), params)
		}
		if !objectSQLResultCompatible(result, expression) {
			return objectSQLPlanError("backend.report.object_sql_result_schema_invalid", fmt.Sprintf("object_sql_v1.result_schema[%d].type", resultPathIndex), map[string]string{
				"field": "type", "invalid_field": "type", "result_key": resultKey, "declared": result.Type, "actual": expression.Type, "reason": "expression_type_incompatible",
			})
		}
		if (result.Type == "decimal" || result.Type == "currency") && expression.Precision > 0 {
			if result.Precision == 0 {
				result.Precision = expression.Precision
			}
			if result.Scale == 0 {
				result.Scale = expression.Scale
			}
		}
		usedResults[alias] = true
		c.projections[alias] = expression
		plan.Projections = append(plan.Projections, reportmodel.ReportObjectSQLProjection{Alias: alias, Expression: expression})
		plan.ResultSchema = append(plan.ResultSchema, result)
	}
	for key := range c.results {
		if !usedResults[key] {
			return objectSQLPlanError("backend.report.object_sql_result_schema_invalid", "object_sql_v1.result_schema", map[string]string{
				"field": "key", "invalid_field": "key", "result_key": key, "actual": key, "reason": "result schema alias is not selected by SQL",
			})
		}
	}
	var err error
	if statement.Where != nil {
		expression, compileErr := c.compileExpression(statement.Where.Expr, false)
		err = compileErr
		plan.Where = &expression
	}
	if err != nil {
		return err
	}
	if statement.GroupBy != nil {
		for _, raw := range statement.GroupBy.Exprs {
			expression, compileErr := c.compileExpression(raw, false)
			if compileErr != nil {
				return compileErr
			}
			plan.GroupBy = append(plan.GroupBy, expression)
		}
	}
	if statement.Having != nil {
		expression, compileErr := c.compileExpression(statement.Having.Expr, false)
		if compileErr != nil {
			return compileErr
		}
		plan.Having = &expression
	}
	plan.ExplicitOrderBy = len(statement.OrderBy) > 0
	for _, order := range statement.OrderBy {
		expression, compileErr := c.compileExpression(order.Expr, true)
		if compileErr != nil {
			return compileErr
		}
		direction := strings.ToLower(strings.TrimSpace(order.Direction.ToString()))
		if direction == "" {
			direction = "asc"
		}
		if direction != "asc" && direction != "desc" {
			return objectSQLPlanError("backend.report.object_sql_order_invalid", "object_sql_v1.sql.order_by", nil)
		}
		plan.OrderBy = append(plan.OrderBy, reportmodel.ReportObjectSQLOrder{Expression: expression, Direction: direction})
	}
	plan.Limit = reportObjectSQLDefaultLimit
	plan.ExplicitLimit = statement.Limit != nil
	if statement.Limit != nil {
		if statement.Limit.Offset != nil {
			return objectSQLPlanError("backend.report.object_sql_limit_invalid", "object_sql_v1.sql.limit", nil)
		}
		literal, ok := statement.Limit.Rowcount.(*sqlparser.Literal)
		if !ok {
			return objectSQLPlanError("backend.report.object_sql_limit_invalid", "object_sql_v1.sql.limit", nil)
		}
		value, parseErr := strconv.Atoi(literal.Val)
		if parseErr != nil || value <= 0 || value > reportObjectSQLMaximumLimit {
			return objectSQLPlanError("backend.report.object_sql_limit_invalid", "object_sql_v1.sql.limit", map[string]string{"maximum": strconv.Itoa(reportObjectSQLMaximumLimit)})
		}
		plan.Limit = value
	}
	appendObjectSQLStableOrder(plan)
	c.populateSourceFields(plan)
	return nil
}

// appendObjectSQLStableOrder is a Report definition-compiler detail. Authors
// keep writing native SQL; the contract adds only the deterministic tie terms
// needed by opaque cursor traversal and never exposes a report AST or ordering DSL.
func appendObjectSQLStableOrder(plan *reportmodel.ReportObjectSQLPlan) {
	if len(plan.GroupBy) > 0 {
		for _, expression := range plan.GroupBy {
			if !objectSQLOrderContains(plan.OrderBy, expression) {
				plan.OrderBy = append(plan.OrderBy, reportmodel.ReportObjectSQLOrder{Expression: expression, Direction: "asc"})
			}
		}
		return
	}
	for _, projection := range plan.Projections {
		if objectSQLExpressionAggregate(projection.Expression) {
			return // aggregate without GROUP BY has at most one result row
		}
	}
	for _, source := range plan.Sources {
		expression := reportmodel.ReportObjectSQLExpression{Kind: "field", Alias: source.Alias, FieldKey: "id", Type: "text"}
		if !objectSQLOrderContains(plan.OrderBy, expression) {
			plan.OrderBy = append(plan.OrderBy, reportmodel.ReportObjectSQLOrder{Expression: expression, Direction: "asc"})
		}
	}
}

// ReportObjectSQLDefaultLimitRows and ReportObjectSQLMaximumLimitRows expose
// the compiled limit contract to validation layers that explain why authored
// SQL must carry its own literal LIMIT.
const (
	ReportObjectSQLDefaultLimitRows = reportObjectSQLDefaultLimit
	ReportObjectSQLMaximumLimitRows = reportObjectSQLMaximumLimit
)

// ReportObjectSQLPlanSingleRow reports whether the compiled plan can only
// produce a single result row: an aggregate projection without GROUP BY. Such
// plans need no authored ORDER BY or LIMIT; every other plan can return
// multiple rows and silently truncates at the implicit default limit.
func ReportObjectSQLPlanSingleRow(plan reportmodel.ReportObjectSQLPlan) bool {
	if len(plan.GroupBy) > 0 {
		return false
	}
	for _, projection := range plan.Projections {
		if objectSQLExpressionAggregate(projection.Expression) {
			return true
		}
	}
	return false
}

func objectSQLOrderContains(orders []reportmodel.ReportObjectSQLOrder, expression reportmodel.ReportObjectSQLExpression) bool {
	for _, order := range orders {
		candidate := order.Expression
		if candidate.Kind == expression.Kind && candidate.Alias == expression.Alias && candidate.FieldKey == expression.FieldKey && candidate.Name == expression.Name {
			return true
		}
	}
	return false
}

func objectSQLExpressionAggregate(expression reportmodel.ReportObjectSQLExpression) bool {
	if expression.Kind == "aggregate" {
		return true
	}
	for _, argument := range expression.Arguments {
		if objectSQLExpressionAggregate(argument) {
			return true
		}
	}
	for _, when := range expression.Whens {
		if objectSQLExpressionAggregate(when.Condition) || objectSQLExpressionAggregate(when.Value) {
			return true
		}
	}
	return expression.Else != nil && objectSQLExpressionAggregate(*expression.Else)
}

func objectSQLPlanError(code, path string, params map[string]string) error {
	return &reportmodel.ReportObjectSQLPlanError{Code: code, Path: path, Params: params}
}

func objectSQLValueType(value string) bool {
	switch value {
	case "text", "integer", "number", "decimal", "boolean", "date", "datetime":
		return true
	}
	return false
}
func objectSQLResultType(value string) bool { return objectSQLValueType(value) || value == "currency" }
func objectSQLResultTypes() []string {
	return []string{"text", "integer", "number", "decimal", "boolean", "date", "datetime", "currency"}
}
