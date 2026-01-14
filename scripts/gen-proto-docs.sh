#!/usr/bin/env bash
set -euo pipefail

proto_root="${1:-api}"
out_dir="${2:-docs/proto}"
out_file="${3:-api.md}"

mkdir -p "$out_dir"

# Use the current user's UID/GID so generated files are not owned by root.
uid_gid="$(id -u):$(id -g)"

# proto_root is mounted to /protos. In this repository, hello.proto lives at api/v1/hello.proto,
# so inside the container it becomes /protos/v1/hello.proto.
docker run --rm \
  --user "$uid_gid" \
  -v "$PWD/$proto_root:/protos" \
  -v "$PWD/$out_dir:/out" \
  --entrypoint protoc \
  pseudomuto/protoc-gen-doc \
  -I /protos \
  -I /usr/include \
  --doc_out=/out \
  --doc_opt=markdown,"$out_file" \
  v1/hello.proto

echo "Generated: $out_dir/$out_file"
