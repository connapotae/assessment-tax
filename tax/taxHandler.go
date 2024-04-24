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
	for _, val := range allLevels {
		net := netIncome
		eachtax := 0.0
		if net > val.MinAmount && net <= val.MaxAmount {
			net = net - val.MinAmount
			eachtax = (net * float64(val.TaxPercent)) / 100
		} else if net > val.MaxAmount {
			eachtax = ((val.MaxAmount - val.MinAmount) * float64(val.TaxPercent)) / 100
		}
		tax += eachtax
		taxLevel = append(taxLevel, TaxLevel{
			Level: val.Label,
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

	var eachTax []TaxesDetail
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

		for _, val := range allLevels {
			net := netIncome
			if net > val.MinAmount && net <= val.MaxAmount {
				net = net - val.MinAmount
				tax = tax + (net*float64(val.TaxPercent))/100
			} else if net > val.MaxAmount {
				tax = tax + ((val.MaxAmount-val.MinAmount)*float64(val.TaxPercent))/100
			}
		}

		tax = tax - wht

		if tax < 0 {
			eachTax = append(eachTax, TaxesDetail{
				TotalIncome: t.TotalIncome,
				Tax:         0.0,
				TaxRefund:   math.Abs(tax),
			})
		} else {
			eachTax = append(eachTax, TaxesDetail{
				TotalIncome: t.TotalIncome,
				Tax:         tax,
			})
		}
	}

	res := Taxes{
		Taxes: eachTax,
	}

	return c.JSON(http.StatusOK, res)
}
