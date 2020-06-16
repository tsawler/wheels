package clientdb

import (
	"context"
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
