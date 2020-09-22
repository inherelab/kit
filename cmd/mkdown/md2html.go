package mkdown

import (
	"bytes"
	"flag"
	"os"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	gmparser "github.com/gomarkdown/markdown/parser"
	"github.com/gookit/color"
	"github.com/gookit/gcli/v2"
	"github.com/russross/blackfriday"
)

// filetypes: [".md", ".markdown", ".mdown"]
type md2html struct {
	cmd *gcli.Command
	// options
	toc   bool
	page  bool
	latex bool

	tocOnly   bool
	fractions bool

	smartyPants bool
	latexDashes bool
	// Sets HTML output to a simple form:
	//  - No head
	//  - No body tags
	//  - ids, classes, and style are stripped out
	//  - Just bare minimum HTML tags and attributes
	//  - extension modifications included
	htmlSimple bool

	css string
	// driver:
	// gm    gomarkdown
	// bf 	 blackfriday
	driver string
	output string
	// "markdown", "github", "gitlab"
	style string
}

const (
	defaultTitle = ""
	// driver names
	driverBF = "bf"
	driverGM = "gm"
)

var (
	drivers = map[string]string{
		"bf": "blackfriday",
		"gm": "gomarkdown",
	}
)

/*
DOC: https://developer.github.com/v3/markdown/#render-an-arbitrary-markdown-document
curl https://api.github.com/markdown/raw -X "POST" -H "Content-Type: text/plain" -d "Hello world github/linguist#1 **cool**, and #1!"

DOC: https://docs.gitlab.com/ee/api/markdown.html#render-an-arbitrary-markdown-document
curl --header Content-Type:application/json --data '{"text":"Hello world! :tada:", "gfm":true, "project":"group_example/project_example"}' https://gitlab.example.com/api/v4/markdown
*/

// ConvertMD2html Convert Markdown to HTML
// styles from https://github.com/facelessuser/MarkdownPreview
//
// "image_path": "https://github.githubassets.com/images/icons/emoji/unicode/",
// "non_standard_image_path": "https://github.githubassets.com/images/icons/emoji/"
func ConvertMD2html() *gcli.Command {
	var mh = md2html{}

	c := &gcli.Command{
		Name:    "md:html",
		UseFor:  "convert one or multi markdown file to html",
		Aliases: []string{"md2html", "md:html"},
		// Config:  nil,
		// Examples: "",
		Func: mh.Handle,
	}

	c.BoolOpt(&mh.toc, "toc", "", false,
		"Generate a table of contents (implies --latex=false)")
	flag.BoolVar(&mh.tocOnly, "toconly", false,
		"Generate a table of contents only (implies -toc)")
	c.BoolOpt(&mh.page, "page", "", false,
		"Generate a standalone HTML page (implies --latex=false)")
	c.BoolOpt(&mh.latex, "latex", "", false,
		"Generate LaTeX output instead of HTML")

	c.BoolOpt(&mh.smartyPants, "smartypants", "", true,
		"Apply smartypants-style substitutions")
	c.BoolOpt(&mh.latexDashes, "latexdashes", "", true,
		"Use LaTeX-style dash rules for smartypants")
	c.BoolOpt(&mh.fractions, "fractions", "", true,
		"Use improved fraction rules for smartypants")
	c.BoolOpt(&mh.htmlSimple, "html-simple", "", true,
		"Sets HTML output to a simple, just bare minimum HTML tags and attributes")

	c.StrOpt(&mh.css, "css", "", "",
		"Link to a CSS stylesheet (implies --page)")
	c.StrOpt(&mh.output, "output", "", "",
		"the rendered content output, default output STDOUT")
	c.StrOpt(&mh.driver, "driver", "", "bf",
		"set the markdown renderer driver.\nallow:\n bf - blackfriday,\n gm - gomarkdown")

	c.AddArg("files", "the listed files will be render to html", false, true)

	// save
	mh.cmd = c
	return c
}

func (mh md2html) Handle(c *gcli.Command, args []string) (err error) {
	// enforce implied options
	if mh.css != "" {
		mh.page = true
	}
	if mh.page {
		mh.latex = false
	}
	if mh.toc {
		mh.latex = false
	}

	color.Info.Println("Work Dir:", c.WorkDir())
	color.Info.Println("Use Driver:", mh.driverName())

	mdString := `
# title

## h2

hello

### h3
`

	if mh.driver == driverBF {
		err = mh.blackFriday([]byte(mdString), args)
	} else {
		err = mh.goMarkdown([]byte(mdString), args)
	}

	// color.Success.Println("Complete")
	return
}

func (mh md2html) driverName() string {
	if name, ok := drivers[mh.driver]; ok {
		return name
	}

	return drivers[driverBF]
}

func (mh md2html) blackFriday(input []byte, args []string) (err error) {
	// set up options
	// extensions := 0
	// extensions |= blackfriday.EXTENSION_NO_INTRA_EMPHASIS
	// extensions |= blackfriday.EXTENSION_TABLES
	// extensions |= blackfriday.EXTENSION_FENCED_CODE
	// extensions |= blackfriday.EXTENSION_AUTOLINK
	// extensions |= blackfriday.EXTENSION_STRIKETHROUGH
	// extensions |= blackfriday.EXTENSION_SPACE_HEADERS

	// if mh.latex {
	// render the data into LaTeX
	// renderer = blackfriday.La(0)
	// } else {
	// render the data into HTML
	htmlFlags := blackfriday.HTMLFlagsNone
	// if xhtml {
	// 	htmlFlags |= blackfriday.HTML_USE_XHTML
	// }
	if mh.smartyPants {
		htmlFlags |= blackfriday.Smartypants
	}
	if mh.fractions {
		htmlFlags |= blackfriday.SmartypantsFractions
	}
	if mh.latexDashes {
		htmlFlags |= blackfriday.SmartypantsLatexDashes
	}

	title := ""
	if mh.page {
		htmlFlags |= blackfriday.CompletePage
		title = getTitle(input)
	}
	// if mh.tocOnly {
	// 	htmlFlags |= blackfriday.HTML_OMIT_CONTENTS
	// }
	if mh.toc {
		htmlFlags |= blackfriday.TOC
	}

	r := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
		Flags: blackfriday.CommonHTMLFlags,
		Title: title,
	})
	optList := []blackfriday.Option{
		blackfriday.WithRenderer(r),
		blackfriday.WithExtensions(blackfriday.CommonExtensions),
	}
	parser := blackfriday.New(optList...)
	ast := parser.Parse(input)

	var buf bytes.Buffer
	r.RenderHeader(&buf, ast)
	ast.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		return r.RenderNode(&buf, node, entering)
	})
	r.RenderFooter(&buf, ast)

	return mh.outToWriter(buf.Bytes())
}

func (mh md2html) goMarkdown(input []byte, args []string) (err error) {
	// set up options
	var extensions = gmparser.NoIntraEmphasis |
		gmparser.Tables |
		gmparser.FencedCode |
		gmparser.Autolink |
		gmparser.Strikethrough |
		gmparser.SpaceHeadings

	var renderer markdown.Renderer
	if mh.latex {
		// render the data into LaTeX
		// renderer = markdown.LatexRenderer(0)
		color.Comment.Println("unsupported")
		return
	} else {
		// render the data into HTML
		var htmlFlags html.Flags
		// if xhtml {
		// 	htmlFlags |= html.UseXHTML
		// }
		if mh.smartyPants {
			htmlFlags |= html.Smartypants
		}
		if mh.fractions {
			htmlFlags |= html.SmartypantsFractions
		}
		if mh.latexDashes {
			htmlFlags |= html.SmartypantsLatexDashes
		}

		title := ""
		if mh.page {
			htmlFlags |= html.CompletePage
			title = getTitle(input)
		}
		if mh.toc {
			htmlFlags |= html.TOC
		}

		params := html.RendererOptions{
			Flags: htmlFlags,
			Title: title,
			CSS:   mh.css,
		}
		renderer = html.NewRenderer(params)
	}

	// parse and render
	psr := gmparser.NewWithExtensions(extensions)

	htmlBts := markdown.ToHTML(input, psr, renderer)

	return mh.outToWriter(htmlBts)
}

func (mh md2html) outToWriter(htmlText []byte) (err error) {
	// output the result
	var out *os.File
	if mh.output == "" {
		color.Info.Println("OUTPUT:")
		out = os.Stdout
	} else {
		if out, err = os.Create(mh.output); err != nil {
			return mh.cmd.Errorf("Error creating %s: %v", mh.output, err)
		}
		defer out.Close()
	}

	if _, err = out.Write(htmlText); err != nil {
		err = mh.cmd.Errorf("Error writing output: %s", err.Error())
	}
	return
}

// try to guess the title from the input buffer
// just check if it starts with an <h1> element and use that
func getTitle(input []byte) string {
	i := 0

	// skip blank lines
	for i < len(input) && (input[i] == '\n' || input[i] == '\r') {
		i++
	}
	if i >= len(input) {
		return defaultTitle
	}
	if input[i] == '\r' && i+1 < len(input) && input[i+1] == '\n' {
		i++
	}

	// find the first line
	start := i
	for i < len(input) && input[i] != '\n' && input[i] != '\r' {
		i++
	}
	line1 := input[start:i]
	if input[i] == '\r' && i+1 < len(input) && input[i+1] == '\n' {
		i++
	}
	i++

	// check for a prefix header
	if len(line1) >= 3 && line1[0] == '#' && (line1[1] == ' ' || line1[1] == '\t') {
		return strings.TrimSpace(string(line1[2:]))
	}

	// check for an underlined header
	if i >= len(input) || input[i] != '=' {
		return defaultTitle
	}
	for i < len(input) && input[i] == '=' {
		i++
	}
	for i < len(input) && (input[i] == ' ' || input[i] == '\t') {
		i++
	}
	if i >= len(input) || (input[i] != '\n' && input[i] != '\r') {
		return defaultTitle
	}

	return strings.TrimSpace(string(line1))
}