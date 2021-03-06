package version

// 当version改变，vue编译缓存就会失效。
const Version = "0.0.25"

// 0.0.9
// fix <!doctype html>

// 0.0.10
// fix unsafe string in attr

// 0.0.11
// use github.com/robertkrimen/otto to parse js code

// 0.0.12
// support watch file and recompile
// use the next package to watch file: github.com/radovskyb/watcher

// 0.0.13
// optimization code: scope

// 0.0.14
// support inject and provide

// 0.0.15
// fix empty slot

// 0.0.16
// fix panic when nil slot called

// 0.0.18
// 1. use strings.buffer to build string
// 2. you can custom Writer to receive result

// 0.0.19
// added *Options arg when call function

// 0.0.20
// reduce costs of NewRender()

// 0.0.21
// support v-html directive on <template>

// 0.0.22
// render voidElements

// 0.0.23
// ordered props

// 0.0.24
// support operate writer in directives

// 0.0.25
// exec directives on root tag of custom component
