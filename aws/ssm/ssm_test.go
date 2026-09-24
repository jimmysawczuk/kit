package ssm

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeClient struct {
	pages []*ssm.GetParametersByPathOutput
	err   error
	calls []*ssm.GetParametersByPathInput
}

func (f *fakeClient) GetParametersByPath(_ context.Context, in *ssm.GetParametersByPathInput, _ ...func(*ssm.Options)) (*ssm.GetParametersByPathOutput, error) {
	f.calls = append(f.calls, in)
	if f.err != nil {
		return nil, f.err
	}
	return f.pages[len(f.calls)-1], nil
}

func param(name, value string) types.Parameter {
	return types.Parameter{Name: aws.String(name), Value: aws.String(value)}
}

func TestGetParametersFromPath(t *testing.T) {
	for _, path := range []string{"/app", "/app/"} {
		t.Run(path, func(t *testing.T) {
			client := &fakeClient{pages: []*ssm.GetParametersByPathOutput{
				{Parameters: []types.Parameter{param("/app/DB_HOST", "db"), param("/app/DB_USER", "u")}, NextToken: aws.String("t1")},
				{Parameters: []types.Parameter{param("/app/DB_PASS", "p")}},
			}}

			params, err := GetParametersFromPath(context.Background(), client, path)
			require.NoError(t, err)
			assert.Equal(t, map[string]string{"DB_HOST": "db", "DB_USER": "u", "DB_PASS": "p"}, params)

			require.Len(t, client.calls, 2)
			assert.Nil(t, client.calls[0].NextToken)
			assert.Equal(t, "t1", aws.ToString(client.calls[1].NextToken))
			assert.Equal(t, path, aws.ToString(client.calls[0].Path))
			assert.True(t, aws.ToBool(client.calls[0].WithDecryption))
		})
	}
}

func TestGetParametersFromPathError(t *testing.T) {
	wantErr := errors.New("boom")
	_, err := GetParametersFromPath(context.Background(), &fakeClient{err: wantErr}, "/app")
	assert.ErrorIs(t, err, wantErr)
}

func TestLoadIntoEnv(t *testing.T) {
	t.Setenv("KIT_SSM_TEST_EXISTING", "old")
	t.Setenv("KIT_SSM_TEST_NEW", "")
	os.Unsetenv("KIT_SSM_TEST_NEW")

	require.NoError(t, LoadIntoEnv(map[string]string{
		"KIT_SSM_TEST_EXISTING": "new",
		"KIT_SSM_TEST_NEW":      "v",
	}))

	assert.Equal(t, "new", os.Getenv("KIT_SSM_TEST_EXISTING"))
	assert.Equal(t, "v", os.Getenv("KIT_SSM_TEST_NEW"))
}
