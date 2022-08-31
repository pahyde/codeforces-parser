package main

import (
    "fmt"
    "log"
    "os"
//    "os/exec"
)

func main() {

    contest := os.Args[1]
    problem := os.Args[2]

    tests, err := ScrapeTests(contest, problem)
    if err != nil {
        log.Fatal(err)
    }

    for _, test := range tests {
        fmt.Println(test.input)
        fmt.Println(test.output)
    }

/*
    cmd := exec.Command("vim", "test.txt")
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    err = cmd.Run()
    fmt.Println(err) 
*/

}
