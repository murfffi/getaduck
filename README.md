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
package main

import (
	"fmt"

	"github.com/murfffi/getaduck/download"
)

func main() {
	// Download the latest DuckDB release for your platform
	res, err := download.Do(download.DefaultSpec())
	if err != nil {
		panic(err)
	}
	fmt.Println("Downloaded DuckDB to:", res.OutputFile)
}
```

## Contributing

Contributions are welcome! Please fork the repository and open a pull request with your proposed changes. Make sure your code follows Go best practices. Bug reports and feature requests are also appreciated.
