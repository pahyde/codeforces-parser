package main

import (
    "github.com/pahyde/forces/cmd"
)

func main() {
    cmd.Execute()
/*
    cmd := exec.Command("vim", "test.txt")
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    err = cmd.Run()
    fmt.Println(err) 
*/
}
