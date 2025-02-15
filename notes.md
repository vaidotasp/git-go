# write-tree cmd notes

Program will create random files and dirs and program will output a hash of the tree

```js
each “line” in a tree object is “mode space path” as text, then a NUL byte, then the binary SHA-1 hash

  100644 blob 4aab5f560862b45d7a9f1370b1c163b74484a24d    LICENSE.txt
  100644 blob 43ab992ed09fa756c56ff162d5fe303003b5ae0f    README.md
  100644 blob c10cb8bc2c114aba5a1cb20dea4c1597e5a3c193    pygit.py
```

it is basically a snapshot of the current state of the repository
