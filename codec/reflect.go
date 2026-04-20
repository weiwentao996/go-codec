package codec

import (
	"errors"
	"reflect"
)

func getReflectAndInitObj(v interface{}) (reflect.Value, error) {
	var vv reflect.Value
	if value, ok := v.(reflect.Value); ok {
		vv = value
	} else {
		vv = reflect.ValueOf(v)
	}

	for vv.Kind() == reflect.Ptr {
		if vv.IsNil() && vv.CanSet() {
			vv.Set(reflect.New(vv.Type().Elem()))
		}
		vv = vv.Elem()
	}

	for vv.Kind() == reflect.Interface {
		vv = vv.Elem()
		vv, _ = getReflectAndInitObj(vv)
	}

	if !vv.IsValid() {
		return vv, errors.New("reflect struct err, decode fail")
	}

	return vv, nil
}
