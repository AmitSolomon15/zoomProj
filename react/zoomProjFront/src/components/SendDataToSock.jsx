function SendDataToSock(){
  console.log("IM HERE1");
  
   
  const username = document.querySelector(".name").innerText;
  const socket = new WebSocket(`wss://zoomproj-back-ws.onrender.com/ws?username=${username}`);
  socket.binaryType = "arraybuffer";
  const video = document.querySelector(".vid");
  const mediaSource = new MediaSource();
  video.src = URL.createObjectURL(mediaSource);
  
  let source;

  mediaSource.addEventListener("sourceopen", () => {
    console.log("MEDIA SOURCE OPENED");
    source = mediaSource.addSourceBuffer('video/mp4; codecs="avc1.42E01E, mp4a.40.2"');
    //video.play();
  });

  socket.addEventListener("message", (event)=>{
    console.log("CHANGES SAVED");
    console.log("RECIVING MP4 ",event.data);
    const chunk = new Uint8Array(event.data);
    if (!source) return;

    if (!source.updating) {
      try {
        source.appendBuffer(chunk);
      } catch (err) {
        console.error("appendBuffer failed:", err);
      }
      } 
    else {
      source.addEventListener("updateend", () => {
        try {
          source.appendBuffer(chunk);
        }
        catch(e) {}
      }, { once: true });
    }
    
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
      

      //startRecord();
    })
    .catch(error => {
      console.error('Error accessing media devices:', error);
    });
};

export default SendDataToSock