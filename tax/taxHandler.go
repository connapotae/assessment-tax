package tax

import (
	"io"
	"math"
	"net/http"

	"github.com/gocarina/gocsv"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	store Storer
}

type Storer interface {
	GetTaxLevels() ([]TBTaxLevel, error)
	GetDeduct() ([]TBDeduct, error)
}

func New(db Storer) *Handler {
	return &Handler{store: db}
}

func mapDeduct(deducts []TBDeduct) map[string]float64 {
	m := make(map[string]float64)
	for _, val := range deducts {
		m[val.DeductType] = val.DeductAmount
	}
	return m
}

func personalDeduct(m map[string]float64) float64 {
	return m["personal"]
}
func kReceiptDeduct(m map[string]float64) float64 {
	return m["k-receipt"]
}
func donationDeduct(m map[string]float64) float64 {
	return m["donation"]
}

func calcDeduct(allowances Allowances, m map[string]float64) float64 {
	result := 0.0
	kReceiptDeduction := kReceiptDeduct(m)
	donateDeduction := donationDeduct(m)

	switch allowances.AllowanceType {
	case "donation":
		if allowances.Amount > donateDeduction {
			result += donateDeduction
		} else {
			result += allowances.Amount
		}
	case "k-receipt":
		if allowances.Amount > kReceiptDeduction {
			result += kReceiptDeduction
		} else {
			result += allowances.Amount
		}
	default:
		result = 0.0
	}

	return result
}

func calcTaxByLevel(tbTax TBTaxLevel, income float64) float64 {
	result := 0.0
	if income > tbTax.MinAmount && income <= tbTax.MaxAmount {
		income = income - tbTax.MinAmount
		result = (income * float64(tbTax.TaxPercent)) / 100
	} else if income > tbTax.MaxAmount {
		result = ((tbTax.MaxAmount - tbTax.MinAmount) * float64(tbTax.TaxPercent)) / 100
	}
	return result
}

func (h *Handler) TaxCalculationsHandler(c echo.Context) error {
	var tax float64
	var t TaxCalcualtions
	err := c.Bind(&t)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: err.Error()})
	}

	deducts, err := h.store.GetDeduct()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: err.Error()})
	}
	m := mapDeduct(deducts)

	personalDeduction := personalDeduct(m)
	totalIncome := t.TotalIncome
	wht := t.Wht
	allowances := t.Allowances
	deduct := 0.0

	for _, a := range allowances {
		deduct += calcDeduct(a, m)
	}
	netIncome := (totalIncome - personalDeduction) - deduct

	allLevels, err := h.store.GetTaxLevels()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: err.Error()})
	}

	var taxLevel []TaxLevel
	for _, t := range allLevels {
		eachtax := calcTaxByLevel(t, netIncome)
		tax += eachtax
		taxLevel = append(taxLevel, TaxLevel{
			Level: t.Label,
			Tax:   eachtax,
		})
	}

	tax = tax - wht

	var res Tax
	if tax < 0 {
		res = Tax{
			Tax:       0.0,
			TaxRefund: math.Abs(tax),
			TaxLevel:  taxLevel,
		}
	} else {
		res = Tax{
			Tax:      tax,
			TaxLevel: taxLevel,
		}
	}

	return c.JSON(http.StatusOK, res)
}

func (h *Handler) TaxCalculationsCSVHandler(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: err.Error()})
	}

	f, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: err.Error()})
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: err.Error()})
	}

	deducts, err := h.store.GetDeduct()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: err.Error()})
	}
	m := mapDeduct(deducts)

	personalDeduction := personalDeduct(m)
	// kReceiptDeduction := kReceiptDeduct(m)
	donateDeduction := donationDeduct(m)

	var taxCsv []TaxCSV
	err = gocsv.UnmarshalBytes(data, &taxCsv)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: err.Error()})
	}

	var taxes []TaxesDetail
	for _, t := range taxCsv {
		var tax float64
		deduct := 0.0

		totalIncome := t.TotalIncome
		wht := t.Wht
		donation := t.Donation

		if donation > donateDeduction {
			deduct = deduct + donateDeduction
		} else {
			deduct = deduct + donation
		}

		netIncome := (totalIncome - personalDeduction) - deduct
		allLevels, err := h.store.GetTaxLevels()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, Err{Message: err.Error()})
		}

		for _, t := range allLevels {
			eachtax := calcTaxByLevel(t, netIncome)
			tax += eachtax
		}

		tax = tax - wht

		if tax < 0 {
			taxes = append(taxes, TaxesDetail{
				TotalIncome: t.TotalIncome,
				Tax:         0.0,
				TaxRefund:   math.Abs(tax),
			})
		} else {
			taxes = append(taxes, TaxesDetail{
				TotalIncome: t.TotalIncome,
				Tax:         tax,
			})
		}
	}

	res := Taxes{
		Taxes: taxes,
	}

	return c.JSON(http.StatusOK, res)
}
