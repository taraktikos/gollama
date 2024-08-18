# Go + ollama Service

### How to import data

```
// Download archive from https://www.kaggle.com/datasets/conjuring92/wiki-stem-corpus?resource=download
export IMPORT_FILE_PATH=/<path_to_wiki_archive>/archive.zip
go run .
```

### How to run server

```
go run .
```

### How to use

```
curl \
    --header "Content-Type: application/json" \
    --data '{"text": "referencing engine speed"}' \
    http://localhost:8080/gollama.v1.GollamaService/Search
```

```
curl \
    --header "Content-Type: application/json" \
    --data '{"prompt": "Who was the first man to walk on the moon?"}' \
    http://localhost:8080/gollama.v1.GollamaService/GenerateFromSinglePrompt
```
