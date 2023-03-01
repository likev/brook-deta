import { WebSocketServer } from 'ws';

const wss = new WebSocketServer({ port: process.env.PORT });

wss.on('connection', function connection(ws) {
  ws.on('error', console.error);

  ws.on('message', function message(data) {
    console.log('received: %s', data);
    ws.send(data);
  });

  ws.send('something');
});