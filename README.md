# ghtopdep

`ghtopdep` is a CLI tool that sorts GitHub repository dependents by the number of stars. This tool can be useful for developers who want to analyze the popularity of repositories that depend on a specific repository.

## Features

- Sort dependents repositories or packages by the number of stars.
- Output results in a table format or as JSON.
- Filter dependents based on a minimum number of stars.
- Limit the number of dependents displayed.

## Installation

To install `ghtopdep`, ensure you have Go installed on your machine, then run:

```sh
go get github.com/yourusername/ghtopdep
```

## Usage

```sh
ghtopdep [flags] URL
```

## Flags

- **packages**: Sort dependents packages instead of repositories.
- **json**: Output the results as JSON.
- **rows**: Number of repositories to display (default is 10).
- **minstar**: Minimum number of stars for the dependents (default is 5).

## Examples

```sh
ghtopdep --minstar 50 --rows 10 https://github.com/yourusername/yourrepository
```

## Build from Source

To build ghtopdep from source, clone the repository and build the binary:

```sh
git clone https://github.com/yourusername/ghtopdep
cd ghtopdep
go build -o ghtopdep

./ghtopdep [flags] URL
```

## Contributing

If you find any issues or have suggestions for improvements, please open an issue or submit a pull request on GitHub.

## License

This project is licensed under the MIT License. See the ./LICENSE file for details.

Enjoy using **ghtopdep**!
