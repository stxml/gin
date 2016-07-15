// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin/binding/example"
	"github.com/golang/protobuf/proto"

	"github.com/stretchr/testify/assert"
)

type FooStruct struct {
	Foo string `json:"foo" form:"foo" xml:"foo" binding:"required"`
}

type FooBarStruct struct {
	FooStruct
	Bar string `json:"bar" form:"bar" xml:"bar" binding:"required"`
}

type StructWithPointers struct {
	Int     *int
	Int8    *int8
	Int16   *int16
	Int32   *int32
	Int64   *int64
	UInt    *uint
	UInt8   *uint8
	UInt16  *uint16
	UInt32  *uint32
	UInt64  *uint64
	Bool    *bool
	Float32 *float32
	Float64 *float64
	String  *string
}

func TestBindingDefault(t *testing.T) {
	assert.Equal(t, Default("GET", ""), Form)
	assert.Equal(t, Default("GET", MIMEJSON), Form)

	assert.Equal(t, Default("POST", MIMEJSON), JSON)
	assert.Equal(t, Default("PUT", MIMEJSON), JSON)

	assert.Equal(t, Default("POST", MIMEXML), XML)
	assert.Equal(t, Default("PUT", MIMEXML2), XML)

	assert.Equal(t, Default("POST", MIMEPOSTForm), Form)
	assert.Equal(t, Default("PUT", MIMEPOSTForm), Form)

	assert.Equal(t, Default("POST", MIMEMultipartPOSTForm), Form)
	assert.Equal(t, Default("PUT", MIMEMultipartPOSTForm), Form)

	assert.Equal(t, Default("POST", MIMEPROTOBUF), ProtoBuf)
	assert.Equal(t, Default("PUT", MIMEPROTOBUF), ProtoBuf)
}

func TestBindingJSON(t *testing.T) {
	testBodyBinding(t,
		JSON, "json",
		"/", "/",
		`{"foo": "bar"}`, `{"bar": "foo"}`)
}

func TestBindingForm(t *testing.T) {
	testFormBinding(t, "POST",
		"/", "/",
		"foo=bar&bar=foo", "bar2=foo")
}

func TestBindingForm2(t *testing.T) {
	testFormBinding(t, "GET",
		"/?foo=bar&bar=foo", "/?bar2=foo",
		"", "")
}

func TestBindingXML(t *testing.T) {
	testBodyBinding(t,
		XML, "xml",
		"/", "/",
		"<map><foo>bar</foo></map>", "<map><bar>foo</bar></map>")
}

func createFormPostRequest() *http.Request {
	req, _ := http.NewRequest("POST", "/?foo=getfoo&bar=getbar", bytes.NewBufferString("foo=bar&bar=foo"))
	req.Header.Set("Content-Type", MIMEPOSTForm)
	return req
}

func createFormMultipartRequest() *http.Request {
	boundary := "--testboundary"
	body := new(bytes.Buffer)
	mw := multipart.NewWriter(body)
	defer mw.Close()

	mw.SetBoundary(boundary)
	mw.WriteField("foo", "bar")
	mw.WriteField("bar", "foo")
	req, _ := http.NewRequest("POST", "/?foo=getfoo&bar=getbar", body)
	req.Header.Set("Content-Type", MIMEMultipartPOSTForm+"; boundary="+boundary)
	return req
}

func TestBindingFormPost(t *testing.T) {
	req := createFormPostRequest()
	var obj FooBarStruct
	FormPost.Bind(req, &obj)

	assert.Equal(t, obj.Foo, "bar")
	assert.Equal(t, obj.Bar, "foo")
}

func TestBindingFormMultipart(t *testing.T) {
	req := createFormMultipartRequest()
	var obj FooBarStruct
	FormMultipart.Bind(req, &obj)

	assert.Equal(t, obj.Foo, "bar")
	assert.Equal(t, obj.Bar, "foo")
}

func TestBindingProtoBuf(t *testing.T) {
	test := &example.Test{
		Label: proto.String("yes"),
	}
	data, _ := proto.Marshal(test)

	testProtoBodyBinding(t,
		ProtoBuf, "protobuf",
		"/", "/",
		string(data), string(data[1:]))
}

func TestValidationFails(t *testing.T) {
	var obj FooStruct
	req := requestWithBody("POST", "/", `{"bar": "foo"}`)
	err := JSON.Bind(req, &obj)
	assert.Error(t, err)
}

func TestValidationDisabled(t *testing.T) {
	backup := Validator
	Validator = nil
	defer func() { Validator = backup }()

	var obj FooStruct
	req := requestWithBody("POST", "/", `{"bar": "foo"}`)
	err := JSON.Bind(req, &obj)
	assert.NoError(t, err)
}

func TestExistsSucceeds(t *testing.T) {
	type HogeStruct struct {
		Hoge *int `json:"hoge" binding:"exists"`
	}

	var obj HogeStruct
	req := requestWithBody("POST", "/", `{"hoge": 0}`)
	err := JSON.Bind(req, &obj)
	assert.NoError(t, err)
}

func TestExistsFails(t *testing.T) {
	type HogeStruct struct {
		Hoge *int `json:"foo" binding:"exists"`
	}

	var obj HogeStruct
	req := requestWithBody("POST", "/", `{"boen": 0}`)
	err := JSON.Bind(req, &obj)
	assert.Error(t, err)
}

func testFormBinding(t *testing.T, method, path, badPath, body, badBody string) {
	b := Form
	assert.Equal(t, b.Name(), "form")

	obj := FooBarStruct{}
	req := requestWithBody(method, path, body)
	if method == "POST" {
		req.Header.Add("Content-Type", MIMEPOSTForm)
	}
	err := b.Bind(req, &obj)
	assert.NoError(t, err)
	assert.Equal(t, obj.Foo, "bar")
	assert.Equal(t, obj.Bar, "foo")

	obj = FooBarStruct{}
	req = requestWithBody(method, badPath, badBody)
	err = b.Bind(req, &obj)
	assert.Error(t, err)
}

func TestFormBindingWithPointers1(t *testing.T) {
	obj := StructWithPointers{}
	req := requestWithBody("GET", "/?Int=1&Int8=1&Int16=1&Int32=1&Int64=1&UInt=1&UInt8=1&UInt16=1&UInt32=1&UInt64=1&Bool=true&Float32=0.1&Float64=0.1&String=foo", "")

	err := Form.Bind(req, &obj)
	assert.NoError(t, err)
	assert.EqualValues(t, *obj.Int, 1)
	assert.EqualValues(t, *obj.Int8, 1)
	assert.EqualValues(t, *obj.Int16, 1)
	assert.EqualValues(t, *obj.Int32, 1)
	assert.EqualValues(t, *obj.Int64, 1)
	assert.EqualValues(t, *obj.UInt, 1)
	assert.EqualValues(t, *obj.UInt8, 1)
	assert.EqualValues(t, *obj.UInt16, 1)
	assert.EqualValues(t, *obj.UInt32, 1)
	assert.EqualValues(t, *obj.UInt64, 1)
	assert.EqualValues(t, *obj.Bool, true)
	assert.EqualValues(t, *obj.Float32, float32(0.1))
	assert.EqualValues(t, *obj.Float64, 0.1)
	assert.EqualValues(t, *obj.String, "foo")
}

func TestFormBindingWithPointers2(t *testing.T) {
	obj := StructWithPointers{}
	req := requestWithBody("POST", "/", "Int=1&Int8=1&Int16=1&Int32=1&Int64=1&UInt=1&UInt8=1&UInt16=1&UInt32=1&UInt64=1&Bool=true&Float32=0.1&Float64=0.1&String=foo")
	req.Header.Add("Content-Type", MIMEPOSTForm)

	err := Form.Bind(req, &obj)
	assert.NoError(t, err)
	assert.EqualValues(t, *obj.Int, 1)
	assert.EqualValues(t, *obj.Int8, 1)
	assert.EqualValues(t, *obj.Int16, 1)
	assert.EqualValues(t, *obj.Int32, 1)
	assert.EqualValues(t, *obj.Int64, 1)
	assert.EqualValues(t, *obj.UInt, 1)
	assert.EqualValues(t, *obj.UInt8, 1)
	assert.EqualValues(t, *obj.UInt16, 1)
	assert.EqualValues(t, *obj.UInt32, 1)
	assert.EqualValues(t, *obj.UInt64, 1)
	assert.EqualValues(t, *obj.Bool, true)
	assert.EqualValues(t, *obj.Float32, float32(0.1))
	assert.EqualValues(t, *obj.Float64, 0.1)
	assert.EqualValues(t, *obj.String, "foo")
}

func testBodyBinding(t *testing.T, b Binding, name, path, badPath, body, badBody string) {
	assert.Equal(t, b.Name(), name)

	obj := FooStruct{}
	req := requestWithBody("POST", path, body)
	err := b.Bind(req, &obj)
	assert.NoError(t, err)
	assert.Equal(t, obj.Foo, "bar")

	obj = FooStruct{}
	req = requestWithBody("POST", badPath, badBody)
	err = JSON.Bind(req, &obj)
	assert.Error(t, err)
}

func testProtoBodyBinding(t *testing.T, b Binding, name, path, badPath, body, badBody string) {
	assert.Equal(t, b.Name(), name)

	obj := example.Test{}
	req := requestWithBody("POST", path, body)
	req.Header.Add("Content-Type", MIMEPROTOBUF)
	err := b.Bind(req, &obj)
	assert.NoError(t, err)
	assert.Equal(t, *obj.Label, "yes")

	obj = example.Test{}
	req = requestWithBody("POST", badPath, badBody)
	req.Header.Add("Content-Type", MIMEPROTOBUF)
	err = ProtoBuf.Bind(req, &obj)
	assert.Error(t, err)
}

func requestWithBody(method, path, body string) (req *http.Request) {
	req, _ = http.NewRequest(method, path, bytes.NewBufferString(body))
	return
}
