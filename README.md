# Task runner in GOLANG

Example:

* Copy the contents below and save in a file called my-tasks.go
* Open your terminal and run go run my-tasks.go SR2

```go
package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/snichme/gotrun"
)

func addString(str string) gotrun.Task {
	return func(in <-chan gotrun.TaskResult) <-chan gotrun.TaskResult {
		out := make(chan gotrun.TaskResult, 0)
		go func() {
			r := <-in
			res := gotrun.TaskResult(str)
			out <- append(r, res...)
			close(out)
		}()
		return out
	}
}

func writeTo(writer io.Writer) gotrun.Task {
	return func(in <-chan gotrun.TaskResult) <-chan gotrun.TaskResult {
		out := make(chan gotrun.TaskResult, 0)
		go func() {
			r := <-in
			_, err := writer.Write(r)
			if err != nil {
				gotrun.TaskError("WriteTo", err)
			}
			close(out)
		}()
		return out
	}
}

func writeToFile(filename string) gotrun.Task {
	f, err := os.Create(filename)
	if err != nil {
		gotrun.TaskError("WriteToFile", err)
	}
	return writeTo(f)
}

func main() {
	tasksToRun := os.Args[1:]
	runner := gotrun.NewRunner("MyTestRunner")
	runner.Task("SR2", gotrun.Combine(gotrun.DownloadFile("http://www.google.com/robots.txt"), gotrun.Printer(), writeToFile("f.txt")))
	runner.Task("SR", gotrun.Combine(gotrun.DownloadFile("http://www.google.com/robots.txt"), writeTo(os.Stdout)))
	runner.Task("ls", gotrun.Exec("ls"))
	runner.Task("uptime", gotrun.Exec("uptime"))
	runner.Task("GR", gotrun.DownloadFile("http://www.google.com/robots.txt"))
	runner.Task("GH", gotrun.DownloadFile("http://www.google.com/humans.txt"))

	res, err := runner.Run(tasksToRun)
	if err != nil {
		log.Fatal(err.Error())
	} else {
		fmt.Printf("%s", res)
	}
}
```
