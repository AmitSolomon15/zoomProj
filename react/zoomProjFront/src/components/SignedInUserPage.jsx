import axios from "axios";
function SignedInUserPage(){
    const username = localStorage.getItem("username");
    console.log(username);
    const tableStyle = {
        position:"absolute",
        right: "10px",
        width:"15%",
        fontSize: "30px",
    }


    function inviteUser(e){
      axios.post("https://zoomproj-back.onrender.com/connect-user-udp",e.innerText)
        .then(Response =>{
          console.log(Response.data);
        })
        .catch(error =>{
          console.log(error);
        })
    }

    window.addEventListener("beforeunload", function () {
    //const username = localStorage.getItem("username");

      const blob = new Blob([JSON.stringify({ username })], {
        type: "application/json"
      });
      
      navigator.sendBeacon("https://zoomproj-back.onrender.com/disconnect", blob);
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
            let tr = document.createElement("tr");
            let td = document.createElement("td");
            //td.onclick()
            td.innerText = element;
            tr.appendChild(td);
            tableBody.appendChild(tr);
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
      <div style={{ color: "white", fontSize: "20px",position: "absolute", top: 0,left: "10px" }}>
        {`${username}`}
      </div>
      <button onClick={getUsers}>show all users</button>
      </div>
        );
}

export default SignedInUserPage; 