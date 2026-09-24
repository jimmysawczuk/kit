package ssm

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// GetParametersFromPath returns the decrypted parameters directly under path,
// keyed by name with the path prefix stripped.
func GetParametersFromPath(ctx context.Context, client ssm.GetParametersByPathAPIClient, path string) (map[string]string, error) {
	prefix := strings.TrimSuffix(path, "/") + "/"
	params := map[string]string{}

	pager := ssm.NewGetParametersByPathPaginator(client, &ssm.GetParametersByPathInput{
		Path:           aws.String(path),
		WithDecryption: aws.Bool(true),
	})
	for pager.HasMorePages() {
		res, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("ssm: get parameters by path: %w", err)
		}

		for _, p := range res.Parameters {
			params[strings.TrimPrefix(aws.ToString(p.Name), prefix)] = aws.ToString(p.Value)
		}
	}

	return params, nil
}

// LoadIntoEnv sets each param as an environment variable, overwriting any
// existing value.
func LoadIntoEnv(params map[string]string) error {
	for k, v := range params {
		if err := os.Setenv(k, v); err != nil {
			return fmt.Errorf("os: setenv: %w", err)
		}
	}

	return nil
}
