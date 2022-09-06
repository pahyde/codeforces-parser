package cmd

import (
    "fmt"
    "strings"
    "log"
    "net/http"
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
    path      string
    verdicts  []Verdict
}

// Test and Submission verdicts for problem w/ id = problemId
type Verdict struct {
    problemId     string
    tests         TestVerdict
    submission    SubmitVerdict
}

// Test verdict
type TestVerdict struct {
    passed    int // num
    total     int // den
}

// Submit verdict
type SubmitVerdict struct {
    label   SVLabel
    message string
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

// forces train contest
// forces train contest problem
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
        //dir := contestId

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

        for _, problem := range contest.problems {
            fmt.Printf("tests problem %s: %s\n", problem.id, problem.name)
            for _, test := range problem.tests {
                fmt.Println(test.input)
                fmt.Println(test.output)
            }
        }
    },
}

func init() {
    rootCmd.AddCommand(trainCmd)
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
