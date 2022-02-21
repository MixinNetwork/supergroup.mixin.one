supergroup.mixin.one is the source code of "Mixin 中文群"，which is an unlimited member group base on Mixin bot.

## NOTICE !!!

Before you upgrade your group, please checkout [CHANGELOG.md](https://github.com/MixinNetwork/supergroup.mixin.one/blob/master/CHANGELOG.md) first. 

## Prepare

1. copy `./config/config.tpl.yaml` to `./config/config.yaml`
2. Replace configurations in `config.yaml`, we use PostgreSQL as our database.
3. `cd client` and exec `yarn install`
4. `cp env.example .env.local`, you can find more document in https://cli.vuejs.org/guide/mode-and-env.html
5. Replace configs in `.env.local`

## Run

#### Server Side

1. `./supergroup.mixin.one` handle http request
2. `./supergroup.mixin.one -service message` handle messages

#### Front-end

Generate static assets `cd client && yarn serve`
