package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"

	constant "github.com/e-harsley/go-mycore/constants"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/segmentio/fasthash/fnv1a"
)

type DataOperation struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	IsError bool        `json:"is_error"`
	ID      string      `json:"id"`
	Data    interface{} `json:"data"`
	Level   string      `json:"level"`
}

func GetEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}

func PanicException(err error, message string) {
	if err != nil {
		panic(message)
	}
}

func BaseException(err error, ctx *gin.Context) {
	ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error(), "error": err})
}

func RaiseException(ctx *gin.Context, err string) {
	ctx.JSON(http.StatusInternalServerError, gin.H{"message": err, "status": constant.ERROR_STATUS})
}

type Error struct {
	Errors map[string]interface{} `json:"errors"`
}

func NewError(err error) Error {
	e := Error{}
	e.Errors = make(map[string]interface{})
	switch v := err.(type) {

	default:
		e.Errors["body"] = v.Error()
	}
	return e
}

func NewValidatorError(err error) Error {
	e := Error{}
	e.Errors = make(map[string]interface{})
	errs := err.(validator.ValidationErrors)
	for _, v := range errs {
		e.Errors[v.Field()] = fmt.Sprintf("%v", v.Tag())
	}
	return e
}

func AccessForbidden() Error {
	e := Error{}
	e.Errors = make(map[string]interface{})
	e.Errors["body"] = "access forbidden"
	return e
}

func NotFound() Error {
	e := Error{}
	e.Errors = make(map[string]interface{})
	e.Errors["body"] = "resource not found"
	return e
}

func HashPassword(password string) (string, error) {
	if len(password) == 0 {
		return "", errors.New("password should not be empty")
	}
	return strconv.FormatUint(fnv1a.HashString64(password), 32), nil
}

func BoolAddr(b bool) *bool {
	boolVar := b
	return &boolVar
}

func StringToUint(s string) uint {
	u64, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		fmt.Println(err)
	}
	wd := uint(u64)
	return wd
}

func ToJsonMap(data interface{}) (response map[string]interface{}) {
	res := map[string]interface{}{}
	res["is_null"] = false
	fmt.Println("response", data)
	if data == nil {
		res["error"] = "Null interface provided"
		res["is_error"] = true
		res["is_empty"] = false
		res["is_null"] = true
		return res
	}
	jsonArray, _ := json.Marshal(data)

	err := json.Unmarshal([]byte(jsonArray), &res)
	fmt.Println(res)
	if err != nil {
		res["error"] = err.Error()
		res["is_empty"] = false
		res["is_error"] = true
		return res
	}

	if len(res) == 0 {
		res["is_empty"] = true
	}
	return res
}

func BindDataOperationStruct(data interface{}, output interface{}) (res DataOperation) {
	fmt.Println("i am", data)
	json_string, _ := json.Marshal(data)
	val := reflect.ValueOf(output)
	var err error
	if reflect.TypeOf(data).Kind() == reflect.Map {
		fmt.Println("dddddddd")
		if val.Kind() == reflect.Ptr {
			fmt.Println("sss")

			err = json.Unmarshal(json_string, &output)
		} else {
			fmt.Println("here dd")
			ptr := reflect.New(val.Type())
			ptr.Elem().Set(val)
			err = json.Unmarshal(json_string, ptr.Interface())
			fmt.Println(ptr.Elem())
			val = ptr.Elem()
			fmt.Println(val)

		}
	}

	if reflect.TypeOf(data).Kind() == reflect.Slice {
		dd := reflect.ValueOf(data)
		sliceType := reflect.SliceOf(val.Type())
		if sliceType.Kind() == reflect.Ptr {
			err = json.Unmarshal(json_string, &sliceType)
		} else {
			// fmt.Println(">>>>>>>>>>>>>>>", val.Type())
			// elemType := val.Type().Elem()
			fmt.Println("hhee", len(json_string))
			sliceType := reflect.SliceOf(val.Type())
			slice := reflect.MakeSlice(sliceType, dd.Len(), dd.Len())

			// Set the element of the slice to the output struct value
			slice.Index(0).Set(val)
			ptr := reflect.New(slice.Type())
			ptr.Elem().Set(slice)
			err = json.Unmarshal(json_string, ptr.Interface())
			// Iterate over the elements of the slice
			fmt.Println(slice, slice.Len())
			val = slice
			// for i := 0; i < slice.Len(); i++ {
			// 	val = slice.Index(i)
			// }
		}
	}

	fail := DataOperation{}

	if err != nil {

		fail.Success = false
		fail.Message = err.Error()
		fail.IsError = true

		return fail
	}
	if val.Kind() == reflect.Struct {
		fmt.Println("am i here", val)
		m := make(map[string]interface{})
		for i := 0; i < val.NumField(); i++ {
			field := val.Type().Field(i)
			fmt.Println(field.Name, val.Field(i).Interface())
			m[strings.ToLower(field.Name)] = val.Field(i).Interface()
		}

		fail.IsError = false
		fail.Data = m

		return fail
	}

	if val.Kind() == reflect.Slice {

		// Get the type of a zero value of the desired element type
		mapType := reflect.TypeOf(map[string]interface{}{})

		// Create a slice type with the element type obtained above
		sliceType := reflect.SliceOf(mapType)

		// Create a new inner slice of maps
		innerSlice1 := reflect.MakeSlice(sliceType, 0, 0)

		// Iterate over the elements of the input slice
		for i := 0; i < val.Len(); i++ {
			// Get the current element value
			va := val.Index(i)

			// Create a new map
			m := make(map[string]interface{})

			// Iterate over the fields of the struct value
			for j := 0; j < va.NumField(); j++ {
				// Get the current field
				field := va.Type().Field(j)

				// Add the field name and value to the map
				m[strings.ToLower(field.Name)] = va.Field(j).Interface()
			}

			// Append the map to the inner slice
			innerSlice1 = reflect.Append(innerSlice1, reflect.ValueOf(m))
		}
		// innerSlice2 := innerSlice1.Elem()

		// Convert the element value to an interface{} type
		innerSlice3 := innerSlice1.Interface().([]map[string]interface{})
		fail.IsError = false
		fail.Data = innerSlice3
		return fail
	}

	return fail
}

func GetMapField(i interface{}, fieldName string) (map[string]interface{}, bool) {
	val := reflect.ValueOf(i)
	if val.Kind() == reflect.Struct {
		fieldVal := val.FieldByName(fieldName)

		if !fieldVal.IsValid() {
			return map[string]interface{}{}, true
		}

		return fieldVal.Interface().(map[string]interface{}), false
	}

	return map[string]interface{}{}, true
}

func BindToMap(ctx *gin.Context, us interface{}) (map[string]interface{}, error) {

	val := reflect.ValueOf(us)
	if val.Kind() == reflect.Ptr {
		if err := ctx.ShouldBindJSON(us); err != nil {
			return nil, err
		}
	} else {
		ptr := reflect.New(val.Type())
		ptr.Elem().Set(val)
		if err := ctx.ShouldBindJSON(ptr.Interface()); err != nil {
			// Handle binding error

			RaiseException(ctx, err.Error())
			return nil, err
		}
		val = ptr.Elem()
	}
	m := make(map[string]interface{})
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		m[field.Name] = val.Field(i).Interface()
	}
	return m, nil

}
