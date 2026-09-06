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

func TestSpanIALEntityBoundaries(t *testing.T) {
	engine := lute.New()
	engine.SetProtyleWYSIWYG(true)
	engine.SetKramdownIAL(true)
	engine.SetAutoSpace(false)
	for _, tc := range []struct {
		input, want string
	}{
		{`<span data-type="text">x</span>{: style="font-family: &quot;Segoe UI&quot;;"}tail &amp; end`, `font-family: "Segoe UI";`},
		{`<span data-type="text">x</span>{: style="font-family: &#34;Segoe UI&#34;;"}tail &amp; end`, `font-family: "Segoe UI";`},
		{`<span data-type="text" style="font-family: &quot;Segoe UI&quot;;">x</span>{: style="font-family: &quot;Segoe UI&quot;;"}tail &amp; end`, `font-family: "Segoe UI";`},
		{`<span data-type="text">x</span>{: title="A &amp; B &#123;C&#125;"}tail &amp; end`, "A & B {C}"},
	} {
		dom := engine.Md2BlockDOM(tc.input, false)
		if strings.Contains(dom, "{:") || !strings.Contains(dom, "tail &amp; end") {
			t.Fatalf("attributes or trailing entities were corrupted: %s", dom)
		}
		tree := parse.Parse("", []byte(tc.input), engine.ParseOptions)
		found := false
		ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
			if entering && n.Type == ast.NodeTextMark && (n.IALAttr("style") == tc.want || n.IALAttr("title") == tc.want) {
				found = true
			}
			return ast.WalkContinue
		})
		if !found {
			t.Fatalf("attribute value was not preserved: %s", dom)
		}
	}
	for _, input := range []string{
		`<span data-type="text">x</span>{: style="font-family: &quot;Segoe UI&quot;;"`,
		`<span data-type="text">x</span>{: style="font-family: &quot;Segoe UI&quot;;"` + "\n\n" + `tail}`,
		`<span data-type="text">x</span>{: title="&quot;a" **bold** }`,
	} {
		dom := engine.Md2BlockDOM(input, false)
		if !strings.Contains(dom, "{:") || strings.Contains(dom, `style="font-family:`) {
			t.Fatalf("incomplete or interrupted attributes were consumed: %s", dom)
		}
	}
}
