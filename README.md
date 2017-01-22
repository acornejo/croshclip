# croshclip

Simple clipboard utility for the chrome OS shell. It requires installing
the crouton [chrome extension](https://chrome.google.com/webstore/detail/crouton-integration/gcpneefbbnfalgjniomfjknbcgkbijom).

# Build from source

```
go get github.com/acornejo/croshclip
go build github.com/acornejo/croshclip
```

A statically linked x64 binary of `croshclip` is hosted in github.

# Installation

First, install the crouton chrome extension (link above), then copy the
`croshclip` binary to your path and make sure to run the `croshclip`
server inside your shell. Example:

```
cp croshclip /usr/bin
echo 'nc -z localhost 30001 || croshclip -serve > /tmp/croshclip.log 2>&1 &' > /etc/profile.d/croshclip.sh
chmod 755 /etc/profile.d/croshclip.sh
```

You will have to restart your shell to make the changes above take
effect.

Now you can use `croshclip` as you would use `xclip`, i.e.

`
echo hello world | croshclip -copy
croshclip -paste > somefile
`

# Vim integration

Here is a very bare-bones integration of croshclip and vim:


```
nnoremap "*p :r !croshclip -paste<CR>
vnoremap "*y :w !croshclip -copy<CR><CR>

nnoremap "+p :r !croshclip -paste<CR>
vnoremap "+y :w !croshclip -copy<CR><CR>
```
