# confloader

## Desciprtion
Confloader는 Configure를 로드 하기 위해 간단히 작성된 패키지 입니다. 파일을 로드하여 설정을 반환하거나, AWS SSM Parameter Store의 값을 읽고 설정을 반환 합니다. 


## Installation
패키지 설치 구문은 아래와 같습니다.
```
go get github.com/sizzlei/confloader
```

## Configure Example
Configure 파일 또는 Parameter Store의 내용은 아래 규격을 기준으로 합니다.
```yaml
Param:  
  - ConfigId : {Configure Name}
    Conf:
      {conf_name} : {conf_value}
      {conf_name} : {conf_value}
      .
      .
      .

  - ConfigId : {Configure Name}
    Conf:
      {conf_name} : {conf_value}
      .
      .
      .
```
## Struct 
Struct는 Interface로 로드합니다. 
```go
  type Param struct {
    Param		[]Conf 				`yaml:"Param"`
  }

  type Conf struct {
    ConfId		string					`yaml:"ConfigId"`
    Conf 		map[string]interface{} 	`yaml:"Conf"`
  }
```

## Functions
| 함수 | 소스 | SDK |
|---|---|---|
| `FileLoader(path)` | YAML 파일 | - |
| `AWSParamLoader(region, key)` | SSM Parameter Store | aws-sdk-go v1 |
| `AWSSecretLoader(region, key)` | Secrets Manager (JSON) | aws-sdk-go v1 |
| `AWSParamLoaderV2(ctx, region, key, optFns...)` | SSM Parameter Store | aws-sdk-go-v2 |
| `AWSSecretLoaderV2(ctx, region, key, optFns...)` | Secrets Manager (JSON) | aws-sdk-go-v2 |

V1·V2 는 반환 타입이 같으므로 `Keyload`, `Conflist`, `InterfaceToSlice` 를 그대로 사용할 수 있습니다.
V1 함수는 기존 사용처 호환을 위해 유지합니다.

## AWS SDK v2
신규 코드와 EKS Pod Identity 환경에서는 V2 함수를 사용합니다.
구버전 v1 SDK 는 Pod Identity 엔드포인트(`169.254.170.23`)를 거부해 자격 증명을 얻지 못합니다.

```go
ctx := context.Background()

// Parameter Store
config, err := confloader.AWSParamLoaderV2(ctx, "ap-northeast-2", "/dba/batch/prod/sample")
if err != nil {
	return err
}
repository := config.Keyload("repository")

// Secrets Manager
secret, err := confloader.AWSSecretLoaderV2(ctx, "ap-northeast-2", "dba/sample")
if err != nil {
	return err
}
username := secret["username"].(string)
```

V1 에서 바꿀 때는 함수 이름에 `V2` 를 붙이고 첫 인자로 `context.Context` 를 넘기면 됩니다.
`optFns` 는 `config.LoadDefaultConfig` 에 그대로 전달되므로 필요할 때만 프로필 등을 지정합니다.

```go
import awsconfig "github.com/aws/aws-sdk-go-v2/config"

config, err := confloader.AWSParamLoaderV2(ctx, "ap-northeast-2", key,
	awsconfig.WithSharedConfigProfile("dev"))
```

V2 함수는 Go 1.22 이상이 필요합니다.