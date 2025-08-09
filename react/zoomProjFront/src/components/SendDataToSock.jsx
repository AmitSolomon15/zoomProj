import axios from "axios";



function SendDataToSock(){
  function init(){
    const socket = new WebSocket("wss://zoomproj-back-ws.onrender.com/ws");
    const username = localStorage.getItem("nUsername");
    socket.addEventListener("open",() =>{
      socket.send(JSON.stringify({username}));
    });
  }
  init();
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