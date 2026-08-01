package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"webProj/internal/models"
	"webProj/internal/tokens"
)

type templateData struct {
	StringMap  	map[string]string
	IntMap     	map[string]int
	FloatMap   	map[string]float32
	Data       	map[string]any
	CSRFToken  	string
	Flash      	string
	Warning    	string
	Error      	string
	IsAuth     	bool
	IsAdmin    	bool
	UserUUID	string
	API        	string
	CSSVersion 	string
}

var functions = template.FuncMap{
	"formatCurrency": FormatCurrency,
}

//go:embed templates
var templateFS embed.FS

func (app *application) addDefaultData(td *templateData, w http.ResponseWriter, r *http.Request) *templateData {

	td.API = app.config.api
	td.CSSVersion = cssVersion

	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Warning = app.Session.PopString(r.Context(), "warning")
	td.Error = app.Session.PopString(r.Context(), "error")

	if user := app.contextGetUser(r); user != nil {
		td.IsAuth = true
		td.IsAdmin = user.Role == models.RoleAdmin
		td.UserUUID = user.UUID
	}

	if td.StringMap == nil {
		td.StringMap = make(map[string]string)
	}

	deviceID := app.getOrCreateDeviceID(w, r)
	td.StringMap["csrf_token"] = tokens.Generate(app.config.csrfSecret, deviceID)
	td.StringMap["device_id"] = deviceID

	return td
}

func (app *application) renderTemplate(w http.ResponseWriter, r *http.Request, page string, templData *templateData, partials ...string) error {

	var templ *template.Template
	var err error

	templateToRender := fmt.Sprintf("templates/%s.page.gohtml", page)

	// check cache, but only for production (to not clear cache when developing)
	app.templateMu.RLock()
	cachedTempl, templateInMap := app.templateCache[templateToRender]
	app.templateMu.RUnlock()

	if app.config.env == "production" && templateInMap {

		templ = cachedTempl

	} else {

		templ, err = app.parseTemplate(partials, page, templateToRender)
		if err != nil {

			app.errorLog.Println(err)
			return err
		}
	}

	if templData == nil {

		templData = &templateData{}
	}

	templData = app.addDefaultData(templData, w, r)

	buf := new(bytes.Buffer)
	err = templ.Execute(buf, templData)
	if err != nil {

		app.errorLog.Println(err)
		return err
	}

	_, err = buf.WriteTo(w)
	if err != nil {

		app.errorLog.Println(err)
		return err
	}

	return nil
}

func (app *application) parseTemplate(partials []string, page, templateToRender string) (*template.Template, error) {

	var templ *template.Template
	var err error

	var partialPatterns []string
	if len(partials) > 0 {

		for _, x := range partials {

			partialPatterns = append(partialPatterns, fmt.Sprintf("templates/%s.partial.gohtml", x))
		}
	}

	if len(partials) > 0 {

		patterns := append([]string{"templates/base.layout.gohtml"}, partialPatterns...)
		patterns = append(patterns, templateToRender)
		templ, err = template.New(fmt.Sprintf("%s.page.gohtml", page)).Funcs(functions).ParseFS(templateFS, patterns...)

	} else {

		templ, err = template.New(fmt.Sprintf("%s.page.gohtml", page)).Funcs(functions).ParseFS(templateFS, "templates/base.layout.gohtml", templateToRender)
	}

	if err != nil {
		app.errorLog.Println(err)
		return nil, err
	}

	if app.config.env == "production" {

		app.templateMu.Lock()
		app.templateCache[templateToRender] = templ
		app.templateMu.Unlock()
	}

	return templ, nil
}
