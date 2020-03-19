package main

import (
  "fmt"
  "flag"
  "path/filepath"
  "os"
  "log"
  "strings"
  "strconv"
  "os/exec"
  "regexp"
)

type node struct {
  Name string
  Size string
  Level int
  IsDir bool
  Id string
  Extension string
  ModTime string
  ParentName string
  ParentId string
  Contributers []string
}
var nodes []node
var processedElements = make(map[string]bool)
var stringBuilder strings.Builder
var verbose bool
var reStr = regexp.MustCompile(`\W`)

// contains tells whether a contains x.
// from https://yourbasic.org/golang/find-search-contains-slice/ by Stefan Nilsson
func contains(a []string, x string) bool {
  for _, n := range a {
    if x == n {
      return true
    }
  }
  return false
}

// initFlags parses the command line flags
func initFlags() {
  flag.BoolVar(&verbose, "verbose", false, "log iteration through file tree")
  flag.Parse()
}

// includePath evaluates whether to include a path in the resulting graph or not
func includePath(path string) bool {
  // TODO: should be a list coming from .gitignore
  return !strings.HasPrefix(path, ".") && !strings.HasPrefix(path, "node_modules")
}

// getGitLog returns the list of contributers of a given path
func getGitLog(path string) string {
  args :=  []string{"log", "--format=\"%an\"", path}
  cmd := exec.Command("git", args...)
  out, errCmd := cmd.CombinedOutput()
  if errCmd != nil {
    log.Fatalf("cmd.Run() failed with %s\n", errCmd)
  }
  return string(out)
}

// getUniqueNameString creates a unique string for a file based on its nested depth in the folder and its name
func getUniqueNameString(index int, element string) string {
  fmt.Fprintf(&stringBuilder, "%d-", index)
  stringBuilder.WriteString(element)
  return stringBuilder.String()
}

// getFileExtension returns the extension for a given file's full name
func getFileExtension (element string) string {
  stringSegments := strings.Split(element, ".")
  return stringSegments[len(stringSegments) - 1]
}

func main() {

  initFlags()

  err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
    if err != nil {
      return err
    }

    if (includePath(path)) {

      if (verbose) {
        fmt.Println("Fullpath: " + path)
      }

      contributers := strings.Split(getGitLog(path), "\"")

      if (len(contributers) > 0 && verbose) {
        fmt.Println(contributers)
      }

      pathSegments := strings.Split(path, "/")
      depth := len(pathSegments) - 1
      element := info.Name()

      if (verbose) {
        fmt.Println("Pathsegment: " + element)
        fmt.Println(info.IsDir())
      }

      elementString := getUniqueNameString(depth, element)
      if (processedElements[elementString] != true) {
        pre := "a_"
        id := strings.Replace(element, ".", "_", -1)
        id = pre + id
        id = reStr.ReplaceAllString(id, "$1")
        id += strconv.Itoa(depth)

        fileExtension := ""
        if (info.IsDir() == false) {
          groups := strings.Split(element, ".")
          fileExtension = groups[len(groups) - 1]
        }
        parentDepth := depth - 1
        if (parentDepth < 0) {
          parentDepth = 0
        }

        ParentId := strings.Replace(pathSegments[parentDepth], ".", "_", -1)
        ParentId = pre + ParentId
        ParentId = reStr.ReplaceAllString(ParentId, "$1")
        ParentId += strconv.Itoa(parentDepth)

        myNode := node{
          Name: element,
          Size: strconv.FormatInt(info.Size(), 10),
          Level: depth,
          Extension: fileExtension,
          Id: id,
          IsDir: info.IsDir(),
          ModTime: info.ModTime().String(),
          ParentName: pathSegments[parentDepth],
          ParentId : ParentId,
          Contributers: contributers,
        }

        if (verbose) {
          fmt.Println(myNode)
        }

        nodes = append(nodes, myNode)

        processedElements[elementString] = true
      }
      stringBuilder.Reset()
    }

    return nil
  })

  if (verbose) {
    fmt.Println("")
    fmt.Println("------------------------------------------------------------------------")
    fmt.Println("")
  }

  // TODO: use make like above with processedElements
  processedNodes := []string{}
  processedContributers := []string{}
  processedContributions := []string{}
  for i := range nodes {
    currentFile := nodes[i]
    if (currentFile.Name != "") {
      label := "directory"
      if (currentFile.IsDir != true) {
        label = "file"
      }

      if (!contains(processedNodes, currentFile.Id)) {
        fmt.Println("CREATE (" + currentFile.Id + ":" + label + " { name: '" + currentFile.Name + "', parentName: '" + currentFile.ParentName + "', isDir: " + strconv.FormatBool(currentFile.IsDir) + ", size: " + currentFile.Size + " , time: '" + currentFile.ModTime + "', extension: '" +  currentFile.Extension + "' })")
        processedNodes = append(processedNodes, currentFile.Id)
      }

      if (label == "file") {
        for _, c := range currentFile.Contributers {
          if (len(c) > 3) {
            contributerId := "c_" + strings.Replace(c, " ", "", -1)
            contributerId = reStr.ReplaceAllString(contributerId, "$1")
            if (!contains(processedContributers, c)) {
              fmt.Println("CREATE (" + contributerId + ":" + "person" + " { name: '" + c + "' })")
              processedContributers = append(processedContributers, c)
            }
            contributionCypherStatement := "CREATE (" + currentFile.Id + ")<-[:EDITED]-(" + contributerId + ")"
            if (!contains(processedContributions, contributionCypherStatement)) {
              processedContributions = append(processedContributions, contributionCypherStatement)
              fmt.Println(contributionCypherStatement)
            }
          }
        }
      }

      if (currentFile.Id != currentFile.ParentId) {
        fmt.Println("CREATE (" + currentFile.Id + ")-[:IN_FOLDER]->(" + currentFile.ParentId + ")")
      }
    }
  }

  if err != nil {
    log.Println(err)
  }
}

