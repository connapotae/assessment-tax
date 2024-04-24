package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

type StubAdmin struct {
	errs error
}

func (s StubAdmin) UpdateDeductionAmount(float64, string) error {
	return s.errs
}

func TestAdmin(t *testing.T) {
	tests := []struct {
		name       string
		deductType string
		req        string
		stub       StubAdmin
		want       any
	}{
		{name: "given unable to setting personal deduction should return 500 and error message", deductType: "personal", req: `{ "amount": 70000.0 }`, stub: StubAdmin{errs: echo.ErrInternalServerError}, want: http.StatusInternalServerError},
		{name: "given unable to setting personal deduction should return 400 and error message", deductType: "personal", req: `{ "amount": 9000.0 }`, stub: StubAdmin{}, want: http.StatusBadRequest},
		{name: "given unable to setting personal deduction with wrong path should return 400 and error message", deductType: "", req: `{ "amount": 70000.0 }`, stub: StubAdmin{}, want: http.StatusBadRequest},
		{name: "given unable to setting k-receipt deduction should return 500 and error message", deductType: "k-receipt", req: `{ "amount": 70000.0 }`, stub: StubAdmin{errs: echo.ErrInternalServerError}, want: http.StatusInternalServerError},
		{name: "given unable to setting k-receipt deduction should return 400 and error message", deductType: "k-receipt", req: `{ "amount": 200000.0 }`, stub: StubAdmin{}, want: http.StatusBadRequest},
		{name: "given unable to setting k-receipt deduction with wrong data type should return 400 and error message", deductType: "k-receipt", req: `{ "amount": "test" }`, stub: StubAdmin{}, want: http.StatusBadRequest},
		{name: "given unable to setting k-receipt deduction with wrong path should return 400 and error message", deductType: "", req: `{ "amount": 70000.0 }`, stub: StubAdmin{}, want: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.req))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/admin/deductions/:deductType")
			c.SetParamNames("deductType")
			c.SetParamValues(tt.deductType)

			p := New(tt.stub)
			p.SetupDeductionHandler(c)

			if rec.Code != tt.want {
				t.Errorf("expected status code %d but got %d", tt.want, rec.Code)
			}
		})
	}

	tests2 := []struct {
		name       string
		deductType string
		req        string
		stub       StubAdmin
		want       any
	}{
		{
			name:       "given user able to setting personal deduction should return personal deduction",
			deductType: "personal",
			req:        `{ "amount": 70000.0 }`,
			stub:       StubAdmin{},
			want:       DeductRes{PersonalDeduction: 70000.0},
		},
		{
			name:       "given user able to setting k-receipt deduction should return k-receipt deduction",
			deductType: "k-receipt",
			req:        `{ "amount": 70000.0 }`,
			stub:       StubAdmin{},
			want:       DeductRes{KReceiptDeduction: 70000.0},
		},
	}
	for _, tt := range tests2 {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.req))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/admin/deductions/:deductType")
			c.SetParamNames("deductType")
			c.SetParamValues(tt.deductType)

			p := New(tt.stub)
			p.SetupDeductionHandler(c)

			gotJson := rec.Body.Bytes()
			var got DeductRes
			if err := json.Unmarshal(gotJson, &got); err != nil {
				t.Errorf("unable to unmarshal json: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("expected %v but got %v", tt.want, got)
			}
		})
	}
}
