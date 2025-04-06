# Configuration

gowebutils uses environment variables for application configuration. The App object provides several methods to access these configuration values safely and consistently.

### Checking for Environment Variables

```go
// Returns true if the environment variable is set
if app.HasEnvVar("DATABASE_URL") {
  // do stuff
}
```

### Retrieving String Values

```go
// Returns the value of the environment variable as a string
// Panics if the variable is not set
connectionString := app.GetEnvVarString("DATABASE_URL")
```

### Retrieving Integer Values

```go
// Returns the value of the environment variable as an integer
// Panics if the variable is not set or cannot be converted to an integer
maxConnections := app.GetEnvVarInt("MAX_CONNECTIONS")
```

[direnv](https://direnv.net/) is recommended to manage your application's environment variables. We recommend using direnv to manage your application's environment variables. This tool automatically loads environment variables from a `.envrc` file when you enter a directory, making it easy to maintain different configurations for different projects.

### Available Environment Variables

For a complete list of supported environment variables and their descriptions, refer to the [.envrc.template](https://github.com/gurch101/gowebutils/blob/main/.envrc.template) file in the repository.
