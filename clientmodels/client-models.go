package clientmodels

import "time"

type VehicleJSON struct {
	ID        int       `json:"id"`
	StockNo   string    `json:"stock_no"`
	Vin       string    `json:"vin"`
	Year      int       `json:"year"`
	Trim      string    `json:"trim"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Make      string    `json:"make"`
	Model     string    `json:"model"`
}

// Vehicle holds a vehicle
type Vehicle struct {
	Body              string           `xml:"-"`
	Cost              float32          `xml:"Price"`
	CreatedAt         time.Time        `xml:"-"`
	Description       string           `xml:"Description"`
	DriveTrain        string           `xml:"-"`
	Engine            string           `xml:"-"`
	ExteriorColour    string           `xml:"exterior_color"`
	HandPicked        int              `xml:"-"`
	ID                int              `xml:"-"`
	Images            []*Image         `xml:"Images"`
	InteriorColour    string           `xml:"interior_color"`
	Make              Make             `xml:"-"`
	Model             Model            `xml:"-"`
	ModelNumber       string           `xml:"-"`
	Odometer          int              `xml:"odometer"`
	Options           string           `xml:"-"`
	PriceForDisplay   string           `xml:"-"`
	SeatingCapacity   string           `xml:"-"`
	Status            int              `xml:"-"`
	StockNo           string           `xml:"StockNo"`
	TotalMSR          float32          `xml:"-"`
	Transmission      string           `xml:"-"`
	Trim              string           `xml:"-"`
	UpdatedAt         time.Time        `xml:"-"`
	Used              int              `xml:"-"`
	VehicleMake       string           `xml:"Make"`
	VehicleMakesID    int              `xml:"-"`
	VehicleModel      string           `xml:"Model"`
	VehicleModelsID   int              `xml:"-"`
	VehicleOptions    []*VehicleOption `xml:"-"`
	VehicleType       int              `xml:"-"`
	VehicleTypeString string           `xml:"vehicle_type"`
	Video             VehicleVideo     `xml:"-"`
	Vin               string           `xml:"Vin"`
	Year              int              `xml:"Year"`
}

// Option holds vehicle options
type Option struct {
	ID         int
	OptionName string
	Active     int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// VehicleOption holds option for a given vehicle
type VehicleOption struct {
	ID         int
	VehicleID  int
	OptionID   int
	CreatedAt  time.Time
	UpdatedAt  time.Time
	OptionName string
}

// Make is vehicle make (i.e. Volvo)
type Make struct {
	ID        int       `json:"id"`
	Make      string    `json:"make"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// Model is vehicle model (i.e. Camry)
type Model struct {
	ID        int       `json:"id"`
	Model     string    `json:"model"`
	MakeID    int       `json:"-"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// Image is a vehicle image
type Image struct {
	ID        int       `xml:"-" json:"id"`
	VehicleID int       `xml:"-" json:"vehicle_id"`
	Image     string    `xml:"image" json:"image"`
	SortOrder int       `xml:"-" json:"sort_order"`
	CreatedAt time.Time `xml:"-" json:"created-at"`
	UpdatedAt time.Time `xml:"-" json:"updated_at"`
}

// VehicleVideo is the join table for vehicles/videos
type VehicleVideo struct {
	ID        int
	VehicleID int
	VideoID   int
	VideoName string
	FileName  string
	Thumb     string
	Is360     int
	Duration  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Video holds a video
type Video struct {
	ID                      int
	VideoName               string
	FileName                string
	Public                  int
	Description             string
	CategoryID              int
	SortOrder               int
	Thumb                   string
	ConvertedForStreamingAt time.Time
	Duration                int
	Is360                   int
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

// SalesStaff holds sales people
type SalesStaff struct {
	ID    int
	Name  string
	Slug  string
	Email string
	Phone string
	Image string
}

// CreditApp holds a credit application
type CreditApp struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"-"`
	Phone     string    `json:"-"`
	Address   string    `json:"-"`
	City      string    `json:"-"`
	Province  string    `json:"-"`
	Zip       string    `json:"-"`
	Vehicle   string    `json:"-"`
	Rent      string    `json:"-"`
	Employer  string    `json:"-"`
	Income    string    `json:"-"`
	DOB       string    `json:"-"`
	Processed int       `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"-"`
}

// TestDrive holds a test drive
type TestDrive struct {
	ID            int
	UsersName     string
	Email         string
	Phone         string
	PreferredDate string
	PreferredTime string
	VehicleID     int
	Processed     int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// QuickQuote holds a quick quote
type QuickQuote struct {
	ID        int       `json:"id"`
	UsersName string    `json:"users_name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	VehicleID int       `json:"-"`
	Vehicle   Vehicle   `json:"-"`
	Processed int       `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"-"`
}
