import axios from "axios"
import { useNavigate } from "react-router-dom";


function SignInPage(){
    const navigate = useNavigate(); 
    function handleSubmit(e) {
    e.preventDefault();

    //console.log("bbbbb");
    const form = new FormData(e.target);

    axios.post("https://zoomproj-back.onrender.com/submit-data-Sign-In",form)
      .then(response => {
        console.log(response)
        console.log(response.data)
        console.log("Success:", response.data["username"]);
        
        var nav = document.querySelector("nav");
        nav.innerHTML = "";

        localStorage.setItem("username",response.data["username"])
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