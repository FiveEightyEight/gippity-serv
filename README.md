# gippity-serv


## Run Dev Server
I recommend using the [Air](https://github.com/air-verse/air) lib for hot reloading the Go server. If not you could run the server via
```shell
go run -v cmd/main.go
```

## Requirements
Requires a `.env` file with the following variables
```shell
# optional
PORT=
# OpenAI API Key
API_KEY=
# Postgres url
DATABASE_URL
# Salt for passwords
HASH_SALT
```

Realized I needed a database to properly send messages to open ai... then I realized I wanted users so auth + db needed. 
