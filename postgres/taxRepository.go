package postgres

import (
	"database/sql"

	"github.com/connapotae/assessment-tax/tax"
)

func (p *Postgres) GetTaxLevel(amount float64) ([]tax.TaxLevel, error) {
	var rows *sql.Rows
	var err error
	sql := `select label, min_amount, max_amount, tax_percent
			from tax_level
			where level <= (select level from tax_level where $1::numeric <@ numrange(min_amount, max_amount, '(]'))`
	rows, err = p.Db.Query(sql, amount)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var levels []tax.TaxLevel
	for rows.Next() {
		var l tax.TaxLevel
		err := rows.Scan(
			&l.Label,
			&l.MinAmount,
			&l.MaxAmount,
			&l.TaxPercent,
		)
		if err != nil {
			return nil, err
		}
		levels = append(levels, tax.TaxLevel{
			Label:      l.Label,
			MinAmount:  l.MinAmount,
			MaxAmount:  l.MaxAmount,
			TaxPercent: l.TaxPercent,
		})
	}
	return levels, nil
}

func (p *Postgres) GetDeduct() ([]tax.Deduct, error) {
	var rows *sql.Rows
	var err error
	sql := `select deduct_type, deduct_amount from deduction`
	rows, err = p.Db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deduct []tax.Deduct
	for rows.Next() {
		var d tax.Deduct
		err := rows.Scan(
			&d.DeductType,
			&d.DeductAmount,
		)
		if err != nil {
			return nil, err
		}
		deduct = append(deduct, tax.Deduct{
			DeductType:   d.DeductType,
			DeductAmount: d.DeductAmount,
		})
	}

	return deduct, nil
}
