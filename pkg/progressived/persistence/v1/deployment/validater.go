package deployment

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"reflect"
	"strconv"
	"strings"
)

var v *validator.Validate

func isOneOfIf(fl validator.FieldLevel) bool {
	params := strings.Split(fl.Param(), " ")
	if len(params) < 3 {
		panic(fmt.Sprintf("Bad param number for oneof_if %s", fl.FieldName()))
	}

	field1 := fl.Parent().FieldByName(params[0])
	switch field1.Kind() {
	case reflect.Invalid:
		return true

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, _ := strconv.ParseInt(params[2], 0, 64)
		if field1.Int() != i {
			return true
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		i, _ := strconv.ParseUint(params[2], 0, 64)
		if field1.Uint() != i {
			return true
		}

	case reflect.Float32, reflect.Float64:
		f, _ := strconv.ParseFloat(params[2], 0)
		if field1.Float() != f {
			return true
		}

	default:
		if field1.String() != params[2] {
			return true
		}
	}

	vals := params[2:]
	field2 := fl.Field()
	var v string
	switch field2.Kind() {
	case reflect.String:
		v = field2.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v = strconv.FormatInt(field2.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v = strconv.FormatUint(field2.Uint(), 10)
	default:
		panic(fmt.Sprintf("Bad field type %T", field2.Interface()))
	}
	for i := 0; i < len(vals); i++ {
		if vals[i] == v {
			return true
		}
	}
	return false
}

func Validate(entity *Deployment) error {
	if v == nil {
		v = validator.New()
	}
	v.RegisterValidation("oneof_if", isOneOfIf)
	if err := v.Struct(entity); err != nil {
		return err
	}

	return nil
}
