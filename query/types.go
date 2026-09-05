// Package query contains the portable metadata types consumed by Report's
// definition compiler. It deliberately excludes host authorization, physical
// storage, and Runtime object-schema details.
package query

type Field struct {
	Key                 string
	Type                string
	Precision           int
	Scale               int32
	Unique              bool
	RelationTarget      string
	RelationCardinality string
}

type Object struct {
	Key    string
	Fields []Field
}
