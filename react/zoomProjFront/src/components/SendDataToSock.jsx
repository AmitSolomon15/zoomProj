import axios from "axios";

const socket = new WebSocket("https://zoomproj-back-ws.onrender.com/ws");
const username = localStorage.getItem("nUsername");
const form = new FormData();
form.append("user",username);
axios.post("https://zoomproj-back-ws.onrender.com/ws",form)
  .then(Response =>{
    console.log("Success");
  })
  .catch(error =>{
          console.log(error);
  })

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