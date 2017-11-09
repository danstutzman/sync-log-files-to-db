#!/bin/bash -ex

SSH_KEY=/Users/dan/.ssh/vultr
REMOTE_HOST=build.danstutzman.com
REMOTE_GOPATH=/root/gopath
REMOTE_DIR=$REMOTE_GOPATH/src/github.com/danielstutzman/sync-logs-from-s3
EXECUTABLE=sync-logs-from-s3

sudo ls # get sudo password at the beginning

if [ `which lsyncd` == "" ]; then
  brew install lsyncd
fi

ssh -i ~/.ssh/vultr root@$REMOTE_HOST <<EOF
sudo apt-get -y update && \
sudo apt-get -y install lsyncd golang
sudo mkdir -p $REMOTE_DIR
EOF

cat >rsync.sh <<EOF
#!/bin/bash -x
/usr/local/bin/rsync "\$@"
RSYNC_RESULT=$?
(
  if [ \$RSYNC_RESULT -eq 0 ]; then
    ssh -i $SSH_KEY root@$REMOTE_HOST <<EOF2
      set -ex
      rm -f $REMOTE_GOPATH/bin/$EXECUTABLE
      cd $REMOTE_DIR
      GOPATH=$REMOTE_GOPATH CGO_ENABLED=0 \
        go install -tags netgo -v -ldflags="-s -w" \
        ./...
      ldd $REMOTE_GOPATH/bin/$EXECUTABLE \
        | grep -q "not a dynamic executable"
EOF2
    osascript -e "display notification \"\" with title \"Build result \$?\""
  fi
)
exit \$RSYNC_RESULT
EOF
chmod +x rsync.sh

cat >lsyncd.conf.lua <<EOF
sync {
  default.rsyncssh,
  source = "$PWD",
  host = "root@$REMOTE_HOST",
  targetdir = "$REMOTE_DIR",
  delay = 1,
  exclude = { '.*.swp', '.*.swx' },
  rsync = {
    binary = "$PWD/rsync.sh",
    archive  = false,
    compress = true,
    rsh = "ssh -i $SSH_KEY -o StrictHostKeyChecking=no"
  }
}
EOF

./rsync.sh -tspozlgD --rsh="ssh -i $SSH_KEY -o StrictHostKeyChecking=no" -r --delete --force --from0 $PWD/ root@build.danstutzman.com:/root/gopath/src/github.com/danielstutzman/sync-logs-from-s3 --progress

sudo lsyncd -nodaemon -log Exec -logfile /dev/null lsyncd.conf.lua

echo Keep this window open to keep syncing.
