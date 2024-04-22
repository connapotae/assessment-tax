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
	err      error
}

func (s StubTax) GetTaxLevel(amount float64) ([]TaxLevel, error) {
	return s.taxLevel, s.err
}

func TestTax(t *testing.T) {
	t.Run("given unable to get tax calculations should return 500 and error message", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/tax/calculations")

		stubError := StubTax{err: echo.ErrInternalServerError}
		p := New(stubError)
		p.TaxCalculationsHandler(c)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status code %d but got %d", http.StatusInternalServerError, rec.Code)
		}
	})

	t.Run("given unable to get tax calculations should return 400 and error message", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`test tax calculations`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/tax/calculations")

		p := New(StubTax{})
		p.TaxCalculationsHandler(c)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status code %d but got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("given user able to getting tax calculations should return tax", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`
		{
			"totalIncome": 500000.0,
			"wht": 0.0,
			"allowances": [
			  {
				"allowanceType": "donation",
				"amount": 0.0
			  }
			]
		}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/tax/calculations")

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
		}

		p := New(stubRefactoring)
		p.TaxCalculationsHandler(c)

		want := Tax{
			Tax: 29000.0,
		}
		gotJson := rec.Body.Bytes()
		var got Tax
		if err := json.Unmarshal(gotJson, &got); err != nil {
			t.Errorf("unable to unmarshal json: %v", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("expected %v but got %v", want, got)
		}
	})

	t.Run("given user able to getting tax calculations with wht should return tax", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`
		{
			"totalIncome": 500000.0,
			"wht": 25000.0,
			"allowances": [
			  {
				"allowanceType": "donation",
				"amount": 0.0
			  }
			]
		}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/tax/calculations")

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
		}

		p := New(stubRefactoring)
		p.TaxCalculationsHandler(c)

		want := Tax{
			Tax: 4000.0,
		}
		gotJson := rec.Body.Bytes()
		var got Tax
		if err := json.Unmarshal(gotJson, &got); err != nil {
			t.Errorf("unable to unmarshal json: %v", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("expected %v but got %v", want, got)
		}
	})

	t.Run("given user able to getting tax calculations with wht should return tax and tax refund", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`
		{
			"totalIncome": 500000.0,
			"wht": 35000.0,
			"allowances": [
			  {
				"allowanceType": "donation",
				"amount": 0.0
			  }
			]
		}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/tax/calculations")

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
		}

		p := New(stubRefactoring)
		p.TaxCalculationsHandler(c)

		want := Tax{
			Tax:       0.0,
			TaxRefund: 6000.0,
		}
		gotJson := rec.Body.Bytes()
		var got Tax
		if err := json.Unmarshal(gotJson, &got); err != nil {
			t.Errorf("unable to unmarshal json: %v", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("expected %v but got %v", want, got)
		}
	})
}
