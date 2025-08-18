import axios from "axios";
import { useRef } from 'react';


function SendDataToSock(){
  const mediaRecorderRef = useRef(null);
  const mediaSourceRef = useRef(null);
  const sourceBufferRef = useRef(null);

  console.log("IM HERE1");
  
   
  const username = document.querySelector(".name").innerText;
  const socket = new WebSocket(`wss://zoomproj-back-ws.onrender.com/ws?username=${username}`);
  
  console.log("IM HERE2");
  navigator.mediaDevices.getUserMedia({ video: true, audio: true })
    .then(stream => {
      console.log("IM HERE3");
      mediaRecorderRef = new MediaRecorder(stream, { mimeType: "video/webm" });

      mediaSourceRef.current = new MediaSource();
      mediaSourceRef.current.addEventListener('sourceopen', () => {
      sourceBufferRef.current = mediaSourceRef.current.addSourceBuffer('video/webm');
      });

      mediaRecorderRef.current.ondataavailable = (event) => {
      if (event.data.size > 0 && sourceBufferRef.current && !sourceBufferRef.current.updating) {
        sourceBufferRef.current.appendBuffer(event.data);
        socket.send(sourceBufferRef.current);
      }
      };
      mediaRecorderRef.current.start(1000); // שולח כל שנייה (chunk)
    })
    .catch(error => {
      console.error('Error accessing media devices:', error);
    });
};

export default SendDataToSock