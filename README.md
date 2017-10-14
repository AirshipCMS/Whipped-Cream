# Whipped Cream

forward proxy cache with memcached

## Memcache

run locally for dev

```sh
run --name memcache -d -p 11211:11211 memcached:alpine
```


## Proof so far

```sh
TARGET_URL=https://swapi.co ./Whipped-Cream
```

```sh
curl -v localhost:8080/api/starships/9
# should be cache miss
curl -v localhost:8080/api/starships/9
# should be cache hit, and fast
```
