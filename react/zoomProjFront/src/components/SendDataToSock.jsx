import axios from "axios";



function SendDataToSock(){
  console.log("IM HERE1");
  
   
  const username = document.querySelector(".name").innerText;
  const socket = new WebSocket(`wss://zoomproj-back-ws.onrender.com/ws?username=${username}`);
  
  console.log("IM HERE2");
  navigator.mediaDevices.getUserMedia({ video: true, audio: true })
    .then(stream => {

      let recorder
      function startRecord(){
      console.log("IM HERE3");
      recorder = new MediaRecorder(stream, { mimeType: "video/webm" });

      recorder.ondataavailable = (event) => {
        console.log("IM HERE3.5");
        if (event.data.size > 0 && socket.readyState === WebSocket.OPEN) {
          console.log("IM HERE4");
          socket.send(event.data);
        }
      };

      recorder.onstop = () =>{
        startRecord();
      };

      recorder.start();
      setTimeout(() => recorder.stop,1000);
      console.log("IM HERE5");
      }

      console.log("IM HERE6")
      startRecord();
    })
    .catch(error => {
      console.error('Error accessing media devices:', error);
    });
};

export default SendDataToSock