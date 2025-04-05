# Configuration

Application configuration is set via environment variables. The `App` object can be used to access the configuration using the following methods:

- `HasEnvVar(key string) bool` returns true if the environment variable is set
- `GetEnvVarString(key string) string` returns the value of the environment variable as a string. If the environment variable is not set, the application will panic.
- `GetEnvVarInt(key string) int` returns the value of the environment variable as an integer. If the environment variable is not set, the application will panic.

I use [direnv](https://direnv.net/) to set application environment variables. The documented list of all environment variables is available in the [.envrc.template](https://github.com/gurch101/gowebutils/blob/main/.envrc.template) file.
