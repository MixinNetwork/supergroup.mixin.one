## Mixin Super Group

supergroup.mixin.one is the source code of "Mixin 中文群"，which is an unlimited member group base on Mixin bot.

## Prepare

1. copy `config.tpl.yaml` to `config.yaml`
2. Replace configurations in `config.go`, we use PostgreSQL as our database.
3. `cd web` and exec `npm install`
4. Replace domain, api address and clientId in `webpack.config.js`

## Run

#### Server Side

1. `./supergroup.mixin.one` handle http request
2. `./supergroup.mixin.one -service message` handle messages

#### Front-end

Generate static assets `cd web && npm run build`
