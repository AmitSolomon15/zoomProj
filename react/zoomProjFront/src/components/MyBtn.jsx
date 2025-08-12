import SendDataToSock from "./SendDataToSock.jsx";
const username = localStorage.getItem("nUsername");
//const username2 = localStorage.getItem("nUsername2");

function MyBtn(){
    return(
      
      <button onClick={SendDataToSock}>
        start call
      </button>  
    );
}

export default MyBtn;