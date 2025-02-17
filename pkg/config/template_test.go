package config

import "testing"

func TestTemplateRender(t *testing.T) {
	tmplCtx := map[string]any{
		"SomeNumber": 1234,
	}
	tests := []struct {
		name string
		tmpl *Template
		want string
	}{
		{
			name: "nil",
			tmpl: nil,
			want: "",
		},
		{
			name: "raw string",
			tmpl: MustTemplate("foo bar\nmoo doo"),
			want: "foo bar\nmoo doo",
		},
		{
			name: "use context",
			tmpl: MustTemplate("the number {{.SomeNumber}} is my favorite"),
			want: "the number 1234 is my favorite",
		},
		{
			name: "use func",
			tmpl: MustTemplate("i dont like the number {{add .SomeNumber 500}}."),
			want: "i dont like the number 1734.",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.tmpl.Render(tmplCtx)
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Errorf("wrong result\nwant: %q\ngot:  %q", test.want, got)
			}
		})
	}
}

func TestNewTemplateContextForPackage(t *testing.T) {
	tests := []struct {
		name string
		pkg  Package
		want TemplateContext
	}{
		{
			name: "empty",
			pkg:  Package{},
			want: TemplateContext{},
		},
		{
			name: "only package name",
			pkg:  Package{Name: "my-pkg"},
			want: TemplateContext{Package: "my-pkg"},
		},
		{
			name: "only package desc",
			pkg:  Package{Description: MustTemplate("Pkg `{{.Package}}` desc")},
			want: TemplateContext{PackageDescription: "Pkg `` desc"},
		},
		{
			name: "package name and desc",
			pkg: Package{
				Name:        "my-pkg",
				Description: MustTemplate("Pkg `{{.Package}}` desc"),
			},
			want: TemplateContext{
				Package:            "my-pkg",
				PackageDescription: "Pkg `my-pkg` desc",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := NewTemplateContextForPackage(test.pkg)
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Errorf("wrong result\nwant: %#v\ngot:  %#v", test.want, got)
			}
		})
	}
}
