import SendDataToSock from "./SendDataToSock.jsx";
const username = localStorage.getItem("nUsername");
//const username2 = localStorage.getItem("nUsername2");

function MyBtn(){
    return(
      <div>
      <div style={{ color: "white", fontSize: "20px",position: "absolute", top: 0,left: "10px" }} className="name">
        {`${username}`}
      </div>
      <button onClick={SendDataToSock}>
        start call
      </button>  
      </div>
    );
}

export default MyBtn;