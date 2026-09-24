package ssm

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

type Config struct {
	Path string `envconfig:"SSM_PATH"`
}

type Param struct {
	Name  string
	Value string
}

func GetParametersFromPath(ctx context.Context, path string) ([]Param, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	ssmClient := ssm.NewFromConfig(awsCfg)

	var params []types.Parameter

	pager := ssm.NewGetParametersByPathPaginator(ssmClient, &ssm.GetParametersByPathInput{
		Path:           aws.String(path),
		WithDecryption: aws.Bool(true),
	})
	for pager.HasMorePages() {
		res, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("ssm: get parameters by path: %w", err)
		}

		params = append(params, res.Parameters...)
	}

	tbr := make([]Param, len(params))
	for i, p := range params {
		tbr[i] = Param{
			Name:  strings.TrimLeft(strings.Replace(aws.ToString(p.Name), path, "", 1), "/"),
			Value: aws.ToString(p.Value),
		}
	}

	return tbr, nil
}

func LoadIntoEnv(in []Param) error {
	for _, v := range in {
		if err := os.Setenv(v.Name, v.Value); err != nil {
			return fmt.Errorf("os: setenv: %w", err)
		}
	}

	return nil
}
