last_tag=`curl -s -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/repos/upfluence/gh-downloader/releases | jq '.[0].tag_name' | tr -d \"`

curl -sL https://github.com/upfluence/gh-downloader/releases/download/$last_tag/gh-downloader-linux-amd64 \
  > ~/bin/gh-downloader
chmod +x ~/bin/gh-downloader

gh-downloader -a bin-go.tar.gz -o bin.tar.gz -repository upfluence/upfluence-if -s bin-vx.x.x

tar -C ~/bin -xvf bin.tar.gz if-go
chmod +x ~/bin/if-go

if-go base

git clean -fd
