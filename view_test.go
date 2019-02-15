package fox

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/labstack/echo"
)

func GenContext(rec *httptest.ResponseRecorder, target string, f url.Values) echo.Context {
	e := echo.New()
	var req *http.Request
	if f != nil {
		req = httptest.NewRequest(http.MethodPost, target, strings.NewReader(f.Encode()))
	} else {
		req = httptest.NewRequest(http.MethodPost, target, nil)
	}

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	//rec = httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c
}

func TestNewModelView(t *testing.T) {
	type args struct {
		model      interface{}
		modelSlice interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    *ModelView
		wantErr bool
	}{
		{"BaseTest", args{&Person{}, []Person{}}, &ModelView{"person", DefaultMessage, &GORMDrive{&Person{}, []Person{}}}, false},
		{"ErrorTest", args{Person{}, []Person{}}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewModelView(tt.args.model, tt.args.modelSlice)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewModelView() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewModelView() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModelView_FuncPOST(t *testing.T) {
	type fields struct {
		Name      string
		Message   map[int]interface{}
		gormDrive *GORMDrive
		rec       *httptest.ResponseRecorder
	}
	type args struct {
		target string
		val    url.Values
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"BaseTest", fields{"person", DefaultMessage, &GORMDrive{&Person{}, []Person{}}, httptest.NewRecorder()}, args{"/person", url.Values{"name": []string{"Jon"}, "age": []string{"18"}}}, false},
		{"ErrorTest", fields{"person", DefaultMessage, &GORMDrive{&Person{}, []Person{}}, httptest.NewRecorder()}, args{"/person", url.Values{"name": []string{"Jon"}, "age": []string{"asd"}}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mv := &ModelView{
				Name:      tt.fields.Name,
				Message:   tt.fields.Message,
				gormDrive: tt.fields.gormDrive,
			}
			tt.fields.rec = httptest.NewRecorder()
			c := GenContext(tt.fields.rec, tt.args.target, tt.args.val)
			mv.FuncPOST(c)
			resp := map[string]interface{}{}
			json.Unmarshal(tt.fields.rec.Body.Bytes(), &resp)
			if tt.wantErr && tt.fields.rec.Code < 400 {
				t.Errorf("rec.Code = %v, want error %v", tt.fields.rec.Code, tt.wantErr)
			} else if !tt.wantErr {
				if tt.fields.rec.Code >= 400 {
					t.Errorf("rec.Code = %v, want error %v", tt.fields.rec.Code, tt.wantErr)
				}
				if _, ok := resp["data"]; !ok {
					t.Errorf("No response data")
				}
			}
		})
	}
}

func TestModelView_FuncDELETE(t *testing.T) {

	type fields struct {
		Name      string
		Message   map[int]interface{}
		gormDrive *GORMDrive
		rec       *httptest.ResponseRecorder
	}
	type args struct {
		val url.Values
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"BaseTest", fields{"person", DefaultMessage, &GORMDrive{&Person{}, []Person{}}, httptest.NewRecorder()}, args{url.Values{"id": []string{"10"}, "id.or": []string{"100"}}}, false},
		{"BaseTest", fields{"person", DefaultMessage, &GORMDrive{&Person{}, []Person{}}, httptest.NewRecorder()}, args{url.Values{"id": []string{"10"}, "id.or": []string{"dwwd"}}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mv := &ModelView{
				Name:      tt.fields.Name,
				Message:   tt.fields.Message,
				gormDrive: tt.fields.gormDrive,
			}
			tt.fields.rec = httptest.NewRecorder()
			c := GenContext(tt.fields.rec, "/person?"+tt.args.val.Encode(), nil)
			mv.FuncDELETE(c)
			resp := map[string]interface{}{}
			json.Unmarshal(tt.fields.rec.Body.Bytes(), &resp)
			if tt.wantErr && tt.fields.rec.Code < 400 {
				t.Errorf("rec.Code = %v, want error %v", tt.fields.rec.Code, tt.wantErr)
			} else if !tt.wantErr {
				if tt.fields.rec.Code >= 400 {
					t.Errorf("rec.Code = %v, want error %v", tt.fields.rec.Code, tt.wantErr)
				}
				if _, ok := resp["data"]; !ok {
					t.Errorf("No response data")
				}
			}
		})
	}
}

func TestModelView_FuncPUT(t *testing.T) {
	type fields struct {
		Name      string
		Message   map[int]interface{}
		gormDrive *GORMDrive
		rec       *httptest.ResponseRecorder
	}
	type args struct {
		query url.Values
		data  url.Values
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"BaseTest", fields{"person", DefaultMessage, &GORMDrive{&Person{}, []Person{}}, httptest.NewRecorder()}, args{url.Values{"id": []string{"1"}, "age.gt": []string{"36"}}, url.Values{"age": []string{"17"}}}, false},
		{"BaseTest", fields{"person", DefaultMessage, &GORMDrive{&Person{}, []Person{}}, httptest.NewRecorder()}, args{url.Values{"id": []string{"1"}, "age.gt": []string{"36"}}, url.Values{"age": []string{"fee"}}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mv := &ModelView{
				Name:      tt.fields.Name,
				Message:   tt.fields.Message,
				gormDrive: tt.fields.gormDrive,
			}
			tt.fields.rec = httptest.NewRecorder()
			c := GenContext(tt.fields.rec, "/person?"+tt.args.query.Encode(), tt.args.data)
			mv.FuncPUT(c)
			resp := map[string]interface{}{}
			json.Unmarshal(tt.fields.rec.Body.Bytes(), &resp)
			if tt.wantErr && tt.fields.rec.Code < 400 {
				t.Errorf("rec.Code = %v, want error %v", tt.fields.rec.Code, tt.wantErr)
			} else if !tt.wantErr {
				if tt.fields.rec.Code >= 400 {
					t.Errorf("rec.Code = %v, want error %v", tt.fields.rec.Code, tt.wantErr)
				}
				if _, ok := resp["data"]; !ok {
					t.Errorf("No response data")
				}
			}
		})
	}
}

func TestModelView_FuncGET(t *testing.T) {
	type fields struct {
		Name      string
		Message   map[int]interface{}
		gormDrive *GORMDrive
		rec       *httptest.ResponseRecorder
	}
	type args struct {
		val url.Values
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"BaseTest", fields{"person", DefaultMessage, &GORMDrive{&Person{}, []Person{}}, httptest.NewRecorder()}, args{url.Values{"id": []string{"10"}, "id.or": []string{"100"}}}, false},
		{"BaseTest", fields{"person", DefaultMessage, &GORMDrive{&Person{}, []Person{}}, httptest.NewRecorder()}, args{url.Values{"id": []string{"10"}, "id.or": []string{"dwwd"}}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mv := &ModelView{
				Name:      tt.fields.Name,
				Message:   tt.fields.Message,
				gormDrive: tt.fields.gormDrive,
			}
			tt.fields.rec = httptest.NewRecorder()
			c := GenContext(tt.fields.rec, "/person?"+tt.args.val.Encode(), nil)
			mv.FuncGET(c)
			resp := map[string]interface{}{}
			json.Unmarshal(tt.fields.rec.Body.Bytes(), &resp)
			if tt.wantErr && tt.fields.rec.Code < 400 {
				t.Errorf("rec.Code = %v, want error %v", tt.fields.rec.Code, tt.wantErr)
			} else if !tt.wantErr {
				if tt.fields.rec.Code >= 400 {
					t.Errorf("rec.Code = %v, want error %v", tt.fields.rec.Code, tt.wantErr)
				}
				if _, ok := resp["data"]; !ok {
					t.Errorf("No response data")
				}
			}
		})
	}
}

func TestModelView_EnableDefault(t *testing.T) {
	type fields struct {
		Name      string
		Message   map[int]interface{}
		gormDrive *GORMDrive
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{"BaseTest", fields{"person", DefaultMessage, &GORMDrive{&Person{}, []Person{}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mv := &ModelView{
				Name:      tt.fields.Name,
				Message:   tt.fields.Message,
				gormDrive: tt.fields.gormDrive,
			}
			mv.EnableDefault()
		})
	}
}
