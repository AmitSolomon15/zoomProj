import axios from "axios";
import { useNavigate } from "react-router-dom";
import SendDataToSock from "./SendDataToSock.jsx";

function SignedInUserPage(){
  const navigate = useNavigate();
    const username = localStorage.getItem("username");

    const socket = new WebSocket("wss://zoomproj-back-ws.onrender.com/wsConn");
    socket.addEventListener("open",() =>{
      socket.send(JSON.stringify({username}));
      console.log("CONNECTED");
    });
    
    


    console.log(JSON.stringify({ username }));

    const tableStyle = {
        position:"absolute",
        right: "10px",
        width:"15%",
        fontSize: "30px",
    }


    function inviteUser(e){
      console.log(`chk1: ${e.target.innerText}`)
      const form = new FormData();
      const username = document.querySelector(".name").innerText;
      console.log(username);
      localStorage.setItem("nUsername",username);
      localStorage.setItem("nUsername2",e.target.innerText);
      form.append("from",username);
      form.append("to",e.target.innerText);
      form.append("msg",`hello world from ${username}`);
      axios.post("https://zoomproj-back.onrender.com/connect-user-udp",form)
        .then(Response =>{
          console.log(Response.data);
          //navigate('/MyBtn')
          const table = document.querySelector("table");
          table.innerHTML = "";
          const buttn = document.querySelector(".btn");
          console.log(buttn);
          buttn.addEventListener("click",SendDataToSock);
          buttn.innerText = "start call";
        })
        .catch(error =>{
          console.log(error);
        })
    }


    

    window.addEventListener("beforeunload", function () {
      const username = document.querySelector(".name").innerText;

      const form = new FormData();
      form.append("username",username);

      navigator.sendBeacon("https://zoomproj-back.onrender.com/disconnect", form);
    });

    


    function getUsers() {
   
    axios.post("https://zoomproj-back.onrender.com/get-users")
      .then(response => {
        console.log(response.data);
        

        var table = document.createElement("table");
        Object.assign(table.style,tableStyle);
        var tableBody = document.createElement("tbody");
        for (let index = 0; index < response.data.length; index++) {
            const element = response.data[index]["username"];
            if(element != document.querySelector(".name").innerText)
            {
            let tr = document.createElement("tr");
            let td = document.createElement("td");
            td.addEventListener("click", inviteUser);
            td.innerText = element;
            tr.appendChild(td);
            tableBody.appendChild(tr);
            }
        }
        table.appendChild(tableBody);
        document.body.appendChild(table);
        

        
        })
      .catch(error => {
        console.error("Error:", error);
        });

    }   

    return(
      <div>
      <div style={{ color: "white", fontSize: "20px",position: "absolute", top: 0,left: "10px" }} className="name">
        {`${username}`}
      </div>
      <button onClick={getUsers} className="btn">show all users</button>
      </div>
        );
}

export default SignedInUserPage; 