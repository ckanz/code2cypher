const acorn = require('acorn')
const fs = require('fs')
const path = require('path')

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

const allFiles = getAllFiles('src/')

allFiles.forEach(f => {
  try {
  console.log('--- ', f)
  const file = fs.readFileSync(f, 'utf8')

  const r = acorn.parse(file, {ecmaVersion: 2020, sourceType: 'module'})

  const logNode = (n, parent) => {
    const d = n.declarations ? n.declarations[0] : n.declaration
    if (!d) return
    if (d.id) {
      console.log(n.type, n.end - n.start, d.id.name)
    } else if (d.declarations) {
      console.log(n.type, n.end - n.start, d.declarations[0].id.name)
    // if (d.declarations[0].body) logNodeBody(d.declarations[0].body.body, d.declarations[0].id.name)
    }
  // if (d.body) logNodeBody(d.body.body, d.id.name)
  }

  const logNodeBody = (nb, parent = '') => {
    nb.forEach(n => {
      logNode(n, parent)
    // if (n.body) logNodeBody(n.body.body, n)
    })
  }

  logNodeBody(r.body)
  } catch {}
})
