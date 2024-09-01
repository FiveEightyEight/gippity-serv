# gippity-serv

## Requirements
Go at least [`>=1.22.5`](https://go.dev/doc/install)

Requires a `.env` file with the following variables
```shell
# optional
PORT=
# OpenAI API Key
API_KEY=
# Postgres url
DATABASE_URL=postgresql://[userspec@][hostspec][/dbname][?paramspec]
# Salt for passwords
HASH_SALT=
# Secrets for tokens
ACCESS_TOKEN_SECRET=
REFRESH_TOKEN_SECRET=
# location for server timezone
LOCATION=America/New_York
# to detect env, currently not designed for prod
ENV=dev
```

## Install
```shell
go mod tidy
```

## Database 
Realized I needed a database to properly send messages to open ai... then I realized I wanted users so auth + db needed. Currently setup to use postgres using `pgx` package for interactions. 


## Run Dev Server
I recommend using the [Air](https://github.com/air-verse/air) package for hot reloading the Go server. If not you could run the server via
```shell
go run -v cmd/main.go
```
