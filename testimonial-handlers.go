package clienthandlers

import (
	"fmt"
	"github.com/tsawler/goblender/pkg/cache"
	"github.com/tsawler/goblender/pkg/helpers"
	"github.com/tsawler/goblender/pkg/models"
	"github.com/tsawler/goblender/pkg/templates"
	"net/http"
	"strconv"
)

// AllWordsPublic displays all word of mouth
func AllWordsPublic(w http.ResponseWriter, r *http.Request) {
	slug := "huggable-word-of-mouth"

	pageIndex, err := strconv.Atoi(r.URL.Query().Get(":pageIndex"))
	if err != nil {
		pageIndex = 1
	}

	perPage := 10
	offset := (pageIndex - 1) * perPage

	var p models.Page
	inCache, err := cache.Has(fmt.Sprintf("page-%s", slug))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	if inCache {
		result, err := cache.Get(fmt.Sprintf("page-%s", slug))
		if err != nil {
			helpers.ServerError(w, err)
			return
		}
		p = result.(models.Page)

	} else {
		p, err = repo.DB.GetPageBySlug(slug)
		if err == models.ErrNoRecord {
			helpers.NotFound(w)
			return
		} else if err != nil {
			helpers.ServerError(w, err)
			return
		}

		err = cache.Set(fmt.Sprintf("page-%s", slug), p)
		if err != nil {
			helpers.ServerError(w, err)
			return
		}
	}

	words, num, err := vehicleModel.AllWordOfMouthPaginated(perPage, offset)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	myMap := make(map[string]interface{})
	myMap["words"] = words

	intMap := make(map[string]int)
	intMap["num"] = num
	intMap["current-page"] = pageIndex

	helpers.Render(w, r, "word-of-mouth.page.tmpl", &templates.TemplateData{
		Page:    p,
		RowSets: myMap,
		IntMap:  intMap,
	})
}

// AllTestimonialsPublic displays all testimonials
func AllTestimonialsPublic(w http.ResponseWriter, r *http.Request) {
	slug := "customer-testimonials"

	pageIndex, err := strconv.Atoi(r.URL.Query().Get(":pageIndex"))
	if err != nil {
		pageIndex = 1
	}

	perPage := 10
	offset := (pageIndex - 1) * perPage

	var p models.Page
	inCache, err := cache.Has(fmt.Sprintf("page-%s", slug))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	if inCache {
		result, err := cache.Get(fmt.Sprintf("page-%s", slug))
		if err != nil {
			helpers.ServerError(w, err)
			return
		}
		p = result.(models.Page)

	} else {
		p, err = repo.DB.GetPageBySlug(slug)
		if err == models.ErrNoRecord {
			helpers.NotFound(w)
			return
		} else if err != nil {
			helpers.ServerError(w, err)
			return
		}

		err = cache.Set(fmt.Sprintf("page-%s", slug), p)
		if err != nil {
			helpers.ServerError(w, err)
			return
		}
	}

	testimonials, num, err := vehicleModel.AllTestimonialsPaginated(perPage, offset)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	myMap := make(map[string]interface{})
	myMap["testimonials"] = testimonials

	intMap := make(map[string]int)
	intMap["num"] = num
	intMap["current-page"] = pageIndex

	helpers.Render(w, r, "testimonials.page.tmpl", &templates.TemplateData{
		Page:    p,
		RowSets: myMap,
		IntMap:  intMap,
	})
}
