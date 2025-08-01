import axios from 'axios'
function SignUpPage(){
    function handleSubmit(e) {
    e.preventDefault();

    const form = new FormData(e.target);

    axios.post("https://zoomproj-back.onrender.com/submit-data-Sign-Up", form)
      .then(response => {
        console.log("Success:", response.data);
        // document.querySelector("#fName").textContent = "";
        var nav = document.querySelector("nav");
        nav.innerHTML = "";

        localStorage.setItem("username",response.data["username"])
        navigate("/SignedInUserPage");
      })
      .catch(error => {
        console.error("Error:", error);
      });
    }   
    return (
      <form onSubmit={handleSubmit} encType="multipart/form-data">
        <label htmlFor="fName">First Name: </label>
        <input type="text" id="fName" name="fName" />
        <br />

        <label htmlFor="lName">Last Name: </label>
        <input type="text" id="lName" name="lName" />
        <br />

        <label htmlFor="uName">User Name: </label>
        <input type="text" id="uName" name="uName" />
        <br />

        <label htmlFor="pass">Password: </label>
        <input type="password" id="pass" name="pass" />
        <br />

        <button type="submit">Submit</button>
      </form>
    );

}

export default SignUpPage