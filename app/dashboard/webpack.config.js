const path = require('path');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');

module.exports = (env, argv) => {
  const isProduction = argv.mode === 'production';

  return {
    mode: isProduction ? 'production' : 'development',
    entry: './src/index.js',
    output: {
      path: path.resolve(__dirname, 'build'),
      filename: 'js/[name].[contenthash:8].js',
      publicPath: '/',
      clean: true, // Clean the output directory before emit
    },
    module: {
      rules: [
        {
          test: /\.jsx?$/,
          exclude: /node_modules/,
          use: {
            loader: 'babel-loader',
            options: {
              presets: ['@babel/preset-env', '@babel/preset-react']
            }
          }
        },
        {
          test: /\.css$/,
          use: [
            isProduction ? MiniCssExtractPlugin.loader : 'style-loader',
            'css-loader',
            'postcss-loader'
          ]
        }
      ]
    },
    plugins: [
      new HtmlWebpackPlugin({
        template: './public/index.html',
        inject: true,
        filename: 'index.html'
      }),
      new MiniCssExtractPlugin({
        filename: 'css/[name].[contenthash:8].css'
      })
    ],
    resolve: {
      extensions: ['.js', '.jsx']
    },
    optimization: {
      runtimeChunk: 'single',
      splitChunks: {
        chunks: 'all',
      }
    }
  };
};