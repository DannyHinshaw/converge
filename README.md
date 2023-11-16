![Build](https://github.com/dannyhinshaw/converge/actions/workflows/build.yaml/badge.svg)
![Lint](https://github.com/dannyhinshaw/converge/actions/workflows/lint.yaml/badge.svg)
![Test](https://github.com/dannyhinshaw/converge/actions/workflows/test.yaml/badge.svg)

# Converge

Converge is a CLI tool designed for merging multiple Go source files into a single file. It provides a streamlined
approach for consolidating Go package structures, simplifying code sharing, and other scenarios where amalgamating Go
files is beneficial.

## Features

- Efficiently merges multiple Go source files from a specified directory into a single consolidated file.
- Allows exclusion of specific files from the merging process.
- Supports an optional timeout setting for the merge operation.
- Employs a configurable worker pool for enhanced concurrency management.

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

Run Converge with the command below, specifying the source directory
and (optionally) the output file:

```bash
converge <source-directory> --file=<output-file>
```

## Options

    -f, --file          Path to the output file where the merged content will be written;
                        defaults to stdout if not specified.
    -v                  Enable verbose logging for debugging purposes.
    -h, --help          Show this help message and exit.
    --version           Show version information.
    -t, --timeout       Maximum time (in seconds) before cancelling the merge operation;
                        if not specified, the command runs until completion.
    -w, --workers       Maximum number of concurrent workers in the worker pool.
    -e, --exclude       Comma-separated list of filenames to exclude from merging.

## Example

To merge all Go files in the 'src' directory into 'merged.go':

```bash
converge ./src --file=./merged.go
```

To merge all Go files in the 'src' directory and pipe to clipboard:

```bash
converge ./src | pbcopy
```

For more detailed usage instructions, refer to the tool's help message:

```bash
converge --help
```

## License

Converge is licensed under the MIT License.
