package ssm

import (
	"context"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/pkg/errors"
)

type Config struct {
	Path string `envconfig:"SSM_PATH"`
}

type Param struct {
	Name  string
	Value string
}

func GetParametersFromPath(ctx context.Context, path string) ([]Param, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, errors.Wrap(err, "session: new session")
	}

	ssmClient := ssm.New(sess)

	var tok *string
	var params []*ssm.Parameter

	for {
		res, err := ssmClient.GetParametersByPathWithContext(ctx, &ssm.GetParametersByPathInput{
			NextToken:      tok,
			Path:           aws.String(path),
			WithDecryption: aws.Bool(true),
		})
		if err != nil {
			return nil, errors.Wrap(err, "ssm: get parameters by path")
		}

		params = append(params, res.Parameters...)

		if res.NextToken == nil {
			break
		}

		tok = res.NextToken
	}

	tbr := make([]Param, len(params))
	for i, p := range params {
		tbr[i] = Param{
			Name:  strings.TrimLeft(strings.Replace(aws.StringValue(p.Name), path, "", 1), "/"),
			Value: aws.StringValue(p.Value),
		}
	}

	return tbr, nil
}

func LoadIntoEnv(in []Param) error {
	for _, v := range in {
		if err := os.Setenv(v.Name, v.Value); err != nil {
			return errors.Wrap(err, "os: setenv")
		}
	}

	return nil
}
