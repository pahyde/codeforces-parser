package main

import (
    "fmt"
    "log"
)

func main() {
    tests, err := ScrapeTests("1720", "D1")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(tests[0].input)
    fmt.Println(tests[0].output)
}

