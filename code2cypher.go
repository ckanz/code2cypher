package main

import (
  "fmt"
  "flag"
  "path/filepath"
  "os"
  "log"
  "strings"
  "strconv"
)

type fileInfo struct {
  Name string
  Url string
  Size int64
  Level int
  IsDir bool
  Id string
  Extension string
  ModTime int64
  ParentName string
  ParentId string
  Contributions []fileContribution
  CommitCount int
}
var nodes []fileInfo
var processedFiles = make(map[string]bool)
var processedNodes = make(map[string]bool)
var processedContributers = make(map[string]bool)
var processedContributions = make(map[string]bool)
var verbose bool
var repoPath string
var gitRepoUrl string

// initFlags parses the command line flags
func initFlags() {
  flag.BoolVar(&verbose, "verbose", false, "log iteration through file tree")
  flag.StringVar(&repoPath, "path", ".", "the full path of the repository")
  flag.Parse()
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

func init() {
  initFlags()
  gitRepoUrl = getGitRemoteUrl(repoPath)
}

func main() {
  err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
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

        contributions := getGitLog(path, repoPath)

        nodes = append(nodes, fileInfo{
          Name: fileName,
          Url: buildGitHubUrl(gitRepoUrl, path, info.IsDir()),
          Size: info.Size(),
          Level: fileDepth,
          Extension: getFileExtension(info),
          Id: createCypherFriendlyVarName(fileName, fileDepth),
          IsDir: info.IsDir(),
          ModTime: info.ModTime().Unix(),
          ParentName: pathSegments[parentDepth],
          ParentId : createCypherFriendlyVarName(pathSegments[parentDepth], parentDepth),
          Contributions: contributions,
          CommitCount: len(contributions),
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
      for _, contribution := range currentFile.Contributions {
        contributerId := createCypherFriendlyVarName(contribution.Name, 0)
        if (!processedContributers[contribution.Name]) {
          // TODO: add 'totalCommits' prop to contributer nodes
          fmt.Println(contributerToCypher(contributerId, contribution.Name, contribution.Email))
          processedContributers[contribution.Name] = true
        }
        // TODO: add array of commit messages as 'commits' prop and length of the array as 'commitCount' prop to the relationship
        contributionCypherStatement := contributionToCypher(currentFile.Id, contributerId)
        if (!processedContributions[contributionCypherStatement]) {
          fmt.Println(contributionCypherStatement)
          processedContributions[contributionCypherStatement] = true
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

