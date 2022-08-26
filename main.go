package main

import (
    "fmt"
    "log"
    "net/http"
    "golang.org/x/net/html"
)


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

func parseTests(doc *html.Node) (*html.Node, error) {

    isSampleTest := func(n *html.Node) bool {
        if n.Type != html.ElementNode { return false }
        for _, attr := range n.Attr {
            if attr.Key == "class" && attr.Val == "sample-test" {
                return true
            }
        }
        return false
    }

    sampleTest, err := dfsNode(doc, isSampleTest)
    if err != nil {
        return nil, err
    }

    c := sampleTest.FirstChild

    for c != nil {
        input  := c
        output := c.NextSibling
        if output == nil {
            return nil, fmt.Errorf("missing sample output for input tests")
        }

        fmt.Println(input.Attr)
        fmt.Println(output.Attr)

        c = output.NextSibling
    }
    return sampleTest, nil
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

    testCases, err := parseTests(doc)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(testCases.Data)
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
