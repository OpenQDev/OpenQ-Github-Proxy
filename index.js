const express = require('express');
const { createProxyMiddleware, fixRequestBody } = require('http-proxy-middleware');
const dotenv = require('dotenv');
dotenv.config();

const app = express();

let patsArray = process.env.PATS.split(',');

async function onProxyReq(proxyReq, req, res) {
	let token = patsArray[Math.floor(Math.random() * patsArray.length)];
	proxyReq.setHeader('Authorization', `Bearer ${token}`);
	fixRequestBody(proxyReq, req)
}

const mapping = {}

app.use(express.json())

app.use('/', createProxyMiddleware({
	target: 'https://api.github.com/graphql',
	pathRewrite: { '^/': '' },
	changeOrigin: true,
	onProxyReq
}));

app.listen(3000, () => {
	console.log('Server started on port 3000')
})

// app.listen with start message