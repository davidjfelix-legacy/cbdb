FROM node

RUN npm install -g webpack
RUN npm install --save react react-dom babel-preset-react
RUN mkdir -p /opt/cbdb

WORKDIR /opt/cbdb
COPY ./main.js /opt/cbdb/main.js
