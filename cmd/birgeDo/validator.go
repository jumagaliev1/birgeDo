package main

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

func (app *application) IsValid(object interface{}) (isValid bool, mp map[string][]string) {
	isValid = true
	values := reflect.ValueOf(object)
	types := reflect.TypeOf(object)
	for i := 0; i < types.NumField(); i++ {
		field := types.Field(i)
		value := values.Field(i).String()
		if valid, hints := app.callMethodsByStructTag(field, value); !valid {
			isValid = valid
			mp[jsonName(field)] = hints
		}
	}
	return
}

func (app *application) callMethodsByStructTag(
	field reflect.StructField, value string,
) (isValid bool, hints []string) {
	isValid = true
	if _, ok := field.Tag.Lookup("consists"); ok {
		valid, consists := checkConsistsTag(value, field.Tag.Get("consists"))
		isValid = valid || isValid
		hints = append(hints, consists...)
	}
	if _, ok := field.Tag.Lookup("existsid"); ok {
		valid := app.IDExistsTag(value, field.Tag.Get("existsid"))
		isValid = valid || isValid
		hints = append(hints, "doesn't exist with this id")
	}
	if _, ok := field.Tag.Lookup("maxlength"); ok {
		valid, consists := maxTag(value, field.Tag.Get("maxlength"))
		isValid = valid || isValid
		hints = append(hints, consists)
	}
	if _, ok := field.Tag.Lookup("minlength"); ok {
		valid, consists := minTag(value, field.Tag.Get("minlength"))
		isValid = valid || isValid
		hints = append(hints, consists)
	}
	return
}

func checkConsistsTag(value, tag string) (isValid bool, hints []string) {
	isValid = true
	strs := strings.Split(tag, ",")
	for _, str := range strs {
		if str == "email" && !isEmail(value) {
			hints = append(hints, "must be a valid email address")
			isValid = false
		} else if str == "digit" && !hasDigit(value) {
			hints = append(hints, "must contain digit")
			isValid = false
		} else if str == "lowercase" && !hasLowerCase(value) {
			hints = append(hints, "must contain lowercase character")
		} else if str == "uppercase" && !hasUpperCase(value) {
			hints = append(hints, "must contain uppercase character")
		} else if str == "symbol" && !hasSymbol(value) {
			hints = append(hints, "must contain symbol")
		}
	}
	return
}

func (app *application) IDExistsTag(value, table string) (isValid bool) {
	id, _ := strconv.Atoi(value)
	var err error
	if table == "rooms" {
		_, err = app.models.Room.GetByID(int64(id))
	} else if table == "users" {
		_, err = app.models.Users.Get(id)
	} else if table == "tasks" {
		_, err = app.models.Task.GetByID(int64(id))
	} else if table == "tokens" {
		_, err = app.models.Task.GetByID(int64(id))
	}
	return err == nil
}

func maxTag(value, tag string) (bool, string) {
	length, _ := strconv.Atoi(tag)
	return len(value) < length, "length should be less than " + value
}

func minTag(value, tag string) (isValid bool, hint string) {
	length, _ := strconv.Atoi(tag)
	return len(value) > length, "length should be more than " + value
}

func hasUpperCase(plaintext string) bool {
	for _, ch := range plaintext {
		if unicode.IsUpper(ch) {
			return true
		}
	}
	return false
}

func hasLowerCase(plaintext string) bool {
	for _, ch := range plaintext {
		if unicode.IsLower(ch) {
			return true
		}
	}
	return false
}

func hasDigit(plaintext string) bool {
	for _, ch := range plaintext {
		if unicode.IsDigit(ch) {
			return true
		}
	}
	return false
}

func hasSymbol(plaintext string) bool {
	for _, ch := range plaintext {
		if unicode.IsSymbol(ch) {
			return true
		}
	}
	return false
}

func isEmail(plaintext string) bool {
	EmailRX := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	return EmailRX.Match([]byte(plaintext))
}

func jsonName(field reflect.StructField) string {
	ans, _ := field.Tag.Lookup("json")
	return ans
}
