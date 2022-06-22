const acorn = require('acorn')
const fs = require('fs')
const path = require('path')

const SystemPath = path.join(__dirname, '/')

const getAllFiles = function (dirPath, arrayOfFiles) {
  const files = fs.readdirSync(dirPath)

  arrayOfFiles = arrayOfFiles || []

  files.forEach(function (file) {
    if (fs.statSync(dirPath + '/' + file).isDirectory()) {
      arrayOfFiles = getAllFiles(dirPath + '/' + file, arrayOfFiles)
    } else {
      const extension = file.split('.').pop()
      if (extension === 'js') arrayOfFiles.push(path.join(__dirname, dirPath, '/', file))
    }
  })

  return arrayOfFiles
}

const toCypher = (type, size, name, varName, index, parentIndex, rawNode) => {
  console.log(`CREATE (${varName}_${parentIndex !== undefined ? parentIndex : ''})-[:DECLARES]->(${parentIndex === undefined ? varName + '_' + index : ''}:${type} { size: ${size}, name: '[${name}]', ${Object.entries(rawNode).map(e => `_${e[0]}:'# ${e[1]}'`)} })`)
}

const allFiles = getAllFiles('src/')

allFiles.forEach((f, i) => {
  try {
    console.log(':BEGIN')
    const varName = `f_${i}`
    console.log(`MERGE (${varName}_:file { path: '${f.split(SystemPath)[1]}' })`)
    const file = fs.readFileSync(f, 'utf8')

    const r = acorn.parse(file, { ecmaVersion: 2020, sourceType: 'module' })

    const logNode = (n, index, parentIndex) => {
      if (n.key) {
        toCypher(n.type, n.end - n.start, n.key.name, varName, index, parentIndex, n)
        return
      }
      const d = n.declarations ? n.declarations[0] : n.declaration
      if (!d) return
      if (d.id) {
        toCypher(n.type, n.end - n.start, d.id.name, varName, index, parentIndex, n)
      } else if (d.declarations) {
        toCypher(n.type, n.end - n.start, d.declarations[0].id.name, varName, index, parentIndex, n)
        if (d.declarations[0].body) logNodeBody(d.declarations[0].body.body, index)
      }
      if (d.body) logNodeBody(d.body.body, index)
    }

    const logNodeBody = (nb, parentIndex) => {
      nb.forEach((n, index) => {
        logNode(n, index, parentIndex)
        // if (n.body) logNodeBody(n.body.body, index)
      })
    }

    logNodeBody(r.body)
    console.log(';')
    console.log(':COMMIT')
  } catch (e) {
    console.log(';')
    console.log(':ROLLBACK')
  }
})
