import axios from "axios";


const socket = new WebSocket("wss://zoomproj-back-ws.onrender.com/ws");
function SendDataToSock(){
  console.log("IM HERE1");
  
    
  const username = localStorage.getItem("nUsername");
  socket.addEventListener("open",() =>{
    socket.send(JSON.stringify({username}));
  });
  
  console.log("IM HERE2");
  navigator.mediaDevices.getUserMedia({ video: true, audio: true })
    .then(stream => {
      console.log("IM HERE3");
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