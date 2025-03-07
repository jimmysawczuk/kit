package dtable_test

import (
	"testing"

	"github.com/jimmysawczuk/kit/aws/dtable/v2"
	"github.com/stretchr/testify/require"
)

type model struct {
	ID       string
	Username string `dynamodbav:"User"`
	Password string `dynamodbav:"-"`
}

func (m model) PK() string {
	return "U"
}

func (m model) SK() string {
	return "U-" + m.ID
}

func TestDefaultColumns(t *testing.T) {
	m := model{}

	require.Equal(t, []string{"ID", "User"}, dtable.Columns(m))
}

type model2 struct {
	ID       string
	Username string `dynamodbav:"User"`
	Password string `dynamodbav:"-"`
}

func (m model2) PK() string {
	return "U"
}

func (m model2) SK() string {
	return "U-" + m.ID
}

func (m model2) Columns() []string {
	return []string{
		"ID",
		"User",
		"Foo",
		"Bar",
	}
}

func TestOverriddenColumns(t *testing.T) {
	m := model2{}

	require.Equal(t, []string{"ID", "User", "Foo", "Bar"}, dtable.Columns(m))
}
