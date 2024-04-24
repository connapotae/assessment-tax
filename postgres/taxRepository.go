package postgres

import (
	"database/sql"

	"github.com/connapotae/assessment-tax/tax"
)

func (p *Postgres) GetTaxLevels() ([]tax.TBTaxLevel, error) {
	var rows *sql.Rows
	var err error
	sql := `select level, label, min_amount, max_amount, tax_percent from tax_level`
	rows, err = p.Db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var levels []tax.TBTaxLevel
	for rows.Next() {
		var l tax.TBTaxLevel
		err := rows.Scan(
			&l.Level,
			&l.Label,
			&l.MinAmount,
			&l.MaxAmount,
			&l.TaxPercent,
		)
		if err != nil {
			return nil, err
		}
		levels = append(levels, tax.TBTaxLevel{
			Level:      l.Level,
			Label:      l.Label,
			MinAmount:  l.MinAmount,
			MaxAmount:  l.MaxAmount,
			TaxPercent: l.TaxPercent,
		})
	}
	return levels, nil
}

func (p *Postgres) GetTaxLevel(amount float64) (int, error) {
	sql := `select level
			from tax_level
			where $1::numeric <@ numrange(min_amount, max_amount, '(]')`
	rows := p.Db.QueryRow(sql, amount)
	var l int
	err := rows.Scan(&l)
	if err != nil {
		return 0, err
	}
	return l, nil
}

func (p *Postgres) GetDeduct() ([]tax.TBDeduct, error) {
	var rows *sql.Rows
	var err error
	sql := `select deduct_type, deduct_amount from deduction`
	rows, err = p.Db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deduct []tax.TBDeduct
	for rows.Next() {
		var d tax.TBDeduct
		err := rows.Scan(
			&d.DeductType,
			&d.DeductAmount,
		)
		if err != nil {
			return nil, err
		}
		deduct = append(deduct, tax.TBDeduct{
			DeductType:   d.DeductType,
			DeductAmount: d.DeductAmount,
		})
	}

	return deduct, nil
}

func (p *Postgres) UpdateDeductionAmount(amount float64, types string) error {
	_, err := p.Db.Exec("UPDATE deduction SET deduct_amount = $1 WHERE deduct_type = $2", amount, types)
	if err != nil {
		return err
	}
	return nil
}
