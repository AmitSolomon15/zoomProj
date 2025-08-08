const socket = new WebSocket("ws://localhost:8080/ws");

function SendDataToSock(){
  navigator.mediaDevices.getUserMedia({ video: true, audio: true })
    .then(stream => {
      const recorder = new MediaRecorder(stream, { mimeType: "video/webm" });

      recorder.ondataavailable = (event) => {
        if (event.data.size > 0 && socket.readyState === WebSocket.OPEN) {
          socket.send(event.data);
          
        }
      };

      recorder.start(1000); // שולח כל שנייה (chunk)
    })
    .catch(error => {
      console.error('Error accessing media devices:', error);
    });
};

export default SendDataToSock