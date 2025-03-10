## App Flags

The application supports the following configuration flags:

### Main Server Configuration

| Flag | Environment Variable | Description | Default Value |
|------|---------------------|-------------|---------------|
| `--main-ep-addr` | `MAIN_ENTRYPOINT_ADDR` | Main entrypoint address | `:8080` |
| `--main-ep-keep-alive` | `MAIN_ENTRYPOINT_KEEP_ALIVE` | Main entrypoint keep alive | `0` |
| `--main-read-timeout` | `MAIN_READ_TIMEOUT` | Main read timeout | `30s` |
| `--main-read-header-timeout` | `MAIN_READ_HEADER_TIMEOUT` | Main read header timeout | `30s` |
| `--main-write-timeout` | `MAIN_WRITE_TIMEOUT` | Main write timeout | `30s` |
| `--main-idle-timeout` | `MAIN_IDLE_TIMEOUT` | Main idle timeout | `30s` |

### Health Check Server Configuration

| Flag | Environment Variable | Description | Default Value |
|------|---------------------|-------------|---------------|
| `--healthz-ep-addr` | `HEALTHZ_ENTRYPOINT_ADDR` | Health check entrypoint address | `:8081` |
| `--healthz-ep-keep-alive` | `HEALTHZ_ENTRYPOINT_KEEP_ALIVE` | Health check entrypoint keep alive | `0` |
| `--healthz-read-timeout` | `HEALTHZ_READ_TIMEOUT` | Health check read timeout | `30s` |
| `--healthz-read-header-timeout` | `HEALTHZ_READ_HEADER_TIMEOUT` | Health check read header timeout | `30s` |
| `--healthz-write-timeout` | `HEALTHZ_WRITE_TIMEOUT` | Health check write timeout | `30s` |
| `--healthz-idle-timeout` | `HEALTHZ_IDLE_TIMEOUT` | Health check idle timeout | `30s` |

All configuration flags can be set via command-line arguments, environment variables, or through a configuration file. The application uses [Viper](https://github.com/spf13/viper) for configuration management, which supports multiple configuration formats.
