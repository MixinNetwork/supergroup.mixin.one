const path = require('path');
const webpack = require('webpack');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const ExtractTextPlugin = require("extract-text-webpack-plugin");
const ScriptExtHtmlWebpackPlugin = require("script-ext-html-webpack-plugin");
const WebappWebpackPlugin = require('webapp-webpack-plugin');

const extractSass = new ExtractTextPlugin({
    filename: "[name]-[hash].css"
});

const webRoot = function (env) {
  if (env === 'production') {
    return 'https://chinese-group.mixin.zone';
  } else {
    return 'http://group-chat.mixin.test';
  }
};

const apiRoot = function (env) {
  if (env === 'production') {
    return 'https://group-chat-api.mixin.zone';
  } else {
    return 'http://localhost:7001';
  }
};

const clientId = function (env) {
  if (env === 'production') {
    return '67a87828-18f5-46a1-b6cc-c72a97a77c43';
  } else {
    return '5fcd897e-e6b2-40d5-93cd-487e2d95d556';
  }
};

module.exports = {
  entry: {
    app: './src/app.js'
  },

  output: {
    publicPath: '/assets/',
    path: path.resolve(__dirname, 'dist'),
    filename: '[name]-[chunkHash].js'
  },

  resolve: {
    alias: {
      jquery: "jquery/dist/jquery",
      handlebars: "handlebars/dist/handlebars.runtime"
    }
  },

  module: {
    rules: [{
      test: /\.html$/, loader: "handlebars-loader?helperDirs[]=" + __dirname + "/src/helpers"
    }, {
      test: /\.(scss|css)$/,
      use: extractSass.extract({
        use: [{
          loader: "css-loader"
        }, {
          loader: "sass-loader"
        }],
        fallback: "style-loader"
      })
    }, {
      test: /\.(woff|woff2|eot|ttf|otf|svg)$/,
      use: [
        'file-loader'
      ]
    }, {
      test: /\.(png|svg|jpg|gif)$/,
      use: [
        'file-loader'
      ]
    }]
  },

  plugins: [
    new webpack.DefinePlugin({
      PRODUCTION: (process.env.NODE_ENV === 'production'),
      WEB_ROOT: JSON.stringify(webRoot(process.env.NODE_ENV)),
      API_ROOT: JSON.stringify(apiRoot(process.env.NODE_ENV)),
      CLIENT_ID: JSON.stringify(clientId(process.env.NODE_ENV)),
      APP_NAME: JSON.stringify('Mixin 中文群'),
      LOCALE: JSON.stringify('zh-CN')
    }),
    new HtmlWebpackPlugin({
      template: './src/layout.html'
    }),
    new WebappWebpackPlugin({
      logo: './src/launcher.png',
      prefix: 'icons-[hash]-',
      background: '#FFFFFF'
    }),
    new ScriptExtHtmlWebpackPlugin({
      defaultAttribute: 'async'
    }),
    extractSass
  ]
};
