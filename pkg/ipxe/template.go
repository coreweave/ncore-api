package ipxe

import (
	"log"
	"path/filepath"
	"text/template"
)

func GetIpxeConfigTemplate(ipxeTemplateFile string) template.Template {
	t, err := template.New(filepath.Base(ipxeTemplateFile)).ParseFiles(ipxeTemplateFile)
	if err != nil {
		log.Printf("GetIpxeConfigTemplate error parsing template: %v", err)
	}
	return *t
}
