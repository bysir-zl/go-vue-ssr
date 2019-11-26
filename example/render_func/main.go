package main

import (
	"github.com/bysir-zl/vue-ssr/genera"
	"net/http"
)

func main() {
	err := http.ListenAndServe(":10000", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// run pkg/vuessr/generator_test.go first
		html := genera.XComponent_helloworld(map[string]interface{}{
			"name":   "bysir",
			"sex":    "男",
			"age":    "18",
			"list":   []interface{}{"1", map[string]interface{}{"a": 1}},
			"isShow": "1",
		}, "")
		w.Write([]byte(html))

		return
	}))

	if err != nil {
		panic(err)
	}
}
