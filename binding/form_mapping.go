// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding

import (
	"errors"
	"reflect"
	"strconv"
)

func mapForm(ptr interface{}, form map[string][]string) error {
	typ := reflect.TypeOf(ptr).Elem()
	val := reflect.ValueOf(ptr).Elem()
	for i := 0; i < typ.NumField(); i++ {
		typeField := typ.Field(i)
		structField := val.Field(i)
		if !structField.CanSet() {
			continue
		}

		structFieldKind := structField.Kind()
		inputFieldName := typeField.Tag.Get("form")
		if inputFieldName == "" {
			inputFieldName = typeField.Name

			// if "form" tag is nil, we inspect if the field is a struct.
			// this would not make sense for JSON parsing but it does for a form
			// since data is flatten
			if structFieldKind == reflect.Struct {
				err := mapForm(structField.Addr().Interface(), form)
				if err != nil {
					return err
				}
				continue
			}
		}
		inputValue, exists := form[inputFieldName]
		if !exists {
			continue
		}

		numElems := len(inputValue)
		if structFieldKind == reflect.Slice && numElems > 0 {
			sliceOf := structField.Type().Elem().Kind()
			slice := reflect.MakeSlice(structField.Type(), numElems, numElems)
			for i := 0; i < numElems; i++ {
				if err := setWithProperType(sliceOf, inputValue[i], slice.Index(i)); err != nil {
					return err
				}
			}
			val.Field(i).Set(slice)
		} else {
			if typeField.Type.Kind() == reflect.Ptr {
				if err := setWithPointerType(typeField.Type.Elem().Kind(), inputValue[0], structField); err != nil {
					return err
				}
			} else {
				if err := setWithProperType(typeField.Type.Kind(), inputValue[0], structField); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func setWithPointerType(valueKind reflect.Kind, val string, structField reflect.Value) error {
	var (
		int64Val   int64
		uint64Val  uint64
		boolVal    bool
		float64Val float64
		err        error
	)
	switch valueKind {
	case reflect.Int:
		if int64Val, err = parseInt(val, 0); err == nil {
			intVal := int(int64Val)
			structField.Set(reflect.ValueOf(&intVal))
		}
	case reflect.Int8:
		if int64Val, err = parseInt(val, 8); err == nil {
			int8Val := int8(int64Val)
			structField.Set(reflect.ValueOf(&int8Val))
		}
	case reflect.Int16:
		if int64Val, err = parseInt(val, 16); err == nil {
			int16Val := int16(int64Val)
			structField.Set(reflect.ValueOf(&int16Val))
		}
	case reflect.Int32:
		if int64Val, err = parseInt(val, 32); err == nil {
			int32Val := int32(int64Val)
			structField.Set(reflect.ValueOf(&int32Val))
		}
	case reflect.Int64:
		if int64Val, err = parseInt(val, 64); err == nil {
			structField.Set(reflect.ValueOf(&int64Val))
		}
	case reflect.Uint:
		if uint64Val, err = parseUint(val, 0); err == nil {
			uintVal := uint(uint64Val)
			structField.Set(reflect.ValueOf(&uintVal))
		}
	case reflect.Uint8:
		if uint64Val, err = parseUint(val, 8); err == nil {
			uint8Val := uint8(uint64Val)
			structField.Set(reflect.ValueOf(&uint8Val))
		}
	case reflect.Uint16:
		if uint64Val, err = parseUint(val, 16); err == nil {
			uint16Val := uint16(uint64Val)
			structField.Set(reflect.ValueOf(&uint16Val))
		}
	case reflect.Uint32:
		if uint64Val, err = parseUint(val, 32); err == nil {
			uint32Val := uint32(uint64Val)
			structField.Set(reflect.ValueOf(&uint32Val))
		}
	case reflect.Uint64:
		if uint64Val, err = parseUint(val, 64); err == nil {
			structField.Set(reflect.ValueOf(&uint64Val))
		}
	case reflect.Bool:
		if boolVal, err = parseBool(val); err == nil {
			structField.Set(reflect.ValueOf(&boolVal))
		}
	case reflect.Float32:
		if float64Val, err = parseFloat(val, 32); err == nil {
			float32Val := float32(float64Val)
			structField.Set(reflect.ValueOf(&float32Val))
		}
	case reflect.Float64:
		if float64Val, err = parseFloat(val, 64); err == nil {
			structField.Set(reflect.ValueOf(&float64Val))
		}
	case reflect.String:
		structField.Set(reflect.ValueOf(&val))
	default:
		err = errors.New("Unknown Pointer type: *" + valueKind.String())
	}
	return err
}

func setWithProperType(valueKind reflect.Kind, val string, structField reflect.Value) error {
	switch valueKind {
	case reflect.Int:
		return setIntField(val, 0, structField)
	case reflect.Int8:
		return setIntField(val, 8, structField)
	case reflect.Int16:
		return setIntField(val, 16, structField)
	case reflect.Int32:
		return setIntField(val, 32, structField)
	case reflect.Int64:
		return setIntField(val, 64, structField)
	case reflect.Uint:
		return setUintField(val, 0, structField)
	case reflect.Uint8:
		return setUintField(val, 8, structField)
	case reflect.Uint16:
		return setUintField(val, 16, structField)
	case reflect.Uint32:
		return setUintField(val, 32, structField)
	case reflect.Uint64:
		return setUintField(val, 64, structField)
	case reflect.Bool:
		return setBoolField(val, structField)
	case reflect.Float32:
		return setFloatField(val, 32, structField)
	case reflect.Float64:
		return setFloatField(val, 64, structField)
	case reflect.String:
		structField.SetString(val)
	default:
		return errors.New("Unknown type")
	}
	return nil
}

func parseInt(val string, bitSize int) (int64, error) {
	if val == "" {
		val = "0"
	}
	return strconv.ParseInt(val, 10, bitSize)
}

func setIntField(val string, bitSize int, field reflect.Value) error {
	intVal, err := parseInt(val, bitSize)
	if err == nil {
		field.SetInt(intVal)
	}
	return err
}

func parseUint(val string, bitSize int) (uint64, error) {
	if val == "" {
		val = "0"
	}
	return strconv.ParseUint(val, 10, bitSize)
}

func setUintField(val string, bitSize int, field reflect.Value) error {
	uintVal, err := parseUint(val, bitSize)
	if err == nil {
		field.SetUint(uintVal)
	}
	return err
}

func parseBool(val string) (bool, error) {
	if val == "" {
		val = "false"
	}
	return strconv.ParseBool(val)
}

func setBoolField(val string, field reflect.Value) error {
	boolVal, err := parseBool(val)
	if err == nil {
		field.SetBool(boolVal)
	}
	return nil
}

func parseFloat(val string, bitSize int) (float64, error) {
	if val == "" {
		val = "0.0"
	}
	return strconv.ParseFloat(val, bitSize)
}

func setFloatField(val string, bitSize int, field reflect.Value) error {
	floatVal, err := parseFloat(val, bitSize)
	if err == nil {
		field.SetFloat(floatVal)
	}
	return err
}

// Don't pass in pointers to bind to. Can lead to bugs. See:
// https://github.com/codegangsta/martini-contrib/issues/40
// https://github.com/codegangsta/martini-contrib/pull/34#issuecomment-29683659
func ensureNotPointer(obj interface{}) {
	if reflect.TypeOf(obj).Kind() == reflect.Ptr {
		panic("Pointers are not accepted as binding models")
	}
}
