package tax

type TaxCalcualtions struct {
	TotalIncome float64      `json:"totalIncome"`
	Wht         float64      `json:"wht"`
	Allowances  []Allowances `json:"allowances"`
}

type Allowances struct {
	AllowanceType string  `json:"allowanceType"`
	Amount        float64 `json:"amount"`
}

type Tax struct {
	Tax       float64       `json:"tax"`
	TaxRefund float64       `json:"taxRefund,omitempty"`
	TaxLevel  []TaxLevelRes `json:"taxLevel"`
}

type TaxLevelRes struct {
	Level string  `json:"level"`
	Tax   float64 `json:"tax"`
}

type Err struct {
	Message string `json:"message"`
}

type TaxLevel struct {
	Id         int     `postgres:"id" json:"id"`
	Level      int     `postgres:"level" json:"level"`
	Label      string  `postgres:"label" json:"label"`
	MinAmount  float64 `postgres:"min_amount" json:"minAmount"`
	MaxAmount  float64 `postgres:"max_amount" json:"maxAmount"`
	TaxPercent int     `postgres:"tax_percent" json:"taxPercent"`
}

type Deduct struct {
	Id           int     `postgres:"id" json:"id"`
	DeductType   string  `postgres:"deduct_type" json:"deductType"`
	DeductAmount float64 `postgres:"deduct_amount" json:"deductAmount"`
}
