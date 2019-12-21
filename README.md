# 
# API Server

## Install

### Requires
- go
- git
- make
- golint

```sh
mkdir -p $GOPATH/src/github.com/rajendraventurit
cd $GOPATH/src/github.com/rajendraventurit
git clone https://github.com/rajendraventurit/radicaapi.git
cd radicaapi
make

# config
mkdir /etc/radica
cp ./assets/config/local/* /etc/radica/.
# For log
touch /var/log/radica.log
```
