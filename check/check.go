package check

import (
	"fmt"
	"reflect"
	"strings"
)

var checks map[string]func(any) error

func init() {
	checks = map[string]func(any) error{}
	checks["required"] = required
}

type Checker interface {
	Check() error
}

func Check(in any) error {
	switch ty := in.(type) {
	case Checker:
		return ty.Check()
	}

	return StructCheck(in)
}

func StructCheck(in any) error {
	if reflect.TypeOf(in).Kind() != reflect.Struct {
		return fmt.Errorf("must pass a struct to StructCheck")
	}

	fields := reflect.TypeOf(in).NumField()
	for i := 0; i < fields; i++ {
		field := reflect.TypeOf(in).Field(i)

		ruleList, ok := field.Tag.Lookup("check")
		if !ok {
			continue
		}

		rules := strings.Split(ruleList, ",")
		for _, rule := range rules {
			f, ok := checks[rule]
			if !ok {
				continue
			}

			v := reflect.ValueOf(in).FieldByName(field.Name)

			err := f(v)
			if err != nil {
				return err
			}
		}

	}

	return nil
}

func required[T any](in T) error {
	// if in.(reflect.Value).IsNil() {
	// 	return fmt.Errorf("field not set")
	// }

	switch ty := in.(type) {
	case reflect.Value:
		ty.String()
	}

	if in.(reflect.Value).IsZero() {
		return fmt.Errorf("field not set")
	}

	return nil
}
