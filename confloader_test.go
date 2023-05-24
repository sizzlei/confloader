package confloader_test 


import (
	cnf "github.com/sizzlei/confloader"
	"fmt"
	"testing"
)

func TestFileLoad(t *testing.T) {
	x, err := cnf.FileLoader("./test_config.yml")
	if err != nil {
		panic(err)
	}

	configure := x.Conflist()
	y := x.Keyload(configure[0])
	
	fmt.Println(y)
}

func TestConvertSlice(t *testing.T) {
	x, err := cnf.FileLoader("./test_config.yml")
	if err != nil {
		panic(err)
	}

	configure := x.Conflist()
	y := x.Keyload(configure[0])
	z := cnf.InterfaceToSlice(y["DB"])

	fmt.Println(z)
}