package goatee

import (
	"fmt"
	"html"
	"reflect"
	"regexp"
	"strconv"
)

// temporary
var _ = fmt.Println

type context struct {
	label      string
	method     string
	start      int
	end        int
	inner      []byte
	innerStart int
	innerEnd   int
	contexts   []*context
}

type Args struct {
	Template []byte
	Data     interface{}
	Partials map[string][]byte
	Globals  interface{}
}

func Fill(a Args) ([]byte, error) {
	regex := regexp.MustCompile("\\{\\{([#!:%\\/-]?)(.*?)\\}\\}")

	c := &context{start: 0, inner: a.Template, innerStart: 0, innerEnd: len(a.Template), end: len(a.Template)}
	currc := c
	prevc := make([]*context, 0)
	matches := regex.FindAllSubmatchIndex(a.Template, -1)
	// fmt.Println("matches:", len(matches))
	for _, value := range matches {
		method := string(a.Template[value[2]:value[3]])
		label := string(a.Template[value[4]:value[5]])

		//fmt.Println("method:", method, "label:", label)
		if method != "/" {
			currc.contexts = append(currc.contexts, &context{label: label, method: method, start: value[0], end: value[1], innerStart: value[1]})

			if method != "" && method != "%" {
				//fmt.Println("oy should not be here")
				prevc = append(prevc, currc)
				currc = currc.contexts[len(currc.contexts)-1]
			}
		} else {
			currc.end = value[1]
			currc.innerEnd = value[0]
			currc.inner = a.Template[currc.innerStart:currc.innerEnd]
			currc = prevc[len(prevc)-1]
			prevc = prevc[:len(prevc)-1]
		}
	}

	//fmt.Println("c.contexts:", len(c.contexts))
	//fmt.Println("currc.contexts:", len(currc.contexts))

	bytes := processContexts(a.Template, c, []interface{}{a.Data}, a.Partials, a.Globals)

	return bytes, nil
}

func processContexts(template []byte, c *context, d []interface{}, partials map[string][]byte, globals interface{}) []byte {
	var mydata reflect.Value
	var hasprop bool
	// var final []byte

	rtn := make([]byte, 0)
	position := c.innerStart
	for _, currc := range c.contexts {
		rtn = append(rtn, template[position:currc.start]...)
		position = currc.end

		if currc.method == "-" {
			// to do later
		}

		first := string(currc.label[0])
		if first == "*" {
			// check for global variables
			mydata, hasprop = getProp(globals, string(currc.label[1:]))
		} else {
			// get last data element
			tempdata := d[len(d)-1]
			mydata, hasprop = getProp(tempdata, currc.label)
		}

		switch currc.method {
		case "", "%":
			//fmt.Println("standard tag")
			if hasprop && !isFalsy(mydata) {
				// convert props to byte arrays to fill template
				var final []byte
				switch mydata.Type().Kind() {
				case reflect.Int:
					final = []byte(strconv.FormatInt(mydata.Int(), 10))
				case reflect.String:
					final = []byte(mydata.String())
				default:
					fmt.Println("NO DATA")
					fmt.Println(currc.label)
					final = make([]byte, 0)
				}

				if currc.method == "%" {
					final = []byte(html.EscapeString(string(final)))
				}

				rtn = append(rtn, final...)
			}
		case "#":
			// fmt.Printf("type: #, label: %+v, hasprop: %+v, isfalsy: %+v, data: %+v", currc.label, hasprop, isFalsy(mydata), mydata)
			if hasprop && !isFalsy(mydata) {
				//switch reflect.TypeOf(mydata.Interface()).Kind() {
				switch mydata.Kind() {
				case reflect.Struct, reflect.Map, reflect.String:
					newData := append(d, mydata.Interface())
					rtn = append(rtn, processContexts(template, currc, newData, partials, globals)...)
				case reflect.Slice:
					for i := 0; i < mydata.Len(); i++ {
						newData := append(d, mydata.Index(i).Interface())
						rtn = append(rtn, processContexts(template, currc, newData, partials, globals)...)
					}
				case reflect.Array:
					// not built yet
				default:
					fmt.Println("DEFAULT CASE")
					fmt.Println(currc.label)
					fmt.Printf("Type: %+v", mydata.Kind())
				}
			}
		case ":":
			if hasprop && !isFalsy(mydata) {
				rtn = append(rtn, processContexts(template, currc, d, partials, globals)...)
			}
		case "!":
			if !hasprop || isFalsy(mydata) {
				rtn = append(rtn, processContexts(template, currc, d, partials, globals)...)
			}
		}
	}

	if position < c.end {
		rtn = append(rtn, template[position:c.innerEnd]...)
	}

	return rtn
}

func getProp(d interface{}, label string) (reflect.Value, bool) {
	var value reflect.Value
	var ptr reflect.Value
	var nullValue reflect.Value

	value = reflect.ValueOf(d)
	if !value.IsValid() {
		return value, false
	} else if value.Kind() == reflect.Ptr {
		ptr = value
		value = ptr.Elem()
	} else {
		ptr = reflect.New(reflect.TypeOf(d))
		temp := ptr.Elem()
		temp.Set(value)
	}

	method := value.MethodByName(label)
	if method.IsValid() {
		return getInner(method.Call([]reflect.Value{})[0]), true
	}

	method = ptr.MethodByName(label)
	if method.IsValid() {
		return getInner(method.Call([]reflect.Value{})[0]), true
	}

	switch value.Kind() {
	case reflect.Struct:
		field := value.FieldByName(label)
		if field.IsValid() {
			return getInner(field), true
		}
	case reflect.Map:
		field := value.MapIndex(reflect.ValueOf(label))
		if field.IsValid() {
			return getInner(field), true
		}
	}

	return nullValue, false
}

func getInner(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Interface {
		return reflect.ValueOf(v.Interface())
	}

	if v.Kind() == reflect.Ptr {
		return v.Elem()
	}

	return v
}

func isFalsy(d reflect.Value) bool {
	switch d.Kind() {
	case reflect.Slice, reflect.Array, reflect.String, reflect.Map:
		return d.Len() == 0
	case reflect.Struct:
		return false
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return d.Int() == 0
	case reflect.Bool:
		return !d.Bool()
	default:
		return true
	}

	return true
}
