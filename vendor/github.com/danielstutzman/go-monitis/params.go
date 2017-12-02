package monitis

import (
	"log"
	"net/url"
	"reflect"
	"strconv"
)

func String(s string) *string {
	return &s
}

func Int(i int) *int {
	return &i
}

func BoolToInt(b bool) *int {
	i := 0
	if b {
		i = 1
	}
	return &i
}

func optsToForm(opts interface{}) url.Values {
	form := url.Values{}
	optsReflect := reflect.ValueOf(opts).Elem()
	for i := 0; i < optsReflect.NumField(); i++ {
		optValueReflect := optsReflect.Field(i)
		optTypeReflect := optsReflect.Type().Field(i)

		paramName := optTypeReflect.Tag.Get("param")
		if paramName == "" {
			log.Fatalf("Field %s is missing 'param' tag", optTypeReflect.Name)
		}

		if !optValueReflect.IsNil() {
			typeString := optTypeReflect.Type.String()
			if typeString == "*string" {
				paramValue := optValueReflect.Elem().String()
				form.Add(paramName, paramValue)
			} else if typeString == "*int" {
				paramValue := optValueReflect.Elem().Int()
				form.Add(paramName, strconv.FormatInt(paramValue, 10))
			} else {
				log.Fatalf("Field %s has unexpected type %s",
					optTypeReflect.Name, typeString)
			}
		}
	}
	return form
}
