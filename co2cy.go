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
  // "encoding/json"
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
var processedNodes = make(map[string]bool)
var b strings.Builder

// Contains tells whether a contains x.
func Contains(a []string, x string) bool {
  for _, n := range a {
    if x == n {
      return true
    }
  }
  return false
}

func main() {
  verbose := flag.Bool("verbose", false, "log iteration through file tree")
  flag.Parse()

  err := filepath.Walk("src/", func(path string, info os.FileInfo, err error) error {
    if err != nil {
      return err
    }

    if (*verbose) {
      fmt.Println("Fullpath: " + path)
    }

    args :=  []string{"log", "--format=\"%an\"", path}
    cmd := exec.Command("git", args...)
    out, errCmd := cmd.CombinedOutput()
    if errCmd != nil {
      log.Fatalf("cmd.Run() failed with %s\n", errCmd)
    }
    gitlog := string(out)
    contribs := strings.Split(gitlog, "\"")
    if (len(contribs) > 0 && *verbose) {
      fmt.Println(contribs)
    }

    pathSegments := strings.Split(path, "/")
    for i, element := range pathSegments {

      if (*verbose) {
        fmt.Println("Pathsegment: " + element)
        fmt.Println(info.IsDir())
      }

      fmt.Fprintf(&b, "%d-", i)
      if (processedNodes[b.String() + element] != true) {
        ext := ""
        pre := "a_"
        id := strings.Replace(element, ".", "_", -1)
        id = pre + id
        id = strings.Replace(id, "-", "_", -1)
        id += strconv.Itoa(i)
        if (info.IsDir() == false) {
          groups := strings.Split(element, ".")
          ext = groups[len(groups) - 1]
        }
        parentIndex := i - 1
        if (parentIndex < 0) {
          parentIndex = 0
        }

        ParentId := strings.Replace(pathSegments[parentIndex], ".", "_", -1)
        ParentId = pre + ParentId
        ParentId = strings.Replace(ParentId, "-", "_", -1)
        ParentId += strconv.Itoa(parentIndex)

        if (ext != "DS_Store") { // TODO: find way to run massive Cypher query before including those
          myNode := node{
            Name: element,
            Size: strconv.FormatInt(info.Size(), 10),
            Level: i,
            Extension: ext,
            Id: id,
            IsDir: info.IsDir(),
            ModTime: info.ModTime().String(),
            ParentName: pathSegments[parentIndex],
            ParentId : ParentId,
            Contributers: contribs,
          }
          if (*verbose) {
            fmt.Println(myNode)
          }

          nodes = append(nodes, myNode)

          processedNodes[b.String() + element] = true
        }
      }
      b.Reset()
    }

    return nil
  })

  if (*verbose) {
    fmt.Println("")
    fmt.Println("------------------------------------------------------------------------")
    fmt.Println("")
  }


  processedNodes := []string{}
  processedContributers := []string{}
  for i := range nodes {
    currentFile := nodes[i]
    if (currentFile.Name != "") {
      label := "directory"
      if (currentFile.IsDir != true) {
        label = "file"
      }

      if (!Contains(processedNodes, currentFile.Id)) {
        fmt.Println("CREATE (" + currentFile.Id + ":" + label + " { name: '" + currentFile.Name + "', parentName: '" + currentFile.ParentName + "', isDir: " + strconv.FormatBool(currentFile.IsDir) + ", size: " + currentFile.Size + " , time: '" + currentFile.ModTime + "', extension: '" +  currentFile.Extension + "' })")
        processedNodes = append(processedNodes, currentFile.Id)
      }

      for _, c := range currentFile.Contributers {
        if (len(c) > 3) {
          contributerId := "c_" + strings.Replace(c, " ", "", -1)
          if (!Contains(processedContributers, c)) {
            fmt.Println("CREATE (" + contributerId + ":" + "person" + " { name: '" + c + "' })")
            processedContributers = append(processedContributers, c)
          }
          fmt.Println("CREATE (" + currentFile.Id + ")<-[:EDITED]-(" + contributerId + ")")
        }
      }

      if (currentFile.Id != currentFile.ParentId) {
        fmt.Println("CREATE (" + currentFile.Id + ")-[:IN_FOLDER]->(" + currentFile.ParentId + ")")
      }
    }
  }

  // z, _ := json.Marshal(nodes)
  // fmt.Println("UNWIND " + string(z) + " AS n")
  // fmt.Println("CREATE (m:file { name: n.name })")
  // fmt.Println("RETURN m")


  if err != nil {
    log.Println(err)
  }
}

