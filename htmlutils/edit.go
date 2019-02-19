package htmlutils

import (
	"bytes"
	"fmt"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/utils"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	kDateField             = "date_field"
	kTimeField             = "time_field"
	kInt64Field            = "int64_field"
	kFloat64Field          = "float64_field"
	kInt64SliceField       = "int64slice_field"
	kInt64SliceHiddenField = "int64slice_hidden_field"
	kRadioField            = "radio_field"
	kStringHiddenField     = "string_hidden_field"
	kStringField           = "string_field"
	kPasswordField         = "password_field"
	kBigStringField        = "big_string_field"
	kInt64HiddenField      = "int64_hidden_field"
	kFloat64HiddenField    = "float64_hidden_field"
	kFileField             = "file_field"
	kImageRadioField       = "image_radio_field"
	kImageField            = "image_field"

	kStringColumn = "string_column"
	kImageColumn  = "image_column"
	kHTMLColumn   = "html_column"
	kActionColumn = "action_column"
)

var saveImageInterface SaveImageInterface

func RegisterSaveImageInterface(theInterface SaveImageInterface) {
	saveImageInterface = theInterface
}

type EditObject struct {
	entries []*EditEntry
	postUrl string
	method  string
}

func NewEditObject(entries []*EditEntry, postUrl string) *EditObject {
	fmt.Println("entries", entries)
	return &EditObject{
		entries: entries,
		postUrl: postUrl,
		method:  "post",
	}
}

func NewEditObjectGet(entries []*EditEntry, postUrl string) *EditObject {
	fmt.Println("entries", entries)
	return &EditObject{
		entries: entries,
		postUrl: postUrl,
		method:  "get",
	}
}

func (editObject *EditObject) HideField(name string) {
	for _, entry := range editObject.entries {
		if entry.name == name {
			if entry.fieldType == kStringField {
				entry.fieldType = kStringHiddenField
			} else if entry.fieldType == kInt64Field {
				entry.fieldType = kInt64HiddenField
			} else if entry.fieldType == kInt64SliceField {
				entry.fieldType = kInt64SliceHiddenField
			} else if entry.fieldType == kPasswordField {
				entry.fieldType = kStringHiddenField
			}
		}
	}
}

func (editObject *EditObject) GetScriptHTML() template.HTML {
	dateFields := make([]string, 0)
	imageFields := make([]string, 0)
	for _, entry := range editObject.entries {
		if entry.fieldType == kDateField {
			dateFields = append(dateFields, entry.name)
		} else if entry.fieldType == kImageRadioField {
			imageFields = append(imageFields, entry.name)
		}
	}

	data := make(map[string]interface{})
	data["date_fields"] = dateFields
	data["image_fields"] = imageFields
	tmpl, err := template.New("").Parse(`
		<script type="text/javascript">
			$(document).ready(function(){
				{{range $index, $element := .date_fields}}
					$( "#{{$element}}" ).datepicker({
						dateFormat: "dd-mm-yy"
					});
				{{end}}

				{{range $index, $element := .image_fields}}
					$('input[class="{{$element}}"]:radio').change(
					    function(){
					    	$('#{{$element}}_display').attr('src',"/images/" +this.value)
					    }
					);
				{{end}}
			});


			 
		</script>
		`)
	if err != nil {
		log.LogSerious("err parse template %v", err)
		return ""
	}
	var htmlBuffer bytes.Buffer
	err = tmpl.Execute(&htmlBuffer, data)
	if err != nil {
		log.LogSerious("err exc template %v", err)
		return ""
	}
	return template.HTML(htmlBuffer.String())
}

func (editObject *EditObject) GetFormHTML() template.HTML {
	dateFields := make([]string, 0)
	for _, entry := range editObject.entries {
		if entry.fieldType == kDateField {
			dateFields = append(dateFields, entry.name)
		}
	}

	data := make(map[string]interface{})
	data["post_url"] = editObject.postUrl
	data["method"] = editObject.method
	data["entries"] = editObject.entries
	tmpl, err := template.New("").Parse(`
		<div class="row">
		<form action="{{.post_url}}" method="{{.method}}" enctype="multipart/form-data" class="col-md-4">
			{{range $index, $element := .entries}}
				{{$element.SerializedData}}
			{{end}}
			<input type="submit" value="Submit" class="btn btn-primary"/>
		</form>
	</div>
		`)
	if err != nil {
		log.LogSerious("err parse template %v", err)
		return ""
	}
	var htmlBuffer bytes.Buffer
	err = tmpl.Execute(&htmlBuffer, data)
	if err != nil {
		log.LogSerious("err exc template %v", err)
		return ""
	}
	return template.HTML(htmlBuffer.String())
}

func (editObject *EditObject) UpdateEntryFromRequestData(data map[string]interface{}) {
	for _, entry := range editObject.entries {
		if entry.fieldType == kInt64Field || entry.fieldType == kInt64HiddenField {
			requestValue := utils.GetInt64AtPath(data, entry.name)
			entry.value = fmt.Sprintf("%v", requestValue)
		} else if entry.fieldType == kStringField || entry.fieldType == kStringHiddenField || entry.fieldType == kBigStringField || entry.fieldType == kPasswordField {
			entry.value = utils.GetStringAtPath(data, entry.name)
		} else if entry.fieldType == kDateField {
			date := data[entry.name].(time.Time)
			dateString, _ := utils.FormatTimeToVietnamTime(date)
			entry.value = dateString
		} else if entry.fieldType == kFloat64Field {
			requestValue := utils.GetFloat64AtPath(data, entry.name)
			entry.value = fmt.Sprintf("%v", requestValue)
		} else if entry.fieldType == kInt64SliceField || entry.fieldType == kInt64SliceHiddenField {
			entry.value = utils.GetStringAtPath(data, entry.name)
		} else if entry.fieldType == kRadioField {
			entry.value = utils.GetStringAtPath(data, entry.name)
		} else if entry.fieldType == kImageRadioField {
			entry.value = utils.GetStringAtPath(data, entry.name)
		} else if entry.fieldType == kImageField {

		}
	}
}

func (editObject *EditObject) ConvertGetRequestToData(request *http.Request) (data map[string]interface{}) {
	data = make(map[string]interface{})
	for _, entry := range editObject.entries {
		requestValue := request.URL.Query().Get(entry.name)
		if entry.fieldType == kInt64Field || entry.fieldType == kInt64HiddenField {
			data[entry.name], _ = strconv.ParseInt(requestValue, 10, 64)
		} else if entry.fieldType == kStringField || entry.fieldType == kStringHiddenField || entry.fieldType == kBigStringField || entry.fieldType == kPasswordField {
			data[entry.name] = requestValue
		} else if entry.fieldType == kDateField {
			dateString := requestValue
			if len(dateString) != 0 {
				data[entry.name] = utils.TimeFromVietnameseDateString(dateString)
			}
		} else if entry.fieldType == kFloat64Field {
			data[entry.name], _ = strconv.ParseFloat(requestValue, 64)
		} else if entry.fieldType == kInt64SliceField || entry.fieldType == kInt64SliceHiddenField {
			rawString := requestValue
			stringSlice := strings.Split(rawString, ",")
			int64Slice := make([]int64, 0)
			for _, stringValue := range stringSlice {
				int64Value, _ := strconv.ParseInt(stringValue, 10, 64)
				int64Slice = append(int64Slice, int64Value)
			}
			data[entry.name] = int64Slice
		} else if entry.fieldType == kRadioField {
			rawString := requestValue
			data[entry.name] = rawString
		} else if entry.fieldType == kImageRadioField {
			rawString := requestValue
			data[entry.name] = rawString
		}
	}
	return data
}

func (editObject *EditObject) ConvertRequestToData(request *http.Request) (data map[string]interface{}) {
	data = make(map[string]interface{})
	for _, entry := range editObject.entries {
		requestValue := request.FormValue(entry.name)
		if entry.fieldType == kInt64Field || entry.fieldType == kInt64HiddenField {
			data[entry.name], _ = strconv.ParseInt(requestValue, 10, 64)
		} else if entry.fieldType == kStringField || entry.fieldType == kStringHiddenField || entry.fieldType == kBigStringField || entry.fieldType == kPasswordField {
			data[entry.name] = requestValue
		} else if entry.fieldType == kDateField {
			dateString := requestValue
			if len(dateString) != 0 {
				data[entry.name] = utils.TimeFromVietnameseDateString(dateString)
			}
		} else if entry.fieldType == kFloat64Field {
			data[entry.name], _ = strconv.ParseFloat(requestValue, 64)
		} else if entry.fieldType == kInt64SliceField || entry.fieldType == kInt64SliceHiddenField {
			rawString := requestValue
			stringSlice := strings.Split(rawString, ",")
			int64Slice := make([]int64, 0)
			for _, stringValue := range stringSlice {
				int64Value, _ := strconv.ParseInt(stringValue, 10, 64)
				int64Slice = append(int64Slice, int64Value)
			}
			data[entry.name] = int64Slice
		} else if entry.fieldType == kRadioField {
			rawString := requestValue
			data[entry.name] = rawString
		} else if entry.fieldType == kImageRadioField {
			rawString := requestValue
			data[entry.name] = rawString
		} else if entry.fieldType == kImageField {
			file, _, err := request.FormFile(entry.name)
			var iconUrl string
			if err != nil || file == nil {
				iconUrl = request.FormValue(fmt.Sprintf("old_%s", entry.name))
			}
			if file != nil {
				data, err := saveImageInterface.SaveImageFile(file)
				file.Close()
				if err != nil {
					log.LogSerious("err saveimage %v", err)
				}
				iconUrl = utils.GetStringAtPath(data, "absolute_url")
			}
			data[entry.name] = iconUrl

		}
	}
	return data
}

type EditEntry struct {
	fieldType   string
	name        string
	placeHolder string
	value       string
	row         int64
	title       string
	options     []string
	texts       []string
}

func NewDateField(title string, name string, placeHolder string, date time.Time) *EditEntry {
	dateString, _ := utils.FormatTimeToVietnamTime(date)
	return &EditEntry{
		fieldType:   kDateField,
		title:       title,
		placeHolder: placeHolder,
		name:        name,
		value:       dateString,
	}
}

func NewInt64SliceField(title string, name string, placeHolder string, slice []int64) *EditEntry {
	stringSlice := make([]string, 0)
	for _, element := range slice {
		stringSlice = append(stringSlice, fmt.Sprintf("%d", element))
	}
	value := strings.Join(stringSlice, ",")
	return &EditEntry{
		fieldType:   kInt64SliceField,
		title:       title,
		placeHolder: placeHolder,
		name:        name,
		value:       value,
	}
}

func NewInt64SliceHiddenField(title string, name string, placeHolder string, slice []int64) *EditEntry {
	stringSlice := make([]string, 0)
	for _, element := range slice {
		stringSlice = append(stringSlice, fmt.Sprintf("%d", element))
	}
	value := strings.Join(stringSlice, ",")
	return &EditEntry{
		fieldType:   kInt64SliceHiddenField,
		title:       title,
		placeHolder: placeHolder,
		name:        name,
		value:       value,
	}
}

func NewStringField(title string, name string, placeHolder string, value string) *EditEntry {
	return &EditEntry{
		fieldType:   kStringField,
		title:       title,
		placeHolder: placeHolder,
		name:        name,
		value:       value,
	}
}

func NewPasswordField(title string, name string, placeHolder string) *EditEntry {
	return &EditEntry{
		fieldType:   kPasswordField,
		title:       title,
		placeHolder: placeHolder,
		name:        name,
		value:       "",
	}
}

func NewBigStringField(title string, name string, row int64, value string) *EditEntry {
	return &EditEntry{
		fieldType: kBigStringField,
		title:     title,
		name:      name,
		row:       row,
		value:     value,
	}
}

func NewStringHiddenField(name string, value string) *EditEntry {
	return &EditEntry{
		fieldType: kStringHiddenField,
		name:      name,
		value:     value,
	}
}

func NewInt64Field(title string, name string, placeHolder string, value int64) *EditEntry {
	return &EditEntry{
		fieldType:   kInt64Field,
		title:       title,
		placeHolder: placeHolder,
		name:        name,
		value:       fmt.Sprintf("%d", value),
	}
}

func NewFloat64Field(title string, name string, placeHolder string, value float64) *EditEntry {
	return &EditEntry{
		fieldType:   kFloat64Field,
		title:       title,
		placeHolder: placeHolder,
		name:        name,
		value:       fmt.Sprintf("%.5f", value),
	}
}

func NewInt64HiddenField(name string, value int64) *EditEntry {
	return &EditEntry{
		fieldType: kInt64HiddenField,
		name:      name,
		value:     fmt.Sprintf("%d", value),
	}
}

func NewRadioField(title string, name string, value string, options []string) *EditEntry {
	return &EditEntry{
		fieldType: kRadioField,
		title:     title,
		name:      name,
		value:     value,
		options:   options,
		texts:     options,
	}
}

func NewRadioFieldDetails(title string, name string, value string, options []string, texts []string) *EditEntry {
	return &EditEntry{
		fieldType: kRadioField,
		title:     title,
		name:      name,
		value:     value,
		options:   options,
		texts:     texts,
	}
}

func NewImageRadioField(title string, name string, value string, options []string) *EditEntry {
	return &EditEntry{
		fieldType: kImageRadioField,
		title:     title,
		name:      name,
		value:     value,
		options:   options,
	}
}

func NewImageField(title string, name string, value string) *EditEntry {
	return &EditEntry{
		fieldType: kImageField,
		title:     title,
		name:      name,
		value:     value,
	}
}

func (entry *EditEntry) SerializedData() template.HTML {
	fmt.Println("entry", entry.fieldType, entry.name)
	if utils.ContainsByString([]string{kInt64Field, kStringField, kDateField, kInt64SliceField, kFloat64Field}, entry.fieldType) {
		format := `
	<div class="form-group">
				<label for="%s">%s</label>
				<input type="text" id="%s" name="%s" class="form-control" placeholder="%s" value="%s" aria-describedby="basic-addon1">
	</div>
	`
		return template.HTML(fmt.Sprintf(format, entry.name, entry.title, entry.name, entry.name, entry.placeHolder, entry.value))
	} else if utils.ContainsByString([]string{kBigStringField}, entry.fieldType) {
		format := `
	<div class="form-group">
				<label for="%s">%s</label>
				<textarea id="%s" name="%s" class="form-control" rows="%d">%s</textarea>
	</div>
	`
		return template.HTML(fmt.Sprintf(format, entry.name, entry.title, entry.name, entry.name, entry.row, entry.value))
	} else if utils.ContainsByString([]string{kPasswordField}, entry.fieldType) {
		format := `
	<div class="form-group">
				<label for="%s">%s</label>
				<input type="password" id="%s" name="%s" class="form-control" placeholder="%s" value="%s" aria-describedby="basic-addon1">
	</div>
	`
		return template.HTML(fmt.Sprintf(format, entry.name, entry.title, entry.name, entry.name, entry.placeHolder, entry.value))
	} else if utils.ContainsByString([]string{kInt64HiddenField, kStringHiddenField, kInt64SliceHiddenField}, entry.fieldType) {

		format := `
			<input type="hidden" id="%s" name="%s" class="form-control" value="%s" aria-describedby="basic-addon1">
	`
		return template.HTML(fmt.Sprintf(format, entry.name, entry.name, entry.value))
	} else if entry.fieldType == kRadioField {
		data := make(map[string]interface{})
		data["name"] = entry.name
		data["title"] = entry.title
		data["value"] = entry.value
		data["options"] = entry.options
		data["texts"] = entry.texts
		tmpl, err := template.New("").Parse(`
			<div class="form-group">
				<label for="{{.name}}">{{.title}}</label> <br/>
				{{range $index, $element := .options}}
				<label class="radio-inline" id="{{$.name}}">
					<input type="radio" name="{{$.name}}" value="{{$element}}" {{if eq $.value $element}} checked="checked" {{ end }}> {{index $.texts $index}}
				</label>
				{{end}}
			</div>
		`)
		if err != nil {
			log.LogSerious("err parse template %v", err)
			return ""
		}
		var htmlBuffer bytes.Buffer
		err = tmpl.Execute(&htmlBuffer, data)
		if err != nil {
			log.LogSerious("err exc template %v", err)
			return ""
		}
		return template.HTML(htmlBuffer.String())
	} else if entry.fieldType == kImageRadioField {
		data := make(map[string]interface{})
		data["name"] = entry.name
		data["title"] = entry.title
		data["value"] = entry.value
		data["options"] = entry.options

		tmpl, err := template.New("").Parse(`
			<div class="row">
				<div class="form-group col-md-6" >
					<label for="image_name">{{.title}}</label>

					{{range $index,$element := .options}}
					<div class="radio">
						<label>
							<input type="radio" class="{{$.name}}" name="{{$.name}}" value="{{$element}}" {{if eq $.value $element}} checked="checked" {{ end }}>
							{{$element}}
						</label>
					</div>
					{{end}}
				</div>

				<div class="col-md-6">
					<img id="{{.name}}_display" src="/images/{{.value}}" width="285px" height="384px"/>
				</div>
			</div>
		`)
		if err != nil {
			log.LogSerious("err parse template %v", err)
			return ""
		}
		var htmlBuffer bytes.Buffer
		err = tmpl.Execute(&htmlBuffer, data)
		if err != nil {
			log.LogSerious("err exc template %v", err)
			return ""
		}
		return template.HTML(htmlBuffer.String())
	} else if entry.fieldType == kImageField {
		format := `
	<div class="form-group">
				<label for="%s">%s</label>
				<img src="%s" width="300px"/>
				<input type="hidden" name="old_%s" class="form-control" value="%s">
				<input type="file" class="form-control" name="%s" id="%s"/>
	</div>
	`
		return template.HTML(fmt.Sprintf(format, entry.name, entry.title, entry.value, entry.name, entry.value, entry.name, entry.name))
	}
	return ""
}
