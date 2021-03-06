package clientdb

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/huandu/go-sqlbuilder"
	"github.com/microcosm-cc/bluemonday"
	"github.com/tsawler/goblender/client/clienthandlers/clientmodels"
	"github.com/tsawler/goblender/pkg/models"
	"html/template"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// VehicleModel holds the db connection
type DBModel struct {
	DB *sql.DB
}

var stripTags = bluemonday.StrictPolicy()

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
	} else {
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
	} else {
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

	where := "vehicle_type < 7"
	if id > 0 {
		where = fmt.Sprintf("vehicle_type = %d", id)
	}

	var years []int
	query := fmt.Sprintf(`
			select distinct 
				v.year
			from 
				vehicles v
			where
				%s
				and v.status = 1
			order by 
				year desc`, where)
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

	where := "and vehicle_type < 7"
	if id > 0 {
		where = fmt.Sprintf("and vehicle_type = %d", id)
	}

	var makes []clientmodels.Make
	query := ""

	query = fmt.Sprintf(`
		select  
			m.id, m.make
		from 
			vehicle_makes m
		where
			m.id in (select v.vehicle_makes_id from vehicles v where status = 1 %s)
		order by 
			m.make`, where)

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

// GetAllVehiclesForWindowStickers returns slice of all Vehicles for sale
func (m *DBModel) GetAllVehiclesForWindowStickers() ([]clientmodels.Vehicle, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var v []clientmodels.Vehicle

	query := `
select 
		       v.id, 
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
		       v.vehicle_makes_id,
		       v.vehicle_models_id,
		       hand_picked,
		       used,
		       coalesce(price_for_display,''),
		       v.created_at,
		       v.updated_at, 
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
		       end as vehicle_type_string,
		       	 vm.make,
		       	 vmod.model
		from 
		     vehicles v 
		      left join vehicle_makes vm on (v.vehicle_makes_id = vm.id)
		      left join vehicle_models vmod on (v.vehicle_models_id = vmod.id)
		where
			v.status = 1
			and v.vehicle_models_id is not null 
			and v.vehicle_makes_id is not null
			and v.vehicle_type < 7
	   order by vm.make, vmod.model
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
			&c.VehicleTypeString,
			&c.VehicleMake,
			&c.VehicleModel,
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
func (m *DBModel) ModelsForMakeID(id, vehicleTypeID int) ([]clientmodels.Model, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	where := "and vehicle_type < 7"
	if vehicleTypeID > 0 {
		where = fmt.Sprintf("and vehicle_type = %d", vehicleTypeID)
	}

	var models []clientmodels.Model
	query := fmt.Sprintf(`select id, model from vehicle_models where vehicle_makes_id = ? and 
		id in (select vehicle_models_id from vehicles where status = 1 %s)
		order by model`, where)
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
func (m *DBModel) MakesForYear(year, vehicleTypeID int) ([]clientmodels.Make, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	where := "and vehicle_type < 7"
	if vehicleTypeID > 0 {
		where = fmt.Sprintf("and vehicle_type = %d", vehicleTypeID)
	}

	var makes []clientmodels.Make
	query := fmt.Sprintf(`select id, make from vehicle_makes 
		where id in (select vehicle_makes_id from vehicles where status = 1 and year = ? %s)
		order by make`, where)
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

// GetTeam gets the team for public
func (m *DBModel) GetStaffForPublic() ([]clientmodels.Employee, error) {
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

// GetOneSalesStaffBySlug gets one sales staff by slug
func (m *DBModel) GetOneSalesStaffBySlug(slug string) (clientmodels.SalesStaff, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var o clientmodels.SalesStaff

	query := `select id, salesperson_name, slug , coalesce(image, ''), coalesce(email, ''), 
		coalesce(phone, ''), active, created_at, updated_at from sales where slug = ?`

	row := m.DB.QueryRowContext(ctx, query, slug)
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
		// Append it to the slice

		w = append(w, *s)
	}

	if err = prows.Err(); err != nil {
		return nil, 0, err
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

	if err = prows.Err(); err != nil {
		return nil, 0, err
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

// CarGurus returns slice for car guru csv
func (m *DBModel) CarGurus() ([][]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var r [][]string

	headers := []string{
		"dealer_id",
		"dealer_name",
		"address",
		"dealer_state",
		"dealer_zip",
		"dealer_phone_number",
		"dealer_crm_email",
		"vin",
		"make",
		"model",
		"year",
		"trim",
		"option",
		"price",
		"msrp",
		"certified",
		"mileage",
		"dealer_comments",
		"stock_number",
		"transmission",
		"image_urls",
		"main_image",
		"exterior_color",
		"dealer_website_url",
	}

	r = append(r, headers)

	query := `
	select
		437564 as dealer_id,
		'Jim Gilbert\'s Wheels and Deals' as dealer_name,
		'402 St. Mary\'s Street, Fredericton NB' as address,
		'New Brunswick' as dealer_state,
		'E3A 8H5' as dealer_zip,
		'5064596832' as dealer_phone_number,
		'salesmanager@wheelsanddeals.ca' as dealer_crm_email,
		v.vin,
		vm.make,
		vmod.model,
		v.year,
		v.trim,
		(select group_concat(o.option_name SEPARATOR ', ') from options o where o.id in (select option_id from vehicle_options where vehicle_id = v.id)) as options,
		v.cost as price,
		v.total_msr as msrp,
		1 as certified,
		v.odometer as mileage,
		REPLACE(v.description, '&nbsp;', ' ') as dealer_comments,
		stock_no as stock_number,
		v.transmission,
	case
	when (select count(id) from vehicle_images vi where vi.vehicle_id = v.id) = 0 then ''
		else
		(select GROUP_CONCAT(CONCAT('https://www.wheelsanddeals.ca/storage/inventory/', v.id, '/',vimages.image) SEPARATOR ',') from vehicle_images vimages where vimages.vehicle_id = v.id order by sort_order)
			end as image_urls,
		case
		when (select count(id) from vehicle_images vi where vi.vehicle_id = v.id) = 0 then 'https://www.wheelsanddeals.ca/vendor/wheelspackage/hug-in-progress.jpg'
			else
			(select concat('https://www.wheelsanddeals.ca/storage/inventory/',v.id,'/',vimages.image) as main_image from vehicle_images vimages where vimages.vehicle_id = v.id and sort_order = 1 limit 1)
				end as main_image,
					v.exterior_color,
					'https://www.wheelsanddeals.ca' as dealer_website_url

				from vehicles v
				left join vehicle_makes vm on (v.vehicle_makes_id = vm.id)
				left join vehicle_models vmod on (v.vehicle_models_id = vmod.id)

				where v.status = 1 and v.vehicle_type < 7`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return r, err
	}

	defer rows.Close()

	for rows.Next() {
		var current []string
		var id, dealderName, address, state, zip, phone, email, vin, vMake, vModel, year, trim, options string
		var price, msrp float32
		var certified, odometer int
		var stock, comments, transmission, images, mainImage, exterior, website string

		err = rows.Scan(
			&id,
			&dealderName,
			&address,
			&state,
			&zip,
			&phone,
			&email,
			&vin,
			&vMake,
			&vModel,
			&year,
			&trim,
			&options,
			&price,
			&msrp,
			&certified,
			&odometer,
			&comments,
			&stock,
			&transmission,
			&images,
			&mainImage,
			&exterior,
			&website,
		)
		current = append(current, id)
		current = append(current, address)
		current = append(current, state)
		current = append(current, zip)
		current = append(current, phone)
		current = append(current, email)
		current = append(current, vin)
		current = append(current, vMake)
		current = append(current, vModel)
		current = append(current, year)
		current = append(current, trim)
		current = append(current, options)
		current = append(current, fmt.Sprintf("%.2f", price))
		current = append(current, fmt.Sprintf("%.2f", msrp))
		current = append(current, fmt.Sprintf("%d", certified))
		current = append(current, fmt.Sprintf("%d", odometer))
		strippedComments := stripTags.Sanitize(comments)
		current = append(current, strippedComments)
		current = append(current, stock)
		current = append(current, transmission)
		current = append(current, images)
		current = append(current, mainImage)
		current = append(current, exterior)
		current = append(current, website)

		r = append(r, current)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return r, nil
}

// Kijiji gets a slice of records for CSV feed
func (m *DBModel) Kijiji() ([][]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var r [][]string

	headers := []string{
		"dealer_id",
		"dealer_name",
		"address",
		"phone",
		"postalcode",
		"email",
		"id",
		"vin",
		"stockid",
		"is_used",
		"is_certified",
		"year",
		"make",
		"model",
		"body",
		"trim",
		"transmission",
		"kilometers",
		"exterior_color",
		"price",
		"model_code",
		"comments",
		"drivetrain",
		"video_url",
		"images",
		"category",
	}

	r = append(r, headers)

	query := `
	select
		86450547 as dealer_id,
		'Jim Gilbert\'s Wheels and Deals' as dealer_name,
		'402 St. Mary\'s Street, Fredericton NB' as address,
		'5064596832' as phone,
		'E3A 8H5' as postalcode,
		'salesmanager@wheelsanddeals.ca' as email,
		v.id, vin, stock_no as stockid, used as is_used, 1 as is_certified, year,
		vm.make, vmod.model, v.body, v.trim, v.transmission, v.odometer as kilometers,
		v.exterior_color, v.cost as price, '' as model_code,
		REPLACE(v.description, '&nbsp;', ' ') as comments,
		v.drive_train as drivetrain, '' as video_url,
	case
	when (select count(id) from vehicle_images vi where vi.vehicle_id = v.id) = 0 then ''
		else
		(select GROUP_CONCAT(CONCAT('https://www.wheelsanddeals.ca/storage/inventory/', v.id, '/',vimages.image) SEPARATOR '|') from vehicle_images vimages where vimages.vehicle_id = v.id order by sort_order)
			end as images,
				0 as category

			from vehicles v
			left join vehicle_makes vm on (v.vehicle_makes_id = vm.id)
			left join vehicle_models vmod on (v.vehicle_models_id = vmod.id)

			where v.status = 1 and v.vehicle_type < 7
			`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return r, err
	}

	defer rows.Close()

	for rows.Next() {
		var current []string
		var id, dealerID, used, certified, year, km, category int
		var dealerName, address, phone, postalCode, email, vin, stockID, vMake, vModel string
		var body, trim, comments, transmission, exteriorColor, modelCode, driveTrain, videoURL, images string
		var price float32

		err = rows.Scan(
			&dealerID,
			&dealerName,
			&address,
			&phone,
			&postalCode,
			&email,
			&id,
			&vin,
			&stockID,
			&used,
			&certified,
			&year,
			&vMake,
			&vModel,
			&body,
			&trim,
			&transmission,
			&km,
			&exteriorColor,
			&price,
			&modelCode,
			&comments,
			&driveTrain,
			&videoURL,
			&images,
			&category,
		)

		current = append(current, strconv.Itoa(dealerID))
		current = append(current, dealerName)
		current = append(current, address)
		current = append(current, phone)
		current = append(current, postalCode)
		current = append(current, email)
		current = append(current, strconv.Itoa(id))
		current = append(current, vin)
		current = append(current, stockID)
		current = append(current, strconv.Itoa(used))
		current = append(current, strconv.Itoa(certified))
		current = append(current, strconv.Itoa(year))
		current = append(current, vMake)
		current = append(current, vModel)
		current = append(current, body)
		current = append(current, trim)
		current = append(current, transmission)
		current = append(current, strconv.Itoa(km))
		current = append(current, exteriorColor)
		current = append(current, fmt.Sprintf("%.2f", price))
		current = append(current, modelCode)
		// remove tags
		strippedComments := stripTags.Sanitize(comments)
		current = append(current, strippedComments)
		current = append(current, driveTrain)
		current = append(current, videoURL)
		current = append(current, images)
		current = append(current, strconv.Itoa(category))

		r = append(r, current)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return r, nil
}

// KijijiPS gets a slice of records for Kijiji Powersports CSV feed
func (m *DBModel) KijijiPS() ([][]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var r [][]string

	headers := []string{
		"dealer_id",
		"dealer_name",
		"address",
		"phone",
		"postalcode",
		"email",
		"id",
		"vin",
		"stockid",
		"engine",
		"is_used",
		"year",
		"make",
		"model",
		"body",
		"trim",
		"transmission",
		"kilometers",
		"exterior_color",
		"price",
		"model_code",
		"comments",
		"drivetrain",
		"video_url",
		"images",
		"category",
	}

	r = append(r, headers)

	query := `
	select
            86450547 as dealer_id, 
            'Jim Gilbert\'s Wheels and Deals' as dealer_name,
            '402 St. Mary\'s Street, Fredericton NB' as address,
            '5064596832' as phone,
            'E3A 8H5' as postalcode,
            'salesmanager@wheelsanddeals.ca' as email,
            v.id, vin, stock_no as stockid, engine, used as is_used,  year,
            vm.make, vmod.model,  v.body, v.trim, v.transmission, v.odometer as kilometers,
            v.exterior_color, v.cost as price, '' as model_code,
            REPLACE(v.description, '&nbsp;', ' ') as comments, 
            v.drive_train as drivetrain, '' as video_url,
            case 
            when (select count(id) from vehicle_images vi where vi.vehicle_id = v.id) = 0 then ''
            else
           	(select GROUP_CONCAT(CONCAT('https://www.wheelsanddeals.ca/storage/inventory/', v.id, '/',vimages.image) SEPARATOR '|') from vehicle_images vimages where vimages.vehicle_id = v.id order by sort_order) 
            end as images,
            case 
            when vehicle_type = 7 and v.vehicle_models_id in (225, 380, 341)  then 303
            when vehicle_type = 7 and v.vehicle_models_id = 223 then 304
            when vehicle_type = 7 and v.vehicle_models_id in (342,228,227,232) then 307
            when vehicle_type = 7 and v.vehicle_models_id not in(225, 380, 341, 223,342,228,227,232) then 306
            
            when vehicle_type = 8 then 311
            when vehicle_type = 11 then 311
            when vehicle_type = 12 then 311
            when vehicle_type = 13 then 330
            when vehicle_type = 15 then 327
            when vehicle_type = 10 then 327
            when vehicle_type = 15 then 327
            when vehicle_type = 9 then 327
            when vehicle_type = 16 then 308
            when vehicle_type = 17 then 308
            
            end as category
        
        from vehicles v
        left join vehicle_makes vm on (v.vehicle_makes_id = vm.id)
        left join vehicle_models vmod on (v.vehicle_models_id = vmod.id)
        
        where v.status = 1 and v.vehicle_type > 6 and v.vehicle_type <> 14
			`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return r, err
	}

	defer rows.Close()

	for rows.Next() {
		var current []string
		var id, dealerID, used, year, km, category int
		var dealerName, engine, address, phone, postalCode, email, vin, stockID, vMake, vModel string
		var body, trim, comments, transmission, exteriorColor, modelCode, driveTrain, videoURL, images string
		var price float32

		err = rows.Scan(
			&dealerID,
			&dealerName,
			&address,
			&phone,
			&postalCode,
			&email,
			&id,
			&vin,
			&stockID,
			&engine,
			&used,
			&year,
			&vMake,
			&vModel,
			&body,
			&trim,
			&transmission,
			&km,
			&exteriorColor,
			&price,
			&modelCode,
			&comments,
			&driveTrain,
			&videoURL,
			&images,
			&category,
		)

		current = append(current, strconv.Itoa(dealerID))
		current = append(current, dealerName)
		current = append(current, address)
		current = append(current, phone)
		current = append(current, postalCode)
		current = append(current, email)
		current = append(current, strconv.Itoa(id))
		current = append(current, vin)
		current = append(current, stockID)
		current = append(current, engine)
		current = append(current, strconv.Itoa(used))
		current = append(current, strconv.Itoa(year))
		current = append(current, vMake)
		current = append(current, vModel)
		current = append(current, body)
		current = append(current, trim)
		current = append(current, transmission)
		current = append(current, strconv.Itoa(km))
		current = append(current, exteriorColor)
		current = append(current, fmt.Sprintf("%.2f", price))
		current = append(current, modelCode)
		// strip html
		strippedComments := stripTags.Sanitize(comments)
		current = append(current, strippedComments)
		current = append(current, driveTrain)
		current = append(current, videoURL)
		current = append(current, images)
		current = append(current, strconv.Itoa(category))

		r = append(r, current)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return r, nil
}

type OldPost struct {
	ID          int
	UserID      int
	BlogID      int
	Title       string
	Content     string
	Thumbnail   string
	Slug        string
	Meta        string
	KeyWords    string
	PostDate    time.Time
	Active      int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Preview     string
	AccessLevel int
}

// postContentStruct is a struct used to create json content for pages
type postContentStruct struct {
	Title   map[string]string        `json:"title"`
	Content map[string]template.HTML `json:"content"`
	Preview map[string]template.HTML `json:"preview"`
	Search  map[string]template.HTML `json:"search"`
}

func (m *DBModel) CopyPosts() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	stripTags := bluemonday.StrictPolicy()

	stmt := "SELECT id, language, code, active, created_at, updated_at FROM languages ORDER BY language"
	lRows, err := m.DB.QueryContext(ctx, stmt)
	if err != nil {
		return err
	}
	defer lRows.Close()

	languages := []*models.Language{}
	for lRows.Next() {
		s := &models.Language{}
		err = lRows.Scan(&s.ID, &s.Language, &s.Code, &s.Active, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			return err
		}
		// Append it to the slice
		languages = append(languages, s)
	}

	query := `select id, 1 as user_id, 1 as blog_id, title, content, coalesce(thumbnail, ''), slug, coalesce(meta, ''), 
			coalesce(keywords, ''),
 			post_date, active, created_at, updated_at, coalesce(preview, ''), access_level from wheelsanddeals.posts order by id`
	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var old OldPost
		err = rows.Scan(
			&old.ID,
			&old.UserID,
			&old.BlogID,
			&old.Title,
			&old.Content,
			&old.Thumbnail,
			&old.Slug,
			&old.Meta,
			&old.KeyWords,
			&old.PostDate,
			&old.Active,
			&old.CreatedAt,
			&old.UpdatedAt,
			&old.Preview,
			&old.AccessLevel,
		)

		fmt.Println(old.ID, old.Title)

		var newContent postContentStruct

		p := models.Post{
			ID:              old.ID,
			UserID:          old.UserID,
			BlogID:          old.BlogID,
			Title:           old.Title,
			Content:         "",
			Thumbnail:       old.Thumbnail,
			Slug:            old.Slug,
			Meta:            old.Meta,
			Keywords:        old.KeyWords,
			PostDate:        old.PostDate,
			Active:          old.Active,
			JS:              "",
			CSS:             "",
			MenuColor:       "navbar-light",
			MenuTransparent: 0,
			PageStyles:      "",
			AccessLevel:     old.AccessLevel,
			CreatedAt:       old.CreatedAt,
			UpdatedAt:       old.UpdatedAt,
			SEOImage:        0,
		}

		// insert new post - first build maps
		titleMap := make(map[string]string)
		contentMap := make(map[string]template.HTML)
		previewMap := make(map[string]template.HTML)
		searchMap := make(map[string]template.HTML)

		for _, lang := range languages {
			newHTML := strings.Replace(old.Content, "<blogarchives></blogarchives>", "</div>", -1)
			newHTML = strings.Replace(newHTML, `<div class="col-md-9" data-gramm="false">`, `<div class="row"><div class="col-md-9">`, 1)
			contentMap[lang.Code] = template.HTML(newHTML)
			titleMap[lang.Code] = old.Title
			previewMap[lang.Code] = ""
			searchMap[lang.Code] = template.HTML(stripTags.Sanitize(old.Content))
		}

		newContent.Title = titleMap
		newContent.Content = contentMap
		newContent.Preview = previewMap
		newContent.Search = searchMap
		contentJson, err := json.MarshalIndent(newContent, "", "    ")
		if err != nil {
			fmt.Println(err)
			return err
		}
		p.Content = string(contentJson)

		// insert
		var postStruct = sqlbuilder.NewStruct(new(models.Post))
		ib := postStruct.InsertInto("posts", p)
		pquery, args := ib.Build()
		_, err = m.DB.ExecContext(ctx, pquery, args...)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	return nil
}

// DeleteUnusedVideos deletes unused videos/thumbnails
func (m *DBModel) DeleteUnusedVideos() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	var ids []int

	query := `select id from vehicles where status = 0 and vehicle_type < 7 
				and id in (select vehicle_id from vehicle_videos)`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var i int
		err = rows.Scan(&i)
		if err != nil {
			return err
		}
		ids = append(ids, i)
	}

	type Item struct {
		VehicleVideoID int
		VideoID        int
		FileName       string
		Thumb          string
	}

	for _, x := range ids {

		stmt := `select vv.id, v.id as video_id, v.file_name, v.thumb
					from vehicle_videos vv 
					left join videos v on (vv.video_id = v.id)
					where vv.vehicle_id = ?`
		vRows, err := m.DB.QueryContext(ctx, stmt, x)
		if err != nil {
			return err
		}
		defer vRows.Close()

		for vRows.Next() {
			var current Item
			err = vRows.Scan(
				&current.VehicleVideoID,
				&current.VideoID,
				&current.FileName,
				&current.Thumb,
			)
			if err != nil {
				return err
			}

			// delete video files
			path := fmt.Sprintf("./ui/static/site-content/videos/%s.mp4", current.FileName)
			err := os.Remove(path)
			if err != nil {
				fmt.Println(err)
			}

			// delete thumb
			path = fmt.Sprintf("./ui/static/site-content/videos/%s", current.Thumb)
			err = os.Remove(path)
			if err != nil {
				fmt.Println(err)
			}

			// update db
			stmt := `delete from vehicle_videos where vehicle_id = ?`
			_, err = m.DB.ExecContext(ctx, stmt, x)
			if err != nil {
				fmt.Println(err)
			}
		}

	}

	return nil
}

// DeleteUnusedInventoryImages deletes images/panoramas for sold vehicles
func (m *DBModel) DeleteUnusedInventoryImages() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var ids []int

	query := `select id from vehicles where status = 0`
	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var i int
		err = rows.Scan(&i)
		if err != nil {
			return err
		}
		ids = append(ids, i)
	}

	for _, x := range ids {
		stmt := `delete from vehicle_images where vehicle_id = ?`
		_, err := m.DB.ExecContext(ctx, stmt, x)
		if err != nil {
			fmt.Println(err)
		}

		// delete vehicle images, if any
		path := fmt.Sprintf("./ui/static/site-content/inventory/%d", x)
		err = os.RemoveAll(path)
		if err != nil {
			fmt.Println(err)
		}

		// delete panorama, if any
		stmt = `delete from vehicle_panoramas where vehicle_id = ?`
		_, err = m.DB.ExecContext(ctx, stmt, x)
		if err != nil {
			fmt.Println(err)
		}

		files, err := filepath.Glob(fmt.Sprintf("./ui/static/site-content/panoramas%d-*", x))
		if err != nil {
			fmt.Println(err)
		}
		for _, f := range files {
			if err := os.Remove(f); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}
