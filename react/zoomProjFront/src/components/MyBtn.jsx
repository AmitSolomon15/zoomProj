import SendDataToSock from "./SendDataToSock";

function MyBtn(){
    return(
      <button onClick={SendDataToSock}>
        start call
      </button>  
      
    );
}

export default MyBtn;