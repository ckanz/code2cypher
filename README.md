# code2cypher

This go packages creates a list of Cypher statements to turn a GIT repository into a Neo4j graph. The data is modeled the following way:

- labels
  - directory
  - file
  - person
- relationships
  - EDITED
  - IN_FOLDER

To turn a GIT repository into a graph, follow these steps:

- build the executable bu running `go build` in the root of this project
- place the executable in the root folder of your repository
- execute the file (the Cypher statements are written to StdOut)
- pipe the Cypher statements to Cypher shell or into a .cypher file you can import to Neo4j Desktop

