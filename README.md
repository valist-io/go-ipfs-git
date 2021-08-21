# go-ipfs-git

IPFS plugin that adds a git remote to your node.

## Building

Go plugins only work in linux.

```bash
git clone https://github.com/valist-io/go-ipfs-git

cd go-ipfs-git

go build -buildmode=plugin -o=git.so ./plugin

mkdir -p ~/.ipfs/plugins

mv git.so ~/.ipfs/plugins/

chmod +x ~/.ipfs/plugins/git.so
```

## Contributing

Found a bug or have an idea for a feature? [Create an issue](https://github.com/valist-io/go-ipfs-git/issues/new).

## License

Valist is licensed under the [Mozilla Public License Version 2.0](https://www.mozilla.org/en-US/MPL/2.0/).
