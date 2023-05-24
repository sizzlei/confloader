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