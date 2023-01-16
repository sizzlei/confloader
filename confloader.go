package confloader

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

type Param struct {
	Param		[]Conf 				`yaml:"Param"`
}

type Conf struct {
	ConfId		string				`yaml:"Config_id"`
	Conf 		map[string]string 	`yaml:"Conf"`
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

func (p Param) Keyload(k string) map[string]string {
	var c map[string]string
	for i, v := range p.Param {
		if v.ConfId == k {
			c = make(map[string]string)
			c = p.Param[i].Conf
		}
	}
	return c

}