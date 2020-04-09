package main

import (
  "strings"
  "os/exec"
  "log"
  "fmt"
)

type fileContributer struct {
  Name string
  Email string
  // commits []string
  // commitCount int
}
type fileContribution struct {
  Name string
  Email string
  Commit string
}

func _prototype_getGitLog(path string) fileContributer {
  args :=  []string{"log", "--format=%an|%ae|%f", path}
  cmd := exec.Command("git", args...)
  out, errCmd := cmd.CombinedOutput()
  if errCmd != nil {
    log.Fatalf("cmd.Run() failed with %s\n", errCmd)
  }
  outArray := strings.Split(string(out), "|")
  c := fileContributer{}
  if (len(outArray) > 1) {
    c = fileContributer{
      Name: outArray[0],
      Email: outArray[1],
    }
    fmt.Println(path)
    fmt.Println(c)
  }
  return c
}

// getGitLog returns the list of contributers of a given path
// TODO: return array of fileContributer instead
func getGitLog(path string) []fileContribution {
  args :=  []string{"log", "--format=%an||%ae||%f", path}
  cmd := exec.Command("git", args...)
  out, errCmd := cmd.CombinedOutput()
  if errCmd != nil {
    log.Fatalf("cmd.Run() failed with %s\n", errCmd)
  }
  contributionLog := strings.Split(string(out), "\n")
  fileContribs := []fileContribution{}
  for _, contribution := range contributionLog {
    splitLog := strings.Split(contribution, "||")
    if (len(splitLog) > 1) {
      fileContribs = append(fileContribs, fileContribution{
        Name: splitLog[0],
        Email: splitLog[1],
        Commit: splitLog[2],
      })
    }
  }
  return fileContribs
}

// includePath evaluates whether to include a path in the resulting graph or not
func includePath(path string) bool {
  // TODO: should be a list coming from .gitignore
  return !strings.HasPrefix(path, ".") && !strings.HasPrefix(path, "node_modules")
}

