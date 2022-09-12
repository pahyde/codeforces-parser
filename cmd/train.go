package cmd

import (
    "fmt"
    "strings"
    "log"
    "net/http"
    "os"
    "time"
    "encoding/json"
    "path/filepath"
    "golang.org/x/net/html"
    "github.com/spf13/cobra"
)

// Types for organizing parsed data from codeforces
// Used to write test cases to disk and generate solution files.
type Contest struct {
    id       string
    problems []Problem
}

type Problem struct {
    id       string
    name     string
    tests    []Test
}

type Test struct {
    input  string
    output string
}


// Types for session data stored in ~/.forces
// Used to store:
//   1) path to test cases and solution files
//   2) contest progress and solution verdicts 
// Persists until the next call to "forces train "

//TODO: start time 

type Session struct {
    Path      string
    Problems  []ProblemState
}

// !ok when problem not found 
func (s Session) getProblemById(id string) (ProblemState, bool) {
    for _, p := range s.Problems {
        if p.ProblemId == id {
            return p, true
        }
    }
    return ProblemState{}, false
}

// returns most recently modified problem from the current session
func (s Session) getLastModifiedProblem() (ProblemState, error) {
    if len(s.Problems) == 0 {
        return ProblemState{}, fmt.Errorf("Can't get current problem. No problems listed for current session.")
    }
    dir := s.Path
    // find problem with the largest unix modification time
    var lastModified ProblemState
    var maxModTime int64
    for _, pstate := range s.Problems {
        path := filepath.Join(dir, pstate.ProblemId)
        info, err := os.Stat(path)
        if err != nil {
            return ProblemState{}, err
        }
        if t := info.ModTime().Unix(); t > maxModTime {
            maxModTime   = t
            lastModified = pstate
        }
    }
    return lastModified, nil
}

// active template and test/submission verdicts for problem w/ id = problemId
// TODO: Template  tname -> Templ Template
// adds redundancy but decouples Session and TemplateRegistry structs
type ProblemState struct {
    ProblemId     string
    Template      tname
    Tests         TestVerdict
    Submission    SubmitVerdict
}

type TestVerdict struct {
    Passed    int // num
    Total     int // den
}

type SubmitVerdict struct {
    Label   SVLabel
    Message string
}

type SVLabel uint8
const (
    NA                      SVLabel = iota
    MemoryLimitExceeded 
    TimeLimitExceeded
    RuntimeError
    WrongAnswer
    IdlenessLimitExceeded
    DenialOfJudgement
    Accepted
)


// Types for template data stored in ~/.config/forces/templates.json
type tname string
type TemplateRegistry struct {
    Starter    tname
    List       []Template
}

type Template struct {
    Name  tname
    Path  string
    Ext   string
    Run   string
}

func (t TemplateRegistry) GetStarter() (Template, bool) {
    for _, templ := range t.List {
        if templ.Name == t.Starter {
            return templ, true
        }
    }
    return Template{}, false
}

// forces train contest
// forces train contest problem
// forces train contest problem --template python
// forces train contest problem -t python
// 1) parse contest problems -> Contest struct
// 2) populate ./{contestid} with a.cpp, b.cpp
//    and ./{contestId}/tests with dirs a,b,c... contianing in0.txt, out0.txt, int1.txt, out1.txt...
// 3) update .forces with:
// {
//   path: ~/phyde/Documents/cp/{contestId}
//   problems: [
//      {name: 'a', testStatus: 0|1|2, submitStatus: 0|1|2},...
//   ]
// }
//   

var trainCmd = &cobra.Command{
    Use: "train",
    Short: "",
    Run: func(cmd *cobra.Command, args []string) {
        //TODO: refactor with cobra arg checks? 
        if len(args) == 0 {
            fmt.Errorf("Must provide a contest id")
        }
        contestId  := args[0]
        problemIds := args[1:]
        if len(problemIds) == 0 {
            // get all problemIds from contestId
            contestUrl := fmt.Sprintf("https://codeforces.com/contest/%s", contestId)
            html, err := getHTMLParseTree(contestUrl)
            if err != nil {
                log.Fatal(err)
            }
            ids, err := parseProblemIds(html)
            if err != nil {
                log.Fatal(err)
            }
            problemIds = ids
        }

        // local directory to store parsed test cases
        //TODO: if duplicate dir, update w/ modifier, i.e. 1130 A -> 1130_0 A
        contestDir := contestId

        contest := Contest{
            contestId, 
            make([]Problem, 0, len(problemIds)),
        }

        for _, id := range problemIds {
            problemUrl := fmt.Sprintf("https://codeforces.com/contest/%s/problem/%s", contestId, id)
            html, err := getHTMLParseTree(problemUrl)
            if err != nil {
                log.Fatal(err)
            }
            // get problem name
            name, err := parseName(html)
            if err != nil {
                log.Fatal(err)
            }
            // get problem sample tests
            tests, err := parseTests(html)
            if err != nil {
                log.Fatal(err)
            }
            problem := Problem{id, name, tests}
            contest.problems = append(contest.problems, problem)
        }

        // For each problem write tests to dir /contestId/tests/problemId/
        for _, problem := range contest.problems {
            // tests directory for problem
            testDir := filepath.Join(contestDir, "tests", problem.id)
            if err := os.MkdirAll(testDir, 0755); err != nil {
                log.Fatal(err)
            }
            for i, test := range problem.tests {
                inputPath  := filepath.Join(testDir, fmt.Sprintf("in%d.txt", i))
                outputPath := filepath.Join(testDir, fmt.Sprintf("out%d.txt", i))
                os.WriteFile(inputPath,  []byte(test.input),  0644)
                os.WriteFile(outputPath, []byte(test.output), 0644)
            }
        }

        // build app directory if it doesn't exist
        //TODO: Possibly wasteful compared to using os.Stat
        configDir, err := os.UserConfigDir()
        appDir := filepath.Join(configDir, "forces")
        if err := os.MkdirAll(appDir, 0700); err != nil {
            log.Fatal(err)
        }

        // read and deserialize TemplateRegistry or create if it doesn't exist
        var registry TemplateRegistry
        p := filepath.Join(appDir, "templates.json")
        registry, err = readTemplateRegistry(p)
        if os.IsNotExist(err) {
            r, err := InitTemplateRegistry(p)
            if err != nil {
                log.Fatal(err)
            }
            registry = r
        }

        // get starter template
        t, ok := registry.GetStarter()
        if !ok {
            log.Fatal("couldn't find starter template in templates list")
        }
        // generate solution from starter template for each problem
        // write to path like contest/A.cpp)
        for _, problem := range contest.problems {
            s, err := generateSolution(t, contest, problem)
            if err != nil {
                log.Fatal(err)
            }
            p := filepath.Join(contestDir, fmt.Sprintf("%s%s", problem.id, t.Ext))
            if err := os.WriteFile(p, s, 0755); err != nil {
                log.Fatal(err)
            }
        }

        // Store session data at os dependent config directory 
        // (e.g. .config/forces for linux).
        // create session struct with path set to curr working directory
        wd, err := os.Getwd()
        if err != nil {
            log.Fatal(err)
        }
        session := Session{Path: filepath.Join(wd, contestDir)}
        // update session with problem templates and initialized verdicts
        for _, problem := range contest.problems {
            template := registry.Starter
            test     := TestVerdict{Passed: 0, Total: len(problem.tests)}
            submit   := SubmitVerdict{}
            state    := ProblemState{problem.id, template, test, submit}
            session.Problems = append(session.Problems, state)
        }

        // JSON encode session data
        dat, err := json.Marshal(&session)
        if err != nil {
            log.Fatal(err)
        }

        // write session data to ...appdir/session.json
        sessionPath := filepath.Join(appDir, "session.json")
        os.WriteFile(sessionPath, dat, 0644)

        fmt.Println(session.getLastModifiedProblem())
    },
}

func init() {
    rootCmd.AddCommand(trainCmd)
}

// returns deserialized TemplateRegistry data read from path p (appDir/templates.cpp)
func readTemplateRegistry(p string) (TemplateRegistry, error) {
    // read
    dat, err := os.ReadFile(p)
    if err != nil {
        return TemplateRegistry{}, err
    }
    // unmarshal
    var r TemplateRegistry
    if err := json.Unmarshal(dat, &r); err != nil {
        return TemplateRegistry{}, err
    }
    return r, nil
}

// returns new TemplateRegistry struct after serializing to path p (appDir/templates.cpp)
// also restores default.cpp template if doesn't exist
func InitTemplateRegistry(p string) (TemplateRegistry, error) {

    cppPath := filepath.Join(filepath.Dir(p), "default.cpp")
    if _, err := os.Stat(cppPath); err != nil {
        // create a new default.cpp template
        if err := InitDefaultTemplate(cppPath); err != nil {
            return TemplateRegistry{}, err
        }
    }

    init := Template{
        Name: "default",
        Path: cppPath,
        Ext: ".cpp",
        Run: "g++ -o sol {{path}}.cpp & ./sol",
    }
    r := TemplateRegistry{Starter: "default", List: []Template{init}}

    dat, err := json.Marshal(&r)
    if err != nil {
        return TemplateRegistry{}, err
    }
    if err := os.WriteFile(p, dat, 0644); err != nil {
        return TemplateRegistry{}, err
    }
    return r, nil
}


func InitDefaultTemplate(p string) error {
    cpp := `
Test CPP file

`
    return os.WriteFile(p, []byte(cpp), 0644)
}


func generateSolution(t Template, c Contest, p Problem) ([]byte, error) {
    // header
    contest := c.id
    name    := p.name
    url     := fmt.Sprintf("https://codeforces.com/contest/%s/problem/%s", c.id, p.id)
    date    := time.Now().String()
    header  := fmt.Sprintf( "// contest: %s\n// problem name: %s\n// url: %s\n// date: %s\n\n", contest, name, url, date)
    // template
    templ, err := os.ReadFile(t.Path)
    if err != nil {
        return nil, err
    }
    return append([]byte(header), templ...), nil
}


// depth-first search for first html node satisfying isMatch function
func dfsNode(n *html.Node, isMatch func(*html.Node) bool) (*html.Node, error) {
    if isMatch(n) {
        // success. 
        return n, nil
    }
    for c := n.FirstChild; c != nil; c = c.NextSibling {
        found, err := dfsNode(c, isMatch)
        // if error, continue search. 
        if err != nil {
            continue
        }
        // bubble up found child node. 
        return found, err
    }
    return nil, fmt.Errorf("No matching node found.")
}

// scrapes text chunks from each child html.TextNode 
// and returns as a single newline delimited string
func scrapeText(n *html.Node) (string, error) {
    if n == nil {
        return "", fmt.Errorf("*html.Node with nil value passed to function scrapeText")
    }

    chunks := make([]string, 0)

    var dfsTextNodes func(n *html.Node)
    dfsTextNodes = func(n *html.Node) {
        if n.Type == html.TextNode {
            chunks = append(chunks, n.Data)
        }
        for c := n.FirstChild; c != nil; c = c.NextSibling {
            dfsTextNodes(c)
        }
    }
    dfsTextNodes(n)

    if len(chunks) == 0 {
        return "", fmt.Errorf("0 text nodes found in *html.Node passed to scrapeText func")
    }
    return strings.Join(chunks, "\n"), nil
}

// returns root node of html parse tree for the given url
func getHTMLParseTree(url string) (*html.Node, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    doc, err := html.Parse(resp.Body)
    if err != nil {
        return nil, err
    }
    return doc, nil
}


// returns true if html node n contains key-value pair k,v
func containsAttr(n *html.Node, k, v string) bool {
    for _, attr := range n.Attr {
        if attr.Key == k && attr.Val == v {
            return true
        }
    }
    return false
}

// parses problem ids from the html parse tree of a CF contest page
// input "contest" is an html root node corresponding to a url of the form:
// https://codeforces.com/contest/{contestId}/
func parseProblemIds(contest *html.Node) ([]string, error) {
    problems, err := dfsNode(contest, func(n *html.Node) bool {
        if n.Type != html.ElementNode {
            return false
        }
        return containsAttr(n, "class", "problems")
    })
    if err != nil {
        return nil, err
    }
    //TODO: figure out why this works but doesn't seem match inspected html 
    tbody := problems.FirstChild.NextSibling
    titleRow := tbody.FirstChild
    firstRow := titleRow.NextSibling.NextSibling

    ids := make([]string, 0)
    for r := firstRow; r != nil; r = r.NextSibling {
        aHref, err := dfsNode(r, func(n *html.Node) bool {
            return n.Data == "a"
        })
        innerText, err := scrapeText(aHref)
        if err != nil {
            log.Fatal(err)
        }
        id := strings.TrimSpace(innerText)
        ids = append(ids, id)
    }
    return ids, nil
}

// parses the name of a codeforces problem from an html parse tree
// input "problem" is an html root node corresponding to a url of the form:
// https://codeforces.com/contest/{contestId}/problem/{problemId}
func parseName(problem *html.Node) (string, error) {
    name, err := dfsNode(problem, func(n *html.Node) bool {
        if n.Type != html.TextNode { 
            return false 
        }
        parent := n.Parent
        if parent == nil {
            return false 
        }
        return containsAttr(parent, "class", "title")
    })
    if err != nil {
        // <div class="title">{some text}</div> node not found
        return "", fmt.Errorf("problem name not found")
    }
    return name.Data, nil
}

// parses the sample tests of a codeforces problem from an html parse tree
// input: "problem" is an html root node corresponding to a url of the form:
// https://codeforces.com/contest/{contestId}/problem/{problemId}
func parseTests(problem *html.Node) ([]Test, error) {
    // dfs for <div class="sample-test">
    // contains input and output for each sample test
    // On success, sampleTest has html structure:
    // <div class="sample-test">
    //     <div class="input">...</div>
    //     <div class="output">...</div>
    //     <div class="input">...</div>
    //     <div class="output">...</div>
    //     ...
    // </div>
    sampleTest, err := dfsNode(problem, func(n *html.Node) bool {
        if n.Type != html.ElementNode { 
            return false 
        }
        return containsAttr(n, "class", "sample-test")
    })
    if err != nil {
        // sample-test node not found
        return nil, fmt.Errorf("<div class=\"sample-test\"><\\div> not found")
    }

    tests := make([]Test, 0)

    c := sampleTest.FirstChild
    for c != nil {
        // <div class="input">
        inputNode  := c
        // <div class="ouput">
        outputNode := c.NextSibling

        if outputNode == nil {
            return nil, fmt.Errorf("missing sample output for input tests")
        }

        // html <pre> tag containing program input
        inputPre := inputNode.LastChild
        // html <pre> tag containing program ouput
        outputPre := outputNode.LastChild
        
        // utf-8 encoded program input
        input, err := scrapeText(inputPre)
        if err != nil {
            return nil, err
        }
        // utf-8 encoded program output
        output, err := scrapeText(outputPre)
        if err != nil {
            return nil, err
        }
        tests = append(tests, Test{input, output})
        c = outputNode.NextSibling
    }
    return tests, nil
}
