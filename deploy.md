如何部署 Mixin 大群，项目主要为分两部分：
1. 后端代码，可以打包成一个二进制文件
2. 前端代码，在 ./web 下面

## 准备工作
1. 一台服务器，需要根据用户量，初期 2C4G 的足够用，golang 本身占内存很少。
2. 需要安装 postgrsql

## 后端部署
1. git pull https://github.com/MixinNetwork/supergroup.mixin.one.git 相关的代码
2. 把 ./config/config.tpl.yaml 复制到 ./config/config.yaml, 这个文件会直接打包项目里， 其中 mixin 的内容需要到 https://developers.mixin.one/dashboard 来生成
3. 导入数据库，可以直接通过 pg 命令导入，想着文件在 https://github.com/MixinNetwork/supergroup.mixin.one/blob/master/models/schema.sql
4. 以上准备完，可以用打包完的文件在本地执行测试，有两个服务，一个是 api， 另一个是消息服务，分别通过下面两个命令开启
    a. ./supergroup.mixin.one -service http -e development
    b. ./supergroup.mixin.one -service message -e development

## 前端部署
1. 前端是用 vue 实现，需要把 env.example，复制成 .env.local (本地)，.env.production.local （生产环境）
2. 启动命令 yarn serve, 可以在 `package.json` 中查看

## 注意事项

后期所有的修改都会放到 [CHANGELOG.md](https://github.com/MixinNetwork/supergroup.mixin.one/blob/master/CHANGELOG.md) 里

生产环境推荐用 systemd 来管理线程 [相关示例](https://github.com/MixinNetwork/supergroup.mixin.one/tree/master/config)
