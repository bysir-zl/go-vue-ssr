package vuessr

import (
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// 用来生成模板字符串代码
// 目的是为了解决递归渲染节点造成的性能问题

func genComponentRenderFunc(app *Compiler, pkgName, name string, file string) string {
	ve, err := ParseVue(file)
	if err != nil {
		panic(err)
	}
	code, _ := ve.GenCode(app)

	// 处理多余的纯字符串拼接: "a"+"b" => "ab"
	code = strings.Replace(code, `"+"`, "", -1)

	return fmt.Sprintf("package %s\n\n"+
		"func (r *Render)Component_%s(options *Options)string{\n"+
		"%s:= extendMap(r.Prototype, options.Props)\n_ = %s\n"+ // 声明this
		"return %s"+
		"}", pkgName, name, DataKey, DataKey, code)
}

func tuoFeng2SheXing(src string) (outStr string) {
	l := len(src)
	var out []byte
	for i := 0; i < l; i = i + 1 {
		// 大写变小写
		if 97-32 <= src[i] && src[i] <= 122-32 {
			if i != 0 {
				out = append(out, '-')
			}
			out = append(out, src[i]+32)
		} else {
			out = append(out, src[i])
		}
	}

	return string(out)
}

func sheXing2TuoFeng(src string) (outStr string) {
	l := len(src)
	out := make([]byte, l)

	// 首字母
	out[0] = src[0]

	del := 0
	for i := 1; i < l; i = i + 1 {
		// 是下划线
		if '-' == src[i] {
			// 下划线的下一个是小写字母
			if 97 <= src[i+1] && src[i+1] <= 122 {
				out[i-del] = src[i+1] - 32
			} else {
				out[i-del] = src[i+1]
			}
			del++
			i++
		} else {
			out[i-del] = src[i]
		}
	}
	out = out[0 : l-del]
	return string(out)
}

func genNew(app *Compiler, pkgName string) string {
	m := map[string]string{}
	for tagName, comName := range app.Components {
		m[tagName] = fmt.Sprintf(`r.Component_%s`, comName)
	}

	return fmt.Sprintf("package %s\n\n"+
		"func NewRender() *Render{"+
		"r:=&Render{}\n"+
		"r.components = %s\n"+
		"return r"+
		"}",
		pkgName, mapGoCodeToCode(m, "ComponentFunc"))
}

// 组件名字, 驼峰
func componentName(src string) string {
	return sheXing2TuoFeng(src)
}

type VueFile struct{
	ComponentName string // xText
	Path string
	Filename string // x-text.vue
}

// 生成并写入文件夹
func GenAllFile(src, desc string) (err error) {
	// 生成文件夹
	err = os.MkdirAll(desc, os.ModePerm)
	if err != nil {
		return
	}

	// 删除老的.vue.go文件
	oldComp, err := walkDir(desc, ".vue.go")
	if err != nil {
		return
	}
	oldCompMap := map[string]struct{}{}

	for _, v := range oldComp {
		_, fileName := filepath.Split(v)
		name := componentName(strings.Split(fileName, ".")[0])
		oldCompMap[name] = struct{}{}
	}

	// 生成新的
	vueFiles, err := walkDir(src, ".vue")
	if err != nil {
		return
	}

	c := NewCompiler()

	var vs []VueFile

	for _, v := range vueFiles {
		_, fileName := filepath.Split(v)
		name := componentName(strings.Split(fileName, ".")[0])

		vs = append(vs, VueFile{
			ComponentName: name,
			Path:          v,
			Filename:      fileName,
		})
	}

	_, pkgName := filepath.Split(desc)

	// 注册vue组件代码
	code := genNew(c, pkgName)
	err = ioutil.WriteFile(desc+string(os.PathSeparator)+"new.go", []byte(code), 0666)
	if err != nil {
		return
	}

	// 生成vue组件
	for _, v := range vueFiles {
		// text.vue
		_, fileName := filepath.Split(v)
		name := strings.Split(fileName, ".")[0]
		code := genComponentRenderFunc(c, pkgName, sheXing2TuoFeng(name), v)
		var codeBs []byte
		codeBs, err = format.Source([]byte(code))
		if err != nil {
			return
		}

		if oldCompMap[]

		ioutil.ReadFile()

		err = ioutil.WriteFile(desc+string(os.PathSeparator)+name+".vue.go", codeBs, 0666)
		if err != nil {
			return
		}
	}

	// buildin代码
	code = fmt.Sprintf("package %s\n", pkgName) + buildInCode
	err = ioutil.WriteFile(desc+string(os.PathSeparator)+"buildin.go", []byte(code), 0666)
	if err != nil {
		return
	}

	return
}

func walkDir(dirPth string, suffix string) (files []string, err error) {
	files = make([]string, 0, 30)

	err = filepath.Walk(dirPth, func(filename string, fi os.FileInfo, err error) error {
		//遍历目录
		if err != nil {
			return err
		}
		if fi.IsDir() {
			// 忽略目录
			return nil
		}

		if strings.HasSuffix(filename, suffix) {
			files = append(files, filename)
		}

		return nil
	})

	return
}

const buildInCode = `
import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/html"
	"sort"
	"strings"
)

type Render struct {
	// 模拟原型链, 每个组件中都可以直接读取到这个对象中的值. 如果和组件上层传递的props冲突, 则上层传递的props优先.
	// 其中可以写签名为function的方法, 可以供{{func(a)}}语法使用.
	Prototype map[string]interface{}
	// 注册的动态组件
	components map[string]ComponentFunc
	// 指令
	directives map[string]DirectivesFunc
}

// for {{func(a)}}
type Function func(args ...interface{}) interface{}

type DirectivesBinding struct {
	Value interface{}
	Arg   string
	Name  string
}

type DirectivesFunc func(b DirectivesBinding, options *Options)

func emptyFunc(args ...interface{}) interface{} {
	if len(args) != 0 {
		return args[0]
	}
	return nil
}

// 注册指令
func (r *Render) Directive(name string, f DirectivesFunc) {
	if r.directives == nil {
		r.directives = map[string]DirectivesFunc{}
	}

	r.directives[name] = f
}

// 内置组件
func (r *Render) Component_slot(options *Options) string {
	name := options.Attrs["name"]
	if name == "" {
		name = "default"
	}
	props := options.Props
	injectSlotFunc := options.P.Slot[name]

	// 如果没有传递slot 则使用默认的code
	if injectSlotFunc == nil {
		return options.Slot["default"](nil)
	}

	return injectSlotFunc(props)
}

func (r *Render) Component_component(options *Options) string {
	is, ok := options.Props["is"].(string)
	if !ok {
		return ""
	}
	if c, ok := r.components[is]; ok {
		return c(options)
	}

	return fmt.Sprintf("<p>not register com: %s</p>", is)
}

// 动态tag
// 何为动态tag:
// - 每个组件的root层tag(attr受到上层传递的props影响)
// - 有自己定义指令(自定义指令需要修改组件所有属性, 只能由动态tag实现)
func (r *Render) Tag(tagName string, isRoot bool, options *Options) string {
	// exec directive
	if len(options.Directives) != 0 {
		for _, d := range options.Directives {
			if f, ok := r.directives[d.Name]; ok {
				f(DirectivesBinding{
					Value: d.Value,
					Arg:   d.Arg,
					Name:  d.Name,
				}, options)
			}
		}
	}

	var p *Options
	if isRoot {
		p = options.P
	}

	// attr
	attr := mixinClass(p, options.Class, options.PropsClass) +
		mixinStyle(p, options.Style, options.PropsStyle) +
		mixinAttr(p, options.Attrs, options.Props.CanBeAttr())

	eleCode := fmt.Sprintf("<%s%s>%s</%s>", tagName, attr, options.Slot["default"](nil), tagName)
	return eleCode
}

// 渲染组件需要的结构
// tips: 此结构应该尽量的简单, 方便渲染才能性能更好.
type Options struct {
	Props      Props       // 本节点的数据(不包含class和style)
	PropsClass interface{} // :class
	PropsStyle interface{} // :style
	// PropsAttr  map[string]interface{}   // 可以被生成attr的Props, 由Props.CanBeAttr而来
	Attrs     map[string]string        // 本节点静态的attrs (除去class和style)
	Class     []string                 // 本节点静态class
	Style     map[string]string        // 本节点静态style
	StyleKeys []string                 // 样式的key, 用来保证顺序, 只会作用在root节点
	Slot      map[string]namedSlotFunc // 当前组件所有的插槽代码(v-slot指令和默认的子节点), 支持多个不同名字的插槽, 如果没有名字则是"default"
	// 父级options
	// - 在渲染插槽会用到. (根据name取到父级的slot)
	// - 读取上层传递的PropsClass, 作用在root tag
	P          *Options
	Directives []directive // 指令值
}

type directive struct {
	Name  string
	Value interface{}
	Arg   string
}

type Props map[string]interface{}

func (p Props) CanBeAttr() Props {
	html := map[string]struct{}{
		"id":  {},
		"src": {},
	}

	a := Props{}
	for k, v := range p {
		if _, ok := html[k]; ok {
			a[k] = v
			continue
		}

		if strings.HasPrefix(k, "data-") {
			a[k] = v
			continue
		}
	}
	return a
}

// 组件的render函数
type ComponentFunc func(options *Options) string

// 用来生成slot的方法
// 由于slot具有自己的作用域, 所以只能使用闭包实现(而不是字符串).
type namedSlotFunc func(props map[string]interface{}) string

// 混合动态和静态的标签, 主要是style/class需要混合
// todo) 如果style/class没有冲突, 则还可以优化
// tip: 纯静态的class应该在编译时期就生成字符串, 而不应调用这个
// classProps: 支持 obj, array, string
// options: 上层组件的options
func mixinClass(options *Options, staticClass []string, classProps interface{}) (str string) {
	var class []string
	// 静态
	for _, c := range staticClass {
		if c != "" {
			class = append(class, c)
		}
	}

	// 本身的props
	for _, c := range getClassFromProps(classProps) {
		class = append(class, c)
	}

	if options != nil {
		// 上层传递的props
		if options.Props != nil {
			if options.PropsClass != nil {
				for _, c := range getClassFromProps(options.PropsClass) {
					class = append(class, c)
				}
			}
		}

		// 上层传递的静态class
		if len(options.Class) != 0 {
			for _, c := range options.Class {
				if c != "" {
					class = append(class, c)
				}
			}
		}
	}

	if len(class) != 0 {
		str = fmt.Sprintf(" class=\"%s\"", strings.Join(class, " "))
	}

	return
}

// 构建style, 生成如style="color: red"的代码, 如果style代码为空 则只会返回空字符串
func mixinStyle(options *Options, staticStyle map[string]string, styleProps interface{}) (str string) {
	style := map[string]string{}

	// 静态
	for k, v := range staticStyle {
		style[k] = v
	}

	// 当前props
	ps := getStyleFromProps(styleProps)
	for k, v := range ps {
		style[k] = v
	}

	if options != nil {
		// 上层传递的props
		if options.Props != nil {
			if options.PropsStyle != nil {
				ps := getStyleFromProps(options.PropsStyle)
				for k, v := range ps {
					style[k] = v
				}
			}
		}

		// 上层传递的静态style
		for k, v := range options.Style {
			style[k] = v
		}
	}

	styleCode := genStyle(style)
	if styleCode != "" {
		str = fmt.Sprintf(" style=\"%s\"", styleCode)
	}

	return
}

// 生成除了style和class的attr
func mixinAttr(options *Options, staticAttr map[string]string, propsAttr map[string]interface{}) string {
	attrs := map[string]string{}

	// 静态
	for k, v := range staticAttr {
		attrs[k] = v
	}

	// 当前props
	ps := getStyleFromProps(propsAttr)
	for k, v := range ps {
		attrs[k] = v
	}

	if options != nil {
		// 上层传递的props
		if options.Props != nil {
			for k, v := range (Props(options.Props)).CanBeAttr() {
				attrs[k] = fmt.Sprintf("%v", v)
			}
		}

		// 上层传递的静态style
		for k, v := range options.Attrs {
			attrs[k] = v
		}
	}

	c := genAttr(attrs)
	if c == "" {
		return ""
	}

	return fmt.Sprintf(" %s", c)
}

func getSortedKey(m map[string]string) (keys []string) {
	keys = make([]string, len(m))
	index := 0
	for k := range m {
		keys[index] = k
		index++
	}
	if len(m) < 2 {
		return keys
	}

	sort.Strings(keys)

	return
}

func genStyle(style map[string]string) string {
	sortedKeys := getSortedKey(style)

	st := ""
	for _, k := range sortedKeys {
		v := style[k]
		st += fmt.Sprintf("%s: %s; ", k, v)
	}

	st = strings.Trim(st, " ")
	return st
}

func genAttr(attr map[string]string) string {
	sortedKeys := getSortedKey(attr)

	st := ""
	for _, k := range sortedKeys {
		v := attr[k]
		st += fmt.Sprintf("%s=\"%s\" ", k, v)
	}

	st = strings.Trim(st, " ")
	return st
}

func getStyleFromProps(styleProps interface{}) map[string]string {
	pm, ok := styleProps.(map[string]interface{})
	if !ok {
		return nil
	}
	st := map[string]string{}
	for k, v := range pm {
		st[k] = fmt.Sprintf("%v", v)
	}
	return st
}

// classProps: 支持 obj, array, string
func getClassFromProps(classProps interface{}) []string {
	if classProps == nil {
		return nil
	}
	switch t := classProps.(type) {
	case []string:
		return t
	case string:
		return []string{t}
	case map[string]interface{}:
		var c []string
		for k, v := range t {
			if interfaceToBool(v) {
				c = append(c, k)
			}
		}
		return c
	case []interface{}:
		var c []string
		for _, v := range t {
			cc := getClassFromProps(v)
			c = append(c, cc...)
		}

		return c
	}

	return nil
}

func lookInterface(data interface{}, key string) (desc interface{}) {
	m, ok := shouldLookInterface(data, key)
	if !ok {
		return nil
	}

	return m
}

var LookInterface = lookInterface

// 扩展map, 实现作用域
func extendMap(src map[string]interface{}, ext ...map[string]interface{}) (desc map[string]interface{}) {
	desc = make(map[string]interface{}, len(src))
	for k, v := range src {
		desc[k] = v
	}
	for _, m := range ext {
		for k, v := range m {
			desc[k] = v
		}
	}
	return desc
}

func lookInterfaceToSlice(data interface{}, key string) (desc []interface{}) {
	m, ok := shouldLookInterface(data, key)
	if !ok {
		return nil
	}

	return interface2Slice(m)
}

func interfaceToStr(s interface{}, escaped ...bool) (d string) {
	switch a := s.(type) {
	case int, string, float64:
		d = fmt.Sprintf("%v", a)
	default:
		bs, _ := json.Marshal(a)
		d = string(bs)
	}

	if len(escaped) == 1 && escaped[0] {
		d = escape(d)
	}
	return
}

var InterfaceToStr = interfaceToStr

// 字符串false,0 会被认定为false
func interfaceToBool(s interface{}) (d bool) {
	if s == nil {
		return false
	}
	switch a := s.(type) {
	case bool:
		return a
	case int, float64, float32, int8, int64, int32, int16:
		return a != 0
	case string:
		return a != "" && a != "false" && a != "0"
	default:
		return true
	}

	return
}

// 用于{{func(a)}}语法
func interfaceToFunc(s interface{}) (d Function) {
	if s == nil {
		return emptyFunc
	}

	switch a := s.(type) {
	case func(args ...interface{}) interface{}:
		return a
	case Function:
		return a
	default:
		panic(a)
		return emptyFunc
	}
}

func interface2Slice(s interface{}) (d []interface{}) {
	switch a := s.(type) {
	case []interface{}:
		return a
	case []map[string]interface{}:
		d = make([]interface{}, len(a))
		for i, v := range a {
			d[i] = v
		}
	case []int:
		d = make([]interface{}, len(a))
		for i, v := range a {
			d[i] = v
		}
	case []int64:
		d = make([]interface{}, len(a))
		for i, v := range a {
			d[i] = v
		}
	case []int32:
		d = make([]interface{}, len(a))
		for i, v := range a {
			d[i] = v
		}
	case []string:
		d = make([]interface{}, len(a))
		for i, v := range a {
			d[i] = v
		}
	case []float64:
		d = make([]interface{}, len(a))
		for i, v := range a {
			d[i] = v
		}
	}
	return
}

func shouldLookInterface(data interface{}, key string) (desc interface{}, exist bool) {
	m, isObj := data.(map[string]interface{})

	kk := strings.Split(key, ".")
	currKey := kk[0]

	// 如果是对象, 则继续查找下一级
	if len(kk) != 1 && isObj {
		c, ok := m[currKey]
		if !ok {
			return
		}
		return shouldLookInterface(c, strings.Join(kk[1:], "."))
	}

	if len(kk) == 1 {
		if isObj {
			c, ok := m[currKey]
			if !ok {
				return
			}
			return c, true
		} else {
			switch currKey {
			case "length":
				switch t := data.(type) {
				// string
				case string:
					return len(t), true
				default:
					// slice
					return len(interface2Slice(t)), true
				}
			}
		}
	} else {
		// key不只有一个, 但是data不是对象, 说明出现了undefined的问题, 直接return
		return
	}

	return
}

func escape(src string) string {
	return html.EscapeString(src)
}
`
