package main

import "reflect"

func GetType(val interface{}) (res string) {
	typeof := reflect.TypeOf(val)
	for typeof.Kind() == reflect.Ptr {
		typeof = typeof.Elem()
		res += "*"
	}
	return res + typeof.Name()
}
