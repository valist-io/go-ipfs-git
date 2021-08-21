# Plugin

IPFS plugin that adds a git support.

## Building

Go plugins only work in linux.

```bash
go build -buildmode=plugin -o=git.so

mkdir -p ~/.ipfs/plugins

mv git.so ~/.ipfs/plugins/

chmod +x ~/.ipfs/plugins/git.so
```
