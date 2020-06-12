package clientdb

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/tsawler/goblender/client/clienthandlers/clientmodels"
	"strings"
	"time"
)

// VehicleModel holds the db connection
type DBModel struct {
	DB *sql.DB
}

// AllActiveOptions returns slice of all active options
func (m *DBModel) AllActiveOptions() ([]clientmodels.Option, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var o []clientmodels.Option

	query := `select id, option_name, active, created_at, updated_at from options where active = 1 order by option_name`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		s := &clientmodels.Option{}
		err = rows.Scan(&s.ID, &s.OptionName, &s.Active, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			return nil, err
		}
		// Append it to the slice of .
		o = append(o, *s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return o, nil
}

// VehicleJSON generates JSON for searching vehicles in admin tool
func (m *DBModel) VehicleJSON(query, baseQuery string, extra ...string) ([]*clientmodels.VehicleJSON, int, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rowCount := 0
	filterCount := 0
	where := ""

	if len(extra) > 0 {
		query = strings.Replace(query, "ORDER BY", fmt.Sprintf("%s ORDER BY", extra[0]), 1)
		baseQuery = fmt.Sprintf("%s %s", baseQuery, extra[0])
		where = extra[0]
	}

	// count all rows
	allRows, err := m.DB.QueryContext(ctx, fmt.Sprintf("select count(id) as all_rows from v_all_vehicles %s", where))
	if err != nil {
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
		return nil, 0, 0, err
	}
	defer filteredRows.Close()

	for filteredRows.Next() {
		_ = filteredRows.Scan(&filterCount)
	}

	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, 0, 0, err
	}
	defer rows.Close()

	v := []*clientmodels.VehicleJSON{}

	for rows.Next() {
		s := &clientmodels.VehicleJSON{}
		err = rows.Scan(&s.ID, &s.Year, &s.Make, &s.Model, &s.Trim, &s.StockNo, &s.Vin, &s.Status, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			return nil, 0, 0, err
		}
		// Append it to the slice of .
		v = append(v, s)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, 0, err
	}

	return v, rowCount, filterCount, nil
}

// GetAllVehicles returns slice of vehicles by type
func (m *DBModel) GetAllVehicles() ([]clientmodels.Vehicle, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var v []clientmodels.Vehicle

	query := `
		select 
		       id, 
		       stock_no, 
		       coalesce(cost, 0),
		       vin, 
		       coalesce(odometer, 0),
		       coalesce(year, 0),
		       coalesce(trim, ''),
		       vehicle_type,
		       coalesce(body, ''),
		       coalesce(seating_capacity,''),
		       coalesce(drive_train,''),
		       coalesce(engine,''),
		       coalesce(exterior_color,''),
		       coalesce(interior_color,''),
		       coalesce(transmission,''),
		       coalesce(options,''),
		       coalesce(model_number, ''),
		       coalesce(total_msr,0.0),
		       v.status,
		       coalesce(description, ''),
		       vehicle_makes_id,
		       vehicle_models_id,
		       hand_picked,
		       used,
		       coalesce(price_for_display,''),
		       created_at,
		       updated_at
		from 
		     vehicles v 
		

		order by year desc
		`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		fmt.Println(err)
		return v, err
	}
	defer rows.Close()

	for rows.Next() {
		c := &clientmodels.Vehicle{}
		err = rows.Scan(
			&c.ID,
			&c.StockNo,
			&c.Cost,
			&c.Vin,
			&c.Odometer,
			&c.Year,
			&c.Trim,
			&c.VehicleType,
			&c.Body,
			&c.SeatingCapacity,
			&c.DriveTrain,
			&c.Engine,
			&c.ExteriorColour,
			&c.InteriorColour,
			&c.Transmission,
			&c.Options,
			&c.ModelNumber,
			&c.TotalMSR,
			&c.Status,
			&c.Description,
			&c.VehicleMakesID,
			&c.VehicleModelsID,
			&c.HandPicked,
			&c.Used,
			&c.PriceForDisplay,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
		if err != nil {
			fmt.Println(err)
			return v, err
		}

		// get make
		vehicleMake := clientmodels.Make{}

		query = `
			SELECT 
				id, 
				make, 
				created_at, 
				updated_at 
			FROM 
				vehicle_makes 
			WHERE 
				id = ?`
		makeRow := m.DB.QueryRowContext(ctx, query, c.VehicleMakesID)

		err = makeRow.Scan(
			&vehicleMake.ID,
			&vehicleMake.Make,
			&vehicleMake.CreatedAt,
			&vehicleMake.UpdatedAt)
		if err != nil {
			fmt.Println("*** Error getting make:", err)
			//return v, err
		}
		c.Make = vehicleMake

		// get model
		model := clientmodels.Model{}

		query = `
			SELECT 
				id, 
				model, 
				vehicle_makes_id,
				created_at, 
				updated_at 
			FROM 
				vehicle_models 
			WHERE 
				id = ?`
		modelRow := m.DB.QueryRowContext(ctx, query, c.VehicleModelsID)

		err = modelRow.Scan(
			&model.ID,
			&model.Model,
			&model.MakeID,
			&model.CreatedAt,
			&model.UpdatedAt)
		if err != nil {
			fmt.Println("*** Error getting model:", err)
			//return v, err
		}
		c.Model = model

		// get options
		query = `
			select 
				vo.id, 
				vo.vehicle_id,
				vo.option_id,
				vo.created_at,
				vo.updated_at,
				o.option_name
			from 
				vehicle_options vo
				left join options o on (vo.option_id = o.id)
			where
				vo.vehicle_id = ?
				and o.active = 1
			order by 
				o.option_name`
		oRows, err := m.DB.QueryContext(ctx, query, c.ID)
		if err != nil {
			fmt.Println("*** Error getting options:", err)
		}

		var vehicleOptions []*clientmodels.VehicleOption
		for oRows.Next() {
			o := &clientmodels.VehicleOption{}
			err = oRows.Scan(
				&o.ID,
				&o.VehicleID,
				&o.OptionID,
				&o.CreatedAt,
				&o.UpdatedAt,
				&o.OptionName,
			)

			if err != nil {
				fmt.Println(err)
			} else {
				vehicleOptions = append(vehicleOptions, o)
			}
		}
		c.VehicleOptions = vehicleOptions
		oRows.Close()

		// get images
		query = `
			select 
				id, 
				vehicle_id,
				image,
				created_at,
				updated_at,
				sort_order
			from 
				vehicle_images 
			where
				vehicle_id = ?
			order by 
				sort_order`
		iRows, err := m.DB.QueryContext(ctx, query, c.ID)
		if err != nil {
			iRows.Close()
			fmt.Println(err)
		}

		var vehicleImages []*clientmodels.Image
		for iRows.Next() {
			o := &clientmodels.Image{}
			err = iRows.Scan(
				&o.ID,
				&o.VehicleID,
				&o.Image,
				&o.CreatedAt,
				&o.UpdatedAt,
				&o.SortOrder,
			)

			if err != nil {
				fmt.Println(err)
			} else {
				vehicleImages = append(vehicleImages, o)
			}
		}
		c.Images = vehicleImages
		iRows.Close()

		current := *c
		v = append(v, current)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return v, nil
}

// GetVehiclesForSaleByType returns slice of vehicles by type
func (m *DBModel) GetVehiclesForSaleByType(vehicleType int) ([]clientmodels.Vehicle, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var v []clientmodels.Vehicle

	query := `
		select 
		       id, 
		       stock_no, 
		       coalesce(cost, 0),
		       vin, 
		       coalesce(odometer, 0),
		       coalesce(year, 0),
		       coalesce(trim, ''),
		       vehicle_type,
		       coalesce(body, ''),
		       coalesce(seating_capacity,''),
		       coalesce(drive_train,''),
		       coalesce(engine,''),
		       coalesce(exterior_color,''),
		       coalesce(interior_color,''),
		       coalesce(transmission,''),
		       coalesce(options,''),
		       coalesce(model_number, ''),
		       coalesce(total_msr,0.0),
		       v.status,
		       coalesce(description, ''),
		       vehicle_makes_id,
		       vehicle_models_id,
		       hand_picked,
		       used,
		       coalesce(price_for_display,''),
		       created_at,
		       updated_at
		from 
		     vehicles v 
		where
			vehicle_type = ?
			and status = 1

		order by year desc`

	rows, err := m.DB.QueryContext(ctx, query, vehicleType)
	if err != nil {
		rows.Close()
		fmt.Println(err)
		return v, err
	}
	defer rows.Close()

	for rows.Next() {
		c := &clientmodels.Vehicle{}
		err = rows.Scan(
			&c.ID,
			&c.StockNo,
			&c.Cost,
			&c.Vin,
			&c.Odometer,
			&c.Year,
			&c.Trim,
			&c.VehicleType,
			&c.Body,
			&c.SeatingCapacity,
			&c.DriveTrain,
			&c.Engine,
			&c.ExteriorColour,
			&c.InteriorColour,
			&c.Transmission,
			&c.Options,
			&c.ModelNumber,
			&c.TotalMSR,
			&c.Status,
			&c.Description,
			&c.VehicleMakesID,
			&c.VehicleModelsID,
			&c.HandPicked,
			&c.Used,
			&c.PriceForDisplay,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
		if err != nil {
			fmt.Println(err)
			return v, err
		}

		// get make
		vehicleMake := clientmodels.Make{}

		query = `
			SELECT 
				id, 
				make, 
				created_at, 
				updated_at 
			FROM 
				vehicle_makes 
			WHERE 
				id = ?`
		makeRow := m.DB.QueryRowContext(ctx, query, c.VehicleMakesID)

		err = makeRow.Scan(
			&vehicleMake.ID,
			&vehicleMake.Make,
			&vehicleMake.CreatedAt,
			&vehicleMake.UpdatedAt)
		if err != nil {
			fmt.Println("*** Error getting make:", err)
			//return v, err
		}
		c.Make = vehicleMake

		// get model
		model := clientmodels.Model{}

		query = `
			SELECT 
				id, 
				model, 
				vehicle_makes_id,
				created_at, 
				updated_at 
			FROM 
				vehicle_models 
			WHERE 
				id = ?`
		modelRow := m.DB.QueryRowContext(ctx, query, c.VehicleModelsID)

		err = modelRow.Scan(
			&model.ID,
			&model.Model,
			&model.MakeID,
			&model.CreatedAt,
			&model.UpdatedAt)
		if err != nil {
			fmt.Println("*** Error getting model:", err)
			//return v, err
		}
		c.Model = model

		// get options
		query = `
			select 
				vo.id, 
				vo.vehicle_id,
				vo.option_id,
				vo.created_at,
				vo.updated_at,
				o.option_name
			from 
				vehicle_options vo
				left join options o on (vo.option_id = o.id)
			where
				vo.vehicle_id = ?
				and o.active = 1
			order by 
				o.option_name`
		oRows, err := m.DB.QueryContext(ctx, query, c.ID)
		if err != nil {
			fmt.Println("*** Error getting options:", err)
		}

		var vehicleOptions []*clientmodels.VehicleOption
		for oRows.Next() {
			o := &clientmodels.VehicleOption{}
			err = oRows.Scan(
				&o.ID,
				&o.VehicleID,
				&o.OptionID,
				&o.CreatedAt,
				&o.UpdatedAt,
				&o.OptionName,
			)

			if err != nil {
				fmt.Println(err)
			} else {
				vehicleOptions = append(vehicleOptions, o)
			}
		}
		c.VehicleOptions = vehicleOptions
		oRows.Close()

		// get images
		query = `
			select 
				id, 
				vehicle_id,
				image,
				created_at,
				updated_at,
				sort_order
			from 
				vehicle_images 
			where
				vehicle_id = ?
			order by 
				sort_order`
		iRows, err := m.DB.QueryContext(ctx, query, c.ID)
		if err != nil {
			iRows.Close()
			fmt.Println(err)
		}

		var vehicleImages []*clientmodels.Image
		for iRows.Next() {
			o := &clientmodels.Image{}
			err = iRows.Scan(
				&o.ID,
				&o.VehicleID,
				&o.Image,
				&o.CreatedAt,
				&o.UpdatedAt,
				&o.SortOrder,
			)

			if err != nil {
				fmt.Println(err)
			} else {
				vehicleImages = append(vehicleImages, o)
			}
		}
		c.Images = vehicleImages
		iRows.Close()

		current := *c
		v = append(v, current)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return v, nil
}

// AllVehiclesPaginated returns paginated slice of vehicles, by type
func (m *DBModel) AllVehiclesPaginated(vehicleTypeID, perPage, offset, year, make, model, price int) ([]clientmodels.Vehicle, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var v []clientmodels.Vehicle

	where := ""
	orderBy := "order by year desc"
	if year > 0 {
		where = fmt.Sprintf("and v.year = %d", year)
	}

	if make > 0 {
		where = fmt.Sprintf("%s and v.vehicle_makes_id = %d", where, make)
	}

	if model > 0 {
		where = fmt.Sprintf("%s and v.vehicle_models_id = %d", where, model)
	}

	if price == 1 {
		orderBy = "order by v.cost asc"
	} else if price == 2 {
		orderBy = "order by v.cost desc"
	}

	stmt := ""
	var nRows *sql.Row

	if vehicleTypeID < 1000 {
		stmt = fmt.Sprintf(`
		select 
			count(v.id) 
		from 
			vehicles v 
		where 
			status = 1 
			and vehicle_type = ? %s`, where)
		nRows = m.DB.QueryRowContext(ctx, stmt, vehicleTypeID)
	} else if vehicleTypeID == 1000 {
		stmt = fmt.Sprintf(`
		select 
			count(v.id) 
		from 
			vehicles v 
		where 
			status = 1 
			and vehicle_type in (8, 11, 12, 16, 13, 10, 7, 9, 15, 17, 14) %s 
			and v.used = 1`, where)
		nRows = m.DB.QueryRowContext(ctx, stmt)
	} else if vehicleTypeID == 1001 {
		stmt = fmt.Sprintf(`
		select 
			count(v.id) 
		from 
			vehicles v 
		where 
			status = 1 
			and vehicle_type in (13, 10, 9, 15) %s 
			and v.used = 1`, where)
		nRows = m.DB.QueryRowContext(ctx, stmt)
	}

	var num int
	err := nRows.Scan(&num)
	if err != nil {
		fmt.Println(err)
	}

	query := ""
	var rows *sql.Rows

	if vehicleTypeID < 1000 {
		query = fmt.Sprintf(`
		select 
		       id, 
		       stock_no, 
		       coalesce(cost, 0),
		       vin, 
		       coalesce(odometer, 0),
		       coalesce(year, 0),
		       coalesce(trim, ''),
		       vehicle_type,
		       coalesce(body, ''),
		       coalesce(seating_capacity,''),
		       coalesce(drive_train,''),
		       coalesce(engine,''),
		       coalesce(exterior_color,''),
		       coalesce(interior_color,''),
		       coalesce(transmission,''),
		       coalesce(options,''),
		       coalesce(model_number, ''),
		       coalesce(total_msr,0.0),
		       v.status,
		       coalesce(description, ''),
		       vehicle_makes_id,
		       vehicle_models_id,
		       hand_picked,
		       used,
		       coalesce(price_for_display,''),
		       created_at,
		       updated_at
		from 
		     vehicles v 
		where
			vehicle_type = ?
			and status = 1
			%s
			%s
		limit ? offset ?`, where, orderBy)
		rows, err = m.DB.QueryContext(ctx, query, vehicleTypeID, perPage, offset)
		if err != nil {
			fmt.Println(err)
			return nil, 0, err
		}

	} else if vehicleTypeID == 1000 {
		query = fmt.Sprintf(`
		select 
		       id, 
		       stock_no, 
		       coalesce(cost, 0),
		       vin, 
		       coalesce(odometer, 0),
		       coalesce(year, 0),
		       coalesce(trim, ''),
		       vehicle_type,
		       coalesce(body, ''),
		       coalesce(seating_capacity,''),
		       coalesce(drive_train,''),
		       coalesce(engine,''),
		       coalesce(exterior_color,''),
		       coalesce(interior_color,''),
		       coalesce(transmission,''),
		       coalesce(options,''),
		       coalesce(model_number, ''),
		       coalesce(total_msr,0.0),
		       v.status,
		       coalesce(description, ''),
		       vehicle_makes_id,
		       vehicle_models_id,
		       hand_picked,
		       used,
		       coalesce(price_for_display,''),
		       created_at,
		       updated_at
		from 
		     vehicles v 
		where
			vehicle_type in (8, 11, 12, 16, 7, 17, 14)
			and status = 1
			and v.used = 1
			%s
			%s
		limit ? offset ?`, where, orderBy)

		rows, err = m.DB.QueryContext(ctx, query, perPage, offset)
		if err != nil {
			fmt.Println(err)
			rows.Close()
			return nil, 0, err
		}
	} else if vehicleTypeID == 1001 {
		query = fmt.Sprintf(`
		select 
		       id, 
		       stock_no, 
		       coalesce(cost, 0),
		       vin, 
		       coalesce(odometer, 0),
		       coalesce(year, 0),
		       coalesce(trim, ''),
		       vehicle_type,
		       coalesce(body, ''),
		       coalesce(seating_capacity,''),
		       coalesce(drive_train,''),
		       coalesce(engine,''),
		       coalesce(exterior_color,''),
		       coalesce(interior_color,''),
		       coalesce(transmission,''),
		       coalesce(options,''),
		       coalesce(model_number, ''),
		       coalesce(total_msr,0.0),
		       v.status,
		       coalesce(description, ''),
		       vehicle_makes_id,
		       vehicle_models_id,
		       hand_picked,
		       used,
		       coalesce(price_for_display,''),
		       created_at,
		       updated_at
		from 
		     vehicles v 
		where
			vehicle_type in (13, 10, 9, 15)
			and status = 1
			and v.used = 1
			%s
			%s
		limit ? offset ?`, where, orderBy)

		rows, err = m.DB.QueryContext(ctx, query, perPage, offset)
		if err != nil {
			fmt.Println(err)
			rows.Close()
			return nil, 0, err
		}
	}

	defer rows.Close()

	for rows.Next() {
		c := &clientmodels.Vehicle{}
		err = rows.Scan(
			&c.ID,
			&c.StockNo,
			&c.Cost,
			&c.Vin,
			&c.Odometer,
			&c.Year,
			&c.Trim,
			&c.VehicleType,
			&c.Body,
			&c.SeatingCapacity,
			&c.DriveTrain,
			&c.Engine,
			&c.ExteriorColour,
			&c.InteriorColour,
			&c.Transmission,
			&c.Options,
			&c.ModelNumber,
			&c.TotalMSR,
			&c.Status,
			&c.Description,
			&c.VehicleMakesID,
			&c.VehicleModelsID,
			&c.HandPicked,
			&c.Used,
			&c.PriceForDisplay,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
		if err != nil {
			fmt.Println(err)
			return nil, 0, err
		}

		// get make
		vehicleMake := clientmodels.Make{}

		query = `
			SELECT 
				id, 
				make, 
				created_at, 
				updated_at 
			FROM 
				vehicle_makes 
			WHERE 
				id = ?`
		makeRow := m.DB.QueryRowContext(ctx, query, c.VehicleMakesID)

		err = makeRow.Scan(
			&vehicleMake.ID,
			&vehicleMake.Make,
			&vehicleMake.CreatedAt,
			&vehicleMake.UpdatedAt)
		if err != nil {
			fmt.Println("*** Error getting make:", err)
			//return v, err
		}
		c.Make = vehicleMake

		// get model
		model := clientmodels.Model{}

		query = `
			SELECT 
				id, 
				model, 
				vehicle_makes_id,
				created_at, 
				updated_at 
			FROM 
				vehicle_models 
			WHERE 
				id = ?`
		modelRow := m.DB.QueryRowContext(ctx, query, c.VehicleModelsID)

		err = modelRow.Scan(
			&model.ID,
			&model.Model,
			&model.MakeID,
			&model.CreatedAt,
			&model.UpdatedAt)
		if err != nil {
			fmt.Println("*** Error getting model:", err)
			//return v, err
		}
		c.Model = model

		// get options
		query = `
			select 
				vo.id, 
				vo.vehicle_id,
				vo.option_id,
				vo.created_at,
				vo.updated_at,
				o.option_name
			from 
				vehicle_options vo
				left join options o on (vo.option_id = o.id)
			where
				vo.vehicle_id = ?
				and o.active = 1
			order by 
				o.option_name`
		oRows, err := m.DB.QueryContext(ctx, query, c.ID)
		if err != nil {
			oRows.Close()
			fmt.Println("*** Error getting options:", err)
		}

		var vehicleOptions []*clientmodels.VehicleOption
		for oRows.Next() {
			o := &clientmodels.VehicleOption{}
			err = oRows.Scan(
				&o.ID,
				&o.VehicleID,
				&o.OptionID,
				&o.CreatedAt,
				&o.UpdatedAt,
				&o.OptionName,
			)

			if err != nil {
				fmt.Println(err)
				oRows.Close()
			} else {
				vehicleOptions = append(vehicleOptions, o)
			}
		}
		c.VehicleOptions = vehicleOptions
		oRows.Close()

		// get images
		query = `
			select 
				id, 
				vehicle_id,
				image,
				created_at,
				updated_at,
				sort_order
			from 
				vehicle_images 
			where
				vehicle_id = ?
			order by 
				sort_order`
		iRows, err := m.DB.QueryContext(ctx, query, c.ID)
		if err != nil {
			iRows.Close()
			fmt.Println(err)
		}

		var vehicleImages []*clientmodels.Image
		for iRows.Next() {
			o := &clientmodels.Image{}
			err = iRows.Scan(
				&o.ID,
				&o.VehicleID,
				&o.Image,
				&o.CreatedAt,
				&o.UpdatedAt,
				&o.SortOrder,
			)

			if err != nil {
				fmt.Println(err)
				iRows.Close()
			} else {
				vehicleImages = append(vehicleImages, o)
			}
		}
		c.Images = vehicleImages
		iRows.Close()

		current := *c
		v = append(v, current)
	}

	if err = rows.Err(); err != nil {
		return nil, num, err
	}

	return v, num, nil
}

// GetYearsForVehicleType gets years for vehicle type
func (m *DBModel) GetYearsForVehicleType(id int) ([]int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var years []int
	query := `
			select distinct 
				v.year
			from 
				vehicles v
			where
				vehicle_type = ?
				and v.status = 1
			order by 
				year desc`
	rows, err := m.DB.QueryContext(ctx, query, id)
	if err != nil {
		fmt.Println(err)
	}

	defer rows.Close()

	for rows.Next() {
		var y int
		err = rows.Scan(&y)
		if err != nil {
			fmt.Println(err)
		}
		years = append(years, y)
	}

	if err = rows.Err(); err != nil {
		return years, err
	}
	return years, nil
}

// GetMakes gets makes
func (m *DBModel) GetMakes() ([]clientmodels.Make, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var makes []clientmodels.Make
	query := `
			select  
				m.id, m.make
			from 
				vehicle_makes m
			order by m.make`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		rows.Close()
		fmt.Println(err)
	}

	defer rows.Close()

	for rows.Next() {
		var y clientmodels.Make
		err = rows.Scan(
			&y.ID,
			&y.Make)
		if err != nil {
			fmt.Println(err)
		}
		makes = append(makes, y)
	}

	if err = rows.Err(); err != nil {
		return makes, err
	}

	return makes, nil
}

// GetMakesForVehicleType gets makes for vehicle type
func (m *DBModel) GetMakesForVehicleType(id int) ([]clientmodels.Make, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var makes []clientmodels.Make
	query := ""

	if id < 1000 {
		query = `
			select  
				m.id, m.make
			from 
				vehicle_makes m
			where
				m.id in (select v.vehicle_makes_id from vehicles v where status = 1 and vehicle_type = ?)
			order by 
				m.make`
	} else if id == 1000 {
		query = `
			select  
				m.id, m.make
			from 
				vehicle_makes m
			where
				m.id in (select v.vehicle_makes_id from vehicles v where status = 1 and used = 1 and vehicle_type  in
				(8, 11, 12, 16, 7, 17, 14, ?)
)
			order by 
				m.make`
	} else if id == 1001 {
		query = `
			select  
				m.id, m.make
			from 
				vehicle_makes m
			where
				m.id in (select v.vehicle_makes_id from vehicles v where status = 1 and used = 1 and vehicle_type  in
				(13, 10, 9, 15, ?)
)
			order by 
				m.make`
	}
	rows, err := m.DB.QueryContext(ctx, query, id)
	if err != nil {
		rows.Close()
		fmt.Println(err)
	}

	defer rows.Close()

	for rows.Next() {
		var y clientmodels.Make
		err = rows.Scan(
			&y.ID,
			&y.Make)
		if err != nil {
			fmt.Println(err)
		}
		makes = append(makes, y)
	}

	if err = rows.Err(); err != nil {
		return makes, err
	}

	return makes, nil
}

// GetModelsForVehicleType gets models for vehicle type
func (m *DBModel) GetModelsForVehicleType(id int) ([]clientmodels.Model, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var models []clientmodels.Model
	query := `
			select  
				m.id, m.model
			from 
				vehicle_models m
			where
				m.id in (select v.vehicle_models_id from vehicles v where status = 1 and vehicle_type = ?)
			order by 
				m.model`
	rows, err := m.DB.QueryContext(ctx, query, id)
	if err != nil {
		rows.Close()
		fmt.Println(err)
	}

	defer rows.Close()

	for rows.Next() {
		var y clientmodels.Model
		err = rows.Scan(
			&y.ID,
			&y.Model)
		if err != nil {
			fmt.Println(err)
		}
		models = append(models, y)
	}

	if err = rows.Err(); err != nil {
		return models, err
	}

	return models, nil
}

// GetModelsForMakeID gets models for vehicle type
func (m *DBModel) GetModelsForMakeID(id int) ([]clientmodels.Model, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var models []clientmodels.Model
	query := `
			select  
				m.id, m.model
			from 
				vehicle_models m
			where
				m.vehicle_makes_id = ?
			order by 
				m.model`
	rows, err := m.DB.QueryContext(ctx, query, id)
	if err != nil {
		rows.Close()
		fmt.Println(err)
	}

	defer rows.Close()

	for rows.Next() {
		var y clientmodels.Model
		err = rows.Scan(
			&y.ID,
			&y.Model)
		if err != nil {
			fmt.Println(err)
		}
		models = append(models, y)
	}

	if err = rows.Err(); err != nil {
		return models, err
	}

	return models, nil
}

// GetVehicleByID gets a complete record for a vehicle
func (m *DBModel) GetVehicleByID(id int) (clientmodels.Vehicle, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var c clientmodels.Vehicle

	query := `
		select 
		       id, 
		       stock_no, 
		       coalesce(cost, 0),
		       vin, 
		       coalesce(odometer, 0),
		       coalesce(year, 0),
		       coalesce(trim, ''),
		       vehicle_type,
		       coalesce(body, ''),
		       coalesce(seating_capacity,''),
		       coalesce(drive_train,''),
		       coalesce(engine,''),
		       coalesce(exterior_color,''),
		       coalesce(interior_color,''),
		       coalesce(transmission,''),
		       coalesce(options,''),
		       coalesce(model_number, ''),
		       coalesce(total_msr,0.0),
		       v.status,
		       coalesce(description, ''),
		       vehicle_makes_id,
		       vehicle_models_id,
		       hand_picked,
		       used,
		       coalesce(price_for_display,''),
		       created_at,
		       updated_at
		from 
		     vehicles v 
		where
			id = ?`

	row := m.DB.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&c.ID,
		&c.StockNo,
		&c.Cost,
		&c.Vin,
		&c.Odometer,
		&c.Year,
		&c.Trim,
		&c.VehicleType,
		&c.Body,
		&c.SeatingCapacity,
		&c.DriveTrain,
		&c.Engine,
		&c.ExteriorColour,
		&c.InteriorColour,
		&c.Transmission,
		&c.Options,
		&c.ModelNumber,
		&c.TotalMSR,
		&c.Status,
		&c.Description,
		&c.VehicleMakesID,
		&c.VehicleModelsID,
		&c.HandPicked,
		&c.Used,
		&c.PriceForDisplay,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		fmt.Println(err)
		return c, err
	}

	// get make
	vehicleMake := clientmodels.Make{}

	query = `
			SELECT 
				id, 
				make, 
				created_at, 
				updated_at 
			FROM 
				vehicle_makes 
			WHERE 
				id = ?`
	makeRow := m.DB.QueryRowContext(ctx, query, c.VehicleMakesID)

	err = makeRow.Scan(
		&vehicleMake.ID,
		&vehicleMake.Make,
		&vehicleMake.CreatedAt,
		&vehicleMake.UpdatedAt)
	if err != nil {
		fmt.Println("*** Error getting make:", err)
	}
	c.Make = vehicleMake

	// get model
	model := clientmodels.Model{}

	query = `
			SELECT 
				id, 
				model, 
				vehicle_makes_id,
				created_at, 
				updated_at 
			FROM 
				vehicle_models 
			WHERE 
				id = ?`
	modelRow := m.DB.QueryRowContext(ctx, query, c.VehicleModelsID)

	err = modelRow.Scan(
		&model.ID,
		&model.Model,
		&model.MakeID,
		&model.CreatedAt,
		&model.UpdatedAt)
	if err != nil {
		fmt.Println("*** Error getting model:", err)
	}
	c.Model = model

	// get options
	query = `
			select 
				vo.id, 
				vo.vehicle_id,
				vo.option_id,
				vo.created_at,
				vo.updated_at,
				o.option_name
			from 
				vehicle_options vo
				left join options o on (vo.option_id = o.id)
			where
				vo.vehicle_id = ?
				and o.active = 1
			order by 
				o.option_name`
	oRows, err := m.DB.QueryContext(ctx, query, c.ID)
	if err != nil {
		fmt.Println("*** Error getting options:", err)
	}

	var vehicleOptions []*clientmodels.VehicleOption
	for oRows.Next() {
		o := &clientmodels.VehicleOption{}
		err = oRows.Scan(
			&o.ID,
			&o.VehicleID,
			&o.OptionID,
			&o.CreatedAt,
			&o.UpdatedAt,
			&o.OptionName,
		)

		if err != nil {
			fmt.Println(err)
		} else {
			vehicleOptions = append(vehicleOptions, o)
		}
	}

	c.VehicleOptions = vehicleOptions

	oRows.Close()

	// get images
	query = `
			select 
				id, 
				vehicle_id,
				image,
				created_at,
				updated_at,
				sort_order
			from 
				vehicle_images 
			where
				vehicle_id = ?
			order by 
				sort_order`
	iRows, err := m.DB.QueryContext(ctx, query, c.ID)
	if err != nil {
		fmt.Println(err)
		iRows.Close()
	}

	defer iRows.Close()

	var vehicleImages []*clientmodels.Image
	for iRows.Next() {
		o := &clientmodels.Image{}
		err = iRows.Scan(
			&o.ID,
			&o.VehicleID,
			&o.Image,
			&o.CreatedAt,
			&o.UpdatedAt,
			&o.SortOrder,
		)

		if err != nil {
			fmt.Println(err)
		} else {
			vehicleImages = append(vehicleImages, o)
		}
	}
	c.Images = vehicleImages
	iRows.Close()

	// get video, if any
	query = `
			select 
				v.id, 
				v.video_name,
				v.file_name,
				v.thumb
			from 
				vehicle_videos vv
				left join videos v on (vv.video_id = v.id)
			where
				vv.vehicle_id = ?
			limit 1`
	vRow := m.DB.QueryRowContext(ctx, query, c.ID)

	var vehicleVideo clientmodels.Video

	err = vRow.Scan(
		&vehicleVideo.ID,
		&vehicleVideo.VideoName,
		&vehicleVideo.FileName,
		&vehicleVideo.Thumb,
	)

	if err == nil {
		c.Video = vehicleVideo
	}

	c.Images = vehicleImages
	return c, nil
}

// GetSales gets max six sales people
func (m *DBModel) GetSales() ([]clientmodels.SalesStaff, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var s []clientmodels.SalesStaff

	query := `
		select 
		       id, 
		       coalesce(salesperson_name, ''),
		       coalesce(slug, ''),
		       coalesce(email, ''),
		       coalesce(phone, ''),
		       coalesce(image, '')
		from 
		     sales 
		where
			active = 1


		order by RAND() limit 6`

	rows, err := m.DB.QueryContext(ctx, query)

	if err != nil {
		fmt.Println(err)
		rows.Close()
		return s, err
	}
	defer rows.Close()

	for rows.Next() {
		c := &clientmodels.SalesStaff{}
		err = rows.Scan(
			&c.ID,
			&c.Name,
			&c.Slug,
			&c.Email,
			&c.Phone,
			&c.Image,
		)
		if err != nil {
			fmt.Println(err)
			return s, err
		}
		staff := *c
		s = append(s, staff)
	}
	return s, nil
}

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

// InsertTestDrive saves a test drive application
func (m *DBModel) InsertTestDrive(a clientmodels.TestDrive) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
	INSERT INTO test_drives (users_name, email, phone, preferred_date, preferred_time, vehicle_id, 
	                   processed, created_at, updated_at)
    VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)
    `

	_, err := m.DB.ExecContext(ctx, stmt,
		a.UsersName,
		a.Email,
		a.Phone,
		a.PreferredDate,
		a.PreferredTime,
		a.VehicleID,
		a.Processed,
		a.CreatedAt,
		a.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

// InsertQuickQuote saves a quick quote to remote dataabase
func (m *DBModel) InsertQuickQuote(a clientmodels.QuickQuote) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
	INSERT INTO quick_quotes (users_name, email, phone, vehicle_id, 
	                   processed, created_at, updated_at)
    VALUES(?, ?, ?, ?, ?, ?, ?)
    `

	_, err := m.DB.ExecContext(ctx, stmt,
		a.UsersName,
		a.Email,
		a.Phone,
		a.VehicleID,
		a.Processed,
		a.CreatedAt,
		a.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

// GetAllVehiclesForSale returns slice of all Vehicles for sale
func (m *DBModel) GetAllVehiclesForSale() ([]clientmodels.Vehicle, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var v []clientmodels.Vehicle

	query := `
		select 
		       id, 
		       stock_no, 
		       coalesce(cost, 0),
		       vin, 
		       coalesce(odometer, 0),
		       coalesce(year, 0),
		       coalesce(trim, ''),
		       vehicle_type,
		       coalesce(body, ''),
		       coalesce(seating_capacity,''),
		       coalesce(drive_train,''),
		       coalesce(engine,''),
		       coalesce(exterior_color,''),
		       coalesce(interior_color,''),
		       coalesce(transmission,''),
		       coalesce(options,''),
		       coalesce(model_number, ''),
		       coalesce(total_msr,0.0),
		       v.status,
		       coalesce(description, ''),
		       vehicle_makes_id,
		       vehicle_models_id,
		       hand_picked,
		       used,
		       coalesce(price_for_display,''),
		       created_at,
		       updated_at, 
		       case when vehicle_type in (8, 11, 12) then 'ATV'
		       when vehicle_type = 1 then 'Car'
		       when vehicle_type = 16 then 'Electric Bike'
		       when vehicle_type = 13 then 'Jetski'
		       when vehicle_type = 10 then 'Outboard Motor'
		       when vehicle_type = 7 then 'Motorcycle'
		       when vehicle_type = 9 then 'Pontoon Boat'
		       when vehicle_type = 15 then 'Power Boat'
		       when vehicle_type = 17 then 'Scooter'
		       when vehicle_type = 5 then 'SUV'
		       when vehicle_type = 14 then 'Trailer'
		       when vehicle_type = 2 then 'Truck'
		       when vehicle_type = 4 then 'Other'
		       when vehicle_type = 6 then 'Van'
		       else 'Other'
		       end as vehicle_type_string 
		from 
		     vehicles v 
		where
			v.status = 1
			and v.vehicle_models_id is not null 
			and v.vehicle_makes_id is not null
		
		order by stock_no`

	rows, err := m.DB.QueryContext(ctx, query)

	if err != nil {
		fmt.Println(err)
		return v, err
	}
	defer rows.Close()

	for rows.Next() {
		c := &clientmodels.Vehicle{}
		err = rows.Scan(
			&c.ID,
			&c.StockNo,
			&c.Cost,
			&c.Vin,
			&c.Odometer,
			&c.Year,
			&c.Trim,
			&c.VehicleType,
			&c.Body,
			&c.SeatingCapacity,
			&c.DriveTrain,
			&c.Engine,
			&c.ExteriorColour,
			&c.InteriorColour,
			&c.Transmission,
			&c.Options,
			&c.ModelNumber,
			&c.TotalMSR,
			&c.Status,
			&c.Description,
			&c.VehicleMakesID,
			&c.VehicleModelsID,
			&c.HandPicked,
			&c.Used,
			&c.PriceForDisplay,
			&c.CreatedAt,
			&c.UpdatedAt,
			&c.VehicleTypeString,
		)
		if err != nil {
			fmt.Println(err)
			return v, err
		}

		// get make
		vehicleMake := clientmodels.Make{}

		query = `
			SELECT 
				id, 
				make, 
				created_at, 
				updated_at 
			FROM 
				vehicle_makes 
			WHERE 
				id = ?`
		makeRow := m.DB.QueryRowContext(ctx, query, c.VehicleMakesID)

		err = makeRow.Scan(
			&vehicleMake.ID,
			&vehicleMake.Make,
			&vehicleMake.CreatedAt,
			&vehicleMake.UpdatedAt)
		if err != nil {
			fmt.Println("*** Error getting make:", err)
			//return v, err
		}
		c.Make = vehicleMake
		c.VehicleMake = vehicleMake.Make

		// get model
		model := clientmodels.Model{}

		query = `
			SELECT 
				id, 
				model, 
				vehicle_makes_id,
				created_at, 
				updated_at 
			FROM 
				vehicle_models 
			WHERE 
				id = ?`
		modelRow := m.DB.QueryRowContext(ctx, query, c.VehicleModelsID)

		err = modelRow.Scan(
			&model.ID,
			&model.Model,
			&model.MakeID,
			&model.CreatedAt,
			&model.UpdatedAt)
		if err != nil {
			fmt.Println("*** Error getting model:", err)
			//return v, err
		}
		c.Model = model
		c.VehicleModel = model.Model

		// get images
		query = `
			select 
				id, 
				vehicle_id,
				image,
				created_at,
				updated_at,
				sort_order
			from 
				vehicle_images 
			where
				vehicle_id = ?
			order by 
				sort_order asc`
		iRows, err := m.DB.QueryContext(ctx, query, c.ID)
		if err != nil {
			iRows.Close()
			fmt.Println(err)
		}

		var vehicleImages []*clientmodels.Image

		for iRows.Next() {
			o := &clientmodels.Image{}
			err = iRows.Scan(
				&o.ID,
				&o.VehicleID,
				&o.Image,
				&o.CreatedAt,
				&o.UpdatedAt,
				&o.SortOrder,
			)

			ph := fmt.Sprintf("https://www.ca/storage/inventory/%d/%s", c.ID, o.Image)
			o.Image = ph

			if err != nil {
				fmt.Println(err)
			} else {
				vehicleImages = append(vehicleImages, o)
			}

		}
		iRows.Close()

		c.Images = vehicleImages
		iRows.Close()

		current := *c

		v = append(v, current)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return v, nil
}
