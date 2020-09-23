/**
 * auto generated by https://github.com/gookit/Kite
 *
 * @author https://github.com/inhere
 */
package {{ .PkgName }}

import (
	"github.com/gookit/rux"
)

// {{ upFirst .GroupName }} {{ .GroupDesc }}
type {{ upFirst .GroupName }} struct {}

// AddRoutes register routes to the router
func (grp *{{ .GroupName }}) AddRoutes(r *rux.Router) { {{range $i, $m := .Actions }}
	r.{{ $m.METHOD }}("{{ $m.Path }}", grp.{{ $m.MethodName }}){{end}}
}
{{range $i, $m := .Actions }}
// {{ $m.MethodName }} {{ $m.MethodDesc }}{{ $m.TagComments }}
func (*{{ upFirst $.GroupName}}) {{ $m.MethodName }}(c *rux.Context) {
	c.Text(200, "hello")
}
{{end}}