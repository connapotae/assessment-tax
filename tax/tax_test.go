package tax

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

type StubTax struct {
	taxLevel []TaxLevel
	deduct   []Deduct
	err      error
}

func (s StubTax) GetTaxLevel(amount float64) ([]TaxLevel, error) {
	return s.taxLevel, s.err
}

func (s StubTax) GetDeduct() ([]Deduct, error) {
	return s.deduct, s.err
}

func TestTax(t *testing.T) {
	stubRefactoring := StubTax{
		taxLevel: []TaxLevel{
			{
				MinAmount:  0,
				MaxAmount:  150000,
				TaxPercent: 0,
			},
			{
				MinAmount:  150000,
				MaxAmount:  500000,
				TaxPercent: 10,
			},
		},
		deduct: []Deduct{
			{
				DeductType:   "personal",
				DeductAmount: 60000,
			},
			{
				DeductType:   "donation",
				DeductAmount: 100000,
			},
			{
				DeductType:   "k-receipt",
				DeductAmount: 50000,
			},
		},
	}

	tests := []struct {
		name string
		req  string
		stub StubTax
		want any
	}{
		{name: "given unable to get tax calculations should return 500 and error message", req: "", stub: StubTax{err: echo.ErrInternalServerError}, want: http.StatusInternalServerError},
		{name: "given unable to get tax calculations should return 400 and error message", req: "test tax calculations", stub: StubTax{}, want: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.req))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/tax/calculations")

			p := New(tt.stub)
			p.TaxCalculationsHandler(c)

			if rec.Code != tt.want {
				t.Errorf("expected status code %d but got %d", tt.want, rec.Code)
			}
		})
	}

	tests2 := []struct {
		name string
		req  string
		stub StubTax
		want any
	}{
		{
			name: "given user able to getting tax calculations should return tax",
			req:  `{ "totalIncome": 500000.0, "wht": 0.0, "allowances": [ { "allowanceType": "donation", "amount": 0.0 }]}`,
			stub: stubRefactoring,
			want: Tax{Tax: 29000.0},
		},
		{
			name: "given user able to getting tax calculations with wht should return tax",
			req:  `{ "totalIncome": 500000.0, "wht": 25000.0, "allowances": [ { "allowanceType": "donation", "amount": 0.0 }]}`,
			stub: stubRefactoring,
			want: Tax{Tax: 4000.0},
		},
		{
			name: "given user able to getting tax calculations with wht should return tax and tax refund",
			req:  `{ "totalIncome": 500000.0, "wht": 35000.0, "allowances": [ { "allowanceType": "donation", "amount": 0.0 }]}`,
			stub: stubRefactoring,
			want: Tax{Tax: 0.0, TaxRefund: 6000.0},
		},
		{
			name: "given user able to getting tax calculations with deduct donation should return tax",
			req:  `{ "totalIncome": 500000.0, "wht": 0.0, "allowances": [ { "allowanceType": "donation", "amount": 200000.0 }]}`,
			stub: stubRefactoring,
			want: Tax{Tax: 19000.0},
		},
	}
	for _, tt := range tests2 {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.req))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/tax/calculations")

			p := New(tt.stub)
			p.TaxCalculationsHandler(c)

			gotJson := rec.Body.Bytes()
			var got Tax
			if err := json.Unmarshal(gotJson, &got); err != nil {
				t.Errorf("unable to unmarshal json: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("expected %v but got %v", tt.want, got)
			}
		})
	}
}
