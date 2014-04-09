package gherkin

import (
    "testing"
    . "github.com/tychofreeman/go-matchers"
    "regexp"
)

func TestReportsNumberOfPendingSteps(t *testing.T) {
    scen := &scenario{}
    scen.AddStep(step{line:".", isPending:true})
    regex, _ := regexp.Compile(".")
    sd := stepdef{r:regex, f:func(w *World, ctx *Context){ }}
    rpt := scen.Execute([]stepdef{sd}, nil, &Context{})

    AssertThat(t, rpt.pendingSteps, Equals(1))
}

func TestReportsNumberOfSkippedSteps(t *testing.T) {
    scen := &scenario{}
    scen.AddStep(step{line:".", isPending:true})
    scen.AddStep(step{line:".", isPending:true})
    regex, _ := regexp.Compile(".")
    sd := stepdef{r:regex, f:func(w *World, ctx *Context){ }}
    rpt := scen.Execute([]stepdef{sd}, nil, &Context{})

    AssertThat(t, rpt.skippedSteps, Equals(1))
}

func TestReportsNumberOfPassedSteps(t *testing.T) {
    scen := &scenario{}
    scen.AddStep(step{line:"."})
    regex, _ := regexp.Compile(".")
    sd := stepdef{r:regex, f:func(w *World, ctx *Context){ }}
    rpt := scen.Execute([]stepdef{sd}, nil, &Context{})

    AssertThat(t, rpt.passedSteps, Equals(1))
}

func TestReportsNumberOfFailedSteps(t *testing.T) {
    scen := &scenario{}
    scen.AddStep(step{line:"."})
    regex, _ := regexp.Compile(".")
    sd := stepdef{r:regex, f:func(w *World, ctx *Context){ AssertThat(w, true, IsFalse) }}
    rpt := scen.Execute([]stepdef{sd}, nil, &Context{})

    AssertThat(t, rpt.failedSteps, Equals(1))
}

func TestReportsNumberOfUndefinedSteps(t *testing.T) {
    scen := &scenario{}
    scen.AddStep(step{line:"."})
    rpt := scen.Execute([]stepdef{}, nil, &Context{})

    AssertThat(t, rpt.undefinedSteps, Equals(1))
}
