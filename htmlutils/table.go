package htmlutils

import (
	"bytes"
	"fmt"
	"github.com/vic/vic_go/log"
	"html/template"
)

type TableObject struct {
	headers []string
	columns [][]*TableColumn
}

func NewTableObject(headers []string, columns [][]*TableColumn) *TableObject {
	return &TableObject{
		headers: headers,
		columns: columns,
	}
}

func (table *TableObject) SerializedData() template.HTML {
	data := make(map[string]interface{})
	data["headers"] = table.headers
	data["columns"] = table.columns
	tmpl, err := template.New("").Parse(`
	<table class="table">
		<tr>
			{{ range $index, $element := .headers }}
			<th>{{$element}}</th>
			{{ end }}
		</tr>
		{{range $index, $element := .columns}}
		<tr>
			{{range $indexCol, $elementCol := $element}}
			<td>{{$elementCol.SerializedData}}</td>
			{{end}}
			<td>
		</tr>
		{{end}}
	</table>
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

type TableColumn struct {
	columnType  string
	value       string
	buttonName  string
	buttonClass string
}

func NewTableColumn(columnType, value string) *TableColumn {
	return &TableColumn{
		columnType: columnType,
		value:      value,
	}
}

func NewStringTableColumn(value string) *TableColumn {
	return &TableColumn{
		columnType: kStringColumn,
		value:      value,
	}
}

func NewImageTableColumn(url string) *TableColumn {
	return &TableColumn{
		columnType: kImageColumn,
		value:      url,
	}
}

func NewRawHtmlTableColumn(html string) *TableColumn {
	return &TableColumn{
		columnType: kHTMLColumn,
		value:      html,
	}
}

func NewActionTableColumn(buttonClass string, buttonName string, link string) *TableColumn {
	return &TableColumn{
		columnType:  kActionColumn,
		buttonName:  buttonName,
		buttonClass: buttonClass,
		value:       link,
	}
}

func (column *TableColumn) SerializedData() template.HTML {
	if column.columnType == kStringColumn {
		return template.HTML(column.value)
	} else if column.columnType == kImageColumn {
		return template.HTML(fmt.Sprintf("<img src='%s' width='100px'/>", column.value))
	} else if column.columnType == kActionColumn {
		return template.HTML(fmt.Sprintf("<a class='btn btn-%s' href='%s'>%s</a>", column.buttonClass, column.value, column.buttonName))
	} else if column.columnType == kHTMLColumn {
		return template.HTML(column.value)
	}
	return ""
}
