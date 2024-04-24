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

type validates struct {
	condition string
	errString string
	res       any
}

func condition(m map[string]validates, dtype string) string {
	return m[dtype].condition
}

func errString(m map[string]validates, dtype string) string {
	return m[dtype].errString
}

func response(m map[string]validates, dtype string) any {
	return m[dtype].res
}

func (h *Handler) SetupDeductionHandler(c echo.Context) error {
	var a AdminDeduction
	if err := c.Bind(&a); err != nil {
		return c.JSON(http.StatusBadRequest, tax.Err{Message: err.Error()})
	}

	m := map[string]validates{
		"personal":  {condition: "required,gte=10000,lte=100000", errString: "amount must between 10,000 and 100,000", res: DeductRes{PersonalDeduction: a.Amount}},
		"k-receipt": {condition: "required,gte=0,lte=100000", errString: "amount must between 0 and 100,000", res: DeductRes{KReceiptDeduction: a.Amount}},
	}

	deductType := c.Param("deductType")

	if _, ok := m[deductType]; !ok {
		return c.JSON(http.StatusBadRequest, Err{Message: "deduct type not support"})
	}

	v := validates{
		condition: condition(m, deductType),
		errString: errString(m, deductType),
		res:       response(m, deductType),
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Var(a.Amount, v.condition); err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: v.errString})
	}

	if err := h.store.UpdateDeductionAmount(a.Amount, deductType); err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: err.Error()})
	}

	return c.JSON(http.StatusCreated, v.res)

}
