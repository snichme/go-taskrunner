package gotrun

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
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
type TaskResult []byte

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

//NewRunner Create a new runner
func NewRunner(name string) *Runner {
	return &Runner{
		name:  name,
		tasks: make(map[string]Task),
	}
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
	for _, taskName := range tasks {
		t, exists := r.tasks[taskName]
		if !exists {
			fmt.Fprintf(r, red("Task %s does not exists"), taskName)
			return nil, fmt.Errorf("Task %s does not exists", taskName)
		}
		ttr[i] = t
		i++
	}
	if len(ttr) == 0 {
		return nil, errors.New("No tasks to run")
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

//DownloadFile put the response of the request on the stream
func DownloadFile(url string) Task {
	return func(in <-chan TaskResult) <-chan TaskResult {
		out := make(chan TaskResult, 0)
		go func() {
			<-in
			res, err := http.Get(url)
			if err != nil {
				log.Fatal(err)
			}
			robots, err := ioutil.ReadAll(res.Body)
			res.Body.Close()
			if err != nil {
				log.Fatal(err)
			}
			out <- TaskResult(robots)
			close(out)
		}()
		return out
	}
}

//Exec execute a command and put the result on the stream
func Exec(command string) Task {
	return func(in <-chan TaskResult) <-chan TaskResult {
		out := make(chan TaskResult, 0)
		go func() {
			<-in
			cmd := exec.Command(command)
			res, err := cmd.Output()
			if err != nil {
				panic(err)
			}

			out <- TaskResult(res)
			close(out)
		}()
		return out
	}
}

//Printer prints the current data on the stream to stdout
func Printer() Task {
	return func(in <-chan TaskResult) <-chan TaskResult {
		out := make(chan TaskResult, 0)
		go func() {
			r := <-in
			fmt.Printf("%s [%s] %s", time.Now(), "Printer", r)
			out <- r
			close(out)
		}()
		return out
	}
}

//TaskError Print a task error
func TaskError(name string, err error) {
	log.Fatalf("[%s] %s", name, err.Error())
}
