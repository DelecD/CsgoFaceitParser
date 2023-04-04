go get -u github.com/markus-wa/demoinfocs-golang/v2/pkg/demoinfocs
SET GOOS=windows
go build
SET GOOS=linux
go build
SET GOOS=windows