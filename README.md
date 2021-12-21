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


```
query Find($src: String!){
  find(oid: $src) {
    refs: offshoot(group:"", kind:"Pod") {
      namespace
      name
    }
  }
}

# variables
{
  "src":  "G=apps,K=Deployment,NS=kube-system,N=coredns",
  "dstGroup": "",
  "dstKind": "Pod"
}


# result
{
  "data": {
    "find": {
      "refs": [
        {
          "name": "coredns-64897985d-4s8fh",
          "namespace": "kube-system"
        },
        {
          "name": "coredns-64897985d-rpjmr",
          "namespace": "kube-system"
        }
      ]
    }
  }
}
```