// Lute - 一款结构化的 Markdown 引擎，支持 Go 和 JavaScript
// Copyright (c) 2019-present, b3log.org
//
// Lute is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//         http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package test

import (
	"strings"
	"testing"

	"github.com/88250/lute"
	"github.com/88250/lute/ast"
	"github.com/88250/lute/parse"
)

func TestHTMLTextStyle(t *testing.T) {
	engine := lute.New()
	engine.SetProtyleWYSIWYG(true)
	engine.SetHTMLTag2TextMark(true)
	engine.SetKramdownIAL(true)
	for _, tc := range []struct {
		name, input string
		expected    []string
	}{
		{"feishu", `<div data-lark-html-role="root"><div>222<span style="color:rgb(216,57,49);background-color:rgb(255,165,61)">333<s>33</s><strong>3</strong></span><strong>33444</strong><u>4444</u><code>231</code></div></div>`, []string{`color: rgb(216,57,49);background-color: rgb(255,165,61);`, `data-type="s"`, `data-type="strong"`, `data-type="u"`, `data-type="code"`}},
		{"family", `<p><span style="font-family: 'Noto Sans', serif; font-size:18pt;color:#123456">A<strong>B</strong><a href="https://example.com" title="link">C</a><code>D</code></span></p>`, []string{`font-family:`, `font-size: 24.000000px;`, `data-href="https://example.com"`, `data-type="code"`}},
		{"small", `<p style="font-size:24px">A<small>B<small>C</small><small style="font-size:18px">D</small></small>E</p>`, []string{`font-size: 20.000000px;`, `font-size: 16.666667px;`, `font-size: 18.000000px;`}},
		{"escaping", `<p><span style="color:red">*&lt;img src=x&gt;&amp;</span></p>`, []string{`*&lt;img src=x&gt;&amp;`}},
		{"link", `<p><a href="https://example.com" title="link"><span style="color:red">C</span></a></p>`, []string{`data-href="https://example.com"`, `data-title="link"`, `color: red;`}},
		{"table", `<table><tr><td><small style="color:red">A<strong>B</strong></small></td><td>C</td></tr></table>`, []string{`data-type="NodeTable"`, `font-size: 0.833333em;`, `color: red;`}},
		{"font", `<p><font color="#123456" face="Arial" size="4">A<strong>B</strong></font></p>`, []string{`font-family: Arial;`, `font-size: 1.200000rem;`, `data-type="strong"`}},
		{"styled code", `<p><code style="color:red;font-size:12px">A</code></p>`, []string{`data-type="code"`, `color: red;`, `font-size: 12.000000px;`}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			output := engine.HTML2BlockDOM(tc.input)
			if strings.Contains(output, `data-href=""`) {
				t.Errorf("empty link in %s", output)
			}
			for _, expected := range tc.expected {
				if !strings.Contains(output, expected) {
					t.Errorf("missing %q in %s", expected, output)
				}
			}
		})
	}
}

func TestHTMLTextStyleRoundTrip(t *testing.T) {
	engine := lute.New()
	engine.SetProtyleWYSIWYG(true)
	engine.SetHTMLTag2TextMark(true)
	engine.SetKramdownIAL(true)
	input := `<p><span style="color:red;font-family:Arial;font-size:24px">A<strong>B</strong><a href="https://example.com">C</a><small>D</small></span></p>`
	blockDOM := engine.HTML2BlockDOM(input)
	for i := 0; i < 3; i++ {
		blockDOM = engine.SpinBlockDOM(blockDOM)
		for _, expected := range []string{`color: red;`, `font-family: Arial;`, `font-size: 24.000000px;`, `font-size: 20.000000px;`, `data-href="https://example.com"`} {
			if !strings.Contains(blockDOM, expected) {
				t.Fatalf("round %d lost %q: %s", i, expected, blockDOM)
			}
		}
	}
}

func TestHTMLQuotedFontFamilyRoundTrip(t *testing.T) {
	engine := lute.New()
	engine.SetHeadingID(false)
	engine.SetProtyleWYSIWYG(true)
	engine.SetHTMLTag2TextMark(true)
	engine.SetKramdownIAL(true)
	engine.SetAutoSpace(false)
	for _, input := range []string{
		`<p style='color:red;font-family:"Mona Sans VF", "Segoe UI", Arial;font-size:14px'>before <strong>bold</strong> after</p>`,
		`<h2 style="color:red;font-family:&quot;Mona Sans VF&quot;, &quot;Segoe UI&quot;, Arial;font-size:14px">before <strong>bold</strong> after</h2>`,
		`<table><tr><td style='color:red;font-family:"Mona Sans VF", "Segoe UI", Arial;font-size:14px'>before <strong>bold</strong> after</td></tr></table>`,
	} {
		dom := engine.HTML2BlockDOM(input)
		for i := 0; i < 3; i++ {
			if strings.Contains(dom, "{: style=") {
				t.Fatalf("round %d leaked attributes: %s", i, dom)
			}
			tree := engine.BlockDOM2Tree(dom)
			styled := 0
			ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
				if entering && n.Type == ast.NodeTextMark {
					styled++
					style := n.IALAttr("style")
					for _, want := range []string{`font-family: "Mona Sans VF", "Segoe UI", Arial;`, "color: red;", "font-size: 14.000000px;"} {
						if !strings.Contains(style, want) {
							t.Errorf("round %d lost %q: %s", i, want, dom)
						}
					}
					if n.TextMarkTextContent == "bold" && !n.ContainTextMarkTypes("strong") {
						t.Errorf("bold was lost: %s", dom)
					}
				}
				return ast.WalkContinue
			})
			if styled < 3 {
				t.Fatalf("text fragments were lost: %s", dom)
			}
			dom = engine.SpinBlockDOM(dom)
		}
	}
}

func TestMarkdownHTMLStyleBoundaries(t *testing.T) {
	engine := lute.New()
	for _, input := range []string{"`<small>x</small>`", "```html\n<small>x</small>\n```", `before <small>unclosed`} {
		tree := parse.Parse("", []byte(input), engine.ParseOptions)
		parse.NormalizeInlineHTMLTextStyles(tree)
		ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
			if entering && n.Type == ast.NodeTextMark && n.IALAttr("style") != "" {
				t.Fatalf("unexpected style in %s", input)
			}
			return ast.WalkContinue
		})
	}
	input := `<small>A &amp; <a href="java&#x09;script:alert(1)">B</a></small>`
	tree := parse.Parse("", []byte(input), engine.ParseOptions)
	parse.NormalizeInlineHTMLTextStyles(tree)
	ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
		if entering && n.Type == ast.NodeTextMark && n.TextMarkAHref != "" {
			t.Fatal("unsafe link retained")
		}
		return ast.WalkContinue
	})
	input = `<span data-type="block-ref" data-id="20260101000000-abcdefg">reference</span> <small>small</small>`
	tree = parse.Parse("", []byte(input), engine.ParseOptions)
	parse.NormalizeInlineHTMLTextStyles(tree)
	protected := false
	ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
		if entering && n.Type == ast.NodeInlineHTML && strings.Contains(n.TokensStr(), `data-type="block-ref"`) {
			protected = true
		}
		return ast.WalkContinue
	})
	if !protected {
		t.Fatal("native text mark was consumed by external style conversion")
	}
}

func TestMarkdownHTMLTextStyle(t *testing.T) {
	engine := lute.New()
	for _, input := range []string{`A<small>B **C** <small>D</small> [E](https://example.com)</small>F`, `A<span style="color:red;font-family:Arial">B<strong>C</strong></span>F`} {
		tree := parse.Parse("", []byte(input), engine.ParseOptions)
		parse.NormalizeInlineHTMLTextStyles(tree)
		styled := 0
		ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
			if entering && n.Type == ast.NodeTextMark && n.IALAttr("style") != "" {
				styled++
			}
			if entering && n.Type == ast.NodeInlineHTML {
				t.Errorf("unconverted HTML: %s", n.Tokens)
			}
			return ast.WalkContinue
		})
		if styled == 0 {
			t.Fatalf("no styled text: %s", input)
		}
	}
}
