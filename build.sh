echo "Building package..."
go build

echo "Adding package to /user/local/bin/ ..."
sudo cp code2cypher /usr/local/bin/

echo "Running tests..."
go test -v -cover
