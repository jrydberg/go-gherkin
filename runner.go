package gherkin

import (
    re "regexp"
    "strings"
    "fmt"
    "io"
    "io/ioutil"
    "path/filepath"
    "os"
    matchers "github.com/tychofreeman/go-matchers"
)

type Runner struct {
    steps []stepdef
    background Scenario
    isExample bool
    setUp func()
    tearDown func()
    currScenario Scenario
    scenarios []Scenario
    output io.Writer
}

func (r *Runner) addStepLine(line, orig string) {
    r.currScenario.AddStep(StepFromStringAndOrig(line, orig))
}

func (r *Runner) currStepLine() step {
    l := r.currStep()
    if l == nil {
        return StepFromString("")
    }
    return *l
}

// Register a set-up function to be called at the beginning of each scenario
func (r *Runner) SetSetUpFn(setUp func()) {
    r.setUp = setUp
}

// Register a tear-down function to be called at the end of each scenario
func (r *Runner) SetTearDownFn(tearDown func()) {
    r.tearDown = tearDown
}

// The recommended way to create a gherkin.Runner object.
func CreateRunner() *Runner {
    s := []Scenario{}
    return &Runner{[]stepdef{}, nil, false, nil, nil, nil, s, os.Stdout}
}

func createWriterlessRunner() *Runner {
    r := CreateRunner()
    r.output = nil
    return r
}

// Register a step definition. This requires a regular expression
// pattern and a function to execute.
func (r *Runner) RegisterStepDef(pattern string, f interface{}) {
    r.steps = append(r.steps, createstepdef(pattern, f))
}

func (r *Runner) callSetUp() {
    if r.setUp != nil {
        r.setUp()
    }
}

func (r *Runner) callTearDown() {
    if r.tearDown != nil {
        r.tearDown()
    }
}

func (r *Runner) runBackground() {
    if r.background != nil {
        r.background.Execute(r.steps, r.output)
    }
}

func parseAsStep(line string) (bool, string) {
    givenMatch, _ := re.Compile(`^\s*(Given|When|Then|And|But|\*)\s+(.*?)\s*$`)
    if s := givenMatch.FindStringSubmatch(line); s != nil && len(s) > 1 {
        return true, s[2]
    }
    return false, ""
}

func isScenarioOutline(line string) bool {
    return lineMatches(`^\s*Scenario Outline:\s*(.*?)\s*$`, line)
}

func isExampleLine(line string) bool {
    return lineMatches(`^\s*Examples:\s*(.*?)\s*$`, line)
}

func isScenarioLine(line string) (bool) {
    return lineMatches(`^\s*Scenario:\s*(.*?)\s*$`, line)
}

func isFeatureLine(line string) bool {
    return lineMatches(`Feature:\s*(.*?)\s*$`, line)
}
func isBackgroundLine(line string) bool {
    return lineMatches(`^\s*Background:`, line)
}

func lineMatches(spec, line string) bool {
    featureMatch, _ := re.Compile(spec)
    if s := featureMatch.FindStringSubmatch(line); s != nil {
        return true
    }
    return false
}

func parseTableLine(line string) (fields []string) {
    mlMatch, _ := re.Compile(`^\s*\|.*\|\s*$`)
    if mlMatch.MatchString(line) {
        tmpFields := strings.Split(line, "|")
        fields = tmpFields[1:len(tmpFields)-1]
        for i, f := range fields {
            fields[i] = strings.TrimSpace(f)
        }
    }
    return
}

func createTableMap(keys []string, fields []string) (l map[string]string) {
    l = map[string]string{}
    for i, k := range keys {
        l[k] = fields[i]
    }
    return
}

func (r *Runner) resetWithScenario(s Scenario) {
    r.isExample = false
    r.scenarios = append(r.scenarios, s)
    r.currScenario = r.scenarios[len(r.scenarios)-1]
}

func (r *Runner) startScenarioOutline() {
    r.resetWithScenario(&scenario_outline{})
}

func (r *Runner) startBackground(orig string) {
    r.resetWithScenario(&scenario{orig: orig, isBackground: true})
}

func (r *Runner) startScenario(orig string) {
    r.resetWithScenario(&scenario{orig: orig})
}

func (r *Runner) currStep() *step {
    if r.currScenario != nil {
        return r.currScenario.Last()
    }
    return nil
}


func (r *Runner) setMlKeys(data []string) {
    r.currStep().setMlKeys(data)
}

func (r *Runner) addMlStep(data map[string]string) {
    r.currStep().addMlData(data)
}
func (r *Runner) addPrintableLine(line string) {
    r.scenarios = append(r.scenarios, &printable_line{line})
}

func (r *Runner) step(line string) {
    fields := parseTableLine(line)
    isStep, data := parseAsStep(line)
    if r.currScenario != nil && isStep {
        r.addStepLine(data, line)
    } else if isScenarioOutline(line) {
        r.startScenarioOutline()
    } else if isScenarioLine(line) {
        r.startScenario(line)
    } else if isFeatureLine(line) {
        r.addPrintableLine(line)
    } else if isBackgroundLine(line) {
        r.startBackground(line)
        r.background = r.currScenario
    } else if isExampleLine(line) {
        r.addPrintableLine(line)
        r.isExample = true
    } else if r.isExample && len(fields) > 0 {
        r.addPrintableLine(line)
        switch scen := r.currScenario.(type) {
            case *scenario_outline:
                if scen.keys == nil {
                    scen.keys = fields
                } else {
                    newScenario := scen.CreateForExample(createTableMap(scen.keys, fields))
                    r.scenarios = append(r.scenarios, &newScenario)
                }
            default:
        }
    } else if r.currStep() != nil && len(fields) > 0 {
        r.addPrintableLine(line)
        s := *r.currStep()
        if len(s.keys) == 0 {
            r.setMlKeys(fields)
        } else if len(fields) != len(s.keys) {
            panic(fmt.Sprintf("Wrong number of fields in multi-line step [%v] - expected %d fields but found %d", line, len(s.keys), len(fields)))
        } else if len(fields) > 0 {
            l := createTableMap(s.keys, fields)
            r.addMlStep(l)
        }
    } else {
        r.addPrintableLine(line)
    }
}

func (r *Runner) executeScenario(scenario Scenario) Report{
    rpt := Report{}
    if !scenario.IsBackground() {
        if !scenario.IsJustPrintable() {
            r.callSetUp()
            r.runBackground()
        }
        rpt = scenario.Execute(r.steps, r.output)
        if !scenario.IsJustPrintable() {
            r.callTearDown()
        }
    }
    return rpt
}

func (r *Runner) executeScenarios(scenarios []Scenario) Report {
    rpt := Report{}
    for _, scenario := range scenarios {
        scenarioRpt := r.executeScenario(scenario)
        rpt.scenarioCount++
        rpt.skippedSteps += scenarioRpt.skippedSteps
        rpt.pendingSteps += scenarioRpt.pendingSteps
        rpt.passedSteps += scenarioRpt.passedSteps
        rpt.failedSteps += scenarioRpt.failedSteps
        rpt.undefinedSteps += scenarioRpt.undefinedSteps
    }
    return rpt
}

// Once the step definitions are Register()'d, use Execute() to
// parse and execute Gherkin data.
func (r *Runner) Execute(file string) Report {
    lines := strings.Split(file, "\n")
    for _, line := range lines {
        r.step(line)
    }
    return r.executeScenarios(r.scenarios)
}

func generateStepReport(count int, name string) string {
    if count > 0 {
        return fmt.Sprintf("%d %s", count, name)
    }
    return ""
}

func addCount(stepSpecifics []string, count int, name string) []string {
    tmp := generateStepReport(count, name)
    if len(tmp) > 0 {
        stepSpecifics = append(stepSpecifics, tmp)
    }
    return stepSpecifics
}

func PrintReport(rpt Report, output io.Writer) {
    stepSpecifics := []string{}
    stepSpecifics = addCount(stepSpecifics, rpt.skippedSteps, "skipped")
    stepSpecifics = addCount(stepSpecifics, rpt.passedSteps, "passed")
    stepSpecifics = addCount(stepSpecifics, rpt.failedSteps, "failed")
    stepSpecifics = addCount(stepSpecifics, rpt.pendingSteps, "pending")
    stepSpecifics = addCount(stepSpecifics, rpt.undefinedSteps, "undefined")
    subset := strings.Join(stepSpecifics, ", ")
    if len(subset) > 0 {
        subset = "(" + subset + ")"
    }

    totalSteps := rpt.skippedSteps + rpt.passedSteps + rpt.failedSteps + rpt.pendingSteps + rpt.undefinedSteps
    fmt.Fprintf(output, "%d scenarios\n%d steps%s\n", rpt.scenarioCount, totalSteps, subset)
}

func (r *Runner) RunFeature(t matchers.Errorable, filename string) {
    file, err := os.Open(filename)
    if err != nil {
        t.Errorf(err.Error())
    }
    data, _ := ioutil.ReadAll(file)
    rpt := r.Execute(string(data))
    PrintReport(rpt, r.output)
    if rpt.failedSteps > 0 {
        t.Errorf("Failed %s", file)
    }
}

// Once the step definitions are Register()'d, use Run() to
// locate all *.feature files within the feature/ subdirectory
// of the current directory.
func (r *Runner) Run(t matchers.Errorable) {
    featureMatch, _ := re.Compile(`.*\.feature`)
    filepath.Walk("features", func(walkPath string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if info.Name() != "features" && info.IsDir() {
            return filepath.SkipDir
        } else if !info.IsDir() && featureMatch.MatchString(info.Name()) {
            file, _ := os.Open(walkPath)
            data, _ := ioutil.ReadAll(file)
            rpt := r.Execute(string(data))
            PrintReport(rpt, r.output)
            r.scenarios = []Scenario{}
            r.background = nil
            if rpt.failedSteps > 0 {
                t.Errorf("Failed %s", file)
            }
        }
        return nil
    })
}

// By default, Runner uses os.Stdout to write to. However, it may be useful
// to redirect. To do so, provide an io.Writer here.
func (r *Runner) SetOutput(w io.Writer) {
    r.output = w
}
