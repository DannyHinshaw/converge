![Build](https://github.com/dannyhinshaw/converge/actions/workflows/go-build.yaml/badge.svg)
![Lint](https://github.com/dannyhinshaw/converge/actions/workflows/go-lint.yaml/badge.svg)
![Test](https://github.com/dannyhinshaw/converge/actions/workflows/go-test.yaml/badge.svg)

# Converge

Converge is a CLI tool designed for merging multiple Go source files into a single file. It provides a streamlined
approach for consolidating Go package structures, simplifying code sharing, and other scenarios where amalgamating Go
files is beneficial.

## Features

- Efficiently merges multiple Go source files from a specified directory into a single consolidated file.
- Allows exclusion of specific files from the merging process.
- Supports an optional timeout setting for the merge operation.

## Installation

To install Converge, ensure Go is installed on your machine.

```bash
which go
```

You can install Converge using the following command:

```bash
go install github.com/dannyhinshaw/converge
```

## Usage

Reference the `help` command for detailed usage instructions:

```bash
converge --help
```

## Example

To merge all Go files in the 'src' directory into 'merged.go':

```bash
converge --dir=./src --output=./merged.go
```

To merge all Go files in the 'src' directory and pipe to clipboard:

```bash
converge --dir=./src | pbcopy
```

## License

Converge is licensed under the MIT License.
