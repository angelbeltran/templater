package templater

import (
	"html/template"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yosssi/gohtml"
)

func TestTemplater(t *testing.T) {
	tmpl := new(Templater).With(Config{
		Funcs: func(name string, props map[string]any) template.FuncMap {
			return template.FuncMap{}
		},
		Dirs: DirsConfig{
			Base:       "",
			Pages:      "",
			Components: "",
		},
		FileExt: "",
	})

	_ = tmpl
}

func TestTemplater_ExecuteComponent(t *testing.T) {
	type (
		Args struct {
			Config Config
			Name   string
			KVs    []any
		}
		Expected struct {
			Bytes string
			Error error
		}
		Test struct {
			Name     string
			Args     Args
			Expected Expected
		}
	)

	tests := []Test{
		{
			Name: "Given a component " +
				"With simple props " +
				"Then the component is rendered",
			Args: Args{
				Config: Config{
					Dirs: DirsConfig{
						Base:       "test_dir/test_templates",
						Pages:      "test_pages",
						Components: "test_components",
					},
				},
				Name: "component_1",
				KVs: []any{
					"X", "abc",
					"Y", 123,
					"Z", true,
				},
			},
			Expected: Expected{
				Bytes: `<div>
  <div>
    abc
  </div>
  <div>
    123
  </div>
  <div>
    true
  </div>
</div>`,
			},
		},
		{
			Name: "Given a component " +
				"With a nested component " +
				"With simple props " +
				"Then the component is rendered",
			Args: Args{
				Config: Config{
					Dirs: DirsConfig{
						Base:       "test_dir/test_templates",
						Pages:      "test_pages",
						Components: "test_components",
					},
				},
				Name: "component_2",
				KVs: []any{
					"A", "abc",
					"B", 123,
					"C", true,
				},
			},
			Expected: Expected{
				Bytes: `<div>
  <div>
    HEADER
  </div>
  <div>
    abc
  </div>
  <div>
    123
  </div>
  <div>
    true
  </div>
  <div>
    <div>
      abc
    </div>
    <div>
      123
    </div>
    <div>
      true
    </div>
  </div>
</div>`,
			},
		},
		{
			Name: "Given a component " +
				"With path parameters " +
				"With a nested component " +
				"With simple props " +
				"Then the component is rendered " +
				"And the path params are include as props",
			Args: Args{
				Config: Config{
					Dirs: DirsConfig{
						Base:       "test_dir/test_templates",
						Pages:      "test_pages",
						Components: "test_components",
					},
				},
				Name: "top_dir/some-phrase/mid_dir/321/bottom_dir/last-part",
				KVs: []any{
					"Q", "QQQQ",
					"R", 7777,
					"S", "SSSS",
				},
			},
			Expected: Expected{
				Bytes: `<div>
  <div>
    QQQQ
  </div>
  <div>
    7777
  </div>
  <div>
    SSSS
  </div>
  <div>
    <div>
      HEADER
    </div>
    <div>
      some-phrase
    </div>
    <div>
      321
    </div>
    <div>
      last-part
    </div>
    <div>
      <div>
        abc
      </div>
      <div>
        123
      </div>
      <div>
        true
      </div>
    </div>
  </div>
</div>`,
			},
		},
		{
			Name: "Given a component " +
				"With a path param " +
				"Then the component is rendered",
			Args: Args{
				Config: Config{
					Dirs: DirsConfig{
						Base:       "test_dir/test_templates",
						Pages:      "test_pages",
						Components: "test_components",
					},
				},
				Name: "58",
				KVs:  []any{},
			},
			Expected: Expected{
				Bytes: `<div>
  byte: 58
</div>`,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			tm := new(Templater).With(test.Args.Config)

			b, err := tm.ExecuteComponent(test.Args.Name, test.Args.KVs...)

			if test.Expected.Error == nil {
				require.NoError(t, err, "unexpected error returned: %+v", err)
				assert.Equal(t, test.Expected.Bytes, gohtml.Format(string(b)), "unexpected bytes returned")
			} else {
				assert.Equalf(t, test.Expected.Error, err, "unexpected error returned: %+v", err)
			}
		})
	}
}

func TestTemplater_ExecutePage(t *testing.T) {
	type (
		Args struct {
			Config Config
			Name   string
			KVs    []any
		}
		Expected struct {
			Bytes string
			Error error
		}
		Test struct {
			Name     string
			Args     Args
			Expected Expected
		}
	)

	tests := []Test{
		{
			Name: "Given a page " +
				"Then the page is rendered",
			Args: Args{
				Config: Config{
					Dirs: DirsConfig{
						Base:       "test_dir/test_templates",
						Pages:      "test_pages",
						Components: "test_components",
					},
				},
				Name: "simple_page",
			},
			Expected: Expected{
				Bytes: `<!DOCTYPE html>
<html>
  <head>
    <title>
      ABC
    </title>
  </head>
  <body>
    <header>
      HEAD
    </header>
    <div>
      TEST
    </div>
    <footer>
      FOOTER
    </footer>
  </body>
</html>`,
			},
		},
		{
			Name: "Given a page " +
				"With a path param " +
				"Then the page is rendered",
			Args: Args{
				Config: Config{
					Dirs: DirsConfig{
						Base:       "test_dir/test_templates",
						Pages:      "test_pages",
						Components: "test_components",
					},
				},
				Name: "true",
			},
			Expected: Expected{
				Bytes: `<!DOCTYPE html>
<html>
  <head>
    <title>
      ABC
    </title>
  </head>
  <body>
    <header>
      HEAD
    </header>
    <div>
      true or false: true
    </div>
    <footer>
      FOOTER
    </footer>
  </body>
</html>`,
			},
		},
		{
			Name: "Given a page " +
				"With path parameters " +
				"With a nested component " +
				"With simple props " +
				"Then the page is rendered " +
				"And the path params are include as props",
			Args: Args{
				Config: Config{
					Dirs: DirsConfig{
						Base:       "test_dir/test_templates",
						Pages:      "test_pages",
						Components: "test_components",
					},
				},
				Name: "top_dir/asdfasdfasdf/the_page",
				KVs: []any{
					"A", "AAA",
					"B", "BBB",
					"C", "CCC",
				},
			},
			Expected: Expected{
				Bytes: `<!DOCTYPE html>
<html>
  <head>
    <title>
      ABC
    </title>
  </head>
  <body>
    <header>
      HEAD
    </header>
    <div>
      <div>
        <div>
          <div>
            QQQQ
          </div>
          <div>
            7777
          </div>
          <div>
            SSSS
          </div>
          <div>
            <div>
              HEADER
            </div>
            <div>
              some-phrase
            </div>
            <div>
              321
            </div>
            <div>
              last-part
            </div>
            <div>
              <div>
                abc
              </div>
              <div>
                123
              </div>
              <div>
                true
              </div>
            </div>
          </div>
        </div>
      </div>
      <div>
        <div>
          <div>
            HEADER
          </div>
          <div>
            123
          </div>
          <div>
            abc
          </div>
          <div>
            true
          </div>
          <div>
            <div>
              abc
            </div>
            <div>
              123
            </div>
            <div>
              true
            </div>
          </div>
        </div>
      </div>
      <div>
        AAA
      </div>
      <div>
        BBB
      </div>
      <div>
        CCC
      </div>
    </div>
    <footer>
      FOOTER
    </footer>
  </body>
</html>`,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			tm := new(Templater).With(test.Args.Config)

			b, err := tm.ExecutePage(test.Args.Name, test.Args.KVs...)

			if test.Expected.Error == nil {
				require.NoError(t, err, "unexpected error returned: %+v", err)
				assert.Equal(t, test.Expected.Bytes, gohtml.Format(string(b)), "unexpected bytes returned")
			} else {
				assert.Equalf(t, test.Expected.Error, err, "unexpected error returned: %+v", err)
			}
		})
	}
}
