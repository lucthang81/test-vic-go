package crud

import (
	"fmt"
	"github.com/vic/vic_go/datacenter"
	"github.com/vic/vic_go/htmlutils"
	"github.com/vic/vic_go/utils"
	"html/template"
	"net/http"
	"strings"
)

const (
	AttributeTypeInt64Type      = "int64"
	AttributeTypeFloat64Type    = "float64"
	AttributeTypeInt64SliceType = "int64slice"
	AttributeTypeStringType     = "string"
)

type Object interface {
	// db
	DatabaseName() string

	// attribute
	Id() int64
	SetId(id int64)
	GetValueForAttributeName(name string) interface{}
	SetValueForAttributeName(name string, value interface{})
	GetAttributes() []*Attribute

	// url
	GetCreateUrl() string
	GetUpdateUrl() string
	GetDeleteUrl() string
}

type Attribute struct {
	title         string
	name          string
	attributeType string
	defaultValue  interface{}

	isEditable bool
}

func NewAttribute(title string, name string, attributeType string, defaultValue interface{}, isEditable bool) *Attribute {
	return &Attribute{
		title:         title,
		name:          name,
		attributeType: attributeType,
		defaultValue:  defaultValue,
		isEditable:    isEditable,
	}
}

func UpdateAndSave(db datacenter.DBInterface, object Object, request *http.Request) (err error) {
	data := GetHTMLForEditObjectForm(object).ConvertRequestToData(request)
	for _, attribute := range object.GetAttributes() {
		if attribute.isEditable {
			value := utils.GetDataAtPath(data, attribute.name)
			object.SetValueForAttributeName(attribute.name, value)
		}
	}
	err = SaveData(db, object)
	return
}

func SaveData(db datacenter.DBInterface, object Object) (err error) {
	queryString := fmt.Sprintf("UPDATE %s SET ", object.DatabaseName())
	var count int
	count = 1
	var params []interface{}
	params = make([]interface{}, 0)
	for index, attribute := range object.GetAttributes() {
		if index == 0 {
			queryString = fmt.Sprintf("%s %s = $%d", queryString, attribute.name, count)
		} else {
			queryString = fmt.Sprintf("%s, %s = $%d", queryString, attribute.name, count)
		}
		if attribute.attributeType == AttributeTypeInt64SliceType {
			slice := object.GetValueForAttributeName(attribute.name).([]int64)
			raw := utils.ConvertInt64SliceToRawString(slice)
			params = append(params, raw)
		} else {
			params = append(params, object.GetValueForAttributeName(attribute.name))
		}
		count++
	}
	params = append(params, object.Id())
	queryString = fmt.Sprintf("%s WHERE id = $%d", queryString, count)
	fmt.Println(queryString)
	_, err = db.Exec(queryString, params...)
	if err != nil {
		fmt.Println("err", err, params, queryString)
	}
	return err
}

func GetHTMLForCreateObjectForm(object Object) *htmlutils.EditObject {
	rows := make([]*htmlutils.EditEntry, 0)
	for _, attribute := range object.GetAttributes() {
		var row *htmlutils.EditEntry
		if attribute.isEditable {
			if attribute.attributeType == AttributeTypeStringType {
				value := utils.GetStringFromInterface(attribute.defaultValue)
				row = htmlutils.NewStringField(attribute.title, attribute.name, attribute.title, value)
			} else if attribute.attributeType == AttributeTypeInt64Type {
				value := utils.GetInt64FromScanResult(attribute.defaultValue)
				row = htmlutils.NewInt64Field(attribute.title, attribute.name, attribute.title, value)
			} else if attribute.attributeType == AttributeTypeInt64SliceType {
				value := attribute.defaultValue.([]int64)
				row = htmlutils.NewInt64SliceField(attribute.title, attribute.name, attribute.title, value)
			} else if attribute.attributeType == AttributeTypeFloat64Type {
				value := attribute.defaultValue.(float64)
				row = htmlutils.NewFloat64Field(attribute.title, attribute.name, attribute.title, value)
			}
		} else {
			if attribute.attributeType == AttributeTypeStringType {
				row = htmlutils.NewStringHiddenField(attribute.name, object.GetValueForAttributeName(attribute.name).(string))
			} else if attribute.attributeType == AttributeTypeInt64Type {
				row = htmlutils.NewInt64HiddenField(attribute.name, object.GetValueForAttributeName(attribute.name).(int64))
			} else if attribute.attributeType == AttributeTypeInt64SliceType {
				row = htmlutils.NewInt64SliceHiddenField(attribute.title, attribute.name, attribute.title, object.GetValueForAttributeName(attribute.name).([]int64))
			} else if attribute.attributeType == AttributeTypeFloat64Type {
				// no hidden float field yet, need implement in htmlutils
			}
		}
		if row != nil {
			rows = append(rows, row)
		}
	}

	editObject := htmlutils.NewEditObject(rows, object.GetCreateUrl())
	return editObject
}

func GetHTMLForEditObjectForm(object Object) *htmlutils.EditObject {
	rows := make([]*htmlutils.EditEntry, 0)
	for _, attribute := range object.GetAttributes() {
		var row *htmlutils.EditEntry
		if attribute.isEditable {
			if attribute.attributeType == AttributeTypeStringType {
				row = htmlutils.NewStringField(attribute.title, attribute.name, attribute.title, object.GetValueForAttributeName(attribute.name).(string))
			} else if attribute.attributeType == AttributeTypeInt64Type {
				row = htmlutils.NewInt64Field(attribute.title, attribute.name, attribute.title, object.GetValueForAttributeName(attribute.name).(int64))
			} else if attribute.attributeType == AttributeTypeInt64SliceType {
				row = htmlutils.NewInt64SliceField(attribute.title, attribute.name, attribute.title, object.GetValueForAttributeName(attribute.name).([]int64))
			} else if attribute.attributeType == AttributeTypeFloat64Type {
				row = htmlutils.NewFloat64Field(attribute.title, attribute.name, attribute.title, object.GetValueForAttributeName(attribute.name).(float64))
			}
		} else {
			if attribute.attributeType == AttributeTypeStringType {
				row = htmlutils.NewStringHiddenField(attribute.name, object.GetValueForAttributeName(attribute.name).(string))
			} else if attribute.attributeType == AttributeTypeInt64Type {
				row = htmlutils.NewInt64HiddenField(attribute.name, object.GetValueForAttributeName(attribute.name).(int64))
			} else if attribute.attributeType == AttributeTypeInt64SliceType {
				row = htmlutils.NewInt64SliceHiddenField(attribute.title, attribute.name, attribute.title, object.GetValueForAttributeName(attribute.name).([]int64))
			} else if attribute.attributeType == AttributeTypeFloat64Type {
				// no hidden float field yet, need implement in htmlutils
			}
		}
		if row != nil {
			rows = append(rows, row)
		}
	}

	rowId := htmlutils.NewInt64HiddenField("id", object.Id())
	rows = append(rows, rowId)
	editObject := htmlutils.NewEditObject(rows, object.GetUpdateUrl())
	return editObject
}

func GetHTMLForDisplayObjects(objects []Object) template.HTML {
	headers := make([]string, 0)
	headers = append(headers, "ID")
	if len(objects) == 0 {
		return ""
	}
	sampleObject := objects[0]
	for _, attribute := range sampleObject.GetAttributes() {
		headers = append(headers, attribute.title)
	}
	headers = append(headers, "Action")
	headers = append(headers, "")

	columns := make([][]*htmlutils.TableColumn, 0)
	for _, object := range objects {
		row := make([]*htmlutils.TableColumn, 0)
		columnId := htmlutils.NewStringTableColumn(fmt.Sprintf("%d", object.Id()))
		row = append(row, columnId)
		for _, attribute := range object.GetAttributes() {
			var column *htmlutils.TableColumn
			if attribute.attributeType == AttributeTypeStringType {
				column = htmlutils.NewStringTableColumn(object.GetValueForAttributeName(attribute.name).(string))
			} else if attribute.attributeType == AttributeTypeFloat64Type {
				column = htmlutils.NewStringTableColumn(fmt.Sprintf("%.5f", object.GetValueForAttributeName(attribute.name).(float64)))
			} else if attribute.attributeType == AttributeTypeInt64Type {
				column = htmlutils.NewStringTableColumn(utils.FormatWithComma(object.GetValueForAttributeName(attribute.name).(int64)))
			} else if attribute.attributeType == AttributeTypeInt64SliceType {
				stringSlice := make([]string, 0)
				int64Slice := object.GetValueForAttributeName(attribute.name).([]int64)
				for _, intValue := range int64Slice {
					stringSlice = append(stringSlice, fmt.Sprintf("%d", intValue))
				}
				joinString := strings.Join(stringSlice, ",")
				column = htmlutils.NewStringTableColumn(joinString)
			}

			if column != nil {
				row = append(row, column)
			}
		}

		editColumn := htmlutils.NewActionTableColumn("primary",
			"Edit", object.GetUpdateUrl())
		row = append(row, editColumn)

		if object.GetDeleteUrl() != "" {
			deleteColumn := htmlutils.NewActionTableColumn("danger",
				"Delete", object.GetDeleteUrl())
			row = append(row, deleteColumn)
		}
		columns = append(columns, row)
	}
	table := htmlutils.NewTableObject(headers, columns)
	return table.SerializedData()
}

func DeleteObject(db datacenter.DBInterface, object Object) (err error) {
	fmt.Println("delete", object.Id())
	queryString := fmt.Sprintf("DELETE FROM %s WHERE id = $1", object.DatabaseName())
	_, err = db.Exec(queryString, object.Id())
	if err != nil {
		return err
	}
	return nil
}

func CreateObject(db datacenter.DBInterface, object Object, request *http.Request) (resultObject Object, err error) {
	data := GetHTMLForCreateObjectForm(object).ConvertRequestToData(request)
	var paramsString string
	var paramsNumberString string
	var params []interface{}
	params = make([]interface{}, 0)

	for index, attribute := range object.GetAttributes() {
		if index == 0 {
			paramsString = attribute.name
			paramsNumberString = fmt.Sprintf("$%d", index+1)
		} else {
			paramsString = fmt.Sprintf("%s, %s", paramsString, attribute.name)
			paramsNumberString = fmt.Sprintf("%s, $%d", paramsNumberString, index+1)
		}
		value := utils.GetDataAtPath(data, attribute.name)
		if attribute.attributeType == AttributeTypeInt64SliceType {
			rawString := utils.ConvertInt64SliceToRawString(value.([]int64))
			params = append(params, rawString)
		} else {
			params = append(params, value)
		}
		object.SetValueForAttributeName(attribute.name, value)
	}

	queryString := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING id", object.DatabaseName(), paramsString, paramsNumberString)

	row := db.QueryRow(queryString, params...)
	var id int64
	err = row.Scan(&id)
	if err != nil {
		return nil, err
	}
	object.SetId(id)
	return object, nil
}
