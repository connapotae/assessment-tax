package tax

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	store Storer
}

type Storer interface {
	GetTaxLevel(amount float64) ([]TaxLevel, error)
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

	personalDeduction := 60000.0
	totalIncome := t.TotalIncome
	// wht := t.Wht
	allowances := t.Allowances
	deduct := 0.0

	for _, val := range allowances {
		switch val.AllowanceType {
		case "donation":
			deduct = deduct + val.Amount
		default:
			deduct = 0.0
		}
	}

	netIncome := (totalIncome - personalDeduction)
	levels, err := h.store.GetTaxLevel(netIncome)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: err.Error()})
	}

	for i, val := range levels {
		net := netIncome
		if i == len(levels)-1 {
			net = net - val.MinAmount
			tax = tax + (net*float64(val.TaxPercent))/100
		} else {
			net = val.MaxAmount - val.MinAmount
			tax = tax + (net*float64(val.TaxPercent))/100
		}
	}

	res := Tax{
		Tax: tax,
	}
	return c.JSON(http.StatusOK, res)
}
