package clientdb

import (
	"context"
	"fmt"
	"github.com/tsawler/goblender/client/clienthandlers/clientmodels"
	"time"
)

// InsertCreditApp saves a credit application
func (m *DBModel) InsertCreditApp(a clientmodels.CreditApp) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
	INSERT INTO credit_applications (first_name, last_name, email, phone, address, city, province, zip, 
	                   vehicle, processed, created_at, updated_at)
    VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `

	_, err := m.DB.ExecContext(ctx,
		stmt,
		a.FirstName,
		a.LastName,
		a.Email,
		a.Phone,
		a.Address,
		a.City,
		a.Province,
		a.Zip,
		a.Vehicle,
		a.Processed,
		a.CreatedAt,
		a.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

// GetCreditApp gets an app
func (m *DBModel) GetCreditApp(id int) (clientmodels.CreditApp, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var c clientmodels.CreditApp
	query := `select id, first_name, last_name, email, phone, address, city, province, zip, vehicle, 
			created_at, updated_at from credit_applications where id = ?`
	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&c.ID,
		&c.FirstName,
		&c.LastName,
		&c.Email,
		&c.Phone,
		&c.Address,
		&c.City,
		&c.Province,
		&c.Zip,
		&c.Vehicle,
		&c.CreatedAt,
		&c.UpdatedAt,
	)

	if err != nil {
		fmt.Println(err)
		return c, err
	}

	return c, nil
}

// CreditJSON generates JSON for searching credit apps in admin tool
func (m *DBModel) CreditJSON(query, baseQuery string) ([]*clientmodels.CreditApp, int, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rowCount := 0
	filterCount := 0

	// count all rows
	allRows, err := m.DB.QueryContext(ctx, "select count(id) as all_rows from credit_applications")
	if err != nil {
		fmt.Println("Error getting all rows", err)
		return nil, 0, 0, err
	}
	defer allRows.Close()

	for allRows.Next() {
		err = allRows.Scan(&rowCount)
		if err != nil {
			fmt.Println(err)
		}
	}

	// count filtered rows
	filteredRows, err := m.DB.QueryContext(ctx, baseQuery)
	if err != nil {
		fmt.Println("Error getting filtered rows", err)
		return nil, 0, 0, err
	}
	defer filteredRows.Close()

	for filteredRows.Next() {
		_ = filteredRows.Scan(&filterCount)
	}

	//fmt.Println("Query:", query)
	rows, err := m.DB.Query(query)
	if err != nil {
		fmt.Println("Error running query", err)
		return nil, 0, 0, err
	}
	defer rows.Close()

	v := []*clientmodels.CreditApp{}

	for rows.Next() {
		s := &clientmodels.CreditApp{}
		err = rows.Scan(&s.ID, &s.FirstName, &s.LastName, &s.CreatedAt)
		if err != nil {
			return nil, 0, 0, err
		}
		v = append(v, s)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, 0, err
	}

	return v, rowCount, filterCount, nil
}
