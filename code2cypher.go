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
var processedFiles = make(map[string]bool)
var processedNodes = make(map[string]bool)
var processedContributers = make(map[string]bool)
var processedContributions = make(map[string]bool)
var verbose bool
var reStr = regexp.MustCompile(`\W`)

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
// TODO: instead of depth, modified timestamp might be a better value to create unique variable names with
func getUniqueNameString(index int, element string) string {
  var stringBuilder strings.Builder
  fmt.Fprintf(&stringBuilder, "%d-", index)
  stringBuilder.WriteString(element)
  uniqueNameString := stringBuilder.String()
  stringBuilder.Reset()
  return uniqueNameString
}

// getFileExtension returns the extension for a given file's full name
func getFileExtension (info os.FileInfo) string {
  if (info.IsDir() == false) {
    stringSegments := strings.Split(info.Name(), ".")
    return stringSegments[len(stringSegments) - 1]
  }
  return ""
}

// verboseLog writes a string to stdOut if the verbose flag is set
func verboseLog(toLog string) {
  if (verbose) {
    fmt.Println(toLog)
  }
}

// createCypherFriendlyVarName produces a string for a filename and its nested depth that can be safely used in cypher as a variable name
func createCypherFriendlyVarName(s string, i int) string {
  id := strings.Replace(s, ".", "_", -1)
  id = "a_" + id
  id = reStr.ReplaceAllString(id, "$1")
  id += strconv.Itoa(i)
  return id
}

func main() {

  initFlags()

  err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
    if err != nil {
      return err
    }

    if (includePath(path)) {
      verboseLog("Fullpath: " + path)

      pathSegments := strings.Split(path, "/")
      fileDepth := len(pathSegments) - 1
      fileName := info.Name()
      uniqueNameString := getUniqueNameString(fileDepth, fileName)

      if (processedFiles[uniqueNameString] != true) {
        parentDepth := fileDepth - 1
        if (parentDepth < 0) {
          parentDepth = 0
        }

        myNode := node{
          Name: fileName,
          Size: strconv.FormatInt(info.Size(), 10),
          Level: fileDepth,
          Extension: getFileExtension(info),
          Id: createCypherFriendlyVarName(fileName, fileDepth),
          IsDir: info.IsDir(),
          ModTime: info.ModTime().String(), // TODO: should be available as numeric timestamp too so it can be used in conditional styling
          ParentName: pathSegments[parentDepth],
          ParentId : createCypherFriendlyVarName(pathSegments[parentDepth], parentDepth),
          Contributers: strings.Split(getGitLog(path), "\""),
        }

        nodes = append(nodes, myNode)
        processedFiles[uniqueNameString] = true
      }
    }

    return nil
  })

  verboseLog("")
  verboseLog("------------------------------------------------------------------------")
  verboseLog("")

  for i := range nodes {
    currentFile := nodes[i]
    label := "directory"
    if (currentFile.IsDir != true) {
      label = "file"
    }

    if (!processedNodes[currentFile.Id]) {
      fmt.Println("CREATE (" + currentFile.Id + ":" + label + " { name: '" + currentFile.Name + "', parentName: '" + currentFile.ParentName + "', isDir: " + strconv.FormatBool(currentFile.IsDir) + ", size: " + currentFile.Size + " , time: '" + currentFile.ModTime + "', extension: '" +  currentFile.Extension + "' })")
      processedNodes[currentFile.Id] = true
    }

    if (label == "file") {
      for _, c := range currentFile.Contributers {
        if (len(c) > 3) {
          contributerId := "c_" + strings.Replace(c, " ", "", -1)
          contributerId = reStr.ReplaceAllString(contributerId, "$1")
          if (!processedContributers[c]) {
            fmt.Println("CREATE (" + contributerId + ":" + "person" + " { name: '" + c + "' })")
            processedContributers[c] = true
          }
          contributionCypherStatement := "CREATE (" + currentFile.Id + ")<-[:EDITED]-(" + contributerId + ")"
          if (!processedContributions[contributionCypherStatement]) {
            fmt.Println(contributionCypherStatement)
            processedContributions[contributionCypherStatement] = true
          }
        }
      }
    }

    if (currentFile.Id != currentFile.ParentId) {
      fmt.Println("CREATE (" + currentFile.Id + ")-[:IN_FOLDER]->(" + currentFile.ParentId + ")")
    }
  }

  if err != nil {
    log.Println(err)
  }
}

