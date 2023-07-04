package dtable

import (
	"fmt"
	"runtime"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"go.uber.org/zap"
)

type Table struct {
	conn *dynamodb.DynamoDB
	log  *zap.Logger

	Name string
}

func New(conn *dynamodb.DynamoDB, name string, log *zap.Logger) *Table {
	return &Table{
		conn: conn,
		log:  log,
		Name: name,
	}
}

func (t *Table) Conn() *dynamodb.DynamoDB {
	return t.conn
}

func debug(log *zap.Logger, name string, in any) {
	if log != nil {
		_, file, line, _ := runtime.Caller(3)

		log.With(
			zap.String("table", name),
			zap.String("file", file),
			zap.Int("line", line),
			// zap.Any("in", in),
		).Debug(name)
	}
}

func (t *Table) PutItem(ctx aws.Context, in *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	debug(t.log, "PutItem", in)

	in.TableName = aws.String(t.Name)
	return t.conn.PutItemWithContext(ctx, in)
}

func (t *Table) GetItem(ctx aws.Context, in *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	debug(t.log, "PutItem", in)

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

func (t *Table) BatchGetItem(ctx aws.Context, in *BatchGetItemInput) (*BatchGetItemOutput, error) {
	debug(t.log, "BatchGetItem", in)

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

func (t *Table) UpdateItem(ctx aws.Context, in *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error) {
	debug(t.log, "UpdateItem", in)

	in.TableName = aws.String(t.Name)
	return t.conn.UpdateItemWithContext(ctx, in)
}

func (t *Table) DeleteItem(ctx aws.Context, in *dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error) {
	debug(t.log, "DeleteItem", in)

	in.TableName = aws.String(t.Name)
	return t.conn.DeleteItemWithContext(ctx, in)
}

func (t *Table) Query(ctx aws.Context, in *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
	debug(t.log, "Query", in)

	in.TableName = aws.String(t.Name)
	return t.conn.QueryWithContext(ctx, in)
}
