# open_cache
Just a go server and a db that can act as a cache based on an uut4


## moar to fix:
- set a max size to the server via CLI
    - if max size -> return 500+ code regarding full, wait 
- add remove cache func with id as parameter
- add helm chart
- fix an interface that supports different dbs, like redis etc

## Prompt 1:
```text
i want to create a duckDB golang cache webserver.

I want a path /add - that you add a content which is just a string that you want to cache and then is automatically generates a uut4 id with that cached string, the post request to this path returns the id. All is json.

you can /fetch the cached string with the id, which returns the string. 

then i want everything older than 24 hours to be removed either by a trigger or a scheduler, you choose. 

```

**result** 
`curl -X POST http://localhost:8080/add -H "Content-Type: application/json" -d '{"content":"Hello, DuckDB!"}'`
`curl http://localhost:8080/fetch/<your-uuid>`


## Prompt 2
```text
create unit tests for this app
```

## Prompt 3
```text
Create a docker compose file i can use to try my server
```


## BOL
```bash
go get github.com/gin-gonic/gin
go get github.com/google/uuid
go get github.com/marcboeker/go-duckdb
go get github.com/stretchr/testify
```

## run 
`docker compose up --build`
`docker compose down`
