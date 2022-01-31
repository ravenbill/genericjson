package genericjson

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
)

// Any - just amother name for interface{}
type Any interface{}

// GenJSON -  base type for this package
type GenJSON struct{ Any }

// ErrJSONPathDoesNotExists - error - JSON path doen not exists
var ErrJSONPathDoesNotExists = errors.New("incorrect or not existent json path")

// FromGeneric - convert interface{} into GenJSON
func FromGeneric(any interface{}) GenJSON { return GenJSON{Any: any} }

// UnmarshalJSON - unmarshal bytes into GenJSON
func (jsonTree *GenJSON) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &jsonTree.Any)
}

// MarshalJSON - marshal GenJSON into bytes
func (jsonTree GenJSON) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonTree.Any)
}

// Bool - find boolean value by json path
func (jsonTree GenJSON) Bool(args ...interface{}) (retval bool, err error) {
	s, err := jsonTree.Unwind(args...)
	if err == nil {
		var ok bool
		retval, ok = s.Any.(bool)
		debug(fmt.Sprintf("Bool: s is '%v', retval is %v", s.Any, retval))
		if !ok {
			return false, fmt.Errorf("value %s '%v' is not a bool", args, s.Any)
		}
	}
	return
}

// Int - find integer value by json path
func (jsonTree GenJSON) Int(args ...interface{}) (retval int, err error) {
	s, err := jsonTree.Unwind(args...)
	if err == nil {
		r, ok := s.Any.(float64)
		debug(fmt.Sprintf("Int: s is '%v', retval is %v, %T", s.Any, r, s.Any))
		if !ok {
			return 0, fmt.Errorf("value %s '%v' is not an int", args, s.Any)
		}
		if math.Abs(r-math.Trunc(r)) > math.SmallestNonzeroFloat64 {
			return 0, fmt.Errorf("value '%v' is a float and not an int", s.Any)
		}
		retval = int(r)
	}
	return
}

// Float - find float value by json path
func (jsonTree GenJSON) Float(args ...interface{}) (retval float64, err error) {
	s, err := jsonTree.Unwind(args...)
	if err == nil {
		var ok bool
		retval, ok = s.Any.(float64)
		debug(fmt.Sprintf("Float: s is '%v', retval is %v", s.Any, retval))
		if !ok {
			return math.NaN(), fmt.Errorf("value %s '%v' is not an float64", args, s.Any)
		}
	}
	return
}

// Empty - check if GenJSON is empty
func (jsonTree GenJSON) Empty() bool {
	return jsonTree.Any == nil
}

// String - find string by json path
func (jsonTree GenJSON) String(args ...interface{}) (retval string, err error) {
	s, err := jsonTree.Unwind(args...)
	if err == nil {
		var ok bool
		retval, ok = s.Any.(string)
		debug(fmt.Sprintf("String: s is '%v', retval is %s", s.Any, retval))
		if !ok {
			return "", fmt.Errorf("value %s '%v' is not a string", args, s.Any)
		}
	}
	return
}

// Array - find array by Json path
func (jsonTree GenJSON) Array(args ...interface{}) (retval []interface{}, err error) {
	s, err := jsonTree.Unwind(args...)
	if err == nil {
		var ok bool
		retval, ok = s.Any.([]interface{})
		debug(fmt.Sprintf("String: s is '%v', retval is %s", s.Any, retval))
		if !ok {
			return []interface{}{}, fmt.Errorf("value %s '%v' is not an array", args, s.Any)
		}
	}
	return
}

// ArrayOrEmpty - find array by Json path or return empty array
func (jsonTree GenJSON) ArrayOrEmpty(args ...interface{}) []interface{} {
	if a, e := jsonTree.Array(args...); e == nil {
		return a
	}
	return []interface{}{}
}

// Unwind - return objest by json path
func (jsonTree GenJSON) Unwind(args ...interface{}) (s GenJSON, err error) {
	s = jsonTree
	for _, arg := range args {
		switch v := s.Any.(type) {
		case []interface{}:
			i, ok := arg.(int)
			if !ok {
				return s, fmt.Errorf("expected integer, found `%v`", arg)
			}
			if i < 0 || i >= len(v) {
				return s, fmt.Errorf("index out of bounds %d", i)
			}
			s.Any = v[i]
		case map[string]interface{}:
			str, ok := arg.(string)
			if !ok {
				return s, fmt.Errorf("expected string, found `%v`", arg)
			}
			s.Any = v[str]
		default:
			return s, fmt.Errorf("incorrect or not existent json path '%v'", args)
		}
	}
	return
}

// Delete - delete object by json path
func (jsonTree GenJSON) Delete(args ...interface{}) (err error) {
	if len(args) == 0 {
		return ErrJSONPathDoesNotExists
	}
	s := jsonTree
	if len(args) > 1 {
		s, err = jsonTree.Unwind(args[0 : len(args)-1]...)
		if err != nil {
			return
		}
	}
	arg := args[len(args)-1]
	switch v := s.Any.(type) {
	case []interface{}:
		i, ok := arg.(int)
		if !ok {
			return fmt.Errorf("expected integer, found `%v`", arg)
		}
		if i < 0 || i >= len(v) {
			return fmt.Errorf("index out of bounds %d", i)
		}
		vn := make([]interface{}, len(v)-1)
		for k, idx := 0, 0; k < len(v); k++ {
			if i != k {
				vn[idx] = v[k]
				idx++
			}
		}
		jsonTree.Set(append([]interface{}{vn}, args[0:len(args)-1]...)...)
	case map[string]interface{}:
		str, ok := arg.(string)
		if !ok {
			return fmt.Errorf("expected string, found `%v`", arg)
		}
		delete(v, str)
	default:
		return ErrJSONPathDoesNotExists
	}
	return nil
}

// Clone an whole tree
func (jsonTree GenJSON) Clone() (s GenJSON) {
	b, _ := jsonTree.MarshalJSON()
	s.UnmarshalJSON(b)
	return
}

// ScanObject - scan Json tree and select objects by predicate
func (jsonTree GenJSON) ScanObject(predicate func(interface{}) bool,
	args ...interface{}) (GenJSON, []interface{}, bool) {
	path, s := []interface{}{}, jsonTree
	for idx, arg := range args {
		if len(args)-1 == idx {
			if predicate(s.Any) {
				return s, path, true
			}
		}
		path = append(path, arg)
		//log.Printf("ScanObject %v %v", arg, s )
		switch v := s.Any.(type) {
		case []interface{}:
			i, ok := arg.(int)
			if !ok {
				return s, path, false
			}
			if i == -1 {
				if len(args)-1 == idx {
					for j, obj := range v {
						s.Any = obj
						if predicate(obj) {
							path[len(path)-1] = j
							return s, path, true
						}
					}
				} else {
					rest := args[(idx + 1):]
					for j, obj := range v {
						s.Any = obj
						retval, par, ok := s.ScanObject(predicate, rest...)
						if ok {
							path[len(path)-1] = j
							path = append(path, par...)
							return retval, path, ok
						}
					}
				}
			} else if i < 0 || i >= len(v) {
				return s, path, false
			} else {
				s1, p1, ok := s.ScanObject(predicate, args[(idx+1):]...)
				if ok {
					path := append(path, p1...)
					return s1, path, ok
				}
			}
		case map[string]interface{}:
			str, ok := arg.(string)
			if !ok {
				return s, path, false
			}
			s.Any = v[str]
		default:
			return s, path, false
		}
	}
	return s, path, false
}

// Set value by json path, the value is FIRST parameter
func (jsonTree GenJSON) Set(args ...interface{}) error {
	s, err := jsonTree.Unwind(args[1:(len(args) - 1)]...)
	val := args[0]
	if err == nil {
		arg := args[len(args)-1]
		switch v := s.Any.(type) {
		case []interface{}:
			i, ok := arg.(int)
			if !ok {
				return fmt.Errorf("expected integer, found `%v`", arg)
			}
			if i < 0 || i >= len(v) {
				return fmt.Errorf("index out of bounds %d", i)
			}
			v[i] = val
			return nil
		case map[string]interface{}:
			str, ok := arg.(string)
			if !ok {
				return fmt.Errorf("expected string, found `%v`", arg)
			}
			v[str] = val
			return nil
		default:
			return ErrJSONPathDoesNotExists
		}
	}
	return err
}

var d = false

// SetDebug - set Debug flag for this package
func SetDebug(b bool) { d = b }

func debug(message string) {
	if d {
		log.Println(message)
	}
}
