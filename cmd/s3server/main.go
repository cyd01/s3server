package main

import "github.com/cyd01/s3server"

func main() {
	s3server.Main(":9000", "./data")
}
