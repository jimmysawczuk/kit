package dtable

import (
	"context"
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

type Model interface {
	PK() string
	SK() string
}

type Columner interface {
	Columns() []string
}

func Columns(m Model) []string {
	if c, ok := m.(Columner); ok {
		return c.Columns()
	}

	fields := reflect.VisibleFields(reflect.TypeOf(m))
	cols := []string{}

	for _, field := range fields {
		col, ok := field.Tag.Lookup("dynamodbav")
		if ok {
			if col == "-" {
				continue
			}

			cols = append(cols, col)
			continue
		}

		cols = append(cols, field.Name)
	}

	return cols
}

func AsProjectionBuilder(cols []string) expression.ProjectionBuilder {
	in := make([]expression.NameBuilder, len(cols))
	for i := range cols {
		in[i] = expression.Name(cols[i])
	}

	b := expression.ProjectionBuilder{}
	b.AddNames(in...)
	return b
}

type Table struct {
	conn *dynamodb.DynamoDB
	name string
}

func New(conn *dynamodb.DynamoDB, name string) *Table {
	return &Table{
		conn: conn,
		name: name,
	}
}

func (t *Table) Put(ctx context.Context, model Model) error {
	m, err := dynamodbattribute.MarshalMap(model)
	if err != nil {
		return fmt.Errorf("dynamodbattr: marshal: %w", err)
	}

	m["PK"] = &dynamodb.AttributeValue{S: aws.String(model.PK())}
	m["SK"] = &dynamodb.AttributeValue{S: aws.String(model.SK())}

	if _, err := t.conn.PutItemWithContext(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(t.name),
		Item:      m,
	}); err != nil {
		return fmt.Errorf("dynamodb: put item: %w", err)
	}

	return nil
}

func (t *Table) Get(ctx context.Context, model Model) error {
	res, err := t.conn.GetItemWithContext(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(t.name),
		Key: map[string]*dynamodb.AttributeValue{
			"PK": {S: aws.String(model.PK())},
			"SK": {S: aws.String(model.SK())},
		},
	})
	if err != nil {
		return fmt.Errorf("dynamodb: get: %w", err)
	}

	if err := dynamodbattribute.UnmarshalMap(res.Item, model); err != nil {
		return fmt.Errorf("dynamodbattr: unmarshal: %w", err)
	}

	return nil
}

// func (t *Table) Query(ctx context.Context, m Model) error {
// 	expr, err := expression.NewBuilder().
// 		WithKeyCondition(
// 			expression.KeyAnd(
// 				expression.Key("PK").Equal(expression.Value("User")),
// 				expression.Key("Email").Equal(expression.Value(email)),
// 			),
// 		).
// 		WithProjection(AsProjectionBuilder(Columns(m))).
// 		Build()
// 	if err != nil {
// 		return fmt.Errorf("expression: build: %w", err)
// 	}

// 	res, err := t.conn.QueryWithContext(ctx, &dynamodb.QueryInput{
// 		TableName:                 aws.String(t.name),
// 		IndexName:                 emailIndex,
// 		ExpressionAttributeNames:  expr.Names(),
// 		ExpressionAttributeValues: expr.Values(),
// 		FilterExpression:          expr.Filter(),
// 		KeyConditionExpression:    expr.KeyCondition(),
// 		ProjectionExpression:      expr.Projection(),
// 	})
// 	if err != nil {
// 		return fmt.Errorf("dynamodb: query: %w", err)
// 	}

// 	if err := dynamodbattribute.UnmarshalMap(res.Items[0], &user); err != nil {
// 		return fmt.Errorf("dynamodbattr: unmarshal: %w", err)
// 	}

// 	return nil
// }

func (t *Table) Delete(ctx context.Context, model Model) error {
	if _, err := t.conn.DeleteItemWithContext(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(t.name),
		Key: map[string]*dynamodb.AttributeValue{
			"PK": {S: aws.String(model.PK())},
			"SK": {S: aws.String(model.SK())},
		},
	}); err != nil {
		return fmt.Errorf("dynamodb: delete item: %w", err)
	}

	return nil
}
