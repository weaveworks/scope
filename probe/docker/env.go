package docker

import (
	"strings"

	"github.com/weaveworks/scope/report"
)

// EnvPrefix is the key prefix used for Docker environment variables in Node
// (e.g. "TERM=vt200" will get encoded as "docker_env_TERM"="vt200" in the
// metadata)
const EnvPrefix = "docker_env_"

// AddEnv appends Docker environment variables to the Node from a topology.
func AddEnv(node report.Node, env []string) report.Node {
	node = node.Copy()
	for _, value := range env {
		v := strings.SplitN(value, "=", 2)
		if len(v) == 2 {
			key, value := v[0], v[1]
			node = node.WithLatests(map[string]string{
				EnvPrefix + key: value,
			})
		}
	}
	return node
}

// ExtractEnv returns the list of Docker environment variables given a Node from a topology.
func ExtractEnv(node report.Node) map[string]string {
	result := map[string]string{}
	node.Latest.ForEach(func(key, value string) {
		if strings.HasPrefix(key, EnvPrefix) {
			env := key[len(EnvPrefix):]
			result[env] = value
		}
	})
	return result
}
