This is the front-end of Mixin group, you can fork it and modify it to your own. If there's a problem, appreciate submitting an issue or PR.

## Preparation

Before you start it, you should `cp env.example env.local`, you can find more document in https://cli.vuejs.org/guide/mode-and-env.html

## How to deploy

Deploy the project, you need a web server like NGINX (or something else), then build the codes and sync to the remote server.

```
#!/bin/sh

rm -r dist/*
npm run build || exit

rsync -rcv dist/* group:/var/www/html/
```
