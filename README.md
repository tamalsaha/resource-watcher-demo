# resource-watcher-demo

```
query Find($srcKey: String!, $dstGroup: String, $dstKind: String){
  find(key: $srcKey) {
    offshoot(group:$dstGroup, kind:$dstKind) {
      namespace
      name
    }
  }
}

{
  "srcKey":  "G=apps,K=Deployment,NS=kube-system,N=coredns",
  "dstGroup": "",
  "dstKind": "Pod"
}
```