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
	GetTaxLevel(amount float64) (int, error)
	GetDeduct() ([]TBDeduct, error)
}

func New(db Storer) *Handler {
	return &Handler{store: db}
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
	m := make(map[string]float64)
	for _, val := range deducts {
		m[val.DeductType] = val.DeductAmount
	}

	personalDeduction := m["personal"]
	totalIncome := t.TotalIncome
	wht := t.Wht
	allowances := t.Allowances
	deduct := 0.0

	for _, val := range allowances {
		switch val.AllowanceType {
		case "donation":
			if val.Amount > m["donation"] {
				deduct = deduct + m["donation"]
			} else {
				deduct = deduct + val.Amount
			}
		case "k-receipt":
			if val.Amount > m["k-receipt"] {
				deduct = deduct + m["k-receipt"]
			} else {
				deduct = deduct + val.Amount
			}
		default:
			deduct = 0.0
		}
	}
	netIncome := (totalIncome - personalDeduction) - deduct

	allLevels, err := h.store.GetTaxLevels()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: err.Error()})
	}
	level, err := h.store.GetTaxLevel(netIncome)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: err.Error()})
	}

	var taxLevel []TaxLevel
	for _, val := range allLevels {
		net := netIncome
		if val.Level == level {
			net = net - val.MinAmount
			tax = tax + (net*float64(val.TaxPercent))/100
		} else if val.Level < level {
			net = val.MaxAmount - val.MinAmount
			tax = tax + (net*float64(val.TaxPercent))/100
		} else {
			net = 0.0
		}
		taxLevel = append(taxLevel, TaxLevel{
			Level: val.Label,
			Tax:   (net * float64(val.TaxPercent)) / 100,
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
	m := make(map[string]float64)
	for _, val := range deducts {
		m[val.DeductType] = val.DeductAmount
	}

	personalDeduction := m["personal"]

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

		if donation > m["donation"] {
			deduct = deduct + m["donation"]
		} else {
			deduct = deduct + donation
		}

		netIncome := (totalIncome - personalDeduction) - deduct
		allLevels, err := h.store.GetTaxLevels()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, Err{Message: err.Error()})
		}
		level, err := h.store.GetTaxLevel(netIncome)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, Err{Message: err.Error()})
		}

		for _, val := range allLevels {
			net := netIncome
			if val.Level == level {
				net = net - val.MinAmount
				tax = tax + (net*float64(val.TaxPercent))/100
			} else if val.Level < level {
				net = val.MaxAmount - val.MinAmount
				tax = tax + (net*float64(val.TaxPercent))/100
			} else {
				net = 0.0
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
