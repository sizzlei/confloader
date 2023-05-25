package confloader

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"encoding/json"
)

type Param struct {
	Param		[]Conf 				`yaml:"Param"`
}

type Conf struct {
	ConfId		string					`yaml:"ConfigId"`
	Conf 		map[string]interface{} 	`yaml:"Conf"`
}

func FileLoader(p string) (Param, error) {
	// p : FilePath
	var c Param
	yamlFile, _ := ioutil.ReadFile(p)
	err := yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		return c, err
	}

	return c, nil
}

func AWSParamLoader(r string, k string)  (Param, error) {
	// r : Region
	// p : SSM.Parameter Store Key Name
	var c Param
	gs := session.Must(session.NewSession())
	s := ssm.New(gs, aws.NewConfig().WithRegion(r))

	p, err := s.GetParameter(&ssm.GetParameterInput{
		Name: aws.String(k),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return c, err
	}

	cs := *p.Parameter.Value
	err = yaml.Unmarshal([]byte(cs), &c)
	if err != nil {
		return c, err
	}

	return c, nil
}

func (p Param) Keyload(k string) map[string]interface{} {
	var c map[string]interface{}
	for i, v := range p.Param {
		if v.ConfId == k {
			c = make(map[string]interface{})
			c = p.Param[i].Conf
		}
	}
	return c
}

func (p Param) Conflist() []string {
	var cl []string 
	for _, v := range p.Param {
		cl = append(cl,v.ConfId)
	}

	return cl
}

func InterfaceToSlice(i interface{}) []string {
	var returnSlice []string
	for _, v := range i.([]interface{}) {
		returnSlice = append(returnSlice,v.(string))
	}

	return returnSlice
}

func AWSSecretLoader(r string,k string) (map[string]interface{}, error) {
	// Declare Return Data
	data := make(map[string]interface{})

	// Create Session
	gs := session.Must(session.NewSession())
	secs := secretsmanager.New(gs, aws.NewConfig().WithRegion(r))

	// Get Keyvalue
	secv, err := secs.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: aws.String(k),
	})
	if err != nil {
		return nil, err
	}

	// Unmarshal
	err = json.Unmarshal([]byte(*secv.SecretString),&data)
	if err != nil {
		return nil, err
	}

	return data, nil
}