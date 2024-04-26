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

type TaxCSV struct {
	TotalIncome float64 `csv:"totalIncome"`
	Wht         float64 `csv:"wht"`
	Donation    float64 `csv:"donation"`
}

type Tax struct {
	Tax       float64    `json:"tax"`
	TaxRefund float64    `json:"taxRefund,omitempty"`
	TaxLevel  []TaxLevel `json:"taxLevel"`
}

type TaxLevel struct {
	Level string  `json:"level"`
	Tax   float64 `json:"tax"`
}

type Taxes struct {
	Taxes []TaxesDetail `json:"taxes"`
}

type TaxesDetail struct {
	TotalIncome float64 `json:"totalIncome"`
	Tax         float64 `json:"tax"`
	TaxRefund   float64 `json:"taxRefund,omitempty"`
}

type Err struct {
	Message string `json:"message"`
}

type ValidateErr struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidateCSVErr struct {
	Message string        `json:"message"`
	Data    []ValidateErr `json:"data"`
}

type TBTaxLevel struct {
	Id         int     `postgres:"id" json:"id"`
	Level      int     `postgres:"level" json:"level"`
	Label      string  `postgres:"label" json:"label"`
	MinAmount  float64 `postgres:"min_amount" json:"minAmount"`
	MaxAmount  float64 `postgres:"max_amount" json:"maxAmount"`
	TaxPercent int     `postgres:"tax_percent" json:"taxPercent"`
}

type TBDeduct struct {
	Id           int     `postgres:"id" json:"id"`
	DeductType   string  `postgres:"deduct_type" json:"deductType"`
	DeductAmount float64 `postgres:"deduct_amount" json:"deductAmount"`
}
