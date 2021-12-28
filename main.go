package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

func main() {
	parse("docs.html")

}

func parse(src string) {

	f, err := os.Open(src)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	n, err := html.Parse(f)
	if err != nil {
		panic(err)
	}
	body := n.LastChild.LastChild
	body = findChildNodeWithClass(body, "main")
	docNodes := filter(getNodeChildren(body), func(n *html.Node) bool {
		return nodeHasClass(n, "doc-content")
	})

	docs := Docs{
		main:    body,
		Content: make([]Content, len(docNodes)-1),
	}

	for i, v := range docNodes[1:] {
		docs.Content[i] = makeContent(v)
	}

	if false {

		res, err := json.Marshal(docs)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s\n", res)
	}

}

func makeContent(v *html.Node) Content {

	c := Content{
		mainNode:     v,
		LeftContent:  makeLeft(v),
		RightContent: makeRight(v),
	}

	if c.LeftContent.leftNode == nil || c.RightContent.rightNode == nil {
		panic("Node failed to create")
	}

	return c
}

func makeRight(n *html.Node) RightContent {
	rn := findChildNodeWithClass(n, "right-code")
	rc := RightContent{
		rightNode: rn,
		contents:  filter(getNodeChildren(rn), notTextFilter),
	}
	return rc
}

type Docs struct {
	main    *html.Node
	Table   TableContent
	Content []Content
}

type TableContent struct {
	table *html.Node
}

type Content struct {
	mainNode     *html.Node
	LeftContent  LeftContent
	RightContent RightContent
}

type RightContent struct {
	rightNode *html.Node
	contents  []*html.Node
}
type LeftContent struct {
	leftNode *html.Node
	contents []*html.Node

	Title string

	//All P tags before the next header tag/may be description.
	Definition string

	//could have both authentication and authorization.
	Authentication string

	Authorization string

	//URL is the method AND the url as one string.

	URL    string
	Method string

	//the text following the Pagination title.

	PaginationSupport string

	//Table of params bay just be QueryParameters with Required field.

	RequiredQueryParams []map[string]string

	OptionalQueryParams []map[string]string

	RequiredBodyParams []map[string]string //May also be Values instead of parameters, may just be "Body Parameters"

	OptionalBodyParams []map[string]string //May also be Values instead of parameters

	ReturnValues []map[string]string // Response Fields is also possible.

	ResponseCodes []map[string]string
}

func notTextFilter(t *html.Node) (keep bool) {
	return t.Type != html.TextNode
}

func makeLeft(n *html.Node) LeftContent {
	ln := findChildNodeWithClass(n, "left-docs")
	nc := getNodeChildren(ln)
	f := filter(nc, notTextFilter)
	lc := LeftContent{
		leftNode: ln,
		contents: f,
	}

	lc.makeTitle()
	lc.makeDefinition()
	lc.makeAuthentication()
	lc.makeAuthorization()
	lc.makeURL()
	lc.makePagination()
	lc.makeReqQuery()
	lc.makeOptQuery()
	lc.makeReqBody()
	lc.makeOptBody()
	lc.makeReturnValues()
	lc.makeResponseCodes()

	return lc
}

func (lc *LeftContent) makeTitle() {
	lc.Title = getString(lc.contents[0])
	// fmt.Println(lc.Title)
}
func (lc *LeftContent) makeDefinition() {}
func (lc *LeftContent) makeAuthentication() {
	idx := 0
	for i, v := range lc.contents {
		match, err := regexp.MatchString("Authen", getString(v))
		if err != nil {
			panic(err)
		}
		if match {
			idx = i
			break
		}
	}
	if idx == 0 {
		return
	}
	if idx >= len(lc.contents)-1 {
		return
	}
	r, err := regexp.Compile("\\s+")
	if err != nil {
		panic(err)
	}

	lc.Authentication = r.ReplaceAllString(strings.TrimSpace(getString(lc.contents[idx+1])), " ")
	fmt.Println(lc.Authentication)
}
func (lc *LeftContent) makeAuthorization() {
	idx := 0
	for i, v := range lc.contents {
		match, err := regexp.MatchString("Author", getString(v))
		if err != nil {
			panic(err)
		}
		if match {
			idx = i
			break
		}
	}
	if idx == 0 {
		return
	}
	if idx >= len(lc.contents)-1 {
		return
	}
	r, err := regexp.Compile("\\s+")
	if err != nil {
		panic(err)
	}

	lc.Authorization = r.ReplaceAllString(strings.TrimSpace(getString(lc.contents[idx+1])), " ")
	fmt.Println(lc.Authorization)
}
func (lc *LeftContent) makeURL() {
	idx := 0
	for i, v := range lc.contents {
		if strings.EqualFold(getString(v), "url") {
			idx = i
			break
		}
	}

	lcTry := getString(lc.contents[idx+1])
	parts := strings.Split(lcTry, " ")
	u := ""
	switch len(parts) {
	case 0:
		panic(fmt.Sprintf("unknown parts count %+v", parts))
	case 1:
		lc.Method = "GET"
		u = parts[0]
	case 2:
		lc.Method = parts[0]
		u = parts[1]
	default:
		panic(fmt.Sprintf("unknown parts count %+v", parts))
	}

	rl, err := url.Parse(u)
	if err != nil {
		panic(err.Error())
	}
	lc.URL = rl.Path

	// fmt.Println(lcTry)
}
func (lc *LeftContent) makePagination()    {}
func (lc *LeftContent) makeReqQuery()      {}
func (lc *LeftContent) makeOptQuery()      {}
func (lc *LeftContent) makeReqBody()       {}
func (lc *LeftContent) makeOptBody()       {}
func (lc *LeftContent) makeReturnValues()  {}
func (lc *LeftContent) makeResponseCodes() {}

var unk map[string]int = make(map[string]int)

func getString(n *html.Node) string {
	children := getNodeChildren(n)
	str := strings.Builder{}
	for _, v := range children {

		switch v.Type {
		case html.TextNode:
			str.WriteString(v.Data)
			continue
		case html.ElementNode:
			switch strings.ToLower(v.Data) {
			case "p":
				//its a paragraph. Do what I can to parse everything in it.
				str.WriteString(getString(v))
			case "div":
				str.WriteString(getString(v))
			case "ul":
				str.WriteString(" ")
				str.WriteString(getString(v))
			case "li":
				str.WriteString(" ")
				str.WriteString(getString(v))
			case "table":
				str.WriteString(" ")
				str.WriteString(getString(v))
			case "thead":
				str.WriteString(" ")
				str.WriteString(getString(v))
			case "tr":
				str.WriteString(" ")
				str.WriteString(getString(v))
			case "td":
				str.WriteString(" ")
				str.WriteString(getString(v))
			case "th":
				str.WriteString(" ")
				str.WriteString(getString(v))
			case "tbody":
				str.WriteString(" ")
				str.WriteString(getString(v))
			case "code":
				str.WriteString(getString(v))
			case "strong":
				str.WriteString(getString(v))
			case "a":
				str.WriteString(getString(v))
			case "em":
				str.WriteString(getString(v))
			case "br":
				str.WriteString("\n")
			case "pre":
				str.WriteString(getString(v))
			case "span":
				str.WriteString(getString(v))

			default:
				count, ok := unk[v.Data]
				if !ok {
					fmt.Printf("unknown data (%s)\n", v.Data)
					unk[v.Data] = 1
				} else {
					unk[v.Data] = count + 1
				}
			}
		}

	}

	return str.String()
}
