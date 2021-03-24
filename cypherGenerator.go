package main

import (
  "strings"
  "strconv"
  "regexp"
)

var reStr = regexp.MustCompile(`\W`)

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
  properties := ("{ name: '" + currentFile.Name + "', url: '" + currentFile.Url + "'")

  if (!currentFile.IsDir) {
    properties += (", " + "size: " + strconv.FormatInt(currentFile.Size, 10) + ", " +
    "commitCount: " + strconv.Itoa(currentFile.CommitCount) + ", " +
    "lastModifiedDateTime: datetime({ epochseconds: " + strconv.FormatInt(currentFile.ModTime, 10) + " }), " +
    "lastModifiedTimestamp: " + strconv.FormatInt(currentFile.ModTime, 10) + ", " +
    "extension: '" + currentFile.Extension + "'")
  }
  properties += " }"
  return "CREATE (" + currentFile.Id + ":" + label + " " + properties + ")"
}

// contributerToCypher returns a cypher statement to create node for a given contributer
func contributerToCypher(contributerId, contributerName, contributerEmail string) string {
  return ("CREATE (" + contributerId + ":" + "person" + " { name: '" + contributerName + "', email: '" + contributerEmail + "' })")
}

// contributerToCypherUpdate returns a cypher statement to update a given contributer's commitCount
func contributerToCypherUpdate(contributerId string, commitCount int) string {
  return ("SET " + contributerId + ".commitCount = " + strconv.Itoa(commitCount))
}

// contributionToCypher returns to cypher statement to create a relationship between a file and a contributer
func contributionToCypher(fileId, contributerId string, contributionId string) string {
  return "CREATE (" + fileId + ")<-[" + contributionId + ":EDITED]-(" + contributerId + ")"
}

// contributionToCypher returns to cypher statement to create a relationship between a file and a contributer
func commitToCypher(fileId, contributerId string, contribution fileContribution) string {
  properties := "hash: '" + contribution.Hash + "', abbreviatedHash: '" + contribution.AbbreviatedHash + "', dateTime: datetime('" + contribution.DateTime + "')"
  return "CREATE (" + fileId + ")<-[:COMMIT { " + properties + " }]-(" + contributerId + ")"
}

// contributionToCypherUpdate returns a cypher statement to update a given contribution's commitCount
func contributionToCypherUpdate(contributionId string, commitCount int) string {
  return "SET " + contributionId + ".commitCount = " + strconv.Itoa(commitCount)
}

// folderStructureToCypher returns to cypher statement to create a relationship between a file and its parent folder
func folderStructureToCypher(currentFile fileInfo) string {
  return "CREATE (" + currentFile.Id + ")-[:IN_FOLDER]->(" + currentFile.ParentId + ")"
}
