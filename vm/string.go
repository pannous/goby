package vm

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// StringObject represents string instances
// String object holds and manipulates a sequence of characters.
// String objects may be created using as string literals.
// Double or single quotations can be used for representation.
//
// ```ruby
// a = "Three"
// b = 'zero'
// c = '漢'
// d = 'Tiếng Việt'
// e = "😏️️"
// ```
//
// **Note:**
//
// - Currently, manipulations are based upon Golang's Unicode manipulations.
// - Currently, UTF-8 encoding is assumed based upon Golang's string manipulation, but the encoding is not actually specified(TBD).
// - `String.new` is not supported.
type StringObject struct {
	*baseObj
	Value string
}

func (vm *VM) initStringObject(value string) *StringObject {
	replacer := strings.NewReplacer("\\n", "\n", "\\r", "\r", "\\t", "\t", "\\v", "\v", "\\\\", "\\")
	return &StringObject{
		baseObj: &baseObj{class: vm.topLevelClass(stringClass)},
		Value:   replacer.Replace(value),
	}
}

func (vm *VM) initStringClass() *RClass {
	sc := vm.initializeClass(stringClass, false)
	sc.setBuiltInMethods(builtinStringInstanceMethods(), false)
	sc.setBuiltInMethods(builtInStringClassMethods(), true)
	return sc
}

func builtinStringInstanceMethods() []*BuiltInMethodObject {
	return []*BuiltInMethodObject{

		{
			// Returns the concatenation of self and another String
			//
			// ```Ruby
			// "first" + "-second" # => "first-second"
			// ```
			//
			// @return [String]
			Name: "+",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {

					leftValue := receiver.(*StringObject).Value
					right, ok := args[0].(*StringObject)

					if !ok {
						return wrongTypeError(receiver.Class())
					}

					rightValue := right.Value
					return t.vm.initStringObject(leftValue + rightValue)
				}
			},
		},
		{
			// Returns self multiplying another Integer
			//
			// ```Ruby
			// "string " * 2 # => "string string string "
			// ```
			//
			// @return [String]
			Name: "*",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {

					leftValue := receiver.(*StringObject).Value
					right, ok := args[0].(*IntegerObject)

					if !ok {
						return wrongTypeError(receiver.Class())
					}

					if right.Value < 0 {
						return newError("Second argument must be greater than or equal to 0 String#*")
					}

					var result string

					for i := 0; i < right.Value; i++ {
						result += leftValue
					}

					return t.vm.initStringObject(result)
				}
			},
		},
		{
			// Returns a Boolean if first string greater than second string
			//
			// ```Ruby
			// "a" < "b" # => true
			// ```
			//
			// @return [Boolean]
			Name: ">",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {

					leftValue := receiver.(*StringObject).Value
					right, ok := args[0].(*StringObject)

					if !ok {
						return wrongTypeError(receiver.Class())
					}

					rightValue := right.Value

					if leftValue > rightValue {
						return TRUE
					}

					return FALSE
				}
			},
		},
		{
			// Returns a Boolean if first string less than second string
			//
			// ```Ruby
			// "a" < "b" # => true
			// ```
			//
			// @return [Boolean]
			Name: "<",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {

					leftValue := receiver.(*StringObject).Value
					right, ok := args[0].(*StringObject)

					if !ok {
						return wrongTypeError(receiver.Class())
					}

					rightValue := right.Value

					if leftValue < rightValue {
						return TRUE
					}

					return FALSE
				}
			},
		},
		{
			// Returns a Boolean of compared two strings
			//
			// ```Ruby
			// "first" == "second" # => false
			// "two" == "two" # => true
			// ```
			//
			// @return [Boolean]
			Name: "==",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {

					leftValue := receiver.(*StringObject).Value
					right, ok := args[0].(*StringObject)

					if !ok {
						return wrongTypeError(receiver.Class())
					}

					rightValue := right.Value

					if leftValue == rightValue {
						return TRUE
					}

					return FALSE
				}
			},
		},
		{
			// Returns a Integer. If first string is less than second string returns -1, if equal to returns 0, if greater returns 1
			//
			//
			// ```Ruby
			// "abc" <=> "abcd" # => -1
			// "abc" <=> "abc" # => 0
			// "abcd" <=> "abc" # => 1
			// ```
			//
			// @return [Integer]
			Name: "<=>",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {

					leftValue := receiver.(*StringObject).Value
					right, ok := args[0].(*StringObject)

					if !ok {
						return wrongTypeError(receiver.Class())
					}

					rightValue := right.Value

					if leftValue < rightValue {
						return t.vm.initIntegerObject(-1)
					}
					if leftValue > rightValue {
						return t.vm.initIntegerObject(1)
					}

					return t.vm.initIntegerObject(0)
				}
			},
		},
		{
			// Returns a Boolean of compared two strings
			//
			// ```Ruby
			// "first" != "second" # => true
			// "two" != "two" # => false
			// ```
			//
			// @return [Boolean]
			Name: "!=",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {

					leftValue := receiver.(*StringObject).Value
					right, ok := args[0].(*StringObject)

					if !ok {
						return wrongTypeError(receiver.Class())
					}

					rightValue := right.Value

					if leftValue != rightValue {
						return TRUE
					}

					return FALSE
				}
			},
		},
		{
			// Return a new String with the first character converted to uppercase but the rest of string converted to lowercase.
			//
			// ```Ruby
			// "test".capitalize         # => "Test"
			// "tEST".capitalize         # => "Test"
			// "heLlo\nWoRLd".capitalize # => "Hello\nworld"
			// ```
			//
			// @return [String]
			Name: "capitalize",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {

					str := []byte(receiver.(*StringObject).Value)
					start := string(str[0])
					rest := string(str[1:])
					result := strings.ToUpper(start) + strings.ToLower(rest)

					return t.vm.initStringObject(result)
				}
			},
		},
		{
			// Returns a new String with all characters is upcase
			//
			// ```Ruby
			// "very big".upcase # => "VERY BIG"
			// ```
			//
			// @return [String]
			Name: "upcase",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {

					str := receiver.(*StringObject).Value

					return t.vm.initStringObject(strings.ToUpper(str))
				}
			},
		},
		{
			// Returns a new String with all characters is lowercase
			//
			// ```Ruby
			// "erROR".downcase        # => "error"
			// "HeLlO\tWorLD".downcase # => "hello\tworld"
			// ```
			//
			// @return [String]
			Name: "downcase",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {

					str := receiver.(*StringObject).Value

					return t.vm.initStringObject(strings.ToLower(str))
				}
			},
		},
		{
			// Returns the character length of self
			// **Note:** the length is currently byte-based, instead of charcode-based.
			//
			// ```Ruby
			// "zero".size # => 4
			// "".size # => 0
			// ```
			//
			// @return [Integer]
			Name: "size",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {

					str := receiver.(*StringObject).Value

					return t.vm.initIntegerObject(len(str))
				}
			},
		},
		{
			// Returns the character length of self
			// **Note:** the length is currently byte-based, instead of charcode-based.
			//
			// ```Ruby
			// "zero".size # => 4
			// "".size # => 0
			// ```
			//
			// @return [Integer]
			Name: "length",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {

					str := receiver.(*StringObject).Value

					return t.vm.initIntegerObject(len(str))
				}
			},
		},
		{
			// Returns a new String with reverse order of self
			// **Note:** the length is currently byte-based, instead of charcode-based.
			//
			// ```Ruby
			// "reverse".reverse      # => "esrever"
			// "Hello\nWorld".reverse # => "dlroW\nolleH"
			// ```
			//
			// @return [String]
			Name: "reverse",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {

					str := receiver.(*StringObject).Value

					var revert string
					for i := len(str) - 1; i >= 0; i-- {
						revert += string(str[i])
					}

					return t.vm.initStringObject(revert)
				}
			},
		},
		{
			// Returns a new String with self value
			//
			// ```Ruby
			// "string".to_s # => "string"
			// ```
			//
			// @return [String]
			Name: "to_s",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {

					str := receiver.(*StringObject).Value

					return t.vm.initStringObject(str)
				}
			},
		},
		{
			// Returns the result of converting self to Integer
			//
			// ```Ruby
			// "123".to_i # => 123
			// "3d print".to_i # => 3
			// "some text".to_i # => 0
			// ```
			//
			// @return [Integer]
			Name: "to_i",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {

					str := receiver.(*StringObject).Value
					parsedStr, err := strconv.ParseInt(str, 10, 0)

					if err == nil {
						return t.vm.initIntegerObject(int(parsedStr))
					}

					var digits string
					for _, char := range str {
						if unicode.IsDigit(char) {
							digits += string(char)
						} else {
							break
						}
					}

					if len(digits) > 0 {
						parsedStr, _ = strconv.ParseInt(digits, 10, 0)
						return t.vm.initIntegerObject(int(parsedStr))
					}

					return t.vm.initIntegerObject(0)
				}
			},
		},
		{
			// Checks if the specified string is included in the receiver
			//
			// ```Ruby
			// "Hello\nWorld".include("\n") # => true
			// ```
			//
			// @return [Bool]
			Name: "include",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {

					rcv := receiver.(*StringObject).Value
					arg := args[0].(*StringObject).Value

					if strings.Contains(rcv, arg) {
						return TRUE
					}

					return FALSE
				}
			},
		},
	}
}

func builtInStringClassMethods() []*BuiltInMethodObject {
	return []*BuiltInMethodObject{
		{
			Name: "new",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {
					return t.UnsupportedMethodError("#new", receiver)
				}
			},
		},
		{
			// Returns a string which is concatenate with the input string or character
			//
			// ```ruby
			// "Hello ".concat("World") # => "Hello World"
			// ```
			// @return [String]
			Name: "concat",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {

					str := receiver.(*StringObject).Value
					concatStr, ok := args[0].(*StringObject)

					if !ok {
						return wrongTypeError(receiver.returnClass())
					}

					return t.vm.initStringObject(str + concatStr.Value)
				}
			},
		},
		{
			// Returns the character of the string with specified index
			// It will raise error if the input is not an Integer type
			//
			// ```ruby
			// "Hello"[1]        # => "e"
			// "Hello"[5]        # => nil
			// "Hello\nWorld"[5] # => "\n"
			// "Hello"[-1]       # => "o"
			// "Hello"[-6]       # => nil
			// ```
			//
			// @return [String]
			Name: "[]",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {
					if len(args) != 1 {
						return &Error{Message: "Expect 1 arguments. got=%d" + string(len(args))}
					}

					str := receiver.(*StringObject).Value
					i := args[0]
					index, ok := i.(*IntegerObject)
					indexValue := index.Value

					if !ok {
						return initErrorObject(TypeErrorClass, "Expect Integer. got=%T (%+v)", args[0], args[0])
					}

					if indexValue < 0 {
						if -indexValue > len(str) {
							return NULL
						}
						return t.vm.initStringObject(string([]rune(str)[len(str)+indexValue]))
					}

					if len(str) > indexValue {
						return t.vm.initStringObject(string([]rune(str)[indexValue]))
					}
					return NULL
				}
			},
		},
		{
			// Replace character of the string with input string
			// It will raise error if the index is not Integer type or the index value is out of
			// range of the string length
			//
			// ```ruby
			// "Ruby"[1] = "oo" # => "Rooby"
			// "Go"[2] = "by"   # => "Goby"
			// "Hello\nWorld"[5] = " " # => "Hello World"
			// "Ruby"[-3] = "oo" # => "Rooby"
			// ```
			//
			// @return [String]
			Name: "[]=",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {

					str := receiver.(*StringObject).Value
					i := args[0]
					index, ok := i.(*IntegerObject)
					indexValue := index.Value

					if !ok {
						return initErrorObject(TypeErrorClass, "Expect Integer. got=%T (%+v)", args[0], args[0])
					}

					strLength := len(str)

					if strLength < indexValue {
						return newError("Index value out of range. got=%T", i)
					}

					replaceStr := args[1].(*StringObject).Value

					// Negative Index Case
					if indexValue < 0 {
						if -indexValue > strLength {
							return newError("Index value out of range. got=%T", i)
						}

						result := str[:strLength+indexValue] + replaceStr + str[strLength+indexValue+1:]
						return t.vm.initStringObject(result)
					}

					if strLength == indexValue {
						return t.vm.initStringObject(str + replaceStr)
					}
					result := str[:indexValue] + replaceStr + str[indexValue+1:]
					return t.vm.initStringObject(result)
				}
			},
		},
		{
			// Returns an array of characters converted from a string
			// ```ruby
			// "Goby".to_a # => ["G", "o", "b", "y"]
			// ```
			// @return [String]
			Name: "to_a",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {

					str := receiver.(*StringObject)
					elems := []Object{}

					for i := 0; i < len(str.Value); i++ {
						elems = append(elems, t.vm.initIntegerObject(i))
					}

					return t.vm.initArrayObject(elems)
				}
			},
		},
		{
			// Returns the integer that count the string chars as UTF-8
			//
			// ```ruby
			// "abcde".count        # => 5
			// "哈囉！世界！".count   # => 6
			// "Hello\nWorld".count # => 11
			// ```
			//
			// @return [Integer]
			Name: "count",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {

					str := receiver.(*StringObject).Value

					return t.vm.initIntegerObject(utf8.RuneCountInString(str))
				}
			},
		},
		{
			// Returns true if string is empty value
			//
			// ```ruby
			// "".empty      # => true
			// "Hello".empty # => false
			// ```
			//
			// @return [Boolean]
			Name: "empty",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {

					str := receiver.(*StringObject).Value

					if str == "" {
						return TRUE
					}
					return FALSE
				}
			},
		},
		{
			// Returns true if receiver string is equal to argument string
			//
			// ```ruby
			// "Hello".eql("Hello") # => true
			// "Hello".eql("World") # => false
			// ```
			// @return [Boolean]
			Name: "eql",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {

					str := receiver.(*StringObject).Value
					compareStr, ok := args[0].(*StringObject)

					if !ok {
						return wrongTypeError(receiver.returnClass())
					}

					if compareStr.Value == str {
						return TRUE
					}
					return FALSE
				}
			},
		},
		{
			// Returns true if receiver string start with the argument string
			//
			// ```ruby
			// "Hello".start_with("Hel") # => true
			// "Hello".start_with("hel") # => false
			// ```
			// @return [Boolean]
			Name: "start_with",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {

					str := receiver.(*StringObject).Value
					compareStr, ok := args[0].(*StringObject)

					if !ok {
						return wrongTypeError(receiver.returnClass())
					}

					index := len(compareStr.Value) - 1
					if compareStr.Value == str[:index] {
						return TRUE
					}
					return FALSE
				}
			},
		},
		{
			// Returns true if receiver string end with the argument string
			//
			// ```ruby
			// "Hello".end_with("llo") # => true
			// "Hello".end_with("ell") # => false
			// ```
			// @return [Boolean]
			Name: "end_with",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {

					str := receiver.(*StringObject).Value
					compareStr, ok := args[0].(*StringObject)

					if !ok {
						return wrongTypeError(receiver.returnClass())
					}

					index := len(compareStr.Value)
					if compareStr.Value == str[index:] {
						return TRUE
					}
					return FALSE
				}
			},
		},
		{
			// Insert a string input in specified index value of the receiver string
			//
			// It will raise error if index value is not an integer or index value is out
			// of receiver string's range
			//
			// It will also raise error if the input string value is not type string
			//
			// ```ruby
			// "Hello".insert(0, "X") # => "XHello"
			// "Hello".insert(2, "X") # => "HeXllo"
			// "Hello".insert(5, "X") # => "HelloX"
			// "Hello".insert(-1, "X") # => "HelloX"
			// "Hello".insert(-3, "X") # => "HelXlo"
			// ```
			// @return [String]
			Name: "insert",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {

					str := receiver.(*StringObject).Value
					i := args[0]
					index, ok := i.(*IntegerObject)
					indexValue := index.Value

					if !ok {
						return newError("Expect index argument to be Integer. got=%T", i)
					}

					insertStr, ok := args[1].(*StringObject)

					if !ok {
						return newError("Expect insert string to be String type, got=%T", insertStr)
					}
					strLength := len(str)

					if indexValue < 0 {
						if -indexValue > strLength+1 {
							return newError("Index value out of range. got=" + string(indexValue))
						} else if -indexValue == strLength+1 {
							return t.vm.initStringObject(insertStr.Value + str)
						}
						return t.vm.initStringObject(str[:strLength+indexValue] + insertStr.Value + str[strLength+indexValue:])
					}

					if strLength < indexValue {
						return newError("Index value out of range. got=" + string(indexValue))
					}

					return t.vm.initStringObject(str[:indexValue] + insertStr.Value + str[indexValue:])
				}
			},
		},
		{
			// Returns a string which is being partially deleted with specified values
			//
			// ```ruby
			// "Hello hello HeLlo".delete("el") # => "Hlo hlo HeLlo"
			// # TODO: Handle delete intersection of multiple strings' input case
			// "Hello hello HeLlo".delete("el", "e") # => "Hllo hllo HLlo"
			// ```
			//
			// @return [String]
			Name: "delete",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {

					str := receiver.(*StringObject).Value
					deleteStr, ok := args[0].(*StringObject)

					if !ok {
						return wrongTypeError(receiver.returnClass())
					}

					return t.vm.initStringObject(strings.Replace(str, deleteStr.Value, "", -1))
				}
			},
		},
		{
			// Returns a string with the last character chopped
			//
			// ```ruby
			// "Hello".chop # => "Hell"
			// "Hello World\n".chop => "Hello World"
			// ```
			//
			// @return [String]
			Name: "chop",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {

					str := receiver.(*StringObject).Value

					return t.vm.initStringObject(str[:len(str)-1])
				}
			},
		},
		{
			// If input integer is greater than the length of receiver string, returns a new String of
			// length integer with receiver string left justified and padded with default " "; otherwise,
			// returns receiver string.
			//
			// It will raise error if the input string length is not integer type
			//
			// ```ruby
			// "Hello".ljust(2)        # => "Hello"
			// "Hello".ljust(7)        # => "Hello  "
			// "Hello".ljust(10, "xo") # => "Helloxoxox"
			// ```
			//
			// @return [String]
			Name: "ljust",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {

					str := receiver.(*StringObject).Value

					l := args[0]
					strLength, ok := l.(*IntegerObject)
					strLengthValue := strLength.Value

					if !ok {
						return initErrorObject(TypeErrorClass, "Expect justify width is Integer type. got=%T", l)
					}

					var padStringValue string
					if len(args) == 1 {
						padStringValue = " "
					} else {
						p := args[1]
						padString, ok := p.(*StringObject)

						if !ok {
							return newError("Expect second argument is String type. got=%T", p)
						}

						padStringValue = padString.Value

						if strLengthValue > len(str) {
							for i := len(str); i < strLengthValue; i += len(padStringValue) {
								str += padStringValue
							}
							str = str[:strLengthValue]
						}
					}

					return t.vm.initStringObject(str)
				}
			},
		},
		{
			// If input integer is greater than the length of receiver string, returns a new String of
			// length integer with receiver string right justified and padded with default " "; otherwise,
			// returns receiver string.
			//
			// It will raise error if the input string length is not integer type
			//
			// ```ruby
			// "Hello".rjust(2) # => "Hello"
			// "Hello".rjust(7) # => "  Hello"
			// "Hello".rjust(10, "xo") => "xoxoxHello"
			// ```
			//
			// @return [String]
			Name: "rjust",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {

					str := receiver.(*StringObject).Value

					l := args[0]
					strLength, ok := l.(*IntegerObject)
					strLengthValue := strLength.Value

					if !ok {
						return initErrorObject(TypeErrorClass, "Expect justify width is Integer type. got=%T", l)
					}

					var padStringValue string
					if len(args) == 1 {
						padStringValue = " "
					} else {
						p := args[1]
						padString, ok := p.(*StringObject)

						if !ok {
							return newError("Expect second argument is String type. got=%T", p)
						}

						padStringValue = padString.Value
					}

					if strLengthValue > len(str) {
						origin := str
						for i := len(str); i < strLengthValue; i += len(padStringValue) {
							str = padStringValue + str
						}
						if len(str) > strLengthValue {
							chopLength := len(str) - strLengthValue
							str = str[:len(str)-len(origin)-chopLength] + origin
						}
					}

					return t.vm.initStringObject(str)
				}
			},
		},
		{
			// Returns a copy of str with leading and trailing whitespace removed.
			// Whitespace is defined as any of the following characters: null, horizontal tab,
			// line feed, vertical tab, form feed, carriage return, space.
			//
			// ```ruby
			// "  Goby Lang  ".strip   # => "Goby Lang"
			// "\nGoby Lang\r\t".strip # => "Goby Lang"
			// ```
			//
			// @return [String]
			Name: "strip",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {

					str := receiver.(*StringObject).Value

					for {
						str = strings.Trim(str, " ")

						if strings.HasPrefix(str, "\\n") || strings.HasPrefix(str, "\\t") || strings.HasPrefix(str, "\\r") {
							str = str[2:]
							continue
						}
						if strings.HasSuffix(str, "\\n") || strings.HasSuffix(str, "\\t") || strings.HasSuffix(str, "\\r") {
							str = str[:len(str)-4]
							continue
						}
						break
					}
					return t.vm.initStringObject(str)
				}
			},
		},
		{
			// Returns an array of strings separated by the given separator
			//
			// ```ruby
			// "Hello World".split("o") # => ["Hell", " W", "rld"]
			// "Goby".split("")         # => ["G", "o", "b", "y"]
			// # TODO: Whitespace carriage return case
			// ```
			//
			// @return [Array]
			Name: "split",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {

					s := args[0]
					seperator, ok := s.(*StringObject)

					if !ok {
						return wrongTypeError(receiver.returnClass())
					}

					str := receiver.(*StringObject).Value
					arr := strings.Split(str, seperator.Value)

					var elements []Object
					for i := 0; i < len(arr); i++ {
						elements = append(elements, t.vm.initStringObject(arr[i]))
					}

					return t.vm.initArrayObject(elements)
				}
			},
		},
		{
			// Returns a string sliced according to the input range
			//
			// ```ruby
			// "Hello World".slice(1..6)    # => "ello W"
			// "1234567890".slice(6..1)     # => ""
			// "1234567890".slice(-10..1)   # => "12"
			// "1234567890".slice(-10..-1)  # => "1234567890"
			// "1234567890".slice(-10..-11) # => ""
			// "1234567890".slice(1..-1)    # => "234567890"
			// "1234567890".slice(1..-1234) # => ""
			// "1234567890".slice(-11..5)   # => nil
			// "1234567890".slice(-11..-12) # => nil
			// "Hello World".slice(4)       # => "o"
			// "Hello\nWorld".slice(6)      # => "\n"
			// ```
			//
			// @return [String]
			Name: "slice",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {

					str := receiver.(*StringObject).Value
					strLength := len(str)

					switch args[0].(type) {
					case *RangeObject:
						ran := args[0].(*RangeObject)
						switch {
						case ran.Start >= 0 && ran.End >= 0:
							if ran.Start > strLength {
								return NULL
							} else if ran.Start > ran.End {
								return t.vm.initStringObject("")
							}
							return t.vm.initStringObject(str[ran.Start: ran.End+1])
						case ran.Start < 0 && ran.End >= 0:
							positiveStart := strLength + ran.Start
							if -ran.Start > strLength {
								return NULL
							} else if positiveStart > ran.End {
								return t.vm.initStringObject("")
							}
							return t.vm.initStringObject(str[positiveStart: ran.End+1])
						case ran.Start >= 0 && ran.End < 0:
							positiveEnd := strLength + ran.End
							if ran.Start > strLength {
								return NULL
							} else if positiveEnd < 0 || ran.Start > positiveEnd {
								return t.vm.initStringObject("")
							}
							return t.vm.initStringObject(str[ran.Start: positiveEnd+1])
						case ran.Start < 0 && ran.End < 0:
							positiveStart := strLength + ran.Start
							positiveEnd := strLength + ran.End
							if positiveStart > strLength {
								return NULL
							} else if positiveStart > positiveEnd {
								return t.vm.initStringObject("")
							}
							return t.vm.initStringObject(str[positiveStart: positiveEnd+1])
						default:
							return NULL
						}

					case *IntegerObject:
						intValue := args[0].(*IntegerObject).Value
						if intValue < 0 {
							if -intValue > strLength {
								return NULL
							}
							return t.vm.initStringObject(string([]rune(str)[strLength+intValue]))
						}
						if intValue > strLength {
							return NULL
						}
						return t.vm.initStringObject(string([]rune(str)[intValue]))

					default:
						return newError("Expect first argument is Range or Integer type. got=%T", args[0])
					}
				}
			},
		},
		{
			// Return a string replaced by the input string
			//
			// ```ruby
			// "Hello".replace("World")          # => "World"
			// "你好"replace("再見")              # => "再見"
			// "Ruby\nLang".replace("Goby\nLang") # => "Goby Lang"
			// ```
			//
			// @return [String]
			Name: "replace",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {

					s := args[0]
					replaceStr, ok := s.(*StringObject)

					if !ok {
						return wrongTypeError(receiver.returnClass())
					}

					return t.vm.initStringObject(replaceStr.Value)
				}
			},
		},
		//{
		//	// Doc
		//	Name: "gsub",
		//	Fn: func(receiver Object) builtinMethodBody {
		//		return func(t *thread, args []Object, blockFrame *callFrame) Object {
		//
		//			str := receiver.(*StringObject).Value
		//
		//			return initStringObject(str)
		//		}
		//	},
		//}
	}
}

// Polymorphic helper functions -----------------------------------------

// toString just returns the value of string.
func (s *StringObject) toString() string {
	return s.Value
}

// toJSON converts the receiver into JSON string.
func (s *StringObject) toJSON() string {
	return "\"" + s.Value + "\""
}

func (s *StringObject) equal(e *StringObject) bool {
	return s.Value == e.Value
}
