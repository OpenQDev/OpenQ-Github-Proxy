FROM node:18.12.1-alpine
WORKDIR /app
RUN apk update && apk upgrade && \
	apk add --no-cache bash git
COPY package.json .
RUN yarn
COPY . .
EXPOSE 3000
ENTRYPOINT yarn start