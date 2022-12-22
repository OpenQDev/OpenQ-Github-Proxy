const express = require('express');
const { createProxyMiddleware, fixRequestBody, responseInterceptor } = require('http-proxy-middleware');
const redis = require('redis');
const dotenv = require('dotenv');
const { promisify } = require("util");

dotenv.config();

const app = express();

let patsArray = process.env.PATS.split(',');

const client = redis.createClient({
  host: process.env.REDIS_HOST,
  port: parseInt(process.env.REDIS_PORT)
});

const getAsync = promisify(client.get).bind(client);
const setAsync = promisify(client.set).bind(client);

const mapping = {}

app.use(express.json())
app.use('/', createProxyMiddleware({
	target: 'https://api.github.com/graphql',
	pathRewrite: { '^/': '' },
	selfHandleResponse: true,
	changeOrigin: true,
	onProxyReq: async (proxyReq, req, res) => {
		// only set header if not present by OAuth
		let token = patsArray[Math.floor(Math.random() * patsArray.length)];
		proxyReq.setHeader('Authorization', `Bearer ${token}`);
	
		// combine the query and variables to get a unique key
		const key = `${JSON.stringify(req.body.query)}${JSON.stringify(req.body.variables)}`

		// const response = await getAsync(key)
		const response = mapping[key]

		// response will be null if there's an error or cache miss
		if (response != undefined) {
			return res.json(JSON.parse(response));
		}
		
		// this method provided by http-proxy-middleware fixes the body after bodyParser has it's way with it
		fixRequestBody(proxyReq, req)
	},
	onProxyRes: responseInterceptor(async (responseBuffer, proxyRes, req, res) => {
		const key = `${JSON.stringify(req.body.query)}${JSON.stringify(req.body.variables)}`
    const response = responseBuffer.toString('utf8');
  
		// const hour = 60 * 60
		// await setAsync(key, response, 'EX', hour)

		mapping[key] = response
		
		return response
  })
}));

app.listen(process.env.PORT, () => {
	console.log(`Server started on port ${process.env.PORT}`)
})