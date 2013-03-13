package gherkin

import (
    re "regexp"
    "io"
    "reflect"
    "strconv"
)

type stepdef struct {
    r *re.Regexp
    f interface{}
}

func (s stepdef) call(w *World) {
    t := reflect.TypeOf(s.f)
    in := make([]reflect.Value, t.NumIn())
    in[0] = reflect.ValueOf(w)
    if len(in) != len(w.regexParams) {
        panic("Function type mismatch")
    }
    for i := 1; i < len(in); i++ {
        var val interface{}
        var err error
        itp := w.regexParams[i]
        switch t.In(i).Kind() {
        case reflect.Bool:
            val, err = strconv.ParseBool(itp)
        case reflect.Int8:
            val, err = strconv.ParseInt(itp, 10, 8)
            val = int8(val.(int64))
        case reflect.Int16:
            val, err = strconv.ParseInt(itp, 10, 16)
            val = int16(val.(int64))
        case reflect.Int32:
            val, err = strconv.ParseInt(itp, 10, 32)
            val = int32(val.(int64))
        case reflect.Int:
            val, err = strconv.ParseInt(itp, 10, 64)
            val = int(val.(int64))
        case reflect.Int64:
            val, err = strconv.ParseInt(itp, 10, 64)
            val = val.(int64)
        case reflect.Float32:
            val, err = strconv.ParseFloat(itp, 32)
            val = float32(val.(float64))
        case reflect.Float64:
            val, err = strconv.ParseFloat(itp, 64)
            val = val.(float64)
        case reflect.String:
            val = itp
        default:
            panic("Function type not supported")
        }
        if err != nil {
            panic(err)
        }
        in[i] = reflect.ValueOf(val)
    }
    r := reflect.ValueOf(s.f)
    r.Call(in)
}

func createstepdef(p string, f interface{}) stepdef {
    r, _ := re.Compile(p)
    return stepdef{r, f}
}

func (s stepdef) execute(line *step, output io.Writer) bool {
    if s.r.MatchString(line.String()) {
        if s.f != nil {
            substrs := s.r.FindStringSubmatch(line.String())
            w := &World{regexParams:substrs, MultiStep:line.mldata, output: output} 
            defer func() { line.hasErrors = w.gotAnError }()
            s.call(w)
        }
        return true
    }
    return false
}

func (s stepdef) String() string {
    return s.r.String()
}
