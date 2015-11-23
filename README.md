# Task runner in GOLANG

Example:

```go
tasksToRun := os.Args[1:]

runner := &Runner{
  name:  "Runner",
  tasks: make(map[string]Task),
}

runner.Task("ls", Exec("ls"))
runner.Task("GH", DownloadFile("http://www.google.com/humans.txt"))

runner.Task("AddHello", addString("Hello"))
runner.Task("AddSpace", addString(" "))
runner.Task("AddWorld", addString("World"))
runner.Task("Add!", addString("!"))
runner.Task("HelloWorld", Combine(addString("Hello"), addString(" "), addString("World"), addString("!")))

res, err := runner.Run(tasksToRun)
if err != nil {
  fmt.Fprint(runner, err.Error())
} else {
  fmt.Fprint(runner, res)
}
```
