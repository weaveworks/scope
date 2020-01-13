# billing-client

A client library for sending usage data to the billing system.

Open sourced so it can be imported into our open-source projects.

## Usage

`dep ensure github.com/weaveworks/billing-client`

then

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

- Invite yourself to the <a href="https://weaveworks.github.io/community-slack/" target="_blank"> #weave-community </a> slack channel.
- Ask a question on the <a href="https://weave-community.slack.com/messages/general/"> #weave-community</a> slack channel.
- Send an email to <a href="mailto:weave-users@weave.works">weave-users@weave.works</a>
- <a href="https://github.com/weaveworks/billing-client/issues/new">File an issue.</a>

Your feedback is always welcome!
