run:
  go build -o out && ./out

up:
  goose -dir "sql/schema" postgres "postgres://pd:@localhost:5432/chirpy" up

down:
  goose -dir "sql/schema" postgres "postgres://pd:@localhost:5432/chirpy" down
