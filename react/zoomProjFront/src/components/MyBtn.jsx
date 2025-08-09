import SendDataToSock from "./SendDataToSock";
const username = localStorage.getItem("nUsername");
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