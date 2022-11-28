## How to release

1. Commit all changes.

2. Add `tag` to HEAD.

For example,

```
$ git tag v0.1.0
```

You can delete tag by below way.

```
$ git tag -d v0.1.0
$ git push origin :v0.1.0
```

3. Release using `goreleaser`

```
$ goreleaser release --rm-dist
```
