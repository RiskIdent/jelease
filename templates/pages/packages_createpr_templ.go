// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.707
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
import "context"
import "io"
import "bytes"

import (
	"fmt"

	"github.com/RiskIdent/jelease/pkg/config"
	"github.com/RiskIdent/jelease/pkg/github"
	"github.com/RiskIdent/jelease/templates/components"
)

type PackagesCreatePRModel struct {
	Config       *config.Config
	Package      config.Package
	PullRequests []github.PullRequest
	DryRun       bool
	Version      string
	JiraIssue    string
	IsPost       bool
	Error        error
}

func PackagesCreatePR(model PackagesCreatePRModel) templ.Component {
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
		templ_7745c5c3_Var2 := templ.ComponentFunc(func(ctx context.Context, templ_7745c5c3_W io.Writer) (templ_7745c5c3_Err error) {
			templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templ_7745c5c3_W.(*bytes.Buffer)
			if !templ_7745c5c3_IsBuffer {
				templ_7745c5c3_Buffer = templ.GetBuffer()
				defer templ.ReleaseBuffer(templ_7745c5c3_Buffer)
			}
			templ_7745c5c3_Var3 := templ.ComponentFunc(func(ctx context.Context, templ_7745c5c3_W io.Writer) (templ_7745c5c3_Err error) {
				templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templ_7745c5c3_W.(*bytes.Buffer)
				if !templ_7745c5c3_IsBuffer {
					templ_7745c5c3_Buffer = templ.GetBuffer()
					defer templ.ReleaseBuffer(templ_7745c5c3_Buffer)
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<li>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = components.Linkf("Packages", "/packages").Render(ctx, templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</li><li>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = components.Linkf(model.Package.Name, "/packages/%s", model.Package.NormalizedName()).Render(ctx, templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</li><li>Create PR</li>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				if !templ_7745c5c3_IsBuffer {
					_, templ_7745c5c3_Err = io.Copy(templ_7745c5c3_W, templ_7745c5c3_Buffer)
				}
				return templ_7745c5c3_Err
			})
			templ_7745c5c3_Err = components.Breadcrumbs().Render(templ.WithChildren(ctx, templ_7745c5c3_Var3), templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(" <h2>Create PR for ")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var4 string
			templ_7745c5c3_Var4, templ_7745c5c3_Err = templ.JoinStringErrs(model.Package.Name)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `templates/pages/packages_createpr.templ`, Line: 47, Col: 40}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var4))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</h2>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if model.IsPost && model.Error != nil {
				templ_7745c5c3_Err = components.AlertDangerErr("Unexpected issue while loading page:", model.Error).Render(ctx, templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(" <section><h3>Create GitHub PR</h3><form method=\"POST\" action=\"\" id=\"create-pr-form\"><div class=\"row margin-bottom-none\"><div class=\"col sm-6 md-4 padding-small\"><div class=\"form-group\"><label for=\"form-version\" title=\"* = Required\">Version <strong class=\"text-danger\">*</strong></label> <input type=\"text\" name=\"version\" placeholder=\"Example: v1.0.0\" id=\"form-version\" required value=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var5 string
			templ_7745c5c3_Var5, templ_7745c5c3_Err = templ.JoinStringErrs(model.Version)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `templates/pages/packages_createpr.templ`, Line: 60, Col: 119}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var5))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\"></div><div class=\"form-group margin-bottom-none\"><label for=\"form-jira-issue\">Jira issue</label> <input type=\"text\" name=\"jiraIssue\" class=\"border-2\" placeholder=\"Example: OP-1234\" id=\"form-jira-issue\" value=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var6 string
			templ_7745c5c3_Var6, templ_7745c5c3_Err = templ.JoinStringErrs(model.JiraIssue)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `templates/pages/packages_createpr.templ`, Line: 64, Col: 135}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var6))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\"></div></div><div class=\"col sm-6 md-8 padding-small\"><p class=\"margin-top-none\">Settings</p>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var7 = []any{"form-group", "margin-bottom-none", templ.KV("disabled", model.Config.DryRun)}
			templ_7745c5c3_Err = templ.RenderCSSItems(ctx, templ_7745c5c3_Buffer, templ_7745c5c3_Var7...)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<fieldset class=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var8 string
			templ_7745c5c3_Var8, templ_7745c5c3_Err = templ.JoinStringErrs(templ.CSSClasses(templ_7745c5c3_Var7).String())
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `templates/pages/packages_createpr.templ`, Line: 1, Col: 0}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var8))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\"><label class=\"paper-switch-2\"><input id=\"form-pr-create\" name=\"prCreate\" class=\"border-3\" type=\"checkbox\" value=\"true\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if model.Config.DryRun {
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(" disabled")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(" else")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if !model.DryRun {
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(" checked")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("> <span class=\"paper-switch-slider\"></span></label> <label for=\"form-pr-create\" class=\"paper-switch-2-label\">Create PR ")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if model.Config.DryRun {
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("(forced \"dry run\" via config)")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			} else {
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("(otherwise will do a \"dry run\")")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</label></fieldset></div><div class=\"col padding-small\"><div class=\"form-group margin-bottom-none\"><p class=\"margin-top-none\"></p><button type=\"submit\" class=\"border-5\" id=\"submit\">Submit</button></div></div></div></form></section><section><h3>Created PRs</h3>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = createPRResults(prResults{
				IsPost:       model.IsPost,
				DryRun:       model.DryRun,
				Error:        model.Error,
				PullRequests: model.PullRequests,
			}).Render(ctx, templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</section><script>\n\t\t(function() {\n\t\t\tconst form = document.getElementById(\"create-pr-form\");\n\t\t\tform.addEventListener(\"submit\", function() {\n\t\t\t\tdocument.getElementById(\"results\").innerHTML = `<p><em class=\"text-muted\">Processing, please wait...</em></p>`;\n\t\t\t\tdocument.getElementById(\"submit\").setAttribute('disabled', 'disabled');\n\n\t\t\t\t// Some magic to make URL query params match the POSTed form data\n\t\t\t\tconst url = new URL(location);\n\n\t\t\t\t// Need to Array.from(), as without that this becomes glitchy,\n\t\t\t\t// because we're removing items at the same time we're iterating it.\n\t\t\t\tconst searchParamKeys = Array.from(url.searchParams.keys());\n\t\t\t\tfor (const key of searchParamKeys) {\n\t\t\t\t\turl.searchParams.delete(key);\n\t\t\t\t}\n\t\t\t\tfor (const [key, value] of new FormData(form).entries()) {\n\t\t\t\t\tif (typeof value === \"string\" && value !== \"\") {\n\t\t\t\t\t\turl.searchParams.set(key, value);\n\t\t\t\t\t}\n\t\t\t\t}\n\t\t\t\tconsole.log(\"Pushing state:\", url);\n\t\t\t\thistory.pushState({}, \"\", url);\n\t\t\t});\n\t\t})();\n\t\t</script>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if !templ_7745c5c3_IsBuffer {
				_, templ_7745c5c3_Err = io.Copy(templ_7745c5c3_W, templ_7745c5c3_Buffer)
			}
			return templ_7745c5c3_Err
		})
		templ_7745c5c3_Err = Layout("Create PR").Render(templ.WithChildren(ctx, templ_7745c5c3_Var2), templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if !templ_7745c5c3_IsBuffer {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteTo(templ_7745c5c3_W)
		}
		return templ_7745c5c3_Err
	})
}

type prResults struct {
	IsPost       bool
	Error        error
	DryRun       bool
	PullRequests []github.PullRequest
}

func createPRResults(model prResults) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, templ_7745c5c3_W io.Writer) (templ_7745c5c3_Err error) {
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templ_7745c5c3_W.(*bytes.Buffer)
		if !templ_7745c5c3_IsBuffer {
			templ_7745c5c3_Buffer = templ.GetBuffer()
			defer templ.ReleaseBuffer(templ_7745c5c3_Buffer)
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var9 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var9 == nil {
			templ_7745c5c3_Var9 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div id=\"results\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if !model.IsPost {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<p><em class=\"text-muted\">The results will be shown here, after you press \"Submit\".</em></p>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		} else {
			if model.Error != nil {
				templ_7745c5c3_Err = components.AlertDangerErr("There was an error when creating the Pull Request:", model.Error).Render(ctx, templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			} else {
				templ_7745c5c3_Err = createPRResultsNoError(model).Render(ctx, templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</div>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if !templ_7745c5c3_IsBuffer {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteTo(templ_7745c5c3_W)
		}
		return templ_7745c5c3_Err
	})
}

func createPRResultsNoError(model prResults) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, templ_7745c5c3_W io.Writer) (templ_7745c5c3_Err error) {
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templ_7745c5c3_W.(*bytes.Buffer)
		if !templ_7745c5c3_IsBuffer {
			templ_7745c5c3_Buffer = templ.GetBuffer()
			defer templ.ReleaseBuffer(templ_7745c5c3_Buffer)
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var10 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var10 == nil {
			templ_7745c5c3_Var10 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		if model.DryRun {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div class=\"alert alert-secondary\"><p><strong>Success:</strong> Request completed. However, note that <code>dryrun</code> was enabled, so no Pull Requests has actually been created.</p></div>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		} else {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div class=\"alert alert-success\"><p><strong>Success:</strong> Request completed. See the created Pull Requests below.</p></div>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		if len(model.PullRequests) == 0 {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div class=\"alert alert-warning\"><p><strong>Warning:</strong> No Pull Requests were created. Maybe @components.Linkf(\"look over the configuration\", \"/packages/%s\", model.Package.NormalizedName()), to ensure it's correct?</p></div>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		} else {
			for i, pr := range model.PullRequests {
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<section><h4>PR #")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var11 string
				templ_7745c5c3_Var11, templ_7745c5c3_Err = templ.JoinStringErrs(fmt.Sprint(i + 1))
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `templates/pages/packages_createpr.templ`, Line: 182, Col: 31}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var11))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</h4><dl><dt>Title</dt><dd>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				if pr.Title != "" {
					var templ_7745c5c3_Var12 string
					templ_7745c5c3_Var12, templ_7745c5c3_Err = templ.JoinStringErrs(pr.Title)
					if templ_7745c5c3_Err != nil {
						return templ.Error{Err: templ_7745c5c3_Err, FileName: `templates/pages/packages_createpr.templ`, Line: 187, Col: 17}
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var12))
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				} else {
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<em>(missing title)</em>")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</dd><dt>Branches</dt><dd>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				if pr.Base != "" && pr.Head != "" {
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("into <code>")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					var templ_7745c5c3_Var13 string
					templ_7745c5c3_Var13, templ_7745c5c3_Err = templ.JoinStringErrs(pr.Base)
					if templ_7745c5c3_Err != nil {
						return templ.Error{Err: templ_7745c5c3_Err, FileName: `templates/pages/packages_createpr.templ`, Line: 195, Col: 27}
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var13))
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</code> from <code>")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					var templ_7745c5c3_Var14 string
					templ_7745c5c3_Var14, templ_7745c5c3_Err = templ.JoinStringErrs(pr.Head)
					if templ_7745c5c3_Err != nil {
						return templ.Error{Err: templ_7745c5c3_Err, FileName: `templates/pages/packages_createpr.templ`, Line: 195, Col: 57}
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var14))
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</code>")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				} else {
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<em>(missing branch info)</em>")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</dd><dt>URL</dt><dd>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				if pr.URL != "" {
					templ_7745c5c3_Err = components.ExternalLink(pr.URL, pr.URL).Render(ctx, templ_7745c5c3_Buffer)
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				} else if pr.RepoRef.URL != "" {
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<em>(would've been created on repo:")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					var templ_7745c5c3_Var15 string
					templ_7745c5c3_Var15, templ_7745c5c3_Err = templ.JoinStringErrs(" ")
					if templ_7745c5c3_Err != nil {
						return templ.Error{Err: templ_7745c5c3_Err, FileName: `templates/pages/packages_createpr.templ`, Line: 205, Col: 46}
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var15))
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					templ_7745c5c3_Err = components.ExternalLink(pr.RepoRef.URL, pr.RepoRef.URL).Render(ctx, templ_7745c5c3_Buffer)
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(")</em>")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				} else {
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<em>(missing URL)</em>")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</dd><dt>Description</dt><dd>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				if pr.Description != "" {
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<pre><code class=\"language-markdown\">")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					var templ_7745c5c3_Var16 string
					templ_7745c5c3_Var16, templ_7745c5c3_Err = templ.JoinStringErrs(pr.Description)
					if templ_7745c5c3_Err != nil {
						return templ.Error{Err: templ_7745c5c3_Err, FileName: `templates/pages/packages_createpr.templ`, Line: 216, Col: 24}
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var16))
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</code></pre>")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				} else {
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<em>(missing description)</em>")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</dd><dt>Git diff</dt><dd>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				if pr.Commit.Diff != "" {
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<pre><code class=\"language-diff\">")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					var templ_7745c5c3_Var17 string
					templ_7745c5c3_Var17, templ_7745c5c3_Err = templ.JoinStringErrs(pr.Commit.Diff)
					if templ_7745c5c3_Err != nil {
						return templ.Error{Err: templ_7745c5c3_Err, FileName: `templates/pages/packages_createpr.templ`, Line: 226, Col: 24}
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var17))
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</code></pre>")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				} else {
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<em>(missing Git diff)</em>")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</dd></dl></section>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
		}
		if !templ_7745c5c3_IsBuffer {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteTo(templ_7745c5c3_W)
		}
		return templ_7745c5c3_Err
	})
}
