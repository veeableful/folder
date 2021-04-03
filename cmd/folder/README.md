# Folder CLI

This is a command-line tool created mainly for testing the library. It supports indexing plain-text files, JSON files, and JSONL files.

## Installation

```
go install
```

## Usage

### Indexing

For indexing raw text files:
```
folder index [file / directory]
```

For indexing JSON files:
```
folder index --type json [file / directory]
```

For indexing JSONL files:
```
folder index --type jsonl [file / directory]
```

You can also develop a plugin that provides data. For an example, see `plugins/jmdict`. After that you can run:
```
folder index --type [type] --plugin [plugin name] [optional arguments]
```

### Searching

```
folder search [query]
```