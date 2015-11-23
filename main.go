package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
)

func red(v string) string {
	return "\033[31m" + v + "\033[0m"
}
func blue(v string) string {
	return "\033[34m" + v + "\033[0m"
}
func grey(v string) string {
	return "\033[37m" + v + "\033[0m"
}

//TaskResult Bla bla
type TaskResult string

//Task Bla bla
type Task func(c <-chan TaskResult) <-chan TaskResult

//TaskRunner Bla bla
type TaskRunner interface {
	Task(name string, task Task) *Runner
	Run(tasks []string) (TaskResult, error)
}

//Runner a
type Runner struct {
	name  string
	tasks map[string]Task
}

func (r *Runner) reverse(s []string) []string {
	//r := []rune(s)
	for i, j := 0, len(s)-1; i < len(s)/2; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func (r *Runner) Write(p []byte) (n int, err error) {
	return fmt.Fprintf(os.Stdout, "%s > %s\n", blue(r.name), grey(string(p)))
}

//Task Add a new task to the runner
func (r *Runner) Task(name string, task Task) *Runner {
	r.tasks[name] = task // = append(r.tasks, task)
	return r
}

//Run run tasks matching the given args
func (r *Runner) Run(tasks []string) (TaskResult, error) {
	//in := make(chan TaskResult, 0)
	ttr := make([]Task, len(tasks))
	i := 0
	for _, taskName := range r.reverse(tasks) {
		t, exists := r.tasks[taskName]
		if !exists {
			fmt.Fprintf(r, red("Task %s does not exists"), taskName)
			return "", fmt.Errorf("Task %s does not exists", taskName)

		}
		ttr[i] = t
		i++
	}
	if len(ttr) == 0 {
		return "", errors.New("No tasks to run")
	}

	in := make(chan TaskResult, 0)
	task := Combine(ttr...)
	out := task(in)
	close(in)
	return <-out, nil
}

//Combine combine multiple tasks into one
func Combine(tasks ...Task) Task {
	return func(in <-chan TaskResult) <-chan TaskResult {
		out := make(chan TaskResult, len(tasks))
		go func() {
			var res = in
			for _, task := range tasks {
				res = task(res)
			}
			out <- TaskResult(<-res)
			close(out)
		}()
		return out
	}
}
func addString(str string) Task {
	return func(in <-chan TaskResult) <-chan TaskResult {
		out := make(chan TaskResult, 0)
		go func() {
			r := <-in
			out <- r + TaskResult(str)
			close(out)
		}()
		return out
	}
}

//DownloadFile put the response of the request on the TaskResult
func DownloadFile(url string) Task {
	return func(in <-chan TaskResult) <-chan TaskResult {
		out := make(chan TaskResult, 0)
		go func() {
			r := <-in
			res, err := http.Get(url)
			if err != nil {
				log.Fatal(err)
			}
			robots, err := ioutil.ReadAll(res.Body)
			res.Body.Close()
			if err != nil {
				log.Fatal(err)
			}
			out <- r + TaskResult(robots)
			close(out)
		}()
		return out
	}
}

//Exec execute a command and put the result as TaskResult
func Exec(command string) Task {
	return func(in <-chan TaskResult) <-chan TaskResult {
		out := make(chan TaskResult, 0)
		go func() {
			r := <-in
			cmd := exec.Command(command)
			res, err := cmd.Output()
			if err != nil {
				panic(err)
			}

			out <- r + TaskResult(res)
			close(out)
		}()
		return out
	}

}

func main() {
	tasksToRun := os.Args[1:]

	runner := &Runner{
		name:  "Runner",
		tasks: make(map[string]Task),
	}

	runner.Task("ls", Exec("ls"))
	runner.Task("GR", DownloadFile("http://www.google.com/robots.txt"))
	runner.Task("GH", DownloadFile("http://www.google.com/humans.txt"))

	runner.Task("AddHW", Combine(DownloadFile("http://www.google.com/robots.txt"), addString("Hello"), addString(" "), addString("World"), addString("!"), DownloadFile("http://www.google.com/humans.txt")))
	runner.Task("AddHello", addString("Hello"))
	runner.Task("AddSpace", addString(" "))
	runner.Task("AddWorld", addString("World"))
	runner.Task("Add!", addString("!"))

	res, err := runner.Run(tasksToRun)
	if err != nil {
		fmt.Fprint(runner, red(err.Error()))
	} else {
		fmt.Fprint(runner, res)
	}
}
