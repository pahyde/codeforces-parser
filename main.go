package main

import (
    "fmt"
    "log"
    "net/http"
    "golang.org/x/net/html"
)


type Test struct {
    input  string
    output string
}



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



func parseLines(n *html.Node) string {
    //TODO: returns formated utf-8 encoded text nested in html Node 'n'
    return ""
}


func parseTests(doc *html.Node) ([]Test, error) {

    isSampleTestNode := func(n *html.Node) bool {
        if n.Type != html.ElementNode { return false }
        for _, attr := range n.Attr {
            if attr.Key == "class" && attr.Val == "sample-test" {
                return true
            }
        }
        return false
    }
    sampleTest, err := dfsNode(doc, isSampleTestNode)
    if err != nil {
        return nil, err
    }

    tests := make([]Test, 0)

    c := sampleTest.FirstChild

    for c != nil {

        inputNode  := c
        outputNode := c.NextSibling
        if outputNode == nil {
            return nil, fmt.Errorf("missing sample output for input tests")
        }
        
        input, err := parseLines(input)
        if err != nil {
            return nil, err
        }

        output, err := parseLines(output)
        if err != nil {
            return nil, err
        }

        tests = append(tests, Test{input, output})

        c = output.NextSibling
    }
    return tests, nil
}


func main() {
    resp, err := http.Get("https://codeforces.com/problemset/problem/1720/D1")
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()
    
    doc, err := html.Parse(resp.Body)
    if err != nil {
        log.Fatal(err)
    }

    tests, err := parseTests(doc)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(tests)
}



/*
func dfsnodes(n *html.node, data string, nodetype int) ([]*html.node, error) {

    nodes := make([]*html.node, 0)

    dfs := func(n *html.node, data, string, nodetype int) {
        if n.type == nodetype && n.data == data {
            nodes = append(nodes, n)
        }
        for c := n.firstchild; c != nil; c = c.nextsibling {
            dfs(c)
        }
    }
    dfs(n)

    if len(nodes) == 0 {
        return nodes, fmt.errorf(
            "no text nodes with data: %s and type: %d found.", 
            data,
            nodetype,
        )
    }
    return nodes, nil
}
*/
