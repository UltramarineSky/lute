package test

import (
	"strings"
	"testing"

	"github.com/88250/lute/html"
	"github.com/88250/lute/parse"
)

func TestTabsTaskRoundTrip(t *testing.T) {
	for _, marker := range []string{" ", "X", "x", "/", "-", "?", "!", "\"", "&", "<"} {
		l := tabsEngine()
		dom := l.Md2BlockDOM("::: tabs\n@tab **Task**\nBody\n:::\n", false)
		dom = strings.Replace(dom, `class="tab-item"`, `class="tab-item" tabs-task="`+html.EscapeAttrVal(marker)+`"`, 1)
		for round := 0; round < 3; round++ {
			tree := l.BlockDOM2Tree(dom)
			_, items := tabNodes(tree.Root)
			if len(items) != 1 || items[0].IALAttr("tabs-task") != marker {
				t.Fatalf("marker %q round %d lost in AST: %s", marker, round, dom)
			}
			exported := l.BlockDOM2HTML(dom)
			_, htmlItems := tabNodes(l.BlockDOM2Tree(l.HTML2BlockDOM(exported)).Root)
			if len(htmlItems) != 1 || htmlItems[0].IALAttr("tabs-task") != marker {
				t.Fatalf("marker %q lost in HTML: %s", marker, exported)
			}
			md := l.BlockDOM2StdMd(dom)
			_, items = tabNodes(parse.Parse("", []byte(md), l.ParseOptions).Root)
			if len(items) != 1 || items[0].IALAttr("tabs-task") != marker {
				t.Fatalf("marker %q lost in Markdown: %s", marker, md)
			}
			dom = l.Md2BlockDOM(md, false)
		}
	}
}
