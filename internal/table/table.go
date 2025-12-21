// Package table provides table metadata structures for sqlcredo.
// It defines the Info type that holds table name and ID column information.
package table

type Info struct {
	Name     string
	IDColumn string
}
