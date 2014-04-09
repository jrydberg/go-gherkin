package gherkin

import (
    "testing"
    . "github.com/tychofreeman/go-matchers"
)

type Context struct {
    wasCalled bool
    firstWasCalled bool
    secondWasCalled bool
    actionWasCalled bool
    secondActionCalled bool
    wasGivenRun bool
    wasThenRun bool
    givenData []map[string]string
    thenData []map[string]string
    whenData []map[string]string
    captured string
    timesRun int
    wasRun bool
    setUpWasCalled bool
    tearDownWasCalled bool
    setUpCalledBeforeStep bool
}

var featureText = `Feature: My Feature
    Scenario: Scenario 1
        Given the first setup
        When the first action
        Then the first result
        But not the other first result
    Scenario: Scenario 2
        Given the second setup
        When the second action
        Then the second result
        And the other second result
    Scenario: Scenario 3
        * the third setup
        When     the third action has leading spaces
        When the third action has trailing spaces
    This is ignored`

func assertMatchCalledOrNot(t *testing.T, step string, pattern string, isCalled bool) {
    f := func(w *World, ctx *Context) {
        ctx.wasCalled = true
    }

    g := createWriterlessRunner()
    g.RegisterStepDef(pattern, f)

    ctx := &Context{}
    g.Execute(step, ctx)
    AssertThat(t, ctx.wasCalled, Equals(isCalled))
}

func matchingFunctionIsCalled(t *testing.T, step string, pattern string) {
    assertMatchCalledOrNot(t, step, pattern, true)
}

func matchingFunctionIsNotCalled(t *testing.T, step string, pattern string) {
    assertMatchCalledOrNot(t, step, pattern, false)
}

func TestExecutesMatchingMethod(t *testing.T) {
    matchingFunctionIsCalled(t, featureText, ".")
}

func TestAvoidsNonMatchingMethod(t *testing.T) {
    matchingFunctionIsNotCalled(t, featureText, "^A")
}

func TestCallsOnlyFirstMatchingMethod(t *testing.T) {
    first := func(w *World, ctx *Context) { }
    second := func(w *World, ctx *Context) {
        ctx.wasCalled = true
    }

    c := &Context{}
    g := createWriterlessRunner()
    g.RegisterStepDef(".", first)
    g.RegisterStepDef(".", second)
    g.Execute("Given only the first step is called", c)
    AssertThat(t, c.wasCalled, Equals(false))
}

func TestRemovesGivenFromMatchLine(t *testing.T) {
    matchingFunctionIsCalled(t, featureText, "^the first setup$")
}

func TestRemovesWhenFromMatchLine(t *testing.T) {
    matchingFunctionIsCalled(t, featureText, "^the first action$")
}

func TestRemovesThenFromMatchLine(t *testing.T) {
    matchingFunctionIsCalled(t, featureText, "^the first result$")
}

func TestRemovesAndFromMatchLine(t *testing.T) {
    matchingFunctionIsCalled(t, featureText, "^the other second result$")
}

func TestRemovesButFromMatchLine(t *testing.T) {
    matchingFunctionIsCalled(t, featureText, "^not the other first result$")
}

func TestRemovesStarFromMatchLine(t *testing.T) {
    matchingFunctionIsCalled(t, featureText, "^the third setup$")
}

func TestRemovesLeadingSpacesFromMatchLine(t *testing.T) {
    matchingFunctionIsCalled(t, featureText, "^the third action has leading spaces$")
}

func TestRemovesTrailingSpacesFromMatchLine(t *testing.T) {
    matchingFunctionIsCalled(t, featureText, "^the third action has trailing spaces$")
}

func TestMultipleStepsAreCalled(t *testing.T) {
    g := createWriterlessRunner()

    g.RegisterStepDef("^the first setup$", func(w *World, ctx *Context) {
        ctx.firstWasCalled = true
    })

    g.RegisterStepDef("^the first action$", func(w *World, ctx *Context) {
        ctx.secondWasCalled = true
    })

    c := &Context{}
    g.Execute(featureText, c)
    AssertThat(t, c.firstWasCalled, IsTrue)
    AssertThat(t, c.secondWasCalled, IsTrue)
}

func TestPendingSkipsTests(t *testing.T) {
    g := createWriterlessRunner()

    g.RegisterStepDef("^the first setup$", func(w *World, ctx *Context) { Pending() })
    g.RegisterStepDef("^the first action$", func(w *World, ctx *Context) { ctx.actionWasCalled = true })

    c := &Context{}
    g.Execute(featureText, c)
    AssertThat(t, c.actionWasCalled, IsFalse)
}

func TestPendingDoesntSkipSecondScenario(t *testing.T) {
    g := createWriterlessRunner()

    g.RegisterStepDef("^the first setup$", func(w *World, ctx *Context) { Pending() })
    g.RegisterStepDef("^the second setup$", func(w *World, ctx *Context) { } )
    g.RegisterStepDef("^the second action$", func(w *World, ctx *Context) { ctx.secondActionCalled = true })

    c := &Context{}
    g.Execute(featureText, c)
    AssertThat(t, c.secondActionCalled, Equals(true))
}

func TestBackgroundIsRunBeforeEachScenario(t *testing.T) {
    c := &Context{}
    g := createWriterlessRunner()
    g.RegisterStepDef("^background$", func(w *World, ctx *Context) { ctx.wasCalled = true })
    g.Execute(`Feature:
        Background:
            Given background
        Scenario:
            Then this
    `, c)

    AssertThat(t, c.wasCalled, IsTrue)
}

func TestCallsSeUptBeforeScenario(t *testing.T) {
    c := &Context{}
    g := createWriterlessRunner()
    g.SetSetUpFn(func(ctx *Context) { ctx.setUpWasCalled = true })

    g.RegisterStepDef(".", func(w *World, ctx *Context) { ctx.setUpCalledBeforeStep = ctx.setUpWasCalled })
    g.Execute(`Feature:
        Scenario:
            Then this`, c)

    AssertThat(t, c.setUpCalledBeforeStep, IsTrue)
}

func TestCallsTearDownBeforeScenario(t *testing.T) {
    c := &Context{}
    g := createWriterlessRunner()
    g.SetTearDownFn(func(ctx *Context) { ctx.tearDownWasCalled = true })

    g.Execute(`Feature:
        Scenario:
            Then this`, c)

    AssertThat(t, c.tearDownWasCalled, IsTrue)
}

func TestPassesTableListToMultiLineStep(t *testing.T) {
    c := &Context{}
    g := createWriterlessRunner()
    g.RegisterStepDef(".", func(w *World, ctx *Context) { ctx.thenData = w.MultiStep })
    g.Execute(`Feature:
        Scenario:
            Then you should see these people
                |name|email|
                |Bob |bob@bob.com|
    `, c)

    expectedData := []map[string]string{
        map[string]string{"name":"Bob", "email":"bob@bob.com"},
    }
    AssertThat(t, c.thenData, Equals(expectedData))
}

func TestErrorsIfTooFewFieldsInMultiLineStep(t *testing.T) {
    c := &Context{}
    g := createWriterlessRunner()
    // Assertions before end of test...
    defer func() {
        recover()
        AssertThat(t, c.wasGivenRun, IsFalse)
        AssertThat(t, c.wasThenRun, IsFalse)
    }()

    g.RegisterStepDef("given", func(w *World, ctx *Context) { ctx.wasGivenRun = true })
    g.RegisterStepDef("then", func(w *World, ctx *Context) { ctx.wasThenRun = true })

    g.Execute(`Feature:
        Scenario:
            Given given
                |name|addr|
                |bob|
            Then then`, c)
}

func TestSupportsMultipleMultiLineStepsPerScenario(t *testing.T) {
    c := &Context{}
    g := createWriterlessRunner()
    g.RegisterStepDef("given", func(w *World, ctx *Context) { ctx.givenData = w.MultiStep })
    g.RegisterStepDef("when", func(w *World, ctx *Context) { ctx.whenData = w.MultiStep })

    g.Execute(`Feature:
        Scenario:
            Given given
                |name|email|
                |Bob|bob@bob.com|
                |Jim|jim@jim.com|
            When when
                |breed|height|
                |wolf|2|
                |shihtzu|.5|
    `, c)

    expectedGivenData := []map[string]string{
        map[string]string{ "name":"Bob", "email":"bob@bob.com"},
        map[string]string{ "name":"Jim", "email":"jim@jim.com"},
    }

    expectedWhenData := []map[string]string{
        map[string]string{"breed":"wolf", "height":"2"},
        map[string]string{"breed":"shihtzu", "height":".5"},
    }

    AssertThat(t, c.givenData, Equals(expectedGivenData))
    AssertThat(t, c.whenData, Equals(expectedWhenData))
}

func TestAllowsAccessToFirstRegexCapture(t *testing.T) {
    c := &Context{}
    g := createWriterlessRunner()
    g.RegisterStepDef("(thing)", func(w *World, ctx *Context, thing string) {
        ctx.captured = thing
    })
    g.Execute(`Feature:
        Scenario:
            Given thing
    `, c)

    AssertThat(t, c.captured, Equals("thing"))
}

func TestFailsGracefullyWithOutOfBoundsRegexCaptures(t *testing.T) {
    panicked := false
    g := createWriterlessRunner()
    g.RegisterStepDef(".", func(w *World, ctx *Context, x string) { })

    func() {
        defer func() {
            r := recover()
            AssertThat(t, r, Equals("Function type mismatch"))
            panicked = true
        }()

        g.Execute(`Feature:
            Scenario:
                Given .
        `, &Context{})
    }()

    AssertThat(t, panicked, IsTrue)
}

func TestFailsGracefullyWithInvalidFunctionType(t *testing.T) {
    panicked := false
    g := createWriterlessRunner()
    g.RegisterStepDef("(.)", func(w *World, ctx *Context, x interface{}) { })

    func() {
        defer func() {
            r := recover()
            AssertThat(t, r, Equals("Function type not supported"))
            panicked = true
        }()

        g.Execute(`Feature:
            Scenario:
                Given .
        `, &Context{})
    }()

    AssertThat(t, panicked, IsTrue)
}

func TestFailsGracefullyWithInvalidArguments(t *testing.T) {
    panicked := false
    g := createWriterlessRunner()
    g.RegisterStepDef("(.)", func(w *World, ctx *Context, x int) {
        t.Fail()
    })

    func() {
        defer func() {
            recover()
            panicked = true
        }()

        g.Execute(`Feature:
            Scenario:
                Given x
        `, &Context{})
    }()

    AssertThat(t, panicked, IsTrue)
}

func TestSupportsArguments(t *testing.T) {
    g := createWriterlessRunner()
    g.RegisterStepDef("(.*),(.*),(.*),(.*),(.*),(.*),(.*),(.*)",
        func(w *World, ctx *Context,
            b1 bool, i8 int8, i16 int16, i32 int32, i64 int64, i int,
            f32 float32, f64 float64) {
    })

    g.Execute(`Feature:
        Scenario:
            Given true,127,255,255,255,255,0.3,0.4
    `, &Context{})
}


func DISABLED_TestOnlyExecutesStepsBelowScenarioLine(t *testing.T) {
    c := &Context{}
    g := createWriterlessRunner()
    g.RegisterStepDef(".", func(w *World, ctx *Context) { ctx.wasRun = true })
    g.Execute(`Feature:
        Given .`, c)

    AssertThat(t, c.wasRun, IsFalse)
}

func TestScenarioOutlineWithoutExampleDoesNotExecute(t *testing.T) {
    c := &Context{}
    g := createWriterlessRunner()
    g.RegisterStepDef(".", func(w *World, ctx *Context) { ctx.wasRun = true})
    g.Execute(`Feature:
        Scenario Outline:
            Given .
    `, c)

    AssertThat(t, c.wasRun, IsFalse)
}

func TestScenarioOutlineReplacesFieldWithValueInExample(t *testing.T) {
    so := ScenarioOutline()
    so.AddStep(StepFromString(`<count> pops`))
    scenario := so.CreateForExample(map[string]string{"count":"5"})

    AssertThat(t, scenario.steps[0].line, Equals(`5 pops`))
}

func TestScenarioOutlineReplacesManyFieldsWithValuesInExample(t *testing.T) {
    so := ScenarioOutline()
    so.AddStep(StepFromString(`<count> <name>`))
    scenario := so.CreateForExample(map[string]string{"count":"5", "name":"pops"})

    AssertThat(t, scenario.steps[0].line, Equals(`5 pops`))
}

func TestScenarioOutlineSupportsMultipleLines(t *testing.T) {
    so := ScenarioOutline()
    so.AddStep(StepFromString(`<count> <name>`))
    so.AddStep(StepFromString(`<name> <type>`))
    scenario := so.CreateForExample(map[string]string{"count":"5", "name":"pops", "type":"music"})

    AssertThat(t, scenario.steps[0].line, Equals(`5 pops`))
    AssertThat(t, scenario.steps[1].line, Equals(`pops music`))
}

func TestExecutesScenarioOncePerLineInExample(t *testing.T) {
    c := &Context{}
    g := createWriterlessRunner()
    g.RegisterStepDef(".", func(w *World, ctx *Context) { ctx.timesRun++ })
    g.Execute(`Feature:
        Scenario Outline:
            Given .
        Examples:
            |scenario num|
            |first|
            |second|
    `, c)

    AssertThat(t, c.timesRun, Equals(2))
}

func TestBackgroundDoesntExecuteBackgroundWhenRun(t *testing.T) {
    c := &Context{}
    g := createWriterlessRunner()
    g.RegisterStepDef(".", func(w *World, ctx *Context) { ctx.wasRun = true })
    g.Execute(`Feature:
        Background:
            Given .
   `, c)

   AssertThat(t, c.wasRun, IsFalse)
}

// Support PyStrings?
// Support tags?
// Support reporting.
