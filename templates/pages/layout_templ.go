// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.771
// SPDX-FileCopyrightText: 2022 Risk.Ident GmbH <contact@riskident.com>

//

// SPDX-License-Identifier: GPL-3.0-or-later

//

// This program is free software: you can redistribute it and/or modify it

// under the terms of the GNU General Public License as published by the

// Free Software Foundation, either version 3 of the License, or

// (at your option) any later version.

//

// This program is distributed in the hope that it will be useful, but WITHOUT

// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or

// FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for

// more details.

//

// You should have received a copy of the GNU General Public License along

// with this program.  If not, see <http://www.gnu.org/licenses/>.

package pages

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

func Layout(title string) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<!doctype html><html lang=\"en\"><head><title>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if title == "Jelease" || title == "" {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("Jelease")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		} else {
			var templ_7745c5c3_Var2 string
			templ_7745c5c3_Var2, templ_7745c5c3_Err = templ.JoinStringErrs(title)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `templates/pages/layout.templ`, Line: 28, Col: 12}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var2))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(" - Jelease")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</title><meta charset=\"UTF-8\"><meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\"><meta http-equiv=\"X-UA-Compatible\" content=\"ie=edge\"><link rel=\"apple-touch-icon\" sizes=\"180x180\" href=\"/apple-touch-icon.png\"><link rel=\"icon\" type=\"image/png\" sizes=\"32x32\" href=\"/favicon-32x32.png\"><link rel=\"icon\" type=\"image/png\" sizes=\"16x16\" href=\"/favicon-16x16.png\"><link rel=\"manifest\" href=\"/site.webmanifest\"><link rel=\"mask-icon\" href=\"/safari-pinned-tab.svg\" color=\"#5bbad5\"><meta name=\"apple-mobile-web-app-title\" content=\"Jelease\"><meta name=\"application-name\" content=\"Jelease\"><meta name=\"msapplication-TileColor\" content=\"#4ecbdd\"><meta name=\"theme-color\" content=\"#ffffff\"><link rel=\"stylesheet\" href=\"https://unpkg.com/papercss@1.9.1/dist/paper.min.css\" integrity=\"sha384-xmINuyCPKMw/MdIfiUNHXvyZesszhJcD4A7OmXnQOCbcoV+V1lSd7Xx70OfMpX4f\" crossorigin=\"anonymous\" referrerpolicy=\"no-referrer\"><link rel=\"stylesheet\" href=\"https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.8.0/styles/github.min.css\" integrity=\"sha512-0aPQyyeZrWj9sCA46UlmWgKOP0mUipLQ6OZXu8l4IcAmD2u31EPEy9VcIMvl7SoAaKe8bLXZhYoMaE/in+gcgA==\" crossorigin=\"anonymous\" referrerpolicy=\"no-referrer\"></head><body><main class=\"paper container margin-top\"><nav class=\"border split-nav margin-bottom\"><div class=\"nav-brand\"><h3><a href=\"/\">Jelease</a></h3></div><div class=\"collapsible\"><input id=\"collapsible-nav\" type=\"checkbox\" name=\"collapsible-nav\"> <label for=\"collapsible-nav\"><div class=\"bar1\"></div><div class=\"bar2\"></div><div class=\"bar3\"></div></label><div class=\"collapsible-body\"><ul class=\"inline\"><li><a href=\"/packages\">Packages</a></li><li><a href=\"/config\">Config</a></li><li><a href=\"https://github.com/RiskIdent/jelease\" target=\"_blank\">Github</a></li></ul></div></div></nav><article class=\"margin-bottom\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templ_7745c5c3_Var1.Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</article><hr><p class=\"text-muted text-center\">Risk.Ident GmbH | Created by the Platform team</p></main><script src=\"https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.8.0/highlight.min.js\" integrity=\"sha512-rdhY3cbXURo13l/WU9VlaRyaIYeJ/KBakckXIvJNAQde8DgpOmE+eZf7ha4vdqVjTtwQt69bD2wH2LXob/LB7Q==\" crossorigin=\"anonymous\" referrerpolicy=\"no-referrer\"></script><script src=\"https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.8.0/languages/yaml.min.js\" integrity=\"sha512-uzLMr+y2UfIJhZodXJJXGUgSWhOTdT1FABVEECjTZ8rPNQ5mT8AJUldpJVPnxUYjT052sB8ddJwiB56MtAQA3g==\" crossorigin=\"anonymous\" referrerpolicy=\"no-referrer\"></script><script src=\"https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.8.0/languages/diff.min.js\" integrity=\"sha512-5JQbFkPWbNPf+CE8ImQrqzCQ58zMzEV+mhYAzZikMIIjTn7WvWem3PP2BuIMeRezXObaVtFnfoUrkL07zvXKdQ==\" crossorigin=\"anonymous\" referrerpolicy=\"no-referrer\"></script><script src=\"https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.8.0/languages/markdown.min.js\" integrity=\"sha512-X9tsqdwNQPU9yGZHcHgZjGUQhify0RLqbPchbFiFzIOh1VKG1VEV3CeASutshbPf5bkSgj+hWBN4c2YcUXZ69w==\" crossorigin=\"anonymous\" referrerpolicy=\"no-referrer\"></script><script>hljs.highlightAll();</script></body></html>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}

var _ = templruntime.GeneratedTemplate
