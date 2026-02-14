# Templater

A go package that provides HTML template **components** and **slots**, built on [html/template](https://pkg.go.dev/html/template).


## Example

### Components:
*section.html.tmpl*
```html
<section id="{{.ID}}" style="{{.Style}}">
    <header>
        {{ slot "section-header" }}
    </header>

    {{ slot "section-body" }}
</table>
```

### Component used:
*my-shop.html.tmpl*
```html
<nav>
    <a href="/">Home</a>
    <a href="/contact">Contact</a>
    <a href="/about">About</a>
</nav>

{{/* component slot content */}}
{{ define "my-shop-header" }}
    <h1>My Shop</h1>
    <p>The best prices around</p>
{{ end }}

{{ define "my-shop-body" }}
    <p>
        Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
    </p>
    <p>
        Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
    </p>
{{ end }}

{{/* component used */}}
{{ component "section" "ID" "my-shop" "Style" "background-color: yellow;" "#section-header" "my-shop-header" "#section-body" "my-shop-body" }}
```

### Resulting in
```html
<nav>
    <a href="/">Home</a>
    <a href="/contact">Contact</a>
    <a href="/about">About</a>
</nav>

<section id="my-shop" style="background-color: yellow;">
    <header>
        <h1>My Shop</h1>
        <p>The best prices around</p>
    </header>

    <p>
        Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
    </p>
    <p>
        Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
    </p>
</table>
```


## Concepts

A **component** is an html template that can be used in other templates.

A **slot** is a specified place in a component where arbitrary content can be set by the parent component.



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

Unlike the standard practice of compiling templates, compiling template dependencies first, then top-level templates, no template compilation is required.
All compilation is done at runtime.

This has the downside of performance costs and error risks due to runtime compilation,
but has the upside of being able to modify templates at runtime, obtaining faster feedback,
and allowing any template to import any other template.


## Full Example

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
