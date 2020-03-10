# billing-client

A client library for sending usage data to the billing system.

Open sourced so it can be imported into our open-source projects.

## Usage

```Go
import billing "github.com/weaveworks/billing-client"

func init() {
  billing.MustRegisterMetrics()
}

func main() {
  var cfg billing.Config
  cfg.RegisterFlags(flag.CommandLine)
  flag.Parse()

  client, err := billing.NewClient(cfg)
  defer client.Close()

  err = client.AddAmounts(
    uniqueKey, // Unique hash of the data, or a uuid here for deduping
    internalInstanceID,
    timestamp,
    billing.Amounts{
      billing.ContainerSeconds: 1234,
    },
    map[string]string{
      "metadata": "goes here"
    },
  )
}

```

## <a name="help"></a>Getting Help

If you have any questions about, feedback for or problems with `billing-client`:

- Invite yourself to the <a href="https://slack.weave.works/" target="_blank">Weave Users Slack</a>.
- Ask a question on the [#general](https://weave-community.slack.com/messages/general/) slack channel.
- [File an issue](https://github.com/weaveworks/billing-client/issues/new).

Your feedback is always welcome!
