const socket = new WebSocket("https://zoomproj-back-ws.onrender.com/ws");
const form = new FormData()
const username = document.querySelector(".name").innerText;
form.append("username",username)

function SendDataToSock(){
  navigator.mediaDevices.getUserMedia({ video: true, audio: true })
    .then(stream => {
      const recorder = new MediaRecorder(stream, { mimeType: "video/webm" });

      recorder.ondataavailable = (event) => {
        if (event.data.size > 0 && socket.readyState === WebSocket.OPEN) {
          socket.send(event.data);
          socket.send(form)
        }
      };

      recorder.start(1000); // שולח כל שנייה (chunk)
    })
    .catch(error => {
      console.error('Error accessing media devices:', error);
    });
};

export default SendDataToSock