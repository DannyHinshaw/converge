# Converge

Converge is a CLI tool for merging multiple Go source files into a single file. It's useful for simplifying package
structures, easy sharing of Go code, and other scenarios where combining multiple Go files is desirable.

## Features

- Merges multiple Go source files from a specified directory into a single file.
- Supports exclusion of specific files.
- Optional timeout for merge operation.
- Configurable concurrency with a worker pool.

## Installation

To install Converge, you need to have Go installed on your machine. You can then install it using the following command:

```bash
go get github.com/dannyhinshaw/converge
```

## Usage

To use Converge, run the following command:

```bash
converge --src=<source-directory> --out=<output-file>
```

## Options

    --src: Source directory containing Go files.
    --out: Destination file for the merged output. Defaults to stdout if not specified.
    --exclude: Comma-separated list of filenames to exclude from merging.
    --timeout: Maximum time in seconds before cancelling the merge operation.
    --workers: Maximum number of concurrent workers.

## Example

Merging all Go files in the 'src' directory into 'merged.go':

```bash
converge --src=./src --out=./merged.go
```

## License

Converge is licensed under the MIT License.