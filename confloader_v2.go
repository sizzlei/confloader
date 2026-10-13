package confloader

// aws-sdk-go-v2 기반 로더입니다.
//
// 기존 AWSParamLoader·AWSSecretLoader(aws-sdk-go v1)는 사용 중인 프로젝트가 많아 그대로 두고,
// 같은 규격을 v2 SDK 로 읽는 함수를 별도로 제공합니다. 반환 타입이 같으므로
// Keyload·Conflist·InterfaceToSlice 는 V1·V2 결과에 똑같이 사용할 수 있습니다.
//
// v2 기본 자격 증명 체인은 EKS Pod Identity(169.254.170.23)를 지원합니다.
// 구버전 v1 SDK 는 이 엔드포인트를 거부하므로 Pod Identity 환경에서는 V2 함수를 사용해야 합니다.

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"gopkg.in/yaml.v3"
)

// AWSParamLoaderV2 는 AWSParamLoader 의 aws-sdk-go-v2 버전입니다.
// r 은 리전, k 는 Parameter Store 키 이름이며, 값은 복호화(SecureString)해서 읽습니다.
// optFns 는 config.LoadDefaultConfig 에 그대로 전달되므로 프로필·엔드포인트 등을 덧붙일 수 있습니다.
func AWSParamLoaderV2(ctx context.Context, r string, k string, optFns ...func(*config.LoadOptions) error) (Param, error) {
	var c Param

	cfg, err := loadAWSConfig(ctx, r, optFns)
	if err != nil {
		return c, err
	}

	p, err := ssm.NewFromConfig(cfg).GetParameter(ctx, &ssm.GetParameterInput{
		Name:           aws.String(k),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return c, fmt.Errorf("SSM 파라미터 조회 실패(key=%s): %w", k, err)
	}
	if p.Parameter == nil || p.Parameter.Value == nil {
		return c, fmt.Errorf("SSM 파라미터 값 없음(key=%s)", k)
	}

	err = yaml.Unmarshal([]byte(*p.Parameter.Value), &c)
	if err != nil {
		return c, err
	}

	return c, nil
}

// AWSSecretLoaderV2 는 AWSSecretLoader 의 aws-sdk-go-v2 버전입니다.
// r 은 리전, k 는 Secret 이름 또는 ARN 이며, SecretString(JSON)을 map 으로 반환합니다.
// optFns 는 config.LoadDefaultConfig 에 그대로 전달됩니다.
func AWSSecretLoaderV2(ctx context.Context, r string, k string, optFns ...func(*config.LoadOptions) error) (map[string]interface{}, error) {
	cfg, err := loadAWSConfig(ctx, r, optFns)
	if err != nil {
		return nil, err
	}

	secv, err := secretsmanager.NewFromConfig(cfg).GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(k),
	})
	if err != nil {
		return nil, fmt.Errorf("Secret 조회 실패(key=%s): %w", k, err)
	}
	// SecretBinary 로 저장된 Secret 은 SecretString 이 비어 있습니다.
	// V1 은 이 경우 nil 역참조로 panic 하므로 V2 는 오류로 반환합니다.
	if secv.SecretString == nil {
		return nil, fmt.Errorf("Secret 문자열 값 없음(key=%s)", k)
	}

	data := make(map[string]interface{})
	err = json.Unmarshal([]byte(*secv.SecretString), &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// loadAWSConfig 는 리전을 지정해 v2 기본 설정을 로드합니다.
// 리전 옵션을 먼저 두므로 optFns 에 config.WithRegion 이 있으면 그 값이 우선합니다.
func loadAWSConfig(ctx context.Context, r string, optFns []func(*config.LoadOptions) error) (aws.Config, error) {
	opts := append([]func(*config.LoadOptions) error{config.WithRegion(r)}, optFns...)

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return cfg, fmt.Errorf("AWS 설정 로드 실패(region=%s): %w", r, err)
	}

	return cfg, nil
}
