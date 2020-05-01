package main

import "testing"

func TestCreateCypherFriendlyVarName(t *testing.T) {
  testTables := []struct {
    name string
    index int
    result string
  }{
    { "bla", 42, "a_bla42" },
    { "myFile.test.java", 2, "a_myFile_test_java2" },
    { "inS4n£$...$%£name.java", 2, "a_inS4n___name_java2" },
  }

  for _, table := range testTables {
    result := createCypherFriendlyVarName(table.name, table.index)
    if (result != table.result) {
      t.Errorf("createCypherFriendlyVarName was incorrect, got: %s, want: %s.", result, table.result)
    }
  }
}

func TestFileInfoToCypher(t *testing.T) {
  testTables := []struct {
    testFile fileInfo
    testLabel string
    cypherResult string
  }{
    {
      fileInfo { Id: "fileId", Name: "someDir", IsDir: true, },
      "testLabel",
      "CREATE (fileId:testLabel { name: 'someDir', url: '' });",
    },
    {
      fileInfo { Name: "someDir", IsDir: true, Url: "https://github.com/someName/someRepo/tree/master/someDir" },
      "testLabel",
      "CREATE (:testLabel { name: 'someDir', url: 'https://github.com/someName/someRepo/tree/master/someDir' });",
    },
    {
      fileInfo { Id: "fileId", Name: "someFile", IsDir: false, Size: 42, CommitCount: 23, ModTime: 111222333, Extension: "go" },
      "testLabel",
      "CREATE (fileId:testLabel { name: 'someFile', url: '', size: 42, commitCount: 23, lastModifiedDateTime: datetime({ epochseconds: 111222333 }), lastModifiedTimestamp: 111222333, extension: 'go' });",
    },
  }
  for _, table := range testTables {
    result := fileInfoToCypher(table.testFile, table.testLabel)
    if (result != table.cypherResult) {
      t.Errorf("fileInfoToCypher was incorrect, got: %s, want: %s.", result, table.cypherResult)
    }
  }
}

func TestContributerToCypher(t *testing.T) {
  testTables := []struct {
    contributerId string
    contributerName string
    contributerEmail string
    cypherResult string
  }{
    {
      "someId",
      "William T. Riker",
      "wt.riker@starfleet.gov",
      "CREATE (someId:person { name: 'William T. Riker', email: 'wt.riker@starfleet.gov' });",
    },
  }
  for _, table := range testTables {
    result := contributerToCypher(table.contributerId, table.contributerName, table.contributerEmail)
    if (result != table.cypherResult) {
      t.Errorf("contributerToCypher was incorrect, got: %s, want: %s.", result, table.cypherResult)
    }
  }
}

func TestContributionToCypher(t *testing.T) {
  testTables := []struct {
    fileId string
    contributerId string
    cypherResult string
  }{
    {
      "someFile_java",
      "William_T__Riker",
      "CREATE (someFile_java)<-[:EDITED]-(William_T__Riker);",
    },
  }
  for _, table := range testTables {
    result := contributionToCypher(table.fileId, table.contributerId)
    if (result != table.cypherResult) {
      t.Errorf("contributionToCypher was incorrect, got: %s, want: %s.", result, table.cypherResult)
    }
  }
}

func TestFolderStructureToCypher(t *testing.T) {
  testTables := []struct {
    file fileInfo
    cypherResult string
  }{
    {
      fileInfo { Id: "someFileId", ParentId: "someParentId" },
      "CREATE (someFileId)-[:IN_FOLDER]->(someParentId);",
    },
  }
  for _, table := range testTables {
    result := folderStructureToCypher(table.file)
    if (result != table.cypherResult) {
      t.Errorf("folderStructureToCypher was incorrect, got: %s, want: %s.", result, table.cypherResult)
    }
  }
}
