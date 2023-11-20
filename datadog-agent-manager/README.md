# Datadog Agent for Home Assistant

This add-on allows you to run the Datadog Agent on your Home Assistant instance.

**NOTE**: Make sure to turn off `Protection mode` in the addon's settings before starting the addon. It will not work otherwise.

## Configuration

The only required configuration is a Datadog `API key`. You can find it in your [Datadog account settings](https://app.datadoghq.com/organization-settings/api-keys).

Make sure to also set the appropriate Datadog `Site` depending on where your account lives. The addon defaults to US1. See the possible site parameters (here)(https://docs.datadoghq.com/getting_started/site/.
