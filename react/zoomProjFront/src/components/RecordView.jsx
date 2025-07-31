import { useReactMediaRecorder } from "react-media-recorder";
import useWebSocket, { ReadyState } from "react-use-websocket"

const RecordView = () => {
  const { status, startRecording,pauseRecording ,stopRecording, mediaBlobUrl } =
    useReactMediaRecorder({ video: true }); // Set video: true for video recording

    
  return (
    <div>
      <p>Status: {status}</p>
      <button onClick={startRecording}>Start Recording</button>
      <button onClick={pauseRecording}>Pause Recording</button>
      <button onClick={stopRecording}>Stop Recording</button>
      {mediaBlobUrl && <video src={mediaBlobUrl} controls autoPlay loop />}
    </div>
  );
};

export default RecordView 