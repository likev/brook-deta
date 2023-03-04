import { WebSocketServer } from 'ws';

const wss = new WebSocketServer({ port: process.env.PORT });

wss.on('connection', function connection(ws) {
    ws.on('error', console.error);

    ws.on('message', function message(data) {
        console.log('received: %s', data);
        ws.send(data);
    });

    let count = 0, interval = 5; //5s

    setInterval(() => {
        count++;
        ws.send({ count, time: interval * count });
    }, interval);

});