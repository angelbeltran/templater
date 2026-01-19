# Templater

A go package that provides HTML template *components*, built on the html/template package.

A *component* is an html template that can be used by any other component or any *page* template.

Unlike the standard practice of compiling templates, compiling template dependencies first, then top-level templates, no template compilation is required. All compilation is done at runtime.

This has the downside of performance costs and error risks due to runtime compilation,
but has the upside of the ability to modify templates at runtime to gain immediate feedback
and allowing any template to import any other template.


## Example

```golang
package main

import (
    "net/http"

    "github.com/angelbeltran/templater"
)

func main() {
    tmpl := new(Templater).Config(templater.Config{
        Dirs: templater.DirConfig {
            Base: "./templates", // templates directory holding all html templates
            Pages: "pages",
            Components: "components",
        },
    })

    http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
        b, err := tmpl.ExecutePage(r.URL.Path)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        } else {
            w.Write(b)
        }
    })

    http.ListenAndServe(":8000", nil)
}
```

**Directories**

```
/templates
    layout.html.tmpl
    /pages
        /home.html.tmpl  # page templates
    /components
        /my-table.html.tmpl  # component templates
```

**Templates**

*layout.html.tmpl*
```html
<!DOCTYPE html>
<html>
    <head>
        <title>The Title</title>
    </head>
    <body>
        <header>
            Welcome!
        </header>
        <main>
            {{ define "block" . }}{{ end }}
        </main>
        <footer>
            Contact us! ...
        </footer>
</html>
```

*home.html.tmpl*
```html
<div>
    Home page!

    {{ component "my-table" "Prop1" "abc123" }}
</div>
```

*my-table.html.tmpl*
```html
<table>
    <tr>
        <th>
            Prop 1
        </th>
    </tr>
    <tr>
        <td>
            {{ .Prop1 }}
        </td>
    </tr>
</table>
```


## Key Features
- serve webpages from template
- easy template composition
- minimal setup
- runtime template editing

## Additional Features
- path parameters, eg `/pets/{name}`
    - type parameter support: eg `/store/{storeID.int}`,
    - automatic injection into templates via "PathParams" argument `<div>Store ID: {{ .PathParams.storeID }}</div>`
- configurable
    - template directories
    - template file extensions
    - template functions
- index.html / index.html.tmpl support
