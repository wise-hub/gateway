go run ./cmd


go build -o ./bin/api-gateway-ws ./cmd

CGO_ENABLED=1 GOARCH=arm64 GOOS=darwin go build -o ./bin/api-gateway-ws ./cmd
CGO_ENABLED=1 GOARCH=amd64 GOOS=linux go build -o ./bin/api-gateway-ws ./cmd


curl --location 'http://localhost:8200/api/public/accounts' \
--header 'Authorization: generated_token_from_db'


curl --location 'http://localhost:8200/api/public/login' \
--header 'Content-Type: application/json' \
--data '{
  "user": "uuu",
  "pass": "ppp"
}'


curl --location 'http://localhost:8200/api/cache'


ab -n 10000 -c 100 -H "Authorization: generated_token_from_db" http://localhost:8200/api/cache