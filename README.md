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

```go
import (
    "github.com/murfffi/getaduck"
    "fmt"
)

func main() {
    // Download the latest DuckDB release for your platform
    duckdbPath, err := getaduck.Download(getaduck.Options{})
    if err != nil {
        panic(err)
    }
    fmt.Println("Downloaded DuckDB to:", duckdbPath)
}
```

## Contributing

Contributions are welcome! Please fork the repository and open a pull request with your proposed changes. Make sure your code follows Go best practices. Bug reports and feature requests are also appreciated.
