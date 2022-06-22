package main

import (
  "fmt"
  "flag"
  "path/filepath"
  "os"
  "os/exec"
  "log"
  "strings"
)

type fileInfo struct {
  Name string
  Path string
  Url string
  Size int64
  Level int
  IsDir bool
  Id string
  Extension string
  ModTime int64
  ParentPath string
  Contributions []fileContribution
  CommitCount int
}
var nodes []fileInfo
var processedFiles = make(map[string]bool)
var processedNodes = make(map[string]bool)
var processedContributersSum = make(map[string]int)
var processedContributionsSum = make(map[string]int)
var verbose bool
var simplified bool
var repoPath string
var gitRepoUrl string

// initFlags parses the command line flags
func initFlags() {
  flag.BoolVar(&verbose, "verbose", false, "logs iteration through file tree to console")
  flag.BoolVar(&simplified, "simplified", false, "won't create a distinct relationship for each commit when set")
  flag.StringVar(&repoPath, "path", ".", "the full path of the repository")
  flag.Parse()
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

func main() {

  initFlags()
  verboseLog("repoPath: " + repoPath)
  gitRepoUrl = getGitRemoteUrl(repoPath)
  verboseLog("gitRepoUrl: " + gitRepoUrl)

  err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
    if err != nil {
      return err
    }
    if repoPath != "." {
      path = path[len(repoPath):len(path)]
    }
    if (len(path) > 0 && includePath(path)) {
      verboseLog("path: " + path)

      pathSegments := strings.Split(path, "/")
      fileDepth := len(pathSegments) - 1
      fileName := info.Name()
      verboseLog("fileName: " + fileName)

      if (processedFiles[path] != true) {
        parentPath := strings.Join(pathSegments[:len(pathSegments)-1], "/")
        verboseLog("parentPath: " + parentPath)
        parentDepth := fileDepth - 1
        if (parentDepth < 0) {
          parentDepth = 0
        }

        contributions := getGitLog(path, repoPath)

        nodes = append(nodes, fileInfo{
          Name: fileName,
          Path: path,
          Url: buildGitHubUrl(gitRepoUrl, path, info.IsDir()),
          Size: info.Size(),
          Level: fileDepth,
          Extension: getFileExtension(info),
          Id: createCypherFriendlyVarName(path, fileDepth),
          IsDir: info.IsDir(),
          ModTime: info.ModTime().Unix(),
          ParentPath: parentPath,
          Contributions: contributions,
          CommitCount: len(contributions),
        })
        processedFiles[path] = true
      }
    }

    return nil
  })

  verboseLog("")
  verboseLog("------------------------------------------------------------------------")
  verboseLog("")

  for _, currentFile := range nodes {
    var processedContributers = make(map[string]int)
    var processedContributions = make(map[string]int)
    fmt.Println(":BEGIN")
    label := getLabelForFileNode(currentFile)

    if (!processedNodes[currentFile.Path]) {
      fmt.Println(fileInfoToCypher(currentFile, label))
      processedNodes[currentFile.Path] = true
    }

    if (label == "file") {
      for _, contribution := range currentFile.Contributions {
        contributerId := createCypherFriendlyVarName(contribution.Name, 0)
        if (processedContributers[contributerId] < 1) {
          fmt.Println(contributerToCypher(contributerId, contribution.Name, contribution.Email))
          processedContributers[contributerId] = 0
        }
        processedContributers[contributerId] += 1
        processedContributersSum[contributerId] += 1

        contributionId := currentFile.Id + "__" + contributerId
        if (processedContributions[contributionId] < 1) {
          fmt.Println(contributionToCypher(currentFile.Id, contributerId, contributionId))
          processedContributions[contributionId] = 0
        }
        if (simplified != true) {
          fmt.Println(commitToCypher(currentFile.Id, contributerId, contribution))
        }
        processedContributions[contributionId] += 1
        processedContributionsSum[contributionId] += 1
      }
    }

    fmt.Println(";")
    fmt.Println(":COMMIT")

    if (len(currentFile.ParentPath) > 0) {
      fmt.Println(":BEGIN")
      fmt.Println(folderStructureToCypher(currentFile))
      fmt.Println(";")
      fmt.Println(":COMMIT")
    }
  }

  for contributerId, contributionCount := range processedContributersSum {
    fmt.Println(":BEGIN")
    fmt.Println(contributerToCypherUpdate(contributerId, contributionCount))
    fmt.Println(";")
    fmt.Println(":COMMIT")
  }

  for contributionId, commitCount := range processedContributionsSum {
    fmt.Println(":BEGIN")
    fmt.Println(contributionToCypherUpdate(contributionId, commitCount))
    fmt.Println(";")
    fmt.Println(":COMMIT")
  }

  fmt.Println(":BEGIN")
  fmt.Println(removeProperty("_tempId"))
  fmt.Println(";")
  fmt.Println(":COMMIT")


  if err != nil {
    log.Println(err)
  }
}

