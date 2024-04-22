package postgres

import (
	"database/sql"

	"github.com/connapotae/assessment-tax/tax"
)

func (p *Postgres) GetTaxLevel(amount float64) ([]tax.TaxLevel, error) {
	var rows *sql.Rows
	var err error
	sql := `select min_amount, max_amount, tax_percent
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
			&l.MinAmount,
			&l.MaxAmount,
			&l.TaxPercent,
		)
		if err != nil {
			return nil, err
		}
		levels = append(levels, tax.TaxLevel{
			MinAmount:  l.MinAmount,
			MaxAmount:  l.MaxAmount,
			TaxPercent: l.TaxPercent,
		})
	}
	return levels, nil
}
