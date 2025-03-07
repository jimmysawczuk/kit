package dtable

import (
	"context"
	"fmt"
	"reflect"
	"time"

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

func (t *Table) Create(ctx context.Context, model Model) error {
	m, err := dynamodbattribute.MarshalMap(model)
	if err != nil {
		return fmt.Errorf("dynamodbattr: marshal: %w", err)
	}

	m["PK"] = &dynamodb.AttributeValue{S: aws.String(model.PK())}
	m["SK"] = &dynamodb.AttributeValue{S: aws.String(model.SK())}

	expr, err := expression.NewBuilder().WithCondition(
		expression.ConditionBuilder(expression.And(
			expression.AttributeNotExists(expression.Name("PK")),
			expression.AttributeNotExists(expression.Name("SK")),
		)),
	).Build()
	if err != nil {
		return fmt.Errorf("expression: build: %w", err)
	}

	if _, err := t.conn.PutItemWithContext(ctx, &dynamodb.PutItemInput{
		TableName:                aws.String(t.name),
		Item:                     m,
		ExpressionAttributeNames: expr.Names(),
		ConditionExpression:      expr.Condition(),
	}); err != nil {
		return fmt.Errorf("dynamodb: put item: %w", err)
	}

	return nil
}

func (t *Table) Update(ctx context.Context, model Model) error {
	update := expression.Set(expression.Name("UpdatedAt"), expression.Value(time.Now().UTC()))
	rv := reflect.ValueOf(model)
	for i := range rv.NumField() {
		field := rv.Field(i)
		tag, ok := rv.Type().Field(i).Tag.Lookup("dynamodbav")
		if ok {
			if tag == "-" {
				continue
			}

			update.Set(expression.Name(tag), expression.Value(field.Interface()))
			continue
		}

		update.Set(expression.Name(field.Type().Name()), expression.Value(field.Interface()))
	}

	expr, err := expression.NewBuilder().WithUpdate(update).
		WithCondition(
			expression.And(
				expression.Equal(expression.Name("PK"), expression.Value(model.PK())),
				expression.Equal(expression.Name("SK"), expression.Value(model.SK())),
			),
		).Build()
	if err != nil {
		return fmt.Errorf("expression: build: %w", err)
	}

	if _, err := t.conn.UpdateItemWithContext(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(t.name),
		Key: map[string]*dynamodb.AttributeValue{
			"PK": {S: aws.String(model.PK())},
			"SK": {S: aws.String(model.SK())},
		},
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ConditionExpression:       expr.Condition(),
		UpdateExpression:          expr.Update(),
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
