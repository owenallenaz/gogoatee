package goatee

import (
	"fmt"
	"testing"
)

var _ = fmt.Println

type testStruct1 struct {
	Test       string
	TestInt    int
	TestStruct testStruct1inner
}

func (t *testStruct1) TestFuncPointerString() string {
	return t.Test + "FuncPointerString"
}

func (t testStruct1) TestFuncValueString() string {
	return t.Test + "FuncValueString"
}

type testStruct1inner struct {
	Inner string
}

func TestValueStructString(t *testing.T) {
	template := []byte("{{Test}}")
	result, _ := Fill(Args{Template: template, Data: testStruct1{Test: "success"}})
	assert(t, "success", result, "")
}

func TestGlobalStructString(t *testing.T) {
	args := Args{
		Template: []byte("{{*Test}}"),
		Data:     testStruct1{Test: "fail"},
		Globals:  map[string]string{"Test": "success"},
	}
	result, _ := Fill(args)
	assert(t, "success", result, "")
}

func TestValueStructInt(t *testing.T) {
	data := testStruct1{TestInt: 10}
	template := []byte("{{TestInt}}")
	result, _ := Fill(Args{Template: template, Data: data})
	assert(t, "10", result, "")
}

func TestValueStructFuncPointerString(t *testing.T) {
	args := Args{
		Template: []byte("{{TestFuncPointerString}}{{TestFuncValueString}}"),
		Data:     &testStruct1{Test: "success"},
	}
	result, _ := Fill(args)
	assert(t, "successFuncPointerStringsuccessFuncValueString", result, "")
}

func TestValueStructFuncValueString(t *testing.T) {
	args := Args{
		Template: []byte("{{TestFuncPointerString}}{{TestFuncValueString}}"),
		Data:     testStruct1{Test: "success"},
	}
	result, _ := Fill(args)
	assert(t, "successFuncPointerStringsuccessFuncValueString", result, "")
}

func TestConditionalStructString(t *testing.T) {
	template := []byte("{{:Test}}success{{/Test}}")
	result, _ := Fill(Args{Template: template, Data: testStruct1{Test: "success"}})
	assert(t, "success", result, "")
}

func TestConditionalStructNotExists(t *testing.T) {
	template := []byte("{{:Test}}success{{/Test}}")
	result, _ := Fill(Args{Template: template, Data: testStruct1{}})
	assert(t, "", result, "")
}

func TestConditionalMapNil(t *testing.T) {
	template := []byte("{{:Test}}fail{{/Test}}")
	result, _ := Fill(Args{Template: template, Data: map[string]interface{}{"Test": nil}})
	assert(t, "", result, "")
}

func TestConditionalMapTrue(t *testing.T) {
	template := []byte("{{:Test}}success{{/Test}}")
	result, _ := Fill(Args{Template: template, Data: map[string]interface{}{"Test": true}})
	assert(t, "success", result, "")
}

func TestConditionalMapFalse(t *testing.T) {
	template := []byte("{{:Test}}fail{{/Test}}")
	result, _ := Fill(Args{Template: template, Data: map[string]interface{}{"Test": false}})
	assert(t, "", result, "")
}

func TestSectionStructStruct(t *testing.T) {
	template := []byte("{{#TestStruct}}{{Inner}}{{/TestStruct}}")
	result, _ := Fill(Args{Template: template, Data: testStruct1{TestStruct: testStruct1inner{Inner: "success"}}})
	assert(t, "success", result, "")
}

func TestSectionStructNotExists(t *testing.T) {
	template := []byte("{{#TestStruct}}{{Inner}}{{/TestStruct}}")
	result, _ := Fill(Args{Template: template, Data: testStruct1{Test: "fail"}})
	assert(t, "", result, "")
}

func TestSectionMapString(t *testing.T) {
	args := Args{
		Template: []byte("{{Test}}"),
		Data:     map[string]string{"Test": "success"},
	}
	result, _ := Fill(args)
	assert(t, "success", result, "")
}

func TestSectionMapSlice(t *testing.T) {
	data := map[string]interface{}{"TestStruct": []testStruct1{{Test: "success1"}, {Test: "success2"}}}
	template := []byte("{{#TestStruct}}{{Test}}{{/TestStruct}}")
	result, _ := Fill(Args{Template: template, Data: data})
	assert(t, "success1success2", result, "")
}

func TestSectionMapNotExists(t *testing.T) {
	data := make(map[string]interface{})
	template := []byte("{{#TestStruct}}fail{{/TestStruct}}")
	result, _ := Fill(Args{Template: template, Data: data})
	assert(t, "", result, "")
}

func TestSectionMapMap(t *testing.T) {
	data := map[string]interface{}{"TestMap": map[string]string{"Test": "success1", "Test2": "success2"}}
	template := []byte("{{#TestMap}}{{Test}}{{Test2}}{{/TestMap}}")
	result, _ := Fill(Args{Template: template, Data: data})
	assert(t, "success1success2", result, "")
}

func TestNegativeStructString(t *testing.T) {
	data := testStruct1{Test: "success"}
	template := []byte("{{!Test}}fail{{/Test}}")
	result, _ := Fill(Args{Template: template, Data: data})
	assert(t, "", result, "")
}

func TestNegativeStructStringFalsy(t *testing.T) {
	data := testStruct1{Test: ""}
	template := []byte("{{!Test}}success{{/Test}}")
	result, _ := Fill(Args{Template: template, Data: data})
	assert(t, "success", result, "")
}

func TestNegativeStructNotExists(t *testing.T) {
	data := testStruct1{}
	template := []byte("{{!Test}}success{{/Test}}")
	result, _ := Fill(Args{Template: template, Data: data})
	assert(t, "success", result, "")
}

func assert(t *testing.T, expected string, result []byte, message string) {
	if string(result) != expected {
		t.Fatalf("Expected: %+v Result: %+v Message: %+v", expected, string(result), message)
	}
}

//TemplateContainerKeyExists
