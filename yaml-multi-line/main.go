package main

import (
	"fmt"
	"sigs.k8s.io/yaml"
)

type Query struct {
	Raw string `json:"raw,omitempty"`
}

func main() {
	q := Query{
		Raw: `query Find($src: String!, $targetGroup: String!, $targetKind: String!) {
  find(oid: $src) {
    backup_via(group: "stash.appscode.com", kind: "BackupConfiguration") {
      refs: offshoot(group: $targetGroup, kind: $targetKind) {
        namespace
        name
      }
    }
  }
}`,
	}

	data, err := yaml.Marshal(q)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
}
