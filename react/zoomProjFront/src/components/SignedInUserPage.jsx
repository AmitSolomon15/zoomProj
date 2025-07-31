import axios from "axios";
function SignedInUserPage(){
    const tableStyle = {
        position:"absolute",
        right: "10px",
        width:"15%",
        fontSize: "30px",
    }

    function getUsers() {
   
    axios.post("http://localhost:8080/get-users")
      .then(response => {
        console.log(response.data);
        var table = document.createElement("table");
        Object.assign(table.style,tableStyle);
        var tableBody = document.createElement("tbody");
        for (let index = 0; index < response.data.length; index++) {
            const element = response.data[index]["username"];
            let tr = document.createElement("tr");
            let td = document.createElement("td");
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
        <button onClick={getUsers}>press me</button>
        );
}

export default SignedInUserPage; 