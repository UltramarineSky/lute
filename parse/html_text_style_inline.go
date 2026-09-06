// Lute - 一款结构化的 Markdown 引擎，支持 Go 和 JavaScript
// Copyright (c) 2019-present, b3log.org
//
// Lute is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//         http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package parse

import (
	"strings"
	"unicode"

	"github.com/88250/lute/ast"
	"github.com/88250/lute/html"
	"github.com/88250/lute/html/atom"
	"github.com/88250/lute/util"
)

type htmlStyleFrame struct {
	node      *ast.Node
	dom       *html.Node
	style     HTMLTextStyle
	protected bool
}

func inlineStyleTag(n *ast.Node) (dom *html.Node, closing bool) {
	if n.Type != ast.NodeInlineHTML {
		return
	}
	z := html.NewTokenizer(strings.NewReader(n.TokensStr()))
	typ := z.Next()
	if typ != html.StartTagToken && typ != html.EndTagToken {
		return
	}
	tag := z.Token()
	switch tag.DataAtom {
	case atom.Span, atom.Small, atom.Font, atom.Strong, atom.B, atom.Em, atom.I, atom.S, atom.Del, atom.Strike, atom.U, atom.Code, atom.A, atom.Sub, atom.Sup, atom.Mark, atom.Kbd:
		return &html.Node{Type: html.ElementNode, Data: tag.Data, DataAtom: tag.DataAtom, Attr: tag.Attr}, typ == html.EndTagToken
	}
	return
}

// NormalizeInlineHTMLTextStyles 只转换配对的行内标签，代码块和未闭合标签保持原样。
func NormalizeInlineHTMLTextStyles(tree *Tree) {
	var blocks []*ast.Node
	ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
		if entering && (n.Type == ast.NodeParagraph || n.Type == ast.NodeHeading || n.Type == ast.NodeTableCell) {
			blocks = append(blocks, n)
			return ast.WalkSkipChildren
		}
		return ast.WalkContinue
	})
	for _, block := range blocks {
		normalizeInlineHTMLTextStyleBlock(tree, block)
	}
}

func normalizeInlineHTMLTextStyleBlock(tree *Tree, block *ast.Node) {
	pairs := map[*ast.Node]*ast.Node{}
	var stack []htmlStyleFrame
	hasStyle := false
	ast.Walk(block, func(n *ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.WalkContinue
		}
		dom, closing := inlineStyleTag(n)
		if dom == nil {
			return ast.WalkContinue
		}
		if !closing {
			stack = append(stack, htmlStyleFrame{node: n, dom: dom})
		} else if len(stack) > 0 && stack[len(stack)-1].dom.DataAtom == dom.DataAtom {
			open := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			pairs[open.node], pairs[n] = n, open.node
			if ResolveHTMLTextStyle(open.dom, HTMLTextStyle{}).CSS() != "" {
				hasStyle = true
			}
		}
		return ast.WalkContinue
	})
	if !hasStyle {
		return
	}
	// 先展开 Markdown 组合格式，使 HTML 样式能够覆盖每一个独立的文本片段。
	NestedInlines2FlattedSpansHybrid(&Tree{Root: block, Context: tree.Context}, false)
	stack = nil
	var unlinks []*ast.Node
	ast.Walk(block, func(n *ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.WalkContinue
		}
		if pairs[n] != nil {
			dom, closing := inlineStyleTag(n)
			protected := false
			if len(stack) > 0 {
				protected = stack[len(stack)-1].protected
			}
			if closing {
				if len(stack) > 0 && stack[len(stack)-1].node == pairs[n] {
					stack = stack[:len(stack)-1]
				}
			} else {
				// 原生文本标记和行级备注继续交由其专用解析器处理。
				protected = protected || util.DomAttrValue(dom, "data-type") != "" || (dom.DataAtom == atom.Span && util.DomAttrValue(dom, "title") != "")
				parent := HTMLTextStyle{}
				if len(stack) > 0 {
					parent = stack[len(stack)-1].style
					dom.Parent = stack[len(stack)-1].dom
				}
				stack = append(stack, htmlStyleFrame{node: n, dom: dom, style: ResolveHTMLTextStyle(dom, parent), protected: protected})
			}
			if !protected {
				unlinks = append(unlinks, n)
			}
			return ast.WalkContinue
		}
		if len(stack) == 0 || stack[len(stack)-1].protected {
			return ast.WalkContinue
		}
		style := stack[len(stack)-1].style
		if n.Type != ast.NodeText && n.Type != ast.NodeLinkText && n.Type != ast.NodeTextMark && n.Type != ast.NodeHTMLEntity {
			return ast.WalkContinue
		}
		if ownStyle := n.IALAttr("style"); ownStyle != "" {
			style = ResolveHTMLTextStyle(&html.Node{Attr: []*html.Attribute{{Key: "style", Val: ownStyle}}}, style)
		}
		ApplyHTMLTextStyle(n, style.CSS())
		if n.Type == ast.NodeTextMark {
			ApplyHTMLTextSemantics(n, stack[len(stack)-1].dom)
		}
		for _, frame := range stack {
			typ := map[atom.Atom]string{atom.Strong: "strong", atom.B: "strong", atom.Em: "em", atom.I: "em", atom.S: "s", atom.Del: "s", atom.Strike: "s", atom.U: "u", atom.Code: "code", atom.A: "a", atom.Sub: "sub", atom.Sup: "sup", atom.Mark: "mark", atom.Kbd: "kbd"}[frame.dom.DataAtom]
			if typ == "" {
				continue
			}
			if typ == "a" {
				unsafe := false
				for _, attr := range frame.dom.Attr {
					if attr.Key != "href" {
						continue
					}
					href := strings.ToLower(strings.Map(func(r rune) rune {
						if unicode.IsSpace(r) || unicode.IsControl(r) {
							return -1
						}
						return r
					}, attr.Val))
					unsafe = strings.HasPrefix(href, "javascript:") || strings.HasPrefix(href, "vbscript:") || strings.HasPrefix(href, "data:")
				}
				if unsafe {
					continue
				}
			}
			if n.Type != ast.NodeTextMark {
				ensureHTMLTextMark(n)
			}
			if !n.ContainTextMarkTypes(typ) {
				n.TextMarkType = strings.TrimSpace(strings.TrimSuffix(n.TextMarkType, "text") + " " + typ)
			}
			if typ == "a" {
				for _, attr := range frame.dom.Attr {
					if attr.Key == "href" {
						n.TextMarkAHref = attr.Val
					}
					if attr.Key == "title" {
						n.TextMarkATitle = attr.Val
					}
				}
			}
		}
		return ast.WalkContinue
	})
	for _, n := range unlinks {
		n.Unlink()
	}
}
