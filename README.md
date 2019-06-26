# Whipped Cream

## Usage

see [GomaGames/stem.airshipcms.io #281](https://github.com/GomaGames/stem.airshipcms.io/issues/281)

## Docker

### Build

```sh
docker build -t quay.io/airshipcms/whipped-cream:latest .
```

### Run

```sh
docker run --rm -t -v $(pwd)/data:/data quay.io/airshipcms/whipped-cream:latest
```
