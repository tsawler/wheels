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
		if strings.Contains(query, " lower(") {
			query = strings.Replace(query, "WHERE", "WHERE (", 1)
			query = strings.Replace(query, "ORDER BY", fmt.Sprintf(") and %s ORDER BY", extra[0]), 1)
		} else {
			query = strings.Replace(query, "ORDER BY", fmt.Sprintf("where true and %s ORDER BY", extra[0]), 1)
		}

		baseQuery = strings.Replace(baseQuery, "WHERE", "WHERE true and", 1)

		if strings.Contains(baseQuery, " lower(") {
			baseQuery = fmt.Sprintf("%s and %s", baseQuery, extra[0])
		} else {
			baseQuery = fmt.Sprintf("%s where true and %s", baseQuery, extra[0])
		}

		where = extra[0]
	}

	// count all rows
	allRows, err := m.DB.QueryContext(ctx, fmt.Sprintf("select count(id) as all_rows from v_all_vehicles where true and %s", where))
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
func (m *DBModel) AllVehiclesPaginated(vehicleTypeID, perPage, offset, year, make, model, price int, handPicked bool) ([]clientmodels.Vehicle, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	extraWhere := ""
	if handPicked {
		extraWhere = " and hand_picked = 1"
	}
	var v []clientmodels.Vehicle

	where := ""
	orderBy := "order by year desc"
	if year > 0 {
		where = fmt.Sprintf("and v.year = %d", year)
	}

	if make > 0 {
		where = fmt.Sprintf("%s and v.vehicle_makes_id = %d %s", where, make, extraWhere)
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

	if vehicleTypeID == 0 {
		stmt = fmt.Sprintf(`
		select 
			count(v.id) 
		from 
			vehicles v 
		where 
			status = 1 
			and vehicle_type < 7 %s %s`, where, extraWhere)
		nRows = m.DB.QueryRowContext(ctx, stmt)
	} else if vehicleTypeID < 1000 {
		// suvs
		stmt = fmt.Sprintf(`
		select 
			count(v.id) 
		from 
			vehicles v 
		where 
			status = 1 
			and vehicle_type = ? %s %s`, where, extraWhere)
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

	if vehicleTypeID == 0 {
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
			vehicle_type < 7
			and status = 1
			%s
			%s
			%s
		limit ? offset ?`, where, extraWhere, orderBy)
		rows, err = m.DB.QueryContext(ctx, query, perPage, offset)
		if err != nil {
			fmt.Println(err)
			return nil, 0, err
		}
	} else if vehicleTypeID < 1000 {
		// suvs
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
			%s
		limit ? offset ?`, where, extraWhere, orderBy)
		rows, err = m.DB.QueryContext(ctx, query, vehicleTypeID, perPage, offset)
		if err != nil {
			fmt.Println(err)
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
				vehicle_type < 7
				and v.status = 1
			order by 
				year desc`
	rows, err := m.DB.QueryContext(ctx, query)
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

	query = `
		select  
			m.id, m.make
		from 
			vehicle_makes m
		where
			m.id in (select v.vehicle_makes_id from vehicles v where status = 1 and vehicle_type < 7)
		order by 
			m.make`

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
				vv.id,
				vv.video_id,
				coalesce(v.video_name, ''),
				coalesce(v.file_name, ''),
				coalesce(v.thumb, ''),
				coalesce(v.is_360, 0),
				coalesce(v.duration, 0),
				coalesce(vv.created_at, now()),
				coalesce(vv.updated_at, now())
			from 
				vehicle_videos vv
				left join videos v on (vv.video_id = v.id)
			where
				vv.vehicle_id = ?
			limit 1`
	vRow := m.DB.QueryRowContext(ctx, query, id)

	var vehicleVideo clientmodels.VehicleVideo

	err = vRow.Scan(
		&vehicleVideo.ID,
		&vehicleVideo.VideoID,
		&vehicleVideo.VideoName,
		&vehicleVideo.FileName,
		&vehicleVideo.Thumb,
		&vehicleVideo.Is360,
		&vehicleVideo.Duration,
		&vehicleVideo.CreatedAt,
		&vehicleVideo.UpdatedAt,
	)

	if err == nil {
		c.Video = vehicleVideo
	} else {
		fmt.Println(err)
	}

	// get panorama, if any
	query = `select id, vehicle_id, panorama, created_at, updated_at 
		from vehicle_panoramas where vehicle_id = ?`

	panRow := m.DB.QueryRowContext(ctx, query, id)
	var pan clientmodels.Panorama
	err = panRow.Scan(&pan.ID, &pan.VehicleID, &pan.Panorama, &pan.CreatedAt, &pan.UpdatedAt)

	if err == nil {
		c.Panorama = pan
	} else {
		fmt.Println(err)
	}

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

// CheckIfVehicleExists checks to see if we have a vehicle, by stock number
func (m *DBModel) CheckIfVehicleExists(stockNumber string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	count := 0
	query := `
		select 
		       count(id) as counter
		from 
		     vehicles v 
		where
			stock_no = ?`

	row := m.DB.QueryRowContext(ctx, query, stockNumber)
	err := row.Scan(
		&count,
	)
	if err != nil {
		fmt.Println(err)
		return false
	}

	if count > 0 {
		return true
	}
	return false
}

// InsertVehicle inserts a vehicle
func (m *DBModel) InsertVehicle(v clientmodels.Vehicle) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		INSERT INTO vehicles (stock_no, cost, vin, odometer, year, trim, vehicle_type, 
			body, seating_capacity, drive_train, engine, exterior_color, interior_color,
			transmission, options, model_number, total_msr, status, description, vehicle_makes_id,
			vehicle_models_id, hand_picked, used, price_for_display, created_at, updated_at)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `

	result, err := m.DB.ExecContext(ctx,
		stmt,
		v.StockNo,
		v.Cost,
		v.Vin,
		v.Odometer,
		v.Year,
		v.Trim,
		v.VehicleType,
		v.Body,
		v.SeatingCapacity,
		v.DriveTrain,
		v.Engine,
		v.ExteriorColour,
		v.InteriorColour,
		v.Transmission,
		v.Options,
		v.ModelNumber,
		v.TotalMSR,
		v.Status,
		v.Description,
		v.VehicleMakesID,
		v.VehicleModelsID,
		v.HandPicked,
		v.Used,
		v.PriceForDisplay,
		v.CreatedAt,
		v.UpdatedAt,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// UpdateVehicle updates a vehicle in the database
func (m *DBModel) UpdateVehicle(v clientmodels.Vehicle) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `update vehicles set 
		stock_no = ?,
		cost = ?,
		vin = ?,
		odometer = ?,
		year = ?,
		trim = ?,
		vehicle_type = ?,
		body = ?,
		seating_capacity = ?,
		drive_train = ?,
		engine = ?,
		exterior_color = ?,
		interior_color = ?,
		transmission = ?,
		options = ?,
		model_number = ?,
		total_msr = ?,
		status = ?,
		description = ?,
		vehicle_makes_id = ?,
		vehicle_models_id = ?,
		hand_picked = ?,
		used = ?,
		price_for_display = ?,
		updated_at = ?
		where id = ?`
	_, err := m.DB.ExecContext(ctx, query,
		v.StockNo,
		v.Cost,
		v.Vin,
		v.Odometer,
		v.Year,
		v.Trim,
		v.VehicleType,
		v.Body,
		v.SeatingCapacity,
		v.DriveTrain,
		v.Engine,
		v.ExteriorColour,
		v.InteriorColour,
		v.Transmission,
		v.Options,
		v.ModelNumber,
		v.TotalMSR,
		v.Status,
		v.Description,
		v.VehicleMakesID,
		v.VehicleModelsID,
		v.HandPicked,
		v.Used,
		v.PriceForDisplay,
		v.UpdatedAt,
		v.ID,
	)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

// InsertVideoForVehicle inserts a video for a vehicle
func (m *DBModel) InsertVideoForVehicle(v clientmodels.VehicleVideo) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	stmt := `
		INSERT INTO vehicle_videos (vehicle_id, video_id, created_at, updated_at)
		VALUES(?, ?, ?, ?)
    `

	_, err := m.DB.ExecContext(ctx,
		stmt,
		v.VehicleID,
		v.VideoID,
		v.CreatedAt,
		v.UpdatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}

// UpdateVideoForVehicle updates video for a vehicle, or removes it entirely
func (m *DBModel) UpdateVideoForVehicle(v clientmodels.VehicleVideo) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if v.VideoID > 0 {
		query := `update vehicle_videos set 
		video_id = ?
		where vehicle_id = ?`
		_, err := m.DB.ExecContext(ctx, query, v.VideoID, v.VehicleID)
		if err != nil {
			fmt.Println(err)
			return err
		}
	} else {
		query := `delete from vehicle_videos where vehicle_id = ?`
		_, err := m.DB.ExecContext(ctx, query, v.VehicleID)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	return nil
}

// InsertVehicleImage inserts a vehicle image
func (m *DBModel) InsertVehicleImage(vi clientmodels.Image) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		INSERT INTO vehicle_images (vehicle_id, image, sort_order, created_at, updated_at)
		VALUES(?, ?, ?, ?, ?)
    `

	_, err := m.DB.ExecContext(ctx,
		stmt,
		vi.VehicleID,
		vi.Image,
		vi.SortOrder,
		vi.CreatedAt,
		vi.UpdatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}

// DeleteEvent deletes a vehicle image
func (m *DBModel) DeleteVehicleImage(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := "delete from vehicle_images where id = ?"
	_, err := m.DB.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}
	return nil
}

// GetVehicleImageByID gets one image by id
func (m *DBModel) GetVehicleImageByID(id int) (clientmodels.Image, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var i clientmodels.Image

	query := `select image, vehicle_id from vehicle_images where id = ?`

	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&i.Image,
		&i.VehicleID,
	)
	if err != nil {
		fmt.Println(err)
		return i, err
	}

	return i, nil
}

// DeleteAllVehicleOptions deletes all options
func (m *DBModel) DeleteAllVehicleOptions(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := "delete from vehicle_options where vehicle_id = ?"
	_, err := m.DB.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}
	return nil
}

// InsertVehicleOption inserts a vehicle option
func (m *DBModel) InsertVehicleOption(vo clientmodels.VehicleOption) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		INSERT INTO vehicle_options (vehicle_id, option_id, created_at, updated_at)
		VALUES(?, ?, ?, ?)
    `

	_, err := m.DB.ExecContext(ctx,
		stmt,
		vo.VehicleID,
		vo.OptionID,
		vo.CreatedAt,
		vo.UpdatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}

// GetMaxSortOrderForVehicleID gets one image by id
func (m *DBModel) GetMaxSortOrderForVehicleID(id int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	max := 0

	query := `select coalesce(max(sort_order), 0) from vehicle_images where vehicle_id = ?`

	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&max,
	)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}

	return max, nil
}

// UpdateSortOrderForImage updates sort order for image
func (m *DBModel) UpdateSortOrderForImage(id, order int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `update vehicle_images set 
		sort_order = ?
		where id = ?`
	_, err := m.DB.ExecContext(ctx, query, order, id)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

// ModelsForMakeID returns json for available models for specified make
func (m *DBModel) ModelsForMakeID(id int) ([]clientmodels.Model, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var models []clientmodels.Model
	query := `select id, model from vehicle_models where vehicle_makes_id = ? and 
		id in (select vehicle_models_id from vehicles where status = 1)
		order by model`
	rows, err := m.DB.QueryContext(ctx, query, id)
	if err != nil {
		fmt.Println(err)
		return models, err
	}

	defer rows.Close()

	for rows.Next() {
		var y clientmodels.Model
		err = rows.Scan(
			&y.ID,
			&y.Model,
		)
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

// ModelsForMakeID returns json for available models for specified make
func (m *DBModel) ModelsForMakeIDAdmin(id int) ([]clientmodels.Model, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var models []clientmodels.Model
	query := `select id, model from vehicle_models where vehicle_makes_id = ? 
		order by model`
	rows, err := m.DB.QueryContext(ctx, query, id)
	if err != nil {
		fmt.Println(err)
		return models, err
	}

	defer rows.Close()

	for rows.Next() {
		var y clientmodels.Model
		err = rows.Scan(
			&y.ID,
			&y.Model,
		)
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

// MakesForYear returns json for available makes for a specified year
func (m *DBModel) MakesForYear(year int) ([]clientmodels.Make, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var makes []clientmodels.Make
	query := `select id, make from vehicle_makes 
		where id in (select vehicle_makes_id from vehicles where status = 1 and year = ? and vehicle_type < 7)
		order by make`
	rows, err := m.DB.QueryContext(ctx, query, year)
	if err != nil {
		fmt.Println(err)
		return makes, err
	}

	defer rows.Close()

	for rows.Next() {
		var y clientmodels.Make
		err = rows.Scan(
			&y.ID,
			&y.Make,
		)
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

// CountSoldThisMonth counts sold this month
func (m *DBModel) CountSoldThisMonth() int {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var soldThisMonth int
	thisMonth := fmt.Sprintf("%d-%d-01", time.Now().Year(), time.Now().Month())

	query := fmt.Sprintf(`select count(id) from vehicles where status = 0 and updated_at >= '%s' and vehicle_type < 7`, thisMonth)
	row := m.DB.QueryRowContext(ctx, query)
	err := row.Scan(
		&soldThisMonth,
	)

	if err != nil {
		fmt.Println(err)
		return 0
	}

	return soldThisMonth
}

// CountPending counts pending
func (m *DBModel) CountPending() int {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var pending int

	query := "select count(id) from vehicles where status = 2"
	row := m.DB.QueryRowContext(ctx, query)
	err := row.Scan(
		&pending,
	)

	if err != nil {
		fmt.Println(err)
		return 0
	}

	return pending
}

// CountTradeIns counts trade ins
func (m *DBModel) CountTradeIns() int {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var pending int

	query := "select count(id) from vehicles where status = 3"
	row := m.DB.QueryRowContext(ctx, query)
	err := row.Scan(
		&pending,
	)

	if err != nil {
		fmt.Println(err)
		return 0
	}

	return pending
}

// CountForSale counts for sale
func (m *DBModel) CountForSale() int {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var forSale int

	query := "select count(id) from vehicles where status = 1 and vehicle_type < 7"
	row := m.DB.QueryRowContext(ctx, query)
	err := row.Scan(
		&forSale,
	)

	if err != nil {
		fmt.Println(err)
		return 0
	}

	return forSale
}

// CountForSalePowerSports counts for sale in power sports
func (m *DBModel) CountForSalePowerSports() int {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var forSale int

	query := "select count(id) from vehicles where status = 1 and vehicle_type >= 7"
	row := m.DB.QueryRowContext(ctx, query)
	err := row.Scan(
		&forSale,
	)

	if err != nil {
		fmt.Println(err)
		return 0
	}

	return forSale
}

// CountSoldThisMonthPowerSports counts sold power sports for thsis mounth
func (m *DBModel) CountSoldThisMonthPowerSports() int {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var soldThisMonth int
	thisMonth := fmt.Sprintf("%d-%d-01", time.Now().Year(), time.Now().Month())

	query := fmt.Sprintf(`select count(id) from vehicles where status = 0 and updated_at >= '%s' and vehicle_type >= 7`, thisMonth)
	row := m.DB.QueryRowContext(ctx, query)
	err := row.Scan(
		&soldThisMonth,
	)

	if err != nil {
		fmt.Println(err)
		return 0
	}

	return soldThisMonth
}

// GetOptions returns slice of vehicles by type
func (m *DBModel) GetOptions() ([]clientmodels.Option, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var v []clientmodels.Option

	query := `
		select 
		       id, 
		       option_name, 
		       active,
		       created_at,
		       updated_at
		from 
		     options 
		order by option_name`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return v, err
	}
	defer rows.Close()

	for rows.Next() {
		c := &clientmodels.Option{}
		err = rows.Scan(
			&c.ID,
			&c.OptionName,
			&c.Active,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
		if err != nil {
			fmt.Println(err)
			return v, err
		}
		v = append(v, *c)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return v, nil
}

// GetOneOption gets one option
func (m *DBModel) GetOneOption(id int) (clientmodels.Option, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var o clientmodels.Option

	query := "select id, option_name, active, created_at, updated_at from options where id = ?"
	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&o.ID,
		&o.OptionName,
		&o.Active,
		&o.CreatedAt,
		&o.UpdatedAt,
	)

	if err != nil {
		fmt.Println(err)
		return o, err
	}

	return o, nil
}

// UpdateOption updates one option
func (m *DBModel) UpdateOption(o clientmodels.Option) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `update options set 
		option_name = ?,
		active = ?,
		updated_at = ?
		where id = ?`

	_, err := m.DB.ExecContext(ctx, query, o.OptionName, o.Active, o.UpdatedAt, o.ID)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

// InsertOption inserts an option
func (m *DBModel) InsertOption(o clientmodels.Option) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
	INSERT INTO options (option_name, active, created_at, updated_at)
    VALUES(?, ?, ?, ?)`

	_, err := m.DB.ExecContext(ctx, stmt,
		o.OptionName,
		o.Active,
		o.CreatedAt,
		o.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

// GetStaffForSorting returns slice of staff
func (m *DBModel) GetStaffForSorting() ([]clientmodels.Employee, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var v []clientmodels.Employee

	query := `
		select 
		       id, 
		       first_name, 
		       last_name, 
		       coalesce(position, ''),
		       coalesce(email, ''),
		       coalesce(image, ''), 
		       coalesce(description, ''),
		       active,
		       sort_order,
		       created_at,
		       updated_at
		from 
		     employees 
		where active = 1
		order by sort_order`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return v, err
	}
	defer rows.Close()

	for rows.Next() {
		c := &clientmodels.Employee{}
		err = rows.Scan(
			&c.ID,
			&c.FirstName,
			&c.LastName,
			&c.Position,
			&c.Email,
			&c.Image,
			&c.Description,
			&c.Active,
			&c.SortOrder,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
		if err != nil {
			fmt.Println(err)
			return v, err
		}
		v = append(v, *c)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return v, nil
}

// GetStaff returns slice of staff
func (m *DBModel) GetStaff() ([]clientmodels.Employee, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var v []clientmodels.Employee

	query := `
		select 
		       id, 
		       first_name, 
		       last_name, 
		       coalesce(position, ''),
		       coalesce(email, ''),
		       coalesce(image, ''), 
		       coalesce(description, ''),
		       active,
		       sort_order,
		       created_at,
		       updated_at
		from 
		     employees 
		order by last_name`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return v, err
	}
	defer rows.Close()

	for rows.Next() {
		c := &clientmodels.Employee{}
		err = rows.Scan(
			&c.ID,
			&c.FirstName,
			&c.LastName,
			&c.Position,
			&c.Email,
			&c.Image,
			&c.Description,
			&c.Active,
			&c.SortOrder,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
		if err != nil {
			fmt.Println(err)
			return v, err
		}
		v = append(v, *c)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return v, nil
}

// UpdateStaff updates staff
func (m *DBModel) UpdateStaff(o clientmodels.Employee) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `update employees e set 
		e.first_name = ?,
		last_name = ?,
		e.position = ?,
		email = ?,
		image = ?,
		description = ?,
		active = ?,
		sort_order = ?,
		updated_at = ?
		where id = ?`

	_, err := m.DB.ExecContext(ctx, query,
		o.FirstName,
		o.LastName,
		o.Position,
		o.Email,
		o.Image,
		o.Description,
		o.Active,
		o.SortOrder,
		o.UpdatedAt,
		o.ID,
	)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

// GetOneStaff gets one staff
func (m *DBModel) GetOneStaff(id int) (clientmodels.Employee, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var o clientmodels.Employee

	query := `select id, first_name, last_name, coalesce(e.position, ''), coalesce(image, ''), coalesce(email, ''), 
		coalesce(description, ''), active, sort_order, created_at, updated_at from employees e where id = ?`

	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&o.ID,
		&o.FirstName,
		&o.LastName,
		&o.Position,
		&o.Image,
		&o.Email,
		&o.Description,
		&o.Active,
		&o.SortOrder,
		&o.CreatedAt,
		&o.UpdatedAt,
	)

	if err != nil {
		fmt.Println(err)
		return o, err
	}

	return o, nil
}

func (m *DBModel) InsertStaff(o clientmodels.Employee) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
	INSERT INTO employees (first_name, last_name, position, image, email, description, active, created_at, updated_at)
    VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := m.DB.ExecContext(ctx, stmt,
		o.FirstName,
		o.LastName,
		o.Position,
		o.Image,
		o.Email,
		o.Description,
		o.Active,
		o.CreatedAt,
		o.UpdatedAt,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

// GetSalesPeople returns slice of sales staff
func (m *DBModel) GetSalesPeople() ([]clientmodels.SalesStaff, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var v []clientmodels.SalesStaff

	query := `
		select 
		       id, 
		       salesperson_name, 
		       slug, 
		       coalesce(email, ''),
		       coalesce(image, ''), 
		       coalesce(phone, ''),
		       active,
		       created_at,
		       updated_at
		from 
		     sales 
		order by salesperson_name`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return v, err
	}
	defer rows.Close()

	for rows.Next() {
		c := &clientmodels.SalesStaff{}
		err = rows.Scan(
			&c.ID,
			&c.Name,
			&c.Slug,
			&c.Email,
			&c.Image,
			&c.Phone,
			&c.Active,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
		if err != nil {
			fmt.Println(err)
			return v, err
		}
		v = append(v, *c)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return v, nil
}

// GetOneSalesStaff gets one sales staff
func (m *DBModel) GetOneSalesStaff(id int) (clientmodels.SalesStaff, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var o clientmodels.SalesStaff

	query := `select id, salesperson_name, slug , coalesce(image, ''), coalesce(email, ''), 
		coalesce(phone, ''), active, created_at, updated_at from sales where id = ?`

	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&o.ID,
		&o.Name,
		&o.Slug,
		&o.Image,
		&o.Email,
		&o.Phone,
		&o.Active,
		&o.CreatedAt,
		&o.UpdatedAt,
	)

	if err != nil {
		fmt.Println(err)
		return o, err
	}

	return o, nil
}

// UpdateSalesStaff updates staff
func (m *DBModel) UpdateSalesStaff(o clientmodels.SalesStaff) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `update sales e set 
		e.salesperson_name = ?,
		slug = ?,
		phone = ?,
		image = ?,
		active = ?,
		updated_at = ?
		where id = ?`

	_, err := m.DB.ExecContext(ctx, query,
		o.Name,
		o.Slug,
		o.Phone,
		o.Image,
		o.Active,
		o.UpdatedAt,
		o.ID,
	)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (m *DBModel) InsertSalesStaff(o clientmodels.SalesStaff) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
	INSERT INTO sales (salesperson_name, slug, email, phone, active, created_at, updated_at)
    VALUES(?, ?, ?, ?, ?, ?, ?)`

	result, err := m.DB.ExecContext(ctx, stmt,
		o.Name,
		o.Slug,
		o.Email,
		o.Phone,
		o.Active,
		o.CreatedAt,
		o.UpdatedAt,
	)
	if err != nil {
		fmt.Println("error inserting", err)
		return 0, err
	}
	fmt.Println("getting new id")

	id, err := result.LastInsertId()
	fmt.Println("returning id of", int(id), "from repo")
	return int(id), nil
}

// DeleteSalesStaff deletes a sales person
func (m *DBModel) DeleteSalesStaff(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := "delete from sales where id = ?"
	_, err := m.DB.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}
	return nil
}

// DeleteStaff deletes a staff person
func (m *DBModel) DeleteStaff(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := "delete from employees where id = ?"
	_, err := m.DB.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}
	return nil
}

// UpdateSortOrderForStaff updates sort order for staff
func (m *DBModel) UpdateSortOrderForStaff(id, order int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `update employees set 
		sort_order = ?
		where id = ?`
	_, err := m.DB.ExecContext(ctx, query, order, id)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

// GetModelByName gets a model by name
func (m *DBModel) GetModelByName(s string) int {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select id from vehicle_models where upper(model) = ?`
	row := m.DB.QueryRowContext(ctx, query, strings.ToUpper(s))

	var modelID int

	err := row.Scan(&modelID)
	if err != nil {
		fmt.Println("*** Error getting make:", err)
		return 0
	}
	return modelID
}

// GetMakeByName gets a make by name
func (m *DBModel) GetMakeByName(s string) int {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select id from vehicle_makes where upper(make) = ?`
	row := m.DB.QueryRowContext(ctx, query, strings.ToUpper(s))

	var modelID int

	err := row.Scan(&modelID)
	if err != nil {
		fmt.Println("*** Error getting model:", err)
		return 0
	}
	return modelID
}

// InsertModel inserts a new model
func (m *DBModel) InsertModel(mid int, s string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `insert into vehicle_models (vehicle_makes_id, model, created_at, updated_at)
			values (?, ?, ?, ?)`

	result, err := m.DB.ExecContext(ctx, query, mid, s, time.Now(), time.Now())
	if err != nil {
		fmt.Print("Error inserting model", err)
		return 0, err
	}

	newID, err := result.LastInsertId()
	if err != nil {
		fmt.Print("Error inserting model", err)
		return 0, err
	}

	return int(newID), nil
}

// InsertMake inserts a new make
func (m *DBModel) InsertMake(s string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `insert into vehicle_makes (make, created_at, updated_at)
			values (?, ?, ?)`

	result, err := m.DB.ExecContext(ctx, query, s, time.Now(), time.Now())
	if err != nil {
		fmt.Print("Error inserting make", err)
		return 0, err
	}

	newID, err := result.LastInsertId()
	if err != nil {
		fmt.Print("Error inserting make", err)
		return 0, err
	}

	return int(newID), nil
}

// GetAllTestimonials returns slice of testimonials
func (m *DBModel) GetAllTestimonials() ([]clientmodels.Testimonial, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var v []clientmodels.Testimonial

	query := `
		select 
		       id, 
		       label, 
		       url, 
		       active,
		       created_at,
		       updated_at
		from 
		     testimonials 
		order by created_at desc`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return v, err
	}
	defer rows.Close()

	for rows.Next() {
		c := &clientmodels.Testimonial{}
		err = rows.Scan(
			&c.ID,
			&c.Label,
			&c.Url,
			&c.Active,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
		if err != nil {
			fmt.Println(err)
			return v, err
		}
		v = append(v, *c)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return v, nil
}

// GetOneTestimonial returns one testimonial
func (m *DBModel) GetOneTestimonial(id int) (clientmodels.Testimonial, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var o clientmodels.Testimonial

	query := "select id, label, url, active, created_at, updated_at from testimonials where id = ?"
	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&o.ID,
		&o.Label,
		&o.Url,
		&o.Active,
		&o.CreatedAt,
		&o.UpdatedAt,
	)

	if err != nil {
		fmt.Println(err)
		return o, err
	}

	return o, nil
}

// UpdateTestimonial updates a testimonial
func (m *DBModel) UpdateTestimonial(o clientmodels.Testimonial) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `update testimonials set 
		label = ?,
		url = ?,
		active = ?,
		updated_at = ?
		where id = ?`

	_, err := m.DB.ExecContext(ctx, query, o.Label, o.Url, o.Active, o.UpdatedAt, o.ID)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

// InsertTestimonial inserts a testimonial
func (m *DBModel) InsertTestimonial(o clientmodels.Testimonial) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
	INSERT INTO testimonials (label, url, active, created_at, updated_at)
    VALUES(?, ?, ?, ?, ?)`

	_, err := m.DB.ExecContext(ctx, stmt,
		o.Label,
		o.Url,
		o.Active,
		o.CreatedAt,
		o.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

// GetAllWordOfMouth returns slice of word of mouth
func (m *DBModel) GetAllWordOfMouth() ([]clientmodels.Word, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var v []clientmodels.Word

	query := `
		select 
		       id, 
		       title, 
		       content, 
		       active,
		       created_at,
		       updated_at
		from 
		     words 
		order by created_at desc`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return v, err
	}
	defer rows.Close()

	for rows.Next() {
		c := &clientmodels.Word{}
		err = rows.Scan(
			&c.ID,
			&c.Title,
			&c.Content,
			&c.Active,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
		if err != nil {
			fmt.Println(err)
			return v, err
		}
		v = append(v, *c)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return v, nil
}

// GetOneWordOfMouth returns one word of mouth
func (m *DBModel) GetOneWordOfMouth(id int) (clientmodels.Word, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var o clientmodels.Word

	query := "select id, title, content, active, created_at, updated_at from words where id = ?"
	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&o.ID,
		&o.Title,
		&o.Content,
		&o.Active,
		&o.CreatedAt,
		&o.UpdatedAt,
	)

	if err != nil {
		fmt.Println(err)
		return o, err
	}

	return o, nil
}

// UpdateWordOfMouth updates a word of mouth
func (m *DBModel) UpdateWordOfMouth(o clientmodels.Word) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `update words set 
		title = ?,
		content = ?,
		active = ?,
		updated_at = ?
		where id = ?`

	_, err := m.DB.ExecContext(ctx, query, o.Title, o.Content, o.Active, o.UpdatedAt, o.ID)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

// InsertWordOfMouth inserts a word of mouth
func (m *DBModel) InsertWordOfMouth(o clientmodels.Word) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
	INSERT INTO words (title, content, active, created_at, updated_at)
    VALUES(?, ?, ?, ?, ?)`

	_, err := m.DB.ExecContext(ctx, stmt,
		o.Title,
		o.Content,
		o.Active,
		o.CreatedAt,
		o.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

// AllWordOfMouthPaginated returns paginated words
func (m *DBModel) AllWordOfMouthPaginated(limit, offset int) ([]clientmodels.Word, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var w []clientmodels.Word

	stmt := "select count(id) from words p where p.active = 1"
	countRow := m.DB.QueryRowContext(ctx, stmt)

	var num int
	err := countRow.Scan(&num)
	if err != nil {
		fmt.Println(err)
	}

	stmt = `
		SELECT 
		p.id,
		p.title,
		p.content,
		p.active,
		p.created_at,
		p.updated_at
		FROM words p 
		where active = 1
		ORDER BY created_at desc
		limit ? offset ?
`
	prows, err := m.DB.QueryContext(ctx, stmt, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer prows.Close()

	for prows.Next() {
		s := &clientmodels.Word{}
		err = prows.Scan(
			&s.ID,
			&s.Title,
			&s.Content,
			&s.Active,
			&s.CreatedAt,
			&s.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		// Append it to the slice.

		w = append(w, *s)
	}

	return w, num, nil
}

// AllTestimonialsPaginated returns paginated list of testimonials
func (m *DBModel) AllTestimonialsPaginated(limit, offset int) ([]clientmodels.Testimonial, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var w []clientmodels.Testimonial

	stmt := "select count(id) from testimonials p where p.active = 1"
	countRow := m.DB.QueryRowContext(ctx, stmt)

	var num int
	err := countRow.Scan(&num)
	if err != nil {
		fmt.Println(err)
	}

	stmt = `
		SELECT 
		p.id,
		p.label,
		p.url,
		p.active,
		p.created_at,
		p.updated_at
		FROM testimonials p 
		where active = 1
		ORDER BY created_at desc
		limit ? offset ?
`
	prows, err := m.DB.QueryContext(ctx, stmt, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer prows.Close()

	for prows.Next() {
		s := &clientmodels.Testimonial{}
		err = prows.Scan(
			&s.ID,
			&s.Label,
			&s.Url,
			&s.Active,
			&s.CreatedAt,
			&s.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		// Append it to the slice.

		w = append(w, *s)
	}

	return w, num, nil
}

// UpdatePanorama updates a panorama
func (m *DBModel) UpdatePanorama(vp clientmodels.Panorama) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `delete from vehicle_panoramas where vehicle_id = ?`
	_, err := m.DB.ExecContext(ctx, query, vp.VehicleID)
	if err != nil {
		fmt.Println(err)
		return err
	}

	stmt := `
	INSERT INTO vehicle_panoramas (vehicle_id, panorama, created_at, updated_at)
    VALUES(?, ?, ?, ?)`

	_, err = m.DB.ExecContext(ctx, stmt,
		vp.VehicleID,
		vp.Panorama,
		vp.CreatedAt,
		vp.UpdatedAt,
	)

	return nil
}

// InsertFinder inserts a vehicle finder request
func (m *DBModel) InsertFinder(o clientmodels.Finder) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
	INSERT INTO finders (first_name, last_name, email, phone, contact_method, year, make, model, created_at, updated_at)
    VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := m.DB.ExecContext(ctx, stmt,
		o.FirstName,
		o.LastName,
		o.Email,
		o.Phone,
		o.ContactMethod,
		o.Year,
		o.Make,
		o.Model,
		o.CreatedAt,
		o.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

// GetOneFinder returns one finder
func (m *DBModel) GetOneFinder(id int) (clientmodels.Finder, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var o clientmodels.Finder

	query := `select 
		id,
		coalesce(first_name,  ''),
		coalesce(last_name,  ''),
		coalesce(email,  ''),
		coalesce(phone, ''), 
		coalesce(contact_method, ''), 
		coalesce(year, ''), 
		coalesce(make, ''), 
		coalesce(model, ''),
		created_at, 
		updated_at 
		from finders where id = ?`
	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&o.ID,
		&o.FirstName,
		&o.LastName,
		&o.Email,
		&o.Phone,
		&o.ContactMethod,
		&o.Year,
		&o.Make,
		&o.Model,
		&o.CreatedAt,
		&o.UpdatedAt,
	)

	if err != nil {
		fmt.Println(err)
		return o, err
	}

	return o, nil
}

// GetAllFinders returns slice of finders
func (m *DBModel) GetAllFinders() ([]clientmodels.Finder, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var v []clientmodels.Finder

	query := `
			select 
		       	id, 
		       	coalesce(first_name,  ''),
				coalesce(last_name,  ''),
				coalesce(email,  ''),
				coalesce(phone, ''), 
				coalesce(contact_method, ''), 
				coalesce(year, ''), 
				coalesce(make, ''), 
				coalesce(model, ''),
		       created_at,
		       updated_at
			from 
				 finders 
			order by created_at desc`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return v, err
	}
	defer rows.Close()

	for rows.Next() {
		c := &clientmodels.Finder{}
		err = rows.Scan(
			&c.ID,
			&c.FirstName,
			&c.LastName,
			&c.Email,
			&c.Phone,
			&c.ContactMethod,
			&c.Year,
			&c.Make,
			&c.Model,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
		if err != nil {
			fmt.Println(err)
			return v, err
		}
		v = append(v, *c)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return v, nil
}
