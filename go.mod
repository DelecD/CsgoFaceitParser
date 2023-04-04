module main

require (
	github.com/dustin/go-heatmap v0.0.0-20180603032536-b89dbd73785a
	github.com/go-sql-driver/mysql v1.6.0
	github.com/gogo/protobuf v1.3.2
	github.com/golang/geo v0.0.0-20210211234256-740aa86cb551
	github.com/llgcode/draw2d v0.0.0-20200930101115-bfaf5d914d1e
	github.com/markus-wa/demoinfocs-golang/v2 v2.13.0 // indirect
	github.com/markus-wa/go-unassert v0.1.2
	github.com/markus-wa/gobitread v0.2.3
	github.com/markus-wa/godispatch v1.4.1
	github.com/markus-wa/quickhull-go/v2 v2.1.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
)

replace github.com/dustin/go-heatmap => github.com/markus-wa/go-heatmap v1.0.0

go 1.13
