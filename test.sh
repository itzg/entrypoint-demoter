#!/bin/sh

set -e

assert() {
  name="$1"
  expected="$2"
  got="$3"

  if [ "$got" != "$expected" ]; then
    echo "FAILED $name: expected '$expected', got '$got'" >&2
    exit 1
  else
    echo "PASSED $name"
  fi
}

docker build -f Dockerfile.test -t entrypoint-demoter-test .

assert "default" "uid=0(root) gid=0(root) groups=0(root)" "$(docker run --rm  entrypoint-demoter-test id)"
docker run --rm -u 1000 entrypoint-demoter-test id
assert "run 1000" "uid=1000 gid=0(root) groups=0(root)" "$(docker run --rm -u 1000 entrypoint-demoter-test id)"
assert "run 1000:1000" "uid=1000 gid=1000 groups=1000" "$(docker run --rm -u 1000:1000 entrypoint-demoter-test id)"
assert "match" "uid=1024(test1024) gid=100(users) groups=100(users)" "$(docker run --rm entrypoint-demoter-test --match /test/test1024 id)"
assert "command args" "/test/test1024" "$(docker run --rm entrypoint-demoter-test --match /test/test1024 ls -d /test/test1024)"
assert "UID" "uid=1000 gid=0(root) groups=0(root)" "$(docker run --rm -e UID=1000 entrypoint-demoter-test id)"
assert "GID" "uid=0(root) gid=1000 groups=1000" "$(docker run --rm -e GID=1000 entrypoint-demoter-test id)"
assert "UID+GID" "uid=1000 gid=1000 groups=1000" "$(docker run --rm -e UID=1000 -e GID=1000 entrypoint-demoter-test id)"

id=$(docker run -d entrypoint-demoter-test --stdin-on-term "hello" cat -)
# shellcheck disable=SC2086
docker stop ${id}
# shellcheck disable=SC2086
assert "stdin-on-term" "hello" "$(docker logs ${id})"
# shellcheck disable=SC2086
docker rm ${id}

echo "ALL PASSED"
