package easyworker

import (
	"fmt"
	"log"
	"reflect"
)

/*
call user's function througth reflect.
*/
func invokeFun(fun any, args ...any) (ret []any, err error) {
	// catch if panic by user code.
	defer func() {
		if r := recover(); r != nil {
			log.Println("user function was panic, ", r)
			err = fmt.Errorf("user function was panic, %s", r)
		}
	}()

	//log.Println("list args: ", args)

	fn := reflect.ValueOf(fun)
	fnType := fn.Type()
	numIn := fnType.NumIn()
	if numIn > len(args) {
		return nil, fmt.Errorf("function must have minimum %d params. Have %d", numIn, len(args))
	}
	if numIn != len(args) && !fnType.IsVariadic() {
		return nil, fmt.Errorf("func must have %d params. Have %d", numIn, len(args))
	}
	params := make([]reflect.Value, len(args))
	for i := 0; i < len(args); i++ {
		var inType reflect.Type
		if fnType.IsVariadic() && i >= numIn-1 {
			inType = fnType.In(numIn - 1).Elem()
		} else {
			inType = fnType.In(i)
		}
		argValue := reflect.ValueOf(args[i])
		if !argValue.IsValid() {
			return nil, fmt.Errorf("func Param[%d] must be %s. Have %s", i, inType, argValue.String())
		}
		argType := argValue.Type()
		if argType.ConvertibleTo(inType) {
			params[i] = argValue.Convert(inType)
		} else {
			return nil, fmt.Errorf("method Param[%d] must be %s. Have %s", i, inType, argType)
		}
	}

	result := fn.Call(params)

	ret = make([]any, len(result))

	for i, r := range result {
		ret[i] = r.Interface()
	}

	//log.Println("invoke result:", result)

	return ret, nil
}

/*
verify if interface is a function.
if interface is not a function, it will return an error.
*/
func verifyFunc(fun any) error {
	if v := reflect.ValueOf(fun); v.Kind() != reflect.Func {
		return fmt.Errorf("not a function")
	}
	return nil
}
