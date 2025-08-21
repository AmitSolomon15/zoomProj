import axios from "axios";



function SendDataToSock(){
  console.log("IM HERE1");
  
   
  const username = document.querySelector(".name").innerText;
  const socket = new WebSocket(`wss://zoomproj-back-ws.onrender.com/ws?username=${username}`);

  socket.addEventListener("message", (event)=>{
    console.log("CHANGES SAVED");
    console.log("RECIVING MP4 ",event.data);
    return(
      <div>
      <video width="750" height="500" controls className="vid">
        <source src={event.data} type="video/mp4" className="vidSrc"/>
      </video>
      </div>
    );
  });
  
  console.log("IM HERE2");
  navigator.mediaDevices.getUserMedia({ video: true, audio: true })
    .then(stream => {
      console.log("IM HERE3");
      const recorder = new MediaRecorder(stream, { mimeType: "video/webm" });

      recorder.ondataavailable = (event) => {
        console.log("IM HERE3.5");
        if (event.data.size > 0 && socket.readyState === WebSocket.OPEN) {
          console.log("IM HERE4");
          socket.send(event.data);
        }
      };


      recorder.start(1000);
      
      console.log("IM HERE5");
      

      startRecord();
    })
    .catch(error => {
      console.error('Error accessing media devices:', error);
    });
};

export default SendDataToSock