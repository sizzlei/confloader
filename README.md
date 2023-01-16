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
  - Config_id : {Configure Name}
    Conf:
      {conf_name} : {conf_value}
      {conf_name} : {conf_value}
      .
      .
      .

  - Config_id : {Configure Name}
    Conf:
      {conf_name} : {conf_value}
      .
      .
      .
```

## Function
현재 2가지 함수를 지원합니다. 
### FileLoader
```
func FileLoader(p string) (Param, error)
```
- p : File Path (ex. ../etc/conf/fileconf.yaml)

### AWSParamLoader
```
func AWSParamLoader(r string, k string)  (Param, error)
```
- r : region code (ex. ap-northeast-2   )
- k : aws parameter store Key name