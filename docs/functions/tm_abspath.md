---
title: tm_abspath - Functions - Configuration Language
description: The tm_abspath function converts the argument to an absolute filesystem path.
---

# `tm_abspath` Function

`tm_abspath` takes a string containing a filesystem path and converts it
to an absolute path. That is, if the path is not absolute, it will be joined
with the current working directory.

Referring directly to filesystem paths in resource arguments may cause
spurious diffs if the same configuration is applied from multiple systems or on
different host operating systems. We recommend using filesystem paths only
for transient values, such as the argument to [`tm_file`](./tm_file.md) (where
only the contents are then stored) or in `connection` and `provisioner` blocks.

## Examples

```
tm_abspath("stacks/vpc")
/home/user/repo/stacks/vpc
```
