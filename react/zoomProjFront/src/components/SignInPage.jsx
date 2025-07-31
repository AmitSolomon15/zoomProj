import axios from "axios"
import { Navigate, useNavigate } from 'react-router-dom';

function SignInPage(){
    const navigate = useNavigate(); 
    function handleSubmit(e) {
    e.preventDefault();

    const form = new FormData(e.target);

    axios.post("http://localhost:8080/submit-data-Sign-In",form)
      .then(response => {
      
        console.log("Success:", response.data["username"]);
        
        document.body.innerHTML = "";
        var name = document.createElement("div");
        name.innerText = response.data["username"];
        name.style.position = "absolute";
        name.style.top = 0;
        name.style.left = "10px";
        name.style.color = "white";
        name.style.fontSize = "20px";
        document.body.appendChild(name);
        navigate("/SignedInUserPage");
      })
      .catch(error => {
        console.error("Error:", error);
      });
    }   
    return(
        <form onSubmit={handleSubmit} encType="multipart/form-data">
            <label htmlFor="uName">User Name: </label>
            <input type="text" id="uName" name="uName" />
            <br />
            <label htmlFor="pass">Password: </label>
            <input type="password" id="pass" name="pass" />
            <br />
            <input type="submit" id="send" name="send" />
            <br />
        </form>
    );
}

export default SignInPage