package tax

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

type StubTax struct {
	taxLevel []TBTaxLevel
	deduct   []TBDeduct
	err      error
}

func (s StubTax) GetTaxLevels() ([]TBTaxLevel, error) {
	return s.taxLevel, s.err
}

func (s StubTax) GetDeduct() ([]TBDeduct, error) {
	return s.deduct, s.err
}

func TestTax(t *testing.T) {
	stubRefactoring := StubTax{
		taxLevel: []TBTaxLevel{
			{
				Level:      1,
				Label:      "0-150,000",
				MinAmount:  0,
				MaxAmount:  150000,
				TaxPercent: 0,
			},
			{
				Level:      2,
				Label:      "150,001-500,000",
				MinAmount:  150000,
				MaxAmount:  500000,
				TaxPercent: 10,
			},
			{
				Level:      3,
				Label:      "500,001-1,000,000",
				MinAmount:  500000,
				MaxAmount:  1000000,
				TaxPercent: 15,
			},
			{
				Level:      4,
				Label:      "1,000,001-2,000,000",
				MinAmount:  1000000,
				MaxAmount:  2000000,
				TaxPercent: 20,
			},
			{
				Level:      5,
				Label:      "2,000,001 ขึ้นไป",
				MinAmount:  2000000,
				MaxAmount:  999999999999,
				TaxPercent: 35,
			},
		},
		deduct: []TBDeduct{
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
		{name: "given unable to get tax calculations should return 500 and error message", req: `{ "totalIncome": 500000.0, "wht": 0.0, "allowances": [ { "allowanceType": "donation", "amount": 0.0 }]}`, stub: StubTax{err: echo.ErrInternalServerError}, want: http.StatusInternalServerError},
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
			want: Tax{Tax: 29000.0, TaxLevel: []TaxLevel{{Level: "0-150,000", Tax: 0}, {Level: "150,001-500,000", Tax: 29000.0}, {Level: "500,001-1,000,000", Tax: 0.0}, {Level: "1,000,001-2,000,000", Tax: 0.0}, {Level: "2,000,001 ขึ้นไป", Tax: 0.0}}},
		},
		{
			name: "given user able to getting tax calculations with wht should return tax",
			req:  `{ "totalIncome": 500000.0, "wht": 25000.0, "allowances": [ { "allowanceType": "donation", "amount": 0.0 }]}`,
			stub: stubRefactoring,
			want: Tax{Tax: 4000.0, TaxLevel: []TaxLevel{{Level: "0-150,000", Tax: 0}, {Level: "150,001-500,000", Tax: 29000.0}, {Level: "500,001-1,000,000", Tax: 0.0}, {Level: "1,000,001-2,000,000", Tax: 0.0}, {Level: "2,000,001 ขึ้นไป", Tax: 0.0}}},
		},
		{
			name: "given user able to getting tax calculations with wht should return tax and tax refund",
			req:  `{ "totalIncome": 500000.0, "wht": 35000.0, "allowances": [ { "allowanceType": "donation", "amount": 0.0 }]}`,
			stub: stubRefactoring,
			want: Tax{Tax: 0.0, TaxRefund: 6000.0, TaxLevel: []TaxLevel{{Level: "0-150,000", Tax: 0}, {Level: "150,001-500,000", Tax: 29000.0}, {Level: "500,001-1,000,000", Tax: 0.0}, {Level: "1,000,001-2,000,000", Tax: 0.0}, {Level: "2,000,001 ขึ้นไป", Tax: 0.0}}},
		},
		{
			name: "given user able to getting tax calculations with deduct donation should return tax",
			req:  `{ "totalIncome": 500000.0, "wht": 0.0, "allowances": [ { "allowanceType": "donation", "amount": 200000.0 }]}`,
			stub: stubRefactoring,
			want: Tax{Tax: 19000.0, TaxLevel: []TaxLevel{{Level: "0-150,000", Tax: 0}, {Level: "150,001-500,000", Tax: 19000.0}, {Level: "500,001-1,000,000", Tax: 0.0}, {Level: "1,000,001-2,000,000", Tax: 0.0}, {Level: "2,000,001 ขึ้นไป", Tax: 0.0}}},
		},
		{
			name: "given user able to getting tax calculations should return tax and tax by level",
			req:  `{ "totalIncome": 500000.0, "wht": 0.0, "allowances": [ { "allowanceType": "donation", "amount": 200000.0 }]}`,
			stub: stubRefactoring,
			want: Tax{Tax: 19000.0, TaxLevel: []TaxLevel{{Level: "0-150,000", Tax: 0}, {Level: "150,001-500,000", Tax: 19000.0}, {Level: "500,001-1,000,000", Tax: 0.0}, {Level: "1,000,001-2,000,000", Tax: 0.0}, {Level: "2,000,001 ขึ้นไป", Tax: 0.0}}},
		},
		{
			name: "given user able to getting tax calculations with deduct donation and k-receipt more than maximum should return tax and tax by level",
			req:  `{ "totalIncome": 500000.0, "wht": 0.0, "allowances": [{ "allowanceType": "k-receipt", "amount": 200000.0 }, { "allowanceType": "donation", "amount": 100000.0 }]}`,
			stub: stubRefactoring,
			want: Tax{Tax: 14000.0, TaxLevel: []TaxLevel{{Level: "0-150,000", Tax: 0}, {Level: "150,001-500,000", Tax: 14000.0}, {Level: "500,001-1,000,000", Tax: 0.0}, {Level: "1,000,001-2,000,000", Tax: 0.0}, {Level: "2,000,001 ขึ้นไป", Tax: 0.0}}},
		},
		{
			name: "given user able to getting tax calculations with deduct donation and k-receipt less than maximum should return tax and tax by level",
			req:  `{ "totalIncome": 500000.0, "wht": 0.0, "allowances": [{ "allowanceType": "k-receipt", "amount": 3000.0 }, { "allowanceType": "donation", "amount": 100000.0 }]}`,
			stub: stubRefactoring,
			want: Tax{Tax: 18700.0, TaxLevel: []TaxLevel{{Level: "0-150,000", Tax: 0}, {Level: "150,001-500,000", Tax: 18700.0}, {Level: "500,001-1,000,000", Tax: 0.0}, {Level: "1,000,001-2,000,000", Tax: 0.0}, {Level: "2,000,001 ขึ้นไป", Tax: 0.0}}},
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

	t.Run("given unable to get tax calculations from csv contain wrong data should return 500 and error message", func(t *testing.T) {
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", "file.csv")
		if err != nil {
			t.Fatal(err)
		}
		file := strings.NewReader(`totalIncome,wht,donation
		500000,0,0
		600000,40000,20000
		750000,50000,test`)
		io.Copy(part, file)
		writer.Close()

		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/", body)
		req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/tax/calculations/upload-csv")

		p := New(stubRefactoring)
		p.TaxCalculationsCSVHandler(c)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status code %d but got %d", http.StatusInternalServerError, rec.Code)
		}
	})

	t.Run("given unable to get tax calculations from csv have no file should return 400 and error message", func(t *testing.T) {
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		writer.Close()

		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/tax/calculations/upload-csv")

		p := New(stubRefactoring)
		p.TaxCalculationsCSVHandler(c)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status code %d but got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("given user able to getting tax calculations from csv should return list of tax", func(t *testing.T) {
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", "file.csv")
		if err != nil {
			t.Fatal(err)
		}
		file := strings.NewReader(`totalIncome,wht,donation
		500000,0,0
		600000,40000,20000
		750000,50000,200000`)
		io.Copy(part, file)
		writer.Close()

		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/", body)
		req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/tax/calculations/upload-csv")

		p := New(stubRefactoring)
		p.TaxCalculationsCSVHandler(c)

		gotJson := rec.Body.Bytes()
		var got Taxes
		if err := json.Unmarshal(gotJson, &got); err != nil {
			t.Errorf("unable to unmarshal json: %v", err)
		}

		want := Taxes{
			Taxes: []TaxesDetail{
				{TotalIncome: 500000.0, Tax: 29000.0},
				{TotalIncome: 600000.0, Tax: 0.0, TaxRefund: 2000.0},
				{TotalIncome: 750000.0, Tax: 0.0, TaxRefund: 1500.0},
			},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("expected %v but got %v", want, got)
		}
	})
}
