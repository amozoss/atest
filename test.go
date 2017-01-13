package atest

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

type Test struct {
	*testing.T
	dir  string
	skip int
}

func Wrap(t *testing.T, skip int) *Test {
	test := &Test{
		T:    t,
		skip: skip,
	}
	return test
}

func (t *Test) AssertNotEqual(val1, val2 interface{}) {
	if !reflect.DeepEqual(val1, val2) {
		_, file, line, _ := runtime.Caller(t.skip)
		t.Logf("** %s:%d **", path.Base(file), line)
		t.Logf("want: %v", val2)
		t.Logf("got:  %v", val1)
		t.FailNow()
	}
}

func (t *Test) AssertEqual(val1, val2 interface{}) {
	if !reflect.DeepEqual(val1, val2) {
		t.printTraceAndFail(val2, val1)
	}
}

func (t *Test) printTraceAndFail(val1, val2 interface{}) {
	_, file, line, _ := runtime.Caller(t.skip)
	t.Logf("** %s:%d **", path.Base(file), line)
	_, file, line, _ = runtime.Caller(t.skip + 1)
	t.Logf("** %s:%d **", path.Base(file), line)
	t.Logf("want: %v", val2)
	t.Logf(" got: %v", val1)
	t.FailNow()
}

func (t *Test) AssertError(err error) {
	t.AssertNotEqual(err, nil)
}

func (t *Test) AssertNoError(err error) {
	t.AssertEqual(err, nil)
}

func (t *Test) AssertNil(obj interface{}) {
	v1 := reflect.ValueOf(obj)
	if !v1.IsNil() {
		t.printTraceAndFail(obj, nil)
	}
}

func (t *Test) Assert(cond bool) {
	t.AssertEqual(cond, true)
}

func (t *Test) AssertJSONEqual(got_str, want_str string) {
	var w interface{}
	err := json.Unmarshal([]byte(got_str), &w)
	t.AssertNoError(err)
	got, err := json.Marshal(w)
	t.AssertNoError(err)

	var v interface{}
	err = json.Unmarshal([]byte(want_str), &v)
	t.AssertNoError(err)
	want, err := json.Marshal(v)
	t.AssertNoError(err)

	if !bytes.Equal(got, want) {
		_, file, line, _ := runtime.Caller(t.skip)
		t.Logf("** %s:%d **", path.Base(file), line)
		t.Logf("want: %s", want)
		t.Logf("got:  %s", got)
		t.Fail()
	}
}

func TestWithDir(t *testing.T, dir string, skip int) *Test {
	test := Wrap(t, skip)
	dir, err := ioutil.TempDir("", dir)
	test.AssertNoError(err)
	test.dir = dir
	return test
}

func (t *Test) Dir() string {
	return t.dir
}

func (t *Test) CreateFile(name string) *os.File {
	f, err := os.OpenFile(filepath.Join(t.dir, name),
		os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	t.AssertNoError(err)
	return f
}

func (t *Test) Close() {
	err := os.RemoveAll(t.dir)
	t.AssertNoError(err)
}

type Response struct {
	Resp *httptest.ResponseRecorder
	Json map[string]interface{}
	Code int
	Body string
}

func (t *Test) PerformRequest(server http.Handler, method, endpoint string,
	headers http.Header, json_str string) *Response {
	req, err := http.NewRequest(method, endpoint, bytes.NewBuffer([]byte(json_str)))
	ctx := context.TODO()
	req.WithContext(ctx)
	if headers != nil {
		req.Header = headers
	}
	t.AssertNoError(err)
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)

	var params map[string]interface{}
	if resp.Body.String() != "" {
		if err := json.Unmarshal([]byte(resp.Body.String()), &params); err != nil {
			//t.AssertNoError(err)
			t.Logf(resp.Body.String())
		}
	}

	return &Response{
		Resp: resp,
		Code: resp.Code,
		Json: params,
		Body: resp.Body.String(),
	}
}
