package ipxe

import (
	"os"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/Masterminds/sprig"
)

func TestIpxe_GetIpxeConfigTemplate(t *testing.T) {
	ipxeTemplateFile := "../../pkg/ipxe/template_iscsi_test.ipxe"
	tmpl := template.Must(template.New(filepath.Base(ipxeTemplateFile)).Funcs(sprig.FuncMap()).ParseFiles(ipxeTemplateFile))
	tmpl.Execute(os.Stdout, "test")
	t.Logf("template: %v", tmpl)
}
