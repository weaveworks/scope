package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"text/template"

	socks5 "github.com/armon/go-socks5"
	"github.com/weaveworks/docker/pkg/mflag"
	"github.com/weaveworks/weave/common/mflagext"
	"golang.org/x/net/context"
)

type pacFileParameters struct {
	HostMatch string
	Aliases   map[string]string
}

const (
	pacfile = `
function FindProxyForURL(url, host) {
	if(shExpMatch(host, "{{.HostMatch}}")) {
		return "SOCKS5 localhost:8000";
	}
	{{range $key, $value := .Aliases}}
	if (host == "{{$key}}") {
		return "SOCKS5 localhost:8000";
	}
	{{end}}
	return "DIRECT";
}
`
)

func main() {
	var (
		as        []string
		hostMatch string
	)
	mflagext.ListVar(&as, []string{"a", "-alias"}, []string{}, "Specify hostname aliases in the form alias:hostname.  Can be repeated.")
	mflag.StringVar(&hostMatch, []string{"h", "-host-match"}, "*.weave.local", "Specify main host shExpMatch expression in pacfile")
	mflag.Parse()

	var aliases = map[string]string{}
	for _, a := range as {
		parts := strings.SplitN(a, ":", 2)
		if len(parts) != 2 {
			fmt.Printf("'%s' is not a valid alias.\n", a)
			mflag.Usage()
			os.Exit(1)
		}
		aliases[parts[0]] = parts[1]
	}

	go socksProxy(aliases)

	t := template.Must(template.New("pacfile").Parse(pacfile))
	http.HandleFunc("/proxy.pac", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-ns-proxy-autoconfig")
		t.Execute(w, pacFileParameters{hostMatch, aliases})
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

type aliasingResolver struct {
	aliases map[string]string
	socks5.NameResolver
}

func (r aliasingResolver) Resolve(ctx context.Context, name string) (context.Context, net.IP, error) {
	if alias, ok := r.aliases[name]; ok {
		return r.NameResolver.Resolve(ctx, alias)
	}
	return r.NameResolver.Resolve(ctx, name)
}

func socksProxy(aliases map[string]string) {
	conf := &socks5.Config{
		Resolver: aliasingResolver{
			aliases:      aliases,
			NameResolver: socks5.DNSResolver{},
		},
	}
	server, err := socks5.New(conf)
	if err != nil {
		panic(err)
	}
	if err := server.ListenAndServe("tcp", ":8000"); err != nil {
		panic(err)
	}
}
