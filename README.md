# GO Pod Executor
### A library that helps to execute commands in kubernetes pods

### Usage
Execute method manages a blocking call to the container and returns a result with the output of the command
```go
executor, err := podexecutor.NewCommandExecutor(masterURL, kubeConfig)
if err != nil {
    panic(err)
}

result, err := executor.Execute(context.TODO(), &podexecutor.Request{
    Pod:       "nginx",
    Namespace: "executor",
    Command:   []string{"ls", "-a"},
})
if err != nil {
    panic(err)
}

// a raw list of folders as a single string
// requires some further processing and formatting
outputAsStr := result.Output()
```