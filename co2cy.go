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
}
var nodes []node
var processedNodes = make(map[string]bool)
var b strings.Builder

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
    // args := []string{"shortlog", "-n", "-s", "src"}
    // cmd := exec.Command("git", args...)
    cmd := exec.Command("git", "shortlog", "-s", path)
    out, errCmd := cmd.CombinedOutput()
    if errCmd != nil {
      // log.Fatalf("cmd.Run() failed with %s\n", errCmd)
    }
    gitlog := string(out)

    pathSegments := strings.Split(path, "/")
    for i, element := range pathSegments {

      if (*verbose) {
        fmt.Println("Pathsegment: " + element)
        fmt.Println(info.IsDir())
      }

      fmt.Fprintf(&b, "%d-", i)
      if (processedNodes[b.String() + element] != true) {
        ext := ""
        id := strings.Replace(element, ".", "_", -1)
        id = strings.Replace(id, "-", "_", -1)
        // id = strings.Replace(id, "3", "A", -1) // TODO: replace find replacing \d regex
        // id = strings.Replace(id, "8", "A", -1) // TODO: replace find replacing \d regex
        id += strconv.Itoa(i)
        if (info.IsDir() == false) {
          groups := strings.Split(element, ".")
          ext = groups[len(groups) - 1]
        }
        parentIndex := i - 1
        if (parentIndex < 0) {
          parentIndex = 0
        }

        Parentid := strings.Replace(pathSegments[parentIndex], ".", "_", -1)
        Parentid = strings.Replace(Parentid, "-", "_", -1)
        Parentid += strconv.Itoa(parentIndex)

        if (ext != "svg") { // TODO: find way to run massive Cypher query before including those
        // if (true) {
          myNode := node{
            Name: element,
            Size: strconv.FormatInt(info.Size(), 10),
            Level: i,
            Extension: ext,
            Id: id,
            IsDir: info.IsDir(),
            ModTime: info.ModTime().String(),
            ParentName: pathSegments[parentIndex],
            ParentId : Parentid,
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
    fmt.Println(gitlog)
    return nil
  })

  if (*verbose) {
    fmt.Println("")
    fmt.Println("------------------------------------------------------------------------")
    fmt.Println("")
  }

  for i := range nodes {
    currentFile := nodes[i]
    if (currentFile.Name != "") {
      label := "directory"
      if (currentFile.IsDir != true) {
        label = "file"
      }

      fmt.Println("CREATE (" + currentFile.Id + ":" + label + " { name: '" + currentFile.Name + "', parentName: '" + currentFile.ParentName + "', isDir: " + strconv.FormatBool(currentFile.IsDir) + ", size: " + currentFile.Size + " , time: '" + currentFile.ModTime + "', extension: '" +  currentFile.Extension + "' })")

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

