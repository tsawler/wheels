package clienthandlers

import (
	"encoding/json"
	"fmt"
	"github.com/tsawler/goblender/client/clienthandlers/clientmodels"
	channel_data "github.com/tsawler/goblender/pkg/channel-data"
	"github.com/tsawler/goblender/pkg/forms"
	"github.com/tsawler/goblender/pkg/helpers"
	"github.com/tsawler/goblender/pkg/templates"
	"html/template"
	"net/http"
	"strconv"
	"time"
)

// JSONResponse is a generic struct to hold json responses
type JSONResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

// QuickQuote sends a quick quote request
func QuickQuote(w http.ResponseWriter, r *http.Request) {
	name := r.Form.Get("name")
	email := r.Form.Get("email")
	phone := r.Form.Get("phone")
	interest := r.Form.Get("interested")
	vid, _ := strconv.Atoi(r.Form.Get("vehicle_id"))
	msg := r.Form.Get("msg")
	if msg != "" {
		interest = msg
	}

	content := fmt.Sprintf(`
		<p>
			<strong>Wheels and Deals Quick Quote Request</strong>:<br><br>
			<strong>Name:</strong> %s <br>
			<strong>Email:</strong> %s <br>
			<strong>Phone:</strong> %s <br>
			<strong>Interested In:</strong><br><br>
			%s
		</p>
`, name, email, phone, interest)

	var cc []string
	cc = append(cc, "wheelsanddeals@pbssystems.com")

	mailMessage := channel_data.MailData{
		ToName:      "",
		ToAddress:   "alex.gilbert@wheelsanddeals.ca",
		FromName:    app.PreferenceMap["smtp-from-name"],
		FromAddress: app.PreferenceMap["smtp-from-email"],
		Subject:     "Wheels and Deals Quick Quote Request",
		Content:     template.HTML(content),
		Template:    "bootstrap.mail.tmpl",
		CC:          cc,
	}

	helpers.SendEmail(mailMessage)

	qq := clientmodels.QuickQuote{
		UsersName: name,
		Email:     email,
		Phone:     phone,
		VehicleID: vid,
		Processed: 0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := vehicleModel.InsertQuickQuote(qq)
	if err != nil {
		errorLog.Println(err)
	}

	theData := JSONResponse{
		OK: true,
	}

	// build the json response from the struct
	out, err := json.MarshalIndent(theData, "", "    ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// send json to client
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(out)
	if err != nil {
		errorLog.Println(err)
	}
}

// TestDrive sends a test drive request
func TestDrive(w http.ResponseWriter, r *http.Request) {
	name := r.Form.Get("name")
	email := r.Form.Get("email")
	phone := r.Form.Get("phone")
	interest := r.Form.Get("interested")
	pDate := r.Form.Get("preferred_date")
	pTime := r.Form.Get("preferred_time")
	vid, _ := strconv.Atoi(r.Form.Get("vehicle_id"))

	content := fmt.Sprintf(`
		<p>
			<strong>PowerSports Test Drive Request</strong>:<br><br>
			<strong>Name:</strong> %s <br>
			<strong>Email:</strong> %s <br>
			<strong>Phone:</strong> %s <br>
			<strong>Preferred Date:</strong> %s<br>
			<strong>Preferred Time:</strong> %s<br>
			<strong>Interested In:</strong><br><br>
			%s
		</p>
`, name, email, phone, pDate, pTime, interest)

	var cc []string
	cc = append(cc, "wheelsanddeals@pbssystems.com")
	//cc = append(cc, "john.eliakis@wheelsanddeals.ca")

	mailMessage := channel_data.MailData{
		ToName:      "",
		ToAddress:   "alex.gilbert@wheelsanddeals.ca",
		FromName:    app.PreferenceMap["smtp-from-name"],
		FromAddress: app.PreferenceMap["smtp-from-email"],
		Subject:     "PowerSports Test Drive Request",
		Content:     template.HTML(content),
		Template:    "generic-email.mail.tmpl",
		CC:          cc,
	}

	helpers.SendEmail(mailMessage)

	// save
	td := clientmodels.TestDrive{
		UsersName:     name,
		Email:         email,
		Phone:         phone,
		PreferredDate: pDate,
		PreferredTime: pTime,
		VehicleID:     vid,
		Processed:     0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := vehicleModel.InsertTestDrive(td)
	if err != nil {
		errorLog.Println(err)
	}

	theData := JSONResponse{
		OK: true,
	}

	// build the json response from the struct
	out, err := json.MarshalIndent(theData, "", "    ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// send json to client
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(out)
	if err != nil {
		errorLog.Println(err)
	}
}

// SendFriend sends to a friend
func SendFriend(w http.ResponseWriter, r *http.Request) {
	name := r.Form.Get("name")
	email := r.Form.Get("email")
	interest := r.Form.Get("interested")
	url := r.Form.Get("url")
	infoLog.Println("url is ", url)

	content := fmt.Sprintf(`
		<p>
			Hi:
			<br>
			<br>
			%s thought you might be interested in this item at Jim Gilbert's PowerSports:
			<br><br>
			%s
			<br><br>
			You can see the item by following this link:
			<a href='http://%s'>Click here to see the item!</a>
		</p>
`, name, interest, url)

	mailMessage := channel_data.MailData{
		ToName:      "",
		ToAddress:   email,
		FromName:    app.PreferenceMap["smtp-from-name"],
		FromAddress: app.PreferenceMap["smtp-from-email"],
		Subject:     fmt.Sprintf("%s thought you might be intersted in this item from Jim Gilbert's PowerSports", name),
		Content:     template.HTML(content),
		Template:    "generic-email.mail.tmpl",
	}

	helpers.SendEmail(mailMessage)

	theData := JSONResponse{
		OK: true,
	}

	// build the json response from the struct
	out, err := json.MarshalIndent(theData, "", "    ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// send json to client
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(out)
	if err != nil {
		errorLog.Println(err)
	}
}

// CreditApp displays credit app page
func CreditApp(w http.ResponseWriter, r *http.Request) {
	pg, err := repo.DB.GetPageBySlug("credit-application")

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	rowSets := make(map[string]interface{})
	var years []int
	for y := time.Now().Year(); y > (time.Now().Year() - 100); y-- {
		years = append(years, y)
	}

	rowSets["years"] = years

	helpers.Render(w, r, "credit-app.page.tmpl", &templates.TemplateData{
		Form:    forms.New(nil),
		Page:    pg,
		RowSets: rowSets,
	})
}

// PostCreditApp handles ajax post of credit application
func PostCreditApp(w http.ResponseWriter, r *http.Request) {
	form := forms.New(r.PostForm)
	form.Required("first_name", "last_name", "email", "y", "m", "y", "phone", "address", "city", "province", "zip", "rent", "income", "vehicle", "g-recaptcha-response")

	form.RecaptchaValid(r.RemoteAddr)

	if !form.Valid() {
		theData := JSONResponse{
			OK:      false,
			Message: "Form error",
		}

		// build the json response from the struct
		out, err := json.MarshalIndent(theData, "", "    ")
		if err != nil {
			helpers.ServerError(w, err)
			return
		}

		// send json to client
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(out)
		if err != nil {
			errorLog.Println(err)
		}
		return
	}

	// create email
	content := fmt.Sprintf(`
		<p>
			<strong>Wheels and Deals Credit Application</strong>:<br><br>
			<strong>Name:</strong> %s  %s<br>
			<strong>Date of birth:</strong> %s <br>
			<strong>Email:</strong> %s <br>
			<strong>Phone:</strong> %s <br>
			<strong>Address:</strong> %s %s, %s, %s<br>
			<strong>Rent/Mortgage</strong>: %s<br>
			<strong>Employer</strong>: %s<br>
			<strong>Income</strong>: %s<br>
			<strong>Interested In:</strong><br><br>
			%s
		</p>
`,
		form.Get("first_name"),
		form.Get("last_name"),
		fmt.Sprintf("%s-%s-%s", form.Get("y"), form.Get("m"), form.Get("d")),
		form.Get("phone"),
		form.Get("email"),
		form.Get("address"),
		form.Get("city"),
		form.Get("province"),
		form.Get("zip"),
		form.Get("rent"),
		form.Get("employer"),
		form.Get("income"),
		form.Get("vehicle"),
	)

	var cc []string
	cc = append(cc, "wheelsanddeals@pbssystems.com")
	//cc = append(cc, "john.eliakis@wheelsanddeals.ca")
	cc = append(cc, "chelsea.gilbert@wheelsanddeals.ca")

	mailMessage := channel_data.MailData{
		ToName:      "",
		ToAddress:   "alex.gilbert@wheelsanddeals.ca",
		FromName:    app.PreferenceMap["smtp-from-name"],
		FromAddress: app.PreferenceMap["smtp-from-email"],
		Subject:     "Wheels and Deals Credit Application",
		Content:     template.HTML(content),
		Template:    "bootstrap.mail.tmpl",
		CC:          cc,
	}

	infoLog.Println("Sending email")

	helpers.SendEmail(mailMessage)

	// save the application
	creditApp := clientmodels.CreditApp{
		FirstName: form.Get("first_name"),
		LastName:  form.Get("last_name"),
		Email:     form.Get("email"),
		Phone:     form.Get("phone"),
		Address:   form.Get("address"),
		City:      form.Get("city"),
		Province:  form.Get("province"),
		Zip:       form.Get("zip"),
		Vehicle:   form.Get("vehicle"),
		Processed: 0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := vehicleModel.InsertCreditApp(creditApp)
	if err != nil {
		errorLog.Println(err)
	}

	theData := JSONResponse{
		OK: true,
	}

	// build the json response from the struct
	out, err := json.MarshalIndent(theData, "", "    ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// send json to client
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(out)
	if err != nil {
		errorLog.Println(err)
	}

}
