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

type fileContributer struct {
  Name string
  Email string
  // commits []string
  // commitCount int
}
type fileInfo struct {
  Name string
  Size int64
  Level int
  IsDir bool
  Id string
  Extension string
  ModTime int64
  ParentName string
  ParentId string
  Contributers []string
  CommitCount int
}
var nodes []fileInfo
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
func getGitLog(path string) []string {
  args :=  []string{"log", "--format=%an", path}
  cmd := exec.Command("git", args...)
  out, errCmd := cmd.CombinedOutput()
  if errCmd != nil {
    log.Fatalf("cmd.Run() failed with %s\n", errCmd)
  }
  return strings.Split(string(out), "\n")
}

// getUniqueNameString creates a unique string for a file based on its nested depth in the folder and its name
// TODO: instead of depth, modified timestamp might be a better value to create unique variable names with
func getUniqueNameString(index int, element string) string {
  return strconv.Itoa(index) + "-" + element
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

// getLabelForFileNode returns the correct label for a given element
func getLabelForFileNode(currentFile fileInfo) string {
  if (currentFile.IsDir) {
    return "directory"
  }
  return "file"
}

// fileInfoToCypher returns a cypher statement to create a node for a given file
func fileInfoToCypher(currentFile fileInfo, label string) string {
  properties := (
    "{ name: '" + currentFile.Name + "', " +
    "size: " + strconv.FormatInt(currentFile.Size, 10) + ", " +
    "commitCount: " + strconv.Itoa(currentFile.CommitCount) + ", " +
    "lastModifiedDateTime: datetime({ epochseconds: " + strconv.FormatInt(currentFile.ModTime, 10) + " }), " +
    "lastModifiedTimestamp: " + strconv.FormatInt(currentFile.ModTime, 10) + ", " +
    "extension: '" + currentFile.Extension + "' " +
    "}")
  return "CREATE (" + currentFile.Id + ":" + label + " " + properties + ")"
}

// contributerToCypher returns a cypher statement to create node for a given contributer
func contributerToCypher(contributerId, contributer string) string {
  return ("CREATE (" + contributerId + ":" + "person" + " { name: '" + contributer + "' })")
}

// contributionToCypher returns to cypher statement to create a relationship between a file and a contributer
func contributionToCypher(fileId, contributerId string) string {
  return "CREATE (" + fileId + ")<-[:EDITED]-(" + contributerId + ")"
}

// folderStructureToCypher returns to cypher statement to create a relationship between a file and its parent folder
func folderStructureToCypher(currentFile fileInfo) string {
  return "CREATE (" + currentFile.Id + ")-[:IN_FOLDER]->(" + currentFile.ParentId + ")"
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

        contributers := getGitLog(path)

        nodes = append(nodes, fileInfo{
          Name: fileName,
          Size: info.Size(),
          Level: fileDepth,
          Extension: getFileExtension(info),
          Id: createCypherFriendlyVarName(fileName, fileDepth),
          IsDir: info.IsDir(),
          ModTime: info.ModTime().Unix(),
          ParentName: pathSegments[parentDepth],
          ParentId : createCypherFriendlyVarName(pathSegments[parentDepth], parentDepth),
          Contributers: contributers,
          CommitCount: len(contributers),
        })
        processedFiles[uniqueNameString] = true
      }
    }

    return nil
  })

  verboseLog("")
  verboseLog("------------------------------------------------------------------------")
  verboseLog("")

  for _, currentFile := range nodes {
    label := getLabelForFileNode(currentFile)

    if (!processedNodes[currentFile.Id]) {
      fmt.Println(fileInfoToCypher(currentFile, label))
      processedNodes[currentFile.Id] = true
    }

    if (label == "file") {
      for _, contributer := range currentFile.Contributers {
        if (len(contributer) > 1) {
          contributerId := createCypherFriendlyVarName(contributer, 0)
          if (!processedContributers[contributer]) {
            fmt.Println(contributerToCypher(contributerId, contributer))
            processedContributers[contributer] = true
          }
          contributionCypherStatement := contributionToCypher(currentFile.Id, contributerId)
          if (!processedContributions[contributionCypherStatement]) {
            fmt.Println(contributionCypherStatement)
            processedContributions[contributionCypherStatement] = true
          }
        }
      }
    }

    if (currentFile.Id != currentFile.ParentId) {
      fmt.Println(folderStructureToCypher(currentFile))
    }
  }

  if err != nil {
    log.Println(err)
  }
}

