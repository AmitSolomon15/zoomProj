import WebSocket, { WebSocketServer } from 'ws';

const wss = new WebSocketServer({ port: 800 });

wss.on('connection', function connection(ws) {
  console.log('A client connected!');

  ws.on('message', function incoming(message) {
    console.log('received:', message);

    // Echo to all connected clients
    wss.clients.forEach(client => {
      if (client.readyState === WebSocket.OPEN) {
        client.send(message);
      }
    });
  });

  ws.send(JSON.stringify({ message: 'Welcome to the WebSocket server!' }));
});