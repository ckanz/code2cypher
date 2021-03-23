# code2cypher

This [Go](https://golang.org/) package creates a list of [Cypher](https://neo4j.com/developer/cypher-query-language/) statements to turn a GIT repository into a [Neo4j](https://neo4j.com/developer/get-started/) graph. The data is modeled the following way:

- labels
  - directory
  - file
  - person
- relationships
  - EDITED
  - IN_FOLDER

To turn a GIT repository into a graph, follow these steps:

- build the executable by running `go build` in the root of this project
- place the executable in the root folder of your repository
- execute the file (the Cypher statements are written to StdOut)
- pipe the Cypher statements to [Cypher Shell](https://neo4j.com/docs/operations-manual/current/tools/cypher-shell/) or into a .cypher file you can import to [Neo4j Desktop](https://neo4j.com/developer/neo4j-desktop/)

Below is a screenshot of what the graph from the [hellogitworld](https://github.com/githubtraining/hellogitworld) repository looks like in Neo4j Desktop

![Neo4j Desktop Screenshot](https://github.com/ckanz/code2cypher/blob/master/screenshot.png)
