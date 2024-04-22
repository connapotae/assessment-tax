package admin

type Err struct {
	Message string `json:"message"`
}

type AdminDeduction struct {
	Amount float64 `json:"amount" validate:"required"`
}

type DeductRes struct {
	PersonalDeduction float64 `json:"personalDeduction,omitempty"`
}
