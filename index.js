const express = require('express');
const { createProxyMiddleware } = require('http-proxy-middleware');
const dotenv = require('dotenv');
dotenv.config();

const app = express();

let patsArray = process.env.PATS.split(',');

function onProxyReq(proxyReq, req, res) {
	let token = patsArray[Math.floor(Math.random() * patsArray.length)];
	proxyReq.setHeader('Authorization', `Bearer ${token}`);
}

app.use('/', createProxyMiddleware({
	target: 'https://api.github.com/graphql',
	pathRewrite: { '^/': '' },
	changeOrigin: true,
	onProxyReq
}));

app.listen(3000);