# gormzerolog

A simple GORM logger implementation for logging with zerolog.

## Basic usage

Initialize the gorm connection with the default logger configuration.

```go
import "github.com/lime008/gormzerolog"

conf := &gorm.Config{
	Logger: gormzerolog.Default,
}
db, err := gorm.Open(postgres.Open(dbConnetionString), conf)
```
