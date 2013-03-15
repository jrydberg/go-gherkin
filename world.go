package gherkin

import (
    "fmt"
    "io"
)

// Passed to each step-definition
type World struct {
    regexParams []string
    regexParamIndex int
    MultiStep []map[string]string
    output io.Writer
    gotAnError bool
}

// Allows World to be used with the go-matchers AssertThat() function.
func (w *World) Errorf(format string, args ...interface{}) {
    w.gotAnError = true
    if w.output != nil {
        if len(args) == 0 {
            fmt.Fprintf(w.output, format)
        } else {
            fmt.Fprintf(w.output, format, args)
        }
    }
}
