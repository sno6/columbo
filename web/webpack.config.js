var path = require('path');
var webpack = require('webpack');
var ExtractTextPlugin = require('extract-text-webpack-plugin');

module.exports = {
	entry: './public/index.jsx',
	output: {
		filename: 'bundle.js',
		path: path.resolve(__dirname, 'public/bundle/')
	},
	module: {
		loaders: [
			{
				test: /\.jsx$/,
				include: path.resolve(__dirname, 'public/'),
				loader: 'babel-loader',
			},
			{
				test: /\.scss$/,
				loader: ExtractTextPlugin.extract('css-loader!sass-loader')
			},
			{
				test: /\.css$/,
				loader: ExtractTextPlugin.extract('css-loader!sass-loader')
			}
		]
	},
	devServer: {
		contentBase: 'public/bundle',
		compress: true,
		port: 8080
	},
	plugins: [
		new webpack.ProvidePlugin({
			$: "jquery",
			jQuery: "jquery"
		}),
		new ExtractTextPlugin('./styles/style.css', {
			allChunks: true
		}),
	]
};
