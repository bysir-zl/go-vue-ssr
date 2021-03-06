// Code generated by go-vue-ssr: https://github.com/zbysir/go-vue-ssr
// src_hash:af2dfadd1b22cde6a44c3361cc04e301

package tplgo

import (
	"strings"
)

type _ strings.Builder

func (r *Render) Component_xattr(options *Options) string {
	scope := extendScope(r.Global.Scope, options.Props)
	options.Directives.Exec(r, options)
	_ = scope
	return r.tag("div", true, &Options{
		PropsClass: scope.Get("customClass"),
		Props:      map[string]interface{}{"id": "id"},
		Attrs:      map[string]string{"data-src": "//baidu.jpg"},
		Class:      []string{"b"},
		Slot: map[string]NamedSlotFunc{"default": func(props map[string]interface{}) string {
			return "\n        test attr\n        <img" + mixinAttr(nil, map[string]string{"alt": "标题"}, map[string]interface{}{"src": interfaceToFunc(scope.Get("img"))(scope.Get("imgUrl"))}) + "></img>"
		}},
		P:     options,
		Scope: scope,
	})
}
