#!/bin/bash -ex

USERNAME=root
HOSTNAME=207.246.91.97
TARGET_PATH=/root/gopath/src/github.com/danielstutzman/sync-log-files-to-db/
SSH_KEY=/Users/dan/.ssh/vultr

ssh -i $SSH_KEY $USERNAME@$HOSTNAME "cd $TARGET_PATH && git config user.email dtstutz@gmail.com && git config user.name 'Dan Stutzman' && git reset --hard && git clean -f -d && GIT_SSH_COMMAND=\"ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no\" git pull origin master"

git status --porcelain --ignored \
  | sed 's/^.. //' \
  | xargs -J {} \
    rsync -r --rsh="ssh -i $SSH_KEY" -v \
     --exclude=backed_up --exclude=.git -z --delete \
     --no-owner --force --relative --delete-missing-args \
      .git/ {} $USERNAME@$HOSTNAME:$TARGET_PATH || \
  if [ "$?" != 0 ]; then
    if [ $(ssh -i $SSH_KEY $USERNAME@$HOSTNAME "[ -e $TARGET_PATH ]; echo \$?") == 1 ]; then
      # If it's the first time, do some setup and full rsync
      ssh -i $SSH_KEY $USERNAME@$HOSTNAME <<EOF
        set -ex
        mkdir -p $TARGET_PATH

        if [ ! -e go ]; then
          curl -o go1.9.2.linux-amd64.tar.gz \
            https://storage.googleapis.com/golang/go1.9.2.linux-amd64.tar.gz
          tar xzf go1.9.2.linux-amd64.tar.gz
          rm go1.9.2.linux-amd64.tar.gz
        fi
EOF
      rsync -r --rsh="ssh -i $SSH_KEY" -z --progress --exclude=backed_up \
        --no-owner --force --relative --delete-missing-args ./ \
        $USERNAME@$HOSTNAME:$TARGET_PATH
    else
      exit 1
    fi
  fi

ssh -i $SSH_KEY $USERNAME@$HOSTNAME <<EOF
  set -ex

  cd $TARGET_PATH
  GOPATH=/root/gopath CGO_ENABLED=0 \
    /root/go/bin/go install -tags netgo -v -ldflags="-s -w" ./...

  ldd /root/gopath/bin/sync-log-files-to-db | grep -q "not a dynamic executable"
  git diff
EOF
