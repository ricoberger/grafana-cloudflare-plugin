# Grafana Cloudflare Plugin

The Grafana Cloudflare Plugin allows you to explore your Cloudflare metrics
within Grafana.

## Installation

1. Before you can install the plugin, you have to add
   `ricoberger-cloudflare-datasource` to the
   [`allow_loading_unsigned_plugins`](https://grafana.com/docs/grafana/latest/setup-grafana/configure-grafana/#allow_loading_unsigned_plugins)
   configuration option or to the `GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS`
   environment variable.
2. The plugin can then be installed by adding
   `ricoberger-cloudflare-datasource@<VERSION>@https://github.com/ricoberger/grafana-cloudflare-plugin/releases/download/v<VERSION>/ricoberger-cloudflare-datasource-<VERSION>.zip`
   to the
   [`preinstall_sync`](https://grafana.com/docs/grafana/latest/setup-grafana/configure-grafana/#preinstall_sync)
   configuration option or the `GF_PLUGINS_PREINSTALL_SYNC` environment
   variable.

### Configuration File

```ini
[plugins]
allow_loading_unsigned_plugins = ricoberger-cloudflare-datasource
preinstall_sync = ricoberger-cloudflare-datasource@0.1.0@https://github.com/ricoberger/grafana-cloudflare-plugin/releases/download/v0.1.0/ricoberger-cloudflare-datasource-0.1.0.zip
```

### Environment Variables

```bash
export GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=ricoberger-cloudflare-datasource
export GF_PLUGINS_PREINSTALL_SYNC=ricoberger-cloudflare-datasource@0.1.0@https://github.com/ricoberger/grafana-cloudflare-plugin/releases/download/v0.1.0/ricoberger-cloudflare-datasource-0.1.0.zip
```

## Contributing

If you want to contribute to the project, please read through the
[contribution guideline](https://github.com/ricoberger/grafana-cloudflare-plugin/blob/main/CONTRIBUTING.md).
Please also follow our
[code of conduct](https://github.com/ricoberger/grafana-cloudflare-plugin/blob/main/CODE_OF_CONDUCT.md)
in all your interactions with the project.
