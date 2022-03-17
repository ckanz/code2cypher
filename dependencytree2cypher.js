const dependencyTree = require('dependency-tree')
const path = require('path')

const SystemPath = path.join(__dirname, '/')

const tree = dependencyTree({
  filename: 'src/index.js',
  directory: '.',
  webpackConfig: 'webpack.config.js'
})

const createFileCypher = (fileName, varName = '') => {
  return `MERGE (${varName}:file { path: '${fileName.split(SystemPath)[1]}' })`
}

const listFileImports = file => {
  Object.entries(file).forEach(([currentFile, fileImports]) => {
    console.log(':BEGIN')
    console.log(createFileCypher(currentFile, 'f'))
    Object.keys(fileImports).forEach((fileImport, index) => {
      const varName = `i${index}`
      console.log(createFileCypher(fileImport, varName))
      console.log(`MERGE (f)-[:IMPORTS]->(${varName})`)
    })
    console.log(';')
    console.log(':COMMIT')
    listFileImports(fileImports)
  })
}

listFileImports(tree)
