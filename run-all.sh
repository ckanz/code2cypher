cwd=$(pwd)

./code2cypher --path $1

cp dependencytree2cypher.js $1
cp acorn2cypher.js $1
cd $1
node dependencytree2cypher.js $1
node acorn2cypher.js $1
cd $cwd
