import axios from "axios";



function SendDataToSock(){
  console.log("IM HERE1");
  
    
  const username = localStorage.getItem("nUsername");
  const socket = new WebSocket(`wss://zoomproj-back-ws.onrender.com/ws?username=${username}`);
  
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