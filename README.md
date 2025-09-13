# getaduck

`getaduck` - Get-a-Duck - is a CLI tool and Go library that downloads [DuckDB](https://duckdb.org/) releases. As a CLI tool,
`getaduck` is useful in scripts that automate provisioning DuckDB. As a Go library, it helps Go app access
DuckDB libraries without having to bundle a specific version.

## Usage as CLI

If Go 1.21+ is available, run:

```
go run github.com/murfffi/getaduck@latest -type cli
./duckdb --version
```

Use ```-help``` to see additional options

## Usage as a library

TODO
