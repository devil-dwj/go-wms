package engine_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/devil-dwj/go-wms/examples/pb"
)

func TestRefField(t *testing.T) {
	v := &pb.LoginRsp{}
	refField(v)
}

func refField(v interface{}) error {
	refV := reflect.ValueOf(v).Elem()

	for i := 0; i < refV.NumField(); i++ {
		fieldInfo := refV.Type().Field(i)
		tag := fieldInfo.Tag
		name := tag.Get("json")
		name = strings.Split(name, ",")[0]
		if name == "" {
			continue
		}

		fieldType := fieldInfo.Type.Name()
		if fieldType == "int32" {
			// param := c.Query(name)
			param := int32(1)
			refV.FieldByName(fieldInfo.Name).Set(reflect.ValueOf(param))
		} else {
			param := "1"
			refV.FieldByName(fieldInfo.Name).Set(reflect.ValueOf(param))
		}
	}

	return nil
}
