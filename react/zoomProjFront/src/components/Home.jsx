import useWebSocket, { ReadyState } from "react-use-websocket"
import { useEffect } from 'react';

const Home = () => {
  const WS_URL = "http://localhost:8080/ws"
  const { sendJsonMessage, lastJsonMessage, readyState } = useWebSocket(
    WS_URL,
    {
      share: false,
      shouldReconnect: () => true,
    },
  )

  // Run when the connection state (readyState) changes
  useEffect(() => {
    console.log("Connection state changed")
    if (readyState === ReadyState.OPEN) {
      sendJsonMessage({
        event: "subscribe",
        data: {
          channel: "general-chatroom",
        },
      })
    }
  }, [readyState])

  // Run when a new WebSocket message is received (lastJsonMessage)
  useEffect(() => {
    console.log(`Got a new message: ${lastJsonMessage}`)
  }, [lastJsonMessage])

  const Chat = ({ lastJsonMessage }) => {
  return (
    <div>
      <h2>Chat Messages</h2>
      <pre>{JSON.stringify(lastJsonMessage, null, 2)}</pre>
    </div>
  )
}
}

export default Home