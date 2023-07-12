package utils

import (
	"fmt"
	"reflect"
)

// ==== UTILS
func PrintStruct(s interface{}) {
	utilsPrintStruct(s, "")
}

// iterate over tx and print all fields
func utilsPrintStruct(s interface{}, indent string) {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		fmt.Println("printStruct only accepts structs; got", v.Kind())
		return
	}

	for i := 0; i < v.NumField(); i++ {
		fmt.Printf("\n%s%s\n", indent, v.Type().Field(i).Name)
		field := v.Field(i)

		// If this is a nested struct, call the function recursively
		if field.Kind() == reflect.Struct {
			utilsPrintStruct(field.Interface(), indent+"  ")
		} else if field.Kind() == reflect.Ptr && field.Elem().Kind() == reflect.Struct {
			utilsPrintStruct(field.Interface(), indent+"  ")
		} else {
			// This is not a nested struct, so just print the value
			fmt.Println(field.Interface())
		}
	}
}
