const path = require('path');
const webpack = require('webpack');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const ExtractTextPlugin = require("extract-text-webpack-plugin");
const ScriptExtHtmlWebpackPlugin = require("script-ext-html-webpack-plugin");
const WebappWebpackPlugin = require('webapp-webpack-plugin');

const extractSass = new ExtractTextPlugin({
    filename: "[name]-[hash].css"
});

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
      WEB_ROOT: JSON.stringify(process.env.SUPERGROUP_WEB_ROOT),
      API_ROOT: JSON.stringify(process.env.SUPERGROUP_API_ROOT),
      CLIENT_ID: JSON.stringify(process.env.SUPERGROUP_CLIENT_ID),
      APP_NAME: JSON.stringify(process.env.SUPERGROUP_APP_NAME),
      LOCALE: JSON.stringify(process.env.SUPERGROUP_LOCALE)
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
