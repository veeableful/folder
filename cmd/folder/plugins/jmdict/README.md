# JMdict

This is a plugin for providing JMdict entries data to the Folder CLI.

## Requirements

The [JMdict_e.gz](http://ftp.monash.edu/pub/nihongo/JMdict_e.gz) needs to be downloaded and extracted to the Folder CLI root directory.

## Build

```
go build -buildmode=plugin
```

## Usage

At the Folder CLI root directory, run:
```
folder index --type jsonl --plugin jmdict JMdict_e
```
