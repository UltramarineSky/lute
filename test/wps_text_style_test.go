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
)

// WPS 使用 background 简写，并在下划线标签内部重复声明文本修饰。
const wpsTextStyleHTML = `<p class=MsoNormal style="mso-pagination:widow-orphan;text-align:left;"><span style="mso-spacerun:'yes';font-family:宋体;font-size:12.0000pt;">222</span><span style="font-family:宋体;color:rgb(216,57,49);font-size:12.0000pt;background:rgb(255,165,61);mso-shading:rgb(255,165,61);">333</span><s><span style="font-family:宋体;color:rgb(216,57,49);text-decoration:line-through;font-size:12.0000pt;background:rgb(255,165,61);">33</span></s><span style="font-family:宋体;color:rgb(216,57,49);font-size:12.0000pt;background:rgb(255,165,61);">3</span><b style="mso-bidi-font-weight:normal"><span style="font-family:宋体;color:rgb(216,57,49);mso-ansi-font-weight:bold;font-size:12.0000pt;background:rgb(255,165,61);">3</span></b><b><span style="font-family:宋体;font-size:12.0000pt;">33444</span></b><span style="font-family:宋体;font-size:12.0000pt;">444</span><u><span style="font-family:宋体;text-decoration:underline;text-underline:single;font-size:12.0000pt;">4444</span></u><span style="font-family:宋体;font-size:12.0000pt;">4 4123123123</span></p>`

func TestWPSTextStyle(t *testing.T) {
	engine := lute.New()
	engine.SetProtyleWYSIWYG(true)
	engine.SetHTMLTag2TextMark(true)
	engine.SetKramdownIAL(true)
	dom := engine.HTML2BlockDOM(wpsTextStyleHTML)
	for i := 0; i < 3; i++ {
		for _, expected := range []string{`background-color: rgb(255,165,61);`, `color: rgb(216,57,49);`, `font-family: 宋体;`, `font-size: 16.000000px;`, `data-type="u" style="font-family: 宋体;font-size: 16.000000px;">4444</span>`} {
			if !strings.Contains(dom, expected) {
				t.Errorf("missing %q in %s", expected, dom)
			}
		}
		if strings.Contains(dom, "{: style=") {
			t.Errorf("style syntax leaked into content: %s", dom)
		}
		dom = engine.SpinBlockDOM(dom)
	}
}
