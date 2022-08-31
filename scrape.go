package main

import (
    "fmt"
    "strings"
    "net/http"
    "golang.org/x/net/html"
)


type Test struct {
    input  string
    output string
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

// scrapes sample tests from given contest and problem
func ScrapeTests(contest, problem string) ([]Test, error) {

    url := fmt.Sprintf("https://codeforces.com/problemset/problem/%s/%s", contest, problem)
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    doc, err := html.Parse(resp.Body)
    if err != nil {
        return nil, err
    }

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
    sampleTest, err := dfsNode(doc, func(n *html.Node) bool {
        if n.Type != html.ElementNode { 
            return false 
        }
        for _, attr := range n.Attr {
            if attr.Key == "class" && attr.Val == "sample-test" {
                return true
            }
        }
        return false
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
