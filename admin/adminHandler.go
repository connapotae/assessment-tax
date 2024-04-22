package admin

import (
	"net/http"

	"github.com/connapotae/assessment-tax/tax"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	store Storer
}

type Storer interface {
	UpdateDeductionAmount(amount float64, types string) error
}

func New(db Storer) *Handler {
	return &Handler{store: db}
}

func (h *Handler) SetupDeductionHandler(c echo.Context) error {
	var res any
	var a AdminDeduction
	if err := c.Bind(&a); err != nil {
		return c.JSON(http.StatusBadRequest, tax.Err{Message: err.Error()})
	}
	type validates struct {
		condition string
		err       string
	}

	m := make(map[string]validates)
	deductType := c.Param("deductType")
	switch deductType {
	case "personal":
		m[deductType] = validates{condition: "required,gte=10000,lte=100000", err: "amount must between 10,000 and 100,000"}
		res = DeductRes{PersonalDeduction: a.Amount}
	default:
		return c.JSON(http.StatusBadRequest, Err{Message: "deduct type not support"})
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Var(a.Amount, m[deductType].condition); err != nil {
		return c.JSON(http.StatusBadRequest, m[deductType].err)
	}

	if err := h.store.UpdateDeductionAmount(a.Amount, deductType); err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: err.Error()})
	}

	return c.JSON(http.StatusCreated, res)

}
