// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.707
package templates

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"
import "bytes"

func About() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, templ_7745c5c3_W io.Writer) (templ_7745c5c3_Err error) {
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templ_7745c5c3_W.(*bytes.Buffer)
		if !templ_7745c5c3_IsBuffer {
			templ_7745c5c3_Buffer = templ.GetBuffer()
			defer templ.ReleaseBuffer(templ_7745c5c3_Buffer)
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<p>About page content updated</p><div class=\"p-10\"><div class=\"dropdown inline-block relative\"><button class=\"bg-gray-300 text-gray-700 font-semibold py-2 px-4 rounded inline-flex items-center\"><span class=\"mr-1\">Dropdown</span> <svg class=\"fill-current h-4 w-4\" xmlns=\"http://www.w3.org/2000/svg\" viewBox=\"0 0 20 20\"><path d=\"M9.293 12.95l.707.707L15.657 8l-1.414-1.414L10 10.828 5.757 6.586 4.343 8z\"></path></svg></button><ul class=\"dropdown-menu absolute hidden text-gray-700 pt-1\"><li class=\"\"><a class=\"rounded-t bg-gray-200 hover:bg-gray-400 py-2 px-4 block whitespace-no-wrap\" href=\"#\">One</a></li><li class=\"\"><a class=\"bg-gray-200 hover:bg-gray-400 py-2 px-4 block whitespace-no-wrap\" href=\"#\">Two</a></li><li class=\"\"><a class=\"rounded-b bg-gray-200 hover:bg-gray-400 py-2 px-4 block whitespace-no-wrap\" href=\"#\">Three is the magic number</a></li></ul></div></div>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if !templ_7745c5c3_IsBuffer {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteTo(templ_7745c5c3_W)
		}
		return templ_7745c5c3_Err
	})
}
