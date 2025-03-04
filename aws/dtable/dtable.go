package dtable

import (
	"context"
	"fmt"
	"runtime"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/rs/zerolog"
)

type Table struct {
	conn *dynamodb.DynamoDB

	Name string
}

func New(conn *dynamodb.DynamoDB, name string) *Table {
	return &Table{
		conn: conn,
		Name: name,
	}
}

func (t *Table) Conn() *dynamodb.DynamoDB {
	return t.conn
}

func debug(ctx context.Context, table, name string, in any) {
	log := zerolog.Ctx(ctx)

	if log != nil {
		_, file, line, _ := runtime.Caller(3)

		log.Debug().
			Str("table", table).
			Str("file", file).
			Int("line", line).
			Any("in", in).
			Msgf("dtable: debug: %s", name)
	}
}

func (t *Table) PutItem(ctx context.Context, in *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	debug(ctx, t.Name, "PutItem", in)

	in.TableName = aws.String(t.Name)
	return t.conn.PutItemWithContext(ctx, in)
}

func (t *Table) GetItem(ctx context.Context, in *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	debug(ctx, t.Name, "PutItem", in)

	in.TableName = aws.String(t.Name)
	return t.conn.GetItemWithContext(ctx, in)
}

type BatchGetItemInput struct {
	KeysAndAttributes      *dynamodb.KeysAndAttributes
	ReturnConsumedCapacity *string
}

type BatchGetItemOutput struct {
	ConsumedCapacity []*dynamodb.ConsumedCapacity
	Responses        []map[string]*dynamodb.AttributeValue
	UnprocessedKeys  map[string]*dynamodb.KeysAndAttributes
}

type Key struct {
	PK string
	SK string
}

func GetKeysAndAttributes(pk, sk string, keys []Key) []map[string]*dynamodb.AttributeValue {
	v := make([]map[string]*dynamodb.AttributeValue, len(keys))
	for i, k := range keys {
		v[i] = map[string]*dynamodb.AttributeValue{
			pk: {S: aws.String(k.PK)},
			sk: {S: aws.String(k.SK)},
		}
	}
	return v
}

func (t *Table) BatchGetItem(ctx context.Context, in *BatchGetItemInput) (*BatchGetItemOutput, error) {
	debug(ctx, t.Name, "BatchGetItem", in)

	din := dynamodb.BatchGetItemInput{
		RequestItems: map[string]*dynamodb.KeysAndAttributes{
			t.Name: in.KeysAndAttributes,
		},
		ReturnConsumedCapacity: in.ReturnConsumedCapacity,
	}

	dout, err := t.conn.BatchGetItemWithContext(ctx, &din)
	if err != nil {
		return nil, fmt.Errorf("batch get item with context: %w", err)
	}

	return &BatchGetItemOutput{
		ConsumedCapacity: dout.ConsumedCapacity,
		Responses:        dout.Responses[t.Name],
		UnprocessedKeys:  dout.UnprocessedKeys,
	}, nil
}

func (t *Table) UpdateItem(ctx context.Context, in *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error) {
	debug(ctx, t.Name, "UpdateItem", in)

	in.TableName = aws.String(t.Name)
	return t.conn.UpdateItemWithContext(ctx, in)
}

func (t *Table) DeleteItem(ctx context.Context, in *dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error) {
	debug(ctx, t.Name, "DeleteItem", in)

	in.TableName = aws.String(t.Name)
	return t.conn.DeleteItemWithContext(ctx, in)
}

func (t *Table) Query(ctx context.Context, in *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
	debug(ctx, t.Name, "Query", in)

	in.TableName = aws.String(t.Name)
	return t.conn.QueryWithContext(ctx, in)
}
