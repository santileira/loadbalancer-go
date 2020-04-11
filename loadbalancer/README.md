# loadbalancer

`go build -o loadbalancer .`

`./loadbalancer --port=5000 --backends=http://localhost:3001,http://localhost:3002,http://localhost:3003 --algorithm=round-robin`