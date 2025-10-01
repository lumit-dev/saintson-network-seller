module panel-service

go 1.24.2

require (
	github.com/gorilla/mux v1.8.1
	github.com/saintson-network-seller/additions v0.0.0
	github.com/sirupsen/logrus v1.9.3
)

require (
	github.com/stretchr/testify v1.11.1 // indirect
	golang.org/x/sys v0.35.0 // indirect
)

replace github.com/saintson-network-seller/additions => ../../additions
